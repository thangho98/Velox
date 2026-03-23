# Phase 01: Setup i18n Infrastructure

Status: ⬜ Pending
Dependencies: None

---

## Objective

Setup infrastructure đa ngôn ngữ cho Velox frontend sử dụng react-i18next.

---

## Requirements

### Functional
- [ ] Install dependencies: react-i18next, i18next, i18next-browser-languagedetector
- [ ] Create i18n configuration file
- [ ] Create translation file structure
- [ ] Integrate i18n vào App
- [ ] Create typed useTranslation hook wrapper

### Non-Functional
- [ ] TypeScript type-safe translations
- [ ] Auto-detect language from browser/localStorage
- [ ] Fallback to 'en' for missing translations

---

## Implementation Steps

### Step 1: Install Dependencies
```bash
cd webapp
npm install react-i18next i18next i18next-browser-languagedetector
```

### Step 2: Create i18n Configuration
**File:** `webapp/src/i18n/index.ts`

```typescript
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

import enCommon from '../locales/en/common.json'
import enAuth from '../locales/en/auth.json'
import enNavigation from '../locales/en/navigation.json'
import enSettings from '../locales/en/settings.json'
import enMedia from '../locales/en/media.json'
import enWatch from '../locales/en/watch.json'
import enErrors from '../locales/en/errors.json'

import viCommon from '../locales/vi/common.json'
import viAuth from '../locales/vi/auth.json'
import viNavigation from '../locales/vi/navigation.json'
import viSettings from '../locales/vi/settings.json'
import viMedia from '../locales/vi/media.json'
import viWatch from '../locales/vi/watch.json'
import viErrors from '../locales/vi/errors.json'

const resources = {
  en: {
    common: enCommon,
    auth: enAuth,
    navigation: enNavigation,
    settings: enSettings,
    media: enMedia,
    watch: enWatch,
    errors: enErrors,
  },
  vi: {
    common: viCommon,
    auth: viAuth,
    navigation: viNavigation,
    settings: viSettings,
    media: viMedia,
    watch: viWatch,
    errors: viErrors,
  },
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    defaultNS: 'common',
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'velox-language',
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false, // React already escapes
    },
  })

export default i18n
```

### Step 3: Create Folder Structure
```
webapp/src/locales/
├── en/
│   ├── common.json
│   ├── auth.json
│   ├── navigation.json
│   ├── settings.json
│   ├── media.json
│   ├── watch.json
│   └── errors.json
└── vi/
    ├── common.json
    ├── auth.json
    ├── navigation.json
    ├── settings.json
    ├── media.json
    ├── watch.json
    └── errors.json
```

Create placeholder JSON files:
```json
{
  "placeholder": "Placeholder"
}
```

### Step 4: Integrate vào App
**File:** `webapp/src/main.tsx`

Add import BEFORE React render:
```typescript
import './i18n' // Import i18n configuration
import { StrictMode } from 'react'
// ... rest of imports
```

### Step 5: Create useTranslation Hook
**File:** `webapp/src/hooks/useTranslation.ts`

```typescript
import { useTranslation as useI18nTranslation } from 'react-i18next'

// Define namespaces for type safety
export type Namespace =
  | 'common'
  | 'auth'
  | 'navigation'
  | 'settings'
  | 'media'
  | 'watch'
  | 'errors'

export function useTranslation(ns: Namespace = 'common') {
  return useI18nTranslation(ns)
}
```

---

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `webapp/src/i18n/index.ts` | Create | i18n configuration |
| `webapp/src/locales/en/*.json` | Create | English translations (placeholder) |
| `webapp/src/locales/vi/*.json` | Create | Vietnamese translations (placeholder) |
| `webapp/src/hooks/useTranslation.ts` | Create | Typed wrapper hook |
| `webapp/src/main.tsx` | Modify | Import i18n config |

---

## Test Criteria

- [ ] App runs without errors (`npm run dev`)
- [ ] i18n initialized in console
- [ ] Language auto-detected from browser
- [ ] localStorage stores 'velox-language' key
- [ ] TypeScript knows translation keys (no errors)

---

## Notes

- Placeholder JSON files will be filled in Phase 02 (English) and Phase 03 (Vietnamese)
- Default language: English
- Supported languages: English ('en'), Vietnamese ('vi')

---

Next Phase: [Phase 02 - Extract English Strings](./phase-02-english-strings.md)
