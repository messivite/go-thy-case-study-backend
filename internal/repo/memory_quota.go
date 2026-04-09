package repo

import (
	"context"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

var _ domain.QuotaRepository = (*MemoryQuotaRepository)(nil)

// MemoryQuotaRepository is an in-memory stub for development / tests.
type MemoryQuotaRepository struct{}

func NewMemoryQuotaRepository() *MemoryQuotaRepository { return &MemoryQuotaRepository{} }

func (m *MemoryQuotaRepository) GetUserQuota(_ context.Context, _ string) (domain.UserQuota, error) {
	return domain.UserQuota{
		DailyTokenLimit:  100_000,
		WeeklyTokenLimit: 500_000,
		QuotaBypass:      true,
	}, nil
}

func (m *MemoryQuotaRepository) GetUserTokenUsage(_ context.Context, _ string) (domain.UserTokenUsage, error) {
	return domain.UserTokenUsage{}, nil
}

func (m *MemoryQuotaRepository) FailPendingLog(_ context.Context, _, _, _ string, _ int) error {
	return nil
}

func (m *MemoryQuotaRepository) CancelPendingLog(_ context.Context, _ string) error {
	return nil
}

func (m *MemoryQuotaRepository) SetUsageLog(_ context.Context, _ string, _, _, _ int) error {
	return nil
}
