# Changelog

| Fecha | Cambio |
|-------|--------|
| 2026-04-12 | **v9.5** — HTTP Advanced (UA rotation, cookie jar, referrer spoofing, Sec-Fetch headers), integration tests con build tag, error types (ErrFetchFailed, ErrAllStrategiesFailed, ErrCircuitOpen). 59 tests, 30 archivos. |
| 2026-04-12 | **v9.4** — Output Markdown con templates YAML, HTML→Markdown converter, auto-tag generation, template resolution por extractor+categoria. 38 tests, 25 archivos. |
| 2026-04-12 | **v9.3** — Cache (InMemory + File), custom extractor chain (Extractor/Extractors/NoFallback). 28 tests, 21 archivos. |
| 2026-04-11 | **v9.1** — Código base extraído de Rissy: 18 archivos, 22 tests pasando. PLAN.md reorganizado en docs/. |
| 2026-04-10 | **v9** — Architecture v2 Executable & Scalable: Integración de PLAN2.md, priorización por fases (MVP → Phase 4), labels de status por sección, tabla de fases, lista explícita de "NO incluir en MVP", orden de implementación actualizado. |
| 2026-04-10 | **v8.1** — Mejoras post-review: CRITICAL FIX path_regex (regex en vez de glob), CRITICAL FIX bot detection con max(signals), CRITICAL FIX rate limiter con upsert atómico, NEW circuit breaker, NEW validación de reglas, NEW outputs compuestos, NEW Instagram + TikTok embeds, NEW diagrama de errores, NEW qualityScore documentado, NEW paywall detection, NEW estrategia de testing. |
| 2026-04-10 | **v8** — Architecture v2: Domain rules YAML, bot detection estructural, fix shouldEscalate(rawHTML), RateLimiter con windowing, PriceExtractor + JobExtractor, OutputPipeline, EscalationConfig. |

---

[Volver al índice](README.md)
