# Velox - Home Media Server

## Project Overview
Velox is a self-hosted home media server (like Jellyfin/Emby but lighter).
- **Backend:** Go 1.26 + stdlib `net/http` (Go 1.22+ routing) + SQLite (WAL mode)
- **Frontend:** React 19 + TypeScript + Vite 8 + TailwindCSS 4 + React Compiler
- **Transcoding:** FFmpeg 8.0 / FFprobe

## Project Structure
```
backend/
  cmd/server/          # Entry point, CLI subcommands (migrate, version)
  internal/
    config/            # Env-based configuration
    database/          # SQLite connection + migration runner
      migrate/         # Versioned migrations (001_, 002_, ...)
    handler/           # HTTP handlers (REST)
    middleware/        # CORS, Logger, Recovery
    model/             # Domain structs
    repository/        # SQL queries (data access)
    scanner/           # File discovery + ffprobe
    service/           # Business logic
    transcoder/        # FFmpeg HLS transcoding
  pkg/
    ffprobe/           # FFprobe wrapper (public package)

webapp/
  src/
    components/        # Reusable UI components
    pages/             # Route-level page components
    hooks/             # Custom React hooks
    api/               # API client functions
    types/             # Shared TypeScript types
    lib/               # Utilities
```

## Architecture Decisions
- **Database:** SQLite only. WAL mode, `MaxOpenConns(1)`, `_foreign_keys=on`
- **Routing:** Go stdlib `net/http` with Go 1.22+ patterns (`GET /api/foo/{id}`)
- **No ORM:** Use sqlc (generated from SQL) or raw `database/sql`. Never GORM.
- **Migrations:** All schema changes via `internal/database/migrate/registry.go`. Never inline CREATE TABLE.
- **Auth:** JWT (short-lived 15min access + 7-day refresh). bcrypt cost 12.
- **Playback:** Direct Play first. HLS/transcode only when codec/container/audio incompatible.
- **Metadata:** TMDb as primary provider. NFO/local files override TMDb.

## Backend Rules (Go)

### Code Style
- Follow standard Go conventions (`gofmt`, `go vet`)
- Error handling: always check errors, wrap with context (`fmt.Errorf("doing X: %w", err)`)
- Use `context.Context` as first parameter in service/repository methods
- Name receivers with 1-2 letter abbreviations (`func (s *MediaService)`, `func (r *MediaRepo)`)
- Package names: singular, lowercase (`handler`, `service`, `repository`, `model`)
- No `utils` or `helpers` packages. Put functions where they belong.

### Patterns
- **Handler:** Parse request → call service → write JSON response. No business logic.
- **Service:** Business logic + orchestration. Calls repository. Returns domain errors.
- **Repository:** Pure SQL queries. One repo per table/aggregate. Returns model structs.
- **Model:** Plain structs with `json` tags. No methods beyond simple formatting.
- **Migrations:** Append new `{Version, Name, Up, Down}` to `All()` in registry.go. Each migration is transactional.

### Testing
- Table-driven tests: `tests := []struct{ name string; ... }{ ... }`
- Use `t.Run(tt.name, ...)` for subtests
- Test files next to source: `foo.go` → `foo_test.go`
- In-memory SQLite for DB tests: `sql.Open("sqlite3", ":memory:?_foreign_keys=on")`
- Run: `cd backend && make test`

### Linting
- `go vet` + `golangci-lint` (config: `backend/.golangci.yml`)
- Enabled linters: errcheck, staticcheck, sqlclosecheck, misspell, bodyclose

### Build & Run
```sh
cd backend
make dev          # go run ./cmd/server
make build        # go build -o bin/velox
make test         # go test ./... -v -count=1
make test-short   # go test ./... -short
make lint         # go vet + golangci-lint
make fmt          # gofmt -w -s
make migrate      # run migrations up
```

## Frontend Rules (React/TypeScript)

### Code Style
- TypeScript strict mode (no `any` unless absolutely necessary)
- Functional components only. No class components.
- React Compiler enabled — no manual `useMemo`/`useCallback` needed
- TailwindCSS 4 for styling. No CSS modules, no styled-components.
- Named exports for components, default exports only for pages.

### Patterns
- Pages in `src/pages/`, components in `src/components/`
- API calls in `src/api/` — thin wrappers returning typed data
- Custom hooks in `src/hooks/` — prefix with `use`
- Types in `src/types/` for shared interfaces
- No prop drilling > 2 levels. Use context or composition.

### Formatting & Linting
- Prettier (config: `webapp/.prettierrc`) — no semicolons, single quotes, 100 char width
- ESLint flat config with TypeScript + React Hooks + React Refresh
- Path alias: `@/` maps to `src/` (configured in vite.config.ts + tsconfig.app.json)

### Build & Run
```sh
cd webapp
npm run dev          # Vite dev server (port 3000, proxy /api → backend:8080)
npm run build        # TypeScript check + Vite build
npm run lint         # ESLint
npm run format       # Prettier format src/
npm run format:check # Prettier check (CI)
```

## Git Hooks (Husky)
Pre-commit hook auto-formats staged files:
- `.ts/.tsx` files → Prettier
- `.go` files → gofmt
Config: root `package.json` (lint-staged) + `.husky/pre-commit`

## Key Design Documents
- `docs/database-design.md` — Full schema (22 tables, 11 migrations), ERD, query patterns
- `plans/` — Implementation roadmap (Plans A-G, phased)

## Development Plan Status
- **Plan A Phase 01:** Migration system ✅ DONE
- **Plan A Phase 02:** Core Data Model (migrations 002-004) — NEXT
- **Current schema:** Migration 001 (libraries, media, progress)

## Important Conventions
- Vietnamese comments in plan files are intentional. Code comments in English.
- Commit messages in English.
- API responses: `{"data": ...}` for success, `{"error": "message"}` for errors.
- All timestamps in ISO 8601 format.
- File paths in database are always absolute paths.
