package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/observability"
)

type ClientConfig struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryBaseWait time.Duration
}

var DefaultClientConfig = ClientConfig{
	Timeout:       30 * time.Second,
	MaxRetries:    2,
	RetryBaseWait: 500 * time.Millisecond,
}

var StreamClientConfig = ClientConfig{
	Timeout:       120 * time.Second,
	MaxRetries:    0,
	RetryBaseWait: 0,
}

func NewHTTPClient(cfg ClientConfig) *http.Client {
	return &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

func DoWithRetry(ctx context.Context, client *http.Client, req *http.Request, cfg ClientConfig, providerName string) (*http.Response, error) {
	var lastErr error
	maxAttempts := cfg.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			wait := cfg.RetryBaseWait * time.Duration(1<<(attempt-1))
			observability.Warn("provider.retry", map[string]any{
				"provider": providerName,
				"attempt":  attempt + 1,
				"wait_ms":  wait.Milliseconds(),
			})

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%w: context cancelled during retry", domain.ErrProviderTimeout)
			case <-time.After(wait):
			}

			req = req.Clone(ctx)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 && attempt < maxAttempts-1 {
			resp.Body.Close()
			lastErr = MapHTTPError(resp.StatusCode, providerName)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, lastErr)
}

func MapHTTPError(statusCode int, providerName string) error {
	switch {
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return fmt.Errorf("%w: %s returned %d", domain.ErrProviderAuthFailed, providerName, statusCode)
	case statusCode == http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s returned 429", domain.ErrProviderRateLimited, providerName)
	case statusCode == http.StatusBadRequest:
		return fmt.Errorf("%w: %s returned 400", domain.ErrProviderBadRequest, providerName)
	case statusCode == http.StatusRequestTimeout:
		return fmt.Errorf("%w: %s returned %d", domain.ErrProviderTimeout, providerName, statusCode)
	case statusCode >= 500:
		return fmt.Errorf("%w: %s returned %d", domain.ErrProviderUnavailable, providerName, statusCode)
	default:
		return fmt.Errorf("provider %s returned unexpected status %d", providerName, statusCode)
	}
}

func IsRetryable(statusCode int) bool {
	return statusCode >= 500 || statusCode == http.StatusTooManyRequests
}
