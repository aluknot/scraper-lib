# scraper-lib — Documentación

## Índice

### Diseño y Arquitectura

| Documento | Descripción |
|-----------|-------------|
| [diagrams.md](diagrams.md) | **Diagramas visuales** (Mermaid): arquitectura, pipeline, embeds, chain, packages, errores |
| [overview.md](overview.md) | Descripción general, **API en dos niveles**, estructura de repositorios, fases de implementación |
| [error-flow.md](error-flow.md) | Diagrama de flujo de errores (9 pasos), tipos de error estructurados |
| [architecture.md](architecture.md) | Extractor Chain (DI Hybrid), interfaz Extractor, API pública, estructura del paquete, **cache**, **custom chain**, **output markdown** |
| [embeds.md](embeds.md) | Sistema de embeds como pre/post processor (YouTube, Twitter, Vimeo, Spotify, etc.) |
| [domain-rules.md](domain-rules.md) | Extractores especializados por dominio con configuración YAML dinámico |
| [strategy-pipeline.md](strategy-pipeline.md) | Pipeline de estrategias fetch, escalación, bot detection estructural, circuit breaker |
| [detection.md](detection.md) | Detección de paywalls (MVP) y qualityScore completo (Phase 2+) |
| [output-pipeline.md](output-pipeline.md) | Pipeline de outputs: article, metadata, raw, markdown (obsidian), price, job |

### Servicio y Operaciones

| Documento | Descripción |
|-----------|-------------|
| [scraper-service.md](scraper-service.md) | Servicio HTTP: API, cache PostgreSQL, rate limiter, tablas SQL |
| [migration.md](migration.md) | Plan de migración desde Rissy |
| [testing.md](testing.md) | Estrategia de testing: pirámide, unit tests, fixtures, integration tests |

### Decisiones y Estado

| Documento | Descripción |
|-----------|-------------|
| [adr.md](adr.md) | Architecture Decision Records (7 ADRs documentados) |
| [implementation-status.md](implementation-status.md) | Estado actual de implementación: qué está hecho, qué falta, **issues de calidad de código** |
| [code-quality.md](code-quality.md) | **Revisión completa de código**: hallazgos por severidad (críticos, altos, medios, bajos), orden de resolución recomendado |
| [dependencies.md](dependencies.md) | Lista de dependencias Go para scraper-lib y scraper-service |
| [dependencies-strategy.md](dependencies-strategy.md) | **Plan futuro** para reducir/evaluar dependencias basado en datos de producción |
| [changelog.md](changelog.md) | Historial de versiones (v8 → v10.0) |

---

## Estado rápido

- **Versión actual:** v10.1 (API composable + logs + LRU cache)
- **Tests:** 70+ pasando
- **Archivos de código:** ~35
- **API:** Nivel 1 (Extract) + Nivel 2 (componentes públicos)
- **Features:** Cache, pipeline stages opcionales (NoEmbeds, NoSanitize, etc.), extractors/exportables, output extensible
