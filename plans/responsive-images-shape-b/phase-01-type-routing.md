# Phase 01: Per-Type URL Routing + Response Stub

Status: ✅ Complete
Dependencies: none

## Objective

Route image requests by type (`/api/images/tmdb/{type}/{path}?width=N`), with per-type TMDb size buckets so `?width=500` maps correctly for backdrop/still/logo instead of 404.

No FE changes yet — purely backend. Legacy JSON string fields stay as-is. FE will start migrating in Phase 05-06.

## Changes

### Route

[cmd/server/server_routes.go](../../backend/cmd/server/server_routes.go):
```diff
- mux.HandleFunc("GET /api/images/tmdb/{path...}", app.handlers.image.Serve)
+ mux.HandleFunc("GET /api/images/tmdb/{type}/{path...}", app.handlers.image.Serve)
```

### Handler

[handler/image.go](../../backend/internal/handler/image.go) — rewrite size resolver per type:

```go
var tmdbSizes = map[string][]int{
    "poster":   {92, 154, 185, 342, 500, 780},
    "backdrop": {300, 780, 1280},
    "still":    {92, 185, 300},
    "logo":     {45, 92, 154, 185, 300, 500},
}

var tmdbDefault = map[string]int{"poster": 500, "backdrop": 780, "still": 300, "logo": 500}

func resolveTMDbSize(imageType, widthParam string) (string, bool) {
    buckets, ok := tmdbSizes[imageType]
    if !ok { return "", false }
    if widthParam == "original" { return "original", true }
    n, err := strconv.Atoi(widthParam)
    if err != nil || n <= 0 { n = tmdbDefault[imageType] }
    for _, b := range buckets {
        if n <= b { return fmt.Sprintf("w%d", b), true }
    }
    return "original", true
}
```

Handler:
```go
imageType := r.PathValue("type")
size, ok := resolveTMDbSize(imageType, r.URL.Query().Get("width"))
if !ok { respondError(w, http.StatusBadRequest, "invalid image type"); return }
```

### Normalizer (model/imagepath.go)

Introduce 4 typed paths: `PosterPath`, `BackdropPath`, `StillPath`, `LogoPath`. Each has `MarshalJSON` emitting `/api/images/tmdb/{type}{path}` with the correct segment.

See Phase 02 for the full types — Phase 01 only needs the handler + normalize helper. Temporary measure: keep existing `ImagePath` type functional but extend normalizer to take an optional type hint:

```go
func (p ImagePath) normalize(kind string) string {
    // existing rules + for tmdb relative paths: "/api/images/tmdb/" + kind + path
}
```

Actual model fields stay `ImagePath` until Phase 02.

## Tests

New [handler/image_test.go](../../backend/internal/handler/image_test.go):
- Table test: 4 types × various widths → expected TMDb size.
- Edge: `original` passthrough, invalid width fallback to default, unknown type → 400.

Keep [model/imagepath_test.go](../../backend/internal/model/imagepath_test.go) cases updated for new URL shape `/api/images/tmdb/{type}/{path}`.

## Acceptance

- `go build ./... && go vet ./... && go test ./internal/handler/... ./internal/model/...` green.
- Manual curl:
  ```sh
  curl -I localhost:8098/api/images/tmdb/backdrop/abc.jpg?width=500     # → 200
  curl -I localhost:8098/api/images/tmdb/still/abc.jpg?width=500        # → 200
  curl -I localhost:8098/api/images/tmdb/avatar/abc.jpg                 # → 400
  ```

## Notes

- DB unchanged.
- FE helper (webapp + Android) already handles `/api/` prefix transparently — no code change needed for new URL shape.
- Old cached FE URLs without `/type/` segment 404. Acceptable — reload fixes.

---
Next: [Phase 02 — ImageResource model + image_metadata table](phase-02-image-resource-model.md)
