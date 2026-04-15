# Phase 02: ImageResource Model + image_metadata Table

Status: ⬜ Pending
Dependencies: Phase 01 (per-type routing works)

## Objective

Define the JSON-facing `ImageResource` struct that carries srcset + dimensions + blurhash. Introduce `image_metadata` DB table as the source of blurhash/dimensions. Models expose both legacy string path and new rich object (progressive rollout).

## Changes

### 1. DB migration 034

[migrate/034_image_metadata.go](../../backend/internal/database/migrate/034_image_metadata.go):

```sql
CREATE TABLE image_metadata (
    path        TEXT PRIMARY KEY,
    blurhash    TEXT NOT NULL,
    width       INTEGER NOT NULL,
    height      INTEGER NOT NULL,
    computed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Register in [registry.go](../../backend/internal/database/migrate/registry.go).

### 2. Repository

[repository/image_metadata.go](../../backend/internal/repository/image_metadata.go) — new:
```go
type ImageMetadataRepo struct { db DBTX }

func (r *ImageMetadataRepo) Get(ctx context.Context, path string) (*model.ImageMetadata, error)
func (r *ImageMetadataRepo) Upsert(ctx context.Context, m *model.ImageMetadata) error
func (r *ImageMetadataRepo) GetBatch(ctx context.Context, paths []string) (map[string]*model.ImageMetadata, error)
```

`GetBatch` is the hot-path for list endpoints — single query resolves N paths.

### 3. ImageResource type

[model/imageresource.go](../../backend/internal/model/imageresource.go) — new:

```go
type ImageResource struct {
    URL      string            `json:"url"`
    Srcset   map[string]string `json:"srcset"`
    Type     string            `json:"type"`              // poster|backdrop|still|logo
    Aspect   string            `json:"aspect"`            // "2:3" | "16:9" | "variable"
    Width    int               `json:"width,omitempty"`
    Height   int               `json:"height,omitempty"`
    Blurhash *string           `json:"blurhash,omitempty"`
}

type ImageMetadata struct {
    Path       string
    Blurhash   string
    Width      int
    Height     int
    ComputedAt string
}
```

Builder function:
```go
// BuildImageResource composes an ImageResource from a raw stored path +
// optional metadata. Returns nil when the path is empty.
func BuildImageResource(rawPath string, kind imagePathKind, meta *ImageMetadata) *ImageResource
```

Aspect per kind: poster=2:3, backdrop=16:9, still=16:9, logo="variable".

### 4. Typed paths (finalize from Phase 01)

Swap `ImagePath` field types across all models to kinded variants: `PosterPath`, `BackdropPath`, `StillPath`, `LogoPath`.

**Out of scope — User.AvatarPath and handler/auth.go userInfo.ProfilePath.** These carry user avatars which use TMDb profile size conventions (`w45`, `w185`, `h632`). Since profile kind is deferred (plan.md non-goals), avatar stays on the generic `ImagePath` type and its current normalization. Phase 07 also leaves these legacy fields in place. Revisit in a follow-up iteration.

Affected for this refactor: [model/media.go](../../backend/internal/model/media.go), [model/series.go](../../backend/internal/model/series.go), [model/user.go](../../backend/internal/model/user.go) (UserData.MediaPoster, ContinueWatchingItem, NextUpItem — NOT User.AvatarPath), [model/activity.go](../../backend/internal/model/activity.go).

### 5. Response envelope

Add **parallel** ImageResource fields alongside legacy string fields (progressive). Example for Media:

```go
type Media struct {
    // ... existing fields ...
    PosterPath   PosterPath   `json:"poster_path"`    // legacy
    BackdropPath BackdropPath `json:"backdrop_path"`  // legacy
    LogoPath     LogoPath     `json:"logo_path"`      // legacy
    ThumbPath    BackdropPath `json:"thumb_path"`     // legacy

    // Rich image resources (preferred). Populated by handlers via
    // MediaService.AttachImageResources before respondJSON.
    Poster   *ImageResource `json:"poster,omitempty"`
    Backdrop *ImageResource `json:"backdrop,omitempty"`
    Logo     *ImageResource `json:"logo,omitempty"`
    Thumb    *ImageResource `json:"thumb,omitempty"`
}
```

Service layer helper:
```go
// AttachImageResources enriches a Media with ImageResource objects by looking
// up blurhash/dimensions from image_metadata table. Handles nil meta gracefully.
func (s *MediaService) AttachImageResources(ctx context.Context, m *model.Media) error
```

Handlers call this before `respondJSON`. For list endpoints, batch-load metadata for all paths in one query (`GetBatch`).

## Tests

- [repository/image_metadata_test.go](../../backend/internal/repository/image_metadata_test.go): upsert + get + batch.
- [model/imageresource_test.go](../../backend/internal/model/imageresource_test.go): BuildImageResource covers tmdb/local, nil meta, each kind's aspect.

## Acceptance

- Migration applies + rolls back cleanly.
- `go test ./internal/...` green.
- Sample `GET /api/media/{id}` response has both `poster_path: "..."` (legacy) and `poster: { url, srcset, type, aspect, blurhash: null }` (new, blurhash null pre-compute).

## Notes

- Srcset enumeration uses the `tmdbSizes` table from Phase 01 — shared.
- `width`/`height` in `ImageResource` come from `image_metadata.width/height` (computed alongside blurhash). Nil until computed.
- `blurhash *string` (pointer) so `null` in JSON distinguishes "not computed yet" from "empty string".

---
Next: [Phase 03 — Blurhash computation service](phase-03-blurhash-service.md)
