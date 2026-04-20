# Descripción General

> **Status: v10.5** — Flat metadata structure + MinWords flexible.
> Ver [docs/implementation-status.md](implementation-status.md) para estado completo.

Servicio standalone de scraping con arquitectura extensible. Base de código extraída de Rissy.

## API en Dos Niveles

### Nivel 1: Extract() (convenience)

API de alto nivel con defaults razonables y opciones configurables:

```go
// Extraer artículo completo (requiere 100+ palabras)
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    Outputs: []string{"article", "markdown"},
    Timeout: 30 * time.Second,
})

// Extraer solo metadata (MinWords=0, ultra-rápido)
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    MinWords: 0,  // Acepta cualquier resultado
    UseAdvanced: true,  // Mejor para sitios que bloquean
})

// Pipeline stages opcionales:
result, err := scraperlib.Extract(ctx, url, &scraperlib.Options{
    MinWords:           100,  // palabras mínimas (default: 100, 0=aceptar cualquier)
    NoEmbeds:           false, // preservar embeds (default)
    NoSanitize:         false, // sanitizar HTML (default)
    NoPaywallDetection: false, // detectar paywalls (default)
    DisableCache:       false, // usar cache (default)
    Debug:              false, // verbose logging
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
| `extractors/platforms/` | Platform extractors (YouTube, GitHub) |
| `output/` | BuildArticleResult, BuildMarkdownResult |
| `types/` | ExtractResult, Attempt, StrategyAttempt |
| `embeds/` | EmbedExtractor para preserve/restore |
| `sanitize/` | Sanitizer con bluemonday |

---

## Platform Extractors

Extrae metadata y contenido específico de plataformas:

```go
// Auto-detecta plataforma
extractor, err := platforms.Get(httpClient, url)

// O directo
extractor := youtube.New(client)

// Metadata (channel, views, stats)
meta, err := extractor.Metadata(ctx, url)

// Contenido (título, descripción)
content, err := extractor.Content(ctx, url)

// Perfil (info de canal/perfil)
profile, err := extractor.Profile(ctx, url)
```

### YouTube

- **URLs soportadas**: youtube.com, youtu.be
- **Tipos**: video, shorts, channel, playlist
- **Metadata**: VideoMetadata, ChannelMetadata
- **Content**: VideoContent, ShortContent

### GitHub

- **URLs soportadas**: github.com
- **Tipos**: repo, profile, release
- **Metadata**: RepoMetadata, ProfileMetadata
- **Content**: ReadmeContent

---

## Metadata Estructura (Flat)

A partir de v10.5, toda la metadata está a nivel superior (sin anidación):

```go
// ExtractResult y MetadataResult tienen todos los campos a nivel superior
result, _ := scraperlib.Extract(ctx, url, &scraperlib.Options{MinWords: 0})

// result.Metadata tiene:
result.Metadata.Title         // og:title o <title>
result.Metadata.Description  // og:description o meta description
result.Metadata.SiteName     // og:site_name
result.Metadata.ThumbnailURL  // og:image (primera imagen)
result.Metadata.Images       // todas las og:images
result.Metadata.Videos      // todos los og:videos
result.Metadata.Category    // og:type
result.Metadata.URL         // URL original
result.Metadata.Language    // idioma detectado
result.Metadata.Author      // autor
result.Metadata.WordCount  // palabras extraídas
```

### Modos de Extracción

| Modo | MinWords | Uso |
|------|----------|-----|
| Article | 100 (default) | Contenido completo |
| Metadata | 0 | Solo OG tags, ultra-rápido |

```go
// Metadata only (link-lens usa esto)
scraperlib.Extract(ctx, url, &scraperlib.Options{MinWords: 0})
```

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
