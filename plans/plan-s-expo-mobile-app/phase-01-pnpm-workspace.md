# Phase 01: pnpm Workspace + Shared Package Skeleton
Status: ⬜ Pending
Dependencies: None

## Objective
Chuyển từ npm → pnpm workspaces. Tạo skeleton cho `packages/shared/`.

## Context
- Root `package.json` hiện dùng npm với `concurrently` + `env-cmd` + `husky` + `lint-staged`
- Webapp dùng npm (`webapp/package.json`)
- KHÔNG có `pnpm-workspace.yaml`

## Implementation Steps

### 1. Install pnpm + tạo workspace config
- [ ] Install pnpm: `npm i -g pnpm`
- [ ] Tạo `pnpm-workspace.yaml` ở root:
  ```yaml
  packages:
    - 'packages/*'
    - 'webapp'
    - 'mobile'
  ```

### 2. Chuyển root package.json sang pnpm
- [ ] Xóa `node_modules/` ở root + webapp
- [ ] Chạy `pnpm install` (pnpm tự detect workspace)
- [ ] Giữ nguyên scripts (`dev`, `build`) — chỉ thay `npm run` → `pnpm`
- [ ] Giữ nguyên husky + lint-staged config

### 3. Tạo shared package skeleton
- [ ] Tạo cấu trúc thư mục:
  ```
  packages/shared/
  ├── package.json
  ├── tsconfig.json
  ├── types/
  │   └── index.ts          # barrel — empty for now
  ├── api/
  │   └── index.ts
  ├── hooks/
  │   └── index.ts
  ├── stores/
  │   └── index.ts
  ├── lib/
  │   └── index.ts
  └── platform.ts           # PlatformAdapter interface
  ```

- [ ] `packages/shared/package.json`:
  ```json
  {
    "name": "@velox/shared",
    "private": true,
    "version": "0.0.0",
    "type": "module",
    "main": "index.ts",
    "types": "index.ts",
    "exports": {
      "./types": "./types/index.ts",
      "./types/*": "./types/*.ts",
      "./api": "./api/index.ts",
      "./hooks": "./hooks/index.ts",
      "./hooks/*": "./hooks/*.ts",
      "./stores": "./stores/index.ts",
      "./lib/*": "./lib/*.ts",
      "./platform": "./platform.ts"
    },
    "peerDependencies": {
      "@tanstack/react-query": "^5.0.0",
      "react": "^18.0.0 || ^19.0.0",
      "zustand": "^5.0.0"
    }
  }
  ```

- [ ] `packages/shared/tsconfig.json`:
  ```json
  {
    "compilerOptions": {
      "target": "ES2022",
      "module": "ESNext",
      "moduleResolution": "bundler",
      "strict": true,
      "esModuleInterop": true,
      "skipLibCheck": true,
      "declaration": true,
      "declarationMap": true,
      "sourceMap": true,
      "outDir": "./dist",
      "rootDir": ".",
      "jsx": "react-jsx"
    },
    "include": ["./**/*.ts", "./**/*.tsx"],
    "exclude": ["node_modules", "dist"]
  }
  ```

### 4. Tạo PlatformAdapter interface

- [ ] File `packages/shared/platform.ts`:
  ```typescript
  // Platform abstraction — web vs mobile inject their own implementations

  export interface StorageAdapter {
    getItem(key: string): string | null | Promise<string | null>
    setItem(key: string, value: string): void | Promise<void>
    removeItem(key: string): void | Promise<void>
  }

  export interface PlatformAdapter {
    /** Regular storage (settings, UI state) — web: localStorage, mobile: MMKV */
    storage: StorageAdapter

    /** Secure storage (tokens) — web: localStorage, mobile: expo-secure-store */
    secureStorage: StorageAdapter

    /** Device name for X-Device-Name header */
    getDeviceName(): string

    /** Base URL for API — web: '/api' (relative, proxied), mobile: 'http://192.168.1.x:8098/api' */
    getApiBaseUrl(): string
  }

  // Singleton — set once at app init, used by api client + stores
  let _platform: PlatformAdapter | null = null

  export function initPlatform(adapter: PlatformAdapter): void {
    _platform = adapter
  }

  export function getPlatform(): PlatformAdapter {
    if (!_platform) throw new Error('Platform not initialized. Call initPlatform() first.')
    return _platform
  }
  ```

### 5. Verify
- [ ] `pnpm install` thành công, tất cả packages resolve
- [ ] `cd webapp && pnpm build` pass (chưa thay đổi gì trong webapp)
- [ ] Root scripts (`pnpm dev`, `pnpm build`) hoạt động

## Files to Create
- `pnpm-workspace.yaml` — new
- `packages/shared/package.json` — new
- `packages/shared/tsconfig.json` — new
- `packages/shared/platform.ts` — new
- `packages/shared/types/index.ts` — new (empty barrel)
- `packages/shared/api/index.ts` — new (empty barrel)
- `packages/shared/hooks/index.ts` — new (empty barrel)
- `packages/shared/stores/index.ts` — new (empty barrel)
- `packages/shared/lib/index.ts` — new (empty barrel)

## Files to Modify
- `package.json` (root) — update scripts `npm run` → `pnpm`

## Test Criteria
- [ ] `pnpm install` — no errors
- [ ] `cd webapp && pnpm build` — build success
- [ ] `pnpm dev` — both backend + webapp start

---
Next Phase: [phase-02-extract-types.md](phase-02-extract-types.md)
