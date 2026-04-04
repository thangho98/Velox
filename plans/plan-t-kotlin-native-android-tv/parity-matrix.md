# Parity Matrix - RN Mobile -> Kotlin Native Android

Created: 2026-04-04

| Area | Current RN Source | Native Target | Phase | Ship | Status |
|---|---|---|---|---|---|
| Setup/bootstrap | `mobile/App.tsx` auth bootstrap | `app` + `feature:auth` | 03 | 1 | planned |
| Login | `mobile/src/screens/LoginScreen.tsx` | `feature:auth` | 03 | 1 | planned |
| Home | `mobile/src/screens/HomeScreen.tsx` | `feature:home` | 04 | 1 | planned |
| Movies listing | `mobile/src/screens/MoviesScreen.tsx` | `feature:home` / `feature:browse` | 04 | 1 | planned |
| Series listing | `mobile/src/screens/SeriesScreen.tsx` | `feature:home` / `feature:browse` | 04 | 1 | planned |
| Browse | `mobile/src/screens/BrowseScreen.tsx` | `feature:browse` | 04 | 1 | planned |
| Library folder traversal | `mobile/src/screens/LibraryBrowseScreen.tsx` | `feature:browse` | 04 | 1 | planned |
| Search | `mobile/src/screens/SearchScreen.tsx` | `feature:search` | 04 | 1 | planned |
| Media detail | `mobile/src/screens/MediaDetailScreen.tsx` | `feature:detail` | 04 | 1 | planned |
| Series detail | `mobile/src/screens/SeriesDetailScreen.tsx` | `feature:detail` | 04 | 1 | planned |
| Playback core | `mobile/src/screens/VideoPlayerScreen.tsx` | `feature:player` + `core:player` | 05 | 1 | planned |
| Progress sync | `packages/shared/hooks/media/useProgress.ts` | `core:network` + `feature:player` | 05 | 1 | planned |
| Subtitle/audio/quality | `VideoPlayerScreen.tsx`, player store | `feature:player` | 06 | 1/2 | planned |
| Skip intro/credits | playback contract + player UI | `feature:player` | 06 | 2 | planned |
| Chromecast | `mobile/src/hooks/useChromecast.ts` | `feature:cast` | 06 | 3 | planned |
| Android TV shell | `mobile/src/lib/tv.ts` + screen behavior | `feature:tv` | 07 | 2 | planned |
| TV player controls | RN player TV behavior | `feature:tv` + `feature:player` | 07 | 2 | planned |
| Favorites | `mobile/src/screens/FavoritesScreen.tsx` | `feature:favorites` | 08 | 3 | planned |
| Settings | `mobile/src/screens/SettingsScreen.tsx` | `feature:settings` | 08 | 3 | planned |
| Profile | `mobile/src/screens/ProfileScreen.tsx` | `feature:profile` | 08 | 3 | planned |
| Admin dashboard | `mobile/src/screens/AdminScreen.tsx` | stay web-admin for now | later | later | deferred |
| Subtitle translate/search | `packages/shared/hooks/media/useSubtitleOps.ts` + player UI | decide after playback stable | 06/08 | later | deferred |
| Metadata editor | admin/media ops | stay web-admin for now | later | later | deferred |

## Practical Cut Line

Native app is considered good enough for Android cutover when these rows are complete:
- Login
- Home
- Browse
- Search
- Media detail
- Series detail
- Playback core
- Progress sync
- Subtitle/audio/quality
- Android TV shell
- TV player controls

Everything else is important, but should not block the first serious beta if the core viewing experience is already strong.
