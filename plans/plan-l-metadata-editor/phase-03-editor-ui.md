# Phase 03: Frontend — Editor UI
Status: ⬜ Pending
Dependencies: Phase 01 (PATCH API), Phase 02 (Image Upload)

## Objective
Tạo UI cho admin chỉnh sửa metadata trên detail page — giống modal "Edit Metadata" của Emby.
Bao gồm: edit text fields, genre tag editor, cast/crew editor, image upload (drag-and-drop).

## Requirements

### Functional
- [ ] Nút "Edit Metadata" trên MediaDetailPage + SeriesDetailPage (admin only)
- [ ] Full-page editor hoặc slide-over panel với form fields
- [ ] Text fields: title, sort_title, overview (textarea), tagline, release_date
- [ ] Number fields: rating
- [ ] Genre editor: tag-style input, autocomplete từ existing genres
- [ ] Credits editor: list cast/crew, thêm/xóa/reorder
- [ ] Image upload: drag-and-drop hoặc click-to-upload cho poster + backdrop
- [ ] Image preview: hiện ảnh hiện tại + preview ảnh mới trước khi save
- [ ] Toggle "Save NFO" checkbox
- [ ] Toggle "Lock Metadata" checkbox (default: on khi edit)
- [ ] Unlock button: "Unlock Metadata" cho locked items
- [ ] Save → gọi PATCH API → refresh detail page data
- [ ] Loading + error states

### Non-Functional
- [ ] Chỉ hiện cho admin (check auth store `user.is_admin`)
- [ ] Responsive: mobile-friendly layout
- [ ] Keyboard accessible: Tab order, Enter to save, Escape to cancel
- [ ] Không làm mất unsaved changes khi click outside (confirm dialog)
- [ ] Optimistic UI: show success immediately, rollback on error

## Implementation Steps

### 1. API Client Functions
- [ ] `webapp/src/api/metadata.ts` — NEW:
  ```typescript
  export async function patchMediaMetadata(id: number, data: MediaEditRequest): Promise<Media>
  export async function patchSeriesMetadata(id: number, data: SeriesEditRequest): Promise<Series>
  export async function uploadMediaImage(id: number, imageType: string, file: File): Promise<{path: string}>
  export async function uploadSeriesImage(id: number, imageType: string, file: File): Promise<{path: string}>
  export async function deleteMediaImage(id: number, imageType: string): Promise<void>
  export async function unlockMediaMetadata(id: number): Promise<void>
  export async function unlockSeriesMetadata(id: number): Promise<void>
  ```

### 2. Types
- [ ] `webapp/src/types/api.ts` — thêm:
  ```typescript
  interface MediaEditRequest {
    title?: string
    sort_title?: string
    overview?: string
    tagline?: string
    release_date?: string
    rating?: number
    genres?: string[]
    credits?: CreditInput[]
    save_nfo?: boolean
    metadata_locked?: boolean
  }

  interface CreditInput {
    person_name: string
    character?: string
    role: 'cast' | 'director' | 'writer' | 'producer'
    order: number
  }

  interface SeriesEditRequest {
    title?: string
    sort_title?: string
    overview?: string
    status?: string
    network?: string
    first_air_date?: string
    genres?: string[]
    save_nfo?: boolean
    metadata_locked?: boolean
  }
  ```

### 3. React Query Hooks
- [ ] `webapp/src/hooks/stores/useMedia.ts` — thêm mutations:
  ```typescript
  export function useEditMediaMetadata(mediaId: number)
  export function useEditSeriesMetadata(seriesId: number)
  export function useUploadMediaImage(mediaId: number)
  export function useUploadSeriesImage(seriesId: number)
  export function useDeleteMediaImage(mediaId: number)
  export function useUnlockMetadata(type: 'media' | 'series', id: number)
  ```
  - onSuccess: `refetchQueries` (không invalidate — tránh loading flash)

### 4. MetadataEditor Component
- [ ] `webapp/src/components/metadata/MetadataEditor.tsx` — NEW:
  - Props: `media | series`, `type: 'media' | 'series'`, `onClose`, `onSave`
  - Layout: slide-over panel từ phải (hoặc full-page modal)
  - Sections:
    1. **Basic Info**: title, sort_title, tagline, overview (textarea auto-resize)
    2. **Dates & Ratings**: release_date (date picker), rating (number input)
    3. **Images**: poster + backdrop upload areas
    4. **Genres**: tag editor
    5. **Cast & Crew**: sortable list
    6. **Options**: Save NFO checkbox, Lock Metadata toggle
  - Footer: Cancel + Save buttons
  - State: form dirty tracking, unsaved changes warning

### 5. GenreEditor Component
- [ ] `webapp/src/components/metadata/GenreEditor.tsx` — NEW:
  - Tag-style input: hiện genres dạng chips/badges
  - Click X để xóa genre
  - Text input + Enter để thêm genre mới
  - Autocomplete dropdown từ existing genres trong DB
  - Hook `useAllGenres()` để fetch danh sách genres

### 6. CreditEditor Component
- [ ] `webapp/src/components/metadata/CreditEditor.tsx` — NEW:
  - List view: mỗi credit là 1 row (name, character, role, order)
  - Tabs hoặc sections: Cast vs Crew
  - Add button: thêm row mới
  - Delete button: xóa row
  - Drag-to-reorder (hoặc up/down arrows)
  - Role dropdown: cast, director, writer, producer

