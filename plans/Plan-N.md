# Plan N: Multi-language Support (i18n)

Support đa ngôn ngữ cho Velox UI, bắt đầu với English (default) và Vietnamese.

---

## Phase 01: Setup i18n Infrastructure

### 1.1 Install Dependencies
```bash
cd webapp
npm install react-i18next i18next i18next-browser-languagedetector
```

### 1.2 Create i18n Configuration
**File:** `webapp/src/i18n/index.ts`
- Khởi tạo i18next với LanguageDetector
- Default namespace: 'translation'
- Fallback language: 'en'
- Supported languages: ['en', 'vi']

### 1.3 Create Translation File Structure
```
webapp/src/locales/
├── en/
│   ├── common.json       # Shared UI strings
│   ├── auth.json         # Login, setup pages
│   ├── navigation.json   # Navbar, sidebar, menus
│   ├── settings.json     # Settings page (rất lớn)
│   ├── media.json        # Media detail, browse, search
│   ├── watch.json        # Watch page, player controls
│   └── errors.json       # Error messages
└── vi/
    └── (same structure)
```

### 1.4 Integrate vào App
**File:** `webapp/src/main.tsx`
- Import i18n config trước khi render App

### 1.5 Create useTranslation Hook Wrapper
**File:** `webapp/src/hooks/useTranslation.ts`
- Wrapper xung quanh react-i18next useTranslation
- Type-safe namespaces

**Deliverable:** i18n infrastructure chạy được, switch được giữa en/vi

---

## Phase 02: Extract English Strings

### 2.1 Common Strings
**File:** `locales/en/common.json`
```json
{
  "app": {
    "name": "Velox",
    "tagline": "Your personal media server"
  },
  "actions": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "edit": "Edit",
    "create": "Create",
    "close": "Close",
    "confirm": "Confirm",
    "loading": "Loading...",
    "saving": "Saving...",
    "scanning": "Scanning...",
    "searching": "Searching..."
  },
  "states": {
    "empty": "No items found",
    "error": "Something went wrong",
    "success": "Success!",
    "active": "Active",
    "inactive": "Inactive",
    "enabled": "Enabled",
    "disabled": "Disabled"
  }
}
```

### 2.2 Auth Strings
**File:** `locales/en/auth.json`
- Login page: "Sign In", "Username", "Password", "Signing in..."
- Setup page: "Create Admin Account", etc.
- Error messages: "Invalid credentials", etc.

### 2.3 Navigation Strings
**File:** `locales/en/navigation.json`
```json
{
  "nav": {
    "home": "Home",
    "movies": "Movies",
    "series": "Series",
    "browse": "Browse",
    "search": "Search"
  },
  "userMenu": {
    "settings": "Settings",
    "signOut": "Sign Out",
    "admin": "Administrator",
    "user": "User"
  }
}
```

### 2.4 Settings Strings (Largest)
**File:** `locales/en/settings.json`

Cấu trúc phân cấp:
```json
{
  "sections": {
    "profile": { "title": "Profile", "description": "..." },
    "preferences": { "title": "Preferences", "description": "..." },
    "security": { "title": "Security", "description": "..." },
    "sessions": { "title": "Sessions", "description": "..." },
    "metadata": { "title": "Metadata", "description": "..." },
    "subtitles": { "title": "Subtitles", "description": "..." },
    "playback": { "title": "Playback", "description": "..." },
    "cinema": { "title": "Cinema Mode", "description": "..." },
    "general": { "title": "Dashboard", "description": "..." },
    "libraries": { "title": "Libraries", "description": "..." },
    "users": { "title": "Users", "description": "..." },
    "activity": { "title": "Activity", "description": "..." },
    "tasks": { "title": "Tasks", "description": "..." },
    "webhooks": { "title": "Webhooks", "description": "..." }
  },
  "fields": {
    "username": "Username",
    "displayName": "Display Name",
    "password": "Password",
    "confirmPassword": "Confirm Password",
    "currentPassword": "Current Password",
    "newPassword": "New Password",
    "role": "Role",
    "subtitleLanguage": "Subtitle Language",
    "audioLanguage": "Audio Language",
    "maxQuality": "Max Streaming Quality",
    "theme": "Theme"
  },
  "options": {
    "theme": {
      "system": "System",
      "dark": "Dark",
      "light": "Light"
    },
    "quality": {
      "original": "Original",
      "4k": "4K",
      "1080p": "1080p",
      "720p": "720p",
      "480p": "480p"
    },
    "language": {
      "auto": "Auto",
      "en": "English",
      "vi": "Vietnamese"
    }
  },
  "providers": {
    "tmdb": {
      "name": "TMDb (The Movie Database)",
      "description": "TMDb provides posters, backdrops, plot summaries...",
      "customKey": "Custom API Key (v4 Read Access Token)"
    },
    "opensubtitles": {
      "name": "OpenSubtitles.com",
      "description": "Connect your OpenSubtitles account..."
    }
  }
}
```

