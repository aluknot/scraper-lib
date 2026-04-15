# Pipeline de Estrategias

> **Status: MVP** — `http_simple` y `http_advanced` implementados.
> Phase 3: proxy, browser, archive.

## Niveles de escalonamiento

```
[1] http_simple    → net/http estándar ✅ Implementado
     │ falla o contenido sospechoso
[2] http_advanced  → UA rotation, headers, cookies, referrer spoofing ✅ Implementado
     │ falla o bloqueado (Cloudflare, PerimeterX, etc.)
[3] http_proxy     → proxy rotation sobre http_advanced
     │ JS-rendering detectado o proxy bloqueado
[4] browser        → Playwright via browser-service (Python, contenedor separado)
     │ falla
[5] archive        → archive.org / Wayback Machine (también para paywalls)
     │ falla
Error con diagnóstico detallado
```

---

## Detección de Cuándo Escalar

> **Status: Phase 3** — Bot detection estructural completo.
> MVP usa solo word count como señal.

### Problema en v7

`strings.Contains` era muy básico — "Please enable JavaScript" en un artículo sin protección daba falsos positivos.

### Solución

- Matching por estructura DOM, no por texto
- Scoring con umbrales configurables
- Múltiples señales combinadas (AND/OR)
- **MAX de señales** (no promedio) — si detectás Cloudflare (1.0) Y JS render (0.9), el max mantiene la confianza alta

```go
type EscalationConfig struct {
    MinWordCount      int     `yaml:"min_word_count"`
    BotScoreThreshold float64 `yaml:"bot_score_threshold"`
    CheckJSRender     bool    `yaml:"check_js_render"`
    CheckCloudflare   bool    `yaml:"check_cloudflare"`
    CheckPerimeterX   bool    `yaml:"check_perimeterx"`
}

func shouldEscalate(rawHTML string) bool {
    config := getEscalationConfig()

    // Señal 1: word count bajo
    wordCount := countVisibleWords(rawHTML)
    if wordCount < config.MinWordCount {
        return true
    }

    // Señal 2: scoring de detección de bot (usa MAX, no promedio)
    score, _ := detectBotProtection(rawHTML)
    return score >= config.BotScoreThreshold
}
```

### Detección de bot por estructura

| Señal | Qué busca | Score |
|-------|-----------|-------|
| Cloudflare | `<form>` con action que contiene `/cdn-cgi/` | 1.0 |
| PerimeterX | Scripts/iframes con `px` o `datadome` | 1.0 |
| SPA vacía | `<div id="root">` vacío + scripts de React/Vue | 0.9 |

---

## Interfaz Strategy

```go
type Strategy interface {
    Name() string
    Fetch(ctx context.Context, url string, opts *Options) (string, error)
}

type StrategyAttempt struct {
    Strategy   string `json:"strategy"`
    Status     string `json:"status"`    // success, error, escalated, blocked
    DurationMs int64  `json:"duration_ms"`
    Error      string `json:"error,omitempty"`
}

type StrategyChain struct {
    strategies []Strategy
    opts       *Options
}

func (c *StrategyChain) Fetch(ctx context.Context, url string) (string, []StrategyAttempt, error) {
    var attempts []StrategyAttempt

    for _, s := range c.strategies {
        start := time.Now()
        rawHTML, err := s.Fetch(ctx, url, c.opts)
        attempt := StrategyAttempt{
            Strategy:   s.Name(),
            DurationMs: time.Since(start).Milliseconds(),
        }

        if err != nil {
            attempt.Status = "error"
            attempt.Error = err.Error()
            attempts = append(attempts, attempt)
            continue
        }

        if shouldEscalate(rawHTML) {
            attempt.Status = "escalated"
            attempts = append(attempts, attempt)
            continue
        }

        attempt.Status = "success"
        attempts = append(attempts, attempt)
        return rawHTML, attempts, nil
    }

    return "", attempts, ErrAllStrategiesFailed
}
```

---

## Circuit Breaker

> **Status: Phase 3** — NO implementar en MVP. Referencia arquitectónica.

**Problema:** Si un dominio cambia su estructura y el extractor falla consistentemente, el servicio sigue intentando en cada request.

**Solución:** Circuit breaker por dominio que se abre después de N fallos consecutivos y se cierra después de un cooldown.

```go
type State string
const (
    StateClosed   State = "closed"     // Normal operation
    StateOpen     State = "open"       // Rejecting requests
    StateHalfOpen State = "half_open"  // Testing with single request
)

type Config struct {
    FailureThreshold int           `yaml:"failure_threshold"` // Fallos para abrir
    SuccessThreshold int           `yaml:"success_threshold"` // Éxitos para cerrar
    Cooldown         time.Duration `yaml:"cooldown"`          // Tiempo en estado open
}
```

### Configuración

```yaml
# config/circuit_breakers.yaml
defaults:
  failure_threshold: 5
  success_threshold: 3
  cooldown: 5m

overrides:
  mercadopago.com:
    failure_threshold: 3
    cooldown: 10m
  github.com:
    failure_threshold: 10
    cooldown: 2m
```

---

## Browser Service

> **Status: Phase 3**

El browser headless **no va en el binario Go**. Playwright requiere Chromium y librerías de sistema.

```
scraper-service (Go)  ──── HTTP ────►  browser-service (Python)
                                            │
                                        Playwright / crawl4ai
```

---

[Volver al índice](README.md)
