# Velox - Home Media Server

## Project Overview
Velox is a self-hosted home media server (like Jellyfin/Emby but lighter).
- **Backend:** Go 1.26 + stdlib `net/http` (Go 1.22+ routing) + SQLite (WAL mode)
- **Frontend:** React 19 + TypeScript + Vite 8 + TailwindCSS 4 + React Compiler
- **Transcoding:** FFmpeg 8.0 / FFprobe
- **Android App:** Kotlin + Jetpack Compose + Dagger Hilt + Media3 ExoPlayer

## Development Rules

**⚠️ Before writing code in any part of the repo, read the matching rules file:**

- [docs/development-rules-backend.md](docs/development-rules-backend.md) — Go (Handler/Service/Repository/Model, error handling, split pattern, testing)
- [docs/development-rules-webapp.md](docs/development-rules-webapp.md) — React 19 + Compiler, TailwindCSS 4, React Query, Zustand
- [docs/development-rules-mobile.md](docs/development-rules-mobile.md) — Jetpack Compose + Hilt + Media3, Clean Architecture + MVVM
- [docs/development-rules.md](docs/development-rules.md) — Index + shared conventions (commit, naming, anti-abstractions)

These files are the source of truth for conventions, patterns, and anti-patterns. The sections below in this file only cover project-wide context (structure, architecture decisions, build commands).

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

android/               # Native Android application (Kotlin + Jetpack Compose)
  app/src/main/java/com/velox/app/
    data/              # API clients, repositories, DTOs, DataStore
    domain/            # Pure Kotlin models, repository interfaces, use cases
    presentation/      # Compose UI, ViewModels, Navigation, Cast
    di/                # Hilt modules (Network, Repository)
    ui/theme/          # Netflix-inspired Compose theme tokens
    utils/             # Extension functions, helpers
```

## Architecture Decisions
- **Database:** SQLite only. WAL mode, `MaxOpenConns(1)`, `_foreign_keys=on`
- **Routing:** Go stdlib `net/http` with Go 1.22+ patterns (`GET /api/foo/{id}`)
- **No ORM:** Use sqlc (generated from SQL) or raw `database/sql`. Never GORM.
- **Migrations:** All schema changes via `internal/database/migrate/registry.go`. Never inline CREATE TABLE.
- **Auth:** JWT (short-lived 15min access + 7-day refresh). bcrypt cost 12.
- **Stream URLs (mobile/external players):** Jellyfin-style `api_key` (32-char hex, 2h) via `POST /api/stream/{id}/url`.
- **Playback:** Direct Play first. HLS/transcode only when codec/container/audio incompatible.
- **Metadata:** TMDb as primary provider. NFO/local files override TMDb.
- **Android architecture:** Clean Architecture (data/domain/presentation) + MVVM with Hilt.

## Build & Run Quick Reference

```sh
# Backend
cd backend
make dev             # go run ./cmd/server
make test            # go test ./... -v -count=1
make lint            # go vet + golangci-lint

# Webapp
cd webapp
npm run dev          # Vite dev server (port 3000, proxy /api → backend:8080)
npm run build        # TypeScript check + Vite build
npm run lint         # ESLint

# Android
cd android
./gradlew build          # Compile + lint + unit tests
./gradlew installDebug   # Install debug APK on connected device
./gradlew test           # Unit tests
```

Full build/lint/test commands for each platform live in the matching rules file.

## Git Hooks (Husky)
Pre-commit hook auto-formats staged files:
- `.ts/.tsx` files → Prettier
- `.go` files → gofmt
- `.kt` files → ktlint (if configured)

Config: root `package.json` (lint-staged) + `.husky/pre-commit`

## Key Design Documents
- [docs/development-rules.md](docs/development-rules.md) — Rules index (backend / webapp / mobile)
- [docs/database-design.md](docs/database-design.md) — Full schema, ERD, query patterns
- `plans/` — Implementation roadmap (Plans A-S)

## Important Conventions
- Vietnamese comments in plan files are intentional. Code comments in English.
- Commit messages in English, prefixed: `Add(scope)`, `Fix(scope)`, `Enhance(scope)`, `Refactor(scope)`, `Chore:`.
- API responses: `{"data": ...}` for success, `{"error": "message"}` for errors.
- All timestamps in ISO 8601 format.
- File paths in database are always absolute paths.