### 2.5 Media Strings
**File:** `locales/en/media.json`
- Browse page: filters, sorts, "No results"
- Media detail: "Play", "Add to Favorites", "More Info", "Cast", "Crew", "Similar"
- Series detail: "Seasons", "Episodes", "Up Next"
- Search: "Search movies, series...", "Results for"

### 2.6 Watch Strings
**File:** `locales/en/watch.json`
- Player controls: Play, Pause, Volume, Fullscreen, "Next Episode"
- Subtitle picker: "Subtitles", "Off", "Custom..."
- Audio picker: "Audio", "Auto"
- Quality picker: "Quality", "Auto"
- Skip intro: "Skip Intro"
- Chromecast: "Cast to"

### 2.7 Update Components to Use Translations

Các files cần update (theo thứ tự ưu tiên):

**Priority 1 (User-facing):**
1. `Navbar.tsx` - navigation labels, user menu
2. `LoginPage.tsx` - form labels, errors
3. `SetupPage.tsx` - setup flow
4. `SettingsPage.tsx` - tất cả sections
5. `HomePage.tsx` - section titles
6. `BrowsePage.tsx` - filters, sorts
7. `SearchPage.tsx` - search UI
8. `MediaDetailPage.tsx` - action buttons, tabs
9. `SeriesDetailPage.tsx` - seasons, episodes
10. `WatchPage.tsx` - player controls

**Priority 2:**
- Các components trong `components/` folder
- Modal, Toast, Dialog content
- ActionMenu items
- Error messages

**Pattern:**
```tsx
// Before
<button>Sign In</button>

// After
import { useTranslation } from '@/hooks/useTranslation'
const { t } = useTranslation('auth')
<button>{t('signIn')}</button>
```

**Deliverable:** Tất cả strings trong UI đều lấy từ translation files

---

## Phase 03: Vietnamese Translation

### 3.1 Dịch Common
**File:** `locales/vi/common.json`
```json
{
  "app": {
    "name": "Velox",
    "tagline": "Máy chủ phương tiện cá nhân của bạn"
  },
  "actions": {
    "save": "Lưu",
    "cancel": "Hủy",
    "delete": "Xóa",
    "edit": "Sửa",
    "create": "Tạo",
    "close": "Đóng",
    "confirm": "Xác nhận",
    "loading": "Đang tải...",
    "saving": "Đang lưu...",
    "scanning": "Đang quét...",
    "searching": "Đang tìm..."
  }
}
```

### 3.2 Dịch Navigation & Auth
**File:** `locales/vi/navigation.json`
```json
{
  "nav": {
    "home": "Trang chủ",
    "movies": "Phim lẻ",
    "series": "Phim bộ",
    "browse": "Duyệt phim",
    "search": "Tìm kiếm"
  },
  "userMenu": {
    "settings": "Cài đặt",
    "signOut": "Đăng xuất",
    "admin": "Quản trị viên",
    "user": "Người dùng"
  }
}
```

