# Prompt: Implement Plan B — Auth & Sessions

## Bối cảnh dự án

Velox là self-hosted home media server (Go backend + React frontend). Backend dùng Go 1.26 + stdlib `net/http` (Go 1.22+ routing) + SQLite (WAL mode, `mattn/go-sqlite3`). Không dùng ORM, tất cả raw `database/sql`.

**Đã hoàn thành:** Plan A (migrations 001-006: libraries, media, media_files, series, seasons, episodes, genres, people, credits, scan_jobs, subtitles, audio_tracks). Migration system, DBTX interface, repositories, services, handlers đều đã có.

**Cần implement:** Plan B — Auth & Sessions (3 phases, chi tiết bên dưới).

---

## Quy tắc PHẢI tuân theo

### Architecture
- **Layer pattern:** Handler (parse request → call service → write JSON) → Service (business logic) → Repository (pure SQL) → Model (plain structs)
- **DBTX interface:** Tất cả repositories dùng `DBTX` interface (đã định nghĩa ở `internal/repository/db.go`) thay vì `*sql.DB` trực tiếp
- **Error pattern:** Dùng `service.ErrNotFound` (sentinel error wrapping `sql.ErrNoRows`, đã có ở `internal/service/errors.go`)
- **Response format:** `{"data": ...}` cho success, `{"error": "message"}` cho errors. Dùng `respondJSON()` và `respondError()` đã có ở `internal/handler/respond.go`
- **Context:** Luôn dùng `context.Context` là param đầu tiên trong service/repository methods
- **Receiver names:** 1-2 chữ viết tắt (`func (s *AuthService)`, `func (r *UserRepo)`)
- **Error wrapping:** `fmt.Errorf("doing X: %w", err)`

### Migrations
- Append new migrations vào `All()` trong `internal/database/migrate/registry.go`
- Mỗi migration là 1 function `up007(tx *sql.Tx) error` / `down007(tx *sql.Tx) error`
- Tham khảo schema chính xác từ `docs/database-design.md`

### Cái đã có (KHÔNG tạo lại)
- `internal/repository/db.go` — DBTX interface
- `internal/service/errors.go` — ErrNotFound
- `internal/handler/respond.go` — respondJSON, respondError, parseID, parseJSON
- `internal/middleware/middleware.go` — CORS, Logger, Recovery
- `internal/config/config.go` — Config struct
- `internal/database/` — Open, Migrate functions

---

## Phase 01: User Model & First-Run Setup

### Migration 007: users, user_preferences, user_library_access

Schema từ `docs/database-design.md`:

```sql
-- users
CREATE TABLE users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_admin      INTEGER DEFAULT 0,
    avatar_path   TEXT DEFAULT '',
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- user_preferences
CREATE TABLE user_preferences (
    user_id                INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    subtitle_language      TEXT DEFAULT '',
    audio_language         TEXT DEFAULT '',
    max_streaming_quality  TEXT DEFAULT 'auto',
    theme                  TEXT DEFAULT 'dark'
);

-- user_library_access
CREATE TABLE user_library_access (
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, library_id)
);
```

### Files to create

**`internal/auth/password.go`** — Password hashing:
- `HashPassword(plain string) (string, error)` — bcrypt cost 12
- `CheckPassword(hash, plain string) bool`
- Dependency: `golang.org/x/crypto/bcrypt`

**`internal/model/user.go`** — User struct:
```go
type User struct {
    ID           int64  `json:"id"`
    Username     string `json:"username"`
    DisplayName  string `json:"display_name"`
    PasswordHash string `json:"-"` // never expose
    IsAdmin      bool   `json:"is_admin"`
    AvatarPath   string `json:"avatar_path"`
    CreatedAt    string `json:"created_at"`
    UpdatedAt    string `json:"updated_at"`
}
```

**`internal/repository/user.go`** — UserRepo:
- `Create(ctx, user *model.User) (int64, error)`
- `GetByID(ctx, id int64) (*model.User, error)`
- `GetByUsername(ctx, username string) (*model.User, error)`
- `List(ctx) ([]*model.User, error)`
- `Update(ctx, user *model.User) error`
- `Delete(ctx, id int64) error`
- `Count(ctx) (int, error)` — dùng để detect first-run
- Constructor nhận DBTX, có WithTx method

**`internal/service/auth.go`** — AuthService:
- `Login(ctx, username, password string) (*model.User, error)`
- `ChangePassword(ctx, userID int64, oldPass, newPass string) error`
- `CreateUser(ctx, username, password, displayName string, isAdmin bool) (*model.User, error)` — validate input, hash password
- `IsConfigured(ctx) (bool, error)` — `Count() > 0`

