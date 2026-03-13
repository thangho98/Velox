# Phase 02: Provider Clients (Backend)
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Implement 2 provider clients như 2 package độc lập trong `pkg/`.

---

## pkg/opensubs — OpenSubtitles.com REST v3

### Structs

```go
type Client struct {
    apiKey   string
    username string
    password string
    token    string        // JWT, cached
    tokenExp time.Time
    http     *http.Client
}

type SearchResult struct {
    ID           string   // file_id (dùng để download)
    SubtitleID   string   // subtitle_id
    Title        string
    Language     string
    Format       string   // srt, vtt, ass...
    Downloads    int
    Rating       float64
    FPS          float64
    HearingImp   bool     // SDH/CC
    Forced       bool
    AITranslated bool
}
```

### Methods
- `New(apiKey, username, password string) *Client`
- `Login(ctx) error` — POST /api/v1/login → cache token (24h)
- `ensureToken(ctx) error` — check expiry, re-login nếu cần
- `Search(ctx, params SearchParams) ([]SearchResult, error)` — GET /api/v1/subtitles
- `Download(ctx, fileID string) ([]byte, string, error)` — POST /api/v1/download → fetch link → return bytes + filename

### SearchParams
```go
type SearchParams struct {
    ImdbID   string // "tt1234567" hoặc "1234567"
    TmdbID   int
    Query    string // fallback: tên film
    Language string // "en", "vi"...
    Year     int
}
```

### API Endpoints
- `POST https://api.opensubtitles.com/api/v1/login`
  - Body: `{"username":"...","password":"..."}`
  - Headers: `Api-Key: {apiKey}`, `Content-Type: application/json`
  - Response: `{"token":"..."}`
- `GET https://api.opensubtitles.com/api/v1/subtitles?imdb_id=...&languages=...`
  - Headers: `Api-Key: {apiKey}`, `Authorization: Bearer {token}`
- `POST https://api.opensubtitles.com/api/v1/download`
  - Body: `{"file_id": 123}`
  - Response: `{"link":"https://..."}`

---

## pkg/podnapisi — Podnapisi JSON API

### Structs
```go
type Client struct {
    http *http.Client
}

type SearchResult struct {
    PID      string
    Title    string
    Language string
    Format   string
    Year     int
    Season   int
    Episode  int
    Rating   float64
    Downloads int
    URL      string // direct download URL
}
```

### Methods
- `New() *Client`
- `Search(ctx, params SearchParams) ([]SearchResult, error)` — GET /subtitles/search/advanced
- `Download(ctx, pid string) ([]byte, string, error)` — fetch zip → extract .srt

### API
- `GET https://www.podnapisi.net/subtitles/search/advanced?keywords={title}&year={year}&language={lang}&format=json`
- Response: `{"data": [{"pid":"...","title":"...","url":"/subtitles/{pid}/download",...}]}`
- Download: `GET https://www.podnapisi.net/subtitles/{pid}/download` → .zip file → extract .srt

---

## Shared type

```go
// pkg/subprovider/result.go
type Result struct {
    Provider   string // "opensubtitles" | "podnapisi"
    ExternalID string // provider's subtitle ID
    Title      string
    Language   string
    Format     string // "srt", "vtt", "ass"
    Downloads  int
    Rating     float64
    Forced     bool
    HearingImp bool
}
```

## Files
- `backend/pkg/opensubs/client.go`
- `backend/pkg/podnapisi/client.go`
- `backend/pkg/subprovider/result.go` — shared Result type
