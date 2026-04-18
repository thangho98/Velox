# Plan: Anime Metadata Integration (AniList)
Created: 2026-04-17
Status: 🟡 In Progress

## Overview
Tích hợp AniList API bằng GraphQL để gỡ bỏ sự phụ thuộc vào TMDb đối với kho Anime. Plan đã được điều chỉnh bám sát kiến trúc Velox hiện tại (tầng scanner, model chuẩn và settings provider).

## Tech Stack
- **Backend:** Go (GraphQL Client, Scanner Dispatcher, Migration)
- **Database:** SQLite (Mở rộng schema bảng `media`, `series`, library type)
- **Frontend:** React / TypeScript (Giao diện cấu hình API tại Provider settings. Thêm Type `anime` vào Create Library)

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Database & Models | ⬜ Pending | 0% |
| 02 | AniList GraphQL Client | ⬜ Pending | 0% |
| 03 | Scanner & Routing Logic | ⬜ Pending | 0% |
| 04 | Settings & FrontEnd | ⬜ Pending | 0% |
| 05 | Testing & Fallback | ⬜ Pending | 0% |

## Quick Commands
- Chạy Phase 1: `/code docs/plans/260417-2222-anime-metadata-anilist/phase-01-database.md`
- Xem tiến độ: `/next`
- Update design: `/design`
