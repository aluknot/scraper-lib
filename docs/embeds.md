# Embeds: Pre/Post Processor

> **Status: Implementado.** No modificar. El orden es crítico.

**Este es el punto más importante de la arquitectura defensiva.**

Extractores heurísticos como Readability o Trafilatura eliminan `<iframe>`,
`<script>` y `<blockquote class="twitter-tweet">` porque los tratan como
widgets o anuncios. La solución es actuar *antes* de pasarle el HTML al
extractor, no después.

Los embeds **no son un extractor en el chain**. Son un pre/post procesador
que envuelve al chain completo.

## Por qué el orden de sanitización es crítico

```text
fetchHTML(url)
      │
      ▼
extractAndReplaceEmbeds(html)       ← PRE: reemplaza embeds con placeholders
      │
      ▼
chain.Extract(processedHTML, url)   ← readability / trafilatura / fallback
      │  Los placeholders sobreviven porque son texto plano.
      │
      ▼
sanitizeHTML(result.Content)        ← bluemonday ANTES de restore (crítico)
      │  Si sanitizás DESPUÉS de restore, bluemonday elimina los iframes
      │  recién restaurados. El orden no es negociable.
      │
      ▼
restoreEmbeds(result.Content, embedMap)
      │  Restaura el outerHTML original en su posición exacta.
      │
      ▼
Result con embeds en posición original
```

## Implementación

```go
// internal/fetch/embeds.go

// knownPlatforms lists URL patterns that indicate an iframe should be preserved.
var knownPlatforms = []string{
    "youtube.com",
    "youtu.be",
    "vimeo.com",
    "player.vimeo.com",
    "spotify.com",
    "open.spotify.com",
    "soundcloud.com",
    "w.soundcloud.com",
    "codepen.io",
    "instagram.com/embed",
    "tiktok.com/embed",
    "www.tiktok.com/embed",
    "twitter.com/embed",
    "x.com/embed",
    "platform.twitter.com",
}

// isSupportedEmbed determina si un nodo HTML es un embed conocido que debe
// preservarse. La estrategia es conservadora (lista explícita de plataformas)
// para evitar preservar ads o widgets no deseados.
func isSupportedEmbed(n *html.Node) bool {
    // Caso 1: Iframes de plataformas conocidas
    if n.Type == html.ElementNode && n.Data == "iframe" {
        src := getAttr(n, "src")
        for _, platform := range knownPlatforms {
            if strings.Contains(src, platform) {
                return true
            }
        }
        return false
    }

    // Caso 2: Blockquote-based embeds (Twitter/X, Instagram, TikTok)
    if n.Type == html.ElementNode && n.Data == "blockquote" {
        class := getAttr(n, "class")
        return strings.Contains(class, "twitter-tweet") ||
            strings.Contains(class, "instagram-media") ||
            strings.Contains(class, "tiktok-preserve")
    }

    return false
}

// ExtractAndReplace extrae embeds conocidos y los reemplaza con placeholders
// únicos. Usa UUID en el placeholder para prevenir colisiones.
// Guarda el outerHTML completo (no solo src) para preservar todos los atributos.
func (e *EmbedExtractor) ExtractAndReplace(rawHTML string) (processedHTML string, embedMap map[string]string) {
    doc, err := html.Parse(strings.NewReader(rawHTML))
    if err != nil {
        return rawHTML, nil
    }

    embedMap = map[string]string{}
    runID := uuid.New().String()
    counter := 0

    var walk func(*html.Node)
    walk = func(n *html.Node) {
        if isSupportedEmbed(n) {
            var buf strings.Builder
            html.Render(&buf, n)
            outerHTML := buf.String()

            placeholder := fmt.Sprintf("[[EMBED_%s_%d]]", runID, counter)
            embedMap[placeholder] = outerHTML

            textNode := &html.Node{Type: html.TextNode, Data: placeholder}
            n.Parent.InsertBefore(textNode, n)
            n.Parent.RemoveChild(n)
            counter++
        } else {
            for c := n.FirstChild; c != nil; {
                next := c.NextSibling
                walk(c)
                c = next
            }
        }
    }
    walk(doc)

    var buf strings.Builder
    html.Render(&buf, doc)
    return buf.String(), embedMap
}

// Restore reemplaza los placeholders con el HTML original del embed,
// envuelto en <figure class="embedded-content">.
// Llamar DESPUÉS de sanitizeHTML.
func (e *EmbedExtractor) Restore(extractedHTML string, embedMap map[string]string) string {
    result := extractedHTML
    for placeholder, originalHTML := range embedMap {
        wrapper := fmt.Sprintf(`<figure class="embedded-content">%s</figure>`, originalHTML)
        result = strings.Replace(result, placeholder, wrapper, 1)
    }
    return result
}
```

## Uso en scraper.go

```go
func Extract(ctx context.Context, url string, opts *Options) (*Result, error) {
    rawHTML, err := fetch.GetHTML(ctx, url, opts)
    if err != nil {
        return nil, err
    }

    // 1. PRE: extraer embeds y reemplazar con placeholders únicos
    processedHTML, embedMap := embeds.ExtractAndReplace(rawHTML)

    // 2. Extracción con chain
    extracted, err := defaultChain.Extract(ctx, processedHTML, url)
    if err != nil {
        return nil, err
    }

    // 3. Sanitizar ANTES de restaurar
    extracted.Content = sanitize.Clean(extracted.Content)

    // 4. POST: restaurar embeds en su posición original
    extracted.Content = embeds.Restore(extracted.Content, embedMap)

    return buildOutput(extracted, opts.Output), nil
}
```

## Tests

**7 tests pasando** en `internal/fetch/embeds_test.go`:

| Test | Qué verifica |
|------|-------------|
| `TestExtractAndReplace_SingleYouTube` | Placeholder generado, outerHTML con atributos preservados |
| `TestExtractAndReplace_ConsecutiveEmbeds` | 2 placeholders, mismo runID, NextSibling preservado |
| `TestExtractAndReplace_PlaceholderTextInPage` | Texto literal `[[EMBED_...]]` no colisiona |
| `TestSanitizeBeforeRestore_IframesPreserved` | bluemonday no elimina iframes restaurados |
| `TestIsSupportedEmbed_TwitterVariants` | Detección de clases variants (twitter-tweet, aligned, instagram, tiktok) |
| `TestExtractAndRestore_Vimeo` | Vimeo iframe extraído correctamente |
| `TestRestore_WrapsInFigure` | Wrapper `<figure class="embedded-content">` presente |

---

[Volver al índice](README.md)
