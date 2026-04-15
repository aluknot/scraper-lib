# Code Quality Review — Hallazgos y Recomendaciones

> **Fecha:** 2026-04-15
> **Versión:** v10.0
> **Alcance:** Revisión completa de los archivos Go del proyecto
> **Metodología:** Five-Axis Review (correctness, readability, architecture, security, performance)

---

## Prioridad de Resolución

Los issues están organizados por severidad. El orden recomendado de resolución es:

1. **Críticos** → Corregir antes de cualquier nueva feature
2. **Altos** → Corregir antes del próximo release
3. **Medios** → Planificar en Phase 2 o sprint de calidad
4. **Bajos** → Opportunistic, cuando se toque código relacionado

---

## 🔴 Críticos (Correctness)

### C1: CollyExtractor hace HTTP request redundante ✅

**Estado:** RESUELTO en v10.0

El extractor ahora usa `goquery.NewDocumentFromReader()` para parsear el HTML en memoria sin requests adicionales.

---

### C2: `defer resp.Body.Close()` dentro de retry loop ✅

**Estado:** RESUELTO en v9.5

El `defer` ahora está inmediatamente después de confirmar éxito, antes de `io.ReadAll`.
    // Aplicar mismos selectores: article, [class*='article'], main, div[class*='content']
    // ...
}
```

---

### C2: `defer resp.Body.Close()` dentro de retry loop

**Archivos:** `internal/fetch/http.go:85`

**Problema:** El `defer resp.Body.Close()` está dentro del bloque `if resp.StatusCode >= 200 && resp.StatusCode < 300`. Si en el futuro se agrega un path de retorno temprano después de este punto, el body se cerraría correctamente por el defer. Sin embargo, la posición actual es inusual y podría confundir.

En el código actual no hay leak porque los paths de error del retry loop cierran el body explícitamente antes del `continue`. Pero es mejor práctica mover el defer inmediatamente después de confirmar éxito:

```go
// Antes (dentro del if de éxito)
if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    defer resp.Body.Close()  // ← posición actual
    // ...
}

// Después (inmediatamente después de confirmar éxito)
if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    // OK — cerrar body al retornar
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    // ...
}
```

---

## 🟠 Altos (Code Quality)

### H1: `getAttr()` duplicada en 2 archivos ✅

**Estado:** RESUELTO en v10.0

Extraída a `internal/dom/attr.go` y usada en `embeds/` y `detection/paywall.go`.

---

### H2: Domain extraction duplicada 3 veces ✅

**Estado:** RESUELTO en v9.5

Consolidada en `internal/urlutil/domain.go` y `urlutil/path.go`. Usada en `extractors/domain_specific.go` y `output/article.go`.

---

### H3: Dos implementaciones de HTML→Markdown ✅

**Estado:** RESUELTO en v10.0

`domain_specific.go` ahora usa `markdown.HTMLToMarkdown()` en vez de `htmlToMarkdown()` local.

**Fix recomendado:** Unificar a una sola implementación. La versión regex-based en `html2md.go` es más simple y ya es usada por el output pipeline. Reemplazar `htmlToMarkdown()` en `domain_specific.go` para que use `markdown.HTMLToMarkdown()`.

---

### H4: `isValidSelector()` no-op ✅

**Estado:** RESUELTO en v10.0

Función eliminada. Goquery no retorna error para selectores inválidos.

---

### H5: Regex inválidos silenciados en YAML ✅

**Estado:** RESUELTO en v10.0

`loadConfig()` ahora loguea a stderr cuando un regex falla.

---

## 🟡 Medios (Maintainability)

### M1: Default cache sin límite de tamaño ✅

**Estado:** RESUELTO en v10.1

- `NewInMemoryCache()` ahora tiene límite de 10000 entradas
- `NewInMemoryCacheWithLimit(maxEntries)` permite configurar límite custom
- Política LRU para evict cuando se alcanza el límite
- Tests para LRU eviction incluidos

---

### M2: Funciones teaser.go preparadas para Phase 2

**Archivos:** `internal/detection/teaser.go`

**Funciones sin caller:** `IsLikelyTeaser()`, `IsKnownTeaserDomain()`, `NeedsScraping()`

**Decision:** Mantener — preparado para Phase 2. Funciones útiles para:
- `IsLikelyTeaser()` — detectar contenido truncado
- `IsKnownTeaserDomain()` — dominios soft-paywall
- `NeedsScraping()` — decidir si hacer scraping completo

---

### M3: ErrCircuitOpen preparado para Phase 3

**Archivos:** `errors.go`

**Problema:** Definido pero nunca retornado (circuit breaker es Phase 3).

**Decision:** Mantener — preparado para Phase 3 cuando se implemente circuit breaker por dominio.

---

### M4: PriceInfo, JobInfo preparados para Phase 2

**Archivos:** `internal/types/types.go`

**Problema:** Tipos definidos pero ningún extractor los pobla (Phase 2+).

**Decision:** Mantener — preparado para cuando se necesiten extractores especializados:
- `PriceInfo` → e-commerce (Amazon, MercadoLibre)
- `JobInfo` → job postings (LinkedIn, Indeed)

---

### M5: `var _ = embed.FS{}` dead code ✅

**Estado:** RESUELTO en v10.0

Línea eliminada.

---

### M6: `contains()` manual vs `strings.Contains` ✅

**Estado:** RESUELTO en v10.0

Usando `strings.Contains()` directamente.

---

### M7: `normalizeWhitespace()` O(n²) ✅

**Estado:** RESUELTO en v10.0

Usando regex `regexp.MustCompile(`\n{3,}`)` para normalizar whitespace en O(n).

---

### M8: `testdata/` vacío ✅

**Estado:** RESUELTO en v10.1

Fixtures creados:
- `testdata/article_with_embeds.html` - Artículo con YouTube y Twitter embeds
- `testdata/github_readme.html` - README de GitHub con estructura típica
- `testdata/wikipedia_article.html` - Artículo Wikipedia con referencias
- `testdata/paywall_detected.html` - Página con señales de paywall
- `testdata/cloudflare_challenge.html` - Cloudflare challenge page
- `testdata/empty_spa.html` - SPA vacía (JS-rendered)
- `testdata/price_product.html` - Página de producto e-commerce

---

### M9: Errores ignorados ✅

**Estado:** RESUELTO en v10.0

`trafilatura.go` ahora retorna error de `nodeToHTML()`. `template.go loadDir()` loguea a stderr.

---

## 🟢 Bajos (Nice-to-have)

### L1: `MemoryCache.Get()` usa Lock vs RLock

**Archivos:** `internal/cache/memory.go`

**Problema:** Usa `mu.Lock()` (write lock) para una operación de lectura. Correcto cuando necesita borrar entry expirada, pero innecesario para hits puros.

**Fix:** Usar `mu.RLock()` primero, verificar existencia y expiración, solo adquirir `mu.Lock()` si necesita delete.

---

### L2: Cero logging en todo el pipeline ✅

**Estado:** RESUELTO en v10.1

Usando `log/slog` (built-in en Go 1.21+):

```go
// Logs en pipeline
slog.Debug("fetch_start", "url", url, "strategy", "simple")
slog.Info("cache_hit", "url", url, "duration_ms", ms)
slog.Warn("paywall_detected", "url", url, "signals", signals)
slog.Info("extraction_complete", "url", url, "extractor", name, "word_count", wc)
slog.Error("fetch_failed", "url", url, "error", err)
```

---

### L3: Retry config hardcoded

**Archivos:** `internal/fetch/http.go:10-12`

```go
const (
    MaxRetries     = 3
    baseRetryDelay = 1 * time.Second
    maxRetryDelay  = 10 * time.Second
)
```

**Fix:** Exponer en `Options.FetchOptions` para callers que necesiten comportamiento distinto.

---

### L4: `github.com` en teaser domains

**Archivos:** `internal/detection/teaser.go`

**Problema:** GitHub es principalmente un host de código, no un site que sirva teaser/resumen de contenido. La inclusión parece cuestionable.

**Fix:** Revisar si tiene sentido concreto. Si no, remover de `knownTeaserDomains`.

---

### L5: `GenerateFilename()` truncación a 80 chars

**Archivos:** `internal/markdown/template.go`

**Problema:** Truncar a 80 caracteres puede producir filenames no únicos para títulos similares:
- "Understanding Go Concurrency Patterns in Modern..." → same prefix as
- "Understanding Go Concurrency Patterns for Web..."

**Fix:** Agregar sufijo hash corto para unicidad:

```go
func GenerateFilename(title, url string) string {
    slug := sanitizeTitle(title)
    if len(slug) > 80 {
        slug = slug[:70]
        hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))[:8]
        slug = slug + "-" + hash
    }
    return slug + ".md"
}
```

---

## Recomendaciones Generales de Organización

### 1. Agregar golangci-lint

Configurar `.golangci.yml` con linters básicos:

```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - misspell
    - gofmt
