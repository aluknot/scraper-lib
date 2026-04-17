# Diagramas — scraper-lib v10.0

> Diagramas Mermaid renderizables en GitHub, Obsidian, VS Code preview.

---

## 1. API en Dos Niveles

```mermaid
flowchart TB
    subgraph L1["Nivel 1: Extract() — Convenience API"]
        E["Extract(ctx, url, &Options{...})"]
        EH["ExtractHTML(ctx, html, url, &Options{...})"]
    end

    subgraph L2["Nivel 2: Componentes Públicos — Composable API"]
        FE["fetch (custom)"]
        EMB["embeds.ExtractAndReplace()"]
        CHA["extractors.Chain.Extract()"]
        SAN["sanitize.Clean()"]
        RES["embeds.Restore()"]
        OUT["output.BuildArticleResult()"]
    end

    L1 -->|"Usa internamente"| L2

    L2 --> FE
    FE --> EMB
    EMB --> CHA
    CHA --> SAN
    SAN --> RES
    RES --> OUT

    classDef l1 fill:#2d6a4f,stroke:#1b4332,color:#fff
    classDef l2 fill:#457b9d,stroke:#1d3557,color:#fff
    classDef pub fill:#e9ecef,stroke:#495057
    class E,EH l1
    class FE,EMB,CHA,SAN,RES,OUT l2
```

### Paquetes Públicos

| Paquete | Componentes |
|---------|-------------|
| `extractors/` | `Chain`, `NewChain()`, `DefaultChain()`, `Extractor` interface |
| `output/` | `BuildArticleResult()`, `BuildMetadataResult()`, `BuildRawResult()`, `BuildMarkdownResult()` |
| `types/` | `ExtractResult`, `Attempt`, `StrategyAttempt`, `PriceInfo`, `JobInfo` |
| `embeds/` | `EmbedExtractor`, `NewEmbedExtractor()`, `ExtractAndReplace()`, `Restore()` |
| `sanitize/` | `Sanitizer`, `NewSanitizer()`, `Clean()` |

---

## 2. Arquitectura General

```mermaid
flowchart LR
    lib["scraper-lib\n(github.com/aluknot/scraper-lib)"]

    lib --> svc["scraper-service\nHTTP API (chi + PostgreSQL)"]
    lib --> rissy["Rissy\nRSS aggregator"]
    lib --> otros["Otros proyectos\nFinanzas, Obsidian, etc."]

    classDef lib fill:#2d6a4f,stroke:#1b4332,color:#fff
    classDef consumer fill:#457b9d,stroke:#1d3557,color:#fff
    class lib lib
    class svc,risy,otros consumer
```

## 3. Pipeline `Extract()` con Stages Opcionales

```mermaid
flowchart TD
    Start(["Extract(url)"])

    Start --> C0["0. Check cache"]
    C0 -->|hit| RetC["Return cached ✅"]
    C0 -->|miss| S1["1. fetchHTML\n(con retry + exponential backoff)"]

    S1 -->|fail| Err1["ErrFetchFailed ❌"]
    S1 -->|ok| S2{"NoPaywallDetection?"}

    S2 -->|false (default)| S2A["2a. detectPaywall\n(marker, no bypass)"]
    S2A --> S3
    S2 -->|true| S3["3. extractAndReplaceEmbeds"]

    S3{"NoEmbeds?"}
    S3 -->|false (default)| S3A["3a. extractAndReplaceEmbeds\n(PRE: placeholders + embedMap)"]
    S3A --> S4
    S3 -->|true| S4["4. chain.Extract"]

    S4["4. chain.Extract\ndomain_specific → readability → trafilatura → fallback"]

    S4 -->|fail| Err5["ErrAllExtractorsFailed ❌"]
    S4 -->|ok| S6{"NoSanitize?"}

    S6 -->|false (default)| S6A["6a. sanitizeHTML\n(bluemonday UGCPolicy)"]
    S6A --> S7
    S6 -->|true| S7["7. restoreEmbeds"]

    S7{"NoEmbeds?"}
    S7 -->|false (default)| S7A["7a. restoreEmbeds\n(POST: iframes en posición original)"]
    S7A --> S9
    S7 -->|true| S9["9. buildResult"]

    S9["9. buildResult + cache store"] --> RetR["Return Result ✅"]

    classDef step fill:#e9ecef,stroke:#495057
    classDef opt fill:#fff3cd,stroke:#ffc107
    classDef error fill:#f8d7da,stroke:#dc3545,color:#721c24
    classDef success fill:#d1e7dd,stroke:#198754,color:#0f5132
    classDef pending fill:#a8dadc,stroke:#1d3557
    class S1,S2A,S3A,S4,S6A,S7A,S9,C0 step
    class S2,S3,S6,S7 opt
    class Err1,Err5 error
    class RetC,RetR success
```

