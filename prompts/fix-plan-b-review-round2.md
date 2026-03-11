# Fix Plan B Review — Round 2

Đây là 5 issues còn lại từ re-review Plan B Auth & Sessions. N6 đã accepted (không cần fix).

## Checklist

Đọc review tại `plans/plan-b-auth-sessions/review.md` (section "Round 2"). Fix theo thứ tự dưới đây. Sau mỗi fix, đánh dấu `[x]` trong review.md.

---

## N1: Session/RefreshToken model thiếu JSON tags

**File:** `internal/model/session.go`

Thêm json tags cho cả 2 structs. `TokenHash` phải có `json:"-"`.

```go
// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TokenHash  string    `json:"-"`
	DeviceName string    `json:"device_name"`
	IPAddress  string    `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// Session represents a session in the database
type Session struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	RefreshTokenID *int64    `json:"refresh_token_id,omitempty"`
	DeviceName     string    `json:"device_name"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastActiveAt   time.Time `json:"last_active_at"`
	CreatedAt      time.Time `json:"created_at"`
}
```

---

## N2: Refresh() không xóa old session → orphaned sessions

**File:** `internal/service/auth.go` — method `Refresh()`

Hiện tại (line ~170):
```go
// Delete old token (rotation)
if err := s.refreshTokenRepo.Delete(ctx, rt.ID); err != nil {
    return nil, fmt.Errorf("deleting old token: %w", err)
}
```

Thêm delete old session TRƯỚC khi delete old token:
```go
// Delete old session and token (rotation)
if err := s.sessionRepo.DeleteByRefreshTokenID(ctx, rt.ID); err != nil {
    return nil, fmt.Errorf("deleting old session: %w", err)
}
if err := s.refreshTokenRepo.Delete(ctx, rt.ID); err != nil {
    return nil, fmt.Errorf("deleting old token: %w", err)
}
```

**Lý do xóa session trước:** Nếu delete token trước, ON DELETE SET NULL sẽ set session.refresh_token_id = NULL → không thể tìm session bằng refresh_token_id nữa.

---

## N3: Xóa `PasswordHash = ""` thừa ở handlers

Model `User` đã có `json:"-"` trên PasswordHash field — JSON encoder tự động skip field này. Gán `""` là thừa.

**Xóa các dòng sau:**

1. `internal/handler/user.go:67` — trong method `Create`:
   ```
   user.PasswordHash = ""    ← XÓA dòng này
   ```

2. `internal/handler/user.go:119` — trong method `Update`:
   ```
   user.PasswordHash = ""    ← XÓA dòng này
   ```

3. `internal/handler/profile.go:106` — trong method `UpdateProfile`:
   ```
   user.PasswordHash = ""    ← XÓA dòng này
   ```

4. `internal/handler/profile.go:128` — trong method `GetProfile`:
   ```
   user.PasswordHash = ""    ← XÓA dòng này
   ```

---

## N4: `DeleteUser` "cannot delete self" — thêm sentinel error

### Step 1: Thêm sentinel error

**File:** `internal/service/auth.go` — thêm vào block `var (...)` ở đầu file (cùng chỗ với `ErrInvalidCredentials`, `ErrUserExists`, etc.):

```go
ErrDeleteSelf = errors.New("cannot delete your own account")
```

### Step 2: Service dùng sentinel

**File:** `internal/service/auth.go` — method `DeleteUser`, đổi:
```go
// Current (line ~362):
return errors.New("cannot delete your own account")

// Đổi thành:
return ErrDeleteSelf
```

### Step 3: Handler thêm case

**File:** `internal/handler/user.go` — method `Delete`, thêm case trong switch:
```go
if err := h.authSvc.DeleteUser(r.Context(), id, currentUserID); err != nil {
    switch {
    case errors.Is(err, service.ErrNotFound):
        respondError(w, http.StatusNotFound, "user not found")
    case errors.Is(err, service.ErrLastAdmin):
        respondError(w, http.StatusBadRequest, "cannot remove the last admin")
    case errors.Is(err, service.ErrDeleteSelf):                              // ← THÊM
        respondError(w, http.StatusBadRequest, "cannot delete your own account") // ← THÊM
    default:
        respondError(w, http.StatusInternalServerError, err.Error())
    }
    return
}
```

---

## N5: Test rollback count sai

**File:** `internal/database/migrate/migrate_test.go` — `TestRealMigrations_Rollback`

Hiện có 9 migrations. Để rollback đến trước migration 004, cần rollback 6 lần (009, 008, 007, 006, 005, 004).

```go
// Current:
// Rollback migrations 007, 006, 005, and 004 to test rollback of 004
for i := 0; i < 4; i++ {

// Đổi thành:
// Rollback migrations 009, 008, 007, 006, 005, and 004 to test rollback of 004
for i := 0; i < 6; i++ {
```

---

## Verify

Sau khi fix xong, chạy:

```sh
cd backend
go build ./...
go test ./... -v -count=1
```

Đảm bảo:
- Build pass (zero errors)
- All tests pass (đặc biệt `TestRealMigrations_Rollback`)
- `go vet ./...` clean
