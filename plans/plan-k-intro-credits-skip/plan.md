# Plan K: Intro / Credits Skip
Created: 2026-03-15
Status: ✅ Complete (4/4 phases)
Priority: 🟡 Medium
Dependencies: Plans A-F done, overlaps Plan G / Phase 02 (Chapter Support)

## Overview
Thêm `Skip Intro` và `Skip Credits` cho player bằng cách detect marker theo 2 nguồn:
- **V1:** chapter marker từ file video, vì repo đã có `ffprobe`, scan pipeline, và playback API phù hợp để ship nhanh.
- **V2:** audio fingerprint backfill cho episode không có chapter, theo hướng tương tự ConfusedPolarBear/Jellyfin.

Feature này nên được triển khai theo hướng **marker-first** chứ không phải **plugin-first**:
- Backend chỉ cần biết có segment nào cần skip, từ nguồn nào, độ tin cậy bao nhiêu
- Player chỉ tiêu thụ `skip_segments` và show CTA đúng thời điểm
- Nguồn marker (`chapter`, `fingerprint`, `manual`) là implementation detail, không làm rối API contract

## Hiện trạng
- Backend đã có scan pipeline tại `backend/internal/scanner/pipeline.go`
- Metadata media file đang lấy bằng `backend/pkg/ffprobe/ffprobe.go`
- Player dùng `POST /api/playback/{id}/info` để lấy playback metadata
- Frontend player nằm ở `webapp/src/pages/WatchPage.tsx`
- Chưa có table/model/repo cho marker kiểu intro/credits
- `backend/cmd/server/main.go` đang wire repo/service/handler theo pattern rõ ràng, nên có thể thêm marker repo/service mà không phải đổi kiến trúc
- `webapp/src/hooks/stores/useMedia.ts` đang dùng một query key chung cho playback info, thuận tiện để nhét `skip_segments` vào payload hiện có

## Product Goal
Khi user xem movie/episode trong web player:
1. Nếu thời gian phát đang nằm trong segment intro hoặc credits đã detect, player hiện CTA tương ứng
2. Click CTA sẽ seek tới `segment.end` ngay lập tức
3. Nếu media có nhiều file version, marker phải bám theo đúng file đang phát
4. Không làm phát sinh thêm round-trip API riêng cho skip marker

## Success Criteria
- Với file có chapter tên `Intro` hoặc `Credits`, backend trả `skip_segments` chính xác qua `POST /api/playback/{id}/info`
- `WatchPage` hiện `Skip Intro` trong khoảng thời gian đúng, click xong tua đến cuối segment
- HLS và direct play đều hoạt động như nhau vì logic skip nằm ở `video.currentTime`, không phụ thuộc playback mode
- Không regress subtitle/audio selection hoặc playback info cache hiện tại
- Có test backend cho parsing + payload và test frontend cho CTA visibility/seek

## Why This Design
### Chapter-first
- Nhanh ship nhất vì dữ liệu đã nằm trong container metadata
- Không cần worker/audio analysis ngay từ đầu
- Có thể tận dụng luôn cho feature chapter listing sau này

### Marker per `media_file`
- Cùng một media có thể có bản remux, encode, theatrical cut, director's cut
- Intro window có thể khác nhau giữa các version
- DB hiện tại cũng đang gắn subtitle/audio track theo `media_file_id`, nên marker theo `media_file_id` là nhất quán

### Extend existing playback payload
- `POST /api/playback/{id}/info` đã là nguồn truth cho stream URL, subtitle/audio track, resume position
- Thêm `skip_segments` vào đây giúp frontend không phải call thêm endpoint
- Tránh race condition giữa playback info và marker info

