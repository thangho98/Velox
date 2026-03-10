# Phase 01: Migration System
Status: ✅ Complete
Plan: A - Core Domain & Ingestion

## Mục tiêu
Thay thế CREATE IF NOT EXISTS bằng versioned migration system có rollback.

## Tasks

### 1. Migration Runner
- [x] Tạo `internal/database/migrate/migrate.go`
- [x] Table `schema_migrations` (version INT, applied_at DATETIME, name TEXT)
- [x] Func `Up()` - apply pending migrations in order (transactional per migration)
- [x] Func `Rollback()` - rollback last applied migration
- [x] Func `Status()` - show all migrations + applied state
- [x] Log mỗi migration applied

### 2. Migration File Convention
- [x] Mỗi migration = Go func trong `registry.go`
- [x] Convention: version int + name string, registered in `All()`
- [x] Struct `Migration { Version, Name, Up(tx), Down(tx) }`
- [x] Migrations auto-sorted by version

### 3. Initial Migration (001)
- [x] `001_initial_schema` trong `registry.go`
- [x] Tables: `libraries`, `media`, `progress`
- [x] Down: drop all tables
- [x] Baseline cho DB hiện tại

### 4. Refactor database.go
- [x] Xóa inline CREATE TABLE statements
- [x] `database.Migrate(db)` → gọi migration runner
- [x] `database.MigrateRollback(db)` + `database.MigrateStatus(db)` exposed
- [x] `database.Open()` giữ nguyên

### 5. CLI: Migration Commands
- [x] `velox migrate up` - apply all pending
- [x] `velox migrate status` - show applied migrations (table format)
- [x] `velox migrate rollback` - rollback last migration
- [x] Server startup auto-migrates
- [x] Makefile: `make migrate`, `make migrate-status`, `make migrate-rollback`

### 6. Tests
- [x] `migrate_test.go` - 10 test cases:
  - Fresh DB apply all
  - Idempotent (double-up)
  - Incremental (add migrations later)
  - Rollback + verify table dropped
  - Rollback empty DB
  - Status shows applied/pending
  - Transaction rollback on error (bad SQL)
  - Migration ordering (out-of-order input)
  - Real migrations: fresh DB
  - Real migrations: rollback

## Files Created/Modified
- `internal/database/migrate/migrate.go` - NEW (Runner, Up, Rollback, Status)
- `internal/database/migrate/registry.go` - NEW (All(), 001_initial_schema)
- `internal/database/migrate/migrate_test.go` - NEW (10 tests)
- `internal/database/database.go` - Refactored
- `cmd/server/main.go` - Added subcommands (migrate, version)
- `Makefile` - Added migrate targets

---
Next: phase-02-core-data-model.md
