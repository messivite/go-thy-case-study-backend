package repo

import (
	"context"
	"testing"
)

func TestMemoryQuotaRepository_Bypass(t *testing.T) {
	r := NewMemoryQuotaRepository()
	q, err := r.GetUserQuota(context.Background(), "test-user")
	if err != nil {
		t.Fatal(err)
	}
	if !q.QuotaBypass {
		t.Fatal("memory quota should bypass by default")
	}
}

func TestMemoryQuotaRepository_Usage(t *testing.T) {
	r := NewMemoryQuotaRepository()
	u, err := r.GetUserTokenUsage(context.Background(), "test-user")
	if err != nil {
		t.Fatal(err)
	}
	if u.DailyTotal != 0 || u.WeeklyTotal != 0 {
		t.Fatalf("expected zero usage, got %+v", u)
	}
}

func TestMemoryQuotaRepository_NoOps(t *testing.T) {
	r := NewMemoryQuotaRepository()
	if err := r.FailPendingLog(context.Background(), "id", "err", "code", 500); err != nil {
		t.Fatal(err)
	}
	if err := r.CancelPendingLog(context.Background(), "id"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetUsageLog(context.Background(), "id", 10, 20, 30); err != nil {
		t.Fatal(err)
	}
}
