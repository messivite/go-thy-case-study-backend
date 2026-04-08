# Release notes — nasıl kullanıyorum?

Bu repo iki şeyi ayırıyorum:

| Ne | Amaç |
|----|------|
| **CHANGELOG.md** | Kullanıcı / geliştirici için sürüm bazlı, okunaklı değişiklik listesi. Elle güncellersin (veya tag öncesi özet burada durur). |
| **GitHub Release sayfası** | Tag ile çıkan sürümün duyurusu; GoReleaser şu an commit’lerden otomatik özet çekiyor (`.goreleaser.yaml` → `changelog: use: github`). |

## Tag atınca ne oluyor?

1. `v*.*.*` formatında tag push (ör. `v0.0.4`).
2. `.github/workflows/release.yml` çalışır: `go mod tidy` diff yok, `build`, `test`, `vet`.
3. GoReleaser binary’leri üretir, GitHub Release oluşturur / günceller.

Manuel publish için workflow’da `publish` + `tag` alanlarını doldurman yeterli (detay workflow dosyasında).

## Yeni sürüm çıkarırken checklist

1. `CHANGELOG.md` içinde `[Unreleased]` altına yaptıklarını yaz.
2. Aşağıdaki **Unreleased — kullanıcı özeti** bölümünü güncelle veya tag sonrası sil / sürüme taşı.
3. Tag vurmadan önce `CHANGELOG` maddelerini yeni bir `## [x.y.z] - YYYY-AA-GG` başlığı altına taşı (tarih = release günü).
4. `git tag -a vX.Y.Z -m "vX.Y.Z"` → `git push origin vX.Y.Z`.
5. GitHub Release açıklamasında kısa özet (bu dosyadaki madde işleri veya CHANGELOG’dan seçtiklerin) + `CHANGELOG` linki.

## Unreleased — kullanıcı özeti

GitHub Release metnine kısa yapıştırmalık özet (tag atınca sürüm numarasıyla güncelle veya `CHANGELOG`’a taşıdıktan sonra temizle):

- **`thy-case-llm v0.3.0+`:** `deploy list | show <id> | init <id>` — Railway / Fly.io / Vercel (örnek ön yüz rewrite) şablonlarını repoya yazar; `thy-case-llm deploy` veya `thy-case-llm help` ile detay.
- Yerelde tüm binary’ler: repo kökünde `go install ./cmd/...` → `GOPATH/bin` içinde `api`, `server`, `thy-case-llm` (release tarball’larında API binary adı GoReleaser ile **`thy-case-study-backend`**).
- Uzun dokümantasyon: README sonundaki **[Deploy](README.md#deploy)** bölümü.

## Release açıklaması şablonu (kopyala-yapıştır)

```markdown
## Öne çıkanlar
- …

## Binary’ler (GoReleaser / tarball)
- `thy-case-study-backend` — API (`cmd/api`)
- `thy-case-llm` — LLM provider + **deploy** şablon CLI (`cmd/thy-case-llm`)

Yerel `go install ./cmd/...`: `api`, `server`, `thy-case-llm`

Tam liste: [CHANGELOG.md](CHANGELOG.md)
```

## Not

İlk stabil sürüme kadar `CHANGELOG.md` içinde “tag öncesi özet” bloğu durabilir; `v1.0.0` veya senin belirlediğin çizgiden sonra her şeyi semver başlıklarına bölmek yeterli.
