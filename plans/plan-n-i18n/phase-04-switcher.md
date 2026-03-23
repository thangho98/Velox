# Phase 04: Language Switcher & Persistence

Status: ⬜ Pending
Dependencies: Phase 03

---

## Objective

Tạo UI để chuyển đổi ngôn ngữ và lưu preference vào database.

---

## Requirements

### Functional
- [ ] Language switcher component (dropdown)
- [ ] Add `language` column to `user_preferences` table
- [ ] API endpoint to update language preference
- [ ] Load language from preference on app start
- [ ] Sync language change to backend

### Non-Functional
- [ ] Smooth transition (no page reload)
- [ ] Persist in localStorage (for guests)
- [ ] Persist in DB (for logged-in users)

---

## Implementation Steps

### Step 1: Database Migration
**File:** `backend/internal/database/migrate/022_language_preference.sql`

```sql
-- Migration: 022_language_preference
-- Adds language preference to user_preferences

ALTER TABLE user_preferences
ADD COLUMN language TEXT DEFAULT 'en'
CHECK (language IN ('en', 'vi'));

-- Update existing rows to default
UPDATE user_preferences SET language = 'en' WHERE language IS NULL;
```

**Update registry:** `backend/internal/database/migrate/registry.go`

```go
{
    Version: 22,
    Name:    "Add language preference",
    Up:      mustReadFile("022_language_preference.sql"),
    Down:    "ALTER TABLE user_preferences DROP COLUMN language;",
}
```

### Step 2: Backend Model Update
**File:** `backend/internal/model/user.go`

```go
type UserPreferences struct {
    // ... existing fields
    Language string `json:"language" db:"language"`
}
```

### Step 3: Repository Update
**File:** `backend/internal/repository/preferences.go`

Add to `UpdatePreferences` method:
```go
func (r *PreferencesRepo) UpdatePreferences(ctx context.Context, userID int64, prefs *model.UserPreferences) error {
    query := `
        UPDATE user_preferences
        SET subtitle_language = ?, audio_language = ?, max_streaming_quality = ?, theme = ?, language = ?
        WHERE user_id = ?
    `
    result, err := r.db.ExecContext(ctx, query,
        prefs.SubtitleLanguage,
        prefs.AudioLanguage,
        prefs.MaxStreamingQuality,
        prefs.Theme,
        prefs.Language, // NEW
        userID,
    )
    // ... error handling
}
```

### Step 4: API Endpoint
**File:** `backend/internal/handler/user.go`

Update PATCH /api/users/preferences to accept language:
```go
type UpdatePreferencesRequest struct {
    // ... existing fields
    Language string `json:"language,omitempty"`
}
```

### Step 5: Language Switcher Component
**File:** `webapp/src/components/LanguageSwitcher.tsx`

```tsx
import { useTranslation } from 'react-i18next'
import { usePreferences, useUpdatePreferences } from '@/hooks/stores/useAuth'
import { LuGlobe, LuCheck } from 'react-icons/lu'

const LANGUAGES = [
  { code: 'en', label: 'English', flag: '🇺🇸' },
  { code: 'vi', label: 'Tiếng Việt', flag: '🇻🇳' },
]

export function LanguageSwitcher() {
  const { i18n } = useTranslation()
  const { data: preferences } = usePreferences()
  const { mutate: updatePreferences } = useUpdatePreferences()

  const currentLang = i18n.language

  const handleChange = (lang: string) => {
    // Update i18n
    i18n.changeLanguage(lang)

    // Persist to localStorage
    localStorage.setItem('velox-language', lang)

    // Sync to backend if logged in
    if (preferences) {
      updatePreferences({ language: lang })
    }
  }

  return (
    <div className="relative">
      <select
        value={currentLang}
        onChange={(e) => handleChange(e.target.value)}
        className="rounded bg-netflix-gray px-3 py-2 text-sm text-white"
      >
        {LANGUAGES.map((lang) => (
          <option key={lang.code} value={lang.code}>
            {lang.flag} {lang.label}
          </option>
        ))}
      </select>
    </div>
  )
}
```

### Step 6: Add to Settings Page
**File:** `webapp/src/pages/SettingsPage.tsx`

Add to PreferencesSection:
```tsx
import { LanguageSwitcher } from '@/components/LanguageSwitcher'

// In PreferencesSection form:
<Field label={t('settings.fields.language')}>
  <LanguageSwitcher />
</Field>
```

Or add to Navbar user menu:
```tsx
// In Navbar.tsx user menu dropdown
<div className="border-t border-netflix-gray">
  <div className="px-4 py-2 text-xs text-gray-500">
    Language
  </div>
  <LanguageSwitcher compact />
</div>
```

### Step 7: Load on App Start
**File:** `webapp/src/providers/Providers.tsx`

```tsx
import { useEffect } from 'react'
import { usePreferences } from '@/hooks/stores/useAuth'
import { useTranslation } from 'react-i18next'

function LanguageProvider({ children }: { children: React.ReactNode }) {
  const { data: preferences, isLoading } = usePreferences()
  const { i18n } = useTranslation()

  useEffect(() => {
    if (!isLoading && preferences?.language) {
      // Only change if different from current
      if (i18n.language !== preferences.language) {
        i18n.changeLanguage(preferences.language)
        localStorage.setItem('velox-language', preferences.language)
      }
    }
  }, [preferences, isLoading, i18n])

  return <>{children}</>
}
```

### Step 8: Update i18n Config
**File:** `webapp/src/i18n/index.ts`

Ensure detection order:
```typescript
detection: {
  order: ['localStorage', 'navigator', 'htmlTag'],
  lookupLocalStorage: 'velox-language',
  caches: ['localStorage'],
}
```

---

## API Changes

### GET /api/users/preferences
Response:
```json
{
  "data": {
    "user_id": 1,
    "subtitle_language": "vi",
    "audio_language": "vi",
    "max_streaming_quality": "original",
    "theme": "dark",
    "language": "vi"  // NEW
  }
}
```

### PATCH /api/users/preferences
Request:
```json
{
  "language": "vi"
}
```

---

## Test Criteria

- [ ] Migration runs successfully
- [ ] Language switcher appears in UI
- [ ] Switch language updates UI immediately
- [ ] Preference saved to localStorage
- [ ] Preference saved to backend (for logged-in users)
- [ ] On refresh, language persists
- [ ] On login, language loaded from user preferences

---

## Files to Modify

| File | Change |
|------|--------|
| `backend/internal/database/migrate/022_language_preference.sql` | Create migration |
| `backend/internal/database/migrate/registry.go` | Register migration |
| `backend/internal/model/user.go` | Add Language field |
| `backend/internal/repository/preferences.go` | Update queries |
| `backend/internal/handler/user.go` | Accept language in API |
| `webapp/src/components/LanguageSwitcher.tsx` | Create component |
| `webapp/src/pages/SettingsPage.tsx` | Add switcher |
| `webapp/src/hooks/stores/useAuth.ts` | Update types |

---

## Notes

- localStorage key: `velox-language`
- DB default: `'en'`
- Valid values: `'en'`, `'vi'`
- Sync: localStorage (immediate) + DB (async)

---

Next Phase: [Phase 05 - Polish & Testing](./phase-05-polish.md)
