# Arquitectura de scraper-lib

## Extractor Chain: DI Hybrid

> **Status: Implementado.**

Registro explícito con inyección de dependencias. Más fácil de testear que auto-registro con `init()`.

```go
// internal/extractors/chain.go
type Chain struct {
    extractors []Extractor
}

func NewChain(extractors ...Extractor) *Chain {
    sort.Slice(extractors, func(i, j int) bool {
        return extractors[i].Priority() < extractors[j].Priority()
    })
    return &Chain{extractors: extractors}
}

func (c *Chain) Extract(ctx context.Context, html, url string) (*types.ExtractResult, error) {
    var attempts []types.Attempt

    for _, e := range c.extractors {
        start := time.Now()
        result, err := e.Extract(ctx, html, url)
        attempt := types.Attempt{
            Extractor:  e.Name(),
            DurationMs: time.Since(start).Milliseconds(),
        }

        if err != nil {
            attempt.Status = "error"
            attempt.Error = err.Error()
            attempts = append(attempts, attempt)
            continue
        }

        if !result.IsValid() {
            attempt.Status = "low_quality"
            attempts = append(attempts, attempt)
            continue
        }

        attempt.Status = "success"
        attempts = append(attempts, attempt)
        result.Attempts = attempts
        return result, nil
    }

    return &types.ExtractResult{Attempts: attempts}, ErrAllExtractorsFailed{URL: url, Attempts: attempts}
}

// DefaultChain crea un chain con los extractores estándar.
// domain-specific va primero; si no hay regla para el dominio, retorna
// ErrNoRuleForDomain y el chain continúa normalmente con readability.
func DefaultChain() *Chain {
    return NewChain(
        NewDomainSpecificExtractor(), // priority 0 — sabe cuándo no aplica
        NewReadabilityExtractor(),    // priority 1
        NewTrafilaturaExtractor(),    // priority 2
        NewCollyExtractor(),          // priority 3 — último recurso
    )
}
```

---

## Interfaz Extractor

> **Status: Implementado.**

```go
// internal/extractors/interface.go
type Extractor interface {
    Name() string
    Priority() int
    Extract(ctx context.Context, html string, url string) (*types.ExtractResult, error)
}
```

---

## Estructura del paquete

```text
scraper-lib/
├── scraper.go                      # API pública: Extract(), ExtractHTML()
├── scraper_test.go                 # Tests de integración (7 tests)
│
├── config/
│   └── extractors.yaml             # Reglas de domain-specific (dinámico)
│
├── internal/
│   ├── extractors/
│   │   ├── interface.go            # Extractor interface
│   │   ├── chain.go                # Chain con attempts y diagnóstico
│   │   ├── domain_specific.go      # Extractor por dominio (carga desde YAML)
│   │   ├── readability.go          # go-readability v2
│   │   ├── trafilatura.go          # go-trafilatura
│   │   └── colly.go                # colly último recurso
│   │
│   ├── fetch/
│   │   ├── http.go                 # fetchWithRetry, exponential backoff
│   │   ├── embeds.go               # ExtractAndReplace() + Restore()
│   │   └── sanitize.go             # bluemonday wrapper
│   │
│   ├── output/
│   │   └── article.go              # BuildArticleResult, BuildMetadataResult, BuildRawResult
│   │
│   ├── detection/
│   │   ├── teaser.go               # isLikelyTeaser(), isKnownTeaserDomain()
│   │   └── paywall.go              # detectPaywall() — marca, no evade
│   │
│   └── types/
│       └── types.go                # ExtractResult, Attempt, PriceInfo, JobInfo, StrategyAttempt
```

---

## API Pública

> **Status: Implementado.**

```go
// scraper.go
package scraperlib

type Options struct {
    Timeout       time.Duration
    UserAgent     string
    ExtractImages bool
    Outputs       []string // Preferred: permite múltiples ["article", "metadata", "price"]
    Output        string   // DEPRECATED: usar Outputs. Mantenido para backwards compat.

    // Control del extractor chain — si no se setean, usa DefaultChain().

    Extractors []extractors.Extractor  // chain custom (reemplaza DefaultChain)
    Extractor  string                  // fuerza un extractor por nombre
    NoFallback bool                    // si el primero falla, no intenta el siguiente
}
```

### Control del Extractor Chain

**Por defecto**, `Extract()` usa `DefaultChain()` (domain_specific → readability → trafilatura → colly).

**Opción A: Forzar un solo extractor por nombre**

```go
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Extractor: "readability",
})
```

Nombres válidos: `"domain_specific"`, `"readability"`, `"trafilatura"`, `"colly"`.

