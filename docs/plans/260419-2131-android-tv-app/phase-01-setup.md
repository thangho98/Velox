# Phase 01: Setup Environment & Manifest
Status: ⬜ Pending

## Objective
Tạo module (hoặc flavor/variant) riêng biệt cho Android TV, cấu hình AndroidManifest để Google Play nhận diện đây là app TV, và nạp các thư viện Jetpack Compose for TV.

## Requirements
### Functional
- [x] Bổ sung TV flag vào `AndroidManifest.xml` (`<uses-feature android:name="android.software.leanback" android:required="true" />` và banner).
- [x] Khởi tạo Activity chính cho TV (`TvMainActivity`).
- [x] Import các thư viện dependencies của `androidx.tv:tv-foundation` và `androidx.tv:tv-material`.

### Non-Functional
- [ ] Không làm gãy build của bản Mobile hiện tại (Sử dụng chung codebase nhưng tách navigation/screens).
- [ ] Đặt font chữ và theme Material3 cơ bản cho bản TV.

## Implementation Steps
1. [x] Cập nhật `build.gradle.kts` của `:app` với thư viện Compose TV.
2. [x] Thêm file Banner (`banner_tv.png`) vào resources để làm icon app trên launcher TV.
3. [x] Tạo `TvMainActivity.kt` định tuyến tới `TvAppNavigation`.

## Files to Create/Modify
- `android/app/build.gradle.kts`
- `android/app/src/main/AndroidManifest.xml`
- `android/app/src/main/java/com/velox/app/TvMainActivity.kt`
- `android/app/src/main/java/com/velox/app/presentation/tv/TvAppNavigation.kt`

## Test Criteria
- [ ] App build và chạy được trên Emulator Android TV.
- [ ] App hiển thị logo Banner đúng chuẩn trên màn hình Home của TV.

---
Next Phase: Phase 02 (Login & Home Dashboard)
