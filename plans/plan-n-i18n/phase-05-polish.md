# Phase 05: Polish & Testing

Status: ⬜ Pending
Dependencies: Phase 04

---

## Objective

Polish & testing cho i18n feature. Xử lý edge cases, RTL (future-proof), và testing.

---

## Requirements

### Functional
- [ ] Test language switch in Settings
- [ ] Test refresh page - language persists
- [ ] Test API error messages display correctly
- [ ] RTL markup support (dir="auto")
- [ ] Number/date/time formatting (Intl)
- [ ] Pluralization support

### Non-Functional
- [ ] No console warnings
- [ ] Smooth language transition
- [ ] All edge cases handled

---

## Implementation Steps

### Step 1: RTL Considerations (Future-proof)
**File:** `webapp/src/providers/Providers.tsx`

Add RTL support:
```tsx
function RTLProvider({ children }: { children: React.ReactNode }) {
  const { i18n } = useTranslation()

  // Set dir attribute for RTL languages (future: Arabic, Hebrew)
  useEffect(() => {
    const rtlLanguages = ['ar', 'he']
    const dir = rtlLanguages.includes(i18n.language) ? 'rtl' : 'ltr'
    document.documentElement.dir = dir
    document.documentElement.lang = i18n.language
  }, [i18n.language])

  return <>{children}</>
}
```

### Step 2: Date/Time Formatting
**File:** `webapp/src/lib/i18n.ts`

```tsx
export function formatDate(date: string | Date, locale: string = 'en'): string {
  const d = new Date(date)
  return d.toLocaleDateString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatTime(date: string | Date, locale: string = 'en'): string {
  const d = new Date(date)
  return d.toLocaleTimeString(locale, {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatDateTime(date: string | Date, locale: string = 'en'): string {
  return `${formatDate(date, locale)} ${formatTime(date, locale)}`
}

export function formatBytes(bytes: number, locale: string = 'en'): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const value = bytes / Math.pow(1024, i)
  return `${value.toLocaleString(locale, { maximumFractionDigits: i > 1 ? 2 : 0 })} ${units[i]}`
}
```

### Step 3: Number Formatting
**File:** `webapp/src/lib/i18n.ts`

```tsx
export function formatNumber(num: number, locale: string = 'en'): string {
  return num.toLocaleString(locale)
}

export function formatRating(rating: number, locale: string = 'en'): string {
  return rating.toLocaleString(locale, { minimumFractionDigits: 1, maximumFractionDigits: 1 })
}
```

### Step 4: Pluralization
**File:** `webapp/src/locales/en/common.json`

```json
{
  "items": {
    "one": "{{count}} item",
    "other": "{{count}} items"
  },
  "episodes": {
    "one": "{{count}} episode",
    "other": "{{count}} episodes"
  }
}
```

**File:** `webapp/src/locales/vi/common.json`

```json
{
  "items": "{{count}} mục",
  "episodes": "{{count}} tập"
}
```

**Usage:**
```tsx
const { t } = useTranslation('common')
t('items', { count: 1 })   // "1 item" / "1 mục"
t('items', { count: 5 })   // "5 items" / "5 mục"
```

### Step 5: Test Cases

#### Test 1: Language Switch
```
Steps:
1. Login vào Velox
2. Vào Settings → Preferences
3. Thay đổi Language sang Vietnamese
4. Kiểm tra:
   - UI chuyển sang tiếng Việt
   - localStorage có 'velox-language' = 'vi'
   - API call PATCH /api/users/preferences với { language: 'vi' }
```

#### Test 2: Persistence
```
Steps:
1. Đổi language sang Vietnamese
2. Refresh trang (F5)
3. Kiểm tra:
   - Language vẫn là Vietnamese
   - Không cần đăng nhập lại
```

#### Test 3: Re-login
```
Steps:
1. Đổi language sang Vietnamese
2. Logout
3. Login lại
4. Kiểm tra:
   - Language vẫn là Vietnamese (từ DB preferences)
```

#### Test 4: Fallback
```
Steps:
1. Xóa key 'items' trong vi/common.json
2. Refresh trang
3. Kiểm tra:
   - Hiển thị English (fallback)
   - Không crash
```

#### Test 5: Error Messages
```
Steps:
1. Tắt backend
2. Thực hiện action (save settings)
3. Kiểm tra:
   - Error message hiển thị đúng ngôn ngữ
   - Có thể dùng i18n cho error messages từ backend
```

### Step 6: Edge Cases

#### Edge Case 1: Invalid Language
**Handling:**
```tsx
// In i18n config
fallbackLng: 'en',
```

#### Edge Case 2: Browser Language
**Handling:**
```typescript
detection: {
  order: ['localStorage', 'navigator', 'htmlTag'],
  // If browser is 'zh-CN', fallback to 'en' (not supported)
}
```

#### Edge Case 3: Guest User (not logged in)
**Handling:**
- localStorage only
- No DB sync
- On login: overwrite with user preference

#### Edge Case 4: Concurrent Users
**Handling:**
- Each browser/tab has own localStorage
- No conflict (server-side is per-user)

### Step 7: Accessibility

#### Screen Reader
```tsx
// Add aria-label for language switcher
<select
  aria-label={t('settings.fields.language')}
  // ...
>
```

#### Keyboard Navigation
- Tab order: Language switcher focusable
- Enter/Space: Open dropdown
- Arrow keys: Navigate options

### Step 8: Performance

#### Lazy Loading (Future)
```typescript
// Instead of loading all languages upfront
import i18n from 'i18next'

i18n.loadNamespaces('settings').then(() => {
  // Namespace loaded
})
```

#### Bundle Size
- Current: ~50KB per language (7 JSON files)
- Acceptable for MVP
- Future: Split by namespace if needed

---

## Test Checklist

### Unit Tests
- [ ] useTranslation hook works
- [ ] LanguageSwitcher renders correctly
- [ ] formatDate/formatBytes work correctly

### Integration Tests
- [ ] Language switch updates UI
- [ ] Preference syncs to backend
- [ ] localStorage persists correctly

### E2E Tests
- [ ] Full flow: Switch → Refresh → Re-login
- [ ] Edge cases: Invalid language, guest user
- [ ] Accessibility: Keyboard navigation, screen reader

---

## Final Checklist

### Code Quality
- [ ] All console warnings fixed
- [ ] TypeScript strict mode passes
- [ ] ESLint passes
- [ ] No hardcoded strings

### Documentation
- [ ] i18n setup documented
- [ ] Translation guidelines (for future translators)
- [ ] README updated

### Deployment
- [ ] Migration tested
- [ ] Build passes
- [ ] Docker image works

---

## Translation Guidelines (Future Reference)

### Adding New Language
1. Copy `locales/en/` → `locales/{lang}/`
2. Translate all JSON files
3. Add to i18n config resources
4. Update LanguageSwitcher options
5. Update DB constraint

### Adding New String
1. Add to `locales/en/{namespace}.json`
2. Add to `locales/vi/{namespace}.json`
3. Use in component
4. Test both languages

### Key Naming Convention
```
# Good
"settings.fields.username"
"media.actions.play"
"errors.notFound"

# Bad
"settings_username"
"playButton"
"not_found_error"
```

---

## Notes

- Keep translations organized by namespace
- Use nested keys for grouping
- Test with real users
- Update translations when UI changes

---

## Done! 🎉

Plan N complete. Velox now supports English and Vietnamese.
