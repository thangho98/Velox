# Velox - Home Media Server (Jellyfin-lite)
Created: 2026-03-10
Status: 🟡 In Progress

## Scope Definition
**Target:** Single-server self-hosted media server cho cá nhân/gia đình.
**Goal:** Ship bản usable đầu tiên (browse + play) càng sớm càng tốt, rồi iterate.
**NOT in scope (yet):** Multi-node, mobile native app, DLNA, plugin marketplace.

## Tech Stack
- Backend: Go 1.26 (stdlib net/http) + SQLite (WAL mode)
- Frontend: React + Vite + TailwindCSS 4
- Transcoding: FFmpeg 8.0
- Metadata: TMDb API

## Current State
Backend Phase 1 done: Library CRUD, Media scan (basic), Direct Play, HLS transcode (basic), Progress tracking (global, no user).
Frontend: Vite + Tailwind scaffolded, no UI yet.

## Plans

| Plan | Name | Focus | Status |
|------|------|-------|--------|
| A | Core Domain & Ingestion | Data model, scan pipeline, file identity, migrations | ⬜ |
| B | Auth & Sessions | First-run setup, JWT, per-user state, library ACL | ⬜ |
| C | Web App MVP | Login → Browse → Detail → Player → Resume | ⬜ |
| D | Playback Decision Engine | Direct play/stream/remux/transcode matrix | ⬜ |
| E | Streaming Enhancement | HW accel, ABR, trickplay, session management | ⬜ |
| F | Admin & Operations | Dashboard, scheduler, health, webhooks | ⬜ |
| G | Nice-to-have | SyncPlay, chapters, plugin system | ⬜ |

## Execution Order & Milestones

```
Plan A ──→ Plan B ──→ Plan C ──→ 🎯 MILESTONE: Usable MVP
                                      │
                                 Plan D ──→ Plan E ──→ 🎯 MILESTONE: Full Streaming
                                                           │
                                                      Plan F ──→ Plan G
```

### 🎯 Milestone 1: Usable MVP (Plans A + B + C)
- User có thể: đăng nhập → browse library → xem chi tiết phim → play video → resume
- Metadata: poster, overview, cast, genres từ TMDb
- TV Series: season/episode navigation
- Direct play hoạt động, basic transcode cho incompatible codecs

### 🎯 Milestone 2: Full Streaming (Plans D + E)
- Smart playback decisions (direct play vs transcode)
- Hardware accelerated transcoding
- Adaptive bitrate
- Trickplay thumbnails

### 🎯 Milestone 3: Production Ready (Plans F + G)
- Admin dashboard, scheduled tasks, health monitoring
- SyncPlay, chapters, extensibility

## Architecture Principles
1. **Scan pipeline là xương sống** - mọi thứ bắt đầu từ file → identity → metadata → playable
2. **Migration versioning** - không dùng CREATE IF NOT EXISTS cho evolution
3. **Playback decision trước optimization** - hiểu khi nào cần transcode trước khi tối ưu transcode
4. **Validate sớm bằng UI** - ship web client MVP ngay sau auth để kiểm chứng mọi thứ
