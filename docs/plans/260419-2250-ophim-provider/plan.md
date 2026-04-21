# Plan: OPhim Third-Party Media Provider
Created: 2026-04-19 22:50
Status: 🟡 In Progress

## Overview
Tích hợp trực tiếp public API của `ophim1.com` vào Backend của Velox. Mục tiêu là cho phép hệ thống đồng bộ danh sách phim (kèm metadata) từ dịch vụ này vào thư viện Velox. Khi người dùng bấm Play, backend sẽ trực tiếp trả về link luồng HLS (`.m3u8`) từ CDN của OPhim cho client player, thay vì sử dụng local transcoding hay FShare.

## Tech Stack
- Backend: Golang (REST API client, Database Job)
- Database: SQLite (Lưu Origin Source `ophim` và Media Items)

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | OPhim Package API Client | ✅ Complete | 100% |
| 02 | Metadata Sync Job | ✅ Complete | 100% |
| 03 | Playback Stream Routing | ✅ Complete | 100% |
| 04 | Testing & Integration | ⬜ Pending | 0% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
- Save context: `/save-brain`
