package chat

import "context"

// SupportedModel tek bir kullanılabilir provider + model çifti (katalog satırı).
type SupportedModel struct {
	Provider       string
	ModelID        string
	DisplayName    string
	SupportsStream bool
}

// SupportedModelsCatalog API açılışında sync + isteklerde doğrulama + GET /api/models.
// Bellek modunda MemoryRepository; Supabase’de aynı struct üzerinden RPC/REST.
type SupportedModelsCatalog interface {
	SyncSupportedModels(ctx context.Context, models []SupportedModel) error
	ListActiveSupportedModels(ctx context.Context) ([]SupportedModel, error)
	IsModelActive(ctx context.Context, provider, modelID string) (bool, error)
}
