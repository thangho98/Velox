# Phase 02: Extract English Strings

Status: ⬜ Pending
Dependencies: Phase 01

---

## Objective

Extract tất cả UI strings từ components sang English translation files.

---

## Requirements

### Functional
- [ ] Extract common strings (actions, states, app info)
- [ ] Extract auth strings (login, setup, errors)
- [ ] Extract navigation strings (navbar, menus)
- [ ] Extract settings strings (all sections)
- [ ] Extract media strings (browse, detail, search)
- [ ] Extract watch strings (player controls)
- [ ] Extract error strings

### Non-Functional
- [ ] Organize by namespace
- [ ] Nested key structure
- [ ] Consistent naming convention

---

## Files to Update (Priority Order)

### Priority 1: Core UI
1. `Navbar.tsx` - Navigation labels, user menu
2. `LoginPage.tsx` - Form labels, errors, buttons
3. `SetupPage.tsx` - Setup flow text
4. `Sidebar.tsx` - Menu items

### Priority 2: Settings (Largest)
5. `SettingsPage.tsx` - All sections (profile, preferences, security, metadata, subtitles, etc.)

### Priority 3: Media Pages
6. `HomePage.tsx` - Section titles
7. `BrowsePage.tsx` - Filters, sorts
8. `MoviesPage.tsx` - Page content
9. `SeriesPage.tsx` - Page content
10. `SearchPage.tsx` - Search UI
11. `MediaDetailPage.tsx` - Action buttons, tabs
12. `SeriesDetailPage.tsx` - Seasons, episodes

### Priority 4: Watch
13. `WatchPage.tsx` - Player controls, overlays

### Priority 5: Components
14. `ActionMenu.tsx` - Menu items
15. `MediaCard.tsx` - Labels
16. `EpisodeCard.tsx` - Labels
17. `FilterBar.tsx` - Filter labels
18. Modals, Toasts, Dialogs

---

## Translation Structure

### Common (`locales/en/common.json`)
```json
{
  "app": {
    "name": "Velox",
    "tagline": "Your personal media server"
  },
  "actions": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "edit": "Edit",
    "create": "Create",
    "close": "Close",
    "confirm": "Confirm",
    "loading": "Loading...",
    "saving": "Saving...",
    "scanning": "Scanning...",
    "searching": "Searching...",
    "signIn": "Sign In",
    "signOut": "Sign Out"
  },
  "states": {
    "empty": "No items found",
    "error": "Something went wrong",
    "success": "Success!",
    "active": "Active",
    "inactive": "Inactive",
    "enabled": "Enabled",
    "disabled": "Disabled"
  }
}
```

### Auth (`locales/en/auth.json`)
```json
{
  "login": {
    "title": "Sign In",
    "username": "Username",
    "password": "Password",
    "signingIn": "Signing in...",
    "newUser": "New to Velox?",
    "contactAdmin": "Contact your administrator"
  },
  "setup": {
    "title": "Create Admin Account",
    "description": "Set up your Velox server"
  },
  "errors": {
    "required": "Username and password are required",
    "invalid": "Invalid credentials"
  }
}
```

### Navigation (`locales/en/navigation.json`)
```json
{
  "nav": {
    "home": "Home",
    "movies": "Movies",
    "series": "Series",
    "browse": "Browse",
    "search": "Search"
  },
  "userMenu": {
    "settings": "Settings",
    "signOut": "Sign Out",
    "admin": "Administrator",
    "user": "User"
  },
  "search": {
    "placeholder": "Search movies, genres..."
  }
}
```

