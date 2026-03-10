# Phase 01: App Shell & Auth UI
Status: ⬜ Pending
Plan: C - Web App MVP

## Tasks

### 1. API Client
- [ ] `src/lib/api.ts` - fetch wrapper with base URL, auth headers, error handling
- [ ] Auto-attach JWT token from localStorage
- [ ] Auto-refresh token on 401 (intercept → refresh → retry)
- [ ] Type-safe response types matching backend models

### 2. Auth Context & State
- [ ] `src/context/AuthContext.tsx` - React context for auth state
- [ ] `useAuth()` hook: user, login(), logout(), isAuthenticated, isAdmin
- [ ] Persist tokens in localStorage
- [ ] Auto-redirect to login if token expired

### 3. Setup Page
- [ ] Route: `/setup`
- [ ] Show only when `GET /api/setup/status` returns `{configured: false}`
- [ ] Form: username, password, confirm password, display name, server name
- [ ] Validation: password min 8 chars, passwords match
- [ ] On success: auto-login → redirect to home

### 4. Login Page
- [ ] Route: `/login`
- [ ] Form: username, password
- [ ] Error display: "Invalid credentials"
- [ ] Redirect to `/` on success
- [ ] Redirect to `/setup` if server unconfigured

### 5. App Layout
- [ ] `src/components/Layout.tsx` - main layout shell
- [ ] Sidebar/navbar: Home, Movies, Series, Search
- [ ] User menu (top-right): profile name, avatar, settings, logout
- [ ] Responsive: sidebar collapses on mobile
- [ ] Dark theme by default (media server aesthetic)

### 6. Protected Routes
- [ ] Route guard: redirect to `/login` if not authenticated
- [ ] Admin route guard for admin-only pages
- [ ] Loading state while checking auth

### 7. Router Setup
- [ ] React Router v7 with routes:
  - `/setup` - first-run setup
  - `/login` - login
  - `/` - home (protected)
  - `/movies` - movie library
  - `/series` - series library
  - `/movie/:id` - movie detail
  - `/series/:id` - series detail
  - `/series/:id/season/:num` - season detail
  - `/play/:id` - video player
  - `/search` - search results
  - `/settings` - user settings
  - `/admin/*` - admin pages

## Files to Create
- `src/lib/api.ts`
- `src/context/AuthContext.tsx`
- `src/pages/SetupPage.tsx`
- `src/pages/LoginPage.tsx`
- `src/components/Layout.tsx`
- `src/components/ProtectedRoute.tsx`
- `src/App.tsx` - router setup

---
Next: phase-02-browse-detail.md
