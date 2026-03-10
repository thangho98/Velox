# Phase 02: JWT Auth & Middleware
Status: ⬜ Pending
Plan: B - Auth & Sessions
Dependencies: Phase 01

## Tasks

### 1. JWT Token Package
- [ ] `internal/auth/jwt.go` - NEW
- [ ] `GenerateAccessToken(userID, isAdmin) (string, error)` - 15 min expiry
- [ ] `GenerateRefreshToken(userID) (string, error)` - 7 days expiry
- [ ] `ValidateToken(tokenString) (*Claims, error)`
- [ ] JWT secret: env `VELOX_JWT_SECRET`, auto-generate random 32 bytes on first run → persist to `data/.jwt_secret`
- [ ] `github.com/golang-jwt/jwt/v5`

### 2. Refresh Token Storage
- [ ] Table `refresh_tokens`: id, user_id, token_hash, device_name, ip_address, expires_at, created_at
- [ ] Hash refresh token before storing (SHA256)
- [ ] Migration: `008_refresh_tokens.go`
- [ ] `POST /api/auth/refresh` → validate refresh token → issue new access + refresh
- [ ] Rotate refresh token on each refresh (old one invalidated)

### 3. Auth Middleware
- [ ] `middleware.RequireAuth` - extract + validate JWT from `Authorization: Bearer <token>`
- [ ] Set `UserID`, `IsAdmin` in `context.Context`
- [ ] Helper: `auth.UserFromContext(ctx) (userID int64, isAdmin bool)`
- [ ] Return 401 with `{error: "unauthorized"}` if invalid/expired
- **File:** `internal/middleware/auth.go` - NEW

### 4. Admin Middleware
- [ ] `middleware.RequireAdmin` - check IsAdmin from context
- [ ] Return 403 with `{error: "forbidden"}` if not admin
- [ ] Stack: `RequireAuth → RequireAdmin`

### 5. Route Protection
- [ ] **Public:** `/api/setup/*`, `/api/auth/login`, `/api/auth/refresh`
- [ ] **Authenticated:** all other `/api/*`
- [ ] **Admin only:** `POST/DELETE /api/libraries/*`, `/api/users/*`, `/api/admin/*`
- [ ] **Stream auth:** `/api/stream/*` accept `?token=` query param (video player can't set headers)
- **File:** `cmd/server/main.go` - reorganize route groups

### 6. Session Tracking
- [ ] Table `sessions`: id, user_id, device_name, ip_address, user_agent, last_active_at, created_at
- [ ] Update `last_active_at` on each authenticated request (debounced, max 1 update/minute)
- [ ] `GET /api/me/sessions` - list my active sessions
- [ ] `DELETE /api/me/sessions/{id}` - revoke session (delete refresh token)
- [ ] Migration: `008_refresh_tokens.go` (combine)

## Files to Create/Modify
- `internal/auth/jwt.go` - NEW
- `internal/middleware/auth.go` - NEW
- `internal/database/migrate/migrations/008_refresh_tokens.go` - NEW
- `internal/repository/session.go` - NEW
- `cmd/server/main.go` - Route protection

---
Next: phase-03-per-user-state.md
