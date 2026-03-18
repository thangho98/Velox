# Phase 03: Folder Browser
Status: ⬜ Pending
Dependencies: Phase 01 (browse API), Phase 02 (FilterBar component)

## Objective
Tạo trang `/browse` cho phép user navigate cấu trúc folder trong library.
DB-based (chỉ show media đã scan), library-scoped, ACL-aware.

## Hiện trạng
- `GET /api/admin/fs/browse?path=` — admin-only, reads filesystem trực tiếp, returns absolute paths
- `media_files.file_path` — absolute path trên disk, indexed
- `libraries.paths` — JSON array of root folders (e.g., `["/media/movies", "/nas/films"]`)
- `user_library_access` table — controls which users can access which libraries

## Security Constraints
1. **Library-relative paths only** — response NEVER contains absolute server paths
2. **ACL check** — non-admin users only browse libraries they have access to
3. **Path traversal blocked** — reject `..` in path param
4. **DB-only** — does NOT read filesystem, only queries `media_files` table
5. **Scanned media only** — folders without scanned media won't appear

## Implementation Steps

### Task 1: Browse hook

**File:** `webapp/src/hooks/useFolderBrowse.ts` — NEW

```typescript
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'

interface UseFolderBrowseParams {
  libraryId: number
  path: string
}

export function useFolderBrowse({ libraryId, path }: UseFolderBrowseParams) {
  return useQuery({
    queryKey: ['browse', libraryId, path],
    queryFn: () => {
      const params = new URLSearchParams()
      params.set('library_id', String(libraryId))
      if (path) params.set('path', path)
      return api.get<BrowseResult>(`/browse?${params}`)
    },
    enabled: libraryId > 0,
    staleTime: 60_000,
  })
}
```

State management via URL params:
- `?library_id=1&path=Action/Marvel` → synced with navigation
- Click folder → update `path` param
- Click breadcrumb → update to that path segment
- Select library → reset path to ""

- [ ] Create `useFolderBrowse` hook
- [ ] Add browse API function to api layer
- [ ] Query key includes libraryId + path for proper caching

### Task 2: Breadcrumb component

**File:** `webapp/src/components/Breadcrumb.tsx` — NEW

```typescript
interface BreadcrumbProps {
  path: string              // "Action/Marvel/MCU"
  libraryName: string       // "Movies"
  onNavigate: (path: string) => void
}

// Renders: Movies > Action > Marvel > MCU
//          ^click    ^click   ^click  (current, not clickable)
```

Parse path into segments:
```typescript
const segments = path ? path.split('/') : []
// segments = ["Action", "Marvel", "MCU"]
// breadcrumbs:
//   { label: libraryName, path: "" }         ← root
//   { label: "Action", path: "Action" }
//   { label: "Marvel", path: "Action/Marvel" }
//   { label: "MCU", path: "Action/Marvel/MCU" }  ← current (not clickable)
```

- [ ] Create Breadcrumb component
- [ ] Parse path into clickable segments
- [ ] Last segment = current (not clickable, bold)
- [ ] Root = library name

### Task 3: FolderCard component

**File:** `webapp/src/components/FolderCard.tsx` — NEW

```typescript
interface FolderCardProps {
  name: string
  mediaCount: number
  onClick: () => void
}
```

Design:
```
┌──────────────┐
│     📁       │
│   Action     │
│   (12 items) │
└──────────────┘
```

- Same card dimensions as MediaCard for grid consistency
- Folder icon (could use Lucide `Folder` icon)
- Name + media count
- Hover effect: border highlight

- [ ] Create FolderCard component
- [ ] Match MediaCard dimensions
- [ ] Hover + click interaction

### Task 4: BrowsePage

**File:** `webapp/src/pages/BrowsePage.tsx` — NEW