```

### 2. Agregar CI con quality gates

Agregar al pipeline de CI (GitHub Actions o similar):
- `go test ./...` — todos los tests
- `go vet ./...` — análisis estático
- `golangci-lint run` — linting
- `go build ./...` — compilación limpia

### 3. Documentar valores por defecto

Crear `docs/defaults.md` con todos los valores hardcoded y su justificación:
- `DefaultCacheTTL = 24h`
- `Timeout = 30s`
- `MaxRetries = 3`
- `WordCount >= 100` para validez
- UA pool (8 agentes)
- Referrer pool (8 referers)

### 4. Agregar comentarios de fase planificada

Para código Phase 2/3 que existe pero no se usa aún, agregar comentarios consistentes:

```go
// Phase 3: Circuit breaker — activar cuando haya volumen suficiente
// por dominio con fallback a estrategia simple
var ErrCircuitOpen = errors.New("circuit breaker open")
```

---

## Orden Recomendado de Implementación

| # | Acción | Prioridad | Esfuerzo |
|---|--------|-----------|----------|
| 1 | Fix Colly redundant HTTP (C1) | Crítico | Bajo |
| 2 | Fix defer body close (C2) | Crítico | Mínimo |
| 3 | Deduplicar `getAttr()` (H1) | Alto | Mínimo |
| 4 | Deduplicar domain extraction (H2) | Alto | Bajo |
| 5 | Consolidar HTML→Markdown (H3) | Alto | Medio |
| 6 | Fix `isValidSelector()` (H4) | Alto | Mínimo |
| 7 | Log regex inválidos (H5) | Alto | Mínimo |
| 8 | Eliminar dead code (M2-M5) | Medio | Mínimo |
| 9 | Fix `contains()` (M6) | Medio | Mínimo |
| 10 | Fix `normalizeWhitespace()` (M7) | Medio | Mínimo |
| 11 | Crear testdata fixtures (M8) | Medio | Bajo |
| 12 | Agregar structured logging (L2) | Medio | Medio |
| 13 | Agregar golangci-lint | General | Bajo |
| 14 | Cache size limits (M1) | Medio | Bajo |

---

[Volver al índice](README.md)