**Opción B: Chain personalizado**

```go
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Extractors: []extractors.Extractor{
        extractors.NewDomainSpecificExtractor(),
        extractors.NewReadabilityExtractor(),
        // Sin trafilatura ni colly
    },
})
```

**Opción C: Deshabilitar fallback**

Si el extractor elegido falla, retorna error en vez de intentar el siguiente:

```go
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Extractor:  "readability",
    NoFallback: true, // Si readability falla, retorna error
})
```

**Prioridad:** `Extractor` (por nombre) > `Extractors` (lista custom) > `DefaultChain()`.

**Nombres desconocidos:** Si se pasa un nombre que no existe, se hace fallback a `DefaultChain()` en vez de fallar silenciosamente.

```go
// Extract descarga y extrae contenido de una URL.
func Extract(ctx context.Context, url string, opts *Options) (*Result, error)

// ExtractHTML extrae contenido de HTML ya descargado.
// Útil para tests y cuando el caller ya tiene el HTML.
func ExtractHTML(ctx context.Context, html string, baseURL string, opts *Options) (*Result, error)
```

### Normalize

```go
func (o *Options) Normalize() {
    if len(o.Outputs) == 0 && o.Output != "" {
        o.Outputs = []string{o.Output}
    }
    if len(o.Outputs) == 0 {
        o.Outputs = []string{"article"} // default
    }
}
```

---

## Cache

> **Status: Implementado.**

La librería cachea resultados de extracción para evitar fetches redundantes.

### Interfaz

```go
type Cache interface {
    Get(url string) (*Result, bool)
    Set(url string, result *Result, ttl time.Duration)
    Delete(url string) error
    Clear() error
    Stats() Stats  // {Hits, Misses, Size}
}
```

### Implementaciones incluidas

**InMemoryCache** (default):
- Compartido entre todas las llamadas a `Extract()` que no especifican cache custom
- Thread-safe con `sync.RWMutex`
- TTL-based expiration

```go
result, _ := scraperlib.Extract(ctx, url, nil)  // usa InMemoryCache compartido
```

**FileCache**:
- Guarda resultados como JSON en un directorio
- Persiste entre reinicios del proceso
- Ideal para herramientas CLI standalone

```go
fileCache, _ := cache.NewFileCache("~/.cache/myapp")
result, _ := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Cache: fileCache,
})
```

### Opciones de cache en Options

```go
type Options struct {
    // ... otros campos

    Cache    cache.Cache     // nil = InMemoryCache compartido
    CacheTTL time.Duration   // default: 24 horas
}
```

---

## Output Markdown

> **Status: Implementado.**

Genera Markdown con YAML frontmatter, configurable vía templates YAML.

### Interfaz

```go
type MarkdownResult struct {
    Content  string   // frontmatter YAML + cuerpo Markdown
    Filename string   // nombre de archivo sanitizado
    Tags     []string // auto-generados del título e idioma
}
```

### Templates YAML

Las templates definen qué campos van en el frontmatter y cómo se estructura el cuerpo:

```yaml
# templates/article.yaml (default built-in)
frontmatter:
  - title
  - url
  - source
  - author
  - published_at
  - extracted_at
  - extractor
  - language
  - word_count
  - tags

body:
  heading: "{{title}}"
  content: "{{content}}"
  footer: "Extracted on {{extracted_at}} via {{extractor}}"
```

### Template resolution

La template se selecciona automáticamente según el extractor usado y la categoría del contenido:

```
extractor + category  →  "youtube_tutorial"  →  youtube_tutorial.yaml
extractor             →  "youtube"           →  youtube.yaml
fallback              →  "article"           →  article.yaml  (built-in)
```

### Uso

```go
// Default (template article.yaml built-in)
result, _ := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Outputs: []string{"markdown"},
})

// Custom template directory
result, _ := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Outputs:     []string{"markdown"},
    TemplateDir: "~/.config/scraper-lib/templates",
})

// Override de categoría
result, _ := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Outputs:  []string{"markdown"},
    Category: "tutorial",
})
```

### HTML → Markdown

El contenido HTML extraído se convierte automáticamente a Markdown:
- Headings (`<h1>` → `#`, `<h2>` → `##`, etc.)
- Bold (`<strong>` → `**`)
- Italic (`<em>` → `*`)
- Links (`<a href>` → `[text](url)`)
- Code (`<code>` → `` ` ``)
- Blockquotes (`<blockquote>` → `>`)

### Generación de filename

```go
markdown.GenerateFilename("My Great Article", "https://example.com/path")
// → "example.com/my-great-article.md"
```

---

[Volver al índice](README.md)
