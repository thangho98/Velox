# Phase 03: Player UX
Status: ⬜ Pending
Plan: K - Intro / Credits Skip
Dependencies: Phase 02

## Mục tiêu
Thêm UX `Skip Intro` / `Skip Credits` vào web player mà không phá controls hiện có, không spam CTA, và hoạt động nhất quán cho cả direct play lẫn HLS.

## Output của phase này
- `WatchPage` biết segment nào đang active
- CTA hiện/ẩn đúng thời điểm
- Click CTA seek tới cuối segment
- State reset đúng khi đổi media hoặc file version

## Tasks
### 1. Consume Marker Data
- [ ] Update `webapp/src/types/api.ts` với `skip_segments` (field name khác table name `media_markers`)
- [ ] Ensure `usePlaybackInfo` / `useStreamUrls` vẫn share cùng payload cache
- [ ] Derive `activeSkipSegment` từ `playbackInfo.skip_segments` + `currentTime`
- [ ] Keep logic local trong `WatchPage` hoặc extract ra helper hook nếu component phình quá nhanh
- [ ] ⚠️ **WatchPage useEffect gotcha:** Video element bị conditional render (behind `if (mediaLoading || streamLoading)` early return). Effect cần `currentTime` listener trên video element → PHẢI include `streamUrls` trong deps để re-run khi element xuất hiện, tránh lỗi `videoRef.current === null`.

**Verify:** log/debug state cho media có 2 segments và active segment thay đổi đúng theo `currentTime`

### 2. Show Skip CTA
- [ ] In `webapp/src/pages/WatchPage.tsx`, detect segment active theo `currentTime`
- [ ] Show floating button `Skip Intro` hoặc `Skip Credits`
- [ ] Hide button ngoài segment window
- [ ] Place CTA gần nhóm controls hiện có, không che subtitle overlay quan trọng
- [ ] CTA label theo `segment.type`

**Verify:** trong khoảng `[start, end)`, CTA visible; ngoài khoảng này, CTA hidden

### 3. Handle Seek
- [ ] On click, seek video tới `segment.end`
- [ ] Track skipped segment trong session để không hiện lại ngay
- [ ] Reset session state khi đổi media/file
- [ ] Update `currentTime` local state sau seek để UI không lag 1 tick
- [ ] Nếu user seek lùi lại trước segment đã skip, quyết định có hiện lại hay không; khuyến nghị: không hiện lại trong cùng session trừ khi reload media

**Verify:** click CTA làm `video.currentTime` nhảy đúng `end`, CTA biến mất ngay

### 4. Handle Edge Cases
- [ ] Không show CTA khi user đang pause ở ngoài window
- [ ] Không spam CTA nếu seek qua lại gần boundary
- [ ] Ưu tiên 1 CTA tại một thời điểm nếu segment overlap
- [ ] Thêm threshold nhỏ để tránh flicker ở boundary do float precision, ví dụ ±0.25s
- [ ] Nếu `credits` và `up next` cùng xuất hiện, ưu tiên CTA nào cần quyết định sớm; khuyến nghị: `Skip Credits` không đè `Up Next`

**Verify:** boundary seek không làm CTA nhấp nháy; near-end episode vẫn giữ `Up Next` usable

### 5. Verify
- [ ] Add frontend test hoặc interaction coverage cho visibility + seek behavior
- [ ] Manual verify với direct play và HLS playback
- [ ] Manual verify với subtitle overlay đang bật để đảm bảo CTA không đè layout quá xấu

## Files dự kiến
- `webapp/src/pages/WatchPage.tsx`
- `webapp/src/types/api.ts`
- `webapp/src/hooks/stores/useMedia.ts`
- Có thể thêm helper ở `webapp/src/components/watch/watchHelpers.ts`

## UX notes
- Nên ship V1 với CTA đơn giản, không animation cầu kỳ
- Không cần auto-skip; user phải chủ động click
- Với credits, có thể chỉ hiện sau khi intro flow ổn và đã test với `Up Next`
