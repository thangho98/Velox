# Phase 03: Source Resolver Abstraction
Status: ⬜ Pending
Dependencies: Phase 02

## Objective

Tạo `SourceResolver` interface trong `backend/internal/source/` để decouple library logic khỏi storage backend. `LocalResolver` wrap existing filesystem logic. `FshareResolver` mới dùng fshare client từ Phase 01.

## Context

Hiện tại scanner + stream handler hardcoded assumption "file = local absolute path" ở nhiều chỗ:
- [backend/internal/scanner/pipeline.go](backend/internal/scanner/pipeline.go) — `filepath.Walk`
- [backend/internal/handler/stream.go](backend/internal/handler/stream.go) — `http.ServeFile`
- [backend/pkg/ffprobe/ffprobe.go](backend/pkg/ffprobe/ffprobe.go) — probe local file

Target: thêm layer trung gian `SourceResolver` — scanner + stream handler gọi interface, không biết underlying storage.

## Implementation Steps

### 1. Define Interface
- [ ] Tạo `backend/internal/source/resolver.go`:
  ```go
  package source

  import (
      "context"
      "io"
      "time"
  )

  // StreamRef identifies a playable media item in a specific source.
  type StreamRef struct {
      SourcePath string // "fshare://xyz" hoặc "/mnt/data/foo.mkv"
      LibraryID  int64  // để resolve credentials (cloud only)
  }

  // DirectURL is the resolved URL for direct client playback.
  type DirectURL struct {
      URL       string
      ExpiresAt time.Time // zero value = no expiry (local files)
      Headers   map[string]string // e.g., Authorization for some sources
  }

  // FolderEntry represents one item during scan (file or subfolder).
  type FolderEntry struct {
      SourcePath string // full path/URI
      Name       string
      Size       int64
      IsDir      bool
      ParentID   string // native parent identifier
      Modified   time.Time
  }

  // SourceResolver abstracts storage backends.
  type SourceResolver interface {
      Provider() string // "local" | "fshare"

      // ListFolder returns entries inside the given source-native folder identifier.
      // For local: path is filesystem path. For fshare: folderCode.
      ListFolder(ctx context.Context, folderID string) ([]FolderEntry, error)

      // OpenRead opens a file for streaming bytes (used by scanner if probe needed).
      // For cloud sources, may return io.ErrUnsupported — caller handles gracefully.
      OpenRead(ctx context.Context, sourcePath string) (io.ReadCloser, error)

      // ResolveStreamURL returns a URL the client can stream directly.
      // For local: returns internal Velox URL + api_key.
      // For fshare: returns direct fshare CDN URL.
      ResolveStreamURL(ctx context.Context, ref StreamRef) (*DirectURL, error)

      // Exists checks if a source path still exists (for rescan prune).
      Exists(ctx context.Context, sourcePath string) (bool, error)
  }
  ```

### 2. LocalResolver Implementation
- [ ] Tạo `backend/internal/source/local_resolver.go`:
  ```go
  type LocalResolver struct {
      apiKeyStore *auth.APIKeyStore // existing
  }

  func NewLocalResolver(store *auth.APIKeyStore) *LocalResolver

  func (r *LocalResolver) Provider() string { return "local" }

  func (r *LocalResolver) ListFolder(ctx context.Context, dirPath string) ([]FolderEntry, error) {
      // wrap os.ReadDir
  }

  func (r *LocalResolver) OpenRead(ctx context.Context, path string) (io.ReadCloser, error) {
      return os.Open(path)
  }

  func (r *LocalResolver) ResolveStreamURL(ctx context.Context, ref StreamRef) (*DirectURL, error) {
      // Generate api_key (existing logic), return internal /api/stream/{id}?api_key=...
  }

  func (r *LocalResolver) Exists(ctx context.Context, path string) (bool, error) {
      _, err := os.Stat(path); return err == nil, nil
  }
  ```

