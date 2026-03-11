# Prompt: Fix Plan B Review Issues

## Context

Velox là self-hosted home media server. Go 1.26 + stdlib `net/http` (Go 1.22+ routing) + SQLite (WAL mode, `mattn/go-sqlite3`). Không dùng ORM.

Plan B (Auth & Sessions) đã được implement nhưng code review phát hiện 16 issues. File review: `plans/plan-b-auth-sessions/review.md`. Fix tất cả issues theo thứ tự bên dưới.

**Sau khi fix mỗi issue, đánh dấu `[x]` trong `plans/plan-b-auth-sessions/review.md`.**

---

## Quy tắc PHẢI tuân theo

- **Layer pattern:** Handler → Service → Repository → Model
- **DBTX interface:** Tất cả repos dùng `DBTX` (đã có ở `internal/repository/db.go`)
- **Error pattern:** `service.ErrNotFound` wrapping `sql.ErrNoRows` (ở `internal/service/errors.go`)
- **Response format:** `{"data": ...}` cho success, `{"error": "message"}` cho errors (top-level, KHÔNG double-wrap)
- **Context:** `context.Context` là param đầu tiên trong service/repo methods
- **Receiver names:** 1-2 chữ (`func (s *AuthService)`, `func (r *UserRepo)`)
- **Error wrapping:** `fmt.Errorf("doing X: %w", err)`

---

## Fix 1 (🔴): Redesign `user_data` schema theo database-design.md

Đây là fix lớn nhất. Migration 009 hiện tại **sai hoàn toàn** so với design doc.

### Bước 1: Rewrite migration 009

Thay toàn bộ `up009`/`down009` trong `internal/database/migrate/registry.go`:

```sql
-- Drop old progress table (from migration 001)
DROP TABLE IF EXISTS progress;

-- Unified per-user-per-media state (Emby pattern: 1 row = 1 user-media pair)
CREATE TABLE user_data (
    user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_id       INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    position       REAL DEFAULT 0,
    completed      INTEGER DEFAULT 0,
    is_favorite    INTEGER DEFAULT 0,
    rating         REAL DEFAULT NULL CHECK (rating IS NULL OR (rating >= 1.0 AND rating <= 10.0)),
    play_count     INTEGER DEFAULT 0,
    last_played_at DATETIME DEFAULT NULL,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, media_id)
);
CREATE INDEX idx_ud_user ON user_data(user_id);
CREATE INDEX idx_ud_media ON user_data(media_id);
CREATE INDEX idx_ud_favorite ON user_data(user_id) WHERE is_favorite = 1;
CREATE INDEX idx_ud_recent ON user_data(user_id, last_played_at DESC) WHERE last_played_at IS NOT NULL;

-- Series-level favorite/rating
CREATE TABLE user_series_data (
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    series_id   INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    is_favorite INTEGER DEFAULT 0,
    rating      REAL DEFAULT NULL CHECK (rating IS NULL OR (rating >= 1.0 AND rating <= 10.0)),
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, series_id)
);
```

Down migration: `DROP TABLE IF EXISTS user_series_data; DROP TABLE IF EXISTS user_data;`

### Bước 2: Rewrite model `UserData`

File: `internal/model/user.go` — thay `UserData`, `FavoriteItem`, `RecentlyWatchedItem` structs:

```go
// UserData represents unified per-user-per-media state (Emby pattern)
type UserData struct {
    UserID       int64    `json:"user_id"`
    MediaID      int64    `json:"media_id"`
    Position     float64  `json:"position"`
    Completed    bool     `json:"completed"`
    IsFavorite   bool     `json:"is_favorite"`
    Rating       *float64 `json:"rating"`          // nil = not rated, 1.0-10.0
    PlayCount    int      `json:"play_count"`
    LastPlayedAt *string  `json:"last_played_at"`   // nil = never played
    UpdatedAt    string   `json:"updated_at"`
}

// UserSeriesData represents series-level favorite/rating
type UserSeriesData struct {
    UserID     int64    `json:"user_id"`
    SeriesID   int64    `json:"series_id"`
    IsFavorite bool     `json:"is_favorite"`
    Rating     *float64 `json:"rating"`
    UpdatedAt  string   `json:"updated_at"`
}
```

