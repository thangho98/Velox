# Phase 07: Drop Legacy String Fields + Final Verify

Status: ⬜ Pending
Dependencies: Phase 01-06 deployed and verified

## Objective

Remove the transitional `poster_path: string` / `backdrop_path: string` etc. fields from API responses now that all clients consume `ImageResource` objects. Final cross-platform verification pass.

## Changes

### 1. Remove legacy fields

Walk each model file and drop the string-typed image fields that have an `ImageResource` replacement populated in Phase 02:

- [model/media.go](../../backend/internal/model/media.go): Media, MediaListItem
- [model/series.go](../../backend/internal/model/series.go): Series, Season, Episode, SeriesListItem
- [model/user.go](../../backend/internal/model/user.go): UserData.MediaPoster, ContinueWatchingItem, NextUpItem
- [model/activity.go](../../backend/internal/model/activity.go): MostWatchedItem

**Keep as-is (profile scope deferred):**
- [model/user.go](../../backend/internal/model/user.go) `User.AvatarPath` — no `Profile` ImageResource in this refactor; FE continues using the legacy string path for avatars.
- [handler/auth.go](../../backend/internal/handler/auth.go) `userInfo.ProfilePath` — same reason. Contract unchanged for `/api/auth/me`.

Revisit these when the profile kind is added in a follow-up iteration.

Repository scan destinations:
- Scanners still need to ingest the raw string from DB. Add a transient `rawPosterPath string` field on the struct OR scan into a local variable inside `scanMedia` and build `ImageResource` via the service helper before returning.
- Cleanest: keep raw string as **unexported** field `posterPath` + exported method `Poster() *ImageResource` that lazy-computes. But Go's embedding with JSON tags makes this ugly — simpler: leave `PosterPath` as an exported typed field (kept from Phase 02) but drop its JSON tag (`json:"-"`), and add `Poster *ImageResource \`json:"poster"\``.

Final shape:
```go
type Media struct {
    // ... other fields ...
    PosterPath   PosterPath   `json:"-"`   // raw, internal
    BackdropPath BackdropPath `json:"-"`
    LogoPath     LogoPath     `json:"-"`
    ThumbPath    BackdropPath `json:"-"`

    Poster   *ImageResource `json:"poster"`
    Backdrop *ImageResource `json:"backdrop"`
    Logo     *ImageResource `json:"logo"`
    Thumb    *ImageResource `json:"thumb"`
}
```

### 2. Drop type-specific FE helpers (keep base URL resolver)

[packages/shared/lib/image.ts](../../packages/shared/lib/image.ts):
- **Remove:** `tmdbImage`, `mediaImage`, `seriesImage` — redundant now that `ImageResource.url` is authoritative.
- **Keep:** `resolveImageUrl(path)` — still needed for `User.AvatarPath` (profile kind deferred, so avatars remain legacy strings).

```bash
grep -rn "tmdbImage\|mediaImage\|seriesImage" webapp/src
# Expect: no results after Phase 05 migration.

grep -rn "resolveImageUrl" webapp/src
# Expect: only avatar/user-profile call sites. All others must use ImageResource.
```

Android [ImageUrlResolver.kt](../../android/app/src/main/java/com/velox/app/data/util/ImageUrlResolver.kt):
- **Keep** — still serves avatar paths (`user.avatar_path`) until profile kind is added.
- Scope its usage to avatar only: grep for any post-Phase-06 usage with non-avatar paths and migrate those to `ImageResource.url`.

### 3. TypeScript type tightening

**Shared package is source of truth** ([packages/shared/types/](../../packages/shared/types/)) — remove `poster_path`, `backdrop_path`, `logo_path`, `thumb_path`, `still_path`, `series_poster`, `media_poster` string fields from all interfaces that got an `ImageResource` replacement. Callers referencing these get compile errors; fix by using the `ImageResource` field. Webapp picks up the change transparently via the re-export in [webapp/src/types/api.ts](../../webapp/src/types/api.ts).

**Keep `avatar_path` / `profile_path`** on User interfaces — profile kind deferred.

Same for Android DTO classes (remove the corresponding Kotlin fields), except User.avatarPath / userInfo.profilePath.

### 4. Legacy route cleanup

If a shim for the old `GET /api/images/tmdb/{size}/{path...}` route was kept during Phase 01, remove it now.

## Verification

### Automated

```bash
# Backend
cd backend
go build ./... && go vet ./...
go test ./internal/... -count=1
golangci-lint run

# Webapp
cd ../webapp
npx tsc --noEmit
npm run lint

# Android
cd ../android
./gradlew :app:compileDebugKotlin
./gradlew :app:testDebugUnitTest
./gradlew lint
```

### Manual acceptance checklist

- [ ] `GET /api/media/{id}` — response has `poster` object, NO `poster_path` string field.
- [ ] `GET /api/media` (list) — every item has ImageResource objects, correct srcset.
- [ ] `GET /api/series/{id}`, `/api/continue-watching`, `/api/next-up` — same.
- [ ] `/api/images/tmdb/backdrop/abc.jpg?width=500` → 200 with `w780` variant.
- [ ] Webapp: browse all pages — cards render with blurhash fade-in. No console errors. CLS = 0 in Lighthouse.
- [ ] Android: launch app → home/browse/detail/player — all image surfaces working. Notification + lock screen artwork still renders.
- [ ] DB integrity: `SELECT COUNT(*) FROM image_metadata` reasonable number, no orphan rows after media deletion (if cascade intended).

### Performance spot-check

- Open Chrome DevTools → Lighthouse → Mobile preset → run on HomePage.
- LCP should improve vs baseline (blurhash eliminates "pop-in" delay perception).
- Network: verify correct srcset variants fetched (not always `original`).

## Rollback

Since legacy fields were dropped in this phase specifically, revert only this commit to restore them. Previous phases stay intact.

## Done Criteria

- All automated checks green across backend/webapp/Android.
- Manual checklist complete.
- Memory note added: "Shape B shipped. Image paths now served as ImageResource objects with srcset + blurhash. Legacy string fields removed 2026-04-XX."
- Plan.md progress table all ✅.

---
End of plan.
