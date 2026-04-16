# Phase 04: Fshare Scanner + Metadata Matching
Status: ⬜ Pending
Dependencies: Phase 03

## Objective

Viết `FshareWalker` — BFS qua fshare folder tree, insert `media_files` rows với `path = fshare://{code}`. Reuse filename parsing + TMDb matching từ existing scanner. Broadcast scan progress qua websocket cho admin dashboard.

## Context

- Existing scanner: [backend/internal/scanner/pipeline.go](backend/internal/scanner/pipeline.go) — assume local filesystem
- TMDb matcher: `backend/internal/scanner/matcher/` (reuse)
- Filename parser: `backend/pkg/filename/` (reuse)
- Websocket broadcast: `backend/internal/service/task_service.go` (reuse pattern từ marker backfill)
- 10TB library ≈ 2500 files (4GB/file trung bình)
- TMDb rate limit: 40 req/10s → cần throttle

## Implementation Steps

### 1. FshareWalker Struct
- [ ] Tạo `backend/internal/scanner/fshare_walker.go`:
  ```go
  package scanner

  type FshareWalker struct {
      resolver      source.SourceResolver // FshareResolver
      library       *model.Library
      mediaRepo     *repository.MediaFileRepo
      matcher       matcher.TMDbMatcher
      progressCh    chan<- ScanProgress
      pageSize      int // default 100
      tmdbThrottle  *rate.Limiter // 40 req / 10s
  }

  type ScanProgress struct {
      LibraryID       int64
      TotalFiles      int
      ProcessedFiles  int
      CurrentFolder   string
      MatchedByTMDb   int
      Errors          []string
  }

  func NewFshareWalker(opts FshareWalkerOpts) *FshareWalker
  ```

### 2. BFS Folder Traversal
- [ ] Method `Walk(ctx context.Context) error`:
  ```go
  // BFS queue starting từ library.SourceRootID
  queue := []string{library.SourceRootID}
  seen := make(map[string]bool)

  for len(queue) > 0 {
      folderCode := queue[0]
      queue = queue[1:]
      if seen[folderCode] { continue }
      seen[folderCode] = true

      entries, err := w.resolver.ListFolder(ctx, folderCode)
      if err != nil {
          // log + continue — 1 folder fail không abort toàn bộ scan
          w.progressCh <- ScanProgress{Errors: []string{err.Error()}}
          continue
      }

      for _, entry := range entries {
          if entry.IsDir {
              queue = append(queue, source.FshareFileCode(entry.SourcePath))
          } else {
              w.processFile(ctx, entry)
          }
      }
  }
  ```
- [ ] BFS ưu tiên breadth (user thấy top-level folders populate trước)
- [ ] Cycle detection via `seen` map (fshare có shortcuts không? check HAR)

### 3. File Processing + Dedupe
- [ ] Method `processFile(ctx, entry)`:
  ```go
  // 1. Filter non-video MIME (reuse isVideoExt helper)
  if !isVideoExt(entry.Name) { return }

  // 2. Dedupe: check media_files WHERE path = entry.SourcePath
  existing, err := w.mediaRepo.GetByPath(ctx, entry.SourcePath)
  if existing != nil { return } // already scanned

  // 3. Parse filename → title + year + season + episode
  parsed := filename.Parse(entry.Name)

  // 4. TMDb match (rate-limited)
  w.tmdbThrottle.Wait(ctx)
  match, err := w.matcher.Match(ctx, parsed)

  // 5. Insert media_files
  mf := &model.MediaFile{
      LibraryID:  w.library.ID,
      Path:       entry.SourcePath,
      Size:       entry.Size,
      Duration:   0, // unknown for cloud
      Codec:      "",
      IsHDR:      false, // runtime probe not available
      DVProfile:  0,
      TMDbID:     match.TMDbID,
      MediaType:  match.MediaType,
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
      switch lib.SourceType {
      case model.SourceTypeLocal:
          return p.scanLocal(ctx, lib) // existing
      case model.SourceTypeFshare:
          walker := NewFshareWalker(FshareWalkerOpts{
              resolver:   p.resolvers.For(lib.SourceType),
              library:    lib,
              mediaRepo:  p.mediaRepo,
              matcher:    p.matcher,
              progressCh: p.progressCh,
          })
          return walker.Walk(ctx)
      default:
          return fmt.Errorf("unsupported source type: %s", lib.SourceType)
      }
  }
  ```

### 7. Rescan / Prune
- [ ] Rescan logic: rerun BFS, compare existing rows:
  - New path → insert
  - Existing path → skip (no metadata update phase này)
  - Row exists nhưng fshare returned not-found → mark `status = 'missing'` (DON'T delete — user có thể muốn restore)
- [ ] Prune trigger: admin button "Remove missing files" (Phase 06)

## Acceptance Criteria

- [ ] `scanner/fshare_walker_test.go` pass với mocked resolver + matcher:
  - [ ] BFS đúng thứ tự
  - [ ] Dedupe skip existing paths
  - [ ] Non-video extensions skipped
  - [ ] TMDb throttle respected (counter assertion)
  - [ ] Progress events broadcast đúng frequency
- [ ] Integration test: fake fshare resolver với 3-level tree (5 folders, 20 files) → verify all 20 inserted
- [ ] Pipeline integration: `ScanLibrary` dispatch đúng theo `source_type`
- [ ] Manual smoke test (dev env): scan fshare VIP account với 1 folder nhỏ (~50 files) → verify media_files rows
- [ ] Zero regression trên local scanner: existing tests pass

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