Xóa `UserDataType`, `UserDataProgress`, `UserDataFavorite`, `UserDataRating` constants.
Xóa `FavoriteItem`, `RecentlyWatchedItem` structs (hoặc refactor nếu cần cho JOIN queries).

### Bước 3: Rewrite `UserDataRepo`

File: `internal/repository/user_data.go` — methods mới:

```go
// GetProgress returns user data for a media item
func (r *UserDataRepo) GetProgress(ctx, userID, mediaID int64) (*model.UserData, error)

// UpsertProgress creates or updates watch progress (UPSERT on PK user_id+media_id)
// Cũng update last_played_at = CURRENT_TIMESTAMP và play_count = play_count + 1 nếu completed
func (r *UserDataRepo) UpsertProgress(ctx, userID, mediaID int64, position float64, completed bool) error

// ToggleFavorite flips is_favorite (UPSERT: INSERT nếu chưa có, UPDATE nếu có)
func (r *UserDataRepo) ToggleFavorite(ctx, userID, mediaID int64) (isFavorite bool, err error)

// SetRating sets user rating (nil = remove rating). UPSERT.
func (r *UserDataRepo) SetRating(ctx, userID, mediaID int64, rating *float64) error

// ListFavorites returns items where is_favorite = 1, JOIN media for title/poster
func (r *UserDataRepo) ListFavorites(ctx, userID int64, limit, offset int) ([]*model.UserData, error)

// ListRecentlyWatched returns items ordered by last_played_at DESC, JOIN media
func (r *UserDataRepo) ListRecentlyWatched(ctx, userID int64, limit int) ([]*model.UserData, error)
```

UPSERT pattern cho SQLite:
```sql
INSERT INTO user_data (user_id, media_id, position, completed, last_played_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, media_id) DO UPDATE SET
    position = excluded.position,
    completed = excluded.completed,
    last_played_at = CURRENT_TIMESTAMP,
    play_count = CASE WHEN excluded.completed = 1 AND user_data.completed = 0
                      THEN user_data.play_count + 1
                      ELSE user_data.play_count END,
    updated_at = CURRENT_TIMESTAMP
```

ToggleFavorite UPSERT:
```sql
INSERT INTO user_data (user_id, media_id, is_favorite)
VALUES (?, ?, 1)
ON CONFLICT(user_id, media_id) DO UPDATE SET
    is_favorite = CASE WHEN user_data.is_favorite = 1 THEN 0 ELSE 1 END,
    updated_at = CURRENT_TIMESTAMP
RETURNING is_favorite
```

### Bước 4: Update `UserDataService` và `ProfileHandler`

Adapt to new repo method signatures. Remove all `data_type` references.

---

## Fix 2 (🔴): Implement `SetLibraryAccess` handler

File: `internal/handler/user.go:157-175`

Cần `*sql.DB` để tạo transaction. Cách đơn giản nhất: thêm `db *sql.DB` vào `UserHandler` hoặc thêm method `SetLibraryAccess` vào `AuthService` (service handle transaction).

**Khuyến nghị:** Thêm vào `AuthService`:

```go
// In service/auth.go
func (s *AuthService) SetLibraryAccess(ctx context.Context, userID int64, libraryIDs []int64) error {
    // Verify user exists
    _, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) { return ErrNotFound }
        return fmt.Errorf("fetching user: %w", err)
    }

    // Need *sql.DB for transaction - add db field to AuthService
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil { return fmt.Errorf("begin tx: %w", err) }
    defer tx.Rollback()

    if err := s.userRepo.WithTx(tx).SetLibraryAccess(ctx, userID, libraryIDs); err != nil {
        return fmt.Errorf("setting library access: %w", err)
    }
    return tx.Commit()
}
```

Thêm `db *sql.DB` vào `AuthService` struct và constructor. Update handler để gọi service method.

---

## Fix 3 (🔴): `GET /api/profile` handler

File: `internal/handler/profile.go` — thêm method:

```go
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID, _, ok := auth.UserFromContext(r.Context())
    if !ok {
        respondError(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    user, err := h.authSvc.GetUser(r.Context(), userID)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            respondError(w, http.StatusNotFound, "user not found")
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    user.PasswordHash = ""
    respondJSON(w, http.StatusOK, user)
}
```

