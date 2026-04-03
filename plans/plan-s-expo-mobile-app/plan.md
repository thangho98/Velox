# Plan S: Expo Mobile App (Monorepo)

Created: 2026-04-01
Status: ⬜ Pending

## Overview
Android mobile app cho Velox media server. Dùng Expo + expo-video (ExoPlayer) để Direct Play nhiều format hơn browser. Monorepo architecture với shared package để reuse types, hooks, stores giữa web app và mobile app.

Android TV dùng web app hiện tại.

## Architecture

```
velox/
├── backend/                       # Go backend (unchanged)
├── packages/
│   └── shared/                    # @velox/shared — reuse giữa web + mobile
│       ├── types/                 # API types (100% reuse)
│       ├── api/                   # API client + query hooks (90% reuse)
│       ├── stores/                # Zustand store factories (80% reuse)
│       └── lib/                   # Utils: languages, image (100% reuse)
├── webapp/                        # Web app — import from @velox/shared
│   └── src/
│       └── platform/              # Web-specific: localStorage, browser fetch
├── mobile/                        # Expo app — import from @velox/shared
│   └── src/
│       └── platform/              # Mobile-specific: SecureStore, RN fetch
├── package.json                   # pnpm workspace root
└── pnpm-workspace.yaml
```

```
┌──────────────┐     REST API      ┌──────────────┐
│  Velox       │ ◄───────────────► │  Web App     │
│  Backend     │                   │  (React)     │
│  (Go+SQLite) │                   └──────┬───────┘
│              │                          │
│              │     REST API      ┌──────┴───────┐
│              │ ◄───────────────► │  Mobile App  │
└──────────────┘                   │  (Expo)      │
                                   └──────────────┘
                                          │
                                   @velox/shared
                                   (types, hooks,
                                    stores, utils)
```

## Tech Stack

| Layer | Web App | Mobile App | Shared |
|-------|---------|-----------|--------|
| Framework | React 19 + Vite | Expo SDK 53+ | — |
| Language | TypeScript | TypeScript | TypeScript |
| Navigation | React Router | Expo Router | — |
| Video | HTML5 `<video>` + HLS.js | expo-video (ExoPlayer) | — |
| Server State | TanStack React Query | TanStack React Query | Query hooks |
| Client State | Zustand | Zustand | Store factories |
| Styling | TailwindCSS 4 | NativeWind | — |
| Storage | localStorage | SecureStore + MMKV | Storage adapter interface |
| Types | @velox/shared | @velox/shared | ✅ Single source |
| API Client | @velox/shared | @velox/shared | ✅ Platform-agnostic |

## Reuse Analysis (từ code thực tế)

### Shareable (→ @velox/shared)

| File/Module | Lines | Reuse % | Ghi chú |
|------------|-------|---------|---------|
| `types/*.ts` (7 files) | 633 | 100% | common, auth, media, series, playback, admin — no DOM refs |
| `lib/fetch.ts` | 222 | 90% | Cần remove `localStorage.getItem('velox_device_name')` → adapter |
| `lib/languages.ts` | 110 | 100% | Pure data + functions |
| `lib/image.ts` | 49 | 100% | Pure URL builder |
| `hooks/stores/media/*.ts` (9 files) | 842 | 100% | Chỉ import `api` + types, không có DOM |
| `hooks/stores/settings/factory.ts` | 39 | 100% | Generic factory, không có DOM |
| `hooks/stores/settings/*.ts` (4 files) | 316 | 100% | Dùng factory, không có DOM |
| `hooks/stores/useAuth.ts` | 246 | 70% | Query hooks OK. `useRequireAuth`/`useTokenRefresh` dùng react-router → tách ra |
| `hooks/stores/useAdmin.ts` | 162 | 100% | Pure API hooks |
| `hooks/stores/useUsers.ts` | 88 | 100% | Pure API hooks |
| `stores/auth.ts` | 91 | 80% | Logic OK, cần factory thay `localStorage` → adapter |
| `stores/player.ts` | 173 | 80% | Logic OK, cần factory thay `localStorage` → adapter |
| **TOTAL** | **~2,971** | | |

