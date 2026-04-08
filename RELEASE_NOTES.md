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
2. Tag vurmadan önce o maddeleri yeni bir `## [x.y.z] - YYYY-AA-GG` başlığı altına taşı (tarih = release günü).
3. `git tag -a vX.Y.Z -m "vX.Y.Z"` → `git push origin vX.Y.Z`.
4. GitHub Release açıklamasında istersen kısa özet + `CHANGELOG` linki koy.

## Release açıklaması şablonu (kopyala-yapıştır)

```markdown
## Öne çıkanlar
- …

## Binary’ler
- `thy-case-study-backend` — API
- `thy-case-llm` — CLI

Tam liste: [CHANGELOG.md](CHANGELOG.md#changelog)
```

## Not

İlk stabil sürüme kadar `CHANGELOG.md` içinde “tag öncesi özet” bloğu durabilir; `v1.0.0` veya senin belirlediğin çizgiden sonra her şeyi semver başlıklarına bölmek yeterli.
