# Plan U: Refactor HLS sang Pre-Generated Playlist (Jellyfin-style)

Created: 2026-04-07
Updated: 2026-04-08
Status: 🟡 PLANNED — single source of truth

> **Note:** Bản plan trước có 2 nhánh kiến trúc mâu thuẫn (batch encode vs long-running). Bản này đã rewrite hoàn toàn — chỉ còn 1 source of truth, không còn legacy block.

## Vấn đề hiện tại

Velox đang gặp 2 lớp bug khi stream phim bitrate cao (BluRay REMUX ~16 Mbps):

1. **PTS lệch giữa playlist EXTINF và segment thực** — combo `-copyts -start_at_zero` trong long-running ffmpeg + Velox `serveProjectedHLSPlaylist` hack tạo cumulative drift. Đến segment N, playlist nói thời điểm `T`, nhưng moof tfdt là `T-X`. MSE clip hết → buffered=empty → `bufferFullError` loop.
2. **MSE SourceBuffer quota đầy** — segment 12-15 MB × 15-20 segment vượt quota Chrome (~150 MB) → `bufferFullError` ngay cả khi PTS đúng.

Đã thử fix client (`maxBufferSize`, `backBufferLength`) và backend (`-avoid_negative_ts make_zero`) nhưng đều workaround. Phim Friends S01E14 vẫn stuck ở mốc 3:18.

## Giải pháp: Jellyfin-style architecture

