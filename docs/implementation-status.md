# Estado de Implementación

> **Fecha:** 2026-04-18
> **Versión:** v10.3 (Debug logging detallado)

---

## ✅ Implementado (v10.0 - v10.3)

| Componente | Archivos | Tests |
|---|---|---|
| Fetch con retry | `internal/fetch/http.go` | ✅ |
| Embeds genéricos (UUID) | `embeds/` | ✅ 7 tests |
| Sanitización | `sanitize/` | ✅ |
| Extractor Chain | `extractors/chain.go` | ✅ 4 tests |
| Readability | `extractors/readability.go` | ✅ |
| Trafilatura | `extractors/trafilatura.go` | ✅ |
| Fallback Extractor (goquery-based) | `extractors/fallback.go` | ✅ |
| Domain rules YAML | `extractors/domain_specific.go` | ✅ 3 tests |
| Paywall detection | `internal/detection/paywall.go` | ✅ |
| Output pipeline | `output/article.go` | ✅ |
| Types compartidos | `types/` | ✅ |
| API pública (con custom chain) | `scraper.go`, `scraper_test.go` | ✅ 16 tests |
| Cache (InMemory + File) | `internal/cache/` | ✅ 16 tests |
| Output Markdown + Templates | `internal/markdown/` | ✅ 10 tests |
| HTTP Advanced (UA rotation, cookies, referrer) | `internal/fetch/http_advanced.go` | ✅ 16 tests |
| Error types | `errors.go` | ✅ 5 tests |
| **Componentes públicos exportados** | `extractors/`, `output/`, `types/`, `embeds/`, `sanitize/` | Nueva API v10 |
| **Pipeline stages opcionales** | `NoEmbeds`, `NoSanitize`, `NoPaywallDetection`, `DisableCache` | ✅ Tests |
| **Platform extractors** | `extractors/platforms/` | ✅ |
| **YouTube extractor** | `extractors/platforms/youtube/` | ✅ |
| **GitHub extractor** | `extractors/platforms/github/` | ✅ |
| **Debug logging** | `Options.Debug`, chain.go logs | ✅ |
| **Total** | **~40 archivos** | **70+ tests pasando** |

---

## ❌ Pendiente (MVP librería)

| Componente | Notas |
|---|---|
| integration tests | URLs reales con build tag |

---

## ❌ Pendiente (Phase 2+)

| Componente | Fase | Notas |
|---|---|---|
| qualityScore completo | Phase 2 | Scoring ponderado (word count, links, images, metadata, extractor) |
| price extractor | Phase 2 | Patrones e-commerce (Amazon, MercadoLibre) |
| job extractor | Phase 2 | Patrones job postings (LinkedIn, Indeed) |
| circuit breaker | Phase 3 | Por dominio, configurable vía YAML |
| bot detection estructural | Phase 3 | Cloudflare, PerimeterX, SPA vacía |
| proxy rotation | Phase 3 | Sobre http_advanced |
| browser service | Phase 3 | Playwright en contenedor Python separado |
| archive.org fallback | Phase 3 | Wayback Machine |
| métricas Prometheus | Phase 3 | /metrics endpoint |

---

## ❌ Pendiente (scraper-service)

| Componente | Fase | Notas |
|---|---|---|
| HTTP API con chi | MVP | POST /scrape, GET /health |
| PostgreSQL cache | MVP | scrape_cache table + cleanup worker |
| PostgreSQL rate limiter | MVP | rate_limits table con upsert atómico |
| migrations SQL | MVP | 001_init.sql |

---

## ⚠️ Issues de Calidad de Código

> Ver [docs/code-quality.md](code-quality.md) para detalle completo de cada issue.

### ✅ Resueltos en v10.0

| # | Issue | Estado |
|---|-------|--------|
| C1 | CollyExtractor hace HTTP request redundante | ✅ FIXED |
| C2 | `defer resp.Body.Close()` dentro de retry loop | ✅ FIXED |
| H1 | `getAttr()` duplicada en 2 archivos | ✅ FIXED |
| H2 | Domain extraction duplicada 3 veces | ✅ FIXED |
| H3 | Dos implementaciones de HTML→Markdown | ✅ FIXED |
| H4 | `isValidSelector()` no-op | ✅ FIXED |
| H5 | Regex inválidos silenciados en YAML | ✅ FIXED |
| M5 | `var _ = embed.FS{}` dead code | ✅ FIXED |
| M6 | `contains()` manual vs `strings.Contains` | ✅ FIXED |
| M7 | `normalizeWhitespace()` O(n²) | ✅ FIXED |
| M9 | Errores ignorados | ✅ FIXED |

### 🔄 Pendientes

| # | Issue | Prioridad |
|---|-------|-----------|
| M1 | Default cache sin límite de tamaño | ✅ RESUELTO |
| L2 | Cero logging en pipeline | ✅ RESUELTO |
| M2 | teaser.go preparado para Phase 2 | Baja |
| M3 | ErrCircuitOpen preparado para Phase 3 | Baja |
| M4 | PriceInfo/JobInfo preparados para Phase 2 | Baja |
| M8 | `testdata/` vacío | ✅ RESUELTO |
| L1 | `MemoryCache.Get()` usa Lock vs RLock | Baja |
| L3 | Retry config hardcoded | Baja |
| L4 | `github.com` en teaser domains | Baja |
| L5 | `GenerateFilename()` truncación 80 chars | Baja |

[Volver al índice](README.md)