### 3. FshareResolver Implementation
- [ ] Tạo `backend/internal/source/fshare_resolver.go`:
  ```go
  type FshareResolver struct {
      clientFactory func() *fshare.Client // new client per call (session isolation)
      sessionRepo   *repository.CloudSessionRepo
      cryptoKey     []byte
      libraryRepo   *repository.LibraryRepo
  }

  // hydrateClient: load cloud_session → decrypt → inject into fresh client.
  func (r *FshareResolver) hydrateClient(ctx context.Context, credentialsID int64) (*fshare.Client, *model.CloudSession, error) {
      sess, err := r.sessionRepo.GetByID(ctx, credentialsID)
      if err != nil { return nil, nil, err }

      // Decrypt credentials (for auto-relogin)
      credsJSON, err := crypto.Decrypt(r.cryptoKey, sess.CredentialsEncrypted)
      var creds CredentialsPayload; json.Unmarshal(credsJSON, &creds)

      client := r.clientFactory()
      client.SetCredentials(creds.Email, creds.Password)

      // Optional: inject existing token + session_id if not expired
      if sess.TokenEncrypted != nil && sess.SessionIDEncrypted != nil {
          token, _ := crypto.Decrypt(r.cryptoKey, sess.TokenEncrypted)
          sessionID, _ := crypto.Decrypt(r.cryptoKey, sess.SessionIDEncrypted)
          client.RestoreSession(string(token), string(sessionID))
      }
      return client, sess, nil
  }

  // persistSessionIfChanged: if client's session changed during call → save back encrypted.
  func (r *FshareResolver) persistSessionIfChanged(ctx context.Context, sess *model.CloudSession, client *fshare.Client, preCallToken string) error {
      current := client.Session()
      if current == nil || current.Token == preCallToken {
          return nil // no change
      }
      tokEnc, _ := crypto.Encrypt(r.cryptoKey, []byte(current.Token))
      sidEnc, _ := crypto.Encrypt(r.cryptoKey, []byte(current.SessionID))
      now := time.Now()
      expiry := now.Add(25 * time.Minute)
      sess.TokenEncrypted = tokEnc
      sess.SessionIDEncrypted = sidEnc
      sess.TokenExpiresAt = &expiry
      sess.LastRefreshAt = &now
      return r.sessionRepo.Update(ctx, sess)
  }

  func NewFshareResolver(factory func() *fshare.Client, sessRepo *repository.CloudSessionRepo, key []byte, libRepo *repository.LibraryRepo) *FshareResolver

  func (r *FshareResolver) Provider() string { return "fshare" }

  func (r *FshareResolver) ListFolder(ctx context.Context, folderCode string) ([]FolderEntry, error) {
      items, err := r.client.ListFolder(ctx, folderCode)
      // Convert FolderItem → FolderEntry with sourcePath = "fshare://" + linkcode
  }

  func (r *FshareResolver) OpenRead(ctx context.Context, sourcePath string) (io.ReadCloser, error) {
      return nil, io.ErrUnsupported // KHÔNG support remote probe phase này
  }

  func (r *FshareResolver) ResolveStreamURL(ctx context.Context, ref StreamRef) (*DirectURL, error) {
      // 1. Load library → get source_credentials_id
      // 2. Load cloud_session → decrypt credentials (email+password) + token + session_id
      // 3. Inject session into fshare.Client (token in struct, cookie via jar)
      // 4. client.SetCredentials(email, password) — for auto-relogin on code:201
      // 5. Parse fileCode from sourcePath ("fshare://abc" → "abc")
      // 6. Call client.GetDirectLink → client auto-retries on ErrSessionExpired
      // 7. Persist new session back to DB if re-login happened (compare token before/after)
      // 8. Return DirectURL with ExpiresAt = time.Now() + 5min (conservative — URL có thể single-use)
  }

  func (r *FshareResolver) Exists(ctx context.Context, sourcePath string) (bool, error) {
      // Call client.GetFileInfo → check not-found vs error
  }
  ```

### 4. Resolver Factory / Registry
- [ ] Tạo `backend/internal/source/factory.go`:
  ```go
  type ResolverFactory struct {
      resolvers map[string]SourceResolver
  }

  func NewResolverFactory() *ResolverFactory

  func (f *ResolverFactory) Register(r SourceResolver) {
      f.resolvers[r.Provider()] = r
  }

  func (f *ResolverFactory) For(sourceType string) (SourceResolver, error) {
      r, ok := f.resolvers[sourceType]
      if !ok {
          return nil, fmt.Errorf("unknown source type: %s", sourceType)
      }
      return r, nil
  }

  // ByLibrary picks resolver based on library.SourceType.
  func (f *ResolverFactory) ByLibrary(lib *model.Library) (SourceResolver, error) {
      return f.For(lib.SourceType)
  }
  ```