File: `cmd/server/main.go:163` — đổi:
```go
mux.HandleFunc("GET /api/profile", profileHandler.GetProfile)
```

---

## Fix 4 (🔴): `SessionTracker` data race

File: `internal/middleware/auth.go:85-104`

Thêm `sync.Mutex`:

```go
func SessionTracker(sessionUpdateFunc func(userID int64)) func(http.Handler) http.Handler {
    var mu sync.Mutex
    lastUpdates := make(map[int64]time.Time)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID, _, ok := auth.UserFromContext(r.Context())
            if ok {
                now := time.Now()
                mu.Lock()
                lastUpdate, exists := lastUpdates[userID]
                shouldUpdate := !exists || now.Sub(lastUpdate) > time.Minute
                if shouldUpdate {
                    lastUpdates[userID] = now
                }
                mu.Unlock()

                if shouldUpdate && sessionUpdateFunc != nil {
                    go sessionUpdateFunc(userID)
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

Thêm `"sync"` vào imports.

---

## Fix 5 (🔴): `Logout` xóa session

File: `internal/service/auth.go:184-202`

```go
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
    tokenHash := auth.HashToken(refreshToken)

    rt, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil // Already logged out
        }
        return fmt.Errorf("fetching refresh token: %w", err)
    }

    // Delete associated session (by refresh_token_id)
    if err := s.sessionRepo.DeleteByRefreshTokenID(ctx, rt.ID); err != nil {
        return fmt.Errorf("deleting session: %w", err)
    }

    // Delete the refresh token
    if err := s.refreshTokenRepo.Delete(ctx, rt.ID); err != nil {
        return fmt.Errorf("deleting refresh token: %w", err)
    }

    return nil
}
```

Cần thêm method `DeleteByRefreshTokenID` vào `SessionRepo`:
```go
func (r *SessionRepo) DeleteByRefreshTokenID(ctx context.Context, rtID int64) error {
    _, err := r.db.ExecContext(ctx,
        "DELETE FROM sessions WHERE refresh_token_id = ?", rtID)
    return err
}
```

---

## Fix 6 (🔴): Middleware stack order

File: `cmd/server/main.go:208-214`

Đổi thành (sessionTracker phải ở giữa auth và mux):
```go
var h http.Handler = mux
h = sessionTracker(h)      // innermost: track session AFTER auth sets context
h = authMiddleware(h)       // auth: set user context
h = middleware.CORS(cfg.CORSOrigin)(h)
h = middleware.Logger(h)
h = middleware.Recovery(h)  // outermost
```

Request flow: Recovery → Logger → CORS → authMiddleware → sessionTracker → mux ✅

---

## Fix 7 (🟡): Admin routes thiếu `RequireAdmin`

File: `cmd/server/main.go:155-160`

Dùng Go 1.22+ cách wrap individual routes hoặc tạo admin sub-mux:

```go
// User management routes (admin only)
mux.Handle("GET /api/users", middleware.RequireAdmin(http.HandlerFunc(userHandler.List)))
mux.Handle("POST /api/users", middleware.RequireAdmin(http.HandlerFunc(userHandler.Create)))
mux.Handle("PATCH /api/users/{id}", middleware.RequireAdmin(http.HandlerFunc(userHandler.Update)))
mux.Handle("DELETE /api/users/{id}", middleware.RequireAdmin(http.HandlerFunc(userHandler.Delete)))
mux.Handle("PUT /api/users/{id}/library-access", middleware.RequireAdmin(http.HandlerFunc(userHandler.SetLibraryAccess)))
```

Cũng bảo vệ admin library routes:
```go
mux.Handle("POST /api/libraries", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.Create)))
mux.Handle("DELETE /api/libraries/{id}", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.Delete)))
mux.Handle("POST /api/libraries/{id}/scan", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.Scan)))
```

---

## Fix 8 (🟡): `GrantAllLibraries` error nuốt

File: `internal/service/auth.go:110-113`

Đổi:
```go
if err := s.userRepo.GrantAllLibraries(ctx, user.ID); err != nil {
    log.Printf("warning: granting library access for user %d: %v", user.ID, err)
}
```

Thêm `"log"` vào imports nếu chưa có.

---

## Fix 9 (🟡): `respondError` double-wrap

File: `internal/handler/respond.go`

Đổi `respondError` để write `{"error": "message"}` trực tiếp (KHÔNG qua `respondJSON`):
```go
func respondError(w http.ResponseWriter, status int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
        log.Printf("json encode error: %v", err)
    }
}
```

Cũng update `middleware/auth.go` cho consistent:
```go
func respondUnauthorized(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    w.Write([]byte(`{"error":"unauthorized"}`))
}

