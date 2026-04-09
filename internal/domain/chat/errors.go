package chat

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrInvalidRole         = errors.New("invalid message role")
	ErrMissingContent      = errors.New("missing message content")
	ErrUnauthorized        = errors.New("not authorized")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidSessionID    = errors.New("invalid session id")

	ErrProviderAuthFailed  = errors.New("provider authentication failed")
	ErrProviderTimeout     = errors.New("provider request timed out")
	ErrProviderRateLimited = errors.New("provider rate limit exceeded")
	ErrProviderUnavailable = errors.New("provider temporarily unavailable")
	ErrProviderBadRequest  = errors.New("provider rejected request")
	ErrUserCancelled       = errors.New("generation cancelled by user")

	ErrQuotaDailyExceeded  = errors.New("daily token quota exceeded")
	ErrQuotaWeeklyExceeded = errors.New("weekly token quota exceeded")
)
