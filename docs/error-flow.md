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
    │     ├── colly → fallback → extracted
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

**Pasos implementados:** [0] cache check (InMemory + File), [1] fetch con retry + http_advanced, [2] paywall detection, [3] embed extraction, [5] extractor chain (domain_specific → readability → trafilatura → colly), [6] sanitization, [7] embed restoration, [8] error types, [9] result building + cache store + markdown rendering.

**Pendientes:** [4] estrategia completa de escalación (proxy, browser, archive), [8] qualityScore completo.

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
