package chat

import "context"

// UserQuota represents one row from user_llm_usage_quota.
type UserQuota struct {
	UserID           string
	DailyTokenLimit  int
	WeeklyTokenLimit int
	QuotaBypass      bool
}

// UserTokenUsage holds aggregated token counts for a time window.
type UserTokenUsage struct {
	DailyTotal  int
	WeeklyTotal int
}

// QuotaRepository abstracts quota + audit operations that the use case needs.
type QuotaRepository interface {
	GetUserQuota(ctx context.Context, userID string) (UserQuota, error)
	GetUserTokenUsage(ctx context.Context, userID string) (UserTokenUsage, error)
	FailPendingLog(ctx context.Context, userMessageID, errorSummary, errorCode string, httpStatus int) error
	CancelPendingLog(ctx context.Context, userMessageID string) error
	SetUsageLog(ctx context.Context, userMessageID string, prompt, completion, total int) error
}
