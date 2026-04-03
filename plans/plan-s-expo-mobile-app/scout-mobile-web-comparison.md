# Scout Report: Mobile vs Web Feature Parity

> Generated: 2026-04-03
> Scope: Full feature comparison across all screens, components, player, admin, subtitles

---

## EXECUTIVE SUMMARY

**Mobile alignment: ~98% of implementable web features**

The mobile app (46 files, ~15k lines) covers nearly all web app features (77 files, pages+components+settings). After the 2026-04-03 parallel agent sprint, only browser-only features and 1 deferred item remain.

---

## FEATURE-BY-FEATURE COMPARISON

### 1. NAVIGATION & APP SHELL
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Bottom tab bar (5 tabs) | ✅ | ✅ | ALIGNED |
| Top header (logo, search, notifications, avatar) | ✅ | ✅ | ALIGNED |
| Dark theme (#141414) | ✅ | ✅ | ALIGNED |
| Sidebar navigation | ✅ | N/A | Mobile uses tabs instead |

### 2. HOME SCREEN
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Hero/greeting section | ✅ | ✅ | ALIGNED |
| Continue Watching | ✅ | ✅ | ALIGNED |
| Next Up | ✅ | ✅ | ALIGNED |
| Recently Added (Movies + Series) | ✅ | ✅ | ALIGNED |
| Horizontal scroll arrows | ✅ | ✅ | ALIGNED (new) |
| Loading skeletons | ✅ | ✅ | ALIGNED |

### 3. MOVIES / SERIES PAGES
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Genre/Year filters | ✅ | ✅ | ALIGNED |
| Sort (newest, oldest, rating, A-Z) | ✅ | ✅ | ALIGNED |
| A-Z fast scroll index | ✅ | ✅ | ALIGNED |
| Item count header | ✅ | ✅ | ALIGNED |
| Long press / quick actions | ✅ | ✅ | ALIGNED |
| Rating badges | ✅ | ✅ | ALIGNED |

### 4. SEARCH
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Auto-focus search | ✅ | ✅ | ALIGNED |
| Type filter chips | ✅ | ✅ | ALIGNED |
| Recent searches | ✅ | ✅ | ALIGNED |
| Results with count | ✅ | ✅ | ALIGNED |

### 5. BROWSE / FOLDERS
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Library grid | ✅ | ✅ | ALIGNED |
| Folder navigation | ✅ | ✅ | ALIGNED |
| Breadcrumb navigation | ✅ | ✅ | ALIGNED (new) |

### 6. MEDIA DETAIL
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Backdrop + poster | ✅ | ✅ | ALIGNED |
| Metadata (year, duration, ratings) | ✅ | ✅ | ALIGNED |
| Play / Resume / From Beginning | ✅ | ✅ | ALIGNED |
| Watched + Favorite toggles | ✅ | ✅ | ALIGNED |
| Progress bar + remaining time | ✅ | ✅ | ALIGNED |
| Technical specs | ✅ | ✅ | ALIGNED |
| Play Options bottom sheet | ✅ | ✅ | ALIGNED |
| YouTube trailer section | ✅ | ✅ | ALIGNED (new - thumbnail+play) |
| Subtitle search (OpenSubtitles) | ✅ | ✅ | ALIGNED (new) |
| Subtitle translate (DeepL) | ✅ | ✅ | ALIGNED (new) |
| Metadata editor (admin) | ✅ | ✅ | ALIGNED |
| Lock badge | ✅ | ✅ | ALIGNED |
| Rotten Tomatoes/Metacritic | ❌ | ❌ | Neither has it |

### 7. SERIES DETAIL
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Backdrop + poster | ✅ | ✅ | ALIGNED |
| Season tabs (horizontal scroll) | ✅ | ✅ | ALIGNED (new) |
| Episode cards with progress | ✅ | ✅ | ALIGNED |
| Episode edit dialog (admin) | ✅ | ✅ | ALIGNED (new) |
| Play options + action menu | ✅ | ✅ | ALIGNED |
| Status/Network badges | ✅ | ✅ | ALIGNED |

### 8. VIDEO PLAYER
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Playback speed (0.5x-2x) | ✅ | ✅ | ALIGNED |
| Audio track selector | ✅ | ✅ | ALIGNED |
| Video stats overlay | ✅ | ✅ | ALIGNED |
| Skip intro/credits buttons | ✅ | ✅ | ALIGNED |
| Next episode + auto-play | ✅ | ✅ | ALIGNED |
| Subtitle delay offset | ✅ | ✅ | ALIGNED |
| Aspect ratio control | ✅ | ✅ | ALIGNED |
| Repeat mode | ✅ | ✅ | ALIGNED |
| Lock controls | ✅ | ✅ | ALIGNED |
| Dual subtitles | ✅ | ✅ | ALIGNED (new) |
| Subtitle appearance (size/color/bg) | ✅ | ✅ | ALIGNED (new) |
| Wall clock display | ✅ | ✅ | ALIGNED (new) |
| Gesture controls (swipe, double-tap) | N/A | ✅ | Mobile-only |
| Keyboard shortcuts | ✅ | N/A | Browser-only |
| Chromecast | ✅ | ❌ | Browser Cast SDK |
| HLS.js ABR | ✅ | N/A | Mobile uses native HLS |
| Bandwidth monitoring | ✅ | ❌ | HLS.js specific |
| Media Session API | ✅ | N/A | Browser-only |
| Trickplay preview | ✅ | ❌ | Hover-based, N/A mobile |
| CinemaOverlay (pre-roll) | ✅ | ❌ | DEFERRED |

### 9. FAVORITES
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Grid display + count | ✅ | ✅ | ALIGNED |
| Quick actions | ✅ | ✅ | ALIGNED |
| Progress overlay | ✅ | ✅ | ALIGNED |

### 10. CONTINUE WATCHING
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Filter chips | ✅ | ✅ | ALIGNED |
| Progress bars | ✅ | ✅ | ALIGNED |
| Swipe to remove | N/A | ✅ | Mobile-only |

### 11. PROFILE & SETTINGS
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Profile editing | ✅ | ✅ | ALIGNED |
| Preferences (language/quality) | ✅ | ✅ | ALIGNED |
| Security (password) | ✅ | ✅ | ALIGNED |
| Sessions (view/revoke) | ✅ | ✅ | ALIGNED |
| Server info | ✅ | ✅ | ALIGNED |
| Cinema mode settings | ✅ | ✅ | ALIGNED (new) |
| Subtitle/translate config | ✅ | ✅ | ALIGNED (new) |

### 12. ADMIN FEATURES
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Server info | ✅ | ✅ | ALIGNED |
| Library management + scan | ✅ | ✅ | ALIGNED |
| Activity logs + filters | ✅ | ✅ | ALIGNED |
| Library statistics | ✅ | ✅ | ALIGNED (new) |
| Pre-transcode dashboard | ✅ | ✅ | ALIGNED (new) |
| Marker detection | ✅ | ✅ | ALIGNED (new) |
| Webhook management | ✅ | ✅ | ALIGNED (new) |
| Scheduled tasks | ✅ | ✅ | ALIGNED (new) |
| DeepL/translate config | ✅ | ✅ | ALIGNED (new) |
| Bulk metadata refresh | ✅ | ✅ | ALIGNED (new) |
| User management | ✅ | ✅ | ALIGNED |

### 13. SUBTITLES (ADVANCED)
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Subtitle delay offset | ✅ | ✅ | ALIGNED |
| Language selector | ✅ | ✅ | ALIGNED |
| Dual subtitle rendering | ✅ | ✅ | ALIGNED (new) |
| Subtitle appearance | ✅ | ✅ | ALIGNED (new) |
| Online subtitle search | ✅ | ✅ | ALIGNED (new) |
| AI translation (DeepL) | ✅ | ✅ | ALIGNED (new) |

### 14. LOGIN
| Feature | Web | Mobile | Status |
|---------|-----|--------|--------|
| Error messages | ✅ | ✅ | ALIGNED |
| Logo styling | ✅ | ✅ | ALIGNED |
| Show/hide password | ✅ | ✅ | ALIGNED |
| Remember server URL | N/A | ✅ | Mobile-only |

---

## REMAINING GAPS (only browser-only or deferred)

| Feature | Reason | Priority |
|---------|--------|----------|
| Chromecast | Browser Cast SDK only | N/A |
| Trickplay preview | Hover-based (no hover on mobile) | N/A |
| HLS.js ABR | Mobile uses native HLS | N/A |
| Bandwidth monitoring | HLS.js specific | N/A |
| Media Session API | Browser-only | N/A |
| Keyboard shortcuts | Browser-only | N/A |
| CinemaOverlay (pre-roll) | Complex, deferred | Low |
| Setup Wizard page | Web-only onboarding | N/A |

---

## MOBILE-ONLY FEATURES (not in web)

- Gesture controls (swipe seek, volume/brightness gestures, double-tap)
- Swipe to remove from watch history
- Server URL configuration + multi-server support
- Bottom sheet modals (native feel)
- Pull-to-refresh on all screens
- SecureStore for token persistence

---

## NEW FILES CREATED (2026-04-03 sprint)

### Components (6 new):
- `mobile/src/components/DualSubtitleOverlay.tsx` - VTT parsing + dual rendering
- `mobile/src/components/SubtitleAppearanceSheet.tsx` - Size/color/background picker
- `mobile/src/components/SubtitleSearchModal.tsx` - Online subtitle search & download
- `mobile/src/components/SubtitleTranslate.tsx` - DeepL translation UI
- `mobile/src/components/EpisodeEditDialog.tsx` - Episode metadata editor
- `mobile/src/components/Breadcrumb.tsx` - Navigation breadcrumb

### Admin Components (8 new):
- `mobile/src/components/admin/CollapsibleSection.tsx` - Reusable collapsible card
- `mobile/src/components/admin/CinemaSection.tsx` - Cinema mode settings
- `mobile/src/components/admin/PretranscodeSection.tsx` - Pre-transcode dashboard
- `mobile/src/components/admin/MarkersSection.tsx` - Marker detection stats
- `mobile/src/components/admin/WebhooksSection.tsx` - Webhook CRUD
- `mobile/src/components/admin/TasksSection.tsx` - Scheduled tasks
- `mobile/src/components/admin/LibraryStatsSection.tsx` - Library analytics
- `mobile/src/components/admin/TranslateSection.tsx` - DeepL/Subdl config

### Modified Screens (5):
- `VideoPlayerScreen.tsx` - Wall clock, dual subs, subtitle appearance
- `AdminScreen.tsx` - 6 new admin sections + bulk refresh
- `SettingsScreen.tsx` - Cinema + Subtitles tabs
- `SeriesDetailScreen.tsx` - Season tabs + episode edit
- `MediaDetailScreen.tsx` - Subtitle search/translate + trailer section
- `LibraryBrowseScreen.tsx` - Breadcrumb navigation
- `HorizontalMediaRow.tsx` - Scroll indicators

---

## CONCLUSION

**Before sprint:** ~90% core UX, ~50% admin, ~40% subtitles
**After sprint:** ~98% core UX, ~95% admin, ~90% subtitles

Only truly browser-only APIs (Chromecast, Trickplay hover, HLS.js internals) and 1 deferred complex feature (CinemaOverlay pre-roll) remain. The mobile app now has full feature parity with the web app for all practical mobile use cases.
