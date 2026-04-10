# Phase 03: Settings Screen Tabs
Status: ⬜ Pending
Dependencies: None

## Objective
Thêm đầy đủ Settings tabs vào Android Native: Cinema, Subtitles, Admin sections.

## Features to Implement

### 1. Profile Tab (EXISTING - enhance if needed)
```
- Display user info
- Edit display name
- Change password
```

### 2. Preferences Tab
```
- Playback quality
- Auto-play trailers
- Subtitle default language
- Notification preferences
```

### 3. Security Tab
```
- Sessions list (logged in devices)
- Revoke session
- Change password
```

### 4. Cinema Tab (NEW - from Mobile AdminScreen)
```
- Enable/disable cinema mode
- Max trailers setting
- Custom intro upload (future)
```

### 5. Subtitles Tab (NEW)
```
- Default subtitle language
- Subtitle appearance (size, color)
- Download subtitles auto
- Subtitle translate toggle
```

### 6. Admin Tab (NEW - from Mobile AdminScreen components)
```
- Library Stats section
- Tasks section (background scans)
- Markers section (chapter markers)
- Translate section (TMDb metadata)
- Webhooks section
```

## Files to Create/Modify

### New Files:
- `presentation/ui/screens/settings/CinemaSettingsScreen.kt`
- `presentation/ui/screens/settings/SubtitlesSettingsScreen.kt`
- `presentation/ui/screens/settings/AdminSettingsScreen.kt`
- `presentation/viewmodel/CinemaSettingsViewModel.kt`
- `presentation/viewmodel/SubtitlesSettingsViewModel.kt`
- `presentation/viewmodel/AdminSettingsViewModel.kt`

### Modify Files:
- `presentation/ui/screens/settings/SettingsScreen.kt` - Add tab navigation
- `presentation/viewmodel/SettingsViewModel.kt` - Add all settings state
- `presentation/navigation/Screen.kt` - Add new screen routes
- `presentation/navigation/VeloxNavHost.kt` - Add navigation

## Implementation Steps

1. **Create Settings tab structure**
   - [x] Add `selectedTab` state (profile/preferences/security/cinema/subtitles/admin)
   - [x] Create TabRow at top of Settings screen
   - [x] Show/hide tabs based on user isAdmin

2. **Migrate Cinema Settings**
   - [ ] Read Mobile CinemaSection component
   - [ ] Create similar UI in Compose
   - [ ] Connect to API (GET/PUT /api/settings/cinema)

3. **Migrate Subtitles Settings**
   - [ ] Read Mobile subtitle settings logic
   - [ ] Create UI for subtitle preferences
   - [ ] Save to DataStore

4. **Migrate Admin Sections**
   - [ ] LibraryStatsSection
   - [ ] TasksSection
   - [ ] MarkersSection
   - [ ] TranslateSection
   - [ ] WebhooksSection

## Test Criteria
- [ ] All tabs visible and navigable
- [ ] Cinema settings save/load correctly
- [ ] Subtitle preferences persist
- [ ] Admin sections load data from API

## Reference
Mobile implementation:
- `mobile/src/screens/SettingsScreen.tsx` (main container)
- `mobile/src/components/admin/CinemaSection.tsx`
- `mobile/src/components/admin/TranslateSection.tsx`
- `mobile/src/components/admin/MarkersSection.tsx`
- `mobile/src/components/admin/TasksSection.tsx`
- `mobile/src/components/admin/WebhooksSection.tsx`
- `mobile/src/components/admin/LibraryStatsSection.tsx`
