# Phase 08: Verify Webapp — Full Regression Test
Status: ⬜ Pending
Dependencies: Phase 07

## Objective
Đảm bảo webapp hoạt động y hệt trước monorepo migration.

## Implementation Steps

### 1. Build verification
- [ ] `cd webapp && pnpm build` — TypeScript + Vite build clean
- [ ] `cd webapp && pnpm lint` — ESLint pass
- [ ] `cd webapp && pnpm format:check` — Prettier pass

### 2. Manual regression checklist
- [ ] **Auth:** Login / Logout / Token refresh (close tab → reopen → still logged in)
- [ ] **Browse:** Library list, movie grid, series grid
- [ ] **Media Detail:** Poster, backdrop, metadata, genres, credits
- [ ] **Series Detail:** Seasons tab, episode list, episode card
- [ ] **Playback:** Direct play + HLS fallback
- [ ] **Subtitles:** Select subtitle, dual subtitle, size/color settings
- [ ] **Quality:** Resolution selector, pre-transcode ⚡ indicator
- [ ] **Settings:** All 17 sections load (navigate via ?section=X)
- [ ] **Admin:** Activity log, tasks, webhooks
- [ ] **Home:** Continue Watching row, Next Up row
- [ ] **Favorites:** Add/remove favorites
- [ ] **Search:** Search movies + series
- [ ] **Persist:** Volume/mute survive page refresh
- [ ] **Persist:** Subtitle language preference survives refresh
- [ ] **Cinema:** Trailer backdrop (if cinema mode enabled)
- [ ] **Pre-transcode:** Controls (if pre-transcode enabled)

### 3. Commit
- [ ] Commit: `refactor: migrate webapp to @velox/shared monorepo package`
- [ ] This is a safe checkpoint — webapp fully migrated and verified

## Test Criteria
- [ ] All 16 checklist items pass
- [ ] Build + lint + format all clean
- [ ] Committed and stable

---
Next Phase: [phase-09-expo-setup.md](phase-09-expo-setup.md)
