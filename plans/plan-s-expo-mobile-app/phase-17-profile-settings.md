# Phase 17: Profile + Settings + Favorites
Status: ⬜ Pending
Dependencies: Phase 12

## Objective
Profile tab, user settings, favorites list, watch history.

## Context — API Access
⚠️ `useServerInfo()` calls `GET /admin/server` — admin-only endpoint, returns 403 for regular users.
Mobile app phải handle cả admin + regular user. Profile screen chỉ dùng endpoints accessible by all users.

## Implementation Steps

### 1. Profile Screen
- [ ] `mobile/app/(tabs)/profile.tsx`:
  ```typescript
  import { useProfile } from '@velox/shared/hooks/useAuth'
  import { useLogout } from '@velox/shared/hooks/useAuth'
  import { useAuthStore } from '../../src/stores/auth'
  import { getServerUrl } from '../../src/platform/storage'

  export default function ProfileScreen() {
    const { data: profile } = useProfile()
    const user = useAuthStore((s) => s.user)
    const logout = useLogout()

    return (
      <SafeAreaView className="flex-1 bg-black">
        <ScrollView>
          {/* User avatar (initials circle) + display name */}
          <View className="items-center py-8">
            <View className="w-20 h-20 rounded-full bg-red-600 items-center justify-center">
              <Text className="text-white text-2xl font-bold">
                {profile?.display_name?.[0]?.toUpperCase() ?? '?'}
              </Text>
            </View>
            <Text className="text-white text-xl mt-3">{profile?.display_name}</Text>
            <Text className="text-gray-400 text-sm">{user?.username}</Text>
            {user?.is_admin && (
              <View className="bg-red-600/20 px-2 py-1 rounded mt-1">
                <Text className="text-red-400 text-xs">Admin</Text>
              </View>
            )}
          </View>

          {/* Menu items — accessible to all users */}
          <MenuItem icon="heart" title="Favorites" onPress={() => router.push('/favorites')} />
          <MenuItem icon="clock" title="Watch History" onPress={() => router.push('/history')} />
          <MenuItem icon="settings" title="Settings" onPress={() => router.push('/settings')} />

          {/* Server info — basic (no admin endpoint) */}
          <View className="px-4 py-6 border-t border-gray-800">
            <Text className="text-gray-500 text-xs">
              Connected to: {getServerUrl()}
            </Text>
          </View>

          {/* Actions */}
          <Button title="Change Server" variant="outline" onPress={handleChangeServer} />
          <Button title="Log Out" variant="danger" onPress={() => logout.mutate()} />
        </ScrollView>
      </SafeAreaView>
    )
  }
  ```
  ⚠️ **KHÔNG dùng `useServerInfo()`** — admin-only. Show server URL từ MMKV thay thế.
  Nếu user là admin, có thể conditionally show thêm admin info:
  ```typescript
  const { data: serverInfo } = useServerInfo({ enabled: user?.is_admin })
  // Show version only for admin
  ```

### 2. Favorites List
- [ ] Tạo `mobile/app/favorites.tsx`:
  - `useFavorites()` from `@velox/shared/hooks/media`
  - Grid layout (same as library, numColumns=3)
  - Each card: poster + title
  - Long press → unfavorite (with confirmation alert)
  - Empty state: "No favorites yet — tap the heart icon on any movie or show"

### 3. Watch History
- [ ] Tạo `mobile/app/history.tsx`:
  - `useRecentlyWatched()` from `@velox/shared/hooks/media`
  - List layout (not grid — more info per row):
    - Poster thumbnail (small, left)
    - Title + last played date (right)
    - Progress bar
  - Tap → navigate to media detail

### 4. Settings Screen
- [ ] Tạo `mobile/app/settings.tsx`:
  - Read/write from `usePlayerStore` (shared, persisted in MMKV)
  - Sections:

  **Playback:**
  - Default quality: picker (Auto, 480p, 720p, 1080p, Original)
    → `usePlayerStore.setMaxQuality()`
  - Default subtitle language: picker (from `LANG_NAMES` in `@velox/shared/lib/languages`)
    → `usePlayerStore.setSubtitleLanguage()`
  - Default audio language: picker
    → `usePlayerStore.setAudioTrack()`

  **Subtitles:**
  - Subtitle size: Small / Medium / Large
    → `usePlayerStore.setSubtitleSize()`
  - Subtitle color: preset picker (White, Yellow, Cyan)
    → `usePlayerStore.setSubtitleColor()`
  - Subtitle background: Solid / Semi / None
    → `usePlayerStore.setSubtitleBackground()`

  **Server:**
  - Current server URL (display only, from `getServerUrl()`)
  - "Change Server" → navigate to server-config

  **About:**
  - App version (from Constants.expoConfig.version)
  - Server version (only if admin: `useServerInfo()`)

## Files to Create
- `mobile/app/(tabs)/profile.tsx` — profile tab
- `mobile/app/favorites.tsx` — favorites grid
- `mobile/app/history.tsx` — watch history list
- `mobile/app/settings.tsx` — settings screen
- `mobile/src/components/MenuItem.tsx` — reusable menu item row

## Test Criteria
- [ ] Profile shows user info (display name, username, admin badge)
- [ ] Profile works for regular users (no 403 errors)
- [ ] Favorites grid shows favorited items
- [ ] Watch history shows recently watched with timestamps
- [ ] Settings: quality/subtitle/audio defaults persist across app restart
- [ ] Logout clears auth + navigates to login
- [ ] Change server → navigates to server config

---
Next Phase: [phase-18-polish-build.md](phase-18-polish-build.md)
