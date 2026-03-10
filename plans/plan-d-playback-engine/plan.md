# Plan D: Playback Decision Engine
Status: ⬜ Pending
Priority: 🟡 High
Dependencies: Plan C (need working player to validate)

## Mục tiêu
Smart playback decisions: khi nào direct play, khi nào remux, khi nào transcode.
Giống cách Jellyfin/Emby quyết định - không phải "luôn transcode" hay "luôn direct play".

## Core Principle: Direct Play First
**Direct Play là ưu tiên số 1.** Zero CPU cost trên server.
HLS/transcode chỉ kick in khi:
- Video codec incompatible (HEVC trên Chrome)
- Audio codec incompatible (DTS/TrueHD → AAC)
- Container incompatible (MKV → remux MP4)
- User chọn non-default audio track (multi-audio switching)
- Image-based subtitle selected (PGS burn-in)

## Why This Matters
Hiện tại Velox có 2 mode: direct play hoặc full HLS transcode.
Nhưng thực tế có 4 playback paths:
1. **Direct Play** - container + video + audio đều compatible → serve file trực tiếp
2. **Direct Stream** - video codec OK nhưng container sai (MKV→MP4) → remux, no transcode
3. **Transcode Audio** - video OK nhưng audio cần convert (DTS→AAC)
4. **Full Transcode** - video codec incompatible (HEVC→H.264)

## Key Features
- **Multi-audio switching**: HLS với `#EXT-X-MEDIA:TYPE=AUDIO` groups
- **Dual subtitles**: custom overlay renderer (2 VTT tracks cùng lúc, tốt cho học ngôn ngữ)
- **Image subtitle burn-in**: PGS/VobSub → FFmpeg burn-in (forces transcode)

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | Client Capability Profiles | 5 tasks | ⬜ |
| 02 | Playback Decision Matrix | 7 tasks | ⬜ |
| 03 | Subtitle Serving & Dual Subs | 6 tasks | ⬜ |
