# Phase 01: Tiered Gap-Based Splitting
Status: ⬜ Pending
Dependencies: None

## Objective
Thay vòng lặp chia batch cứng `for i := 0; i < len(cues); i += batchSize` trong `TranslateSRT` bằng greedy chunker tận dụng khoảng lặng (gap) giữa các cue. Không đổi `Translator` interface, không đụng prompt format của 3 provider. Phase này đóng gói hoàn toàn trong layer `backend/pkg/translate`.

## Rationale
- Dữ liệu thực: trên Avatar (movie) thuật toán này cho 0 forced cuts / 106 batches, trên Malcolm (sitcom) cho 14 natural / 16 total.
- Không đổi interface = zero risk với các call site hiện tại (`service/subtitle_auto_translate.go`, `service/subtitle.go`, `cmd/aisubtest`).
- Tiered thresholds tự thích nghi: content gap-rich cắt sớm (batch nhỏ, nhiều ranh giới), content dense fill đến gần max rồi mới nhận gap nhỏ hơn.

## Requirements

### Functional
- [ ] Parse timing (`"00:01:23,456 --> 00:01:25,789"`) trong `SRTCue` thành float seconds (start, end).
- [ ] Greedy chunker với tiered thresholds:
    - `gap ≥ 3.0s AND cur_size ≥ 20` → cut (scene break)
    - `gap ≥ 2.0s AND cur_size ≥ 30` → cut (dialog pause rõ)
    - `gap ≥ 1.5s AND cur_size ≥ 40` → cut (gần full, pause nhẹ còn hơn forced)
    - `cur_size ≥ MaxBatchSize` → forced cut (fallback)
- [ ] Áp dụng cùng logic cho retry path (`translateBatchWithRetry` khi downsize batch).
- [ ] Khi cue thiếu timing hợp lệ (parse fail) → fallback về fixed-size batching như cũ (graceful degrade).

### Non-Functional
- [ ] Không đổi signature `Translator.Translate(ctx, texts, targetLang) ([]string, error)`.
- [ ] Không đổi signature `TranslateSRT(ctx, translator, srtContent, targetLang) (string, error)`.
- [ ] Performance: parse timing một lần, không re-parse trong chunker loop.
- [ ] Thread-safe: chunker là pure function, không share state.

## Implementation Steps

1. [ ] **Extend `SRTCue` với timing fields**
    - Thêm `StartSec, EndSec float64` vào struct `SRTCue` trong `translate.go`.
    - Update `ParseSRT` để parse timing string (format `HH:MM:SS,mmm`) → float seconds. Giữ `Timing` string field để `BuildSRT` vẫn hoạt động (no-op change output).
    - Nếu parse fail (malformed timing) → set `StartSec=-1, EndSec=-1` (sentinel cho graceful degrade).

2. [ ] **Viết chunker function — emit metadata về boundary type**
    - File: `translate.go` (hoặc tách `chunker.go` nếu dễ đọc hơn). Name đề xuất: `chunkCuesByGap`.
    - Return struct giàu thông tin thay vì chỉ index pair, để Phase 2 consume được:
      ```go
      type Batch struct {
          Start, End    int     // End exclusive
          LeftForced    bool    // true nếu boundary trái do hit maxBatch (không phải gap tự nhiên)
          RightForced   bool    // true nếu boundary phải do hit maxBatch
          LeftGap       float64 // gap tại boundary trái (seconds, 0 nếu batch đầu hoặc timing không hợp lệ)
          RightGap      float64 // gap tại boundary phải (0 nếu batch cuối hoặc timing không hợp lệ)
      }

      type ChunkResult struct {
          Batches     []Batch
          TimingValid bool  // false nếu bất kỳ cue nào có StartSec<0 (fallback fixed-size)
      }

      func chunkCuesByGap(cues []SRTCue, maxBatch int) ChunkResult
      ```
    - **Phân biệt 2 khái niệm tách rời:**
      - `LeftForced/RightForced` = "boundary này là do hit maxBatch, không phải gap tự nhiên" — dùng khi chunker biết gap giữa 2 cue.
      - `TimingValid=false` = "chunker không tin được vào timing, đã fallback" — là trạng thái toàn cục, Phase 2 dùng để tắt overlap entirely.
    - Logic tiered thresholds như mô tả trên khi `TimingValid=true`.
    - Fallback khi có cue `StartSec<0`:
      - Return `ChunkResult{Batches: fixedSizeChunks(cues, maxBatch), TimingValid: false}`.
      - Trong fixed-size batches: set `LeftForced=false, RightForced=false` cho tất cả boundaries (kể cả interior) — vì không biết chúng là forced hay natural. Flag `TimingValid=false` là tín hiệu chính để Phase 2 consume.
    - **Invariant:** khi `TimingValid==false`, Phase 2 PHẢI ignore `LeftForced/RightForced` và không apply overlap. Điều này giữ behavior legacy khi timing hỏng.

