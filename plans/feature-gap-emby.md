# Feature Gap: Velox vs Emby
Created: 2026-03-13
Status: 📋 Checklist — chờ anh chọn feature để lên plan

## Đã có ✅
- [x] Library scan + file watcher (fsnotify, debounce, rename tracking)
- [x] Metadata (TMDb, OMDb, TheTVDB, Fanart.tv, TVmaze, NFO override)
- [x] Multi-user + JWT auth (access + refresh + session)
- [x] Web player (Direct Play → Direct Stream → Transcode)
- [x] ABR adaptive bitrate (480p/720p/1080p)
- [x] HW accelerated transcode (VideoToolbox/VAAPI/NVENC)
- [x] Trickplay thumbnails (sprite sheets + WebVTT)
- [x] Subtitle discovery + dual subs + online search (OpenSubtitles + Podnapisi)
- [x] Admin dashboard + webhooks (HMAC-SHA256)
- [x] Activity logging + scheduled tasks (in-memory scheduler)
- [x] Watch progress / resume

## Plan G (đã có plan, chưa implement) 🟡
- [ ] SyncPlay — watch party, WebSocket sync
- [ ] Chapter support — chapter markers từ MKV/MP4
- [ ] Favorites / Playlists / Collections

## Đã implement (sau gap analysis) ✅
- [x] **Continue Watching / Next Up** — Plan I ✅ Done
- [x] **Intro/Credits skip** — Plan K ✅ Done (chapter + chromaprint + blackframe)
- [x] **Favorites** — Done (toggle + page + filter)

## Chưa có — Ưu tiên trung bình 🟡
- [ ] **Parental Controls** — Content rating filter per user (G/PG/PG-13/R), PIN lock cho profile. Cần thêm column vào users + filter logic.
- [ ] **Music library** — Hỗ trợ audio-only (FLAC/MP3/AAC), album art, artist pages, gapless playback. Cần media_type='music' + scan logic riêng.
- [x] **Media info editor** — Plan L 🟡 Planned (4 phases: PATCH API + image upload + editor UI + NFO write)
- [ ] **DLNA / Chromecast** — Cast tới Smart TV, Chromecast. DLNA = UPnP server, Chromecast = Google Cast SDK.
- [ ] **Mobile apps** — Android/iOS native. Scope hiện tại web-only, có thể wrap bằng Capacitor/Expo sau.

## Chưa có — Ưu tiên thấp 🟢
- [ ] **Offline sync** — Download media về device xem offline, sync progress khi reconnect.
- [ ] **Photo library** — Browse ảnh theo folder/date, slideshow, EXIF metadata.
- [ ] **Book/Comic library** — EPUB, PDF, CBZ reader trong browser.
- [ ] **Cinema mode** — Trailers + custom intro video trước phim chính.
- [ ] **Live TV & DVR** — TV tuner (HDHomeRun), EPG guide, scheduled recording.
- [ ] **Notifications** — Push notification khi scan xong, media mới được thêm, transcode hoàn tất.

## Gợi ý thứ tự triển khai
1. **Continue Watching / Next Up** — impact lớn nhất, effort thấp nhất (đã có data)
2. **Media info editor** — admin cần sửa metadata sai
3. **Intro/Credits skip** — UX nâng cấp đáng kể
4. **Parental Controls** — cần nếu dùng chung gia đình
5. Plan G (SyncPlay, Chapters, Favorites)
6. Còn lại tùy nhu cầu