## Proposed Data Model
### Table: `media_markers`
- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE`
- `marker_type TEXT NOT NULL CHECK (marker_type IN ('intro','credits'))`
- `start_sec REAL NOT NULL`
- `end_sec REAL NOT NULL`
- `source TEXT NOT NULL CHECK (source IN ('chapter','fingerprint','manual'))`
- `confidence REAL NOT NULL DEFAULT 1`
- `label TEXT NOT NULL DEFAULT ''`
- `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`
- `updated_at DATETIME DEFAULT CURRENT_TIMESTAMP`

### Indexes / constraints
- Unique logical segment:
  - `(media_file_id, marker_type, source, start_sec, end_sec)`
- Lookup index:
  - `(media_file_id, marker_type)`
- Validation rule:
  - `start_sec >= 0`
  - `end_sec > start_sec`

## Proposed API Contract
`POST /api/playback/{id}/info`

Thêm field:

```json
{
  "skip_segments": [
    {
      "type": "intro",
      "start": 12.5,
      "end": 86.4,
      "source": "chapter",
      "confidence": 1
    }
  ]
}
```

### Response rules
- Chỉ trả marker cho đúng `primary_file_id` hoặc `media_file_id` được request
- Sort theo `start ASC`
- Source priority (locked): `manual > chapter > fingerprint`. Nếu cùng `marker_type` có nhiều source, chỉ trả source cao nhất.
- V1 chỉ có `chapter`, nhưng contract phải generic ngay từ đầu
- **Naming:** Table `media_markers` → API field `skip_segments`. Service layer transform `model.MediaMarker` → `SkipSegment` DTO.

## Detection Rules
### V1: Chapter-based
- Parse chapter title + start/end từ ffprobe
- Normalize title về lowercase, strip punctuation
- Match vào danh sách alias:
  - Intro: `intro`, `opening`, `opening credits`, `theme`, `title sequence`
  - Credits: `credits`, `end credits`, `closing credits`
- Reject segment nếu:
  - `end <= start`
  - quá ngắn, ví dụ `< 5s`
  - intro xuất hiện quá muộn, ví dụ `start > 15m` nếu muốn thêm heuristic

### V2: Fingerprint-based
- Chỉ chạy khi chưa có intro marker chất lượng đủ tốt
- Có config/feature flag
- Ghi `source = fingerprint`, `confidence < 1`

## Cần làm
1. Persist marker theo từng `media_file` để support nhiều version của cùng media
2. Trả marker qua playback payload hiện có, tránh mở thêm endpoint nếu không cần
3. Hiện nút skip trong player đúng thời điểm, seek tới cuối segment
4. Chừa abstraction để sau này backfill bằng audio fingerprint mà không phải đổi contract API

## Scope
- In:
  - Chapter marker parsing + persistence
  - Playback API trả `skip_segments`
  - Web player `Skip Intro` / `Skip Credits`
  - Detector abstraction cho fingerprint phase sau
- Out:
  - Plugin Emby/Jellyfin integration
  - Manual marker editor
  - Mobile/native client

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | Marker Ingestion | 5 tasks | ✅ Complete |
| 02 | Playback Contract | 4 tasks | ✅ Complete |
| 03 | Player UX | 5 tasks | ✅ Complete |
| 04 | Fingerprint Backfill | 4 tasks | ✅ Complete |

## Delivery Strategy
### Ship 1: Chapter-backed Skip Intro
- Create table + model/repo/service
- Parse chapter metadata
- Return `skip_segments`
- Show `Skip Intro`

### Ship 2: Skip Credits
- Reuse same plumbing
- Bật CTA thứ hai nếu data đủ tốt
- Verify UX không gây phiền khi user gần hết phim

### Ship 3: Fingerprint Backfill
- Add detector abstraction
- Add rebuild path / scheduler
- Keep optional until accuracy ổn

## Files Expected to Change
### Backend
- `backend/pkg/ffprobe/ffprobe.go`
- `backend/internal/model/*.go`
- `backend/internal/repository/*.go`
- `backend/internal/service/*.go`
- `backend/internal/scanner/pipeline.go`
- `backend/internal/handler/playback.go`
- `backend/internal/database/migrate/registry.go`
- `backend/cmd/server/main.go`

### Frontend
- `webapp/src/types/api.ts`
- `webapp/src/hooks/stores/useMedia.ts`
- `webapp/src/pages/WatchPage.tsx`
- Có thể thêm helper riêng trong `webapp/src/components/watch/`

## Verification Strategy
### Backend
- Unit test parse chapter title normalization
- Unit test marker priority resolution
- Repository tests cho CRUD/upsert/list by `media_file_id`
- Handler test đảm bảo `skip_segments` xuất hiện đúng trong playback response

### Frontend
- Test segment activation theo `currentTime`
- Test click CTA sẽ seek đúng `segment.end`
- Test CTA không hiện lại sau khi skip cùng segment

### Manual
- 1 movie có chapter intro
- 1 episode có chapter credits
- 1 media không có marker
- 1 media HLS transcode
- 1 media direct play

## Risks
- Nhiều file không có chapter hữu ích, nên V1 coverage thấp hơn kỳ vọng
- Chapter title thực tế rất bẩn, cần alias list đủ rộng
- Credits CTA có thể gây phiền nếu hiện quá sát auto-next hoặc kết thúc media
- Nếu source priority không rõ ràng từ đầu, phase fingerprint sẽ khó chồng lên V1

## Decisions Locked
- **Skip Intro + Skip Credits ship cùng V1.** Cùng schema, cùng CTA component, chỉ khác label. Credits CTA không hiện nếu Up Next đang active.
- **Source priority: `manual > chapter > fingerprint`.** Manual là user override (luôn cao nhất). Chapter có sẵn trong container (confidence=1.0). Fingerprint là heuristic (confidence<1.0).
- Fingerprint backfill: admin-triggered rebuild trước, scheduler sau (Phase 04).
- `manual` source có sẵn trong schema từ đầu, dù V1 chưa có UI tạo manual marker.

## Notes
- Plan này hấp thụ phần "chapter support để skip intro" đang ghi rất ngắn trong `plans/plan-g-nice-to-have/phase-02-chapters.md`.
- Nên ship **chapter-first** để giảm scope và có UX sớm; fingerprint chỉ nên bật sau khi contract marker ổn định.
- Nếu muốn scope gọn nhất, có thể tách release đầu thành **Intro only** rồi mới bật Credits.

## Quick Commands
- Start: `/code phase-01`
- Check: `/next`