3. [ ] **Thay vòng lặp trong `TranslateSRT`**
    - Thay block `for i := 0; i < len(cues); i += batchSize` (dòng ~130-140) bằng call `result := chunkCuesByGap(cues, batchSize)`.
    - Giữ nguyên logic concurrency + error aggregation, iterate `result.Batches`.
    - Lưu `result.TimingValid` để pass xuống Phase 2 consumer (hiện tại Phase 1 chưa dùng, nhưng Phase 2 sẽ cần). Có thể lưu trong closure variable hoặc field của batch work struct.

4. [ ] **Retry path — xử lý đúng case non-contiguous**
    - `translateBatchWithRetry` có 2 retry path, cần xử lý khác nhau:

    **(a) Downsize-retry (`nextBatchSize < len(texts)`):** cues vẫn contiguous → OK dùng chunker gap-aware. Pass `[]SRTCue` subset từ caller qua signature mới: `translateBatchWithRetry(ctx, translator, texts []string, cues []SRTCue, targetLang) ([]string, error)`. Caller `TranslateSRT` truyền `cues[b.Start:b.End]`.

    **(b) Partial-retry (missing indexes):** cues KHÔNG contiguous — `missingTexts` chỉ gồm cue thiếu (VD indexes `[2, 7, 13]` trong batch gốc). Chunker gap-aware sẽ thấy gap giả giữa cue 2 và cue 7 (vốn không liền nhau trong SRT). **Quyết định: disable gap-aware cho partial retry** — dùng fixed-size (nextBatchSize cũ) cho path này. Lý do:
        - Partial retry thường chỉ vài cue lẻ → cắt batch thêm gần như không giúp gì.
        - Gap giả có thể gây oversplit (mỗi cue 1 batch → tốn API call vô ích).
        - Implementation: pass `cues=nil` khi gọi retry cho `missingTexts` → fallback fixed-size path trong helper.

    - Helper đề xuất: `splitForRetry(texts []string, cues []SRTCue, size int) [][2]int`. Nếu `cues==nil` hoặc `len(cues)!=len(texts)` → fixed-size. Else → gap-aware qua `chunkCuesByGap`.

