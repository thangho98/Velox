# Phase 02: Playback Contract
Status: ⬜ Pending
Plan: K - Intro / Credits Skip
Dependencies: Phase 01

## Mục tiêu
Expose marker data qua playback payload hiện có để player lấy đầy đủ `stream_url`, `resume position`, `audio/subtitle track`, và `skip_segments` chỉ trong một request.

## Output của phase này
- Backend resolve marker cho đúng file đang phát
- `POST /api/playback/{id}/info` trả `skip_segments`
- TypeScript types được cập nhật đồng bộ
- Có test đảm bảo contract không regress

## Tasks
### 1. Add Service Layer
- [ ] Add marker service để resolve marker theo `media_file_id`
- [ ] Source priority (locked): `manual > chapter > fingerprint` — nếu cùng `marker_type` có nhiều source, chỉ trả source cao nhất
- [ ] Define priority policy trong service thay vì rải rác ở handler
- [ ] Transform `model.MediaMarker` → `SkipSegment` DTO (API field name: `skip_segments`, không dùng tên table `media_markers`)

**Verify:** service test trả marker sorted, filter invalid marker, apply source priority đúng

### 2. Extend Playback Response
- [ ] Add `skip_segments` vào response của `POST /api/playback/{id}/info`
- [ ] Mỗi segment gồm: `type`, `start`, `end`, `source`, `confidence`
- [ ] Giữ field optional để media không có marker vẫn backward-compatible
- [ ] Không mở endpoint mới trừ khi có blocker thực sự

**Verify:** handler response JSON có `skip_segments` khi file có marker và omit/empty hợp lý khi không có

### 3. Wire Current File Context
- [ ] Resolve marker theo `primaryFile.ID` hoặc `media_file_id` được client chọn
- [ ] Chỉ trả segment hợp lệ cho file/version đang phát
- [ ] Hook marker service vào `PlaybackHandler` constructor và `backend/cmd/server/main.go`
- [ ] Đảm bảo playback info cho alternate version trả marker của đúng version đó

**Verify:** request với `media_file_id` khác nhau trả segment khác nhau nếu DB seed khác nhau

### 4. Verify
- [ ] Add handler tests cho payload có và không có `skip_segments`
- [ ] Add TypeScript types cho playback contract mới
- [ ] Add regression test đảm bảo query key trong `useMedia.ts` không đổi semantics

## Files dự kiến
- `backend/internal/service/media_marker.go` - NEW
- `backend/internal/handler/playback.go`
- `backend/internal/handler/playback_test.go`
- `backend/cmd/server/main.go`
- `webapp/src/types/api.ts`
- `webapp/src/hooks/stores/useMedia.ts`

## Notes
- Nếu phase này làm đúng, frontend phase sau gần như chỉ là pure UI state machine.
- Ưu tiên inject service vào handler thay vì cho handler query repo trực tiếp để giữ layering nhất quán.
