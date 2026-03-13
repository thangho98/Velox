# Phase 03: Search + Download API (Backend)
Status: ⬜ Pending
Dependencies: Phase 01 + 02

## Objective
Expose 2 endpoints: search (query cả 2 providers) và download (tải về + lưu disk + tạo DB record).

---

## Service: SubtitleSearchService

```go
type SubtitleSearchService struct {
    openSubs  *opensubs.Client  // nil nếu chưa config
    podnapisi *podnapisi.Client
    mediaRepo *repository.MediaRepo
    mfRepo    *repository.MediaFileRepo
    subRepo   *repository.SubtitleRepo
    settings  *repository.AppSettingsRepo
    cacheDir  string // ~/.velox/subtitles/downloaded/
}
```

### Methods

**Search:**
```go
func (s *SubtitleSearchService) Search(ctx context.Context, mediaID int64, lang string) ([]subprovider.Result, error)
```
- Load media để lấy imdb_id, tmdb_id, title, year
- Gọi OpenSubtitles nếu configured (apiKey + credentials set)
- Gọi Podnapisi luôn
- Merge kết quả: OpenSubs trước, Podnapisi sau
- Dedup theo (language + title similarity)

**Download:**
```go
func (s *SubtitleSearchService) Download(ctx context.Context, mediaID int64, provider, externalID string) (*model.Subtitle, error)
```
- Tải file từ provider
- Detect format (.srt/.vtt/.ass)
- Lưu vào `cacheDir/{mediaID}/{provider}_{externalID}.{ext}`
- Tạo Subtitle record: `media_file_id = primaryFile.ID`, `is_embedded=false`, `file_path=...`
- Return subtitle để frontend dùng ngay

---

## Handler: SubtitleSearchHandler

```go
type SubtitleSearchHandler struct {
    svc *service.SubtitleSearchService
}
```

### Routes

```
GET  /api/media/{id}/subtitles/search
     Query: ?lang=en (default: all)
     Response: { "data": [ ...subprovider.Result ] }
     RequireAuth ✅

POST /api/media/{id}/subtitles/download
     Body: { "provider": "opensubtitles", "external_id": "12345" }
     Response: { "data": { subtitle model } }
     RequireAuth ✅
```

---

## Error cases
| Case | Response |
|------|----------|
| OpenSubtitles không config | Chỉ trả Podnapisi results, không error |
| OpenSubtitles login fail | Log warning, skip, trả Podnapisi results |
| Provider trả 0 kết quả | `{"data": []}` |
| Download fail | 500 + message |
| Provider rate limit | 429, retry sau 1s (1 lần) |

---

## Files
- `backend/internal/service/subtitle_search.go`
- `backend/internal/handler/subtitle_search.go`
- `backend/cmd/server/routes.go` — thêm 2 routes