func respondForbidden(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusForbidden)
    w.Write([]byte(`{"error":"forbidden"}`))
}
```

---

## Fix 10 (🟡): Username normalize trước validate

File: `internal/service/auth.go:68-81`

Đổi thứ tự:
```go
func (s *AuthService) CreateUser(ctx context.Context, username, password, displayName string, isAdmin bool) (*model.User, error) {
    // Normalize FIRST
    username = strings.ToLower(strings.TrimSpace(username))

    // Validate AFTER normalize
    if !isValidUsername(username) {
        return nil, ErrInvalidUsername
    }
    if len(password) < 8 {
        return nil, ErrInvalidPassword
    }
    // ... rest unchanged
```

Xóa dòng normalize cũ (line 80) vì đã normalize ở đầu.

---

## Fix 11 (🟢): Move `Session`/`RefreshToken` structs to model

File: `internal/repository/session.go` — move `RefreshToken` (lines 62-71) và `Session` (lines 178-189) structs sang `internal/model/user.go` (hoặc tạo `internal/model/session.go`). Update imports trong repository và service.

---

## Fix 12 (🟢): Compile regex một lần

File: `internal/service/auth.go`

```go
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidUsername(username string) bool {
    if len(username) < 3 || len(username) > 32 {
        return false
    }
    return usernameRegex.MatchString(username)
}
```

---

## Fix 13 (🟢): Xóa thừa password hash clear trong `UserHandler.List`

File: `internal/handler/user.go:29-31`

Xóa loop `for _, u := range users { u.PasswordHash = "" }` — `json:"-"` trên model đã handle.

---

## Fix 14 (🟢): Thêm cleanup goroutine cho expired tokens/sessions

File: `cmd/server/main.go` — thêm trong `runServer()` sau khởi tạo repos:

```go
// Cleanup expired tokens/sessions every hour
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        if err := refreshTokenRepo.DeleteExpired(ctx); err != nil {
            log.Printf("cleanup expired tokens: %v", err)
        }
        if err := sessionRepo.DeleteExpired(ctx); err != nil {
            log.Printf("cleanup expired sessions: %v", err)
        }
        cancel()
    }
}()
```

---

## Fix 15 (🟢): `LoadOrCreateSecret` validate/trim

File: `internal/auth/jwt.go:44-68`

```go
func LoadOrCreateSecret(dataDir string) ([]byte, error) {
    secretPath := filepath.Join(dataDir, ".jwt_secret")

    if data, err := os.ReadFile(secretPath); err == nil {
        // Use only first 32 bytes, ignore trailing whitespace/newlines
        if len(data) >= 32 {
            return data[:32], nil
        }
        // File exists but too short — regenerate
    }
    // ... generate and save (unchanged)
```

---

## Fix 16 (🟢): Tests — defer to later

Ghi chú: Tests cho middleware, handler, profile sẽ được viết trong phase review riêng. Không cần fix trong lần này.

---

## Thứ tự fix khuyến nghị

1. Fix 9 (`respondError`) — nhỏ, ảnh hưởng tất cả handlers
2. Fix 1 (`user_data` schema) — lớn nhất, cần làm sớm
3. Fix 6 (middleware order) + Fix 4 (data race) — liên quan nhau
4. Fix 5 (logout session) + Fix 2 (SetLibraryAccess) — service fixes
5. Fix 3 (GET profile) + Fix 7 (RequireAdmin) — handler/route fixes
6. Fix 8, 10, 11, 12, 13, 14, 15 — nhỏ, làm nhanh

## Verify

Sau khi fix xong tất cả:
```sh
cd backend && make test && make lint
```

Đảm bảo tất cả tests pass và không có lint errors.