5. [ ] **Unit tests — dùng metric assertable, không manual inspect**
    - File: `translate_test.go` (hoặc `chunker_test.go`).
    - Helper test: `countForcedBoundaries(batches []Batch) int` — đếm số boundary có `RightForced==true` (không tính batch cuối).
    - Test cases:
        - Empty cues → 0 batches.
        - Single cue → 1 batch, `LeftForced=false, RightForced=false`.
        - **Movie-like (synthetic):** 100 cues với gap xen kẽ [0.2s × 20, 3.5s, 0.2s × 20, 3.5s, ...], maxBatch=50 → assert `forcedBoundaries == 0`, số batches trong [2, 4].
        - **Sitcom-like (synthetic):** 60 cues gap 0.2s đều, maxBatch=50 → assert `forcedBoundaries >= 1`, `len(batches) == 2`.
        - **No gap (dense):** 55 cues all gap=0.1s, maxBatch=50 → 2 batches, boundary giữa `RightForced=true`.
        - **Malformed timing:** 1 cue có `StartSec=-1` → `TimingValid=false`, fallback fixed-size, tất cả interior boundaries `LeftForced=false` và `RightForced=false` (không dùng forced flag khi không tin timing).
        - **Tier ordering sanity:** 50 cues + 1 gap 3.5s ở vị trí 25 → 2 batches split tại đó (natural). Move gap về vị trí 10 (size < min_size=20) → chunker KHÔNG cắt, đợi cho tới khi size ≥ 20.
    - Assert: mọi batch có `size ≤ maxBatch`, thứ tự liên tục (`batches[i].End == batches[i+1].Start`), cover toàn bộ input (`batches[0].Start==0`, `batches[-1].End==len(cues)`).

6. [ ] **Integration smoke test — có numeric acceptance**
    - Commit 2 testdata files: `testdata/sitcom_dense.srt` (Malcolm pilot snippet, ~100 cues đầu) + `testdata/movie_sparse.srt` (Avatar snippet, ~200 cues đầu). Đã verify từ NAS, KHÔNG chứa text nhạy cảm — chỉ dialog public TV/cinema.
    - Test `TestChunker_RealSamples`:
        - Sitcom: assert `forcedBoundaries / totalBoundaries < 0.5` (allow tệ nhưng không phải 100%).
        - Movie: assert `forcedBoundaries == 0`.
    - Tắt test nếu testdata missing (skip), không fail build cho repo clone không có file.

## Files to Create/Modify

- `backend/pkg/translate/translate.go` — extend SRTCue, ParseSRT, rewrite batching loop trong TranslateSRT, đổi signature translateBatchWithRetry.
- `backend/pkg/translate/translate_test.go` — thêm chunker tests + integration tests.
- `backend/pkg/translate/testdata/sitcom_dense.srt` *(optional)* — Malcolm pilot snippet (~100 cues đầu), đại diện content dense dialog.
- `backend/pkg/translate/testdata/movie_sparse.srt` *(optional)* — Avatar snippet (~200 cues đầu), đại diện content sparse dialog với gap dồi dào.

## Test Criteria
- [ ] `cd backend && go test ./pkg/translate/...` pass (go.mod nằm ở `backend/`, không chạy từ repo root được).
- [ ] Existing tests (`ai_test.go`, `translate_test.go` cũ) không regression.
- [ ] `TestChunker_RealSamples`: movie sample `forcedBoundaries==0`, sitcom sample `forcedBoundaries/totalBoundaries < 0.5` — assert tự động, KHÔNG manual inspect.
- [ ] Malformed timing SRT vẫn translate được (fallback path). Assert `result.TimingValid==false` và mọi `LeftForced==RightForced==false` trong output chunker.
- [ ] Partial-retry path với non-contiguous missing indexes KHÔNG dùng gap-aware (verified bằng test unit gọi `splitForRetry(texts, nil, size)` → fixed-size output).

## Notes
- Các hằng số threshold (`3.0s`, `2.0s`, `1.5s`, min_size `20/30/40`) đặt thành `const` trong file `translate.go`, có comment chỉ rõ lấy từ đâu (link tới plan.md evidence table).
- **Không** expose thresholds ra config Settings — tuning plan này dựa trên dữ liệu; cho phép user tune sẽ làm plan Phase 2 (overlap) phức tạp hơn khi test.
- Prompt của provider có dòng "Count the input items first..." — giữ nguyên, không động trong phase này.

---
Next Phase: [phase-02-context-overlap.md](phase-02-context-overlap.md)
