# Phase 04: Home Screen & Polish
Status: ⬜ Pending
Plan: C - Web App MVP
Dependencies: Phase 02, 03

## Tasks

### 1. Home Page
- [ ] Route: `/`
- [ ] Hero banner: random featured item with backdrop, title, overview, play button
- [ ] "Continue Watching" row (nếu có progress items)
- [ ] "Next Up" row (next episodes cho series đang xem)
- [ ] "Recently Added" row
- [ ] Genre rows (top 3-5 genres)
- [ ] Backend: `GET /api/home` aggregate endpoint (hoặc multiple calls)
- **File:** `src/pages/HomePage.tsx`

### 2. Horizontal Scroll Row Component
- [ ] `src/components/MediaRow.tsx` - reusable horizontal scroll
- [ ] Section title + "See All" link
- [ ] Scroll buttons (left/right arrows)
- [ ] Snap scroll behavior
- [ ] Responsive: fewer items on smaller screens

### 3. Settings Page
- [ ] Route: `/settings`
- [ ] Change display name
- [ ] Change password
- [ ] Subtitle language preference
- [ ] Audio language preference
- [ ] Theme toggle (dark/light) - dark default
- **File:** `src/pages/SettingsPage.tsx`

### 4. Admin: Library Management
- [ ] Route: `/admin/libraries`
- [ ] List libraries: name, path, item count, last scanned
- [ ] Add library: name + path input + library type (movies/tvshows/mixed)
- [ ] Delete library (confirm dialog)
- [ ] Scan button per library (show progress/status)
- **File:** `src/pages/admin/LibrariesPage.tsx`

### 5. Admin: User Management
- [ ] Route: `/admin/users`
- [ ] List users: username, display name, role, last active
- [ ] Create user form
- [ ] Edit user: toggle admin, set library access
- [ ] Delete user (confirm dialog, cannot delete self)
- **File:** `src/pages/admin/UsersPage.tsx`

### 6. Loading & Error States
- [ ] Skeleton loaders for MediaCard grid
- [ ] Skeleton for detail page (poster + text placeholders)
- [ ] Error boundary: generic error page
- [ ] 404 page
- [ ] Toast notifications (success/error for actions)
- [ ] Empty states with helpful messages

## Files to Create
- `src/pages/HomePage.tsx`
- `src/pages/SettingsPage.tsx`
- `src/pages/admin/LibrariesPage.tsx`
- `src/pages/admin/UsersPage.tsx`
- `src/components/MediaRow.tsx`
- `src/components/Skeleton.tsx`
- `src/components/Toast.tsx`
- `src/pages/NotFoundPage.tsx`

---
✅ End of Plan C
🎯 MILESTONE 1: Usable MVP achieved!
Next Plan: plan-d-playback-engine