### 7. ImageUploader Component
- [ ] `webapp/src/components/metadata/ImageUploader.tsx` — NEW:
  - Drop zone: drag-and-drop hoặc click to browse
  - Preview: hiện ảnh hiện tại (TMDb hoặc local)
  - Upload progress indicator
  - Delete button: xóa custom image
  - Accept: image/jpeg, image/png, image/webp
  - Max size hint: "Max 10MB"
  - Sau upload thành công: hiện ảnh mới ngay lập tức

### 8. Integration vào Detail Pages
- [ ] `webapp/src/pages/MediaDetailPage.tsx`:
  - Thêm "Edit Metadata" button (chỉ hiện cho admin)
  - Thêm "Locked" badge nếu metadata_locked = true
  - State: `showEditor` boolean
  - Render `<MetadataEditor>` khi showEditor = true
  - Sau save: refetch media data
- [ ] `webapp/src/pages/SeriesDetailPage.tsx`:
  - Tương tự MediaDetailPage

## Files to Create/Modify
- `webapp/src/api/metadata.ts` — NEW: API client functions
- `webapp/src/types/api.ts` — thêm edit request types
- `webapp/src/hooks/stores/useMedia.ts` — thêm mutations
- `webapp/src/components/metadata/MetadataEditor.tsx` — NEW: main editor
- `webapp/src/components/metadata/GenreEditor.tsx` — NEW: genre tag editor
- `webapp/src/components/metadata/CreditEditor.tsx` — NEW: cast/crew editor
- `webapp/src/components/metadata/ImageUploader.tsx` — NEW: image upload
- `webapp/src/pages/MediaDetailPage.tsx` — Edit button + editor integration
- `webapp/src/pages/SeriesDetailPage.tsx` — Edit button + editor integration
- `webapp/src/lib/image.ts` — mediaImage() helper (if not done in Phase 02)

## UI Mockup (ASCII)

### Detail Page (với Edit button)
```
┌─────────────────────────────────────────────┐
│ [Backdrop Image                           ] │
│                                             │
│  [Poster]  The Matrix (1999)      [🔒 Locked]│
│            ⭐ 8.7  🍅 87%         [✏️ Edit] │
│            Action, Sci-Fi                   │
│                                             │
│  A computer hacker learns about the true    │
│  nature of his reality...                   │
│                                             │
│  Cast: Keanu Reeves, Laurence Fishburne... │
└─────────────────────────────────────────────┘
```

### Editor Panel (slide-over)
```
┌──────────────────────────────────────┐
│ Edit Metadata                    [✕] │
├──────────────────────────────────────┤
│                                      │
│ Title: [The Matrix              ]    │
│ Sort:  [Matrix, The             ]    │
│ Tagline: [Welcome to the Real Wo]    │
│                                      │
│ Overview:                            │
│ ┌──────────────────────────────┐     │
│ │ A computer hacker learns     │     │
│ │ about the true nature of his │     │
│ │ reality...                   │     │
│ └──────────────────────────────┘     │
│                                      │
│ Release Date: [1999-03-31]           │
│ Rating:       [8.7       ]           │
│                                      │
│ ── Images ──                         │
│ Poster:    [drag image here ▼]       │
│ Backdrop:  [drag image here ▼]       │
│                                      │
│ ── Genres ──                         │
│ [Action ✕] [Sci-Fi ✕] [+ Add]       │
│                                      │
│ ── Cast ──                           │
│ 1. Keanu Reeves → Neo        [🗑️]    │
│ 2. Laurence Fishburne → Morpheus [🗑️]│
│ [+ Add Cast]                         │
│                                      │
│ ── Options ──                        │
│ ☑ Lock Metadata (skip on rescan)     │
│ ☐ Save NFO file                      │
│                                      │
├──────────────────────────────────────┤
│              [Cancel]  [💾 Save]     │
└──────────────────────────────────────┘
```

## Test Criteria
- [ ] Admin sees "Edit Metadata" button, non-admin does not
- [ ] Edit title → Save → detail page shows new title immediately
- [ ] Edit genres (remove + add) → Save → genres updated
- [ ] Add cast member → Save → cast list updated
- [ ] Upload poster → preview shows → Save → detail page shows new poster
- [ ] Upload > 10MB → error message shown
- [ ] Click Cancel with unsaved changes → confirm dialog appears
- [ ] Locked badge visible for locked items
- [ ] Unlock button works → removes locked badge
- [ ] Edit form pre-fills with current metadata values
- [ ] Empty title → validation error, cannot save
- [ ] Network error on save → error toast, form stays open

## Notes
- Dùng slide-over panel (không phải separate page) — UX tốt hơn, không mất context
- React Compiler đã bật — không cần useMemo/useCallback
- Genre autocomplete: fetch all genres once, filter client-side
- Credit reorder: có thể dùng simple up/down buttons thay vì drag-and-drop (giảm complexity)
- Image upload nên dùng Preview trước khi Save — tránh upload rồi mới thấy sai

---
Next Phase: [Phase 04 — NFO Write + Sync](phase-04-nfo-write.md)
