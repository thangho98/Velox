# Plan N: Multi-language Support (i18n)

Support đa ngôn ngữ cho Velox UI, bắt đầu với English (default) và Vietnamese.

Created: 2026-03-22
Status: 🟡 Ready to code

---

## Overview

Thêm hệ thống đa ngôn ngữ (i18n) cho Velox, cho phép người dùng chuyển đổi giữa English và Vietnamese.

**Scope:**
- Frontend i18n infrastructure (React)
- Backend language preference storage
- Full Vietnamese translation for all UI strings
- Language switcher UI

---

## Tech Stack

| Component | Choice |
|-----------|--------|
| i18n Library | `react-i18next` + `i18next` |
| Language Detection | `i18next-browser-languagedetector` |
| Storage | localStorage + DB preference |

---

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Setup i18n Infrastructure | ⬜ Pending | 0% |
| 02 | Extract English Strings | ⬜ Pending | 0% |
| 03 | Vietnamese Translation | ⬜ Pending | 0% |
| 04 | Language Switcher & Persistence | ⬜ Pending | 0% |
| 05 | Polish & Testing | ⬜ Pending | 0% |

---

## Estimated Effort

- **Total:** 17-22 hours
- **Files created:** ~30 files
- **New packages:** 3 (react-i18next, i18next, i18next-browser-languagedetector)

---

## Quick Commands

- Start Phase 1: `/code phase-01`
- Check progress: `/next`

---

## Files to Create

```
webapp/src/
├── i18n/
│   └── index.ts              # i18n config
├── locales/
│   ├── en/
│   │   ├── common.json
│   │   ├── auth.json
│   │   ├── navigation.json
│   │   ├── settings.json
│   │   ├── media.json
│   │   ├── watch.json
│   │   └── errors.json
│   └── vi/
│       ├── common.json
│       ├── auth.json
│       ├── navigation.json
│       ├── settings.json
│       ├── media.json
│       ├── watch.json
│       └── errors.json
└── hooks/
    └── useTranslation.ts     # Typed wrapper
```

## Backend Migration

```sql
-- 022_language_preference.sql
ALTER TABLE user_preferences ADD COLUMN language TEXT DEFAULT 'en' CHECK (language IN ('en', 'vi'));
```

---

## Design Notes

- Nested keys: `t('settings.sections.profile.title')`
- TypeScript: Generate type definitions từ JSON files
- Netflix/Vietnamese style: Giữ tone giống Netflix (thân thiện, đơn giản)
- Fallback: 'en' cho strings chưa dịch
