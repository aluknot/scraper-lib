# Descripción General

> **Status: v10.0** — API composable con componentes públicos exportados.
> Ver [docs/implementation-status.md](implementation-status.md) para estado completo.

Servicio standalone de scraping con arquitectura extensible. Base de código extraída de Rissy.

## API en Dos Niveles

### Nivel 1: Extract() (convenience)

API de alto nivel con defaults razonables y opciones configurables:

```go
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Outputs: []string{"article", "markdown"},
    Timeout: 30 * time.Second,
    // Pipeline stages opcionales:
    NoEmbeds:           false, // preservar embeds (default)
    NoSanitize:         false, // sanitizar HTML (default)
    NoPaywallDetection:  false, // detectar paywalls (default)
    DisableCache:       false, // usar cache (default)
})
```

### Nivel 2: Componentes Públicos (composable)

Para control total, usa los componentes directamente:

```go
// 1. Fetch como quieras
html := myBrowser.GetHTML(url)

// 2. Extrae solo lo que necesitas
chain := extractors.NewChain(myCustomExtractor{})
result, err := chain.Extract(ctx, html, url)

// 3. Formatea como quieras
article := output.BuildArticleResult(result, url)
```

### Paquetes Públicos Exportados

| Paquete | Uso |
|---------|-----|
| `extractors/` | Chain, extractores, interfaz Extractor |
| `output/` | BuildArticleResult, BuildMarkdownResult |
| `types/` | ExtractResult, Attempt, StrategyAttempt |
| `embeds/` | EmbedExtractor para preserve/restore |
| `sanitize/` | Sanitizer con bluemonday |

---

## Estructura de Repositorios

```text
~/proyectos/go/
├── scraper-lib/              # Librería compartida (este repo)
├── scraper-service/          # Servicio HTTP (depende de scraper-lib)
├── rissy/                    # RSS aggregator (cliente de scraper-lib)
└── otros-proyectos/          # Finanzas, Obsidian, etc. (clientes)
```

---

## Pipeline de Extracción

```text
Fetch → [Paywall Detection] → [Embed Extract] → Extractors → [Sanitize] → [Embed Restore] → Output
         (optional)              (optional)                   (optional)     (optional)
```

---

## Estrategia de Ejecución

| Fase | Enfoque | Estado |
|------|---------|--------|
| **v10.0** | API composable + quality fixes | ✅ Completado |
| **Phase 2** | Basado en fallos reales | Pendiente |
| **Phase 3** | Alto volumen, múltiples dominios | Pendiente |

---

[Volver al índice](README.md)
