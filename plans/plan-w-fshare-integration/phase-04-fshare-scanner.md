# Phase 04: Cloud Scanner (Provider-based)
Status: ⬜ Pending
Dependencies: Phase 03

## Objective

Viết `CloudWalker` — generic BFS scanner sử dụng `cloudstorage.Provider` interface. Works cho fshare hôm nay, GDrive/OneDrive khi có driver mới. Insert `media_files` rows với `path = {provider_type}://{native_id}` scheme. Reuse filename parsing + TMDb matching. Broadcast scan progress via websocket.

## Context

- Existing scanner: [backend/internal/scanner/pipeline.go](backend/internal/scanner/pipeline.go) — assume local filesystem
- TMDb matcher: `backend/internal/scanner/matcher/` (reuse)
- Filename parser: `backend/pkg/filename/` (reuse)
- Websocket broadcast: `backend/internal/service/task_service.go` (reuse pattern từ marker backfill)
- 10TB library ≈ 2500 files (4GB/file trung bình)
- TMDb rate limit: 40 req/10s → cần throttle
- Library has `storage_provider_id` + `source_url`; scanner resolves provider via Registry, parses URL to native folder ID, BFS qua Provider interface

## Path Scheme

Use `{provider_type}://{native_id}` so `media_files.path` encodes both provider + ID:

| Provider | Example Path | Encoding |
|---|---|---|
| local | `/mnt/data/foo.mkv` | filesystem path |
| fshare | `fshare://XZWCPAZV3J71` | scheme + linkcode |
| gdrive (future) | `gdrive://1a2b3c...` | scheme + fileId |
| onedrive (future) | `onedrive://01AB...` | scheme + itemId |

Parsing: `strings.SplitN(path, "://", 2)` → `(providerType, nativeID)`.

## Implementation Steps

### 1. CloudWalker Struct (provider-agnostic)
- [ ] Tạo `backend/internal/scanner/cloud_walker.go`:
  ```go
  package scanner

  type CloudWalker struct {
      provider     cloudstorage.Provider // fshare, gdrive, onedrive — walker doesn't care
      library      *model.Library
      mediaRepo    *repository.MediaFileRepo
      matcher      matcher.TMDbMatcher
      progressCh   chan<- ScanProgress
      tmdbThrottle *rate.Limiter // 40 req / 10s
  }

  type ScanProgress struct {
      LibraryID      int64
      TotalFiles     int
      ProcessedFiles int
      CurrentFolder  string
      MatchedByTMDb  int
      Errors         []string
  }

  func NewCloudWalker(opts CloudWalkerOpts) *CloudWalker
  ```

### 2. BFS Folder Traversal (provider-agnostic)
- [ ] Method `Walk(ctx context.Context) error`:
  ```go
  // Parse library.SourceURL via provider → root folder ID
  rootID, err := w.provider.ParseFolderURL(*w.library.SourceURL)
  if err != nil { return err }

  queue := []string{rootID}
  seen := make(map[string]bool)

  for len(queue) > 0 {
      folderID := queue[0]
      queue = queue[1:]
      if seen[folderID] { continue }
      seen[folderID] = true

      items, err := w.provider.ListFolder(ctx, folderID)
      if err != nil {
          // log + continue — 1 folder fail không abort toàn bộ scan
          w.progressCh <- ScanProgress{Errors: []string{err.Error()}}
          continue
      }

      for _, item := range items {
          if item.IsFolder {
              queue = append(queue, item.ID)
          } else {
              w.processFile(ctx, item)
          }
      }
  }
  ```
- [ ] BFS ưu tiên breadth (user thấy top-level folders populate trước)
- [ ] Cycle detection via `seen` map (fshare không có shortcuts; GDrive có — future drivers cần aware)

### 3. File Processing + Dedupe
- [ ] Method `processFile(ctx, item)`:
  ```go
  // 1. Filter non-video MIME (check item.Mimetype + fallback to extension)
  if !isVideoMime(item.Mimetype) && !isVideoExt(item.Name) { return }

  // 2. Encode path with provider scheme: "fshare://XYZ"
  storagePath := w.provider.Type() + "://" + item.ID

  // 3. Dedupe: check media_files WHERE path = storagePath
  existing, err := w.mediaRepo.GetByPath(ctx, storagePath)
  if existing != nil { return } // already scanned

  // 4. Parse filename → title + year + season + episode
  parsed := filename.Parse(item.Name)

  // 5. TMDb match (rate-limited)
  w.tmdbThrottle.Wait(ctx)
  match, err := w.matcher.Match(ctx, parsed)

  // 6. Insert media_files
  mf := &model.MediaFile{
      LibraryID: w.library.ID,
      Path:      storagePath,
      Size:      item.Size,
      Duration:  0, // unknown for cloud — ExoPlayer probes runtime
      Codec:     "",
      IsHDR:     false, // runtime probe not available
      DVProfile: 0,
      TMDbID:    match.TMDbID,
      MediaType: match.MediaType,
      // ... title, year, season, episode từ parsed + match
  }
  w.mediaRepo.Insert(ctx, mf)
  ```