```typescript
export default function BrowsePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const libraryId = Number(searchParams.get('library_id')) || 0
  const path = searchParams.get('path') ?? ''

  const { data: libraries } = useLibraries()
  const { data: result, isLoading } = useFolderBrowse({ libraryId, path })

  // Auto-select first library if none selected
  useEffect(() => {
    if (!libraryId && libraries?.length) {
      setSearchParams({ library_id: String(libraries[0].id) }, { replace: true })
    }
  }, [libraries, libraryId])

  const navigate = (newPath: string) => {
    setSearchParams(prev => {
      if (newPath) {
        prev.set('path', newPath)
      } else {
        prev.delete('path')
      }
      return prev
    }, { replace: true })
  }

  const selectedLibrary = libraries?.find(l => l.id === libraryId)

  return (
    <div>
      {/* Header: title + library selector */}
      <div className="flex items-center justify-between">
        <h1>Browse Folders</h1>
        <select
          value={libraryId}
          onChange={e => setSearchParams({
            library_id: e.target.value,
          }, { replace: true })}
        >
          {libraries?.map(lib => (
            <option key={lib.id} value={lib.id}>{lib.name}</option>
          ))}
        </select>
      </div>

      {/* Breadcrumb */}
      <Breadcrumb
        path={path}
        libraryName={selectedLibrary?.name ?? 'Library'}
        onNavigate={navigate}
      />

      {/* Folders section */}
      {result?.folders.length > 0 && (
        <section>
          <h2>Folders ({result.folders.length})</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {result.folders.map(folder => (
              <FolderCard
                key={folder.path}
                name={folder.name}
                mediaCount={folder.media_count}
                onClick={() => navigate(folder.path)}
              />
            ))}
          </div>
        </section>
      )}

      {/* Media in this folder */}
      {result?.media.length > 0 && (
        <section>
          <h2>Media ({result.media.length})</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {result.media.map(item => (
              <MediaCard key={item.id} media={item} />
            ))}
          </div>
        </section>
      )}

      {/* Empty state */}
      {!isLoading && !result?.folders.length && !result?.media.length && (
        <EmptyState message="No media in this folder" />
      )}
    </div>
  )
}
```

- [ ] Create BrowsePage component
- [ ] Library selector dropdown
- [ ] Breadcrumb navigation
- [ ] Folder grid (FolderCard) + Media grid (MediaCard)
- [ ] URL state: ?library_id=&path=
- [ ] Auto-select first library
- [ ] Loading + empty states

### Task 5: Router integration

**File:** `webapp/src/providers/Router.tsx`

```typescript
// Add route
{ path: '/browse', element: <BrowsePage /> }
```

**File:** Navigation component (Sidebar or Header)

Add "Browse Folders" link:
```typescript
<NavLink to="/browse">
  <FolderIcon /> Browse
</NavLink>
```

- [ ] Add `/browse` route to Router
- [ ] Add navigation link (sidebar or header)

### Task 6: Polish + edge cases

- [ ] Back button: when path is not root, breadcrumb handles going up
- [ ] Large folders: if > 100 media items in one folder, show first 50 + "Show more"
- [ ] Library with multiple paths: for now, browse first path only (document limitation)
- [ ] Series in folder: if media_type is episode, show series instead of individual episode
- [ ] Keyboard navigation: Enter on folder card = navigate

## Files to Create/Modify

### Create
- `webapp/src/hooks/useFolderBrowse.ts` — browse hook
- `webapp/src/pages/BrowsePage.tsx` — browse page
- `webapp/src/components/Breadcrumb.tsx` — breadcrumb navigation
- `webapp/src/components/FolderCard.tsx` — folder card

### Modify
- `webapp/src/providers/Router.tsx` — add route
- `webapp/src/components/Sidebar.tsx` (or navigation) — add link

## UI Mockup

```
┌─────────────────────────────────────────────────────┐
│  📁 Browse Folders          [Library: Movies ▼]     │
├─────────────────────────────────────────────────────┤
│  Movies > Action                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  📁 Folders (3)                                     │
│  ┌──────┐  ┌──────┐  ┌──────┐                      │
│  │  📁  │  │  📁  │  │  📁  │                      │
│  │Marvel │  │DC    │  │Other │                      │
│  │ (12)  │  │ (8)  │  │ (25) │                      │
│  └──────┘  └──────┘  └──────┘                      │
│                                                     │
│  🎬 Media in this folder (5)                        │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐ │
│  │poster│  │poster│  │poster│  │poster│  │poster│  │
│  │Die   │  │John  │  │Speed │  │Heat  │  │Taken │  │
│  │Hard  │  │Wick  │  │      │  │      │  │      │  │
│  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘ │
└─────────────────────────────────────────────────────┘
```

## Verification

### Test Cases
- [ ] Select library → shows root folders (derived from scanned media paths)
- [ ] Click folder → URL updates to `?library_id=1&path=Action`
- [ ] Click subfolder → path deepens: `?path=Action/Marvel`
- [ ] Breadcrumb "Movies" → back to root (path="")
- [ ] Breadcrumb "Action" → back to `?path=Action`
- [ ] Media card click → navigate to detail page
- [ ] Folder with no media → "No media in this folder"
- [ ] Library with no scanned media → empty state
- [ ] Refresh page → same state (URL params preserved)
- [ ] Non-admin user → only sees accessible libraries in dropdown

### Edge Cases
- [ ] Very deep folder paths (5+ levels) → breadcrumb wraps / truncates
- [ ] Unicode folder names → display correctly
- [ ] Library with multiple root paths → browse first path (document in UI)
- [ ] Folder name with special characters → URL encoded properly

---
Previous Phase: [Phase 02 — Frontend Filter UI](phase-02-frontend-filter.md)
