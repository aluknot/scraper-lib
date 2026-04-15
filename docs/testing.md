# Estrategia de Testing

> **Status: MVP** — 59 tests pasando en 30 archivos.
> Integration tests implementados con build tag `integration`.
> Fixtures HTML pendientes de crear (ver sección Fixtures).

## Pirámide de Tests

```
         ┌─────────────┐
         │ Integration │   ← URLs reales (build tag: integration)
         └─────────────┘
        ┌───────────────┐
        │    Service    │  ← StrategyChain + ExtractorChain completos
       └─────────────────┘
      ┌───────────────────┐
      │     Unit          │ ← Cada extractor, detector, util por separado
     └─────────────────────┘
```

## Tests Actuales (v10.1)

| Archivo | Tests | Qué cubre |
|---------|-------|-----------|
| `scraper_test.go` | 11 | Extract completo (simple, YouTube, consecutivos, 404, ExtractHTML, backwards compat, defaults, custom chain, no fallback, markdown, advanced HTTP) |
| `scraper_integration_test.go` | 7 | URLs reales (Wikipedia, GitHub), YouTube embeds, markdown output, advanced HTTP, múltiples outputs, diferentes extractores, cache timing |
| `errors_test.go` | 5 | Error message formatting, `errors.As` compatibility |
| `internal/fetch/embeds_test.go` | 8 | YouTube, Vimeo, consecutivos, placeholder collision, sanitización, Twitter variants, figure wrapper |
| `internal/extractors/chain_test.go` | 10+ | Orden, primer éxito, fallback, todos fallan, domain extractor, validación de reglas |
| `internal/fetch/http_advanced_test.go` | 16+ | UA rotation, referrer, accept language, headers, nil options, cookies, jitter |
| `internal/cache/memory_test.go` | 7 | Get/Set, miss, TTL, delete, clear, stats, concurrent access |
| `internal/cache/file_test.go` | 7+ | Get/Set, miss, TTL, delete, clear, stats, dir creation |
| `internal/markdown/template_test.go` | 7 | Basic conversion, links/images, template resolve, filename, tags, YAML quoting, timestamp |
| **Total** | **59** | |

---

## ⚠️ Pendiente: Fixtures HTML

**Estado:** El directorio `testdata/` está **vacío**. Los siguientes fixtures son requeridos por la estrategia de testing pero aún no se crearon:

```
testdata/
├── article_with_embeds.html     # Artículo con embeds de YouTube/Twitter
├── github_readme.html           # README de GitHub con estructura típica
├── wikipedia_article.html       # Artículo Wikipedia con referencias
├── paywall_detected.html        # Página con señales de paywall
├── cloudflare_challenge.html    # Cloudflare challenge page
├── empty_spa.html               # SPA vacía (JS-rendered sin contenido)
└── price_product.html           # Página de producto e-commerce
```

**Impacto:** Sin estos fixtures, los tests que los necesitan usan httptest servers inline en vez de archivos. Esto funciona pero:
- Los HTML inline en tests son más difíciles de mantener
- No se pueden testear escenarios complejos con HTML real
- La cobertura de edge cases es menor

**Recomendación:** Crear al mínimo los 3 primeros fixtures. Ver [docs/code-quality.md](code-quality.md) issue M8.

---

## Tests Recomendados (post code quality fixes)

Después de resolver los issues de calidad de código, agregar estos tests:

| Test | Qué verifica | Prioridad |
|------|-------------|-----------|
| `TestCollyExtractor_NoHTTPCall` | Verificar que Colly no haga requests HTTP (mock para confirmar) | Alta |
| `TestDomainSpecific_InvalidRegex` | Regex inválido en YAML genera warning | Alta |
| `TestCache_MaxSize` | Cache con límite de tamaño evicta entradas | Media |
| `TestTeaserDetection` | Integrar teaser.go al pipeline y verificar | Media |
| `TestHTTPRetry_BodyClosed` | Verificar que resp.Body se cierra en cada retry | Media |
| `TestConcurrentExtract` | Múltiples goroutines concurrentes con defaultCache | Media |
| `TestExtractHTML_CacheKey` | Verificar que baseURL se usa como cache key | Baja |
| `TestMarkdown_UniqueFilename` | Verificar que títulos similares no colisionan | Baja |

---

## Coverage Goal

| Paquete | Coverage Actual (est.) | Target |
|---------|----------------------|--------|
| `scraperlib` (root) | ~80% | 90% |
| `internal/fetch` | ~75% | 85% |
| `internal/extractors` | ~70% | 85% |
| `internal/cache` | ~85% | 90% |
| `internal/markdown` | ~75% | 85% |
| `internal/detection` | ~60% | 80% |
| `internal/output` | ~70% | 80% |

Para medir coverage real:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out  # abre navegador
```

---

## Unit Tests — Mock HTTP

```go
func TestFetch_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        // HTML conocido con contenido suficiente
    }))
    defer server.Close()

    fetcher := NewHTTPSimple(Options{Timeout: 5 * time.Second})
    html, err := fetcher.Fetch(context.Background(), server.URL, nil)

    assert.NoError(t, err)
    assert.Contains(t, html, "contenido esperado")
}
```

## Fixture-Based Tests

HTML real en `testdata/`:

```go
func TestReadability_Article(t *testing.T) {
    html, err := os.ReadFile("testdata/article_with_embeds.html")
    require.NoError(t, err)

    ext := NewReadabilityExtractor()
    result, err := ext.Extract(context.Background(), string(html), "https://example.com/article")

    require.NoError(t, err)
    assert.GreaterOrEqual(t, result.WordCount, 100)
}
```

### Fixtures requeridos

```
testdata/
├── article_with_embeds.html
├── github_readme.html
├── wikipedia_article.html
├── paywall_detected.html
├── cloudflare_challenge.html
├── empty_spa.html
└── price_product.html
```

## Integration Tests — URLs Reales

```go
//go:build integration

func TestExtract_RealURLs(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    tests := []struct {
        name        string
        url         string
        expectMinWC int
    }{
        {"YouTube video", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 50},
        {"GitHub README", "https://github.com/golang/go", 200},
        {"Wikipedia", "https://en.wikipedia.org/wiki/Go_(programming_language)", 500},
        {"Twitter embed", "https://example.com/article-with-tweet", 100},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := scraperlib.Extract(context.Background(), tt.url, &scraperlib.Options{
                Timeout: 30 * time.Second,
            })
            require.NoError(t, err)
            assert.GreaterOrEqual(t, result.WordCount, tt.expectMinWC)
        })
    }
}
```

## Running Tests

```bash
# Unit tests (rápidos, sin red)
go test ./...

# Con cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Solo tests cortos
go test -short ./...

# Integration tests (requiere red)
go test -tags=integration ./...

# Integration test con verbose
go test -tags=integration -v ./... -run TestExtract_RealURLs
```

---

[Volver al índice](README.md)