### Settings (`locales/en/settings.json`)
```json
{
  "sections": {
    "profile": { "title": "Profile", "description": "Manage your account information" },
    "preferences": { "title": "Preferences", "description": "Customize your viewing experience" },
    "security": { "title": "Security", "description": "Change your password" },
    "sessions": { "title": "Sessions", "description": "Manage your active sessions" },
    "metadata": { "title": "Metadata", "description": "Configure metadata providers" },
    "subtitles": { "title": "Subtitles", "description": "Configure external subtitle providers" },
    "playback": { "title": "Playback", "description": "Server-wide playback policy" },
    "cinema": { "title": "Cinema Mode", "description": "Play trailers before main feature" },
    "general": { "title": "Dashboard", "description": "Server information and status" },
    "libraries": { "title": "Libraries", "description": "Manage media libraries" },
    "users": { "title": "Users", "description": "Manage user accounts" },
    "activity": { "title": "Activity", "description": "Recent server activity" },
    "tasks": { "title": "Tasks", "description": "Background tasks and maintenance" },
    "webhooks": { "title": "Webhooks", "description": "Configure webhook notifications" }
  },
  "fields": {
    "username": "Username",
    "displayName": "Display Name",
    "password": "Password",
    "confirmPassword": "Confirm Password",
    "currentPassword": "Current Password",
    "newPassword": "New Password",
    "role": "Role",
    "subtitleLanguage": "Subtitle Language",
    "audioLanguage": "Audio Language",
    "maxQuality": "Max Streaming Quality",
    "theme": "Theme"
  },
  "options": {
    "theme": {
      "system": "System",
      "dark": "Dark",
      "light": "Light"
    },
    "quality": {
      "original": "Original",
      "4k": "4K",
      "1080p": "1080p",
      "720p": "720p",
      "480p": "480p"
    },
    "language": {
      "auto": "Auto",
      "en": "English",
      "vi": "Vietnamese"
    }
  }
}
```

### Media (`locales/en/media.json`)
```json
{
  "actions": {
    "play": "Play",
    "addToFavorites": "Add to Favorites",
    "removeFromFavorites": "Remove from Favorites",
    "moreInfo": "More Info",
    "trailer": "Trailer",
    "cast": "Cast",
    "crew": "Crew",
    "similar": "Similar"
  },
  "filters": {
    "genre": "Genre",
    "year": "Year",
    "rating": "Rating",
    "sort": "Sort"
  },
  "search": {
    "placeholder": "Search movies, series, genres...",
    "results": "Results for",
    "noResults": "No results found"
  },
  "library": {
    "movies": "Movies",
    "tvshows": "TV Shows",
    "mixed": "Mixed"
  }
}
```

### Watch (`locales/en/watch.json`)
```json
{
  "controls": {
    "play": "Play",
    "pause": "Pause",
    "mute": "Mute",
    "unmute": "Unmute",
    "fullscreen": "Fullscreen",
    "exitFullscreen": "Exit Fullscreen",
    "nextEpisode": "Next Episode",
    "skipIntro": "Skip Intro",
    "skipCredits": "Skip Credits"
  },
  "subtitles": {
    "title": "Subtitles",
    "off": "Off",
    "custom": "Custom...",
    "search": "Search Subtitles"
  },
  "audio": {
    "title": "Audio",
    "auto": "Auto"
  },
  "quality": {
    "title": "Quality",
    "auto": "Auto"
  },
  "casting": {
    "title": "Cast to",
    "disconnect": "Disconnect"
  }
}
```

### Errors (`locales/en/errors.json`)
```json
{
  "generic": "Something went wrong",
  "network": "Network error",
  "notFound": "Not found",
  "unauthorized": "Session expired",
  "forbidden": "Access denied",
  "validation": "Invalid input",
  "server": "Server error"
}
```

---

## Implementation Pattern

### Before:
```tsx
<button>Sign In</button>
```

### After:
```tsx
import { useTranslation } from '@/hooks/useTranslation'

function LoginButton() {
  const { t } = useTranslation('auth')
  return <button>{t('login.signIn')}</button>
}
```

---

## Test Criteria

- [ ] All UI displays English strings correctly
- [ ] No hardcoded strings remaining in components
- [ ] TypeScript validates translation keys
- [ ] Console has no i18n warnings

---

## Notes

- Settings page is the largest - has ~14 sections
- Use nested keys: `t('settings.fields.username')`
- Keep keys descriptive but concise

---

Next Phase: [Phase 03 - Vietnamese Translation](./phase-03-vietnamese.md)
