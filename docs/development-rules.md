# Velox Development Rules & Conventions

> Living document — updated as patterns evolve. For project overview and build commands, see [CLAUDE.md](../CLAUDE.md).

---

## General Rules

### File Size
- **Backend:** Max ~500 lines per file. Split by logical concern when approaching limit.
- **Frontend:** Max ~400 lines per component/hook. Pages with subfolders can be larger if well-organized.
- **Exception:** Migration registry, generated code.

### Naming
- Code comments and variable names in **English**.
- Plan/spec files may contain **Vietnamese**.
- Commit messages in **English**, prefixed: `Add(scope)`, `Fix(scope)`, `Enhance(scope)`, `Refactor(scope)`, `Chore:`.

### No Premature Abstractions
- Don't create helpers/wrappers for one-time operations.
- Three similar lines > one premature abstraction.
- Only abstract when the pattern repeats 3+ times.

---

## Backend Rules (Go)

### Layer Architecture

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

### File Organization (Split Pattern)
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

### Error Handling
- Always wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Check `RowsAffected` for update/delete, return `ErrNotFound` if 0.
- Never swallow errors silently (except intentional fire-and-forget with `_ =`).

### Testing
- Table-driven tests with `t.Run`.
- Test files next to source: `foo.go` → `foo_test.go`.
- In-memory SQLite for DB tests.
- Pass `nil` for optional dependencies in tests (e.g., `*websocket.Hub`).

---

## Frontend Rules (React/TypeScript)

### Component Architecture

| Type | Location | Exports |
|------|----------|---------|
| **Page** | `src/pages/` or `src/pages/feature/index.tsx` | Default export |
| **Component** | `src/components/` | Named export |
| **Hook** | `src/hooks/` | Named export, `use` prefix |
| **Store** | `src/stores/` | Zustand store |
| **Type** | `src/types/` | Type/interface exports |
| **Utility** | `src/lib/` | Named export |

### Page Pattern (Feature Folders)
Large pages use a subfolder with `index.tsx` + `components/`:

```
src/pages/settings/
├── index.tsx                    # Page layout + routing (default export)
└── components/
    ├── shared.tsx               # Shared UI primitives (Field, SaveButton, etc.)
    ├── ProfileSection.tsx       # Individual section
    ├── MetadataSection.tsx
    └── ...
```

**Route registration** uses lazy import:
```typescript
const SettingsPage = lazy(() => import('@/pages/settings'))
```

### Hook Patterns

#### Domain-Grouped Hooks
Large hook files split into domain subfolders with barrel re-export:

```
src/hooks/stores/
├── useMedia.ts          # Barrel: export * from './media/*'
├── media/
│   ├── useLibrary.ts
│   ├── useMediaQuery.ts
│   ├── usePlayback.ts
│   └── ...
├── useSettings.ts       # Barrel: export * from './settings/*'
└── settings/
    ├── factory.ts       # createSettingsHooks factory
    ├── useMetadataSettings.ts
    └── ...
```

**Barrel re-export** preserves backward compatibility:
```typescript
// useMedia.ts — all existing imports still work
export * from './media/useLibrary'
export * from './media/useMediaQuery'
// ...
```

#### Settings Hook Factory
For repetitive get/update pairs, use the factory:

```typescript
import { createSettingsHooks } from './settings/factory'

// One line instead of 15 lines of boilerplate
export const [useTMDbSettings, useUpdateTMDbSettings] =
  createSettingsHooks<TMDbSettings, { api_key: string }>('tmdb')
```

### Type Organization
Types split by domain with barrel re-export:

```
src/types/
├── api.ts          # Barrel: export * from './*'
├── common.ts       # ApiResponse, FsBrowseResponse
├── auth.ts         # Login, User, Session
├── media.ts        # Media, Library, Genre, Search
├── series.ts       # Series, Season, Episode
├── playback.ts     # StreamUrls, PlaybackInfo, Subtitles
└── admin.ts        # ServerInfo, Webhook, Task
```

### React Compiler Rules
React Compiler is enabled — follow these rules:

- **No manual `useMemo`/`useCallback`** — Compiler handles memoization automatically.
- **No reading refs during render** — Use `useState` mirror updated via `useEffect` if needed.
- **setState in effects** — Wrap synchronous `setState` in `queueMicrotask()`:
  ```typescript
  // ❌ Bad — lint error
  useEffect(() => {
    setState(value)
  }, [dep])

  // ✅ Good
  useEffect(() => {
    queueMicrotask(() => setState(value))
  }, [dep])
  ```
- **react-refresh/only-export-components** — Files that export both components and non-components (hooks, constants) need `/* eslint-disable react-refresh/only-export-components */` at top.

### State Management

| Need | Solution |
|------|----------|
| Server data (API) | TanStack React Query |
| Global client state | Zustand (`src/stores/`) |
| Component-local state | `useState` |
| URL state | `useSearchParams` |

**Query key factories** for consistent cache management:
```typescript
export const mediaKeys = {
  all: ['media'] as const,
  list: (params) => [...mediaKeys.all, 'list', params] as const,
  detail: (id) => [...mediaKeys.all, 'detail', id] as const,
}
```

### Shared Utilities
Reusable logic lives in `src/lib/`:

| File | Purpose |
|------|---------|
| `fetch.ts` | API client with auth, refresh, FormData |
| `image.ts` | TMDb image URL builder |
| `capabilities.ts` | Browser/device detection |
| `languages.ts` | Language names, normalization, subtitle helpers |

### Styling
- **TailwindCSS 4 only.** No CSS modules, styled-components, or inline `style` objects (except dynamic values).
- Use Tailwind classes for all static styling.
- Use `className` string interpolation for conditional styles.
- Shared style constants (like `inputClass`) in `shared.tsx` or co-located with components.

### TypeScript
- Strict mode, no `any`.
- Use `interface` for component props and API responses.
- Use `type` for unions and utility types.
- Import types with `import type { ... }` when possible.

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
- `.ts/.tsx` → Prettier auto-format
- `.go` → gofmt auto-format
- Runs automatically on `git commit`

### Verification Before Commit
```bash
# Backend
cd backend && go build ./... && go vet ./...

# Frontend
cd webapp && npx tsc --noEmit && npm run lint
```

---

## Anti-Patterns (Don't Do This)

| ❌ Don't | ✅ Do |
|----------|-------|
| `sql.ErrNoRows` in handler | `repository.ErrNotFound` |
| Raw SQL in handler | Add method to repository |
| Business logic in handler | Move to service layer |
| `database/sql` import in handler | Use typed errors from service/repo |
| 800+ line file | Split by concern into same package |
| Manual `useMemo`/`useCallback` | Let React Compiler handle it |
| `setState` directly in useEffect | Wrap in `queueMicrotask()` |
| Read `ref.current` during render | Mirror to state via useEffect |
| 12 identical hook pairs | Use factory pattern |
| All types in one file | Split by domain, barrel re-export |
| Prop drilling > 2 levels | Context or composition |
| `any` type | Proper interface/type |
