# AI Prompt: Mobile-Web Alignment

> Copy prompt ben duoi vao AI coder. Moi lan chi lam 1 Priority level (P0, P1, P2...).
> Truoc khi bat dau, paste priority level can lam va noi "Lam P0" hoac "Lam P1"...

---

## PROMPT (copy tu day)

```
Ban la senior React Native developer. Nhiem vu cua ban la upgrade mobile app (Expo SDK 55) cho giong voi web app.

## DU AN

Velox la self-hosted home media server (giong Jellyfin/Emby). Co 2 app:
- **Web app** (`webapp/`): React 19, TypeScript, TailwindCSS 4, Vite — DA XONG, lam mau
- **Mobile app** (`mobile/`): Expo SDK 55, React Native 0.83, TypeScript — CAN UPGRADE

Mobile va web dung chung `@velox/shared` package (types, hooks, API client, stores).

## ARCHITECTURE RULES (BAT BUOC)

1. **Doc CLAUDE.md va docs/development-rules.md TRUOC khi code.** Day la source of truth cho conventions.

2. **Dung shared code:** Import types, hooks, API client tu `@velox/shared`. KHONG duplicate logic.
   ```typescript
   import type { MediaListItem } from '@velox/shared/types'
   import { useContinueWatching, useMediaList } from '@velox/shared/hooks'
   import { mediaImage } from '@velox/shared/lib'
   ```

3. **Styling:** Chi dung React Native `StyleSheet.create()`. KHONG dung NativeWind, styled-components, hay inline style cho static values.

4. **Color palette (Netflix dark theme):**
   - Background: `#141414` (screens), `#0a0a0a` (deep bg)
   - Cards: `#1f1f1f`, Hover/tabs: `#2a2a2a`
   - Accent red: `#e50914` (buttons, active states)
   - Text: `#fff` (primary), `#888` (muted)

5. **Naming:** Screens = `[Feature]Screen.tsx`, Components = `PascalCase.tsx`. Named exports, KHONG default export.

6. **State:** React Query (tu shared hooks) cho server data. Zustand (tu shared stores) cho auth/player. useState cho local UI state.

7. **Navigation:** React Navigation voi typed `RootStackParamList`.

8. **Tham khao web app lam mau.** Khi implement 1 feature, doc file web tuong ung de hieu logic va UI. Vi du: implement HomeScreen thi doc `webapp/src/pages/HomePage.tsx`.

## CACH LAM VIEC

1. **Doc checklist:** `plans/plan-s-expo-mobile-app/mobile-web-alignment-checklist.md` — day la master checklist.
2. **Lam tung task** trong priority level duoc yeu cau (P0, P1, P2...).
3. **Moi task:**
   - Doc file mobile hien tai (screen/component can sua)
   - Doc file web tuong ung (de hieu feature can implement)
   - Doc shared hooks/types lien quan
   - Implement
   - Dam bao TypeScript khong loi (`npx tsc --noEmit` trong mobile/)
4. **Sau moi task:** Cap nhat checklist — danh dau [x] cho task da xong.
5. **KHONG tao file moi** tru khi can thiet. Uu tien edit file hien co.
6. **KHONG them dependency moi** tru khi thuc su can va giai thich ly do.

## FILE MAP (de navigate)

### Mobile (can sua):
- `mobile/App.tsx` — Root, navigation, auth gate
- `mobile/src/screens/` — Tat ca screens
- `mobile/src/components/` — Reusable components (MediaCard, SectionHeader, HorizontalMediaRow, LibraryGrid)
- `mobile/src/stores/` — Zustand stores (auth.ts, player.ts)
- `mobile/src/platform/` — Platform adapter (mobile-adapter.ts)

### Web (lam mau, CHI DOC khong sua):
- `webapp/src/pages/` — Tat ca pages
- `webapp/src/components/` — Web components
- `webapp/src/hooks/` — Web hooks (nhieu cai da duoc move vao @velox/shared)
- `webapp/src/stores/` — Web stores

### Shared (CHI DOC, khong sua tru khi can them type/hook moi):
- `packages/shared/types/` — Shared types
- `packages/shared/hooks/` — React Query hooks
- `packages/shared/stores/` — Store factories
- `packages/shared/lib/` — Utilities (mediaImage, languages, etc.)
- `packages/shared/api/` — API client

## CURRENT STATE (Mobile da co)

✅ Dark theme #141414
✅ Stack Navigator (chua co Bottom Tabs)
✅ Continue Watching section voi overlay gradient
✅ Search + Sort trong LibraryBrowse
✅ Play button (red) o MediaDetail
✅ Expo Video player + subtitle/quality modals + PiP
✅ Progress saving 5s + resume
✅ Server URL remember (SecureStore)
✅ Pull-to-refresh

## PRIORITY LEVELS

