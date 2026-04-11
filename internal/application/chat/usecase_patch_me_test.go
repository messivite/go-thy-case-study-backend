package chat

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"testing"

	"github.com/disintegration/imaging"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

func TestUseCase_PatchMe_emptyReturnsErr(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	reg := provider.NewRegistry("x")
	uc := NewUseCase(mem, &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}, reg, mem)

	_, err := uc.PatchMe(ctx, "u1", domain.ProfilePatch{}, nil)
	if err == nil {
		t.Fatal("expected ErrProfilePatchEmpty")
	}
	if err != domain.ErrProfilePatchEmpty {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_PatchMe_displayName(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	reg := provider.NewRegistry("x")
	uc := NewUseCase(mem, &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}, reg, mem)

	name := "Ada"
	p, err := uc.PatchMe(ctx, "u1", domain.ProfilePatch{DisplayName: &name}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.DisplayName != "Ada" {
		t.Fatalf("displayName: %q", p.DisplayName)
	}
}

func TestUseCase_PatchMe_withAvatarPNG(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	reg := provider.NewRegistry("x")
	uc := NewUseCase(mem, &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}, reg, mem)

	img := imaging.New(600, 200, color.NRGBA{R: 10, G: 200, B: 30, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	p, err := uc.PatchMe(ctx, "user-avatar-1", domain.ProfilePatch{}, buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if p.AvatarURL == "" || p.ID != "user-avatar-1" {
		t.Fatalf("profile: %+v", p)
	}
}
