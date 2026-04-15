# Extractores Especializados por Dominio

> **Status: Implementado.**

## El problema con domainRules hardcodeado

En v7 las reglas estaban hardcodeadas en un slice en memoria. Tres problemas:

1. **No se puede cambiar en runtime** — requiere rebuild para agregar reglas
2. **No hay versionado** — cambios en reglas rompen sin rollback
3. **Escalabilidad** — con 20+ sites con reglas propias, el código se vuelve inmanejable

## Solución: Cargar desde configuración YAML

Las reglas se cargan dinámicamente desde `config/extractors.yaml` en tiempo de inicialización.

```go
type domainRule struct {
    Domain    string `yaml:"domain"`      // "github.com", "wikipedia.org"
    PathRegex string `yaml:"path_regex"`  // regex pattern
    Type      string `yaml:"type"`        // "readme", "releases", "article"
    Selector  string `yaml:"selector"`    // CSS selector principal
    Alt       string `yaml:"alt"`         // Selector alternativo
    Transform string `yaml:"transform"`   // "" = HTML, "markdown"
}
```

### Implementación resumida

```go
func NewDomainSpecificExtractor() *DomainSpecificExtractor {
    rules := loadDefaultRules()
    if external, err := loadExternalRules(""); err == nil {
        rules = external
    }
    // Precompilar regexes
    compiledREs := make([]*regexp.Regexp, len(rules))
    for i, r := range rules {
        re, _ := regexp.Compile(r.PathRegex)
        compiledREs[i] = re
    }
    return &DomainSpecificExtractor{rules: rules, compiledREs: compiledREs}
}

func (e *DomainSpecificExtractor) Extract(ctx context.Context, htmlContent, url string) (*types.ExtractResult, error) {
    domain := extractDomain(url)
    path := extractPath(url)

    // Buscar regla que matchee dominio + path
    for i := range e.rules {
        if e.rules[i].Domain != domain {
            continue
        }
        if e.compiledREs[i] != nil && e.compiledREs[i].MatchString(path) {
            // Extraer con CSS selector
            doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
            content, _ := doc.Find(e.rules[i].Selector).Html()
            if content == "" && e.rules[i].Alt != "" {
                content, _ = doc.Find(e.rules[i].Alt).Html()
            }
            if content == "" {
                return nil, ErrSelectorNotFound{URL: url, Selector: e.rules[i].Selector}
            }

            result := &types.ExtractResult{
                Content:       content,
                ExtractorUsed: fmt.Sprintf("domain_specific:%s", e.rules[i].Type),
                WordCount:     countWords(content),
            }
            if e.rules[i].Transform == "markdown" {
                result.Content = htmlToMarkdown(content)
            }
            return result, nil
        }
    }
    return nil, ErrNoRuleForDomain{Domain: domain, Path: path}
}
```

## Validación en Load Time

```go
func ValidateRule(rule domainRule) error {
    if _, err := regexp.Compile(rule.PathRegex); err != nil {
        return fmt.Errorf("invalid path_regex %q: %w", rule.PathRegex, err)
    }
    if !isValidSelector(rule.Selector) {
        return fmt.Errorf("invalid selector %q", rule.Selector)
    }
    validTypes := map[string]bool{
        "readme": true, "releases": true, "article": true,
        "product": true, "job": true, "list": true,
    }
    if !validTypes[rule.Type] {
        return fmt.Errorf("unknown type %q", rule.Type)
    }
    return nil
}
```

## Configuración externa

```yaml
# config/extractors.yaml
rules:
  - domain: "github.com"
    path_regex: "/.*"
    type: "readme"
    selector: "article.markdown-body"
    alt: "article#readme"
    transform: "markdown"

  - domain: "github.com"
    path_regex: "/[^/]+/[^/]+/releases/.*"
    type: "releases"
    selector: "div.markdown-body"

  - domain: "wikipedia.org"
    path_regex: "/.*"
    type: "article"
    selector: "#mw-content-text"
```

## Cuándo agregar nuevas reglas

Agregar una entrada a `config/extractors.yaml` cuando:
- El extractor genérico (readability/trafilatura) falla consistentemente en ese dominio
- La estructura HTML del sitio es estable y conocida
- Hay un caso de uso concreto (no por anticipación)

---

[Volver al índice](README.md)
