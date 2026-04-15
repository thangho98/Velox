# Phase 03: Blurhash Computation Service

Status: ⬜ Pending
Dependencies: Phase 02 (image_metadata table exists)

## Objective

Pure computation layer: given a raw image path, fetch + decode + compute blurhash + extract dimensions + upsert into `image_metadata`. Independent of who calls it (ingestion hooks in Phase 04, backfill CLI in Phase 04).

## Changes

### 1. Dependency

`go get github.com/bbrks/go-blurhash`

### 2. Service

[service/imagemeta/service.go](../../backend/internal/service/imagemeta/service.go) — new package:

```go
type Service struct {
    repo         *repository.ImageMetadataRepo
    storage      *storage.ImageStorage
    httpClient   *http.Client
    tmdbRateLtd  *rate.Limiter  // golang.org/x/time/rate
}

// Compute fetches (if remote), decodes, computes blurhash + dimensions for a
// raw stored path (local://... or TMDb relative /abc.jpg). Upserts into
// image_metadata. Idempotent — no-op if already computed.
func (s *Service) Compute(ctx context.Context, rawPath string) (*model.ImageMetadata, error)

// ComputeBatch processes N paths with concurrency + rate limit. Returns per-path
// errors without aborting the batch.
func (s *Service) ComputeBatch(ctx context.Context, paths []string) map[string]error
```

### 3. Image sourcing

```go
// fetchImage returns a decoded image.Image for any supported path scheme.
func (s *Service) fetchImage(ctx context.Context, rawPath string) (image.Image, error) {
    switch {
    case strings.HasPrefix(rawPath, "local://"):
        // local://{type}/{id}/{filename} — read from ImageStorage
        parts := strings.SplitN(strings.TrimPrefix(rawPath, "local://"), "/", 3)
        // parts = [entityType, id, filename]
        absPath := s.storage.AbsPath(parts[0], parseInt64(parts[1]), parts[2])
        return decodeFile(absPath)
    case strings.HasPrefix(rawPath, "/"):
        // TMDb relative path — fetch small variant
        return s.fetchTMDbImage(ctx, rawPath)
    default:
        return nil, fmt.Errorf("unsupported path scheme: %s", rawPath)
    }
}

func (s *Service) fetchTMDbImage(ctx context.Context, relPath string) (image.Image, error) {
    if err := s.tmdbRateLtd.Wait(ctx); err != nil { return nil, err }
    // Use w185 — smallest decent quality, fastest. Blurhash doesn't need high res.
    url := "https://image.tmdb.org/t/p/w185" + relPath
    // ... http GET, decode
}
```

### 4. Blurhash + dimensions

```go
func computeHash(img image.Image) (string, int, int, error) {
    bounds := img.Bounds()
    w, h := bounds.Dx(), bounds.Dy()
    hash, err := blurhash.Encode(4, 3, img)
    return hash, w, h, err
}
```

Components 4×3 empirically good for movie posters/backdrops. ~30-char hash.

### 5. Idempotency + caching

`Compute` checks repo first:
```go
if existing, _ := s.repo.Get(ctx, rawPath); existing != nil {
    return existing, nil
}
```

### 6. Rate limit config

- TMDb fetches: 20 req/s default (conservative, TMDb allows ~40/s). Configurable via env `TMDB_BLURHASH_RPS`.
- Local reads: no rate limit (disk IO bound).

### 7. Wiring

Follow the existing constructor pattern in [backend/cmd/server/server_app.go](../../backend/cmd/server/server_app.go) (same file that wires `app.services.*` and `app.handlers.image = handler.NewImageHandler()` around lines 261/391). Add:

```go
// In the services wiring section of server_app.go:
app.services.imagemeta = imagemeta.NewService(
    repos.imageMetadata,
    imageStorage,
    httpClient,
    rate.NewLimiter(20, 1),
)
// Pass to services that ingest images (optional setter to keep constructor sigs stable):
app.services.metadata.SetImageMetaService(app.services.imagemeta)
```

`repos.imageMetadata` is the new `ImageMetadataRepo` constructed alongside the other repos.

## Tests

[service/imagemeta/service_test.go](../../backend/internal/service/imagemeta/service_test.go):
- `Compute` for a local JPEG (testdata fixture) → verify blurhash + dims in DB.
- `Compute` idempotent (second call no-op).
- `ComputeBatch` with mixed success/failure returns per-path errors.
- Skip TMDb HTTP calls (integration test w/ `-short` flag).

## Acceptance

- `go test ./internal/service/imagemeta/... -count=1` green.
- Manual: pick a local uploaded poster, call `imagemetaSvc.Compute(ctx, path)` via a one-shot binary → verify row in `image_metadata`.

## Notes

- Blurhash library uses `image.Image` interface — works with decoded JPEG/PNG/WebP out of the box (std lib registrations already in storage package).
- For TMDb: fetching w185 (~15KB) is ~5x faster than w500. Blurhash quality is identical at either input size.
- If TMDb fetch fails (404, timeout), log + skip — don't block. Blurhash is best-effort.
- No retry on TMDb 404 — the path is bad, don't thrash.

---
Next: [Phase 04 — Ingestion hooks + backfill CLI](phase-04-ingestion-backfill.md)