### P0 - App Shell & Core UX
1. **Bottom Tab Navigator** — Tao 5 tabs: Home, Movies, Series, Browse, Favorites. Header: Logo "VELOX" (do, bold) ben trai, Search icon + Avatar ben phai. Tham khao web bottom nav.
2. **MediaCard upgrade** — Overlay gradient, text TREN anh (khong duoi), progress bar overlay, favorite heart, "Xm remaining", "S1E1 · Series Name" format, "Movie"/"Series" badge.
3. **HomeScreen upgrade** — Hero section "Welcome back, [Name]", 2 CTA buttons (Movies/Series), Next Up section, tach Recently Added Movies va Series, "See all →" links, loading skeletons.
4. **MediaDetailScreen upgrade** — Backdrop lon hon + gradient overlay, poster 150x225, metadata line (Year · Duration · "Ends at"), ratings badges (TMDB/IMDb), action buttons row (Play/Resume, Watched toggle, Favorite toggle), progress bar, technical specs.
5. **SeriesDetailScreen upgrade** — Play button "Resume S1E3" / "Play First Episode", full overview (ExpandableText), Status badge, Season tabs full name, episode progress bars + watched indicators.

### P1 - Browse & Discovery
6. **MoviesScreen** — Dedicated screen (Bottom Tab), header voi count, genre/year filter (bottom sheet), sort, A-Z index.
7. **SeriesScreen** — Dedicated screen (Bottom Tab), reuse filter components tu Movies.
8. **SearchScreen** — Global search, type filter chips (All/Movies/Series), ket qua tach rieng, recent searches.
9. **BrowseScreen** — Browse tab, library cards voi poster preview, folder navigation, back button chain.
10. **FavoritesScreen** — Dedicated screen (Bottom Tab), header + count, progress bars, unfavorite action.

### P2 - Player & Playback
11. **Playback speed** — Speed selector modal (0.5x-2x).
12. **Audio track selector** — Multi-audio support modal.
13. **Next Episode** — Auto-play countdown 15s, "Next Episode" button.
14. **Skip intro/credits** — Buttons khi co markers data.
15. **Gesture controls** — Double tap seek ±10s, swipe volume/brightness.

### P3 - Settings & Profile
16. **Settings tab layout** — 4 tabs (Profile, Preferences, Security, Sessions). Persist via API.
17. **Profile editing** — Edit display name, role badge, Save button.
18. **Change password** — Current + new + confirm form.
19. **Sessions management** — Active sessions list, revoke button.

### P4 - Admin & Polish
20. **Admin panel** — Basic: libraries list, trigger scan, manage users.
21. **Metadata editor** — Bottom sheet editor cho media/series (admin only).
22. **Loading skeletons** — Card, row, detail page skeletons.
23. **Toast notifications** — Thay the Alert dialogs.
24. **Animations** — Screen transitions, card press feedback.

## LUU Y QUAN TRONG

- **KHONG sua web app** (webapp/). Chi doc de tham khao.
- **KHONG sua shared package** tru khi can them type/hook moi cho mobile. Neu can them, giai thich ly do.
- **Moi lan chi lam 1 Priority level.** Xong P0 roi moi lam P1.
- **Test TypeScript** sau moi file: `cd mobile && npx tsc --noEmit`
- **Cap nhat checklist** sau moi task: `plans/plan-s-expo-mobile-app/mobile-web-alignment-checklist.md`
- **Uu tien doc code web tuong ung** truoc khi implement. Vi du khi lam HomeScreen, doc `webapp/src/pages/HomePage.tsx` de hieu layout va logic.
- **Giu file < 400 dong.** Neu file lon, split thanh components nho.
```

---

## CACH SU DUNG

### Lan 1 (P0):
```
[Paste prompt tren]

Lam P0 — App Shell & Core UX. Bat dau tu task 1 (Bottom Tab Navigator).
```

### Lan 2 (P1):
```
[Paste prompt tren]

Lam P1 — Browse & Discovery. Bat dau tu task 6 (MoviesScreen).
```

### Neu AI lam sai:
```
DUNG LAI. Doc lai:
- CLAUDE.md (project rules)
- docs/development-rules.md (coding conventions)  
- plans/plan-s-expo-mobile-app/mobile-web-alignment-checklist.md (master checklist)
- File web tuong ung (vi du webapp/src/pages/HomePage.tsx)

Roi lam lai task [so] dung theo web app.
```

### Neu AI them dependency khong can thiet:
```
KHONG them dependency moi. Dung React Native core + Expo SDK 55 built-in.
Neu can bottom sheet, dung Modal + Animated API.
Neu can icons, dung emoji hoac text (chua co icon library).
```

### Neu AI khong doc web app:
```
TRUOC KHI IMPLEMENT, doc file web tuong ung:
- HomeScreen -> doc webapp/src/pages/HomePage.tsx
- MediaDetail -> doc webapp/src/pages/MediaDetailPage.tsx  
- SeriesDetail -> doc webapp/src/pages/SeriesDetailPage.tsx
- Settings -> doc webapp/src/pages/settings/index.tsx
- Video Player -> doc webapp/src/pages/WatchPage.tsx

Copy CHINH XAC layout, logic, va data flow tu web. Chi khac UI framework (React Native thay vi HTML/CSS).
```
