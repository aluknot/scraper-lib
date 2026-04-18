# Changelog

| Fecha | Cambio |
|-------|--------|
| 2026-04-17 | **v10.3** — Debug logging: logs detallados por extractor (extractor_error, extractor_low_quality, extractor_success, all_extractors_failed), HTML size y preview en fetch, nueva opción Debug en Options para verbose logging. |
| 2026-04-17 | **v10.2** — Platform extractors: nuevo paquete extractors/platforms/ con factory y fetcher, YouTube extractor (VideoMetadata, ChannelMetadata, VideoContent, ShortContent), GitHub extractor (RepoMetadata, ProfileMetadata, ReadmeContent), renamed colly.go a fallback.go, extractores registrados automáticamente vía init(). |
| 2026-04-15 | **v10.1** — Structured logging con log/slog (cache_hit/miss, fetch, extraction_complete, paywall_detected), InMemoryCache con límite LRU (10000 entries default, NewInMemoryCacheWithLimit), testdata fixtures (7 HTML files), tests con DisableCache para evitar cache conflicts. 70+ tests, ~40 archivos. |
| 2026-04-15 | **v10.0** — API composable: componentes públicos exportados (extractors/, output/, types/, embeds/, sanitize/), pipeline stages opcionales (NoEmbeds, NoSanitize, NoPaywallDetection, DisableCache), quality fixes (C1-C2, H1-H5, M5-M7, M9). |
| 2026-04-12 | **v9.5** — HTTP Advanced (UA rotation, cookie jar, referrer spoofing, Sec-Fetch headers), integration tests con build tag, error types (ErrFetchFailed, ErrAllStrategiesFailed, ErrCircuitOpen). 59 tests, 30 archivos. |
| 2026-04-12 | **v9.4** — Output Markdown con templates YAML, HTML→Markdown converter, auto-tag generation, template resolution por extractor+categoria. 38 tests, 25 archivos. |
| 2026-04-12 | **v9.3** — Cache (InMemory + File), custom extractor chain (Extractor/Extractors/NoFallback). 28 tests, 21 archivos. |
| 2026-04-11 | **v9.1** — Código base extraído de Rissy: 18 archivos, 22 tests pasando. PLAN.md reorganizado en docs/. |
| 2026-04-10 | **v9** — Architecture v2 Executable & Scalable: Integración de PLAN2.md, priorización por fases (MVP → Phase 4), labels de status por sección, tabla de fases, lista explícita de "NO incluir en MVP", orden de implementación actualizado. |
| 2026-04-10 | **v8.1** — Mejoras post-review: CRITICAL FIX path_regex (regex en vez de glob), CRITICAL FIX bot detection con max(signals), CRITICAL FIX rate limiter con upsert atómico, NEW circuit breaker, NEW validación de reglas, NEW outputs compuestos, NEW Instagram + TikTok embeds, NEW diagrama de errores, NEW qualityScore documentado, NEW paywall detection, NEW estrategia de testing. |
| 2026-04-10 | **v8** — Architecture v2: Domain rules YAML, bot detection estructural, fix shouldEscalate(rawHTML), RateLimiter con windowing, PriceExtractor + JobExtractor, OutputPipeline, EscalationConfig. |

---

[Volver al índice](README.md)
