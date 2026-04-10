package catalog

import (
	"strings"

	"github.com/messivite/go-thy-case-study-backend/internal/config"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
)

// SupportedModelsFromRegistry çalışan provider kayıtları + yerleşik şablonlardan katalog üretir.
// Şablonu olmayan özel provider için yalnızca registry varsayılan modeli eklenir.
func SupportedModelsFromRegistry(reg *provider.Registry) []domain.SupportedModel {
	if reg == nil {
		return nil
	}
	var out []domain.SupportedModel
	for _, meta := range reg.List() {
		name := meta.Name
		tpl, ok := config.GetTemplate(name)
		if ok {
			for _, m := range tpl.Models {
				m = strings.TrimSpace(m)
				if m == "" {
					continue
				}
				dn := m
				out = append(out, domain.SupportedModel{
					Provider:       name,
					ModelID:        m,
					DisplayName:    dn,
					SupportsStream: tpl.SupportsStream,
				})
			}
			continue
		}
		dm := strings.TrimSpace(meta.DefaultModel)
		if dm == "" {
			continue
		}
		out = append(out, domain.SupportedModel{
			Provider:       name,
			ModelID:        dm,
			DisplayName:    dm,
			SupportsStream: meta.SupportsStream,
		})
	}
	return out
}
