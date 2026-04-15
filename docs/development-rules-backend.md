# Velox Backend Development Rules (Go)

> Living document — updated as patterns evolve. For project overview and build commands, see [CLAUDE.md](../CLAUDE.md). See also: [webapp rules](development-rules-webapp.md), [mobile rules](development-rules-mobile.md).

---

## General Rules

### File Size
- Max ~500 lines per file. Split by logical concern when approaching limit.
- **Exception:** migration registry, generated code.

### Naming
- Code comments and variable names in **English**.
- Plan/spec files may contain **Vietnamese**.
- Commit messages in **English**, prefixed: `Add(scope)`, `Fix(scope)`, `Enhance(scope)`, `Refactor(scope)`, `Chore:`.

### No Premature Abstractions
- Don't create helpers/wrappers for one-time operations.
- Three similar lines > one premature abstraction.
- Only abstract when the pattern repeats 3+ times.

---

## Layer Architecture

```
Handler → Service → Repository → Model
   ↓          ↓          ↓
 HTTP      Logic       SQL       Structs
```

| Layer | Responsibility | Imports Allowed |
|-------|---------------|-----------------|
| **Handler** | Parse request → call service → respond JSON | `service`, `model`, `repository` (for types only) |
| **Service** | Business logic, orchestration | `repository`, `model`, `pkg/*` |
| **Repository** | Pure SQL queries, one per table/aggregate | `model`, `database/sql` |
| **Model** | Plain structs with `json` tags | Nothing from `internal/` |

### Handler Rules
- **No `database/sql` imports.** Never use `sql.ErrNoRows` in handlers.
- Use `repository.ErrNotFound` or `service.ErrNotFound` for not-found checks.
- No business logic — only parse → call → respond.
- Standalone functions (like `Health`, `FSBrowse`) are OK for stateless endpoints.

### Repository Rules
- Wrap `sql.ErrNoRows` → `repository.ErrNotFound` in `GetByID`-style methods.
- One repo per table/aggregate. One file per repo.
- Use `DBTX` interface for transaction support (`WithTx` pattern).
- `RowsAffected` check + `ErrNotFound` on update/delete operations.

### Service Rules
- Wrap repo errors into `service.ErrNotFound` where appropriate.
- Use `context.Context` as first parameter.
- `Set*` methods for optional dependencies (notification, transcoder, etc.).

---

## File Organization (Split Pattern)

When a file exceeds ~500 lines, split by concern into multiple files in the **same package**:

```
# Repository example:
repository/
  media.go           # MediaRepo: struct, constructor, CRUD
  media_query.go     # MediaRepo: List, Search, ListFiltered
  media_file.go      # MediaFileRepo: struct, constructor, CRUD
  media_browse.go    # MediaFileRepo: BrowseFolders + types

# Service example:
service/
  pretranscode.go           # struct, constructor, scheduler loop
  pretranscode_worker.go    # job processing, FFmpeg encoding
  pretranscode_admin.go     # enqueue, status, cleanup, profiles

# Transcoder example:
transcoder/
  transcoder.go             # struct, slots, paths, cleanup
  transcoder_hls.go         # HLS generation
  transcoder_abr.go         # ABR variant generation
  transcoder_encoding.go    # HW accel, video encoding args
```

**Rules for splitting:**
- Struct definition + constructor stay in the main file.
- Shared state (mutex, atomic, constants) stays in the main file.
- Split files use receiver methods — same struct, different file.
- Each file gets only the imports it needs.
- `go build ./...` + `go vet ./...` after every split.

---

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Error handling: always check errors, wrap with context (`fmt.Errorf("doing X: %w", err)`).
- Use `context.Context` as first parameter in service/repository methods.
- Name receivers with 1-2 letter abbreviations (`func (s *MediaService)`, `func (r *MediaRepo)`).
- Package names: singular, lowercase (`handler`, `service`, `repository`, `model`).
- No `utils` or `helpers` packages. Put functions where they belong.

### Error Handling
- Always wrap errors with context: `fmt.Errorf("doing X: %w", err)`.
- Check `RowsAffected` for update/delete, return `ErrNotFound` if 0.
- Never swallow errors silently (except intentional fire-and-forget with `_ =`).

---

## Testing

- Table-driven tests: `tests := []struct{ name string; ... }{ ... }`.
- Use `t.Run(tt.name, ...)` for subtests.
- Test files next to source: `foo.go` → `foo_test.go`.
- In-memory SQLite for DB tests: `sql.Open("sqlite3", ":memory:?_foreign_keys=on")`.
- Pass `nil` for optional dependencies in tests (e.g., `*websocket.Hub`).
- Run: `cd backend && make test`.

---

## Linting

- `go vet` + `golangci-lint` (config: `backend/.golangci.yml`).
- Enabled linters: errcheck, staticcheck, sqlclosecheck, misspell, bodyclose.

---

## Build & Run

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

**Verification before commit:**
```sh
cd backend && go build ./... && go vet ./...
```

---

## Git & CI

### Commit Convention
```
Add(scope): new feature
Fix(scope): bug fix
Enhance(scope): improve existing feature
Refactor(scope): structural change, no behavior change
Chore: tooling, deps, config
```

### Pre-commit Hooks (Husky + lint-staged)
- `.go` → gofmt auto-format
- Runs automatically on `git commit`

---

## Anti-Patterns

| ❌ Don't | ✅ Do |
|----------|-------|
| `sql.ErrNoRows` in handler | `repository.ErrNotFound` |
| Raw SQL in handler | Add method to repository |
| Business logic in handler | Move to service layer |
| `database/sql` import in handler | Use typed errors from service/repo |
| 800+ line file | Split by concern into same package |
| Bare `return err` | `fmt.Errorf("doing X: %w", err)` |
| Skip `RowsAffected` on update/delete | Check + return `ErrNotFound` if 0 |
| Swallowing errors silently | Log, wrap, or explicitly `_ =` for fire-and-forget |
| `utils` / `helpers` package | Put function where it belongs |
| Inline `CREATE TABLE` in code | Add migration to `registry.go` |
| GORM or any ORM | Raw `database/sql` or sqlc |
