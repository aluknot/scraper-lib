# Migración desde Rissy

> **Status: MVP**

## Paso 0: Crear scraper-lib ✅

**Completado.** Código extraído y modularizado de `rissy/internal/fetcher/fetcher.go`:

| Código original en Rissy | Paquete en scraper-lib | Mejora |
|---|---|---|
| `fetchWithRetry` | `fetch/http.go` | API independiente, sin depender de config/DB |
| `scrapeWithReadability` | `extractors/readability.go` | Interfaz Extractor, retorna ExtractResult |
| `scrapeWithTrafilatura` | `extractors/trafilatura.go` | Idem |
| `scrapeWithFallback` | `extractors/fallback.go` | Idem |
| YouTube placeholders | `fetch/embeds.go` | → Genérico: YouTube + Vimeo + Twitter + Spotify + Instagram + TikTok con UUIDs |
| `sanitizeHTML` | `fetch/sanitize.go` | Permite iframes para embeds restaurados |
| `isLikelyTeaser` | `detection/teaser.go` | + `NeedsScraping` combinada |

**Nuevo en scraper-lib (no estaba en Rissy):**
- Paywall detection (`detection/paywall.go`)
- Domain rules YAML (`extractors/domain_specific.go`)
- Output pipeline (`output/article.go`)
- Extractor Chain con DI (`extractors/chain.go`)
- API pública `Extract()` / `ExtractHTML()` (`scraper.go`)
- 22 tests unitarios

## Paso 1: Crear scraper-service

1. Importar `scraper-lib`
2. HTTP API con chi (`POST /scrape`, `GET /health`)
3. PostgreSQL: migrations + worker de limpieza de caché
4. Rate limiter PostgreSQL con upsert atómico

**Estado:** Pendiente.

## Paso 2: Integrar Rissy con scraper-lib

```bash
cd rissy
go get github.com/aluknot/scraper-lib@latest
```

Reemplazar el código de scraping interno con llamadas a `scraperlib.Extract()`.

**Estado:** Pendiente.

## Paso 3: Agregar extractores especializados

| Caso de uso | Qué agregar |
|---|---|
| Finanzas | Price extractor + output |
| Markdown | Output formatter con templates YAML (antes "Obsidian") |
| Job scraper | Job extractor + output |

**Estado:** Pendiente — esperar caso concreto.

---

[Volver al índice](README.md)
