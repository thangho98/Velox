# Plan: Responsive Images — Shape B (srcset + blurhash)

Created: 2026-04-14
Status: 🟡 Planned

## Goal

Replace flat `poster_path: "string"` JSON fields with a rich `ImageResource` object per image: type-aware URL, full srcset across TMDb size buckets, aspect ratio, dimensions, and a blurhash LQIP for progressive loading. Supports `<picture><source>` natively and enables Netflix/Instagram-style blurred placeholder transitions.

Subsumes the previous `image-type-aware-routing` plan — per-type URL routing becomes Phase 01 here.

## Target Shape (reference)

```json
{
  "poster": {
    "url": "/api/images/tmdb/poster/abc.jpg?width=500",
    "srcset": {
      "185": "/api/images/tmdb/poster/abc.jpg?width=185",
      "342": "/api/images/tmdb/poster/abc.jpg?width=342",
      "500": "/api/images/tmdb/poster/abc.jpg?width=500",
      "780": "/api/images/tmdb/poster/abc.jpg?width=780",
      "original": "/api/images/tmdb/poster/abc.jpg?width=original"
    },
    "type": "poster",
    "aspect": "2:3",
    "width": 500,
    "height": 750,
    "blurhash": "LKO2?U%2Tw=w]~RBVZRi};RPxuwH"
  }
}
```

`null` when no image. `blurhash` is `null` when not yet computed (async backfill).

## Decisions

- **Per-type URL routing:** `/api/images/tmdb/{type}/{path}?width=N`. Type segments: `poster`, `backdrop`, `still`, `logo` (+ `profile` deferred).
- **Blurhash library:** `github.com/bbrks/go-blurhash` (maintained fork). Components 4×3 — good quality/size trade-off, ~30-char hash.
- **Storage:** Separate `image_metadata(path, blurhash, width, height, computed_at)` table keyed on the raw stored path. Shared across rows reuse same TMDb poster → compute once.
- **Computation timing:**
  - Local uploads: synchronous at upload time (we already have the decoded image).
  - TMDb paths: async worker. Scanner/metadata service enqueues; worker fetches `w185` variant from TMDb (small, fast), decodes, computes, upserts.
  - Backfill: CLI subcommand `velox blurhash backfill` iterates existing media.
- **Rate limit for TMDb fetches:** 20 req/s conservative. Configurable. Backfill 1000 images ≈ 50s.
- **Response field layout (progressive rollout):**
  - Phase 01-06: keep legacy `poster_path: string` **alongside** new `poster: { ... }` fields. FE migrates incrementally.
  - Phase 07: drop legacy string fields after all clients updated.
- **FE blurhash rendering:**
  - Webapp: `react-blurhash` (~8KB gzipped).
  - Android: `io.github.Snowdh:coil-blurhash` or render via custom Compose Painter using blurhash-kotlin.

## Non-Goals

- **Profile kind (cast/crew + user avatars)** — deferred. Concretely:
  - `User.AvatarPath` and `handler/auth.go userInfo.ProfilePath` keep their current `ImagePath`/string shape through all 7 phases.
  - Phase 07 does NOT drop `avatar_path` / `profile_path` fields from responses.
  - FE continues rendering avatars via the legacy `resolveImageUrl(path)` helper (kept for exactly this reason).
  - Follow-up iteration adds `kindProfile` with its own size buckets (`w45`, `w185`, `h632`) when cast/crew surfaces are prioritized.
- Dynamic resize for local uploads at arbitrary widths — local endpoint still serves stored file regardless of `?width`.
- WebP/AVIF format negotiation — future iteration.
- CDN in front of Velox — out of scope.

## Phases

| # | Name | Status | Est |
|---|------|--------|-----|
| 01 | [Per-type URL routing + response stub](phase-01-type-routing.md) | ✅ Complete | ~2h |
| 02 | [ImageResource model + `image_metadata` table](phase-02-image-resource-model.md) | ⬜ Pending | ~3h |
| 03 | [Blurhash computation service](phase-03-blurhash-service.md) | ⬜ Pending | ~3h |
| 04 | [Ingestion hooks + backfill CLI](phase-04-ingestion-backfill.md) | ⬜ Pending | ~3h |
| 05 | [Webapp `<ResponsiveImage>` + migration](phase-05-webapp-component.md) | ✅ Complete | ~3h |
| 06 | [Android Coil blurhash + migration](phase-06-android-component.md) | ⬜ Pending | ~3h |
| 07 | [Drop legacy string fields + verify](phase-07-drop-legacy-verify.md) | ⬜ Pending | ~1h |

**Total estimate:** ~18h (1-2 solid workdays).

## Breaking Change Matrix

| Surface | Phase | Compatible? |
|---------|-------|-------------|
| New JSON fields `poster`/`backdrop`/`logo`/`thumb`/`still` | 02 | Additive |
| `?width=N` + `/type/` route | 01 | Breaking for old FE caches (minor) |
| Legacy `poster_path: string` | 07 | Removed |
| DB: `image_metadata` table added | 02 | Additive |
| FE component migration | 05, 06 | Internal |

Progressive rollout: deploy Phase 01-02 → FE continues using legacy fields while aware of new shape → Phase 05-06 migrate FE usage → Phase 07 drop legacy.

## Dependencies

New deps:
- Backend: `github.com/bbrks/go-blurhash`
- Webapp: `react-blurhash`
- Android: `io.github.Snowdh:coil-blurhash` (or alternative)

## Rollback

- DB migration 034 is reversible (drop table).
- Response shape is additive until Phase 07 — can revert FE changes any time before that.
- Phase 07 removal is a separate commit — revert just that commit to restore legacy strings.

## Quick Commands

```bash
# Backend verify after each phase
cd backend && go build ./... && go test ./internal/... -count=1

# FE verify
cd webapp && npx tsc --noEmit

# Android verify  
cd android && ./gradlew :app:compileDebugKotlin

# Run backfill (after Phase 04)
cd backend && ./bin/velox blurhash backfill
```
