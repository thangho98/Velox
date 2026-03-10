# Plan C: Web App MVP
Status: ⬜ Pending
Priority: 🔴 Critical (validate everything built so far)
Dependencies: Plan A + B

## Mục tiêu
Ship bản web client usable đầu tiên. User flow hoàn chỉnh:
Setup → Login → Browse Library → Movie/Series Detail → Play Video → Resume.

Đây là validation loop: nếu UX sai → fix backend sớm thay vì muộn.

## Target UX
```
First visit → Setup Page (tạo admin)
  ↓
Login → Home (Continue Watching, Recently Added, Genre Rows)
  ↓
Browse → Movie Grid / Series Grid (poster + title + year + rating)
  ↓
Detail → Poster, Backdrop, Overview, Cast, Genres, Play Button
  ↓ (Series: Season picker → Episode list)
Play → Video Player (hls.js, subtitle track selector, resume position)
  ↓
Resume → Next time: pick up where you left off
```

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | App Shell & Auth UI | 7 tasks | ⬜ |
| 02 | Browse & Detail Pages | 8 tasks | ⬜ |
| 03 | Video Player | 7 tasks | ⬜ |
| 04 | Home Screen & Polish | 6 tasks | ⬜ |

## Tech (Frontend)
- React 19 + TypeScript
- Vite 6
- TailwindCSS 4
- React Router v7
- hls.js (HLS playback)
- Axios or fetch (API client)
