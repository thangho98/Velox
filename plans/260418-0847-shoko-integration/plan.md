# Plan: Shoko (AniDB) Integration
Created: 26-04-2026 08:47
Status: 🟡 In Progress

## Overview
Dự án tách biệt metadata matcher hiện tại của Velox thành hệ thống Provider có thể 'plug and play' thay chui (Chain of Responsibility).
Kế hoạch bao gồm việc ưu tiên Shoko (AniDB) dành cho Advanced anime watchers, sau đó fallback về AniList và TMDb. Hệ thống đảm bảo tính tuỳ chọn và không bắt buộc với người dùng phổ thông.

## Tech Stack
- Cốt lõi Backend: Golang
- CSDL: SQLite/GORM cho cấu hình/bảng (có thể)
- Frontend Settings: React & TypeScript

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Foundation (Cấu trúc lại Matcher, NFO & TMDb) | ⬜ Pending | 0% |
| 02 | AniList Provider (Cốt lõi hiển thị Anime) | ⬜ Pending | 0% |
| 03 | Database & Settings UI (Bật tắt tích hợp) | ⬜ Pending | 0% |
| 04 | Shoko Provider (Tích hợp API truy vấn AniDB) | ⬜ Pending | 0% |
| 05 | Testing & Validation | ⬜ Pending | 0% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Start Phase 2: `/code phase-02`
- Check progress: `/next`
- Lưu lại context nếu kết thúc phiên: `/save-brain`
