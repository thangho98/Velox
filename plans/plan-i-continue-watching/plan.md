# Plan I: Continue Watching / Next Up
Created: 2026-03-13
Status: ⬜ Pending
Priority: 🔴 High
Dependencies: Plans A-F (all done)

## Overview
Tách "Recently Watched" hiện tại thành 2 row có ý nghĩa hơn trên homepage:
- **Continue Watching:** Phim/tập đang xem dở (có progress bar, nút resume)
- **Next Up:** Tập tiếp theo của series đang theo dõi (auto-detect)

## Hiện trạng
- `user_data` table đã có: `position`, `completed`, `last_played_at`, `play_count`
- `ListRecentlyWatched` repo/service/handler đã có nhưng trả về TẤT CẢ (cả completed)
- Homepage đã có row "Continue Watching" nhưng dùng `recently-watched` endpoint (chưa filter đúng)
- Episodes chain: `series → seasons → episodes → media → user_data`

## Cần làm
1. **Continue Watching:** Query `user_data` WHERE `position > 0 AND completed = 0` — chỉ lấy đang xem dở
2. **Next Up:** Với mỗi series user đang xem, tìm tập tiếp theo chưa xem (episode_number + 1)
3. **Frontend:** 2 row riêng biệt trên homepage, card hiển thị thông tin phù hợp

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | Backend API | 6 tasks | ⬜ |
| 02 | Frontend UI | 5 tasks | ⬜ |

## Không cần migration mới
- `user_data` đã có đủ columns: `position`, `completed`, `last_played_at`
- Index `idx_ud_recent` đã cover `(user_id, last_played_at DESC) WHERE last_played_at IS NOT NULL`

## Quick Commands
- Start: `/code phase-01`
- Check: `/next`