Refactor theo cách Jellyfin làm: backend kiểm soát toàn bộ playlist + segment naming + ffmpeg lifecycle. Em đã verify từ [Jellyfin DynamicHlsController.cs](https://github.com/jellyfin/jellyfin/blob/master/src/Jellyfin.Api/Controllers/DynamicHlsController.cs) — toàn bộ behavior dưới đây copy chính xác từ Jellyfin, không tự đoán.

### Constants chính xác từ Jellyfin

| Constant | Value | Reference |
|---|---|---|
| Segment length | 6 giây | `state.SegmentLength` default |
| Seek threshold | `24 / segLength` segs (= 4 segs với 6s) | `segmentGapRequiringTranscodingChange` |
| Init segment filename | `{prefix}-1.{ext}` | `-hls_fmp4_init_filename` |
| Segment filename | `{prefix}{N}.{ext}` (N integer thuần) | `outputPrefix + "%d" + outputExtension` |
| Idle timeout | ~10 phút | `_transcodeManager` cleanup |
| ffmpeg combo | `-copyts -avoid_negative_ts disabled` (giữ PTS gốc, không shift) | per Jellyfin source |

### Decision logic (copy nguyên từ Jellyfin)

Khi client request segment N:
```
currentTranscodingIndex = file segment cao nhất đã có trên disk
gap = N - currentTranscodingIndex

if N == -1 (init segment):     → start fresh
if no current ffmpeg:          → start fresh
if N < currentTranscodingIndex: → start fresh (seek lùi)
if gap > 4 segments (= 24s):   → kill current + start fresh từ N
else:                          → wait current ffmpeg encode tới N (poll filesystem)
```

### Khác biệt kiến trúc

**Hiện tại (Velox):**
```
client → master.m3u8
  → ffmpeg với -hls_playlist_type event ghi m3u8 + segments
  → backend "project" m3u8 sang VOD bằng append segment ảo
  → serve client
```
Backend phụ thuộc vào ffmpeg ghi playlist. Mọi bug PTS đều bắt nguồn từ đây.

**Target (v2 — Jellyfin-style):**
```
client → master.m3u8
  → backend tự generate master (chỉ EXT-X-MEDIA + EXT-X-STREAM-INF, KHÔNG có EXTINF)
  → master deterministic, references tới video.m3u8 / audio0.m3u8 / ...
  → serve ngay

client → video.m3u8 (sub-playlist)
  → backend resolve session, prime ffmpeg từ session_start nếu cần
  → backend đọc Session.extinfMap (chỉ chứa segment đã encode + đã parse từ playlist file ffmpeg ghi)
  → render incremental EVENT-style playlist:
    ├─ Có ít nhất 1 segment → list các EXTINF đã có (contiguous từ session_start)
    ├─ Chưa hoàn chỉnh (chưa encode tới segment cuối) → EVENT mode, KHÔNG ENDLIST
    └─ Hoàn chỉnh (encoded tới lastSegNum) → VOD mode, có ENDLIST
  → serve với Cache-Control: no-store (player refetch trong EVENT mode)

client → seg_{N}.m4s
  → check disk
    ├─ có → serve
    └─ chưa → SessionManager.RequestSegment(N):
              if gap > 4 segs: kill old ffmpeg + start fresh từ N
              else: wait current ffmpeg encode tới N
              poll filesystem mỗi 100ms tới khi file xuất hiện
              → serve
```

### Lợi ích

- **Triệt tiêu bug PTS lệch:** Bug v1 đến từ combo (ffmpeg event playlist + project hack append segment ảo với EXTINF synthetic). v2 bỏ project hack hoàn toàn — backend chỉ render EXTINF cho segment **đã được ffmpeg encode**, lấy giá trị duration thật từ playlist file ffmpeg ghi ra. Không có segment ảo, không có EXTINF synthetic.
- **Triệt tiêu bug MSE quota gián tiếp:** Playlist incremental + EVENT mode → hls.js refetch định kỳ, biết chính xác phạm vi seekable hiện tại, không bị "hint" sai về segment chưa tồn tại.
- **Tua mượt hơn:** Filename ↔ time mapping cố định → seek nhỏ chỉ wait poll, seek lớn restart 1 ffmpeg + reload sub-playlist với MEDIA-SEQUENCE mới.
- **Dễ debug:** Segment N luôn ở vị trí `[N×6, (N+1)×6]` giây (xấp xỉ, segment có thể 5.8-6.5s tùy keyframe). Test bằng tay: `ffprobe seg_0050.m4s` → packet PTS đầu ≈ 300.0s.
- **Bỏ được hack:** `serveProjectedHLSPlaylist`, `projectPlaylistToVOD`, `WaitForSegment` polling logic phức tạp.

### Playlist ownership: ffmpeg là source of truth (KHÔNG placeholder synthetic)

Đây là quyết định quan trọng — em chốt rõ ràng để không mâu thuẫn về sau:

**Quy tắc tuyệt đối:** Backend **KHÔNG BAO GIỜ** sinh EXTINF cho segment chưa encode. Chỉ segment đã được ffmpeg encode + ghi playlist + parse vào `extinfMap` mới được xuất hiện trong playlist trả về client.

**Lý do:**
- HLS fmp4 segment duration KHÔNG luôn = `-hls_time` chính xác. ffmpeg cắt segment ở keyframe gần nhất → segment thực có thể là 5.808s, 6.467s, 6.000s tùy GOP structure.
- Nếu backend tự sinh EXTINF = 6.000 cho segment chưa encode, client sẽ thấy timing sai → khi fetch segment thật về, tfdt khác EXTINF → tái tạo bug v1.
- Player thường cache sub-playlist VOD và KHÔNG refetch sau khi `#EXT-X-ENDLIST` được thấy → mismatch sẽ stuck vĩnh viễn.

**Mô hình playlist đúng — "incremental EVENT-style với ENDLIST có điều kiện":**

Backend serve sub-playlist dựa **chỉ trên những gì đã encode**, dùng `#EXT-X-PLAYLIST-TYPE:EVENT` cho phép player refetch:

| Trạng thái session | Playlist render | `#EXT-X-ENDLIST`? |
|---|---|---|
| Mới bắt đầu, chưa encode segment nào | Empty playlist với `#EXT-X-PLAYLIST-TYPE:EVENT`, không có EXTINF | Không |
| Đã encode segment 0..30 (chưa hết phim) | EXTINF cho 0..30 (giá trị thật từ ffmpeg) | Không |
| Đã encode segment 0..227 (= hết phim, lastSegNum = floor(duration/segLength)) | EXTINF cho 0..227 (bao gồm partial segment cuối) | **Có** |

**Phân biệt EVENT vs VOD trong v2:**
- EVENT: player phải refetch playlist định kỳ (default ~3s) → khi backend update `extinfMap` với segment mới, lần refetch tiếp theo sẽ thấy.
- VOD: player KHÔNG refetch sau khi load lần đầu — chỉ dùng khi đã có ENDLIST.
- → v2 dùng EVENT cho mọi playlist chưa hoàn chỉnh, VOD (+ ENDLIST) chỉ khi ffmpeg đã encode tới segment cuối.

**Trường hợp seek ngoài range hiện tại:**

Khi user seek tới segment N >> currentSegmentOnDisk, backend cần đảm bảo playlist refetch sẽ chứa segment N. Có 2 sub-case:

**Case A: Seek nhỏ (gap ≤ 4 segs)** — ffmpeg đang encode tới gần đó
- Đợi ffmpeg encode tới N (poll filesystem)
- Sau khi encode xong, parse playlist file → merge vào `extinfMap`
- Client refetch playlist (do EVENT mode polls every ~3s) → thấy segment N → request segment N
- Latency: ~6s × gap (max 24s)

**Case B: Seek lớn (gap > 4 segs)** — restart ffmpeg
- Kill ffmpeg cũ
- Start ffmpeg mới với `-ss (N × segLength) -start_number N`
- ffmpeg ghi playlist file mới chỉ chứa segment N..N+M
- **KEY:** Backend parse file mới và **MERGE** vào `extinfMap` (không replace) — vẫn giữ EXTINF của segment 0..currentBeforeRestart
- → `extinfMap` có thể có "lỗ hổng": ví dụ segment 0..30 + segment 50..80 (gap 31..49 chưa encode)
- Backend render playlist: chỉ output **các range liên tục** từ segment 0, hoặc dùng `#EXT-X-DISCONTINUITY-SEQUENCE` để báo gap

**Render rule chi tiết:**

```go
func RenderMediaPlaylist(extinfMap map[int]float64, totalDuration, segLength float64, prefix, kind string) string {
    // Sort segment numbers
    segNums := sortedKeys(extinfMap)

    // Find longest contiguous prefix from 0
    // (player chỉ play được liên tục từ start của playlist nó nhận)
    contiguousEnd := -1
    for i, n := range segNums {
        if n == i {
            contiguousEnd = i
        } else {
            break
        }
    }

    // Build playlist với segments [0..contiguousEnd]
    sb := strings.Builder{}
    sb.WriteString("#EXTM3U\n#EXT-X-VERSION:6\n#EXT-X-TARGETDURATION:6\n")
    sb.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")
    sb.WriteString("#EXT-X-START:TIME-OFFSET=0\n") // Ngăn lỗi m3u8_js nhảy tới Live Edge
    sb.WriteString(fmt.Sprintf("#EXT-X-MAP:URI=\"%s%s_-1.mp4\"\n", prefix, kind))

    if contiguousEnd >= 0 {
        // Có ít nhất 1 segment
        for i := 0; i <= contiguousEnd; i++ {
            sb.WriteString(fmt.Sprintf("#EXTINF:%.6f,\n", extinfMap[i]))
            sb.WriteString(fmt.Sprintf("%s%s_%d.m4s\n", prefix, kind, i))
        }
    }

    // Determine if playlist is complete
    lastSegNum := int(math.Floor(totalDuration / segLength))
    isComplete := contiguousEnd == lastSegNum

    if isComplete {
        sb.WriteString("#EXT-X-ENDLIST\n")
    } else {
        // EVENT mode → player will refetch
        sb.WriteString("#EXT-X-PLAYLIST-TYPE:EVENT\n")
    }
    return sb.String()
}
```

**Player behavior với EVENT mode:**
- hls.js sẽ refetch playlist mỗi `targetduration / 2` (= 3s với targetduration=6)
- Khi backend update `extinfMap` với segment mới, lần refetch tiếp theo sẽ thấy
- Khi playlist có ENDLIST, hls.js chuyển sang VOD mode, không refetch nữa

**Vấn đề: contiguous prefix vs seek lớn**

Nếu user seek từ giây 0 → giây 600 (segment 100), trong khi mới có segment 0..30 đã encode:
- Sau seek, ffmpeg encode 100..130
- `extinfMap` = {0..30, 100..130}
- `contiguousEnd` = 30 (vì segment 31 không có)
- Playlist trả về client chỉ có segment 0..30

→ **Vấn đề:** Client không thấy segment 100 → không request → stuck.

**Fix: Per-position playlist (giống Jellyfin)**

Backend không serve "contiguous from 0" mà serve "contiguous around current playback position":

```go
// New approach: lấy từ ffmpeg's currentTranscodingIndex range
func RenderMediaPlaylist(
    extinfMap map[int]float64,
    sessionStartSegment int,    // segment ffmpeg đang encode TỪ đâu (sau seek)
    totalDuration, segLength float64,
    prefix, kind string,
) string {
    // Player cần biết playlist BẮT ĐẦU từ đâu để tính currentTime
    // Dùng MEDIA-SEQUENCE = sessionStartSegment
    sb.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", sessionStartSegment))
    sb.WriteString("#EXT-X-START:TIME-OFFSET=0\n")

    // Render contiguous từ sessionStartSegment
    for i := sessionStartSegment; ; i++ {
        if _, ok := extinfMap[i]; !ok { break }
        sb.WriteString(fmt.Sprintf("#EXTINF:%.6f,\n", extinfMap[i]))
        sb.WriteString(fmt.Sprintf("%s%s_%d.m4s\n", prefix, kind, i))
    }
    // ... ENDLIST nếu encoded tới hết
}
```

→ Mỗi lần seek lớn, client phải reload playlist (get URL mới với `?session_start=N`). Đây giống cách v1 hiện tại reload playlist khi `hlsStartOffset` thay đổi (xem flow resume bên dưới).

**Tóm lại:**
- Backend CHỈ serve EXTINF cho segment đã encode
- Empty playlist hợp lệ ban đầu (player sẽ refetch)
- EVENT mode tới khi encode hết, sau đó ENDLIST
- Seek lớn = client reload playlist với `session_start` mới
- KHÔNG có placeholder synthetic, KHÔNG có bug timing

### Các Edge Cases đã được bổ sung (Core Fixes)

Quá trình research bổ sung đã phát hiện và xử lý 5 edge cases quan trọng khớp với kiến trúc của Jellyfin:

1. **CPU/Disk Throttling:** Để tránh trường hợp FFmpeg ngốn 100% CPU và làm đầy ổ cứng ngay lập tức, `Session Manager` sẽ track khoảng cách giữa `currentSegmentOnDisk` và vị trí playback hiện tại. Nếu FFmpeg đã encode vượt giới hạn (vd: > 20 segments ahead), backend gọi `syscall.Kill(pid, syscall.SIGSTOP)` để "đóng băng" process và `syscall.SIGCONT` để resume khi player chạy gần tới.
2. **Client HLS.js Live Edge:** Do dùng `#EXT-X-PLAYLIST-TYPE:EVENT`, player sẽ hiểu lầm là luồng Live và tự động nhảy tới 3 segment cuối (Live Edge). Backend sẽ thêm dòng `#EXT-X-START:TIME-OFFSET=0` vào Sub-playlist để ép player bắt đầu từ đầu playlist, chống lỗi mất đoạn khi resume.
3. **Burn-in Subtitle Sync:** Seek xa với `-ss` có thể làm mất đồng bộ Subtitle dán cứng (PGS/Text). Việc phụ thuộc bằng `-copyts` sẽ đóng vai trò cốt lõi. Cần track kỹ trong testing Phase 2.
4. **Keyframe Alignment `videoCopy`:** Nếu stream là Video copy (`-c:v copy`), các segment duration có thể lớn hơn 6s. Backend hoàn toàn dùng map `timestamp` (EXTINF) chuẩn do FFmpeg xuất ra để tính thời gian Playlist, chứ không nhân chay theo index.
5. **Race Condition Đọc m3u8:** `ingestPlaylistFile` phải parse an toàn các dòng (thread-safe string parsing) đề phòng đọc ngay khúc FFmpeg đang ghi dở nửa dòng EOF.

### Trade-off

- Code backend phức tạp hơn ban đầu, ~900 LOC mới
- Refactor lớn → cần feature flag + rollback plan
- Migration window 2-4 tuần để verify

## Phạm vi

### IN scope
- Single-quality HLS với video copy (`videoCopy=true`) — path nóng nhất
- Single-quality HLS với re-encode (full transcode khi codec không compatible)
- **Multi-audio HLS với `#EXT-X-MEDIA`** — giữ behavior hiện tại (xem chi tiết Phase 2 + 4)
- Resume / seek trong và ngoài threshold
- Subtitle burn-in (PGS/VobSub image subtitles → re-encode video với overlay filter)
- Max-height quality cap

### OUT of scope (giữ nguyên kiến trúc cũ, không refactor)
- ABR multi-bitrate (`HLSABRMaster`) — đã hoạt động ổn cho pretranscode
- Pretranscode (file đã encode sẵn trên disk)
- Direct play (không qua transcoder)

## Cache key strategy

Đây là phần plan cũ mơ hồ — em chốt chính xác ở đây.

**Session cache key (tuple):**
```go
type SessionKey struct {
    StreamSessionID string  // 12-char hex per viewer (giống v1, từ POST /api/playback/{id}/info)
    MediaID         int64   // movie hoặc episode
    FileID          int64   // file_id (multi-version support)
    SubtitleStreamIdx int   // -1 nếu không burn-in, else stream index của image sub
    VideoCopy       bool    // true = -c:v copy, false = re-encode
    MaxHeight       int     // 0 = nguyên gốc, else 480/720/1080
    // KHÔNG bao gồm: startOffset, audio track (xem dưới)
}
```

**Tại sao PHẢI include `StreamSessionID` (per-viewer isolation):**
- v1 hiện đã encode `streamSessionID` vào prefix tại [transcoder.go:291](backend/internal/transcoder/transcoder.go#L291) (`hlsPrefix`) — mỗi viewer có session riêng.
- Nếu v2 gom mọi viewer cùng `(media, file, sub, vc, height)` vào 1 session chung:
  - User A đang xem giây 200, ffmpeg encode tới giây 250
  - User B mở phim, resume giây 800 → gap > 4 segs → kill ffmpeg + restart từ 800
  - User A bị mất segment đang play → stall
- → **Đây là regression so với v1**, không acceptable.
- v2 giữ nguyên isolation per-viewer: mỗi `streamSessionID` = 1 Session struct riêng, 1 ffmpeg process riêng, 1 folder riêng.
- Trade-off: 2 viewer cùng phim = 2× CPU/disk. Acceptable cho self-hosted (Velox không phải Netflix).

**Tại sao KHÔNG include `startOffset`:**
- Trong v2, `startOffset` chỉ là **gợi ý vị trí khởi đầu** cho ffmpeg session đầu tiên — không ảnh hưởng segment naming.
- Segment N **luôn** = giây `[N×6, (N+1)×6]` của phim, bất kể session bắt đầu encode từ đâu.
- Khi user resume tại 81s, backend tính `requestedSegment = 81 / 6 = 13`, gọi `RequestSegment(13)`. Session manager start ffmpeg với `-ss 78 -start_number 13`.

### Resume flow trong v2 (chốt rõ với `WatchPage` hiện tại)

WatchPage hiện tại có 3 mảnh logic liên quan ([WatchPage.tsx:101](webapp/src/pages/WatchPage.tsx#L101), [WatchPage.tsx:593](webapp/src/pages/WatchPage.tsx#L593), [WatchPage.tsx:765](webapp/src/pages/WatchPage.tsx#L765)):
- `buildHlsSessionUrl(baseUrl, startOffset)` — append `?start=X.XXX` vào URL
- `requestHlsSessionReload(targetTime)` — destroy hls.js + setHlsStartOffset → trigger reload với URL mới
- Effect tái tạo HLS instance khi `hlsStartOffset` thay đổi

**v2 phải tương thích với 3 mảnh trên** (vì plan nói KHÔNG sửa WatchPage). Cách wire:

**1. Master URL từ backend:**
```
POST /api/playback/{id}/info
→ response.hls_v2_url = "/api/stream/v2/{id}/hls/master.m3u8?ssid=xxx&api_key=yyy"
```

WatchPage nhận URL này qua `streamUrls.hls`, rồi:
```ts
const sessionOffset = hlsStartOffset ?? 0
const finalUrl = sessionOffset > 0
  ? buildHlsSessionUrl(streamUrls.hls, sessionOffset)  // append ?start=81.247
  : streamUrls.hls
```
→ Cuối cùng request là: `/api/stream/v2/{id}/hls/master.m3u8?ssid=xxx&api_key=yyy&start=81.247`

**2. Backend `Master()` handler parse `?start=`:**
```go
func (h *StreamV2Handler) Master(w, r) {
    mediaID := parseID(r)
    streamSessionID := streamSessionIDFromValues(r.URL.Query())
    startOffset := parseStartOffset(r.URL.Query().Get("start"))  // 0 nếu không có

    // Compute starting segment number from offset
    sessionStartSeg := int(math.Floor(startOffset / hls.SegLength))  // = 13 với 81.247

    // ...
    // Note: this snippet (Resume flow section) calls GetOrCreate with file only
    // for brevity. The canonical signature accepts audioTracks as well — see
    // the Phase 4 Master() snippet below for the full call.
    sess, err := h.mgr.GetOrCreate(r.Context(), key, file, audioTracks)
    // Kick off ffmpeg from sessionStartSeg if no current job exists, or
    // restart it if current job is too far away (Jellyfin gap logic inside).
    if err := sess.PrimeFromSegment(r.Context(), sessionStartSeg); err != nil { ... }

    // Master playlist tự body deterministic, KHÔNG có sessionStartSeg
    master := hls.GenerateMasterPlaylist(audioTracks, sess.Prefix(), estimatedBandwidth)

    // BUT — sub-playlist URI references trong master phải có ?session_start={N}
    // để khi player fetch sub-playlist, backend biết MEDIA-SEQUENCE start ở đâu
    rewritten := rewriteMasterPlaylistURIs(master, r.URL.Query())  // forward all query params + session_start
    serveText(w, rewritten, "application/vnd.apple.mpegurl")
}
```

**3. Sub-playlist URL có `session_start`:**
```
master.m3u8 references:
  video.m3u8?ssid=xxx&api_key=yyy&session_start=13
  audio0.m3u8?ssid=xxx&api_key=yyy&session_start=13
```

Backend `serveMediaPlaylist` parse `session_start` → render với `MEDIA-SEQUENCE=13`, segment URI `{prefix}video_13.m4s`, `_14.m4s`, ...

**4. Khi user seek lớn (gap > 4 segs):**
- WatchPage gọi `requestHlsSessionReload(targetTime)`
- → `setHlsStartOffset(targetTime)` → effect tái tạo HLS instance với URL mới có `?start=newOffset`
- → backend nhận request master mới với `start={new}` → tính sessionStartSeg mới → kill ffmpeg cũ + start fresh
- → master mới references sub-playlist với `session_start={new}`
- → player load playlist mới → request segment từ vị trí mới

**5. Khi user seek nhỏ (gap ≤ 4 segs):**
- hls.js tự handle: chỉ request segment kế tiếp, không reload playlist
- Backend ffmpeg đang chạy → ffmpeg encode tới đó → backend serve khi sẵn sàng
- KHÔNG đụng `hlsStartOffset` ở client

**Method `Session.PrimeFromSegment` (replace `RequestInitSegment`):**
```go
// PrimeFromSegment ensures ffmpeg is encoding from a position that covers segNum.
// - If no current job: start ffmpeg from segNum
// - If current job's start <= segNum and gap acceptable: no-op (current job will reach it)
// - Otherwise: kill + restart from segNum
// Also ensures init segment file exists.
func (s *Session) PrimeFromSegment(ctx context.Context, segNum int) error
```

→ Đây là API chuẩn dùng bởi cả Master handler (lúc user mở phim) và segment handler (lúc seek lớn).

**Backend KHÔNG cache session theo `startOffset`:**
- 1 session per `(streamSessionID, mediaID, fileID, sub, vc, height)` tuple
- Mỗi lần WatchPage reload với `?start=` mới, **vẫn cùng session** (cùng streamSessionID), nhưng ffmpeg có thể bị restart bởi `PrimeFromSegment(newSeg)`
- `extinfMap` được merge cumulative qua các restart đó (segment cũ vẫn được nhớ)
- Folder không bị xóa, segment files tích lũy → cleanup khi idle 10 phút

**Resume KHÔNG cause "vòng start sai rồi restart":**
- Nếu `Master()` luôn gọi `PrimeFromSegment(sessionStartSeg)` ngay từ đầu, ffmpeg start đúng vị trí ngay lần đầu
- Không có giai đoạn "start từ 0 rồi mới biết cần seek tới 81s"

**Tại sao KHÔNG include `audio track`:**
- Trong v2, mỗi session encode **tất cả audio tracks** dùng `#EXT-X-MEDIA` (giống `GenerateHLSWithAudio` hiện tại).
- Audio switching ở client = chỉ đổi audio rendition trong playlist, không restart ffmpeg.
- Một session phục vụ mọi audio track của file đó.

**Session prefix on disk:**
```go
prefix := fmt.Sprintf("v2_ss%s_%d_f%d_si%d_vc%d_h%d_",
    key.StreamSessionID,
    key.MediaID, key.FileID, key.SubtitleStreamIdx,
    boolToInt(key.VideoCopy), key.MaxHeight)
// e.g. "v2_ss934dd9d39ea4_163_f163_si-1_vc1_h0_"
// → segments: v2_ss934..._video_0.m4s, v2_ss934..._video_1.m4s, ...
// → init:     v2_ss934..._video_-1.mp4
// → audio:    v2_ss934..._audio0_0.m4s, v2_ss934..._audio0_-1.mp4, ...
```

**Cache invalidation:**
- Idle timeout 10 phút → cleanup session, xóa folder
- Subtitle change → key thay đổi (cùng streamSessionID, khác sub idx) → tạo session mới
- Quality change → key thay đổi → tạo session mới
- Audio switch → KHÔNG đổi key, dùng lại session
- New playback (POST /api/playback/{id}/info trả streamSessionID mới) → session mới, session cũ tự GC sau 10 phút

**Lifecycle & GC:**
- Manager track sessions theo `(streamSessionID, mediaID)` → 1 viewer mở 1 phim = 1 session
- 1 viewer mở 2 phim khác nhau = 2 session
- 2 viewer mở cùng phim = 2 session độc lập (KHÔNG share)
- WebSocket disconnect → đánh dấu lastAccess sớm hơn để GC nhanh
- Cap số session đồng thời: dùng chung semaphore với v1 transcoder (`maxConcurrent`)

## Multi-audio architecture (chi tiết)

Đây là phần plan cũ thiếu — em viết rõ.

### Master playlist với `#EXT-X-MEDIA`

Backend tự generate (không phụ thuộc ffmpeg):

```m3u8
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-INDEPENDENT-SEGMENTS

#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="eng",DEFAULT=YES,AUTOSELECT=YES,URI="audio_0.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="Vietnamese",LANGUAGE="vie",DEFAULT=NO,AUTOSELECT=YES,URI="audio_1.m3u8"

#EXT-X-STREAM-INF:BANDWIDTH=8000000,CODECS="avc1.640028,mp4a.40.2",AUDIO="audio"
video.m3u8
```

### Sub-playlist `video.m3u8` (incremental, EVENT-style)

Backend render từ `Session.extinfMap` — chỉ chứa segment đã được ffmpeg encode + parse. EXTINF lấy giá trị thật từ playlist file ffmpeg ghi (không phải synthetic).

**Trạng thái 1: Mới bắt đầu, chưa có segment nào**
```m3u8
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-MAP:URI="v2_ss934..._video_-1.mp4"
```
(không có EXTINF nào — player sẽ refetch sau ~3s)

**Trạng thái 2: ffmpeg đã encode segment 0..30 (chưa hết phim)**
```m3u8
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-TARGETDURATION:7
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-MAP:URI="v2_ss934..._video_-1.mp4"
#EXTINF:5.808000,
v2_ss934..._video_0.m4s
#EXTINF:6.000000,
v2_ss934..._video_1.m4s
#EXTINF:6.467000,
v2_ss934..._video_2.m4s
... (30 entries với EXTINF đúng từ ffmpeg, KHÔNG phải 6.000 cứng)
#EXTINF:6.000000,
v2_ss934..._video_30.m4s
```
(không có ENDLIST — player tiếp tục refetch)

**Trạng thái 3: ffmpeg đã encode hết phim (segment 0..227)**
```m3u8
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-TARGETDURATION:7
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-MAP:URI="v2_ss934..._video_-1.mp4"
#EXTINF:5.808000,
v2_ss934..._video_0.m4s
... (228 entries)
#EXTINF:5.808000,
v2_ss934..._video_227.m4s
#EXT-X-ENDLIST
```
(VOD + ENDLIST → player ngừng refetch, switch sang VOD mode)

**Trạng thái 4: User resume tại giây 81 (segment 13), ffmpeg restart từ seg 13, đã encode 13..40**
```m3u8
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-TARGETDURATION:7
#EXT-X-MEDIA-SEQUENCE:13
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-MAP:URI="v2_ss934..._video_-1.mp4"
#EXTINF:6.467000,
v2_ss934..._video_13.m4s
#EXTINF:6.000000,
v2_ss934..._video_14.m4s
... (segment 13..40)
```
(MEDIA-SEQUENCE = 13 = sessionStartSeg lấy từ `?session_start=13` query param)

### Sub-playlist `audio0.m3u8` (per audio track, cùng pattern)

Cùng pattern như `video.m3u8`, chỉ khác:
- `EXT-X-MAP:URI="v2_ss934..._audio0_-1.mp4"`
- Filename `v2_ss934..._audio0_{N}.m4s`
- EXTINF có thể khác chút so với video (audio frame boundary khác keyframe boundary)
- Cùng `MEDIA-SEQUENCE` với video (đảm bảo sync)
- Cùng EVENT/VOD state với video

### ffmpeg multi-output command

Một ffmpeg invocation encode video + tất cả audio tracks (giống `runMultiOutputHLS` hiện tại):

```
ffmpeg
  -ss {startSec}
  -i input.mkv
  -copyts
  -avoid_negative_ts disabled
  -map_metadata -1 -map_chapters -1
  -max_muxing_queue_size 4096

  # Video output
  -map 0:v:0 -c:v copy
  -f hls -hls_segment_type fmp4 -hls_time 6 -hls_list_size 0
  -hls_playlist_type event              # EVENT để ffmpeg flush EXTINF sau mỗi segment, backend đọc được mid-encode
  -start_number {startSegNum}
  -hls_fmp4_init_filename {prefix}video_-1.mp4
  -hls_segment_filename {dir}/{prefix}video_%d.m4s
  {dir}/{prefix}video.m3u8   # ffmpeg ghi playlist EVENT vào disk thật, backend đọc EXTINF mid-encode

  # Audio 0 output
  -map 0:a:0 -c:a aac -b:a 192k -ac 2
  -f hls -hls_segment_type fmp4 -hls_time 6 -hls_list_size 0
  -hls_playlist_type event
  -start_number {startSegNum}
  -hls_fmp4_init_filename {prefix}audio0_-1.mp4
  -hls_segment_filename {dir}/{prefix}audio0_%d.m4s
  {dir}/{prefix}audio0.m3u8

  # Audio 1 output
  -map 0:a:1 -c:a aac -b:a 192k -ac 2
  ... (tương tự, ghi {prefix}audio1.m3u8 với -hls_playlist_type event)
```

**Key points:**
- `-hls_playlist_type event` → ffmpeg flush playlist sau mỗi segment, backend có thể đọc mid-encode để biết segment nào đã sẵn sàng + EXTINF của nó.
- Khi ffmpeg encode xong toàn bộ và exit cleanly, nó tự thêm `#EXT-X-ENDLIST` vào playlist file.
- Khi ffmpeg bị kill (do seek lớn), playlist file vẫn còn các EXTINF đã ghi tới đó — backend đã merge các giá trị này vào `extinfMap`, không bị mất.
- Backend đọc playlist file → parse → merge vào `Session.extinfMap` (xem "Playlist ownership"). Backend KHÔNG serve trực tiếp file ffmpeg cho client — backend tự render playlist từ `extinfMap` (xem các "Trạng thái 1-4" ở section sub-playlist).
- `-start_number` đảm bảo segment N có filename khớp với playlist position N.
- Single ffmpeg process, single semaphore slot, single read input file.
- Audio switching ở client không restart ffmpeg vì cả 2 audio track đều có sẵn segment.

### Audio switching at runtime

Khi user đổi audio track:
- Client (hls.js) đổi `audioTrack` index → load `audio_1.m3u8` thay vì `audio_0.m3u8`
- Backend serve segment audio mới từ session đã có
- KHÔNG cần restart ffmpeg, KHÔNG cần reload master

## 7 Phase

### Phase 1: Foundation utilities (zero risk)

**Mục tiêu:** Tạo package mới `internal/hls/` với pure functions, có test, chưa wire vào đâu.

**Files mới:**
- `internal/hls/naming.go`
  - `SessionKey` struct (xem "Cache key strategy")
  - `BuildPrefix(key SessionKey) string`
  - `SegmentNumber(time, segLength float64) int`
  - `SegmentTimeRange(num int, segLength float64) (start, end float64)`
  - `SegmentFilename(prefix, kind string, idx int) string` (kind = "video" | "audio0" | ...)
  - `InitFilename(prefix, kind string) string` (= `{prefix}{kind}_-1.mp4`)
  - `MediaPlaylistFilename(prefix, kind string) string` (= `{prefix}{kind}.m3u8`)
- `internal/hls/master.go`
  - `GenerateMasterPlaylist(audioTracks []model.AudioTrack, prefix string, videoBandwidth int) string`
    - Chỉ chứa `#EXT-X-MEDIA` + `#EXT-X-STREAM-INF`, không có EXTINF — 100% deterministic
    - Reference tới `{prefix}video.m3u8` và `{prefix}audio{i}.m3u8`
- `internal/hls/extinf.go`
  - `ParseMediaPlaylist(content []byte) (map[int]float64, error)` — đọc playlist ffmpeg ghi ra, return `segNum → duration`
  - `MergeExtinf(existing, new map[int]float64)` — merge cumulative qua nhiều ffmpeg restart, mutate `existing` in place
  - `RenderMediaPlaylist(opts RenderOpts) string` — sinh playlist từ extinfMap, KHÔNG tự sinh placeholder
  - `RenderOpts` struct:
    ```go
    type RenderOpts struct {
        ExtinfMap         map[int]float64
        SessionStartSeg   int      // MEDIA-SEQUENCE = đây
        TotalDuration     float64  // để biết segment cuối là số mấy
        SegLength         float64
        Prefix            string
        Kind              string   // "video" | "audio0" | ...
    }
    ```
  - Render rule: contiguous từ `SessionStartSeg`, dừng ở segment đầu tiên không có trong map. ENDLIST chỉ khi đã encode tới `floor(TotalDuration/SegLength)`. EVENT mode khi chưa hoàn chỉnh.
- Tests:
  - `internal/hls/naming_test.go`
  - `internal/hls/master_test.go`
  - `internal/hls/extinf_test.go`

**Edge cases test:**
- `BuildPrefix` với streamSessionID rỗng → error
- `SegmentNumber(0, 6) → 0`, `SegmentNumber(5.99, 6) → 0`, `SegmentNumber(6.0, 6) → 1`
- `ParseMediaPlaylist` với input ffmpeg-generated thật (em sẽ paste sample từ run thật làm fixture)
- `MergeExtinf`: existing={0..30}, new={50..80} → merged map có cả 2 dải (không lose data, không dùng range cũ overwrite range mới)
- `MergeExtinf` overlap: existing={0..30}, new={25..40} → segment 25..30 dùng giá trị từ `new` (lần encode mới nhất là source of truth)
- `RenderMediaPlaylist` với map rỗng + sessionStartSeg=0 → playlist EVENT mode không có EXTINF nào
- `RenderMediaPlaylist` với map={0..30} + sessionStartSeg=0 + totalDuration=1367.808s → 31 EXTINF, EVENT mode (chưa complete)
- `RenderMediaPlaylist` với map={0..227} + sessionStartSeg=0 + totalDuration=1367.808s → 228 EXTINF (segment cuối ≈ 5.808s từ ffmpeg), VOD mode + ENDLIST
- `RenderMediaPlaylist` với map={0..30, 50..80} + sessionStartSeg=0 → playlist chỉ chứa 0..30 (contiguous từ start, dừng ở gap)
- `RenderMediaPlaylist` với map={0..30, 50..80} + sessionStartSeg=50 → playlist MEDIA-SEQUENCE=50, chứa 50..80
- Master với 0/1/5 audio tracks

**Acceptance:** Unit tests pass, > 90% coverage. Test fixture cho parse playlist phải dùng output ffmpeg thật (em sẽ generate 1 file mẫu trên NAS rồi commit vào testdata).

**Risk:** Zero. Không ai gọi.

**Rollback:** Xóa package.

### Phase 2: Segment encoder (medium risk)

**Mục tiêu:** Hàm wrap ffmpeg invocation cho 1 session encoding job. Tách biệt khỏi `runHLSFFmpeg` cũ.

**Files mới:**
- `backend/internal/transcoder/v2_encoder.go`
  - `StartHLSV2Encoder(ctx, opts) (*exec.Cmd, error)` — start ffmpeg, return process handle
  - `Opts` struct: `InputPath, OutputDir, Prefix, StartSegNum, AudioTracks, VideoCopy, SubtitleStreamIdx, MaxHeight, HwAccel, SegLength`
- Test bằng tay với 3 phim:
  - h264 SDR (Friends REMUX)
  - HEVC HDR (4K phim test)
  - Phim có 2+ audio tracks

**ffmpeg command building (single + multi audio):**

Em **giữ nguyên kiến trúc multi-output** của `runMultiOutputHLS` hiện tại, chỉ đổi 4 thứ:
1. **Dùng `-hls_playlist_type event`** (KHÔNG phải `vod`). Lý do: backend cần đọc playlist file ffmpeg ghi ra trong khi ffmpeg vẫn đang encode (chưa xong), để parse EXTINF và merge vào `extinfMap`. EVENT mode đảm bảo ffmpeg flush playlist sau mỗi segment, KHÔNG đợi đến khi hoàn tất rồi mới ghi 1 lần. Khi ffmpeg encode xong cuối phim, nó sẽ tự ghi `#EXT-X-ENDLIST` vào file (EVENT → ENDLIST). Nếu ffmpeg bị kill giữa chừng, file vẫn còn các EXTINF đã ghi tới đó (đủ để backend parse). VOD mode chỉ ghi playlist 1 lần khi ffmpeg exit cleanly → backend không có gì để đọc trong khi encode.
2. Thêm `-start_number {startSegNum}` cho mỗi output
3. Đổi `-hls_segment_filename` pattern theo convention v2 (`{prefix}{kind}_%d.m4s`)
4. **Playlist file output ghi vào disk thật** (`{prefix}{kind}.m3u8`), KHÔNG `/dev/null` — backend cần đọc EXTINF (xem "Playlist ownership")

**Lưu ý quan trọng về EVENT mode trong v2 vs v1:**
- v1 dùng `-hls_playlist_type event` + serve trực tiếp file ffmpeg + project hack append segment ảo → bug PTS lệch
- v2 dùng `-hls_playlist_type event` + backend đọc + parse + merge vào `extinfMap` + render lại với rule mới → KHÔNG có segment ảo
- → Cùng flag, khác cách dùng. Bug v1 không phải do `event` mode, mà do project hack ở backend.

**Init filename:**
- `-hls_fmp4_init_filename {prefix}{kind}_-1.mp4` (Jellyfin convention với segment id -1)

**Verify chính xác:**
- Sau khi encode 30s test, chạy `ffprobe -show_packets -select_streams v:0 -read_intervals 0%+5 v2_..._video_50.m4s`
- PTS của packet đầu ≈ 300.0s (= 50 × 6, ± vài ms tùy keyframe)
- Compare với segment 51 → không gap, không overlap > 100ms
- Đọc `v2_..._video.m3u8` (file ffmpeg ghi) → kiểm tra `#EXTINF:` cumulative khớp với PTS thực
- Đọc playlist trong khi ffmpeg vẫn đang chạy (mid-encode) → file phải có các EXTINF cho segment đã encode tới thời điểm đó

**Acceptance:**
- ffmpeg command build đúng cho cả single audio và multi audio
- Encode test 30s video → segment file + playlist file tồn tại trên disk
- Playlist EXTINF khớp với PTS từ ffprobe (sai số < 100ms)
- Mid-encode read playlist → có EXTINF cho segment đã ghi (test bằng cách `tail -f` playlist file trong khi ffmpeg chạy)
- Không leak process khi context cancel
- Test fixture playlist sẽ commit vào `internal/hls/testdata/` cho Phase 1 unit test parser (lấy 2 fixture: 1 playlist incomplete đang chạy + 1 playlist complete với ENDLIST)

**Risk:** Medium. Function mới, chưa wire vào HTTP path.

**Rollback:** Xóa file.

### Phase 3: Session manager (HIGHEST risk)

**Mục tiêu:** Quản lý ffmpeg lifecycle per-session, on-demand segment serving.

**Files mới:**
- `backend/internal/streamv2/session.go`
  ```go
  type Session struct {
      Key            hls.SessionKey
      OutputDir      string
      InputPath      string
      Duration       float64
      AudioTracks    []model.AudioTrack
      SegLength      float64

      mu             sync.Mutex
      cmd            *exec.Cmd
      cmdCancel      context.CancelFunc
      startSegment   int        // segment ffmpeg bắt đầu
      lastAccess     time.Time
      log            *slog.Logger

      // EXTINF tracking — merged across multiple ffmpeg restarts.
      // Key: segment number (0-based, global). Value: duration in seconds.
      extinfMu       sync.RWMutex
      extinfMap      map[int]map[string]float64  // [segNum][kind] -> duration
  }

  // Prefix returns the on-disk filename prefix for this session.
  func (s *Session) Prefix() string { return hls.BuildPrefix(s.Key) }

  // PrimeFromSegment ensures ffmpeg is encoding from a position that covers segNum.
  // Used by Master handler (initial load + seek) and segment handler (large seek fallback).
  // - If no current job: start ffmpeg from segNum
  // - If current job's start <= segNum AND gap (segNum - currentEncoded) <= 4: no-op
  // - Otherwise: kill current + restart from segNum
  // Also ensures init segment {prefix}{kind}_-1.mp4 exists for all kinds.
  func (s *Session) PrimeFromSegment(ctx context.Context, segNum int) error

  // RequestSegment ensures segment {kind}_{N} exists on disk.
  // Blocks until ready or ctx cancelled. Uses Jellyfin gap logic internally.
  func (s *Session) RequestSegment(ctx context.Context, kind string, segNum int) error

  // GetExtinfSnapshot returns a copy of the current EXTINF map for serving playlist.
  // Caller can pass this to hls.RenderMediaPlaylist().
  func (s *Session) GetExtinfSnapshot(kind string) map[int]float64

  // SessionStartSegment returns the segment number that current ffmpeg job
  // started encoding from. Used as MEDIA-SEQUENCE in sub-playlist.
  func (s *Session) SessionStartSegment() int

  func (s *Session) Touch()  // update lastAccess
  func (s *Session) Close()  // kill ffmpeg + cleanup files

  // Internal: after ffmpeg writes/updates a media playlist file, parse it
  // and merge into extinfMap. Called by waitForSegmentOnDisk poll loop.
  func (s *Session) ingestPlaylistFile(kind string) error
  ```
- `backend/internal/streamv2/manager.go`
  ```go
  type Manager struct {
      transcoder  *transcoder.Transcoder
      mediaSvc    MediaService

      mu          sync.Mutex
      sessions    map[hls.SessionKey]*Session
      gcInterval  time.Duration  // 1 phút
      idleTimeout time.Duration  // 10 phút
  }

  func (m *Manager) GetOrCreate(ctx context.Context, key hls.SessionKey, file *model.MediaFile, audioTracks []model.AudioTrack) (*Session, error)
  func (m *Manager) Get(key) *Session
  func (m *Manager) Close(key)
  func (m *Manager) gcLoop()  // background goroutine
  ```
- `backend/internal/streamv2/session_test.go` — test concurrency edge cases

**RequestSegment logic:**
```go
func (s *Session) RequestSegment(ctx, kind, segNum) error {
    s.Touch()

    segPath := s.segPath(kind, segNum)
    if fileExists(segPath) {
        return nil
    }

    s.mu.Lock()
    needRestart := false
    switch {
    case s.cmd == nil:
        needRestart = true
    case segNum < s.startSegment:
        needRestart = true  // seek lùi
    case segNum > s.currentSegmentOnDisk()+gapThreshold:
        needRestart = true  // gap > 4 segments (Jellyfin formula)
    }

    if needRestart {
        s.killFFmpeg()
        if err := s.startFFmpegFrom(segNum); err != nil {
            s.mu.Unlock()
            return err
        }
    }
    s.mu.Unlock()

    // Poll filesystem
    return s.waitForSegmentOnDisk(ctx, segPath, 90*time.Second)
}
```

**`currentSegmentOnDisk()`:** scan thư mục, đếm số segment file đã có cho `kind`. Cheap operation (cached với invalidate khi ffmpeg ghi mới).

**Concurrency safety:**
- 1 mutex per session (coarse-grained)
- Mỗi RequestSegment lock ngắn (chỉ khi check restart), unlock trước khi poll filesystem
- ffmpeg process được track qua `s.cmd`, kill an toàn qua `s.cmdCancel()`
- Manager.GetOrCreate dùng double-check locking

**Throttle & GC logic:**
- Cần thêm 1 vòng lặp (poll / trigger) trong session để check buffer size: `currentEncodedSeg - lastRequestedSeg`.
- Nếu buffer > 20 (khoảng 2 phút), gọi `syscall.Kill(s.cmd.Process.Pid, syscall.SIGSTOP)`.
- Khi client request segment mà gap rút xuống dưới 10, gửi `syscall.SIGCONT` để resume.

**GC logic:**
```go
func (m *Manager) gcLoop() {
    ticker := time.NewTicker(m.gcInterval)
    for range ticker.C {
        m.mu.Lock()
        for key, sess := range m.sessions {
            if time.Since(sess.lastAccess) > m.idleTimeout {
                sess.Close()
                delete(m.sessions, key)
                os.RemoveAll(sess.OutputDir)
            }
        }
        m.mu.Unlock()
    }
}
```

**Acceptance:**
- Test play tuần tự seg 0..30 → mỗi segment serve trong < 100ms sau khi encoder warm
- Test seek nhỏ (gap < 4): instant (hit cache)
- Test seek lớn (gap > 4): ~600ms (kill + restart + first segment)
- Test seek liên tục 5 lần trong 10s → không leak ffmpeg process
- Test cancel: client disconnect → ffmpeg kill trong 2s
- Test concurrent: 2 client request 2 segment khác nhau cùng session → cả 2 OK
- Test audio switch: request `audio_1_50.m4s` sau khi đang play `video_50.m4s` → instant (cùng ffmpeg job)
- Test idle GC: idle 11 phút → folder bị xóa

**Risk:** HIGHEST. Concurrency edge cases, file race, ffmpeg lifecycle, leak risk.

**Rollback:** Xóa package `streamv2`. HTTP handlers v2 (Phase 4) chưa exist nên không có gì gọi.

### Phase 4: HTTP handlers v2 (low risk, behind feature flag)

**Mục tiêu:** Endpoint mới song song với cũ, bật bằng env `VELOX_HLS_V2=true`.

**Files mới:**
- `backend/internal/handler/streamv2.go`
  ```go
  type StreamV2Handler struct {
      mgr      *streamv2.Manager
      mediaSvc MediaService
      log      *slog.Logger
  }

  // Routes:
  //   GET /api/stream/v2/{id}/hls/master.m3u8
  //   GET /api/stream/v2/{id}/hls/{file}     (catch-all dispatcher)
  func (h *StreamV2Handler) Master(w, r)
  func (h *StreamV2Handler) Serve(w, r)  // dispatcher cho mọi file khác master
  ```

**Files sửa (đã verify đường dẫn):**
- `backend/cmd/server/server_app.go` ([server_app.go:354](backend/cmd/server/server_app.go#L354)) — `initHandlers()` thêm `app.handlers.streamV2 = handler.NewStreamV2Handler(...)`. Init `streamv2.Manager` trong `initServices()` (tương tự `app.services.transcoder` ở line 268). KHÔNG sửa `main.go` (chỉ gọi `newServerApp()` rồi start).
- `backend/cmd/server/server_routes.go` ([server_routes.go:235](backend/cmd/server/server_routes.go#L235)) — đăng ký routes mới sau routes v1, gated bởi `app.cfg.HLSV2Enabled`
- `backend/internal/config/config.go` — thêm field `HLSV2Enabled bool`, đọc từ env `VELOX_HLS_V2`

**Routes (chỉ đăng ký nếu flag on):**
```go
if app.cfg.HLSV2Enabled {
    mux.HandleFunc("GET /api/stream/v2/{id}/hls/master.m3u8", app.handlers.streamV2.Master)
    mux.HandleFunc("GET /api/stream/v2/{id}/hls/{file}", app.handlers.streamV2.Serve)
}
```

**`Serve` dispatcher** (dispatch theo filename suffix/pattern):
```go
func (h *StreamV2Handler) Serve(w, r) {
    file := r.PathValue("file")
    // file vd: "v2_ss934..._video.m3u8" hoặc "v2_ss934..._video_50.m4s" hoặc "v2_ss934..._video_-1.mp4"
    switch {
    case strings.HasSuffix(file, ".m3u8"):
        h.serveMediaPlaylist(w, r, file)
    case strings.Contains(file, "_-1.mp4"):
        h.serveInitSegment(w, r, file)
    case strings.HasSuffix(file, ".m4s"):
        h.serveSegment(w, r, file)
    default:
        http.NotFound(w, r)
    }
}
```

`serveMediaPlaylist`:
- Resolve session từ `streamSessionID` query param (giống v1)
- Parse `session_start` query param (do master URL rewrite ghi vào)
- Gọi `sess.PrimeFromSegment(ctx, sessionStart)` để đảm bảo ffmpeg đang encode từ vị trí đúng (no-op nếu đã chạy đúng)
- Compose playlist: `hls.RenderMediaPlaylist(RenderOpts{ExtinfMap: sess.GetExtinfSnapshot(kind), SessionStartSeg: sessionStart, ...})`
- Rewrite segment URI thêm query auth (`api_key`, `token`, etc.) — giống `rewriteHLSPlaylist` v1
- Serve với `Cache-Control: no-store` (vì EVENT mode cần refetch)

`serveSegment`:
- Resolve session
- Parse `kind` + `segNum` từ filename
- Gọi `sess.RequestSegment(ctx, kind, segNum)` (block tới khi có hoặc timeout 90s)
- Serve file

`serveInitSegment`:
- Resolve session
- Init segment được tạo bởi ffmpeg invocation đầu tiên (ngay sau `PrimeFromSegment` khi Master/MediaPlaylist được fetch)
- Nếu file chưa tồn tại → poll filesystem (timeout 30s) trong khi ffmpeg đang start
- Serve file

**`StreamV2Handler` dependencies (verified service methods):**
```go
type StreamV2Handler struct {
    mgr         *streamv2.Manager
    streamSvc   *service.StreamService    // for GetPrimaryFile (stream.go:260)
    audioSvc    *service.AudioTrackService // for ListByMediaFile (subtitle_tracks.go:27)
    log         *slog.Logger
}
```

**Master generation:**
```go
func (h *StreamV2Handler) Master(w, r) {
    mediaID := parseID(r)
    fileID := parseFileID(r)  // optional, 0 = primary
    streamSessionID := streamSessionIDFromValues(r.URL.Query())
    if streamSessionID == "" {
        streamSessionID = newStreamSessionID()
    }
    startOffset := parseStartOffset(r.URL.Query().Get("start"))
    sessionStartSeg := int(math.Floor(startOffset / hls.SegLength))

    // Use existing services (verified in code):
    //   service.StreamService.GetPrimaryFile(ctx, mediaID, fileID) → *model.MediaFile
    //   service.AudioTrackService.ListByMediaFile(ctx, mediaFileID) → []model.AudioTrack
    file, err := h.streamSvc.GetPrimaryFile(r.Context(), mediaID, fileID)
    if err != nil { respondError(...); return }
    audioTracks, err := h.audioSvc.ListByMediaFile(r.Context(), file.ID)
    if err != nil { respondError(...); return }

    key := hls.SessionKey{
        StreamSessionID:   streamSessionID,
        MediaID:           mediaID,
        FileID:            file.ID,
        SubtitleStreamIdx: parseSubIdx(r),
        VideoCopy:         parseVCopy(r),
        MaxHeight:         parseMaxHeight(r),
    }
    sess, err := h.mgr.GetOrCreate(r.Context(), key, file, audioTracks)
    if err != nil { respondError(...); return }

    // Prime ffmpeg from the requested resume position (Jellyfin gap logic inside).
    if err := sess.PrimeFromSegment(r.Context(), sessionStartSeg); err != nil {
        respondError(...); return
    }

    // Master is deterministic — no ffmpeg call needed for the playlist body itself.
    // sess.Prefix() returns hls.BuildPrefix(sess.Key).
    master := hls.GenerateMasterPlaylist(audioTracks, sess.Prefix(), estimatedBandwidth)
    rewritten := rewriteMasterPlaylistURIs(master, r.URL.Query(), sessionStartSeg)
    serveText(w, rewritten, "application/vnd.apple.mpegurl")
}
```

`rewriteMasterPlaylistURIs` thêm query params (`api_key`, `token`, `ssid`) vào sub-playlist URI references **và** thêm `session_start={sessionStartSeg}` để sub-playlist handler biết MEDIA-SEQUENCE.

`serveMediaPlaylist` đọc `session_start` từ query, gọi `sess.GetExtinfSnapshot(kind)`, gọi `hls.RenderMediaPlaylist(RenderOpts{ExtinfMap: snap, SessionStartSeg: sessionStart, ...})`.

→ Snippet đã sửa nhất quán: `Serve` (catch-all), không còn `MediaPlaylist`/`Segment` ảo, `sess.Prefix()` là method, service method names khớp với codebase thật (`GetPrimaryFile` từ `StreamService`, `ListByMediaFile` từ `AudioTrackService`).

**Acceptance:**
- Flag off → tất cả test v1 vẫn pass, không có route v2
- Flag on → cả v1 và v2 endpoint đều hoạt động, test với curl + browser
- Master playlist hợp lệ với HLS validator (`hlsfetch` hoặc `mediastreamvalidator`)
- Sub-playlist (video.m3u8) chứa EXTINF đúng (đọc từ ffmpeg), URI segment có query auth

**Risk:** Low. Feature flag tách biệt hoàn toàn.

**Rollback:** `VELOX_HLS_V2=false` → routes không đăng ký → 404 → frontend tự fallback v1.

### Phase 5: Client integration (medium risk)

**Mục tiêu:** Backend trả `hls_v2_url`, frontend ưu tiên dùng nếu có.

**Files sửa (đã verify đường dẫn):**
- `backend/internal/handler/playback.go` ([playback.go:66](backend/internal/handler/playback.go#L66)) — `PlaybackInfoResponse` struct (đang có `StreamSessionID`) thêm field `HLSV2URL`:
  ```go
  type PlaybackInfoResponse struct {
      // ... existing fields
      StreamURL       string `json:"stream_url"`
      DirectURL       string `json:"direct_url,omitempty"`
      StreamSessionID string `json:"stream_session_id,omitempty"`
      HLSV2URL        string `json:"hls_v2_url,omitempty"`  // chỉ set nếu cfg.HLSV2Enabled
  }
  ```
  Handler set `HLSV2URL = "/api/stream/v2/{id}/hls/master.m3u8?ssid=...&api_key=..."` khi flag bật.

- `packages/shared/types/playback.ts` ([playback.ts:88](packages/shared/types/playback.ts#L88)) — thêm `hls_v2_url?: string` vào `PlaybackInfo` interface (KHÔNG phải `api.ts` như plan cũ ghi sai). Re-export từ `index.ts` không cần đụng vì `PlaybackInfo` đã được export sẵn ở [index.ts:8](packages/shared/types/index.ts#L8).

- `packages/shared/hooks/media/usePlayback.ts` ([usePlayback.ts:31](packages/shared/hooks/media/usePlayback.ts#L31)) — `useStreamUrls` `select` callback ưu tiên `hls_v2_url`:
  ```ts
  select: (info: PlaybackInfo): StreamUrls => {
    // v2 path: backend trả URL sẵn sàng dùng
    if (info.hls_v2_url) {
      return {
        direct: info.direct_url || info.stream_url,
        hls: info.hls_v2_url,
        primary_file_id: info.primary_file_id,
        stream_session_id: info.stream_session_id,
      }
    }
    // v1 fallback: existing logic build URL từ stream_url
    const isHLS = info.stream_url.includes('/hls/')
    let hlsUrl: string | undefined
    if (isHLS) {
      hlsUrl = info.stream_url
    } else {
      // ... existing logic ...
    }
    return { direct: ..., hls: hlsUrl, ... }
  }
  ```

**KHÔNG cần sửa:**
- `webapp/src/pages/WatchPage.tsx` — chỉ consume `streamUrls.hls`, không quan tâm v1/v2
- `mobile/` — chỉ consume tương tự
- `android-native/` — tương tự

→ Đây là điểm em sửa lại từ plan cũ. Chỉ cần sửa hook `useStreamUrls`, mọi consumer tự dùng được.

**Acceptance:**
- Bật flag → DevTools Network thấy request tới `/api/stream/v2/...`
- Tắt flag → DevTools Network thấy request tới `/api/stream/...` (v1)
- Switch flag không cần rebuild client (chỉ cần refresh page)

**Risk:** Low-medium. Hook change ảnh hưởng cả webapp + mobile + android-native cùng lúc.

**Rollback:** Backend không trả `hls_v2_url` → hook tự dùng v1 logic. Không cần đụng client.

### Phase 6: Testing & validation (CRITICAL)

**Mục tiêu:** Verify v2 hoạt động đúng cho mọi case trước Phase 7.

**Test matrix:**

| # | Case | Test media | Expected |
|---|---|---|---|
| 1 | Direct play (không HLS) | h264/aac mp4 | Vẫn dùng direct play, không qua v2 |
| 2 | HLS bitrate thấp | h264/ac3 1080p ~5 Mbps | Play mượt từ đầu tới cuối, không stall |
| 3 | HLS REMUX bitrate cao | Friends S01E14 ~16 Mbps | **Phải fix bug 3:18** |
| 4 | HLS UHD HEVC | 4K HDR phim test | Play mượt nếu HW transcode được |
| 5 | Resume từ giữa | Bất kỳ phim, position > 0 | Resume đúng vị trí, segment number tính đúng |
| 6 | Tua trong gap (≤ 4 segs) | Đang play seg 5, tua → seg 8 | Tua < 200ms |
| 7 | Tua xa gap (> 4 segs) | Đang play seg 5, tua → seg 100 | Tua < 1s |
| 8 | Tua liên tục 5 lần / 10s | | Không leak ffmpeg, không stuck |
| 9 | Multi-audio switch | Phim có 2+ audio | Đổi track instant, không restart ffmpeg |
| 10 | Subtitle burn-in PGS | Phim có PGS sub, bật burn-in | Sub hiển thị đúng, segment encode được |
| 11 | Network slow | Throttle 1 Mbps client | Buffer hợp lý, không stall vô hạn |
| 12 | Long video > 2 tiếng | | Idle GC hoạt động, ffmpeg không leak |
| 13 | Concurrent users | 2 user xem 2 phim khác nhau | Cả 2 đều OK, session không xung đột |
| 14 | Same user multi-tab | 2 tab cùng phim | Session reuse hoặc tách rõ ràng |
| 15 | Session GC | Idle 11 phút | Folder bị xóa, ffmpeg killed |

**Test environment:**
- NAS production thật (Velox Docker container)
- 2 browser: Chrome/Edge + Safari
- 1 mobile (Expo dev build)

**Acceptance:** Tất cả 15 case pass.

**Risk:** Critical phase — bắt bug ở đây tốt hơn production.

**Rollback:** Tắt flag, fix bug, retest.

### Phase 7: Migration & cleanup (low risk, AFTER verify)

**Mục tiêu:** Default-on v2, deprecate v1.

**Steps (4 milestones, không gộp):**

**M1: Default-on v2 (week 1)**
- `VELOX_HLS_V2` default `true` trong `config.go`
- v1 routes vẫn đăng ký song song
- Monitor 1 tuần — nếu user report bug nghiêm trọng → set env `VELOX_HLS_V2=false` rollback ngay

**M2: Deprecation warning (week 2-3)**
- Log warning khi có request tới `/api/stream/{id}/hls/...` (v1 path)
- Tracking metric: số request v1 vs v2
- Verify mọi client đã dùng v2 (webapp, mobile, android, third-party)

**M3: Disable v1 routes (week 4)**
- v1 routes không đăng ký nữa, return 404
- v1 code vẫn còn trong source, có thể re-enable bằng cách revert 1 commit
- Monitor 1 tuần thêm

**M4: Code deletion (week 5+)**
- **Chỉ làm khi M3 đã ổn 1 tuần**
- Xóa code v1:
  - `runHLSFFmpeg()`, `GenerateHLSWithAudio()`, `runMultiOutputHLS()` trong `transcoder_hls.go`
  - `serveProjectedHLSPlaylist()`, `projectPlaylistToVOD()` trong `stream.go`
  - `WaitForSegment()` polling logic phức tạp
- **Sau M4 không thể rollback bằng env nữa** — chỉ có thể rollback bằng git revert
- Tag release `v0.2.0-hls-v2` trước khi M4

**Rollback strategy theo milestone:**
| Milestone | Rollback method | Time to rollback |
|---|---|---|
| M1 | Set env `VELOX_HLS_V2=false` + restart container | 1 phút |
| M2 | Tương tự M1 | 1 phút |
| M3 | Set env + revert 1 commit (re-enable routes) | 5 phút |
| M4 | Git revert + rebuild + redeploy | 30 phút - 1 tiếng |

→ Đây là điểm em sửa lại từ plan cũ. M4 (delete v1) chỉ rollback được bằng git, không bằng env. Plan cũ nói "1 phút rollback" cho mọi giai đoạn là sai.

**Risk:** Low (đã verify ở Phase 6).

## Files chính sẽ tạo/sửa

### Mới
- `backend/internal/hls/naming.go` (~100 LOC)
- `backend/internal/hls/master.go` (~80 LOC)
- `backend/internal/hls/extinf.go` (~150 LOC) — parse + merge + render media playlist
- `backend/internal/hls/naming_test.go` (~120 LOC)
- `backend/internal/hls/master_test.go` (~100 LOC)
- `backend/internal/hls/extinf_test.go` (~180 LOC)
- `backend/internal/hls/testdata/` — fixture playlist từ ffmpeg run thật
- `backend/internal/transcoder/v2_encoder.go` (~200 LOC)
- `backend/internal/streamv2/session.go` (~350 LOC) — Session struct + RequestSegment + ingestPlaylistFile
- `backend/internal/streamv2/manager.go` (~150 LOC)
- `backend/internal/streamv2/session_test.go` (~250 LOC)
- `backend/internal/handler/streamv2.go` (~250 LOC) — Master + Serve dispatcher

### Sửa (đã verify từng đường dẫn)
- `backend/cmd/server/server_app.go` ([server_app.go:354](backend/cmd/server/server_app.go#L354)) — `initServices()` thêm `streamv2.NewManager(...)`, `initHandlers()` thêm `handler.NewStreamV2Handler(...)` (~15 LOC). KHÔNG sửa `main.go`.
- `backend/cmd/server/server_routes.go` ([server_routes.go:235](backend/cmd/server/server_routes.go#L235)) — đăng ký 2 routes v2 sau routes v1, gated bởi `app.cfg.HLSV2Enabled` (~10 LOC)
- `backend/internal/config/config.go` — thêm `HLSV2Enabled bool` field, đọc env `VELOX_HLS_V2` (~5 LOC)
- `backend/internal/handler/playback.go` ([playback.go:66](backend/internal/handler/playback.go#L66)) — `PlaybackInfoResponse` thêm `HLSV2URL` field, handler build URL khi flag bật (~15 LOC)
- `packages/shared/types/playback.ts` ([playback.ts:88](packages/shared/types/playback.ts#L88)) — `PlaybackInfo` thêm `hls_v2_url?: string` (~1 LOC). KHÔNG sửa `api.ts` (file không tồn tại cho type này) hoặc `index.ts` (đã re-export sẵn).
- `packages/shared/hooks/media/usePlayback.ts` ([usePlayback.ts:31](packages/shared/hooks/media/usePlayback.ts#L31)) — `useStreamUrls` ưu tiên `info.hls_v2_url` (~15 LOC)

### Xóa (M4, sau khi v2 ổn định 4+ tuần)
- ~300 LOC trong `transcoder_hls.go`
- ~100 LOC trong `stream.go` (project hack)
- Một phần `transcoder.go` (WaitForSegment phức tạp)

## Tổng kết

**Net code change:** +1500 LOC (Phase 1-6) → -400 LOC (M4) = +1100 LOC

**Effort:** 7 phase + 4 milestone migration. Phase 3 + Phase 6 là 2 phase tốn effort nhất.

**Em sẽ làm:**
1. Phase 1 ngay sau khi anh approve plan này (zero risk)
2. Mỗi phase commit riêng để anh review
3. Mỗi phase em sẽ pause sau commit để anh test trước khi qua phase tiếp

## Open questions đã trả lời

| Question | Answer |
|---|---|
| Source of truth: nhánh batch hay long-running? | **Long-running, theo Jellyfin chính xác.** Plan này không còn nhánh batch. |
| V2 có giữ `#EXT-X-MEDIA` cho multi-audio? | **Có.** Section "Multi-audio architecture" mô tả chi tiết. |
| Cache key tuple? | **`(StreamSessionID, MediaID, FileID, SubtitleStreamIdx, VideoCopy, MaxHeight)`.** PHẢI include `StreamSessionID` để cách ly per-viewer (giống v1). KHÔNG include startOffset (segment naming là deterministic), KHÔNG include audio track (1 session encode tất cả audio). |
| Source of truth cho media playlist EXTINF: ffmpeg hay backend? | **ffmpeg.** Backend đọc playlist file ffmpeg ghi ra, parse, merge cumulative qua mọi ffmpeg restart trong `Session.extinfMap`, render lại khi serve client. Backend CHỈ tự sinh `master.m3u8` (không có EXTINF, deterministic 100%). Lý do: HLS fmp4 segment duration KHÔNG luôn = `-hls_time` chính xác do GOP keyframe alignment — synthetic EXTINF sẽ tái tạo bug v1. |
