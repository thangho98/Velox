# Phase 04: SubtitleSearchModal (Frontend)
Status: ⬜ Pending
Dependencies: Phase 03

## Objective
Modal tìm subtitle được trigger từ "Search for Subtitles" button trong SubtitlePicker. Hiện kết quả từ cả 2 providers, cho phép download và áp dụng ngay.

---

## Components

### SubtitleSearchModal.tsx

**Props:**
```typescript
interface SubtitleSearchModalProps {
  mediaId: number
  defaultLang?: string | null      // ngôn ngữ hiện tại của player
  onClose: () => void
  onSubtitleDownloaded: () => void  // callback để refresh subtitle list
}
```

**State:**
- `lang` — ngôn ngữ đang search (dropdown: en, vi, fr, de, ja, ko, zh, es, pt, ...)
- `results` — kết quả search
- `downloading` — externalID đang download
- `downloaded` — set của externalIDs đã download thành công

**Layout:**
```
┌──────────────────────────────────────────┐
│  🔍 Search Subtitles           [×]       │
│  ──────────────────────────────────────  │
│  Language: [English ▼]   [Search again]  │
│  ──────────────────────────────────────  │
│  [Loading spinner / results list]        │
│                                          │
│  ┌─ OpenSubtitles ──────────────────┐   │
│  │ ✓ English (SRT) · 10k DL · ⭐4.5 │   │
│  │   Some.Movie.2023.BLURAY         │   │
│  │                   [Download ↓]   │   │
│  └──────────────────────────────────┘   │
│  ┌─ Podnapisi ──────────────────────┐   │
│  │   English (SRT) · 5k DL          │   │
│  │   Some.Movie.2023.WEB-DL         │   │
│  │                   [Download ↓]   │   │
│  └──────────────────────────────────┘   │
└──────────────────────────────────────────┘
```

**Behavior:**
- Mở modal → auto-search với `defaultLang` (hoặc "en" nếu null)
- Mỗi result hiện: provider badge, title, language, format, downloads, rating
- Click Download → POST /api/media/{id}/subtitles/download → spinner → checkmark
- Sau download: `onSubtitleDownloaded()` để WatchPage invalidate query và refresh subtitle list

---

## API hooks (useMedia.ts)

```typescript
// Search subtitles from external providers
export function useSubtitleSearch(mediaId: number, lang: string, enabled: boolean) {
  return useQuery({
    queryKey: ['subtitleSearch', mediaId, lang],
    queryFn: () => subtitleSearchApi.search(mediaId, lang),
    enabled: enabled && mediaId > 0,
    staleTime: 2 * 60 * 1000, // 2 min cache
  })
}

// Download + save subtitle
export function useDownloadSubtitle(mediaId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { provider: string; external_id: string }) =>
      subtitleSearchApi.download(mediaId, body),
    onSuccess: () => {
      // Invalidate subtitles so SubtitlePicker refreshes
      qc.invalidateQueries({ queryKey: ['media', mediaId, 'subtitles'] })
    },
  })
}
```

---

## API functions (api/subtitleSearch.ts)

```typescript
export const subtitleSearchApi = {
  search: (mediaId: number, lang: string): Promise<SubtitleSearchResult[]> =>
    apiFetch(`/api/media/${mediaId}/subtitles/search?lang=${lang}`).then(r => r.data),

  download: (mediaId: number, body: { provider: string; external_id: string }): Promise<Subtitle> =>
    apiFetch(`/api/media/${mediaId}/subtitles/download`, {
      method: 'POST',
      body: JSON.stringify(body),
    }).then(r => r.data),
}
```

---

## Types (types/api.ts additions)

```typescript
export interface SubtitleSearchResult {
  provider: 'opensubtitles' | 'podnapisi'
  external_id: string
  title: string
  language: string
  format: string
  downloads: number
  rating: number
  forced: boolean
  hearing_impaired: boolean
}
```

---

## SubtitlePicker modifications

```typescript
// Thêm prop
interface SubtitlePickerProps {
  // ... existing props
  mediaId: number          // cần để truyền vào modal
  onSubtitleAdded?: () => void  // callback sau khi download
}

// "Search for Subtitles" button mở modal thay vì không làm gì
const [showSearch, setShowSearch] = useState(false)

// Render modal
{showSearch && (
  <SubtitleSearchModal
    mediaId={mediaId}
    defaultLang={primaryLanguage}
    onClose={() => setShowSearch(false)}
    onSubtitleDownloaded={() => { onSubtitleAdded?.(); setShowSearch(false) }}
  />
)}
```

---

## Files
- `webapp/src/components/SubtitleSearchModal.tsx` — tạo mới
- `webapp/src/api/subtitleSearch.ts` — tạo mới
- `webapp/src/types/api.ts` — thêm SubtitleSearchResult interface
- `webapp/src/hooks/stores/useMedia.ts` — thêm useSubtitleSearch + useDownloadSubtitle
- `webapp/src/components/SubtitlePicker.tsx` — thêm mediaId prop + wire modal
- `webapp/src/pages/WatchPage.tsx` — pass mediaId vào SubtitlePicker
