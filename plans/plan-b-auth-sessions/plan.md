# Plan B: Auth & Sessions
Status: ⬜ Pending
Priority: 🔴 Critical
Dependencies: Plan A (cần user_id cho progress, watch history)

## Mục tiêu
- First-run setup (NO default credentials)
- JWT auth + refresh tokens
- Per-user progress, watch state, library access
- Session tracking

## Security Principles
- KHÔNG seed default admin/admin. First-run setup wizard forces user tạo admin account.
- Passwords: bcrypt, cost 12
- JWT: short-lived access (15min) + long-lived refresh (7 days)
- Refresh tokens stored in DB, revokable

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | User Model & First-Run Setup | 6 tasks | ⬜ |
| 02 | JWT Auth & Middleware | 6 tasks | ⬜ |
| 03 | Per-User State & Library ACL | 6 tasks | ⬜ |
