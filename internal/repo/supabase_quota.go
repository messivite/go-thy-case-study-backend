package repo

import (
	"context"
	"fmt"
	"net/http"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

var _ domain.QuotaRepository = (*SupabaseRepository)(nil)

// GetUserQuota reads the single user_llm_usage_quota row for userID.
func (r *SupabaseRepository) GetUserQuota(ctx context.Context, userID string) (domain.UserQuota, error) {
	path := fmt.Sprintf("/user_llm_usage_quota?user_id=eq.%s", userID)

	var rows []quotaRow
	if err := r.doRequest(ctx, http.MethodGet, path, nil, &rows); err != nil {
		return domain.UserQuota{}, fmt.Errorf("get quota: %w", err)
	}
	if len(rows) == 0 {
		return domain.UserQuota{}, fmt.Errorf("quota row not found for user %s", userID)
	}
	return rows[0].toDomain(), nil
}

// GetUserTokenUsage returns daily + weekly aggregated total_tokens from llm_interaction_log.
// Daily = today (UTC), weekly = last 7 days (UTC).
func (r *SupabaseRepository) GetUserTokenUsage(ctx context.Context, userID string) (domain.UserTokenUsage, error) {
	path := fmt.Sprintf(
		"/rpc/llm_get_user_token_usage?p_user_id=%s",
		userID,
	)
	body := map[string]any{"p_user_id": userID}
	var rows []usageAggRow
	if err := r.doRequest(ctx, http.MethodPost, "/rpc/llm_get_user_token_usage", body, &rows); err != nil {
		return domain.UserTokenUsage{}, fmt.Errorf("get usage: %w (%s)", err, path)
	}
	if len(rows) == 0 {
		return domain.UserTokenUsage{}, nil
	}
	return domain.UserTokenUsage{
		DailyTotal:  rows[0].DailyTotal,
		WeeklyTotal: rows[0].WeeklyTotal,
	}, nil
}

// FailPendingLog calls llm_fail_pending_for_user_message RPC.
func (r *SupabaseRepository) FailPendingLog(ctx context.Context, userMessageID, errorSummary, errorCode string, httpStatus int) error {
	body := map[string]any{
		"p_user_message_id": userMessageID,
		"p_error_summary":   errorSummary,
	}
	if err := r.doRequest(ctx, http.MethodPost, "/rpc/llm_fail_pending_for_user_message", body, nil); err != nil {
		return fmt.Errorf("fail pending log: %w", err)
	}
	if errorCode != "" || httpStatus != 0 {
		patch := map[string]any{}
		if errorCode != "" {
			patch["error_code"] = errorCode
		}
		if httpStatus != 0 {
			patch["provider_http_status"] = httpStatus
		}
		patchPath := fmt.Sprintf("/llm_interaction_log?user_message_id=eq.%s", userMessageID)
		_ = r.doRequest(ctx, http.MethodPatch, patchPath, patch, nil)
	}
	return nil
}

// SetUsageLog calls llm_set_usage_for_user_message RPC.
func (r *SupabaseRepository) SetUsageLog(ctx context.Context, userMessageID string, prompt, completion, total int) error {
	body := map[string]any{
		"p_user_message_id":   userMessageID,
		"p_prompt_tokens":     prompt,
		"p_completion_tokens": completion,
		"p_total_tokens":      total,
	}
	return r.doRequest(ctx, http.MethodPost, "/rpc/llm_set_usage_for_user_message", body, nil)
}

type quotaRow struct {
	UserID           string `json:"user_id"`
	DailyTokenLimit  int    `json:"daily_token_limit"`
	WeeklyTokenLimit int    `json:"weekly_token_limit"`
	QuotaBypass      bool   `json:"quota_bypass"`
}

func (q quotaRow) toDomain() domain.UserQuota {
	return domain.UserQuota{
		UserID:           q.UserID,
		DailyTokenLimit:  q.DailyTokenLimit,
		WeeklyTokenLimit: q.WeeklyTokenLimit,
		QuotaBypass:      q.QuotaBypass,
	}
}

type usageAggRow struct {
	DailyTotal  int `json:"daily_total"`
	WeeklyTotal int `json:"weekly_total"`
}