> **Nota:** Stages 2, 3, 6, 7 son opcionales según las Options de Extract().

## 4. Embed Pre/Post Processor — Orden Crítico

```mermaid
flowchart TD
    HTML["rawHTML"] --> E["extractAndReplaceEmbeds\nPRE: iframe → [[EMBED_uuid_N]]"]
    E --> CH["chain.Extract\nreadability / trafilatura / fallback\n(placeholders sobreviven = texto)"]
    CH --> SA["sanitizeHTML\n⚠ bluemonday ANTES de restore"]
    SA --> RE["restoreEmbeds\nPOST: [[EMBED_uuid_N]] → iframe original"]
    RE --> OUT["Result con embeds preservados"]

    classDef pre fill:#fff3cd,stroke:#ffc107
    classDef chain fill:#e9ecef,stroke:#495057
    classDef warn fill:#f8d7da,stroke:#dc3545
    classDef post fill:#d1e7dd,stroke:#198754
    class E pre
    class CH chain
    class SA warn
    class RE post
```

## 5. Extractor Chain — Decision Tree

```mermaid
flowchart TD
    Start(["chain.Extract(html, url)"])

    Start --> DS["0. domain_specific\n(carga reglas desde config/extractors.yaml)"]

    DS -->|rule found + content| Done["✅ Return result"]
    DS -->|no rule| RD["1. readability\n(go-readability v2)"]

    RD -->|wordCount >= 100| Done
    RD -->|low quality| TF["2. trafilatura\n(go-trafilatura)"]

    TF -->|wordCount >= 100| Done
    TF -->|low quality| CL["3. fallback\n(goquery CSS selectors)"]

    CL -->|content extracted| Done
    CL -->|no content| Fail["❌ ErrAllExtractorsFailed"]

    classDef ext fill:#e9ecef,stroke:#495057
    classDef ok fill:#d1e7dd,stroke:#198754
    classDef fail fill:#f8d7da,stroke:#dc3545
    class DS,RD,TF,CL ext
    class Done ok
    class Fail fail
```

## 6. Package Dependency Graph (v10.0)

```mermaid
flowchart TD
    scraper["scraper.go\n(API pública: Extract, ExtractHTML)"]

    %% Paquetes públicos (exportados)
    scraper --> extractors["extractors/\nchain.go, interface.go\ndomain_specific.go, readability.go\ntrafilatura.go, fallback.go"]
    scraper --> output["output/\narticle.go"]
    scraper --> types["types/\ntypes.go"]
    scraper --> embeds["embeds/\nembeds.go"]
    scraper --> sanitize["sanitize/\nsanitize.go"]

    %% Paquetes internos
    scraper --> fetch["internal/fetch\nhttp.go, http_advanced.go"]
    scraper --> cache["internal/cache\nmemory.go, file.go\ninterface.go"]
    scraper --> detection["internal/detection\npaywall.go, teaser.go"]

    extractors --> types
    output --> markdown["internal/markdown\nhtml2md.go, template.go"]
    output --> types
    embeds --> dom["internal/dom\nattr.go"]
    extractors --> markdown
    extractors --> urlutil["internal/urlutil"]
    output --> urlutil
    fetch --> types
    detection --> dom

    config["config/extractors.yaml"] -.-> extractors

    classDef root fill:#2d6a4f,stroke:#1b4332,color:#fff
    classDef pub fill:#457b9d,stroke:#1d3557,color:#fff
    classDef internal fill:#6c757d,stroke:#495057,color:#fff
    classDef data fill:#a8dadc,stroke:#1d3557
    class scraper root
    class extractors,output,types,embeds,sanitize pub
    class fetch,cache,detection,markdown,dom,urlutil internal
    class config data
```

> **Leyenda:**
> - 🟢 **verde oscuro**: API raíz pública
> - 🟢 **verde azul**: Paquetes públicos exportados (usables por clientes)
> - ⚪ **gris**: Paquetes internos (implementación)

## 7. Error Types — Cuándo se generan

```mermaid
flowchart TD
    Start(["Extract(url)"])

    Start --> F["fetchHTML"]
    F -->|timeout, DNS, network| EF["ErrFetchFailed\n{URL, Attempts[], LastError}"]
    F -->|circuit open| EC["ErrCircuitOpen\n{Domain, CooldownRemaining}"]
    F -->|ok| E["chain.Extract"]

    E -->|all fail| EA["ErrAllExtractorsFailed\n{URL, Attempts[], QualityScore}"]
    E -->|ok| O["buildResult"]

    O --> R["✅ Result\n{Article/Metadata/Markdown,\n Warnings[], Attempts[], DurationMs}"]

    classDef error fill:#f8d7da,stroke:#dc3545
    classDef ok fill:#d1e7dd,stroke:#198754
    class EF,EC,EA error
    class O,R ok
```

---

[Volver al índice](README.md)
