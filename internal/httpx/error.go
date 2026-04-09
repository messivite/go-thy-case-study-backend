package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func Unauthorized(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
}

func Forbidden(w http.ResponseWriter) {
	WriteError(w, http.StatusForbidden, "forbidden", "Forbidden")
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "bad_request", message)
}

func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "not_found", message)
}

func Internal(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
}

func ProviderAuthFailed(w http.ResponseWriter) {
	WriteError(w, http.StatusBadGateway, "provider_auth_failed", "Provider authentication failed")
}

func ProviderTimeout(w http.ResponseWriter) {
	WriteError(w, http.StatusGatewayTimeout, "provider_timeout", "Provider request timed out")
}

func ProviderRateLimited(w http.ResponseWriter) {
	WriteError(w, http.StatusTooManyRequests, "provider_rate_limited", "Provider rate limit exceeded")
}

func ProviderUnavailable(w http.ResponseWriter) {
	WriteError(w, http.StatusBadGateway, "provider_unavailable", "Provider temporarily unavailable")
}

func QuotaDailyExceeded(w http.ResponseWriter) {
	WriteError(w, http.StatusTooManyRequests, "llm_quota_daily_exceeded", "Daily token quota exceeded")
}

func QuotaWeeklyExceeded(w http.ResponseWriter) {
	WriteError(w, http.StatusTooManyRequests, "llm_quota_weekly_exceeded", "Weekly token quota exceeded")
}

func GenerationCancelled(w http.ResponseWriter) {
	// 499 mirrors client-closed behavior used by common proxies.
	WriteError(w, 499, "generation_cancelled", "Generation cancelled by user")
}
