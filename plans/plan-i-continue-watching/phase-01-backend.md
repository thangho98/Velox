# Phase 01: Backend API
Status: ⬜ Pending
Dependencies: None (existing infrastructure)

## Objective
Thêm 3 thứ: repo method `DismissProgress` (reset progress mà không xóa row), 2 endpoint mới cho Continue Watching và Next Up.

## Semantic Decisions (chốt trước khi code)

### Continue Watching vs Next Up — không trùng nhau
- **Continue Watching:** media (movie hoặc episode) đang xem dở → `position > 0 AND completed = 0`
- **Next Up:** tập TIẾP THEO chưa bắt đầu xem sau tập đã hoàn thành → `completed = 1` rồi tìm tập kế
- **Không overlap:** Next Up chỉ chọn episode chưa có trong Continue Watching (position = 0 hoặc chưa có user_data)
- **Edge case:** Nếu user xem dở S01E05, nó nằm ở Continue Watching. Next Up sẽ không hiện S01E05 mà hiện S01E06 (nếu E04 đã completed), hoặc không hiện series này (nếu chưa completed tập nào)

### Dismiss vs Delete
- **DismissProgress:** chỉ reset `position = 0, completed = 0, last_played_at = NULL` — giữ nguyên `is_favorite`, `rating`, `play_count`
- **DeleteProgress (existing):** xóa cả row — chỉ dùng khi user muốn xóa hoàn toàn tất cả data cho media đó

## Implementation Steps

### 1. Repository: `DismissProgress`
File: `backend/internal/repository/user_data.go`

```sql
-- Reset progress mà giữ favorite/rating/play_count
UPDATE user_data
SET position = 0, completed = 0, last_played_at = NULL, updated_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND media_id = ?
```

- [ ] Thêm method `DismissProgress(ctx, userID, mediaID) → error`

### 2. Repository: `ListContinueWatching`
File: `backend/internal/repository/user_data.go`

```sql
-- Lấy media đang xem dở (có progress, chưa hoàn thành)
SELECT ud.user_id, ud.media_id, ud.position, ud.completed, ud.play_count,
       ud.last_played_at, ud.updated_at,
       m.title, m.poster_path, m.backdrop_path, m.media_type,
       COALESCE(mf.duration, 0) as media_duration,
       e.episode_number, e.series_id,
       s.title as series_title,
       se.season_number
FROM user_data ud
JOIN media m ON ud.media_id = m.id
LEFT JOIN media_files mf ON m.id = mf.media_id AND mf.is_primary = 1
LEFT JOIN episodes e ON e.media_id = m.id
LEFT JOIN series s ON e.series_id = s.id
LEFT JOIN seasons se ON e.season_id = se.id
WHERE ud.user_id = ?
  AND ud.position > 0
  AND ud.completed = 0
ORDER BY ud.last_played_at DESC
LIMIT ?
```

- [ ] Thêm method `ListContinueWatching(ctx, userID, limit) → []*ContinueWatchingItem`
- [ ] Model `ContinueWatchingItem` trong `model/user.go`:
  - UserData fields (media_id, position, last_played_at, ...)
  - Media fields (title, poster_path, backdrop_path, media_type, duration)
  - Episode context (series_title, season_number, episode_number) — nullable

### 3. Repository: `ListNextUp`
File: `backend/internal/repository/user_data.go`

Logic:
1. Tìm series mà user đã completed ít nhất 1 tập
2. Cho mỗi series, tìm tập đầu tiên chưa completed VÀ chưa bắt đầu xem (loại trừ Continue Watching)
3. Dùng subquery với ROW_NUMBER để chọn đúng 1 episode per series

```sql
-- Next Up: tập tiếp theo chưa xem cho mỗi series đang theo dõi
WITH user_series AS (
  -- Series mà user đã completed ít nhất 1 tập
  SELECT e.series_id,
         MAX(ud.last_played_at) as last_watched
  FROM user_data ud
  JOIN episodes e ON e.media_id = ud.media_id
  WHERE ud.user_id = ? AND ud.completed = 1
  GROUP BY e.series_id
),
candidates AS (
  -- Tất cả tập chưa completed VÀ chưa bắt đầu xem dở, sắp xếp theo thứ tự
  SELECT us.series_id, us.last_watched,
         e.media_id,
         se.season_number, e.episode_number,
         ROW_NUMBER() OVER (
           PARTITION BY us.series_id
           ORDER BY se.season_number, e.episode_number
         ) as rn
  FROM user_series us
  JOIN episodes e ON e.series_id = us.series_id
  JOIN seasons se ON e.season_id = se.id
  LEFT JOIN user_data ud ON ud.media_id = e.media_id AND ud.user_id = ?
  WHERE COALESCE(ud.completed, 0) = 0
    AND COALESCE(ud.position, 0) = 0  -- Loại trừ đang xem dở (đã ở Continue Watching)
)
SELECT m.id as media_id, m.title, m.media_type, m.backdrop_path,
       e.episode_number, e.title as episode_title, e.still_path,
       c.season_number, s.title as series_title, s.poster_path as series_poster,
       c.last_watched,
       COALESCE(mf.duration, 0) as media_duration
FROM candidates c
JOIN episodes e ON e.media_id = c.media_id
JOIN media m ON m.id = c.media_id
JOIN series s ON e.series_id = s.id
LEFT JOIN media_files mf ON m.id = mf.media_id AND mf.is_primary = 1
WHERE c.rn = 1  -- Chỉ lấy tập đầu tiên chưa xem per series
ORDER BY c.last_watched DESC
LIMIT ?
```