**File:** `locales/vi/auth.json`
```json
{
  "login": {
    "title": "Đăng nhập",
    "username": "Tên đăng nhập",
    "password": "Mật khẩu",
    "signIn": "Đăng nhập",
    "signingIn": "Đang đăng nhập...",
    "newUser": "Mới dùng Velox?",
    "contactAdmin": "Liên hệ quản trị viên",
    "errors": {
      "required": "Vui lòng nhập tên đăng nhập và mật khẩu",
      "invalid": "Tên đăng nhập hoặc mật khẩu không đúng"
    }
  }
}
```

### 3.3 Dịch Settings (Phần lớn nhất)
**File:** `locales/vi/settings.json`

Các section chính:
- Profile → Hồ sơ
- Preferences → Tùy chọn
- Security → Bảo mật
- Sessions → Phiên đăng nhập
- Metadata → Metadata (hoặc Thông tin phương tiện)
- Subtitles → Phụ đề
- Playback → Phát lại
- Cinema Mode → Chế độ Rạp phim
- Dashboard → Bảng điều khiển
- Libraries → Thư viện
- Users → Người dùng
- Activity → Hoạt động
- Tasks → Tác vụ
- Webhooks → Webhooks

Ví dụ một số fields:
```json
{
  "fields": {
    "username": "Tên đăng nhập",
    "displayName": "Tên hiển thị",
    "password": "Mật khẩu",
    "confirmPassword": "Xác nhận mật khẩu",
    "currentPassword": "Mật khẩu hiện tại",
    "newPassword": "Mật khẩu mới",
    "role": "Vai trò",
    "subtitleLanguage": "Ngôn ngữ phụ đề",
    "audioLanguage": "Ngôn ngữ âm thanh",
    "maxQuality": "Chất lượng phát tối đa",
    "theme": "Giao diện"
  },
  "options": {
    "theme": {
      "system": "Theo hệ thống",
      "dark": "Tối",
      "light": "Sáng"
    },
    "quality": {
      "original": "Gốc",
      "4k": "4K",
      "1080p": "1080p",
      "720p": "720p",
      "480p": "480p"
    }
  }
}
```

### 3.4 Dịch Media & Watch
**File:** `locales/vi/media.json`
```json
{
  "actions": {
    "play": "Xem ngay",
    "addToFavorites": "Thêm vào yêu thích",
    "removeFromFavorites": "Xóa khỏi yêu thích",
    "moreInfo": "Thông tin thêm",
    "trailer": "Trailer",
    "cast": "Diễn viên",
    "crew": "Đoàn làm phim",
    "similar": "Phim tương tự"
  },
  "filters": {
    "genre": "Thể loại",
    "year": "Năm",
    "rating": "Đánh giá",
    "sort": "Sắp xếp"
  },
  "search": {
    "placeholder": "Tìm phim, series, thể loại...",
    "results": "Kết quả cho",
    "noResults": "Không tìm thấy kết quả"
  }
}
```

**File:** `locales/vi/watch.json`
```json
{
  "controls": {
    "play": "Phát",
    "pause": "Tạm dừng",
    "mute": "Tắt tiếng",
    "unmute": "Bật tiếng",
    "fullscreen": "Toàn màn hình",
    "exitFullscreen": "Thoát toàn màn hình",
    "nextEpisode": "Tập tiếp theo",
    "skipIntro": "Bỏ qua giới thiệu",
    "skipCredits": "Bỏ qua credit"
  },
  "subtitles": {
    "title": "Phụ đề",
    "off": "Tắt",
    "custom": "Tùy chỉnh...",
    "search": "Tìm phụ đề"
  },
  "audio": {
    "title": "Âm thanh",
    "auto": "Tự động"
  },
  "quality": {
    "title": "Chất lượng",
    "auto": "Tự động"
  },
  "casting": {
    "title": "Truyền tới",
    "disconnect": "Ngắt kết nối"
  }
}
```

