# Velox Webapp Development Rules (React / TypeScript)

> Living document — updated as patterns evolve. For project overview and build commands, see [CLAUDE.md](../CLAUDE.md). See also: [backend rules](development-rules-backend.md), [mobile rules](development-rules-mobile.md).

---

## General Rules

### File Size
- Max ~400 lines per component/hook. Pages with subfolders can be larger if well-organized.

### Naming
- Code comments and variable names in **English**.
- Plan/spec files may contain **Vietnamese**.
- Commit messages in **English**, prefixed: `Add(scope)`, `Fix(scope)`, `Enhance(scope)`, `Refactor(scope)`, `Chore:`.

### No Premature Abstractions
- Don't create helpers/wrappers for one-time operations.
- Three similar lines > one premature abstraction.
- Only abstract when the pattern repeats 3+ times.

---

## Component Architecture

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

---

## Hook Patterns

### Domain-Grouped Hooks
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

### Settings Hook Factory
For repetitive get/update pairs, use the factory:

```typescript
import { createSettingsHooks } from './settings/factory'

// One line instead of 15 lines of boilerplate
export const [useTMDbSettings, useUpdateTMDbSettings] =
  createSettingsHooks<TMDbSettings, { api_key: string }>('tmdb')
```

---

## Type Organization

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

---

## React Compiler Rules

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

---

## State Management

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

---

## Shared Utilities

Reusable logic lives in `src/lib/`:

| File | Purpose |
|------|---------|
| `fetch.ts` | API client with auth, refresh, FormData |
| `image.ts` | TMDb image URL builder |
| `capabilities.ts` | Browser/device detection |
| `languages.ts` | Language names, normalization, subtitle helpers |

---

## Styling

- **TailwindCSS 4 only.** No CSS modules, styled-components, or inline `style` objects (except dynamic values).
- Use Tailwind classes for all static styling.
- Use `className` string interpolation for conditional styles.
- Shared style constants (like `inputClass`) in `shared.tsx` or co-located with components.

---

## TypeScript

- Strict mode, no `any`.
- Use `interface` for component props and API responses.
- Use `type` for unions and utility types.
- Import types with `import type { ... }` when possible.

---

## Formatting & Linting

- Prettier (config: `webapp/.prettierrc`) — no semicolons, single quotes, 100 char width.
- ESLint flat config with TypeScript + React Hooks + React Refresh.
- Path alias: `@/` maps to `src/` (configured in vite.config.ts + tsconfig.app.json).

---

## Build & Run

```sh
cd webapp
npm run dev          # Vite dev server (port 3000, proxy /api → backend:8080)
npm run build        # TypeScript check + Vite build
npm run lint         # ESLint
npm run format       # Prettier format src/
npm run format:check # Prettier check (CI)
```

**Verification before commit:**
```sh
cd webapp && npx tsc --noEmit && npm run lint
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
- `.ts/.tsx` → Prettier auto-format
- Runs automatically on `git commit`

---

## Anti-Patterns

| ❌ Don't | ✅ Do |
|----------|-------|
| Manual `useMemo`/`useCallback` | Let React Compiler handle it |
| `setState` directly in useEffect | Wrap in `queueMicrotask()` |
| Read `ref.current` during render | Mirror to state via useEffect |
| 12 identical hook pairs | Use factory pattern |
| All types in one file | Split by domain, barrel re-export |
| Prop drilling > 2 levels | Context or composition |
| `any` type | Proper interface/type |
| CSS modules / styled-components | TailwindCSS classes |
| Inline `style={{}}` for static values | Tailwind classes |
| Default export for reusable components | Named export (only pages use default) |
| Class components | Functional components only |
