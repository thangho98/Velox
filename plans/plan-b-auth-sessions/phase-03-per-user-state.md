# Phase 03: Per-User State & Library ACL
Status: ⬜ Pending
Plan: B - Auth & Sessions
Dependencies: Phase 02

## Tasks

### 1. Unified User Data Table (replaces progress + favorites + user_ratings)
- [ ] DROP old `progress` table, CREATE `user_data` (PK: user_id + media_id)
- [ ] Fields: position, completed, is_favorite, rating, play_count, last_played_at, updated_at
- [ ] CREATE `user_series_data` (PK: user_id + series_id) for series favorite/rating
- [ ] Migration: `009_per_user_state.go`
- [ ] Refactor ProgressRepo → UserDataRepo: all queries use unified table
- [ ] Update ProgressHandler: extract user_id from context
- [ ] Each user has independent watch progress + favorites + ratings in 1 row per media

### 2. Library Access Control
- [ ] Table `user_library_access`: user_id, library_id (whitelist)
- [ ] Admin sees all libraries by default
- [ ] New users get access to all existing libraries (configurable)
- [ ] Filter media queries by user's accessible libraries
- [ ] `PUT /api/admin/users/{id}/libraries` → set library access `{library_ids: [1,2,3]}`

### 3. User Management API (Admin)
- [ ] `GET /api/admin/users` - list all users
- [ ] `POST /api/admin/users` - create user `{username, password, display_name, is_admin}`
- [ ] `PUT /api/admin/users/{id}` - update user (display_name, is_admin)
- [ ] `DELETE /api/admin/users/{id}` - delete user + cascade (progress, sessions, favorites)
- [ ] Cannot delete self, cannot remove last admin
- **File:** `internal/handler/user.go` - NEW

### 4. User Preferences
- [ ] Table `user_preferences`: user_id (PK), subtitle_language, audio_language, max_streaming_quality, theme (light|dark|auto)
- [ ] Default prefs created when user created
- [ ] `GET /api/me/preferences`
- [ ] `PUT /api/me/preferences`
- [ ] Migration: `009_per_user_state.go` (combine)

### 5. Watch State Per Episode
- [ ] Progress table already handles per-media-item progress
- [ ] Ensure episode progress tracked individually (media_id = episode's media_id)
- [ ] Helper: `GetSeriesProgress(userID, seriesID)` → map[seasonNum]map[epNum]Progress
- [ ] Used by frontend to show watched/unwatched badges

### 6. Profile Update
- [ ] `PUT /api/me` - update display_name
- [ ] `PUT /api/me/avatar` - upload avatar (multipart, max 2MB, resize 200x200)
- [ ] Serve: `GET /api/avatars/{user_id}` - avatar image (default placeholder if none)
- [ ] Store: `data/avatars/{user_id}.jpg`

## Files to Create/Modify
- `internal/database/migrate/migrations/009_per_user_state.go` - NEW
- `internal/repository/user_data.go` - NEW (replaces progress.go)
- `internal/handler/progress.go` - Refactor (use UserDataRepo)
- `internal/handler/user.go` - NEW
- `internal/repository/user.go` - Add library access queries
- `internal/service/progress.go` - Update

---
✅ End of Plan B
Next Plan: plan-c-webapp-mvp
