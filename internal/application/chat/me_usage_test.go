package chat

import (
	"context"
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
)

type meUsageQuotaStub struct {
	q   domain.UserQuota
	u   domain.UserTokenUsage
	err error
}

func (s *meUsageQuotaStub) GetUserQuota(context.Context, string) (domain.UserQuota, error) {
	if s.err != nil {
		return domain.UserQuota{}, s.err
	}
	return s.q, nil
}

func (s *meUsageQuotaStub) GetUserTokenUsage(context.Context, string) (domain.UserTokenUsage, error) {
	if s.err != nil {
		return domain.UserTokenUsage{}, s.err
	}
	return s.u, nil
}

func (s *meUsageQuotaStub) FailPendingLog(context.Context, string, string, string, int) error {
	return nil
}

func (s *meUsageQuotaStub) CancelPendingLog(context.Context, string) error { return nil }

func (s *meUsageQuotaStub) SetUsageLog(context.Context, string, int, int, int) error { return nil }

func TestUseCase_MeUsage_mergesParallelResults(t *testing.T) {
	stub := &meUsageQuotaStub{
		q: domain.UserQuota{
			UserID:           "u1",
			DailyTokenLimit:  10_000,
			WeeklyTokenLimit: 50_000,
			QuotaBypass:      false,
		},
		u: domain.UserTokenUsage{DailyTotal: 100, WeeklyTotal: 900},
	}
	uc := NewUseCase(nil, stub, provider.NewRegistry("openai"))
	got, err := uc.MeUsage(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if got.QuotaBypass {
		t.Fatal("expected quotaBypass false")
	}
	if got.Daily.LimitTokens != 10_000 || got.Daily.UsedTokens != 100 {
		t.Fatalf("daily: %+v", got.Daily)
	}
	if got.Weekly.LimitTokens != 50_000 || got.Weekly.UsedTokens != 900 {
		t.Fatalf("weekly: %+v", got.Weekly)
	}
}
