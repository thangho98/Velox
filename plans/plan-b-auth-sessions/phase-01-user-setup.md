# Phase 01: User Model & First-Run Setup
Status: ⬜ Pending
Plan: B - Auth & Sessions

## Tasks

### 1. User Model
- [ ] Table `users`: id, username, display_name, password_hash, is_admin, avatar_path, created_at, updated_at
- [ ] Struct `User` với `CheckPassword(plain) bool`
- [ ] Migration: `007_users.go`
- **File:** `internal/model/user.go` - NEW

### 2. Password Hashing
- [ ] `golang.org/x/crypto/bcrypt` - cost 12
- [ ] `internal/auth/password.go` - HashPassword, CheckPassword
- [ ] Never store plaintext, never log passwords

### 3. User Repository
- [ ] `Create`, `GetByID`, `GetByUsername`, `List`, `Update`, `Delete`, `Count`
- [ ] `Count()` dùng cho detect first-run (0 users = unconfigured)
- **File:** `internal/repository/user.go` - NEW

### 4. First-Run Setup (NO default credentials)
- [ ] `GET /api/setup/status` → `{configured: bool, server_name: string}`
- [ ] Nếu `Count() == 0` → server chưa configured
- [ ] `POST /api/setup` → body: `{username, password, display_name, server_name}`
- [ ] Validate: password min 8 chars, username alphanumeric
- [ ] Tạo admin user, save server_name to config
- [ ] **After setup: endpoint trả 403** (one-time only)
- [ ] Frontend redirect tới setup page khi unconfigured
- **File:** `internal/handler/setup.go` - NEW

### 5. Auth Service
- [ ] `Login(username, password) (*User, error)`
- [ ] `ChangePassword(userID, oldPass, newPass) error`
- [ ] Validate old password trước khi cho đổi
- **File:** `internal/service/auth.go` - NEW

### 6. Auth Handlers
- [ ] `POST /api/auth/login` → validate credentials, return tokens
- [ ] `POST /api/auth/change-password` → require current password
- [ ] `GET /api/auth/me` → current user info from token
- [ ] `POST /api/auth/logout` → revoke refresh token
- **File:** `internal/handler/auth.go` - NEW

## Files to Create/Modify
- `internal/database/migrate/migrations/007_users.go` - NEW
- `internal/model/user.go` - NEW
- `internal/auth/password.go` - NEW
- `internal/repository/user.go` - NEW
- `internal/service/auth.go` - NEW
- `internal/handler/auth.go` - NEW
- `internal/handler/setup.go` - NEW

---
Next: phase-02-jwt-middleware.md
