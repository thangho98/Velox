# Phase 09: Expo Project Setup + Development Build
Status: ⬜ Pending
Dependencies: Phase 08

## Objective
Tạo Expo project với **development build** (CNG — Continuous Native Generation). KHÔNG dùng Expo Go vì project cần native modules: `react-native-mmkv`, `expo-video`, `expo-secure-store`.

## Context — Tại sao không dùng Expo Go?
- `react-native-mmkv` yêu cầu custom native code (JSI C++ bridge)
- `expo-video` cần native ExoPlayer integration
- `expo-secure-store` cần Android Keystore
- → Phải dùng **development build** (`npx expo prebuild` + `npx expo run:android`) hoặc **EAS Build**

## Implementation Steps

### 1. Create Expo project
- [ ] ```bash
  npx create-expo-app@latest mobile --template blank-typescript
  cd mobile
  ```

### 2. Update mobile/package.json — add shared + deps
- [ ] ```json
  {
    "dependencies": {
      "@velox/shared": "workspace:*",
      "@tanstack/react-query": "^5.90.0",
      "zustand": "^5.0.0",
      "expo-router": "latest",
      "expo-secure-store": "latest",
      "expo-image": "latest",
      "expo-video": "latest",
      "expo-screen-orientation": "latest",
      "expo-status-bar": "latest",
      "expo-build-properties": "latest",
      "expo-keep-awake": "latest",
      "expo-brightness": "latest",
      "react-native-mmkv": "latest",
      "nativewind": "latest",
      "tailwindcss": "^4.0.0",
      "react-native-safe-area-context": "latest",
      "react-native-screens": "latest",
      "react-native-reanimated": "latest",
      "react-native-gesture-handler": "latest",
      "@react-native-community/netinfo": "latest"
    }
  }
  ```
- [ ] `pnpm install`

### 3. Configure app.json for development build + cleartext
- [ ] `mobile/app.json`:
  ```json
  {
    "expo": {
      "name": "Velox",
      "slug": "velox-mobile",
      "version": "1.0.0",
      "scheme": "velox",
      "newArchEnabled": true,
      "icon": "./assets/icon.png",
      "splash": {
        "image": "./assets/splash.png",
        "resizeMode": "contain",
        "backgroundColor": "#141414"
      },
      "android": {
        "package": "com.velox.mobile",
        "adaptiveIcon": {
          "foregroundImage": "./assets/adaptive-icon.png",
          "backgroundColor": "#141414"
        }
      },
      "ios": {
        "bundleIdentifier": "com.velox.mobile",
        "infoPlist": {
          "UIBackgroundModes": ["audio"],
          "NSAppTransportSecurity": {
            "NSAllowsArbitraryLoads": true
          }
        }
      },
      "plugins": [
        "expo-router",
        "expo-secure-store",
        [
          "expo-build-properties",
          {
            "android": {
              "usesCleartextTraffic": true
            }
          }
        ]
      ]
    }
  }
  ```
  ⚠️ **`usesCleartextTraffic: true`** — bắt buộc cho LAN HTTP (`http://192.168.1.x`). Android 9+ mặc định block cleartext. Dùng `expo-build-properties` plugin để inject vào AndroidManifest.

### 4. Setup Expo Router
- [ ] File structure:
  ```
  mobile/app/
  ├── _layout.tsx         # Root layout: providers
  ├── server-config.tsx   # Step 1: Enter server URL
  ├── login.tsx           # Step 2: Login
  └── (tabs)/             # Protected routes (tab navigation)
      ├── _layout.tsx     # Tab bar config
      ├── index.tsx       # Home tab (placeholder)
      ├── library.tsx     # Library tab (placeholder)
      ├── search.tsx      # Search tab (placeholder)
      └── profile.tsx     # Profile tab (placeholder)
  ```

### 5. Setup NativeWind
- [ ] Follow NativeWind v4 + Expo setup docs
- [ ] `mobile/tailwind.config.ts`
- [ ] `mobile/global.css` with `@tailwind` directives
- [ ] Configure babel/metro for NativeWind

### 6. Setup TanStack Query + root layout
- [ ] `mobile/app/_layout.tsx`:
  ```typescript
  import '../global.css'
  import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
  import { Stack } from 'expo-router'

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: 2, staleTime: 60_000 },
    },
  })

  export default function RootLayout() {
    return (
      <QueryClientProvider client={queryClient}>
        <Stack screenOptions={{ headerShown: false }} />
      </QueryClientProvider>
    )
  }
  ```

### 7. Generate native project + verify build
- [ ] Generate native Android project (CNG):
  ```bash
  npx expo prebuild --platform android
  ```
- [ ] Build + run on emulator:
  ```bash
  npx expo run:android
  ```
  ⚠️ Cần Android SDK + emulator đã cài sẵn.
  
  Hoặc dùng EAS development build:
  ```bash
  eas build --profile development --platform android
  ```

### 8. Verify
- [ ] App launches on Android emulator
- [ ] NativeWind styles render correctly
- [ ] TanStack Query provider active (no runtime errors)
- [ ] `import type { Media } from '@velox/shared/types'` compiles
- [ ] MMKV initializes without crash (JSI bridge works)

## Files to Create
- `mobile/` — entire Expo project (via create-expo-app)
- `mobile/app/_layout.tsx` — root layout with providers
- `mobile/app/(tabs)/_layout.tsx` — tab bar
- `mobile/app.json` — with plugins + cleartext config
- `mobile/tailwind.config.ts`
- `mobile/global.css`

## Important Notes
- **KHÔNG dùng `npx expo start`** với Expo Go — native modules sẽ crash
- **Dùng `npx expo run:android`** hoặc EAS development build
- `npx expo prebuild` generates `android/` folder — add to `.gitignore` (CNG regenerates it)
- Mỗi khi thêm native package mới → re-run `npx expo prebuild --clean`

## Test Criteria
- [ ] `npx expo run:android` — app builds + launches on emulator
- [ ] MMKV read/write works (no JSI crash)
- [ ] SecureStore read/write works
- [ ] expo-video basic test (can create VideoPlayer without crash)
- [ ] Cleartext HTTP works: `fetch('http://10.0.2.2:8098/api/setup/status')` succeeds from emulator

---
Next Phase: [phase-10-platform-adapters.md](phase-10-platform-adapters.md)
