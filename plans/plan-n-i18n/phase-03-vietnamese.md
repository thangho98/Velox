# Phase 03: Vietnamese Translation

Status: ⬜ Pending
Dependencies: Phase 02

---

## Objective

Dịch toàn bộ UI strings sang tiếng Việt (Vietnamese).

---

## Requirements

### Functional
- [ ] Dịch common.json (app info, actions, states)
- [ ] Dịch auth.json (login, setup, errors)
- [ ] Dịch navigation.json (navbar, menus)
- [ ] Dịch settings.json (all sections, fields, options)
- [ ] Dịch media.json (actions, filters, search)
- [ ] Dịch watch.json (player controls)
- [ ] Dịch errors.json (error messages)

### Non-Functional
- [ ] Tone: Netflix-style (thân thiện, đơn giản)
- [ ] Technical terms: giữ nguyên hoặc dịch phổ biến
- [ ] Consistent terminology

---

## Translation Files

### Common (`locales/vi/common.json`)
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
    "searching": "Đang tìm...",
    "signIn": "Đăng nhập",
    "signOut": "Đăng xuất"
  },
  "states": {
    "empty": "Không có mục nào",
    "error": "Đã xảy ra lỗi",
    "success": "Thành công!",
    "active": "Đang hoạt động",
    "inactive": "Không hoạt động",
    "enabled": "Đã bật",
    "disabled": "Đã tắt"
  }
}
```

### Auth (`locales/vi/auth.json`)
```json
{
  "login": {
    "title": "Đăng nhập",
    "username": "Tên đăng nhập",
    "password": "Mật khẩu",
    "signingIn": "Đang đăng nhập...",
    "newUser": "Mới dùng Velox?",
    "contactAdmin": "Liên hệ quản trị viên"
  },
  "setup": {
    "title": "Tạo tài khoản quản trị",
    "description": "Thiết lập máy chủ Velox của bạn"
  },
  "errors": {
    "required": "Vui lòng nhập tên đăng nhập và mật khẩu",
    "invalid": "Tên đăng nhập hoặc mật khẩu không đúng"
  }
}
```

### Navigation (`locales/vi/navigation.json`)
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
  },
  "search": {
    "placeholder": "Tìm phim, thể loại..."
  }
}
```

### Settings (`locales/vi/settings.json`)
```json
{
  "sections": {
    "profile": {
      "title": "Hồ sơ",
      "description": "Quản lý thông tin tài khoản"
    },
    "preferences": {
      "title": "Tùy chọn",
      "description": "Tùy chỉnh trải nghiệm xem phim"
    },
    "security": {
      "title": "Bảo mật",
      "description": "Thay đổi mật khẩu"
    },
    "sessions": {
      "title": "Phiên đăng nhập",
      "description": "Quản lý phiên hoạt động"
    },
    "metadata": {
      "title": "Metadata",
      "description": "Cấu hình nhà cung cấp metadata"
    },
    "subtitles": {
      "title": "Phụ đề",
      "description": "Cấu hình nhà cung cấp phụ đề"
    },
    "playback": {
      "title": "Phát lại",
      "description": "Chính sách phát lại toàn server"
    },
    "cinema": {
      "title": "Chế độ Rạp phim",
      "description": "Phát trailer trước phim chính"
    },
    "general": {
      "title": "Bảng điều khiển",
      "description": "Thông tin và trạng thái máy chủ"
    },
    "libraries": {
      "title": "Thư viện",
      "description": "Quản lý thư viện phương tiện"
    },
    "users": {
      "title": "Người dùng",
      "description": "Quản lý tài khoản người dùng"
    },
    "activity": {
      "title": "Hoạt động",
      "description": "Hoạt động gần đây trên máy chủ"
    },
    "tasks": {
      "title": "Tác vụ",
      "description": "Tác vụ nền và bảo trì"
    },
    "webhooks": {
      "title": "Webhooks",
      "description": "Cấu hình thông báo webhook"
    }
  },
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
    },
    "language": {
      "auto": "Tự động",
      "en": "Tiếng Anh",
      "vi": "Tiếng Việt"
    }
  }
}
```

### Media (`locales/vi/media.json`)
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
  },
  "library": {
    "movies": "Phim lẻ",
    "tvshows": "Phim bộ",
    "mixed": "Hỗn hợp"
  }
}
```

### Watch (`locales/vi/watch.json`)
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

### Errors (`locales/vi/errors.json`)
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

---

## Translation Guidelines

### Tone & Style
- **Netflix-style**: Thân thiện, đơn giản, dễ hiểu
- **Không quá formal**: "Đăng xuất" thay vì "Đăng xuất khỏi hệ thống"
- **Active voice**: "Xem ngay" thay vì "Bấm để xem"

### Technical Terms
| English | Vietnamese | Notes |
|---------|------------|-------|
| Library | Thư viện | |
| Metadata | Metadata | Giữ nguyên (hoặc "Thông tin phương tiện") |
| Trailer | Trailer | Giữ nguyên |
| Webhook | Webhook | Giữ nguyên |
| Dashboard | Bảng điều khiển | Hoặc "Tổng quan" |
| Cast | Diễn viên | |
| Crew | Đoàn làm phim | |
| Episode | Tập | |
| Season | Mùa/Mùa phim | |
| Series | Phim bộ/Series | |
| Movies | Phim lẻ | |

### Common Patterns
- "Save" → "Lưu" (không phải "Lưu lại")
- "Cancel" → "Hủy" (không phải "Hủy bỏ")
- "Delete" → "Xóa"
- "Edit" → "Sửa"
- "Create" → "Tạo"

---

## Test Criteria

- [ ] Switch language sang Vietnamese
- [ ] Tất cả UI hiển thị tiếng Việt
- [ ] Không có string nào còn English
- [ ] Fallback về English nếu key missing
- [ ] localStorage lưu 'vi'

---

## Notes

- Dùng ngôn ngữ đời thường, không quá technical
- Test với người dùng thật nếu có thể
- Sẵn sàng điều chỉnh sau feedback

---

Next Phase: [Phase 04 - Language Switcher & Persistence](./phase-04-switcher.md)
