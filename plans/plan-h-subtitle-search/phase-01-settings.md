# Phase 01: App Settings (Backend)
Status: ⬜ Pending
Dependencies: Migration 010 (library_paths)

## Objective
Tạo bảng `app_settings` để lưu OpenSubtitles credentials. Admin nhập 1 lần, dùng mãi.

## Tasks
- [ ] Migration 011: tạo bảng `app_settings`
- [ ] `internal/model/app_settings.go` — model + constants
- [ ] `internal/repository/app_settings.go` — Get/Set methods
- [ ] `internal/handler/settings.go` — GET + PUT /api/admin/settings
- [ ] Register routes + wire handler trong server startup

## Schema

```sql
-- Migration 011
CREATE TABLE app_settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Keys dùng
| Key | Mô tả |
|-----|-------|
| `opensubs_api_key` | App-level API key (từ opensubtitles.com/consumers) |
| `opensubs_username` | Username của admin |
| `opensubs_password` | Password (stored plaintext, self-hosted ok) |

## API

```
GET  /api/admin/settings          → { opensubs_api_key, opensubs_username, opensubs_password_set: bool }
PUT  /api/admin/settings          → body: { opensubs_api_key, opensubs_username, opensubs_password }
```

Lưu ý: GET không trả password thực, chỉ trả `password_set: true/false`.

## Files
- `backend/internal/database/migrate/registry.go` — thêm migration 011
- `backend/internal/database/migrate/011_app_settings.go` — SQL
- `backend/internal/model/app_settings.go`
- `backend/internal/repository/app_settings.go`
- `backend/internal/handler/settings.go`