**`internal/handler/setup.go`** — First-Run Setup:
- `GET /api/setup/status` → `{configured: bool}`
- `POST /api/setup` → tạo admin user (chỉ khi chưa configured)
  - Body: `{username, password, display_name}`
  - Validate: password >= 8 chars, username alphanumeric 3-32 chars
  - Sau khi configured: endpoint trả 403

**`internal/handler/auth.go`** — Auth endpoints:
- `POST /api/auth/login` → validate credentials, trả tokens (Phase 02 mới implement JWT, Phase 01 trả user info)
- `POST /api/auth/change-password` → require current password
- `GET /api/auth/me` → current user info (cần auth middleware từ Phase 02)
- `POST /api/auth/logout`

---

## Phase 02: JWT Auth & Middleware

### Migration 008: refresh_tokens, sessions

Schema từ `docs/database-design.md`:

```sql
-- refresh_tokens
CREATE TABLE refresh_tokens (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    device_name TEXT DEFAULT '',
    ip_address  TEXT DEFAULT '',
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_rt_user ON refresh_tokens(user_id);
CREATE INDEX idx_rt_expires ON refresh_tokens(expires_at);

-- sessions
CREATE TABLE sessions (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id          INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_id INTEGER REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    device_name      TEXT DEFAULT '',
    ip_address       TEXT DEFAULT '',
    user_agent       TEXT DEFAULT '',
    expires_at       DATETIME NOT NULL,
    last_active_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

### Files to create

**`internal/auth/jwt.go`** — JWT token management:
- `GenerateAccessToken(userID int64, isAdmin bool) (string, error)` — 15 min expiry
- `GenerateRefreshToken() (string, error)` — random 32 bytes, hex encoded
- `ValidateToken(tokenString string) (*Claims, error)`
- JWT secret: env `VELOX_JWT_SECRET`. Nếu không set → auto-generate random 32 bytes, persist to `{dataDir}/.jwt_secret`
- Dependency: `github.com/golang-jwt/jwt/v5`
- Claims struct: `UserID int64`, `IsAdmin bool`, embedded `jwt.RegisteredClaims`

**`internal/auth/context.go`** — Context helpers:
- `UserFromContext(ctx) (userID int64, isAdmin bool, ok bool)`
- `ContextWithUser(ctx, userID int64, isAdmin bool) context.Context`
- Dùng private context key type

**`internal/repository/session.go`** — SessionRepo + RefreshTokenRepo:
- RefreshToken: `Create`, `GetByTokenHash`, `Delete`, `DeleteExpired`, `DeleteByUserID`
- Session: `Create`, `GetByID`, `ListByUserID`, `Delete`, `UpdateLastActive`, `DeleteExpired`
- Hash refresh token bằng SHA256 trước khi lưu

**`internal/middleware/auth.go`** — Auth middleware:
- `RequireAuth(jwtValidator)` — extract Bearer token, validate JWT, set user in context
- `RequireAdmin` — check IsAdmin from context, return 403 if not
- Stream auth: `/api/stream/*` accept `?token=` query param (video player không set headers được)
- Return `{"error": "unauthorized"}` (401) hoặc `{"error": "forbidden"}` (403)

**Update `internal/handler/auth.go`:**
- `POST /api/auth/login` → validate credentials → tạo access token + refresh token + session → trả `{access_token, refresh_token, expires_in, user}`
- `POST /api/auth/refresh` → validate refresh token → rotate (xóa cũ, tạo mới) → trả new tokens
- `POST /api/auth/logout` → xóa refresh token + session
- `GET /api/auth/me` → trả user info từ context

**Update `internal/service/auth.go`:**
- `Login` trả thêm tokens
- `Refresh(ctx, refreshToken string) (*TokenPair, error)` — rotate refresh token
- `Logout(ctx, refreshToken string) error`

**Update `cmd/server/main.go`:**
- Thêm `JWTSecret` vào Config
- Route protection:
  - **Public:** `/api/setup/*`, `/api/auth/login`, `/api/auth/refresh`
  - **Authenticated:** tất cả `/api/*` khác
  - **Admin only:** `POST/DELETE /api/libraries/*`, tương lai `/api/admin/*`
  - **Stream auth:** `/api/stream/*` chấp nhận `?token=` query param
- Session tracking: update `last_active_at` (debounce max 1 update/minute)

**Update `internal/config/config.go`:**
- Thêm `JWTSecret string` — from `VELOX_JWT_SECRET`
- Thêm `DataDir string` — expose cho JWT secret file persistence

---

## Phase 03: Per-User State & Library ACL

### Migration 009: user_data, user_series_data (DROP old progress)

```sql
-- Drop old progress table
DROP TABLE IF EXISTS progress;

-- user_data (unified per-user-per-media state)
CREATE TABLE user_data (
    user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_id       INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    position       REAL DEFAULT 0,
    completed      INTEGER DEFAULT 0,
    is_favorite    INTEGER DEFAULT 0,
    rating         REAL DEFAULT NULL,
    play_count     INTEGER DEFAULT 0,
    last_played_at DATETIME DEFAULT NULL,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, media_id)
);
CREATE INDEX idx_ud_user ON user_data(user_id);
CREATE INDEX idx_ud_media ON user_data(media_id);

-- user_series_data (series-level favorite/rating)
CREATE TABLE user_series_data (
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    series_id   INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    is_favorite INTEGER DEFAULT 0,
    rating      REAL DEFAULT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, series_id)
);
```

### Files to create/update

**`internal/repository/user_data.go`** — UserDataRepo (replaces old ProgressRepo):
- `GetProgress(ctx, userID, mediaID int64) (*model.UserData, error)`
- `UpdateProgress(ctx, userID, mediaID int64, position float64, completed bool) error` — UPSERT
- `ToggleFavorite(ctx, userID, mediaID int64) error`
- `SetRating(ctx, userID, mediaID int64, rating *float64) error`
- `ListFavorites(ctx, userID int64, limit, offset int) ([]*model.UserData, error)`
- `ListRecentlyWatched(ctx, userID int64, limit int) ([]*model.UserData, error)`
- `GetSeriesProgress(ctx, userID, seriesID int64) (map[int]map[int]*model.UserData, error)` — watched/unwatched per season/episode

**`internal/handler/user.go`** — Admin User Management:
- `GET /api/admin/users` — list all users
- `POST /api/admin/users` — create user
- `PUT /api/admin/users/{id}` — update user (display_name, is_admin)
- `DELETE /api/admin/users/{id}` — delete user + cascade
  - Cannot delete self
  - Cannot remove last admin
- `PUT /api/admin/users/{id}/libraries` — set library access `{library_ids: [1,2,3]}`

**Update `internal/repository/user.go`:**
- `SetLibraryAccess(ctx, userID int64, libraryIDs []int64) error` — transaction: DELETE all + INSERT new
- `GetLibraryIDs(ctx, userID int64) ([]int64, error)`
- `GrantAllLibraries(ctx, userID int64) error` — cho user mới access tất cả libraries

**Refactor old progress code:**
- DELETE `internal/handler/progress.go` (đã xóa trong git status)
- DELETE `internal/service/progress.go` (đã xóa)
- DELETE `internal/repository/progress.go` (đã xóa)
- Progress endpoints mới extract `userID` từ context (auth middleware)

**User preferences endpoints:**
- `GET /api/me/preferences`
- `PUT /api/me/preferences`

**Profile endpoints:**
- `PUT /api/me` — update display_name
- `GET /api/me/sessions` — list active sessions
- `DELETE /api/me/sessions/{id}` — revoke session

---

## Thứ tự implement

Implement tuần tự Phase 01 → 02 → 03. Mỗi phase:
1. Tạo migration trước
2. Tạo model structs
3. Tạo repository
4. Tạo service
5. Tạo handler
6. Update `cmd/server/main.go` routes
7. Viết tests (table-driven, in-memory SQLite)

## Testing

- Table-driven tests: `tests := []struct{ name string; ... }{ ... }` + `t.Run(tt.name, ...)`
- In-memory SQLite: `sql.Open("sqlite3", ":memory:?_foreign_keys=on")`
- Test files cạnh source: `foo.go` → `foo_test.go`
- Test ít nhất: repository CRUD, service business logic (login, password change, first-run guard), middleware (valid/invalid/expired token, admin check)
- Run: `cd backend && make test`

## Dependencies cần thêm

```sh
cd backend
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
```

## Lưu ý quan trọng

1. **KHÔNG tạo default admin/admin.** First-run setup wizard bắt user tạo admin account.
2. **respondError format:** Dùng `respondError(w, status, "message")` — nó wrap trong `{"data": {"error": "message"}}`. Xem `internal/handler/respond.go`.
3. **Refresh token rotation:** Mỗi lần refresh → xóa token cũ, tạo token mới. Prevent replay attacks.
4. **Stream auth:** Video player (HTML5 `<video>`) không set Authorization header được → `/api/stream/*` phải accept `?token=<access_token>` query param.
5. **Session debounce:** Update `last_active_at` max 1 lần/phút. Tránh mỗi request đều write DB.
6. **Library ACL:** Admin thấy tất cả libraries. Non-admin chỉ thấy libraries trong `user_library_access`. User mới mặc định có access tất cả libraries hiện có.
7. **Cannot delete last admin:** Phải check trước khi delete/demote. Đếm admin count.
