# Phase 02: Extract Types → @velox/shared
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Move type definitions sang shared package. Webapp vẫn import từ path cũ (barrel re-export).

## Context
- `webapp/src/types/` có 7 files (633 LOC): `common.ts`, `auth.ts`, `media.ts`, `series.ts`, `playback.ts`, `admin.ts`, `api.ts` (barrel)
- Tất cả đều pure TypeScript interfaces/types — KHÔNG có DOM/browser refs
- `api.ts` là barrel re-export từ 6 domain files

## Implementation Steps

### 1. Copy type files sang shared
- [ ] Copy 6 domain type files:
  ```bash
  cp webapp/src/types/common.ts   packages/shared/types/common.ts
  cp webapp/src/types/auth.ts     packages/shared/types/auth.ts
  cp webapp/src/types/media.ts    packages/shared/types/media.ts
  cp webapp/src/types/series.ts   packages/shared/types/series.ts
  cp webapp/src/types/playback.ts packages/shared/types/playback.ts
  cp webapp/src/types/admin.ts    packages/shared/types/admin.ts
  ```

### 2. Tạo barrel `packages/shared/types/index.ts`
- [ ] ```typescript
  export * from './common'
  export * from './auth'
  export * from './media'
  export * from './series'
  export * from './playback'
  export * from './admin'
  ```

### 3. Update webapp barrel — re-export từ shared
- [ ] Update `webapp/src/types/api.ts`:
  ```typescript
  // Re-export all types from shared package
  // Webapp imports still work: import { Media } from '@/types/api'
  export * from '@velox/shared/types'
  ```

### 4. Xóa duplicate domain files
- [ ] Xóa `webapp/src/types/common.ts`, `auth.ts`, `media.ts`, `series.ts`, `playback.ts`, `admin.ts`
- [ ] Giữ `webapp/src/types/api.ts` (barrel re-export)

### 5. Update webapp package.json
- [ ] Thêm dependency:
  ```json
  {
    "dependencies": {
      "@velox/shared": "workspace:*"
    }
  }
  ```

### 6. Update webapp tsconfig nếu cần
- [ ] Nếu Vite không tự resolve workspace package, thêm path alias trong `webapp/tsconfig.app.json`:
  ```json
  {
    "compilerOptions": {
      "paths": {
        "@/*": ["./src/*"],
        "@velox/shared/*": ["../packages/shared/*"]
      }
    }
  }
  ```

## Files to Create
- `packages/shared/types/common.ts` — from webapp
- `packages/shared/types/auth.ts` — from webapp
- `packages/shared/types/media.ts` — from webapp
- `packages/shared/types/series.ts` — from webapp
- `packages/shared/types/playback.ts` — from webapp
- `packages/shared/types/admin.ts` — from webapp
- `packages/shared/types/index.ts` — barrel

## Files to Modify
- `webapp/src/types/api.ts` — replace with re-export
- `webapp/package.json` — add @velox/shared dependency
- `webapp/tsconfig.app.json` — possibly add path alias

## Files to Delete
- `webapp/src/types/common.ts`
- `webapp/src/types/auth.ts`
- `webapp/src/types/media.ts`
- `webapp/src/types/series.ts`
- `webapp/src/types/playback.ts`
- `webapp/src/types/admin.ts`

## Test Criteria
- [ ] `pnpm install` — resolves @velox/shared
- [ ] `cd webapp && pnpm build && pnpm lint` — zero errors
- [ ] Không thay đổi import nào trong webapp components (vẫn `import { X } from '@/types/api'`)

---
Next Phase: [phase-03-extract-api-client.md](phase-03-extract-api-client.md)