### 3.5 Dịch Errors
**File:** `locales/vi/errors.json`
```json
{
  "generic": "Đã xảy ra lỗi",
  "network": "Lỗi kết nối mạng",
  "notFound": "Không tìm thấy",
  "unauthorized": "Phiên đăng nhập hết hạn",
  "forbidden": "Không có quyền truy cập",
  "validation": "Dữ liệu không hợp lệ",
  "server": "Lỗi máy chủ"
}
```

**Deliverable:** File dịch tiếng Việt hoàn chỉnh, có thể chuyển đổi ngôn ngữ

---

## Phase 04: Language Switcher & Persistence

### 4.1 Language Switcher Component
**File:** `webapp/src/components/LanguageSwitcher.tsx`
- Dropdown chọn ngôn ngữ (en/vi)
- Hiển thị flag/icon + tên ngôn ngữ
- Vị trí: Settings → Preferences hoặc User Menu

### 4.2 Persist Language Preference
**Backend:**
- Migration: Thêm column `language` vào table `user_preferences` (default: 'en')
- API: PATCH /api/users/preferences/language
- Response: Include language trong preferences

**Frontend:**
- Hook: `useLanguage()` - đồng bộ với backend
- On mount: Load language từ preferences → set i18n
- On change: Update i18n + call API update preference

### 4.3 Language Detector Config
**File:** `webapp/src/i18n/index.ts`
```ts
import LanguageDetector from 'i18next-browser-languagedetector'

i18n
  .use(LanguageDetector)
  .init({
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'velox-language',
      caches: ['localStorage']
    }
  })
```

**Deliverable:** Người dùng có thể chọn ngôn ngữ, preference được lưu và đồng bộ

---

## Phase 05: Polish & Edge Cases

### 5.1 RTL Considerations (Future-proof)
- Markup hỗ trợ RTL (dir="auto" hoặc dir="rtl")
- CSS logical properties (margin-inline-start thay vì margin-left)

### 5.2 Number/Date/Time Formatting
- Dùng Intl.DateTimeFormat cho locale-aware dates
- Dùng Intl.NumberFormat cho số lượng, dung lượng

### 5.3 Pluralization
- i18next hỗ trợ plural rules
- Ví dụ: "1 library" vs "2 libraries" / "1 thư viện" vs "2 thư viện"

### 5.4 Testing
- Test switch language trong Settings
- Test refresh page - language persist
- Test API error messages hiển thị đúng ngôn ngữ

---

## Estimated Effort

| Phase | Effort | Files Modified |
|-------|--------|----------------|
| Phase 01 | 2-3h | 3-4 files |
| Phase 02 | 6-8h | 15-20 files |
| Phase 03 | 4-5h | 7-8 files (vi) |
| Phase 04 | 3-4h | 5-6 files |
| Phase 05 | 2h | 3-4 files |
| **Total** | **17-22h** | **~30 files** |

---

## Files to Create

```
webapp/src/
├── i18n/
│   └── index.ts              # i18n config
├── locales/
│   ├── en/
│   │   ├── common.json
│   │   ├── auth.json
│   │   ├── navigation.json
│   │   ├── settings.json
│   │   ├── media.json
│   │   ├── watch.json
│   │   └── errors.json
│   └── vi/
│       ├── common.json
│       ├── auth.json
│       ├── navigation.json
│       ├── settings.json
│       ├── media.json
│       ├── watch.json
│       └── errors.json
└── hooks/
    └── useTranslation.ts     # Typed wrapper
```

## Backend Migration

```sql
-- 022_language_preference.sql
ALTER TABLE user_preferences ADD COLUMN language TEXT DEFAULT 'en' CHECK (language IN ('en', 'vi'));
```

---

## Notes

- Dùng nested keys để tổ chức: `t('settings.sections.profile.title')`
- Tránh hardcode strings trong logic (error messages từ API cần xử lý riêng)
- TypeScript: Generate type definitions từ JSON files
- Netflix/Vietnamese style: Giữ tone giống Netflix (thân thiện, đơn giản)
