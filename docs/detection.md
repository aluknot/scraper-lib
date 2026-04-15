# Detección: qualityScore y Paywall

## Paywall Detection

> **Status: Implementado.** Solo marca, no evade.

```go
// internal/detection/paywall.go
func DetectPaywall(rawHTML string) (bool, []string) {
    // Busca señales en el HTML:
    // 1. Clases CSS conocidas: paywall, subscription-wall, premium-content, etc.
    // 2. IDs de overlay/modal: paywall-cta, subscription-prompt, login-wall, etc.
    // 3. Texto en párrafos: "subscribe to continue", "premium content", etc.
}
```

### Uso en el pipeline

```go
if isPaywall, signals := detection.DetectPaywall(rawHTML); isPaywall {
    for _, s := range signals {
        warnings = append(warnings, fmt.Sprintf("paywall_detected:%s", s))
    }
}
```

---

## qualityScore()

> **Status: Phase 2+** — MVP usa validación simple (`WordCount >= 100`).
> Se mantiene implementación completa como referencia.

Calcula un score de 0.0 a 1.0 basado en múltiples factores:

| Factor | Peso | Lógica |
|--------|------|--------|
| Word count | 0.3 | Ideal: 500+ palabras, mínimo: 100 |
| Densidad de enlaces | 0.15 | < 0.05 = buena, 0.05-0.15 = aceptable, > 0.15 = demasiados |
| Presencia de imágenes | 0.1 | 1 o más = score completo |
| Metadata completa | 0.15 | title (0.4), author (0.3), publishedAt (0.3) |
| Extractor usado | 0.3 | domain_specific=1.0, readability=0.8, trafilatura=0.6, colly=0.3 |

```go
func qualityScore(result *types.ExtractResult) float64 {
    var score float64
    var weights float64

    // Factor 1: Word count (peso: 0.3)
    wc := float64(result.WordCount)
    wcScore := math.Min(1.0, wc/500.0)
    score += wcScore * 0.3
    weights += 0.3

    // Factor 2: Densidad de enlaces (peso: 0.15)
    linkRatio := float64(len(result.Links)) / wc
    if linkRatio < 0.05 {
        score += 1.0 * 0.15
    } else if linkRatio < 0.15 {
        score += 0.7 * 0.15
    } else {
        score += 0.2 * 0.15
    }
    weights += 0.15

    // Factor 3: Presencia de imágenes (peso: 0.1)
    if len(result.Images) > 0 {
        score += 1.0 * 0.1
    }
    weights += 0.1

    // Factor 4: Metadata completa (peso: 0.15)
    metadataScore := 0.0
    if result.Title != "" { metadataScore += 0.4 }
    if result.Author != "" { metadataScore += 0.3 }
    if result.PublishedAt != nil { metadataScore += 0.3 }
    score += metadataScore * 0.15
    weights += 0.15

    // Factor 5: Extractor usado (peso: 0.3)
    extractorScores := map[string]float64{
        "domain_specific": 1.0, "readability": 0.8,
        "trafilatura": 0.6, "colly": 0.3,
    }
    if extScore, ok := extractorScores[result.ExtractorUsed]; ok {
        score += extScore * 0.3
    }
    weights += 0.3

    return score / weights
}
```

---

[Volver al índice](README.md)
