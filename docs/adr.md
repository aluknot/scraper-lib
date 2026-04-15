# Architecture Decision Records (ADRs)

## ADR-001: PostgreSQL sobre Redis para MVP

**Decisión:** Usar PostgreSQL como backend de caché y rate limiting en el MVP.

**Razones:**
- Ya tenemos PostgreSQL en scraper-service (migraciones compartidas)
- Menor complejidad operativa: no hay que mantener otro servicio
- El volumen inicial del MVP no justifica Redis

**Cuándo migrar a Redis:**
- Contención de fila en rate_limits (`ON CONFLICT DO UPDATE`) causa latencia
- Necesidad de TTL nativo (PostgreSQL requiere worker de limpieza)
- Volumen > 1000 req/s en caché

---

## ADR-002: Inyección de Dependencias sobre `init()`

**Decisión:** Los extractores se registran explícitamente via `NewChain(...)`, no con auto-registro vía `init()`.

**Razones:**
- Testeable: se puede crear un chain con extractores mock
- Explícito: se ve qué extractores están activos en cada contexto
- Evita side effects silenciosos de `init()` en imports

**Patrón:** `DefaultChain()` crea el chain estándar; tests usan `NewChain(mock1, mock2)`.

---

## ADR-003: Embeds como Pre/Post Processor, no como Extractor

**Decisión:** Los embeds (YouTube, Twitter, etc.) se extraen ANTES del extractor chain y se restauran DESPUÉS.

**Razones:**
- Readability/Trafilatura eliminan iframes y scripts (los tratan como ads)
- No tiene sentido como extractor porque no extrae contenido, lo preserva
- El orden es crítico: sanitizar ANTES de restaurar, o bluemonday elimina iframes

---

## ADR-004: Browser Service Separado (Python + Playwright)

**Decisión:** El browser headless NO va en el binario Go. Se implementa como microservicio Python separado.

**Razones:**
- Playwright requiere Chromium + librerías de sistema (ensucia el entorno Go)
- Escala independientemente (es el más caro en recursos)
- Se puede reemplazar sin tocar el servicio principal
- Contenedor Docker separado evita instalar Chromium en la máquina

---

## ADR-005: Domain Rules desde YAML, no Hardcodeadas

**Decisión:** Las reglas por dominio se cargan desde `config/extractors.yaml`, no desde código.

**Razones:**
- Cambios en runtime sin recompilar
- Versionado y rollback de reglas
- Escalable: con 20+ sites, el código sería inmanejable
- Validación en load time previene errores en runtime

---

## ADR-006: Bot Detection con MAX de señales, no Promedio

**Decisión:** El score de detección de bot usa `max(señales)` en vez de `avg(señales)`.

**Razones:**
- Si detectás Cloudflare (1.0) Y JS render (0.9), el promedio (0.95) baja la confianza
- Con `max`, múltiples señales fuertes refuerzan la decisión
- El promedio diluye señales fuertes con señales débiles

---

## ADR-007: Rate Limiter con Upsert Atómico

**Decisión:** `GetState()` usa `INSERT ... ON CONFLICT DO UPDATE` en vez de SELECT + UPDATE separado.

**Razones:**
- Previene race condition entre SELECT y UPDATE en requests concurrentes
- Cada request lee y actualiza el estado atómicamente
- Sin upsert, dos requests simultáneos leen el mismo contador y ambos pasan

---

[Volver al índice](README.md)

---

## ADR-008: CollyExtractor usa DOM en memoria, no HTTP

**Decisión:** El CollyExtractor no debe hacer una petición HTTP nueva via `c.Visit()`. En su lugar, parsea el HTML ya descargado con goquery.

**Razones:**
- El HTML ya fue obtenido upstream por `fetch.GetHTML()`
- Hacer un segundo request es I/O redundante e innecesario
- El contenido puede cambiar entre el primer y segundo request
- Riesgo de rate limiting por peticiones duplicadas al mismo dominio

**Implementación:** Reemplazar `colly.NewCollector()` + `c.Visit()` con `goquery.NewDocumentFromReader()` aplicando los mismos selectores CSS al DOM en memoria.

---

## ADR-009: Utilidades compartidas en `internal/dom/` y `internal/urlutil/`

**Decisión:** Las funciones duplicadas (`getAttr`, domain extraction) se consolidan en paquetes utilitarios compartidos.

**Razones:**
- `getAttr()` estaba duplicada en `embeds.go` y `paywall.go` — misma implementación
- Domain extraction existía en 3 paquetes distintos con lógica casi idéntica
- Un solo punto de truth evita inconsistencia y facilita testing

**Paquetes creados:**
- `internal/dom/` — utilidades de manipulación de DOM HTML (`Attr`, etc.)
- `internal/urlutil/` — utilidades de URL (`Domain(url)`, `Path(url)`)

---

## ADR-010: Una sola implementación de HTML→Markdown

**Decisión:** Consolidar las dos implementaciones de HTML→Markdown en una sola. La versión regex-based en `internal/markdown/html2md.go` es la canónica.

**Razones:**
- `domain_specific.go` tenía `htmlToMarkdown()` (DOM-walking, ~100 líneas)
- `html2md.go` tenía `HTMLToMarkdown()` (regex-based, más simple)
- Producían output diferente para el mismo input — inconsistencia
- Mantener dos converters es costo doble de testing y bugfixing

**Implementación:** `domain_specific.go` importa y usa `markdown.HTMLToMarkdown()`.

---

## ADR-011: Calidad de código antes de Phase 2

**Decisión:** Agregar una fase "Quality" entre MVP y Phase 2 para resolver issues de correctness y maintainability antes de agregar features nuevas.

**Razones:**
- Issues correctness (Colly HTTP redundante, defer body close) afectan confiabilidad
- Duplicación de código hace más caro mantener y extender
- Código muerto (teaser.go, ErrCircuitOpen) crea confusión sobre qué está activo
- Sin structured logging, es difícil debuggear issues en Phase 2+
- Sin testdata fixtures, la cobertura de edge cases es limitada

**Ver:** [docs/code-quality.md](code-quality.md) para lista completa de issues y orden de resolución.