- [ ] Thêm method `ListNextUp(ctx, userID, limit) → []*NextUpItem`
- [ ] Model `NextUpItem` trong `model/user.go`:
  - media_id, title, media_type, backdrop_path, duration
  - episode_title, still_path, season_number, episode_number
  - series_title, series_poster
  - last_watched_at (series-level, để sort)

### 4. Service methods
File: `backend/internal/service/user_data.go`

- [ ] `DismissProgress(ctx, userID, mediaID) → error` — gọi repo DismissProgress
- [ ] `ContinueWatching(ctx, userID, limit) → []*ContinueWatchingItem` — default limit=20, max=50
- [ ] `NextUp(ctx, userID, limit) → []*NextUpItem` — default limit=20, max=50

### 5. Handler + Routes
File: `backend/internal/handler/profile.go` + `backend/cmd/server/main.go`

- [ ] `GET /api/profile/continue-watching?limit=20` — handler `ContinueWatching`
- [ ] `GET /api/profile/next-up?limit=20` — handler `NextUp`
- [ ] `DELETE /api/profile/progress/{mediaId}/dismiss` — handler `DismissProgress` (reset, không xóa row)

## Response Format

`respondJSON` đã wrap trong `{"data": ...}` — handler chỉ cần trả slice, consistent với `ListRecentlyWatched` hiện tại.

```json
// GET /api/profile/continue-watching → respondJSON(w, 200, items)
// Wire format: {"data": [...]}
[
  {
    "media_id": 42,
    "title": "The One Before the Laundry",
    "poster_path": "/posters/friends.jpg",
    "backdrop_path": "/backdrops/friends.jpg",
    "media_type": "episode",
    "position": 1234.5,
    "duration": 1800,
    "last_played_at": "2026-03-13T15:30:00Z",
    "series_title": "Friends",
    "season_number": 4,
    "episode_number": 23
  }
]

// GET /api/profile/next-up → respondJSON(w, 200, items)
// Wire format: {"data": [...]}
[
  {
    "media_id": 43,
    "title": "The One with Ross's Wedding (2)",
    "episode_title": "The One with Ross's Wedding (2)",
    "still_path": "/stills/friends-s04e24.jpg",
    "media_type": "episode",
    "duration": 1800,
    "series_title": "Friends",
    "series_poster": "/posters/friends.jpg",
    "season_number": 4,
    "episode_number": 24,
    "last_watched_at": "2026-03-13T15:30:00Z"
  }
]
```

## Test Criteria
- [ ] Continue Watching: chỉ trả về items có position > 0 AND completed = 0
- [ ] Continue Watching: không trả về completed items
- [ ] Continue Watching: sort by last_played_at DESC
- [ ] Continue Watching: gồm cả movies và episodes
- [ ] Next Up: chỉ trả về episodes (không có movies)
- [ ] Next Up: chỉ hiện khi user đã completed ít nhất 1 tập của series
- [ ] Next Up: skip series đã xem hết (tất cả episodes completed)
- [ ] Next Up: không trùng với Continue Watching (loại episode đang xem dở)
- [ ] Next Up: nếu xem xong S01E05, next = S01E06
- [ ] Next Up: nếu xem xong season cuối tập, next = season tiếp theo E01
- [ ] DismissProgress: reset position/completed/last_played_at, giữ nguyên is_favorite/rating/play_count
- [ ] DismissProgress: item biến mất khỏi Continue Watching sau khi dismiss

## Notes
- Không sửa endpoint `recently-watched` hiện tại (backward compatible)
- `progress_percent` tính ở frontend: `(position / duration) * 100`
- Handler trả slice trực tiếp vào `respondJSON` — format cuối cùng là `{"data": [...]}`

---
Next: Phase 02 — Frontend UI
