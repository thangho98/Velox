# Plan Q: Setup Wizard (Onboarding)
Created: 2026-03-27
Status: 🟡 In Progress

## Overview
Thêm Setup Wizard sau khi tạo admin account lần đầu. Hướng dẫn user cấu hình metadata providers, subtitle, playback, pretranscode, cinema mode trước khi bắt đầu dùng.

## Flow
```
SetupPage (create admin) → auto-login → SetupWizardPage (6 steps) → HomePage
```

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Backend: wizard completed flag + API | ⬜ Pending | 0% |
| 02 | Frontend: Wizard shell (stepper, navigation) | ⬜ Pending | 0% |
| 03 | Frontend: 6 step components | ⬜ Pending | 0% |
| 04 | Integration: routing, redirect, testing | ⬜ Pending | 0% |

## Steps Detail

### Step 1: Metadata API Keys
- TMDb (has_builtin), OMDb, TheTVDB, Fanart.tv
- Show status badge per provider
- Explain: "Velox dùng TMDb để tự động lấy poster, mô tả phim"

### Step 2: Subtitle Providers
- OpenSubtitles (API key + account), Subdl, DeepL
- Auto-download language preference
- Explain: "Tự động tải phụ đề cho phim"

### Step 3: Playback Settings
- Playback mode: Auto vs Direct Play
- Explain: "Auto = tự chuyển đổi video nếu thiết bị không hỗ trợ"

### Step 4: Pre-transcode
- Enable/disable toggle
- Profile selection (480p/720p/1080p/1440p/4K)
- Schedule selection
- Explain: "Chuyển đổi phim trước để xem không giật lag"

### Step 5: Cinema Mode
- Enable/disable toggle
- Max trailers
- Explain: "Tự động phát trailer trước khi xem phim, giống rạp chiếu"

### Step 6: Summary + Add Library
- Show tất cả settings đã cấu hình
- Nút "Add Library" để tạo thư viện media đầu tiên
- Nút "Finish Setup" để hoàn thành

## Design Decisions
- Skip-friendly: mọi step đều optional, có nút Skip
- Chỉ hiện 1 lần: flag `setup_wizard_completed` trong app_settings
- Reuse hooks: dùng lại useSettings hooks đã có
- Auto-login sau create admin: wizard cần authenticated
