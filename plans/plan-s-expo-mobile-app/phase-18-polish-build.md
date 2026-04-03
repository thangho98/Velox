# Phase 18: Polish + Build APK
Status: ⬜ Pending
Dependencies: Phase 09-17

## Objective
App icon, splash screen, error handling, performance optimization, APK build.

## Implementation Steps

### 1. App Icon + Splash Screen
- [ ] Design app icon (Velox logo, 1024x1024)
- [ ] Configure in `app.json`:
  ```json
  {
    "expo": {
      "icon": "./assets/icon.png",
      "splash": {
        "image": "./assets/splash.png",
        "resizeMode": "contain",
        "backgroundColor": "#141414"
      },
      "android": {
        "adaptiveIcon": {
          "foregroundImage": "./assets/adaptive-icon.png",
          "backgroundColor": "#141414"
        }
      }
    }
  }
  ```

### 2. Loading States
- [ ] Skeleton placeholders cho:
  - Media grid (poster shape skeletons)
  - Home screen rows
  - Detail pages (backdrop + text placeholders)
  - Episode list
- [ ] Pull-to-refresh on all list screens
- [ ] ActivityIndicator for detail page first load

### 3. Error Handling
- [ ] Network error screen component:
  - "Cannot connect to server"
  - "Check your network connection"
  - "Retry" button
  - Auto-detect with `@react-native-community/netinfo`
- [ ] Server timeout handling (AbortController in shared API client)
- [ ] Toast notifications for actions:
  - "Added to favorites"
  - "Removed from favorites"
  - "Progress saved"
  - Error toasts (red)

### 4. EAS Build Configuration
- [ ] Install EAS CLI: `npm i -g eas-cli`
- [ ] `eas.json`:
  ```json
  {
    "cli": { "version": ">= 15.0.0" },
    "build": {
      "development": {
        "developmentClient": true,
        "distribution": "internal"
      },
      "preview": {
        "distribution": "internal",
        "android": { "buildType": "apk" }
      },
      "production": {
        "android": { "buildType": "apk" }
      }
    }
  }
  ```

### 5. Build APK
- [ ] ```bash
  cd mobile
  eas build --platform android --profile preview
  # → Downloads .apk for sideloading
  ```

### 6. Testing Checklist
- [ ] Test on physical Android device (install APK)
- [ ] Test on Android emulator (different screen sizes: phone, tablet)
- [ ] Test landscape + portrait orientation handling
- [ ] Test with slow network (throttle in emulator)
- [ ] Test background playback (play video → home button → audio continues)
- [ ] Test deep linking: `velox://watch/123` (if configured)
- [ ] Test app kill → reopen → still logged in, preferences preserved

### 7. Performance
- [ ] `expo-image` for all images (automatic caching, blurhash placeholders)
- [ ] FlatList optimization:
  - `getItemLayout` for fixed-height items
  - `windowSize={5}` to reduce off-screen rendering
  - `removeClippedSubviews={true}`
- [ ] Memory profiling with Flipper or React DevTools
- [ ] Avoid re-renders: memoize MediaCard with `React.memo`

## Files to Create
- `mobile/assets/icon.png`
- `mobile/assets/splash.png`
- `mobile/assets/adaptive-icon.png`
- `mobile/eas.json`
- `mobile/src/components/Skeleton.tsx` — skeleton placeholder component
- `mobile/src/components/NetworkError.tsx` — network error screen
- `mobile/src/components/Toast.tsx` — toast notification

## Files to Modify
- `mobile/app.json` — icon, splash, permissions
- Various screens — add loading/error states

## Test Criteria
- [ ] APK installs and runs on physical Android device
- [ ] App icon + splash screen display correctly
- [ ] Error states handled gracefully (no crashes)
- [ ] Performance: smooth scrolling in media grid (60fps)
- [ ] Performance: fast image loading with caching
- [ ] All core features work end-to-end: login → browse → play → subtitles → settings
