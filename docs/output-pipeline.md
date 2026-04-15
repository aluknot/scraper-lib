# Outputs Especializados

> **Status: MVP** — Article, Metadata, Raw implementados.
> Price y Job se implementan cuando hay caso concreto.

En v7 los outputs (price, job, markdown/obsidian) estaban definidos pero no había forma de extraer esos datos. En v9 agregamos un output pipeline que popula los campos específicos. El output Markdown (antes "obsidian") genera Markdown con frontmatter YAML, auto-tags y template resolution.

---

## Output Pipeline

```go
// internal/output/article.go

func BuildArticleResult(extracted *types.ExtractResult, url string) *ArticleResult {
    result := &ArticleResult{
        Title:   extracted.Title,
        Content: extracted.Content,
        Author:  extracted.Author,
        URL:     url,
    }
    if extracted.PublishedAt != nil {
        result.PublishedAt = extracted.PublishedAt.Format("2006-01-02")
    }
    if extracted.Language != "" {
        result.Language = extracted.Language
    }
    return result
}

func BuildMetadataResult(extracted *types.ExtractResult, url string) *MetadataResult {
    return &MetadataResult{
        Title:        extracted.Title,
        Author:       extracted.Author,
        Language:     extracted.Language,
        WordCount:    extracted.WordCount,
        Images:       len(extracted.Images),
        Videos:       len(extracted.Videos),
        Links:        len(extracted.Links),
        Extractor:    extracted.ExtractorUsed,
        QualityScore: extracted.QualityScore,
        Warnings:     extracted.Warnings,
    }
}

func BuildRawResult(extracted *types.ExtractResult) *RawResult {
    return &RawResult{
        Content:   extracted.Content,
        Extractor: extracted.ExtractorUsed,
        Attempts:  extracted.Attempts,
        WordCount: extracted.WordCount,
    }
}
```

---

## Outputs Compuestos

```go
// Solo artículo
result, _ := Extract(ctx, url, &Options{Outputs: []string{"article"}})

// Artículo + metadata (caso típico para RSS readers)
result, _ := Extract(ctx, url, &Options{Outputs: []string{"article", "metadata"}})

// Precio + artículo (finanzas: comparar precios con descripción)
result, _ := Extract(ctx, url, &Options{Outputs: []string{"price", "article"}})

// Todo lo disponible (para debugging o uso completo)
result, _ := Extract(ctx, url, &Options{Outputs: []string{"full"}})
```

---

## Price Extractor

> **Status: Phase 2+** — Se implementa cuando hay caso concreto de finanzas.

```go
type PriceInfo struct {
    ProductName  string
    Price        float64
    Currency     string
    Availability string // in_stock, out_of_stock, preorder
    Brand        string
    SKU          string
}

type priceRule struct {
    Domain   string `yaml:"domain"`
    Price    string `yaml:"price"`    // CSS selector
    Product  string `yaml:"product"`  // CSS selector
    Brand    string `yaml:"brand"`    // CSS selector
    Currency string `yaml:"currency"` // Fija o metadata
}

var priceRules = []priceRule{
    {
        Domain:  "amazon.com",
        Price:   ".a-price-whole",
        Product: "#productTitle",
        Brand:   "#brand",
    },
    {
        Domain:   "mercadolibre.com.ar",
        Price:    ".andes-money-amount__fraction",
        Product:  ".ui-pdp-title",
        Currency: "ARS",
    },
}
```

---

## Job Extractor

> **Status: Phase 2+** — Skeleton implementado. Se completa con caso concreto.

```go
type JobInfo struct {
    Company      string
    Location     string
    Salary       string
    JobType      string // full_time, part_time, contract
    Remote       bool
    Requirements []string
    Benefits     []string
    ApplyURL     string
}

type jobRule struct {
    Domain       string `yaml:"domain"`
    Title        string `yaml:"title"`
    Company      string `yaml:"company"`
    Location     string `yaml:"location"`
    Salary       string `yaml:"salary"`
    Description  string `yaml:"description"`
    Requirements string `yaml:"requirements"`
    ApplyURL     string `yaml:"apply_url"`
}

var jobRules = []jobRule{
    {
        Domain:      "linkedin.com",
        Title:       "h1.topcard__title",
        Company:     "a.topcard__flavor",
        Location:    "span.topcard__flavor--bullet",
        Description: "div.show-more-less-html__markup",
    },
    {
        Domain:      "indeed.com",
        Title:       "h1.jobsearch-JobInfoHeader-title",
        Company:     "span.companyName",
        Location:    "div.jobsearch-JobInfoHeader-subtitle",
        Description: "div.jobsearch-jobDescriptionText",
    },
}
```

### Fallback: schema.org/JobPosting

Cuando no hay regla para el dominio, el Job Extractor intenta extraer datos de `<script type="application/ld+json">` con `"@type": "JobPosting"`.

---

[Volver al índice](README.md)
