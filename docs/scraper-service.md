# scraper-service (HTTP API)

> **Status: MVP** — HTTP API básica con PostgreSQL.
> Phase 3+: Redis si es necesario.

## Estructura

```text
scraper-service/
├── cmd/server/main.go
├── internal/
│   ├── api/
│   │   ├── handler.go
│   │   ├── middleware.go
│   │   └── router.go
│   │
│   ├── cache/
│   │   ├── interface.go   # Cache interface — permite migrar a Redis sin tocar lógica
│   │   ├── postgres.go    # Implementación PostgreSQL (día 1)
│   │   └── worker.go      # Background goroutine: DELETE WHERE expires_at < NOW()
│   │
│   ├── ratelimit/
│   │   ├── interface.go
│   │   └── postgres.go    # Implementación PostgreSQL con upsert atómico
│   │
│   ├── config/
│   │   ├── domains.yaml
│   │   └── extractors.yaml
│   │
│   └── migrations/
│       └── 001_init.sql
│
└── go.mod                 # Importa github.com/aluknot/scraper-lib
```

---

## RateLimiter

> **Status: MVP** — PostgreSQL. Migrar a Redis en Phase 3 si la contención es problema.

### Interfaz

```go
type RateLimitConfig struct {
    Requests int           `yaml:"requests"`
    Window   time.Duration `yaml:"window"`
    Burst    int           `yaml:"burst"`
}

type RateLimiter interface {
    Allow(domain string) bool
    GetState(domain string) (*RateLimitState, error)
    Reset(domain string)
    WindowRemaining(domain string) time.Duration
    Config(domain string) *RateLimitConfig
}
```

### Implementación PostgreSQL

```go
type PostgresRateLimiter struct {
    db        *sql.DB
    defaults  RateLimitConfig
    overrides map[string]RateLimitConfig
}

func (r *PostgresRateLimiter) Config(domain string) *RateLimitConfig {
    if cfg, ok := r.overrides[domain]; ok {
        return &cfg
    }
    return &r.defaults
}

// GetState usa upsert atómico para evitar race conditions
func (r *PostgresRateLimiter) GetState(domain string) (*RateLimitState, error) {
    var state RateLimitState
    err := r.db.QueryRow(`
        INSERT INTO rate_limits (domain, requests, window_start)
        VALUES ($1, 1, NOW())
        ON CONFLICT (domain) DO UPDATE
        SET requests = CASE
            WHEN NOW() - rate_limits.window_start > $2 THEN 1
            ELSE rate_limits.requests + 1
        END,
        window_start = CASE
            WHEN NOW() - rate_limits.window_start > $2 THEN NOW()
            ELSE rate_limits.window_start
        END
        RETURNING requests, window_start
    `, domain, r.Config(domain).Window).Scan(&state.Requests, &state.WindowStart)
    return &state, err
}
```

---

## PostgreSQL: Advertencias Operacionales

### Limpieza de caché

PostgreSQL no expira filas automáticamente. Se necesita un worker:

```sql
DELETE FROM scrape_cache WHERE expires_at < NOW()
```

```bash
CACHE_CLEANUP_INTERVAL=1h    # Default
CACHE_CLEANUP_INTERVAL=30m   # Alta rotación
CACHE_CLEANUP_INTERVAL=24h   # Contenido estático
```

### Contención en rate limiting

La interfaz permite migrar a Redis cuando la contención de fila se vuelve problema. El cambio es solo de implementación, no de interfaz.

---

## Tablas PostgreSQL

### Cache de resultados

```sql
CREATE TABLE scrape_cache (
    id         BIGSERIAL PRIMARY KEY,
    url_hash   TEXT NOT NULL UNIQUE,
    url        TEXT NOT NULL,
    output     TEXT NOT NULL,
    result     JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_scrape_cache_url_hash ON scrape_cache(url_hash);
CREATE INDEX idx_scrape_cache_expires  ON scrape_cache(expires_at);
```

### Rate limiting por dominio

```sql
CREATE TABLE rate_limits (
    id           BIGSERIAL PRIMARY KEY,
    domain       TEXT NOT NULL UNIQUE,
    requests     INT NOT NULL DEFAULT 0,
    window_start TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Stats para métricas

```sql
CREATE TABLE scrape_stats (
    id              BIGSERIAL PRIMARY KEY,
    date            DATE NOT NULL,
    output          TEXT NOT NULL,
    extractor       TEXT NOT NULL,
    success_count   INT NOT NULL DEFAULT 0,
    error_count     INT NOT NULL DEFAULT 0,
    avg_duration_ms BIGINT NOT NULL DEFAULT 0,
    UNIQUE(date, output, extractor)
);
```

### Errores para debugging

```sql
CREATE TABLE fetch_errors (
    id          BIGSERIAL PRIMARY KEY,
    url         TEXT NOT NULL,
    error       TEXT NOT NULL,
    extractor   TEXT,
    strategy    TEXT,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

[Volver al índice](README.md)