### 4. Cloud-Aware Field Defaults
- [ ] Fields KHÔNG scan được cho cloud files:
  - `duration`: 0 (ExoPlayer runtime probe)
  - `codec`, `container`, `audio_codec`: empty
  - `is_hdr`, `dv_profile`, `color_transfer`, `color_primaries`: 0/empty
  - `fingerprint`: hash(`path + size`) (no content hash)
- [ ] **Sentinel safety:** Các field này đã default OFF trong existing logic (migration 035 safety net). Không ảnh hưởng runtime decisions.

### 5. Progress Broadcast
- [ ] Reuse websocket pattern từ marker backfill:
  ```go
  // Mỗi 50 files hoặc mỗi folder done → broadcast
  ws.Broadcast("scan.progress", ScanProgress{...})
  ```
- [ ] Events:
  - `scan.started` — tổng folder estimate (có thể chưa có, broadcast unknown=true)
  - `scan.progress` — mỗi 50 files
  - `scan.folder_done` — mỗi folder xong
  - `scan.completed` — summary
  - `scan.failed` — error khi không recover được (ví dụ session expired + re-login fail)

### 6. Pipeline Integration
- [ ] Modify [backend/internal/scanner/pipeline.go](backend/internal/scanner/pipeline.go):
  ```go
  func (p *Pipeline) ScanLibrary(ctx context.Context, lib *model.Library) error {
      if lib.StorageProviderID == nil {
          return p.scanLocal(ctx, lib) // existing filesystem path
      }

      // Cloud library: load provider + credentials from registry
      sp, err := p.providerRepo.GetByID(ctx, *lib.StorageProviderID)
      if err != nil { return fmt.Errorf("load provider: %w", err) }

      driver, err := p.registry.Get(sp.ProviderType)
      if err != nil { return err }

      creds, err := cloudstorage.DecryptCredentials(p.cryptoKey, sp.CredentialsEncrypted)
      if err != nil { return fmt.Errorf("decrypt creds: %w", err) }

      provider, err := driver.NewProvider(creds)
      if err != nil { return err }

      walker := NewCloudWalker(CloudWalkerOpts{
          provider:   provider,
          library:    lib,
          mediaRepo:  p.mediaRepo,
          matcher:    p.matcher,
          progressCh: p.progressCh,
      })
      return walker.Walk(ctx)
  }
  ```

### 7. Rescan / Prune
- [ ] Rescan logic: rerun BFS, compare existing rows:
  - New path → insert
  - Existing path → skip (no metadata update phase này)
  - Row exists nhưng fshare returned not-found → mark `status = 'missing'` (DON'T delete — user có thể muốn restore)
- [ ] Prune trigger: admin button "Remove missing files" (Phase 06)

## Acceptance Criteria

- [ ] `scanner/cloud_walker_test.go` pass với mocked `cloudstorage.Provider` + matcher:
  - [ ] BFS đúng thứ tự
  - [ ] Dedupe skip existing paths (with `{type}://{id}` encoding)
  - [ ] Non-video extensions skipped
  - [ ] TMDb throttle respected
  - [ ] Progress events broadcast đúng frequency
  - [ ] Cycle detection (same folderID visited once)
- [ ] Integration test: fake provider với 3-level tree (5 folders, 20 files) → verify all 20 inserted
- [ ] Pipeline integration: `ScanLibrary` dispatch đúng theo `storage_provider_id` (nil = local, else cloud)
- [ ] Manual smoke test (dev env): scan fshare VIP account với 1 folder nhỏ (~50 files) → verify `media_files` rows có `path = fshare://...`
- [ ] Zero regression trên local scanner: existing tests pass
- [ ] Walker KHÔNG hardcode "fshare" — test swap với mock provider for GDrive-like behavior

## Performance Targets

- 2500 files scan < 60 min (TMDb throttle = bottleneck: 2500 / 4 req/sec ≈ 10 min nếu TMDb là dominant)
- Memory: < 200MB during scan (streaming insert, không batch 10K rows)
- Websocket frequency: max 1 event/sec (batch progress)

## Gotchas

- **TMDb failure**: Không abort file insert — insert với `tmdb_id = NULL`, user có thể manual match sau.
- **Large folder (>1000 items)**: Pagination phải work đúng. Test với fake server returning 10 pages.
- **Session expire giữa scan**: FshareResolver auto-retry 1 lần (Phase 03). Nếu vẫn fail → mark scan failed, user re-authenticate.
- **Concurrent scans**: Cho phép 1 scan per library (mutex). Multiple libraries scan song song OK.
- **Rate limit fshare API**: Observed ~20 req/sec safe. Tune trong Phase 01.

## Out of Scope

- Delta scan (compare modified_at) — phase sau
- ffprobe via partial-download (Range GET first 10MB) — phase sau nếu cần metadata chính xác hơn
- Subtitle sidecar extraction (.srt next to video) — phase sau
- Fingerprint content hash — bỏ hẳn (không practical cho cloud)
