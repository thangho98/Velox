# Phase 02: Frontend UI
Status: ⬜ Pending
Dependencies: Phase 01 (Backend API)

## Objective
Cập nhật homepage với 2 row mới: Continue Watching (thay thế row cũ) và Next Up.

## Implementation Steps

### 1. API client + hooks
File: `webapp/src/hooks/stores/useMedia.ts`

- [ ] Thêm API functions: `getContinueWatching(limit)`, `getNextUp(limit)`, `dismissProgress(mediaId)`
- [ ] Thêm hooks: `useContinueWatching({limit})`, `useNextUp({limit})`, `useDismissProgress()`
- [ ] Cache invalidation khi `updateProgress` thành công: invalidate `continue-watching` + `next-up` + `recently-watched`
- [ ] Cache invalidation khi `dismissProgress` thành công: invalidate `continue-watching`
- [ ] Response format: `respondJSON` đã wrap `{"data": ...}`, API client extract `.data` như các hook hiện tại

### 2. ContinueWatchingCard component
File: `webapp/src/components/ContinueWatchingCard.tsx`

- [ ] Card hiển thị:
  - Backdrop/poster thumbnail (dùng backdrop nếu có, fallback poster)
  - Progress bar (position / duration)
  - Tiêu đề: movie title, hoặc "S04E23 · Friends" cho episode
  - Thời gian còn lại: `Math.ceil((duration - position) / 60)` + "phút còn lại"
  - Click card → navigate tới WatchPage với resume position
  - Nút X (dismiss) → gọi `dismissProgress(mediaId)` — KHÔNG gọi `deleteProgress` (giữ favorite/rating)

### 3. NextUpCard component
File: `webapp/src/components/NextUpCard.tsx`

- [ ] Card hiển thị:
  - Still image (episode thumbnail `still_path`) hoặc `series_poster` fallback
  - "S04E24 · Friends" format
  - Episode title
  - Click card → navigate tới WatchPage (bắt đầu từ đầu, position = 0)
  - Không có progress bar (Next Up = chưa bắt đầu xem)
  - Không có nút X (không cần dismiss)

### 4. Update HomePage
File: `webapp/src/pages/HomePage.tsx`

- [ ] Thay row "Continue Watching" cũ (dùng `useRecentlyWatched`) bằng `useContinueWatching`
- [ ] Thêm row "Next Up" mới bên dưới (dùng `useNextUp`)
- [ ] Thứ tự rows: Continue Watching → Next Up → Movies → Series
- [ ] Ẩn row nếu data rỗng (giữ logic `hasItems` hiện tại)

### 5. Polish
- [ ] Responsive: horizontal scroll trên mobile, grid trên desktop (MediaRow đã có sẵn)
- [ ] Loading skeleton cho cả 2 rows (MediaRow đã có sẵn)
- [ ] Empty state: ẩn row hoàn toàn nếu không có data

## UI Layout

```
┌─────────────────────────────────────────────────┐
│  Welcome back, Thawng                           │
├─────────────────────────────────────────────────┤
│ ▶ Continue Watching                    See All → │
│ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐            │
│ │backdr│ │backdr│ │poster│ │backdr│  ← scroll   │
│ │██░░░░│ │████░░│ │█░░░░░│ │███░░░│  progress   │
│ │title │ │S4E23 │ │movie │ │S2E01 │             │
│ │32min │ │Friends│ │1h02m│ │House │             │
│ │   [X]│ │   [X]│ │   [X]│ │   [X]│  dismiss   │
│ └──────┘ └──────┘ └──────┘ └──────┘             │
├─────────────────────────────────────────────────┤
│ ⏭ Next Up                                       │
│ ┌──────┐ ┌──────┐ ┌──────┐                      │
│ │still │ │still │ │still │                       │
│ │S4E24 │ │S2E02 │ │S1E01 │  ← chỉ episodes     │
│ │Friends│ │House │ │Office│                      │
│ └──────┘ └──────┘ └──────┘                      │
├─────────────────────────────────────────────────┤
│ 🎬 Movies                             See All → │
│ ...                                              │
└─────────────────────────────────────────────────┘
```

## Test Criteria
- [ ] Continue Watching row hiện khi có ít nhất 1 item in-progress
- [ ] Continue Watching card click → WatchPage resume đúng position
- [ ] Dismiss (nút X) → item biến mất, favorite/rating vẫn còn
- [ ] Next Up row hiện khi có ít nhất 1 series có tập tiếp theo
- [ ] Next Up card click → WatchPage bắt đầu từ đầu
- [ ] Next Up không hiện episode đang nằm ở Continue Watching
- [ ] Xem xong 1 tập → biến mất khỏi Continue Watching, Next Up cập nhật tập mới

## Notes
- Giữ nguyên endpoint `recently-watched` + page `/recently-watched` (không breaking change)
- ContinueWatchingCard tách riêng khỏi MediaCard vì layout khác (progress bar, time remaining, dismiss button)
- NextUpCard đơn giản hơn (không progress, không dismiss)

---
Previous: Phase 01 — Backend API
