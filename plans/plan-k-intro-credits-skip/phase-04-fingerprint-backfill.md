# Phase 04: Fingerprint Backfill
Status: ✅ Complete
Plan: K - Intro / Credits Skip
Dependencies: Phase 01-03 stable in production

## Mục tiêu
Thêm khả năng backfill intro/credits marker cho media không có chapter marker, nhưng không làm rối contract đã ship ở các phase trước.

## Output của phase này
- ✅ Có detector abstraction đa nguồn
- ✅ Có execution model rõ ràng cho fingerprint jobs
- ✅ Có source priority policy ổn định
- ✅ Fingerprint marker chỉ là fallback, không phá chapter/manual marker

## Tasks
### 1. Define Detector Boundary
- [x] Add detector interface tách biệt khỏi chapter parser
- [x] Allow multiple sources: `chapter`, `fingerprint`, `manual`
- [x] Tách raw detection khỏi persistence để detector chỉ trả candidate markers
- [x] Define merge policy trong service layer, không embed vào detector

**Verify:** unit test detector interface có thể accept source mới mà không đổi playback contract ✅

### 2. Choose Execution Model
- [x] Decide backfill chạy lúc scan, scheduled task, hoặc admin-triggered rebuild
- [x] Gate bằng config hoặc feature flag
- [x] Khuyến nghị rollout: admin-triggered rebuild trước, scheduler sau
- [x] Xác định nơi đặt config: env var hay app settings

**Verify:** có documented trigger path và default system behavior không đổi khi flag tắt ✅

### 3. Implement Safe Fallback
- [x] Chỉ chạy fingerprint khi file chưa có chapter-based intro marker
- [x] Upsert marker mà không phá source ưu tiên cao hơn
- [x] Nếu cùng `marker_type`, source priority phải deterministic
- [x] Log confidence thấp để tiện audit accuracy

**Verify:** seed DB với chapter + fingerprint marker, service vẫn trả source ưu tiên đúng ✅

### 4. Verify
- [x] Add tests cho source priority và fallback behavior
- [x] Document operational flow trước khi bật mặc định
- [x] Add manual QA checklist cho episode có/không có chapter

**Files implemented:**
- `backend/internal/scanner/detector.go` - Detector interface & registry
- `backend/internal/scanner/detector_test.go` - Source priority tests
- `backend/internal/scanner/fingerprint_detector.go` - Fingerprint detector stub
- `backend/internal/scanner/marker_service.go` - Backfill logic & integration
- `backend/internal/scanner/marker_service_test.go` - Backfill behavior tests
- `backend/internal/handler/marker_admin.go` - Admin endpoints

## Rollout note
- Phase này không nên bắt đầu trước khi V1 chapter-first đã chạy ổn trên sample library thật.
- Nếu accuracy fingerprint không đủ tốt, vẫn giữ feature usable nhờ chapter path đã ship trước đó.
