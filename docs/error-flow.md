# Diagrama de Flujo de Errores

## Pipeline completo

```text
Extract(url)
    │
    ├──[0] Check Cache (si existe)
    │     │
    │     ├── hit → return cached result ✅
    │     └── miss → continue
    │
    ├──[1] fetchHTML(ctx, url)
    │     │
    │     ├── success → rawHTML
    │     │
    │     └── error (network, timeout, DNS)
    │           │
    │           ├── Retry? (exponential backoff, max 3)
    │           │     ├── success → continue
    │           │     └── fail → return ErrFetchFailed{URL, Attempts[], LastError}
    │           │
    │           └── Circuit breaker open? → ErrCircuitOpen{Domain, CooldownRemaining}
    │
    ├──[2] detectPaywall(rawHTML)
    │     │
    │     ├── signals detected → add warnings to result
    │     └── no signals → continue
    │
    ├──[3] extractAndReplaceEmbeds(rawHTML)
    │     │
    │     ├── embeds found → placeholders + embedMap
    │     └── no embeds → processedHTML = rawHTML, embedMap = nil
    │
    ├──[4] StrategyChain.Fetch(processedHTML)
    │     │
    │     ├── http_simple → success → rawHTML
    │     ├── escalated? → try http_advanced
    │     ├── escalated? → try http_proxy (post-MVP)
    │     ├── escalated? → try browser (post-MVP)
    │     ├── escalated? → try archive.org (post-MVP)
    │     │
    │     └── ALL strategies failed
    │           │
    │           └── return ErrAllStrategiesFailed{
    │                 URL,
    │                 Attempts[] (cada strategy + status + error),
    │                 BotDetectionSignals[] (si aplica)
    │               }
    │
    ├──[5] chain.Extract(html, url)
    │     │
    │     ├── domain_specific → match → extracted
    │     ├── readability → success + wordCount >= 100 → extracted
    │     ├── trafilatura → success + wordCount >= 100 → extracted
    │     ├── fallback → extracted
    │     │
    │     └── ALL extractors failed (error o low_quality)
    │           │
    │           └── return ErrAllExtractorsFailed{
    │                 URL,
    │                 Attempts[] (cada extractor + status + duration + error),
    │                 QualityScore (si algún extractor retornó algo)
    │               }
    │
    ├──[6] sanitizeHTML(extracted.Content)
    │     │
    │     ├── normal → cleaned content
    │     └── warning (e.g. suspicious content removed) → add warning, continue
    │
    ├──[7] restoreEmbeds(content, embedMap)
    │     │
    │     ├── embedMap not empty → restore iframes/blockquotes
    │     └── embedMap empty → content unchanged
    │
    ├──[8] qualityScore(result)
    │     │
    │     ├── score >= 0.7 → return result
    │     ├── score 0.4-0.7 → return result + warning("low_quality")
    │     └── score < 0.4 → return result + warning("very_low_quality")
    │
    └──[9] Return Result{
              Article/Metadata/Price/Job/Markdown (según Outputs),
              ExtractorAttempts[],
              StrategyAttempts[],
              QualityScore,
              Warnings[] (paywall, low_quality, sanitize, etc.),
              DurationMs
          }
```

**Pasos implementados:** [0] cache check (InMemory + File), [1] fetch con retry + http_advanced, [2] paywall detection, [3] embed extraction, [5] extractor chain (domain_specific → readability → trafilatura → fallback), [6] sanitization, [7] embed restoration, [8] error types, [9] result building + cache store + markdown rendering.

**Pendientes:** [4] estrategia completa de escalación (proxy, browser, archive), [8] qualityScore completo.

---

## Debugging

### Debug Logs

Habilitar logging detallado con nivel DEBUG:

```go
// Minimal logging (default)
result, err := scraperlib.Extract(ctx, url, nil)

// Verbose logging for troubleshooting
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Debug: true,
})
```

### Log Keys

| Log | Descripción | Campos |
|-----|--------------|--------|
| `fetch_start` | Inicio del fetch | url, strategy |
| `fetch_success` | Fetch completado | url, html_size, duration_ms |
| `fetch_failed` | Fetch falló | url, error |
| `fetch_empty_html` | HTML vacío recibido | url |
| `fetch_html_preview` | Primeros 200 chars del HTML | url, html_preview |
| `extractor_error` | Extractor retornó error | url, extractor, error, duration_ms |
| `extractor_low_quality` | word_count < 100 | url, extractor, word_count, content_length |
| `extractor_success` | Extracción exitosa | url, extractor, word_count, content_length |
| `all_extractors_failed` | Todos los extractores fallaron | url, attempts, attempt_summary |

### Troubleshooting común

**"all extractors failed"** significa que ningún extractor pudo extraer ≥100 words. Causas posibles:

1. **HTML vacío o muy corto** — verificar con `fetch_empty_html` log
2. **Servidor bloquea requests** — intentar con `UseAdvanced: true`
3. **Protección (Cloudflare, CAPTCHA)** — verificar manualmente el HTML
4. **SPA/JavaScript rendering** — requires Playwright (Phase 3)

### MinWords - Umbral Flexible

Para casos donde solo necesitás metadata (como link-lens), podés ajustar el umbral:

```go
// Extraer lo que haya, sin mínimo de palabras
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    MinWords: 0,  // Acepta cualquier resultado
})

//默认值
result, err := scraperlib.Extract(ctx, url, nil) // MinWords: 100 (default)
```

**Cuándo usar MinWords=0:**
- Solo necesitás metadata (título, descripción, OG tags)
- link-lens: guardar URL con info básica
- Preview rápido sin validar contenido

**MetadataExtractor:**
Cuando MinWords < 100, el chain incluye automáticamente `MetadataExtractor` (priority 0) que extrae solo og:title, og:description, author, etc. — ultra-rápido sin usar readability.

---

## Tipos de Error

### Implementados

```go
// ErrFetchFailed — no se pudo obtener el HTML
type ErrFetchFailed struct {
    URL        string
    Attempts   []StrategyAttempt
    LastError  error
}

// ErrAllStrategiesFailed — todas las strategies fallaron
type ErrAllStrategiesFailed struct {
    URL      string
    Attempts []StrategyAttempt
}

// ErrCircuitOpen — circuit breaker abierto para el dominio
type ErrCircuitOpen struct {
    Domain           string
    CooldownRemaining time.Duration
}
```

---

[Volver al índice](README.md)
