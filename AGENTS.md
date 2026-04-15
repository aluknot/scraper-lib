# AGENTS.md

This file defines how the AI agent must operate in this repository.

---

## Overview

**scraper-lib** is a Go library for web scraping and article content extraction.

It uses an extractor chain:
domain_specific → readability → trafilatura → colly

The system includes caching, sanitization, embed preservation, and multiple output formats.

---

## Skill-Based Workflow (CRITICAL)

This project uses a **skill-driven execution model** with OpenCode.

Skills are located in:
.opencode/skills/<skill-name>/SKILL.md

---

### Language Handling

User prompts may be written in Spanish or English.

The agent MUST:
- Infer intent semantically (not keyword-based)
- NOT rely on exact English trigger phrases
- Map meaning → skill regardless of language

Examples:
- "hacé un endpoint" → API design
- "esto no funciona" → debugging
- "armemos un plan" → planning
- "optimizar este código" → refactoring

---

### Core Rules

- If a task matches a skill → you MUST use it
- NEVER implement directly if a skill applies
- ALWAYS follow the skill workflow completely
- DO NOT skip steps (spec, plan, test, etc.)
- DO NOT partially apply a skill

---

### Execution Model

For EVERY request:

1. Interpret intent (semantic, not literal)
2. Check if ANY skill applies (even minimal relevance)
3. If yes → invoke the skill
4. Follow the skill instructions STRICTLY
5. Only proceed after required steps are completed

---

### Lifecycle

The agent must internally follow this lifecycle:

1. DEFINE → `spec-driven-development`
2. PLAN → `planning-and-task-breakdown`
3. BUILD → `incremental-implementation` + `test-driven-development`
4. VERIFY → `debugging-and-error-recovery`
5. REVIEW → `code-review-and-quality`
6. SHIP → `shipping-and-launch`

---

### Anti-Rationalization

The following thoughts are WRONG and must be ignored:

- "This is too small for a skill"
- "I can just implement this quickly"
- "I'll explore first and decide later"
- "No need for full process"

Correct behavior:

- ALWAYS evaluate skill usage first
- ALWAYS prefer structured execution

---

### When NOT to Use a Skill

Only skip skills if:

- The request is purely informational (explanation, theory, concepts)
- No implementation, modification, or execution is required

If there is ANY action:
→ a skill MUST be used

---

## Build, Test, and Development Commands

### Build

go build ./...

### Test (includes integration tests)

go test -tags=integration ./...
go test -tags=integration -v ./...
go test -tags=integration -cover ./...

### Targeted Tests

go test -tags=integration -v -run "TestName" ./package
go test -tags=integration -v -run "^TestPrefix" ./...
go test -tags=integration -v ./specific/file_test.go ./specific/file.go

### Coverage Report

go test -tags=integration -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

---

### Lint and Format

go fmt ./...
go vet ./...
go vet -all ./...
go mod tidy

---

## Code Style (Summary)

Follow Go best practices with these key rules:

### Structure
- Keep implementation under `internal/`
- Expose public API only from root package
- Define interfaces where they are used (consumer side)

### Naming
- Packages: lowercase, short (`cache`, `fetch`)
- Public symbols: PascalCase
- Private symbols: camelCase
- Errors: prefixed with `Err`

### Errors
- Use wrapped errors: `fmt.Errorf("context: %w", err)`
- Return early, avoid nesting
- Define structured errors when needed

### Context
- First argument in I/O functions: `context.Context`
- Respect cancellation (`ctx.Err()`)

### Testing
- Prefer table-driven tests
- Use `httptest.NewServer` for HTTP
- Use `t.TempDir()` for filesystem
- Naming: `Test<Subject>_<Scenario>`

### Concurrency
- Use `sync.RWMutex` for shared state
- Use `sync.WaitGroup` when coordinating goroutines

---

## Architecture (High-Level)

Public API:

- Extract()
- ExtractHTML()
- Options
- Result

Internal structure:

internal/
- cache/
- detection/
- dom/
- extractors/
- fetch/
- markdown/
- output/
- types/
- urlutil/

---

## Documentation Reference

For detailed information, consult the docs folder:

### Core Docs
- docs/overview.md
- docs/architecture.md
- docs/strategy-pipeline.md

### Extraction System
- docs/domain-rules.md
- docs/detection.md
- docs/output-pipeline.md
- docs/embeds.md

### Engineering
- docs/code-quality.md
- docs/testing.md
- docs/error-flow.md

### Additional
- docs/adr.md
- docs/implementation-status.md
- docs/dependencies.md
- docs/migration.md

---

## Git Workflow (REQUIRED)

At the end of each session, the agent MUST commit all changes:

### Commit Rules
- **When:** Before ending session or when explicitly requested
- **Format:** Conventional commits (feat:, fix:, docs:, refactor:, test:)
- **Scope:** Use relevant scope (e.g., `feat/api`, `fix/cache`, `docs/diagrams`)

### Commit Message Format
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Examples
```bash
git add .
git commit -m "feat(api): add NoEmbeds and NoSanitize pipeline flags"
git commit -m "fix(cache): disablecache flag now properly bypasses cache"
git commit -m "docs: update diagrams for v10.0 composable API"
```

### Verification Before Commit
1. `go build ./...` - must pass
2. `go test ./...` - must pass
3. `go vet ./...` - must pass

If any step fails, fix before committing.

---

## Key Principle

The agent is NOT a free-form coder.

It is a **skill executor**.

- No shortcuts
- No improvisation outside skills
- No direct implementation without process

Consistency > speed  
Structure > intuition