### NOT Shareable (giữ trong webapp/)

| File | Lines | Lý do |
|------|-------|-------|
| `lib/capabilities.ts` | 241 | 100% browser APIs (MediaSource, createElement, navigator) |
| `stores/ui.ts` | 91 | Web-specific (sidebar, search modal, toasts with setTimeout) |
| `useAuth.ts` → `useRequireAuth()` | ~12 | Dùng `useNavigate` từ react-router |
| `useAuth.ts` → `useTokenRefresh()` | ~15 | Dùng `useEffect` + `setInterval` — logic OK nhưng mobile cần khác |
| All page components | ~8,476 | UI layer — mobile viết lại hoàn toàn |

## Phases

| Phase | Name | Status | Tasks |
|-------|------|--------|-------|
| **MONOREPO SETUP** | | | |
| 01 | [pnpm Workspace + Shared Package Skeleton](phase-01-pnpm-workspace.md) | ⬜ Pending | 5 |
| 02 | [Extract Types → @velox/shared](phase-02-extract-types.md) | ⬜ Pending | 6 |
| 03 | [Extract Platform-Agnostic API Client](phase-03-extract-api-client.md) | ⬜ Pending | 6 |
| 04 | [Extract Zustand Store Factories](phase-04-extract-stores.md) | ⬜ Pending | 6 |
| 05 | [Extract Shared Libs](phase-05-extract-libs.md) | ⬜ Pending | 4 |
| 06 | [Extract Query Hooks](phase-06-extract-hooks.md) | ⬜ Pending | 7 |
| **WEBAPP MIGRATION** | | | |
| 07 | [Migrate Webapp Imports](phase-07-migrate-webapp.md) | ⬜ Pending | 5 |
| 08 | [Verify Webapp — Regression Test](phase-08-verify-webapp.md) | ⬜ Pending | 3 |
| **EXPO MOBILE APP** | | | |
| 09 | [Expo Project Setup + Dev Build](phase-09-expo-setup.md) | ⬜ Pending | 8 |
| 10 | [Platform Adapters](phase-10-platform-adapters.md) | ⬜ Pending | 5 |
| 11 | [Auth Flow](phase-11-auth-flow.md) | ⬜ Pending | 4 |
| 12 | [Home Screen](phase-12-home-screen.md) | ⬜ Pending | 4 |
| 13 | [Library Browser + Media Grid](phase-13-library-browser.md) | ⬜ Pending | 4 |
| 14 | [Media Detail + Series Detail](phase-14-media-detail.md) | ⬜ Pending | 3 |
| 15 | [Video Player (ExoPlayer)](phase-15-video-player.md) | ⬜ Pending | 5 |
| 16 | [Player Controls + Subtitles + Quality](phase-16-player-controls.md) | ⬜ Pending | 8 |
| 17 | [Profile + Settings + Favorites](phase-17-profile-settings.md) | ⬜ Pending | 4 |
| 18 | [Polish + Build APK](phase-18-polish-build.md) | ⬜ Pending | 7 |

**Total:** 92 tasks across 18 phases

## Backend Changes Needed

Minimal — API already mobile-ready:

| Change | Why | Where | Effort |
|--------|-----|-------|--------|
| Accept `mkv` in container capabilities | ExoPlayer handles MKV natively | `internal/service/playback_config.go` | 1 line |
| Extended codec list | More Direct Play for ExoPlayer | Already works — backend reads client capabilities | 0 changes |

## Future (Post-MVP)

- **Offline download** — expo-file-system + background download
- **Chromecast** — Cast SDK integration
- **Home screen widget** — Continue Watching
- **iOS** — Same Expo codebase, minimal changes (AVPlayer via expo-video)
- **Android TV** — Evaluate after mobile launch (Leanback or enhanced web app)