### 5. Path Prefix Helpers
- [ ] Tạo `backend/internal/source/path.go`:
  ```go
  const FshareScheme = "fshare://"

  // ParseSourcePath returns (provider, nativeID).
  // "fshare://abc" → ("fshare", "abc")
  // "/mnt/data/foo.mkv" → ("local", "/mnt/data/foo.mkv")
  func ParseSourcePath(p string) (provider, id string)

  func IsFsharePath(p string) bool { return strings.HasPrefix(p, FshareScheme) }

  func FshareFileCode(p string) string { return strings.TrimPrefix(p, FshareScheme) }

  func FsharePath(fileCode string) string { return FshareScheme + fileCode }
  ```
- [ ] Unit tests cho các helper

### 6. Wire vào ServerApp
- [ ] `backend/cmd/server/server_app.go`:
  - [ ] Nếu `config.Cloud.Enabled`:
    - [ ] Tạo `fshare.Client`
    - [ ] Tạo `FshareResolver`
    - [ ] Register vào factory
  - [ ] Luôn register `LocalResolver`
  - [ ] Inject `ResolverFactory` vào services cần dùng (Phase 04 scanner, Phase 05 stream handler)

## Acceptance Criteria

- [ ] `backend/internal/source/` compile clean với unit tests
- [ ] `LocalResolver` test: listFolder trên tmpdir, resolveStreamURL returns valid internal URL + api_key
- [ ] `FshareResolver` test (with mock fshare.Client): listFolder converts items, resolveStreamURL decrypts token + calls client
- [ ] `ResolverFactory` test: register + lookup, unknown type error
- [ ] `ParseSourcePath` + helpers unit tests pass (10+ edge cases)
- [ ] ServerApp startup KHÔNG regression khi `VELOX_CLOUD_SOURCES_ENABLED=false`

## Design Notes

- **Interface granularity:** `OpenRead` intentional cho scanner probe (local only); fshare returns `io.ErrUnsupported` — scanner caller handles (skip ffprobe for cloud).
- **ResolveStreamURL placement:** Resolver biết cách generate URL, handler chỉ orchestrate. Keeps handler thin.
- **Session refresh location:** `fshare.Client` internal `withSessionRetry` handles code:201 → relogin automatic (Phase 01). FshareResolver orchestrates: decrypt creds from DB → feed vào client → call API → persist refreshed session back. NO retry logic duplicated in resolver.
- **Credentials access:** Resolver NHẬN `sessionRepo` + `cryptoKey` via constructor — KHÔNG pass qua method args (reduces handler complexity).
- **Error mapping (fshare → source):** FshareResolver wraps `fshare` package errors:
  - `fshare.ErrSessionExpired` (after relogin fail) → `source.ErrSessionExpired` (HTTP 401)
  - `fshare.ErrInvalidCredentials` → `source.ErrInvalidCredentials` (HTTP 401)
  - `fshare.ErrRateLimit` → `source.ErrRateLimit` (HTTP 429)
  - `fshare.ErrLinkDead` → `source.ErrFileNotFound` (HTTP 404)
  - `fshare.ErrNotVIP` → `source.ErrInvalidCredentials` (HTTP 403 with clear message)

## Gotchas

- `io.ErrUnsupported` cần import từ `errors` (Go 1.21+). Chúng ta đang dùng 1.26 — OK.
- Factory map lookup race-condition: register xảy ra một lần ở startup → no mutex needed nếu đúng pattern (construct → register → done → read-only).
- `FshareResolver.ResolveStreamURL` error path: nếu session refresh fail → return typed error để handler biết status code trả về (`403` vs `503`).
- **Persist refreshed session back:** Khi fshare.Client auto-relogin (Phase 01 `withSessionRetry`), session values change. Resolver PHẢI compare `client.Session().Token` với pre-call token; nếu khác → encrypt + update cloud_sessions row. Otherwise stale token gets re-written next call. Pattern: capture token pre-call, compare post-call, conditional write.
- **Client instance per request vs shared:** `fshare.Client` có session state. Don't share across library credentials. Factory để FshareResolver construct new client per library (hoặc per-credential-ID cache với LRU eviction).

## Out of Scope

- Multi-account fshare trong 1 library (1 library = 1 account)
- Stream URL caching (phase 5 optionally adds TTL cache)
- Parallel listFolder optimization (scanner phase 4 handles concurrency)
