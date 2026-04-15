# Dependencias

## scraper-lib (Go)

| Paquete | Uso |
|---------|-----|
| `codeberg.org/readeck/go-readability/v2` | Article extraction |
| `github.com/markusmobius/go-trafilatura` | Article extraction fallback |
| `github.com/gocolly/colly/v2` | HTML scraping último recurso |
| `github.com/PuerkitoBio/goquery` | CSS selectors (domain-specific extractor) |
| `github.com/microcosm-cc/bluemonday` | HTML sanitization |
| `github.com/google/uuid` | RunIDs para placeholders de embeds |
| `golang.org/x/net/html` | DOM parsing |
| `gopkg.in/yaml.v3` | Lectura de config YAML |

## scraper-service (Go)

| Paquete | Uso |
|---------|-----|
| `github.com/go-chi/chi/v5` | HTTP router |
| `github.com/lib/pq` | PostgreSQL driver |
| `github.com/prometheus/client_golang` | Métricas |
| `github.com/aluknot/scraper-lib` | Core scraping |

---

[Volver al índice](README.md)
