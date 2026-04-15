# Phase 04: Ingestion Hooks + Backfill CLI

Status: ✅ Done
Dependencies: Phase 03 (Compute service works)

## Objective

1. Auto-compute blurhash when new images enter the system (TMDb scan, user upload).
2. One-shot CLI to backfill all existing images — the 1000+ media items already in DB.
3. Keep computation async so user uploads don't block HTTP response.

## Part A: Ingestion hooks

### Local uploads

[service/metadata_editor_operations.go](../../backend/internal/service/metadata_editor_operations.go) — the `UploadMediaImage` / `UploadSeriesImage` flow already decodes the image for resize in [storage/image.go](../../backend/internal/storage/image.go). Extend `ImageStorage.Save` to return dimensions + compute blurhash synchronously before DB write:

```go
func (s *ImageStorage) Save(entityType string, id int64, imageType string, data []byte) (path string, meta ImageMetaResult, err error)
```

`ImageMetaResult { Blurhash string, Width int, Height int }` — computed from the already-decoded `img` variable.

Caller persists to `image_metadata` table after storing the file.

Sync is fine here: upload is already slow (decode + resize), blurhash adds <50ms.

### TMDb ingestion (scanner + metadata service)

After every `media.PosterPath = ...` / `series.BackdropPath = ...` assignment in [service/metadata.go](../../backend/internal/service/metadata.go), [metadata_editor.go](../../backend/internal/service/metadata_editor.go), [metadata_enrichment.go](../../backend/internal/service/metadata_enrichment.go), enqueue the path for async compute:

```go
imagemetaSvc.Enqueue(ctx, rawPath)  // fire-and-forget
```

### Async worker

[service/imagemeta/worker.go](../../backend/internal/service/imagemeta/worker.go):
```go
// Worker runs a single goroutine consuming a channel of paths to compute.
// Bounded channel (buffer 256). Overflow drops newest (logged) — backfill CLI
// can sweep missed ones.
type Worker struct {
    svc   *Service
    queue chan string
}

func (w *Worker) Enqueue(path string) { select { case w.queue <- path: default: /* log drop */ } }
func (w *Worker) Run(ctx context.Context) { for p := range w.queue { w.svc.Compute(ctx, p) } }
```

Start in [backend/cmd/server/server_runtime.go](../../backend/cmd/server/server_runtime.go) alongside the existing `go func()` background loops (session cleanup, library scan, etc. — grep for `app.services.scheduler.Register` to find the pattern). Worker goroutine lives for the process lifetime, queue is `app.services.imagemeta.Worker` (constructed in `server_app.go`).

## Part B: Backfill CLI

### Subcommand

[backend/cmd/server/main.go](../../backend/cmd/server/main.go) — add `blurhash` case to `handleCommand` (alongside existing `case "migrate"` and `case "version"` around lines 34/37):

```
velox blurhash backfill              # process all missing
velox blurhash backfill --force      # recompute even if present
velox blurhash backfill --limit=100  # sample mode
```

Subcommand dispatch pattern already uses `os.Args[2]` for subsubcommands — follow the existing convention from the `migrate` branch.

### Implementation

[backend/cmd/server/cmd_blurhash.go](../../backend/cmd/server/cmd_blurhash.go) — new file:

1. Load config + DB.
2. Query all distinct image paths across tables:
   ```sql
   SELECT poster_path FROM media WHERE poster_path != ''
   UNION SELECT backdrop_path FROM media WHERE backdrop_path != ''
   UNION SELECT logo_path FROM media WHERE logo_path != ''
   UNION SELECT thumb_path FROM media WHERE thumb_path != ''
   UNION SELECT poster_path FROM series ...
   -- etc for all image-carrying columns
   ```
3. LEFT JOIN against `image_metadata` to filter out already-computed (unless `--force`).
4. Call `imagemetaSvc.ComputeBatch(ctx, paths)` with progress logging every 50 items.
5. Summary at end: `processed=N computed=N failed=N duration=X`.

Rate limit from Phase 03's limiter applies → 1000 TMDb paths ≈ 50s.

## Tests

- Worker: enqueue → verify Compute called (via mock svc).
- Overflow drop: flood queue → verify non-blocking + drops logged.
- Backfill CLI: integration test with in-memory SQLite + small fixture set. Skip TMDb network with `-short`.

## Acceptance

- Upload a new poster via MetadataEditor → reload `/api/media/{id}` → `poster.blurhash` populated immediately.
- TMDb scan of a new media file → within a few seconds, blurhash appears.
- `./bin/velox blurhash backfill --limit=10` runs, prints progress, populates rows.
- Worker queue overflow logged, not crashed.

## Notes

- Ingestion hooks are "fire-and-forget". If worker is down (shouldn't be — runs in the same process), enqueue is a no-op drop. Backfill CLI catches anything missed.
- After deploy, operator runs `blurhash backfill` once. Subsequent deploys need no manual action.
- Progress log shape: `blurhash: [200/1000] computed=192 failed=8 elapsed=10s`.

---
Next: [Phase 05 — Webapp `<ResponsiveImage>` component](phase-05-webapp-component.md)
