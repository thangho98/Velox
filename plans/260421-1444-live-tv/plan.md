# Plan: Live TV Integration
Created: 260421-1444
Status: 🟡 In Progress

## Overview
Tích hợp trải nghiệm xem Live TV (Truyền hình Trực tiếp) vào Velox.
Cho phép gộp chung danh sách kênh nội địa (qua file m3u nội bộ `velox_livetv.m3u`) và hàng ngàn kênh quốc tế mở (thông qua link URL cập nhật từ `iptv-org`), giúp mang đến trải nghiệm Truyền hình phong phú và miễn phí.

## Tech Stack
- **Backend:** Go, SQLite (Bảng `live_channels`, `live_playlists`).
- **Core Engine:** Định kỳ tải M3U, Regex parse chuẩn #EXTM3U, #EXTINF (tách ID, Logo, Group, Quốc gia).
- **Frontend:** React 19, TailwindCSS 4 (Component danh mục dọc đa tầng).
- **Mobile/TV:** Jetpack Compose, Media3/ExoPlayer (Luồng HLS trực tiếp).

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Database & M3U Parser | ✅ Complete | 100% |
| 02 | Backend API & Sync Task | ⬜ Pending | 0% |
| 03 | Frontend Web UI | ⬜ Pending | 0% |
| 04 | Android Mobile & TV UI | ⬜ Pending | 0% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
- Evaluate Design: `/design`
