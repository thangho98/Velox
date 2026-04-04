# API Contract Checklist - Kotlin Native Port

Created: 2026-04-04

This file is the contract checklist for porting the current TS client into Kotlin native. Source of truth remains the existing TS types/hooks plus backend behavior.

---

## 1. Auth / Session

| Contract | Endpoint | Source | Native Owner | Phase |
|---|---|---|---|---|
| Login | `POST /auth/login` | `packages/shared/hooks/auth.ts` | `AuthRepository` | 03 |
| Refresh | `POST /auth/refresh` | `packages/shared/hooks/auth.ts` | `AuthRepository` | 03 |
| Logout | `POST /auth/logout` | `packages/shared/hooks/auth.ts` | `AuthRepository` | 03 |
| Me | `GET /auth/me` | `packages/shared/hooks/auth.ts` | `SessionRepository` | 03 |
| Change password | `POST /auth/change-password` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08 |
| Sessions list | `GET /auth/sessions` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08/later |
| Revoke session | `DELETE /auth/sessions/{id}` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08/later |
| Profile | `GET /profile` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08 |
| Update profile | `PATCH /profile` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08 |
| Preferences | `GET /profile/preferences` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08 |
| Update preferences | `PUT /profile/preferences` | `packages/shared/hooks/auth.ts` | `ProfileRepository` | 08 |

---

## 2. Libraries / Browse / Search

| Contract | Endpoint | Source | Native Owner | Phase |
|---|---|---|---|---|
| Libraries list | `GET /libraries` | `packages/shared/hooks/media/useLibrary.ts` | `LibraryRepository` | 04 |
| Browse folders | `GET /browse?...` | `packages/shared/hooks/media/useGenres.ts` (`useFolderBrowse`) | `LibraryRepository` | 04 |
| Media list | `GET /media?...` | `packages/shared/hooks/media/useMediaQuery.ts` | `MediaRepository` | 04 |
| Media detail | `GET /media/{id}` | `packages/shared/hooks/media/useMediaQuery.ts` | `MediaRepository` | 04 |
| Media with files | `GET /media/{id}/files` | `packages/shared/hooks/media/useMediaQuery.ts` | `MediaRepository` | 04 |
| Media genres | `GET /media/{id}/genres` | `packages/shared/hooks/media/useGenres.ts` | `MediaRepository` | 04 |
| Media credits | `GET /media/{id}/credits` | `packages/shared/hooks/media/useGenres.ts` | `MediaRepository` | 04 |
| Genres | `GET /genres?...` | `packages/shared/hooks/media/useGenres.ts` | `SearchRepository` / `LibraryRepository` | 04 |
| Unified search | `GET /search?q=...` | `packages/shared/hooks/media/useGenres.ts` | `SearchRepository` | 04 |

---

## 3. Series / Seasons / Episodes

| Contract | Endpoint | Source | Native Owner | Phase |
|---|---|---|---|---|
| Series list | `GET /series?...` | `packages/shared/hooks/media/useSeries.ts` | `SeriesRepository` | 04 |
| Series detail | `GET /series/{id}` | `packages/shared/hooks/media/useSeries.ts` | `SeriesRepository` | 04 |
| Series search | `GET /series/search?...` | `packages/shared/hooks/media/useSeries.ts` | `SeriesRepository` | 04 |
| Series genres | `GET /series/{id}/genres` | `packages/shared/hooks/media/useGenres.ts` | `SeriesRepository` | 04 |
| Series credits | `GET /series/{id}/credits` | `packages/shared/hooks/media/useGenres.ts` | `SeriesRepository` | 04 |
| Seasons | `GET /series/{id}/seasons` | `packages/shared/hooks/media/useSeries.ts` | `SeriesRepository` | 04 |
| Episodes in season | `GET /series/{id}/seasons/{seasonId}/episodes` | `packages/shared/hooks/media/useSeries.ts` | `SeriesRepository` | 04 |
| Episode detail | `GET /episodes/{id}` | `packages/shared/hooks/media/useSeries.ts` | `SeriesRepository` | 04/05 |

---

## 4. User Data / Favorites / Progress

