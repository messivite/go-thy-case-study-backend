package chat

import "errors"

// LLMErrorCode returns a short machine-readable code for an LLM-related error.
func LLMErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrProviderRateLimited):
		return "rate_limited"
	case errors.Is(err, ErrProviderUnavailable):
		return "upstream_5xx"
	case errors.Is(err, ErrProviderTimeout):
		return "timeout"
	case errors.Is(err, ErrProviderAuthFailed):
		return "auth_failed"
	case errors.Is(err, ErrProviderBadRequest):
		return "bad_request"
	case errors.Is(err, ErrUserCancelled):
		return "user_cancelled"
	default:
		return "unknown"
	}
}

// LLMHTTPStatus maps a domain error to the upstream HTTP status (best effort).
func LLMHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrProviderRateLimited):
		return 429
	case errors.Is(err, ErrProviderUnavailable):
		return 503
	case errors.Is(err, ErrProviderTimeout):
		return 504
	case errors.Is(err, ErrProviderAuthFailed):
		return 401
	case errors.Is(err, ErrProviderBadRequest):
		return 400
	default:
		return 0
	}
}
