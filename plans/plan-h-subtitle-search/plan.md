# Plan H: External Subtitle Search
Status: ⬜ Pending
Priority: 🟡 Medium
Dependencies: Plans A-D (core + auth + playback)

## Overview
Cho phép user tìm và tải subtitle từ OpenSubtitles.com (primary) và Podnapisi (fallback) ngay trong player, không cần rời khỏi app.

## Providers
| Provider | Auth | API | Coverage |
|----------|------|-----|----------|
| OpenSubtitles.com | API key (app) + username/password (user) | REST v3 | ✅ Lớn nhất |
| Podnapisi | Không cần | JSON API | ✅ Tốt |

## Flow
1. User bật subtitle picker → click "Search for Subtitles"
2. Modal hiện lên, auto-search theo ngôn ngữ hiện tại
3. User chọn subtitle → backend tải về, lưu disk, tạo DB record
4. SubtitlePicker tự refresh list → user chọn luôn

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | App Settings (Backend) | 5 tasks | ⬜ |
| 02 | Provider Clients (Backend) | 7 tasks | ⬜ |
| 03 | Search + Download API (Backend) | 6 tasks | ⬜ |
| 04 | SubtitleSearchModal (Frontend) | 7 tasks | ⬜ |

## Files Created/Modified

### Backend
- `backend/internal/database/migrate/registry.go` — migration 011 (app_settings)
- `backend/pkg/opensubs/client.go` — OpenSubtitles.com REST v3 client
- `backend/pkg/podnapisi/client.go` — Podnapisi JSON API client
- `backend/internal/model/app_settings.go` — model
- `backend/internal/repository/app_settings.go` — repo
- `backend/internal/service/subtitle_search.go` — orchestrate both providers
- `backend/internal/handler/subtitle_search.go` — HTTP handlers
- `backend/cmd/server/routes.go` — register routes (modify)

### Frontend
- `webapp/src/api/subtitleSearch.ts` — API client functions
- `webapp/src/components/SubtitleSearchModal.tsx` — search UI
- `webapp/src/components/SubtitlePicker.tsx` — wire modal (modify)
- `webapp/src/hooks/stores/useMedia.ts` — add search/download mutations (modify)
