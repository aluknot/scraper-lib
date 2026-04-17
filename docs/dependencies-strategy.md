# Estrategia de Dependencias

> **Fecha:** 2026-04-15
> **Versión:** v10.0

## Estado Actual

**Dependencias directas (8):**
- `go-readability/v2` - extracción de artículos
- `goquery` - CSS selectors para parsear HTML
- `colly/v2` - (legacy, ya no se usa para HTTP)
- `google/uuid` - placeholders para embeds
- `go-trafilatura` - extractor alternativo
- `bluemonday` - sanitización HTML
- `golang.org/x/net` - parsing HTML/DOM
- `yaml.v3` - parsing configs

**Dependencias transitivas:** 35+

---

## Análisis de Tradeoffs

### Pros de múltiples dependencias
- Specialización: cada lib hace una cosa bien
- Mantenimiento externo
- Testeado en producción (libs maduras)
- Ahorro de tiempo

### Contras
- Riesgo de vulnerabilidades (supply chain)
- Tamaño del binario
- API changes pueden romper
- Complexity de updates

---

## Plan de Reducción Futuro

### Fase 1: Evaluar uso real (post-MVP)

Agregar métricas/telemetry para saber:
- ¿Cuántas veces se usa cada extractor?
- ¿Trafilatura aporta valor vs readability?
- ¿Colly realmente extra algo que readability no?

### Fase 2: Decisiones basadas en datos

| Si... | Entonces... |
|-------|------------|
| Trafilatura nunca gana | Eliminar `go-trafilatura` y sus ~30 transitive deps |
| Readability gana siempre | Solo mantener goquery + bluemonday |
| Colly no se usa | Eliminar `gocolly/colly/v2` |

### Fase 3: Reducción agresiva (opcional)

**Opción A: Mantener solo lo esencial**
```
✓ go-readability/v2  - extractor principal
✓ bluemonday        - sanitización
✓ goquery           - domain rules + CSS selectors
✗ go-trafilatura    - si no se usa
✗ colly/v2         - legacy
```

**Opción B: Implementaciones propias (workaround)**
```
✓ golang.org/x/net - parsing HTML nativo
✓ yaml.v3          - configs
✓ google/uuid      - embeds
✗ Parser CSS custom - regex-based selectors
✗ Sanitización custom - allowlist manual
```

**Tradeoff:** Menos deps pero más código propio para mantener.

---

## Recomendación Actual

**Dejar todo como está hasta tener datos de producción.**

Razones:
1. No sabés qué casos reales van a aparecer
2. Trafilatura puede ser necesaria para páginas complejas
3. El overhead de 30 transitive deps es aceptable hoy
4. Medir antes de optimizar

---

## Métricas a Implementar (Phase 2)

```go
// En chain.go, agregar:
result.ExtractorUsed  // "readability", "trafilatura", "fallback"
result.QualityScore   // para comparar

// En logs (futuro):
logger.Info().
    Str("extractor", name).
    Int("word_count", wordCount).
    Msg("extraction_result")
```

Con esto, después de un período de uso, vas a poder responder:
- "¿Trafilatura se usa en producción?"
- "¿Colly alguna vez gana?"
- "¿Vale la pena el peso de las dependencias?"

---

[Volver al índice](README.md)