| Contract | Endpoint | Source | Native Owner | Phase |
|---|---|---|---|---|
| Progress | `GET /profile/progress/{mediaId}` | `packages/shared/hooks/media/useProgress.ts` | `PlaybackRepository` | 05 |
| Update progress | `PUT /profile/progress/{mediaId}` | `packages/shared/hooks/media/useProgress.ts` | `PlaybackRepository` | 05 |
| Favorites list | `GET /profile/favorites?...` | `packages/shared/hooks/media/useProgress.ts` | `FavoritesRepository` | 08 |
| Toggle favorite | `POST /profile/favorites/{mediaId}` | `packages/shared/hooks/media/useProgress.ts` | `FavoritesRepository` | 08 |
| Recently watched | `GET /profile/recently-watched?...` | `packages/shared/hooks/media/useProgress.ts` | `HomeRepository` / `FavoritesRepository` | 04/08 |

---

## 5. Playback

| Contract | Endpoint | Source | Native Owner | Phase |
|---|---|---|---|---|
| Playback info | `POST /playback/{mediaId}/info` | `packages/shared/hooks/media/usePlayback.ts` | `PlaybackRepository` | 05 |
| Stream URL | `POST /stream/{mediaId}/url` | `packages/shared/hooks/media/usePlayback.ts`, `mobile/src/hooks/useChromecast.ts` | `PlaybackRepository` / `CastRepository` | 05/06 |
| Playback types | `PlaybackInfo`, `PlaybackInfoRequest`, `StreamUrls`, `PlaybackAudioTrack`, `PlaybackSubtitleTrack`, `QualityOption` | `packages/shared/types/playback.ts` | `core:model` | 05 |

Playback-specific notes:
- `selected_audio_track` is the request field name, not `selected_audio_track_id`
- `selected_subtitle_id` is used when selecting a specific subtitle
- client should respect `available_qualities`
- client should not assume transcode implies low quality

---

## 6. Subtitle Ops (Advanced / Deferred Candidate)

| Contract | Endpoint/Behavior | Source | Native Owner | Phase |
|---|---|---|---|---|
| Subtitle search | see `useSubtitleSearch` | `packages/shared/hooks/media/useSubtitleOps.ts` | evaluate later | 06/08 |
| Subtitle download | see `useDownloadSubtitle` | `packages/shared/hooks/media/useSubtitleOps.ts` | evaluate later | 06/08 |
| Subtitle translate | see `useTranslateSubtitle` | `packages/shared/hooks/media/useSubtitleOps.ts` | evaluate later | 06/08 |

These should not block the first playback-stable native beta.

---

## 7. Settings (User/Admin)

| Contract | Endpoint | Source | Native Owner | Phase |
|---|---|---|---|---|
| Playback settings | `/admin/settings/playback` via factory | `packages/shared/hooks/settings/usePlaybackSettings.ts` | likely web-admin only or limited native surface | 08 |
| Cinema settings | `GET/PUT /admin/settings/cinema` | `packages/shared/hooks/settings/usePlaybackSettings.ts` | probably defer | later |
| Metadata/subtitle/pretranscode admin settings | `/admin/settings/{slug}` | `packages/shared/hooks/settings/*` | likely web-admin only | later |

Rule:
- user-facing playback preferences live locally in DataStore
- backend admin settings should only be pulled into native if they are truly needed for end users

---

## 8. Admin / Ops (Do Not Block Native MVP)

| Contract | Source | Recommendation |
|---|---|---|
| filesystem browse | `packages/shared/hooks/media/useLibrary.ts` (`useFsBrowse`) | keep web-admin |
| create/delete/scan library | `packages/shared/hooks/media/useLibrary.ts` | keep web-admin |
| metadata editor / image upload | `packages/shared/hooks/media/useMetadataOps.ts` | keep web-admin |
| cinema intro upload | `packages/shared/hooks/settings/usePlaybackSettings.ts` | keep web-admin |

---

## 9. Porting Rules

1. Port exact field names before introducing Kotlin-side aliases.
2. Keep DTOs close to current backend contract.
3. Separate DTOs from richer UI/domain models when needed.
4. Validate each contract with real backend responses before declaring parity.
