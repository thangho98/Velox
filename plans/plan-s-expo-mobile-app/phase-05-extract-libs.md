# Phase 05: Extract Shared Libs (languages, image)
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Move pure utility modules sang shared.

## Implementation Steps

### 1. Move languages.ts
- [ ] Copy `webapp/src/lib/languages.ts` → `packages/shared/lib/languages.ts`
- [ ] Update `webapp/src/lib/languages.ts` → thin re-export:
  ```typescript
  export * from '@velox/shared/lib/languages'
  ```

### 2. Move image.ts
- [ ] Copy `webapp/src/lib/image.ts` → `packages/shared/lib/image.ts`
- [ ] Check nếu `image.ts` import từ `@/types/api` — nếu có, đổi thành `../types` (relative trong shared)
- [ ] Update `webapp/src/lib/image.ts` → thin re-export:
  ```typescript
  export * from '@velox/shared/lib/image'
  ```

### 3. capabilities.ts — KHÔNG move
- [ ] Xác nhận `capabilities.ts` (241 LOC) giữ nguyên trong `webapp/src/lib/`
- [ ] Lý do: dùng `MediaSource`, `document.createElement`, `navigator.userAgent`, `window.screen` — 100% browser APIs
- [ ] Mobile sẽ có `mobile/src/platform/capabilities.ts` riêng (hardcoded ExoPlayer capabilities)

### 4. Verify
- [ ] `cd webapp && pnpm build && pnpm lint`

## Files to Create
- `packages/shared/lib/languages.ts` — from webapp (110 LOC)
- `packages/shared/lib/image.ts` — from webapp (49 LOC)

## Files to Modify
- `webapp/src/lib/languages.ts` — replace with re-export
- `webapp/src/lib/image.ts` — replace with re-export

## Files NOT Modified
- `webapp/src/lib/capabilities.ts` — web-only, stays as-is

## Test Criteria
- [ ] Webapp build pass
- [ ] Subtitle language names display correctly
- [ ] TMDb image URLs resolve correctly

---
Next Phase: [phase-06-extract-hooks.md](phase-06-extract-hooks.md)
