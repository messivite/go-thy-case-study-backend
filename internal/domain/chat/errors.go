package chat

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrInvalidRole         = errors.New("invalid message role")
	ErrMissingContent      = errors.New("missing message content")
	ErrUnauthorized        = errors.New("not authorized")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidSessionID    = errors.New("invalid session id")
	ErrInvalidMessageID    = errors.New("invalid message id")
	ErrMessageNotFound     = errors.New("message not found")
	ErrInvalidSearchCursor = errors.New("invalid search cursor")
	ErrSearchQueryTooShort = errors.New("search query must be at least 2 characters")
	ErrInvalidDirection    = errors.New("invalid direction")

	ErrProviderAuthFailed  = errors.New("provider authentication failed")
	ErrProviderTimeout     = errors.New("provider request timed out")
	ErrProviderRateLimited = errors.New("provider rate limit exceeded")
	ErrProviderUnavailable = errors.New("provider temporarily unavailable")
	ErrProviderBadRequest  = errors.New("provider rejected request")
	ErrUserCancelled       = errors.New("generation cancelled by user")

	ErrQuotaDailyExceeded  = errors.New("daily token quota exceeded")
	ErrQuotaWeeklyExceeded = errors.New("weekly token quota exceeded")

	// ErrModelDiscontinued model katalogda yok veya pasifleştirilmiş (sync / operatör).
	ErrModelDiscontinued = errors.New("bu model artık desteklenmiyor veya kullanıma kapatıldı")

	ErrInvalidImagePayload = errors.New("geçersiz veya desteklenmeyen görsel")
	ErrAvatarTooLarge      = errors.New("avatar dosyası çok büyük")
	ErrProfilePatchEmpty   = errors.New("güncellenecek alan veya avatar yok")
)
