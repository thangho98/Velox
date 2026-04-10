# Velox Mobile & Tablet UI Specifications
# Updated: 2026-04-04

---

## 1. Login Screen (`/login`)

### Layout Structure
```
┌──────────────────────────────────────┐
│           [VELOX logo]               │  ← Header (logo only)
│                                      │
│           [Sign In title]            │  ← h1, large text
│                                      │
│     ┌────────────────────────┐       │
│     │  Username              │       │  ← TextInput, outlined
│     └────────────────────────┘       │
│     ┌────────────────────────┐       │
│     │  Password              │       │  ← TextInput, outlined, secure
│     └────────────────────────┘       │
│     ┌────────────────────────┐       │
│     │       Sign In          │       │  ← Primary Button, full width
│     └────────────────────────┘       │
│                                      │
│   Questions? Contact your server         │  ← Footer text
│                                      │
│   Privacy | Terms of Service | Help  │  ← Footer links
└──────────────────────────────────────┘
```

### Elements Detail
| Element | Type | Behavior |
|---------|------|----------|
| VELOX logo | Text link | Navigates to `/` (homepage route, but not logged in → login) |
| "Sign In" | Heading1 | Static text |
| Username | TextInput | `name="username"`, placeholder="Username", text auto-capitalize |
| Password | TextInput | `name="password"`, placeholder="Password", secureTextEntry, show/hide toggle |
| Sign In button | Button | Primary filled, full width, triggers auth API call |
| "Questions? Contact your server administrator." | Paragraph | Static text, muted color |
| Privacy, Terms of Service, Help Center | Links | Static footer links |

### Mobile (375x667)
- Vertical centering with padding 24px horizontal
- Logo ~48px height
- Input fields: height 48px, border-radius 8px
- Button: height 48px, border-radius 8px, filled primary color

### Tablet (768x1024)
- Centered card/container max-width 400px
- More vertical spacing between elements
- Same proportions otherwise

### API Integration
- `POST /auth/login` with `{ username, password }` → `{ access_token, refresh_token, user, expires_in }`
- On success: redirect to `/`
- On failure: show inline error message below form

---

## 2. Home Screen (`/`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  VELOX    [🔔 Notifications] [🔍 Search]  [A]  │  ← Header
├──────────────────────────────────────┤
│  [Home] [Movies] [Series] [Browse] [Favorites] │  ← Bottom Tab Nav
├──────────────────────────────────────┤
│                                      │
│  Welcome back, Admin                 │  ← h1
│  Your personal media server.         │  ← Subtitle
│                                      │
│  [🎬 Movies]    [📺 Series]       │  ← Quick access link buttons
│                                      │
│  ── Continue Watching ────── See all │  │  ← Section header
│  ┌────┬────┬────┬────┬────┬────┐    │  ← Horizontal scroll
│  │ ▶ │ ▶ │ ▶ │ ▶ │ ▶ │ ▶ │... │    │  │  Media cards with progress
│  └────┴────┴────┴────┴────┴────┘    │
│                                      │
│  ── Next Up ───────────────── See all │  │  ← Section header
│  ┌────────┐  ┌────────┐              │
│  │        │  │        │              │  ← Next up card(s)
│  │ S1E19  │  │ S1E2   │              │
│  └────────┘  └────────┘              │
│                                      │
│  ── Movies ───────────────── See all  │  │
│  ┌────┬────┬────┬────┐              │
│  │    │    │    │    │              │
│  └────┴────┴────┴────┘              │
│                                      │
│  ── Series ───────────────── See all  │  │
│  ┌────┬────┬────┬────┐              │
│  │    │    │    │    │              │
│  └────┴────┴────┴────┘              │
└──────────────────────────────────────┘
```

### Elements Detail

#### Header
| Element | Type | Behavior |
|---------|------|----------|
| VELOX | Text link | Navigate to `/` |
| Notifications | Button with badge | Shows 9+ badge when >9 unread, opens notification panel |
| Search | Button | Opens search modal/overlay |
| User avatar [A] | Button | Opens user menu (settings, logout) |

#### Bottom Tab Navigation
| Tab | Icon | Route |
|-----|------|-------|
| Home | 🏠 house icon | `/` |
| Movies | 🎬 film icon | `/movies` |
| Series | 📺 TV icon | `/series` |
| Browse | 📁 folder icon | `/browse` |
| Favorites | ❤️ heart icon | `/favorites` |

#### Continue Watching Section
- Horizontal ScrollView with snap
- Card: poster image + gradient overlay at bottom
- Card shows: title, episode info (S1E1 · SeriesName), progress bar
- Progress bar: green fill, shows "Xm remaining" or "-Xm remaining" (negative = behind)
- "Dismiss" X button on each card → removes from continue watching

#### Next Up Section
- 2-column grid or horizontal scroll of larger cards
- Shows next unwatched episode based on progress
- Tap → navigate to `/watch/{episodeId}`

#### Movies / Series Section
- Horizontal link buttons with icons ("Movies", "Series") that navigate to respective browse pages
- OR grid of poster cards (2-column on mobile, 4-column on tablet)
- Shows: poster image, rating badge (e.g., "7.3"), type badge "Movie"/"Series", year

### Mobile (375x667)
- Bottom tab bar height: 64px + safe area
- Header height: 56px
- Card width: ~130px (Continue Watching), ~140px (Next Up)
- Grid: 2 columns for Next Up, 2-3 for Movies/Series rows

### Tablet (768x1024)
- Header height: 64px
- Card width: ~160px (Continue Watching)
- Grid: 3-4 columns for Movies/Series rows
- More horizontal padding

### API Integration
- `GET /profile/continue-watching?limit=20` → Continue Watching list
- `GET /profile/next-up?limit=20` → Next Up list
- `GET /media?type=movie&limit=10` → Movies row
- `GET /series?limit=10` → Series row

---

## 3. Movies Screen (`/movies`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  VELOX    [🔔] [🔍]  [A]           │  ← Header
├──────────────────────────────────────┤
│  [Home] [Movies] [Series] [Browse]  │  ← Bottom Tab Nav
├──────────────────────────────────────┤
│                                      │
│  Movies                              │  ← h1 title
│                                      │
│  [🎭 Genre ▼] [📆 Year ▼] [⭐ Rating ▼] [🔄 Sort ▼]  │  ← Filter chips
│                                      │
│  ┌────┬────┬────┬────┬────┐        │
│  │    │    │    │    │    │        │  ← Media grid
│  │ 7.3│    │    │    │    │        │  ← Rating badges on cards
│  │    │    │    │    │    │        │
│  ├────┼────┼────┼────┼────┤        │
│  │    │    │    │    │    │        │
│  │    │    │    │    │    │        │
│  └────┴────┴────┴────┴────┘        │
│                                      │
└──────────────────────────────────────┘
```

### Elements Detail

#### Filter Chips (Horizontal ScrollView)
| Filter | Type | Options |
|--------|------|---------|
| Genre | Dropdown | All, Action, Comedy, Drama, Horror, Sci-Fi, etc. |
| Year | Dropdown | All, 2020s, 2010s, 2000s, 1990s, etc. |
| Rating | Dropdown | All, 8+, 7+, 6+, 5+ |
| Sort | Dropdown | Recently Added, Title A-Z, Title Z-A, Rating, Year |

#### Media Card
- Poster image (aspect ratio 2:3)
- Overlay gradient at bottom
- Rating badge (bottom-left of poster): "7.3" format
- Type badge: "Movie" / "Series" (shown on card overlay or below)
- Title below card: movie name
- Year below title: "2025" format

### Mobile (375x667)
- 2-column grid
- Card width: ~160px
- Filter chips: horizontally scrollable row

### Tablet (768x1024)
- 4-column grid
- Card width: ~170px
- Filter chips: full row visible

### API Integration
- `GET /media?type=movie&genre={genre}&year={year}&sort={sort}&limit={limit}&offset={offset}`

---

## 4. Series Screen (`/series`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  VELOX    [🔔] [🔍]  [A]           │  ← Header
├──────────────────────────────────────┤
│  [Home] [Movies] [Series] [Browse]  │  ← Bottom Tab Nav
├──────────────────────────────────────┤
│                                      │
│  Series                              │  ← h1 title
│                                      │
│  [🎭 Genre ▼] [📆 Year ▼] [⭐ Rating ▼] [🔄 Sort ▼]  │  ← Filter chips
│                                      │
│  ┌────┬────┬────┬────┬────┐        │
│  │    │    │    │    │    │        │
│  │    │    │    │    │    │        │
│  │1994│    │    │    │    │        │  ← Year below card
│  └────┴────┴────┴────┴────┘        │
│                                      │
└──────────────────────────────────────┘
```

### Media Card (Series)
- Poster image (aspect ratio 2:3)
- Type badge: "Series" (shown on hover/long press)
- Title below card: series name
- Year below title: "1994" format

### API Integration
- `GET /series?genre={genre}&year={year}&sort={sort}&limit={limit}&offset={offset}`

---

## 5. Browse Screen (`/browse`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  VELOX    [🔔] [🔍]  [A]           │  ← Header
├──────────────────────────────────────┤
│  [Home] [Movies] [Series] [Browse]  │  ← Bottom Tab Nav
├──────────────────────────────────────┤
│                                      │
│  Browse                              │  ← h1 title
│                                      │
│  ┌──────────────────────────────┐   │
│  │ 📁                          │   │
│  │ Movies Library               │   │  ← Library folder card
│  │ 24 items                    │   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ 📁                          │   │
│  │ TV Shows Library            │   │  ← Library folder card
│  │ 3 items                     │   │
│  └──────────────────────────────┘   │
│                                      │
└──────────────────────────────────────┘
```

### Elements Detail
- Grid of library folder cards
- Each card shows: folder icon, library name, item count
- Tap → navigate to `/browse?library_id={id}&path=/` (root of that library)

### API Integration
- `GET /browse?library_id={id}&path={path}` → `{ folders: [], media: [] }`
- `GET /libraries` → list all libraries

---

## 6. Favorites Screen (`/favorites`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  VELOX    [🔔] [🔍]  [A]           │  ← Header
├──────────────────────────────────────┤
│  [Home] [Movies] [Series] [Browse]  │  ← Bottom Tab Nav
├──────────────────────────────────────┤
│                                      │
│  Favorites                           │  ← h1 title
│                                      │
│  (Empty state - heart icon)          │
│  "No favorites yet"                 │
│  "Start adding movies and series"    │
│                                      │
└──────────────────────────────────────┘
```

### Empty State
- Centered heart outline icon
- "No favorites yet" title
- "Start adding movies and series to see them here" subtitle

### API Integration
- `GET /profile/favorites?limit=20&offset=0` → list of favorited media
- `POST /profile/favorites/{mediaId}` → toggle favorite

---

## 7. Movie Detail Screen (`/movies/:id`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  ← Back                   [A]        │  ← Back header
├──────────────────────────────────────┤
│  ┌────────────────────────────────┐  │
│  │                                │  │
│  │     Backdrop Image             │  │  ← Full-width backdrop
│  │     (YouTube trailer if avail) │  │
│  │                                │  │
│  │  ▶ Play  [+ Favorite] [⋮ More] │  │  ← Action buttons overlay
│  └────────────────────────────────┘  │
│                                      │
│  Avatar: Fire and Ash               │  ← h1 Title
│  2025  │  7.3  │  3h 12m            │  ← Metadata row
│  [Action] [Sci-Fi] [Adventure]      │  ← Genre tags
│                                      │
│  ┌──────────────────────────────┐   │
│  │ Resume                       │   │  ← Resume button (if started)
│  │ "Continue from 1h 45m"       │   │
│  └──────────────────────────────┘   │
│                                      │
│  Overview                            │  ← Section header
│  The story of the Sully family...   │  ← Description text
│                                      │
│  Cast & Crew                        │  ← Section header
│  ┌────┬────┬────┬────┐              │
│  │ 👤 │ 👤 │ 👤 │ 👤 │  →          │  ← Cast horizontal scroll
│  │Name│Name│Name│Name│              │
│  └────┴────┴────┴────┘              │
│                                      │
│  Similar                            │  ← Section header
│  ┌────┬────┬────┐                   │
│  │    │    │    │                   │
│  └────┴────┴────┘                   │
│                                      │
└──────────────────────────────────────┘
```

### Action Buttons
| Button | Icon | Behavior |
|--------|------|----------|
| Play | ▶ | Navigate to `/watch/{id}` |
| Favorite | ♡ / ❤ | Toggle favorite, optimistic update |
| More | ⋮ | Opens action menu (Download, Stream URL, Edit Metadata) |

### Metadata Row
- Year: "2025"
- Rating: "7.3" with star icon
- Duration: "3h 12m" or "191m"

### Resume Button
- Shown if user has watch progress
- Shows "Continue from Xh Xm" or "Resume"
- Tap → navigate to `/watch/{id}?position={position}`

### Cast Card
- Circular avatar image
- Name below
- Character name (if cast)
- Horizontal scroll with "→" indicator

### API Integration
- `GET /media/{id}` → movie details
- `GET /media/{id}/files` → media with files (for playback)
- `GET /media/{id}/genres` → genres
- `GET /media/{id}/credits` → cast & crew
- `GET /media/{id}/cinema` → trailers (YouTube)
- `POST /media/{id}/refresh` → refresh metadata from TMDb
- `PATCH /media/{id}/metadata` → edit metadata (title, overview, genres, etc.)
- `POST /media/{id}/images` → upload poster/backdrop image
- `GET /search?q={query}&limit=5` → similar movies

---

## 8. Series Detail Screen (`/series/:id`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  ← Back                   [A]        │  ← Back header
├──────────────────────────────────────┤
│  ┌────────────────────────────────┐  │
│  │                                │  │
│  │     Backdrop Image             │  │
│  │     (YouTube trailer if avail) │  │
│  │                                │  │
│  │  ▶ Resume  [+ Favorite] [⋮]   │  │
│  └────────────────────────────────┘  │
│                                      │
│  Friends                             │  ← h1 Title
│  1994  │  Ended  │  NBC             │  ← Metadata row
│  10 seasons                         │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ Resume "Continue Friends -   │   │  ← Link card (not button)
│  │ The One With The Princess..." │   │
│  └──────────────────────────────┘   │
│                                      │
│  Overview                            │
│  Six young people from New York...   │
│                                      │
│  Episodes                            │
│  [Season 1 ▼] [Season 2] ...        │  ← Season selector
│  ┌──────────────────────────────┐   │
│  │ E1  │ Pilot                   │   │
│  │     │ An introduction to...    │   │
│  │     │              ▶ Play     │   │
│  ├──────────────────────────────┤   │
│  │ E2  │ The One with the...     │   │
│  │     │ Ross's lesbian ex-wife..│   │
│  │     │              ▶ Play     │   │
│  └──────────────────────────────┘   │
│                                      │
└──────────────────────────────────────┘
```

### Season Selector
- Horizontal scrollable chips
- Active season: filled background
- Tap → filter episode list

### Episode List Item
| Part | Content |
|------|---------|
| Left | Episode thumbnail (small) |
| Middle-top | Episode number badge "1", "2" + Episode title (heading) |
| Middle-bottom | Episode description (2-line clamp) |
| Right | Play button + More button (⋮) |

### Progress Indicator
- If watched: checkmark or progress bar
- If in progress: "Xm remaining" badge

### API Integration
- `GET /series/{id}` → series details
- `GET /series/{id}/seasons` → list seasons
- `GET /series/{id}/seasons/{seasonId}/episodes` → episodes by season
- `GET /series/{id}/genres` → genres
- `GET /series/{id}/credits` → cast & crew
- `GET /series/{id}/cinema` → trailers (YouTube)
- `GET /profile/progress/{mediaId}` → watch progress per episode

---

## 9. Watch Screen (`/watch/:id`)

### Layout Structure
```
┌──────────────────────────────────────┐
│  ←  [Title]           [CC] [🎬] [⋮]  │  ← Top bar (auto-hide)
├──────────────────────────────────────┤
│                                      │
│                                      │
│         ┌──────────────┐             │
│         │              │             │
│         │    Video     │             │  ← Video player (full screen)
│         │   Player     │             │
│         │              │             │
│         │   00:45:30   │             │  ← Current time
│         │ ────●────────│             │  ← Seek bar
│         │   03:12:00   │             │  ← Duration
│         └──────────────┘             │
│                                      │
│    ◁◁  ▐▐  ▷▷      🔊━━━━━━━━━     │  ← Control bar
│                                      │
│  [Primary Subtitle Line 1]           │  ← Subtitle overlay
│  [Secondary Subtitle Line 2]         │  │
│                                      │
└──────────────────────────────────────┘
```

### Video Player Controls (Overlay, auto-hide after 3s)
| Control | Position | Behavior |
|---------|----------|----------|
| Back arrow | Top-left | Navigate back |
| Title | Top-center | Media title |
| CC (subtitles) | Top-right | Toggle subtitle visibility |
| AirPlay | Top-right | Cast to device |
| More menu | Top-right | Audio track, quality, playback speed |
| Play/Pause | Center | Toggle playback |
| Skip backward 10s | Left of play | `currentTime -= 10` |
| Skip forward 30s | Right of play | `currentTime += 30` |
| Progress bar | Bottom | Seek, shows buffered range |
| Current time | Bottom-left | `HH:MM:SS` format |
| Duration | Bottom-right | `HH:MM:SS` format |
| Volume | Bottom-right | Slider with mute toggle |

### Subtitle Overlay
- Position: 10% from bottom (inside video area, above letterbox if any)
- Style: white text, black outline/shadow, font-size ~18sp
- 2-line max, bottom line = current dialogue

### Lock Screen
- Tap lock icon → disable all touch controls except unlock

### Picture-in-Picture
- PiP button → enters PiP mode (iOS/Android native)

### API Integration
- `POST /playback/{mediaId}/info` → `{ stream_url, direct_url, abr_url, subtitle_tracks, audio_tracks }`
- `POST /stream/{mediaId}/url` → `{ direct_url, hls_url, token, api_key, expires_in }` (for Chromecast/external player)
- `GET /profile/progress/{mediaId}` → get current progress
- `PUT /profile/progress/{mediaId}` → `{ position, completed }` save progress

### Mobile (375x667)
- Full screen video, fixed position
- Tap to show/hide controls
- Landscape: full landscape, different control layout

### Tablet (768x1024)
- Same full-screen behavior
- Larger controls

---

## 10. Search Screen

### Layout Structure
```
┌──────────────────────────────────────┐
│  ←  [🔍 Search movies, series...]  [X] │  ← Search input
├──────────────────────────────────────┤
│                                      │
│  Recent Searches                     │  ← Section (if no query)
│  ┌──────────────────────────────┐   │
│  │ 🔍 Avatar                   │   │
│  │ 🔍 Friends                  │   │
│  │ 🔍 Marvel                   │   │
│  └──────────────────────────────┘   │
│                                      │
│  OR (with query results)            │
│                                      │
│  ┌────┬────┬────┬────┬────┐        │
│  │    │    │    │    │    │        │
│  │    │    │    │    │    │        │
│  └────┴────┴────┴────┴────┘        │
│  Movies (12)                        │
│                                      │
│  ┌────┬────┬────┬────┬────┐        │
│  │    │    │    │    │    │        │
│  │    │    │    │    │    │        │
│  └────┴────┴────┴────┴────┘        │
│  Series (5)                         │
│                                      │
└──────────────────────────────────────┘
```

### Search Input
- Auto-focus on open
- Debounced search (300ms)
- Clear button (X) to reset
- Back arrow to close

### Recent Searches
- Shown when query is empty
- Tappable chips
- Clear all option

### Search Results
- Grouped by type: Movies, Series
- Count shown in section header
- Same card layout as browse

### API Integration
- `GET /search?q={query}&limit=20` → `{ movies: Media[], series: Series[] }`

---

## 11. Settings Screen

### Layout Structure
```
┌──────────────────────────────────────┐
│  ←  Settings                        │  ← Header
├──────────────────────────────────────┤
│  ┌────────────────────────────────┐   │
│  │  [Avatar]  Admin             │   │  ← User card
│  │            Administrator       │   │
│  └────────────────────────────────┘   │
│                                      │
│  [Settings combobox ▼]              │  ← Section selector dropdown
│  ┌────────────────────────────────┐   │
│  │                                │   │
│  │     Section Content            │   │
│  │     (changes based on          │   │
│  │      combobox selection)       │   │
│  │                                │   │
│  └────────────────────────────────┘   │
│                                      │
└──────────────────────────────────────┘
```

### Settings Sections (combobox options)

**User Settings (all users):**
1. Profile (section=profile) - username, display name, role
2. Preferences (section=preferences) - language, theme, grid size
3. Security (section=security) - change password
4. Sessions (section=sessions) - active sessions
5. Notifications (section=notifications) - notification toggles
6. Metadata (section=metadata) - metadata language, country
7. Subtitles (section=subtitles) - subtitle settings
8. Playback (section=playback) - quality, skip seconds
9. Cinema Mode (section=cinema) - trailers, intro
10. Pre-transcode (section=pretranscode) - pretranscode settings
11. Skip Intro / Credits (section=markers) - marker settings

**Admin Settings (admin only):**
12. Dashboard (section=general) - server stats
13. Libraries (section=libraries) - library management
14. Users (section=users) - user management
15. Activity (section=activity) - activity feed
16. Tasks (section=tasks) - scheduled tasks
17. Webhooks (section=webhooks) - webhook management

### Settings Tabs Detail

#### 11.1 Profile (`settings?section=profile`)
```
┌──────────────────────────────────────┐
│  Profile                             │
│  Manage your account information     │
│                                      │
│  Username                            │
│  ┌──────────────────────────────┐   │
│  │ admin                    [👻] │   │  ← Disabled input
│  └──────────────────────────────┘   │
│  "Username cannot be changed"       │
│                                      │
│  Display Name                        │
│  ┌──────────────────────────────┐   │
│  │ Admin                        │   │  ← Editable
│  └──────────────────────────────┘   │
│                                      │
│  Role                                │
│  Administrator                       │  ← Read-only
│                                      │
│  [Save Changes]                      │
└──────────────────────────────────────┘
```

#### 11.2 Preferences (`settings?section=preferences`)
| Setting | Type | Default |
|---------|------|---------|
| Language | Dropdown | English |
| Theme | Dropdown | System |
| Default home page | Dropdown | Home |
| Grid size | Slider | Medium |

#### 11.3 Security (`settings?section=security`)
| Setting | Type | Options |
|---------|------|---------|
| Current password | TextInput | - |
| New password | TextInput | - |
| Confirm password | TextInput | - |
| [Update Password] | Button | - |

#### 11.4 Sessions (`settings?section=sessions`)
```
┌──────────────────────────────────────┐
│  Active Sessions (2)                │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ 📱 Chrome on Mac            │   │
│  │ Current session             │   │
│  │ Last active: Just now       │   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ 📱 Safari on iPhone         │   │
│  │ Last active: 2 hours ago    │   │
│  │              [Revoke]       │   │
│  └──────────────────────────────┘   │
└──────────────────────────────────────┘
```

#### 11.5 Notifications (`settings?section=notifications`)
| Setting | Type | Default |
|---------|------|---------|
| New content | Toggle | On |
| Library updates | Toggle | On |
| System notifications | Toggle | On |

#### 11.6 Metadata (`settings?section=metadata`)
| Setting | Type | Default |
|---------|------|---------|
| Preferred metadata language | Dropdown | English |
| Country | Dropdown | US |
| Auto-download missing posters | Toggle | Off |

#### 11.7 Subtitles (`settings?section=subtitles`)
| Setting | Type | Default |
|---------|------|---------|
| Default subtitle language | Dropdown | English |
| Auto-load subtitles | Toggle | On |
| Subtitle font size | Slider | Medium |
| Subtitle background | Toggle | Off |

#### 11.8 Playback (`settings?section=playback`)
| Setting | Type | Default |
|---------|------|---------|
| Default quality | Dropdown | Auto |
| Direct play ( codec compatible) | Toggle | On |
| Skip forward (seconds) | Dropdown | 30 |
| Skip backward (seconds) | Dropdown | 10 |
| Auto-play next episode | Toggle | On |

#### 11.9 Cinema Mode (`settings?section=cinema`)
| Setting | Type | Default |
|---------|------|---------|
| Enable cinema mode | Toggle | Off |
| Max trailers | Number input | 2 |
| Custom intro | File upload | - |

#### 11.10 Pre-transcode (`settings?section=pretranscode`)
```
┌──────────────────────────────────────┐
│  Pre-transcode                       │
│                                      │
│  [⚡ Start Pre-transcode Jobs]       │  ← Primary button
│                                      │
│  ┌──────────────────────────────┐   │
│  │ Avatar: Fire and Ash         │   │
│  │ 4K HEVC → 1080p HLS         │   │
│  │ ████████████░░░░ 75%        │   │
│  └──────────────────────────────┘   │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ The Matrix                  │   │
│  │ 1080p HEVC → 720p HLS       │   │
│  │ Queued                      │   │
│  └──────────────────────────────┘   │
└──────────────────────────────────────┘
```

#### 11.11 Skip Intro / Credits (`settings?section=markers`)
| Setting | Type | Default |
|---------|------|---------|
| Skip intro | Toggle | On |
| Skip credits | Toggle | On |
| Auto-skip sponsor messages | Toggle | On |

#### 11.12 Dashboard (`settings?section=general`) - Admin
```
┌──────────────────────────────────────┐
│  Dashboard                           │
│                                      │
│  📊 Library Stats                    │
│  ┌────────┐ ┌────────┐ ┌────────┐  │
│  │  24    │ │   3    │ │   2    │  │
│  │ Movies │ │ Series │ │Episodes│  │
│  └────────┘ └────────┘ └────────┘  │
│                                      │
│  📈 Recent Activity                  │
│  • Admin watched Avatar (2h ago)    │
│  • Library scan completed (5h ago)  │
│                                      │
│  💾 Storage                          │
│  Used: 1.2 TB / 4 TB (30%)         │
│  ████████░░░░░░░░░░░░░             │
└──────────────────────────────────────┘
```

#### 11.13 Libraries (`settings?section=libraries`) - Admin
| Library | Type | Path | Items |
|---------|------|------|-------|
| Movies | movie | /media/movies | 24 |
| TV Shows | series | /media/series | 3 |

#### 11.14 Users (`settings?section=users`) - Admin
```
┌──────────────────────────────────────┐
│  Users                               │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ 👤 Admin                     │   │
│  │ Administrator                │   │
│  │ Last active: Just now        │   │
│  │                        [Edit]│   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ 👤 Guest                     │   │
│  │ Guest                        │   │
│  │ Last active: 2 days ago      │   │
│  │                        [Edit]│   │
│  └──────────────────────────────┘   │
│                                      │
│  [+ Add User]                        │
└──────────────────────────────────────┘
```

#### 11.15 Activity (`settings?section=activity`) - Admin
```
┌──────────────────────────────────────┐
│  Activity Feed                       │
│                                      │
│  Today                               │
│  • Admin watched Avatar 2h ago      │
│  • Guest played Friends 5h ago      │
│                                      │
│  Yesterday                           │
│  • Admin updated Avatar metadata    │
│  • Library scan completed           │
└──────────────────────────────────────┘
```

#### 11.16 Tasks (`settings?section=tasks`) - Admin
| Task | Status | Schedule |
|------|--------|----------|
| Library Scan | Idle | Daily at 3:00 AM |
| thumbnail Generation | Idle | On media add |
| Pre-transcode | Running | Manual |

#### 11.17 Webhooks (`settings?section=webhooks`) - Admin
```
┌──────────────────────────────────────┐
│  Webhooks                            │
│                                      │
│  ┌──────────────────────────────┐   │
│  │ https://example.com/hook     │   │
│  │ Events: media.play, media.add│   │
│  │ Active                 [Edit]│   │
│  └──────────────────────────────┘   │
│                                      │
│  [+ Add Webhook]                     │
└──────────────────────────────────────┘
```

### API Integration
**Auth:**
- `GET /auth/me` → current user info
- `POST /auth/change-password` → `{ current_password, new_password }`

**Profile:**
- `GET /profile` → user profile
- `PATCH /profile` → update profile `{ display_name }`
- `GET /profile/preferences` → user preferences
- `PUT /profile/preferences` → update preferences

**Sessions:**
- `GET /auth/sessions` → active sessions
- `DELETE /auth/sessions/{id}` → revoke session

**Notifications:**
- `GET /notifications?unread_only=false&limit=20&offset=0` → notifications list
- `PATCH /notifications/read` → `{ ids: number[] }` mark as read
- `PATCH /notifications/read-all` → mark all read
- `POST /notifications/delete` → `{ ids: number[] }` delete
- `GET /notifications/unread-count` → `{ count: number }`

**Libraries (Admin):**
- `GET /libraries` → list libraries
- `POST /libraries` → create library
- `DELETE /libraries/{id}` → delete library
- `POST /libraries/{id}/scan` → scan library
- `GET /admin/fs/browse?path={path}` → filesystem browse

**Users (Admin):**
- `GET /admin/users` → user list
- `POST /admin/users` → create user
- `PUT /admin/users/{id}` → update user
- `DELETE /admin/users/{id}` → delete user
- `GET /admin/users/{id}/libraries` → user library access
- `PUT /admin/users/{id}/libraries/{libraryId}` → set library access

**Admin Settings:**
- `GET /admin/settings/{slug}` / `PUT /admin/settings/{slug}` → settings (tmdb, omdb, tvdb, fanart, opensubtitles, playback, pretranscode, cinema)
- `POST /admin/metadata/refresh-ratings` → bulk refresh ratings

**Pretranscode (Admin):**
- `GET /admin/pretranscode/status` → current status
- `GET /admin/pretranscode/profiles` → profiles list
- `PUT /admin/pretranscode/profiles/{id}` → toggle profile
- `GET /admin/pretranscode/estimate?library_id={id}` → storage estimate
- `POST /admin/pretranscode/start` → start jobs
- `POST /admin/pretranscode/stop` → stop jobs
- `POST /admin/pretranscode/resume` → resume
- `POST /admin/pretranscode/cleanup-files` → cleanup files

**Markers (Admin):**
- `GET /admin/markers/stats` → marker statistics
- `GET /admin/markers/detectors` → available detectors
- `POST /admin/markers/backfill` → backfill markers

**Activity (Admin):**
- `GET /admin/activity` → activity log
- `GET /admin/stats/libraries` → library statistics
- `GET /admin/stats/playback` → playback statistics

**Webhooks (Admin):**
- `GET /admin/webhooks` → webhook list
- `POST /admin/webhooks` → create webhook
- `PUT /admin/webhooks/{id}` → update webhook
- `DELETE /admin/webhooks/{id}` → delete webhook

**Tasks (Admin):**
- `GET /admin/tasks` → scheduled tasks
- `POST /admin/tasks/{name}/run` → run task manually

**Server:**
- `GET /admin/server` → server info (version, uptime)

---

## 12. Component Specifications

### 12.1 MediaCard
```
Props:
  - media: { id, title, type, poster_url, rating?, year?, progress? }
  - variant: 'default' | 'continue' | 'next-up'
  - size: 'small' | 'medium' | 'large'
```

| Variant | Shows | Additional |
|---------|-------|------------|
| default | poster, title, year, rating | - |
| continue | poster, title, episode info, progress bar | dismiss button |
| next-up | poster, title, episode info | - |

### 12.2 BottomTabBar
```
Props:
  - activeTab: string
  - onTabPress: (route) => void
  - notifications?: number
```

- Height: 64px + bottom safe area
- Icons: 24x24
- Active: primary color, filled icon
- Inactive: muted color, outline icon
- Badge: red circle with white text for notifications

### 12.3 ActionMenu
```
Props:
  - items: { label, icon, onPress, danger?, separator? }[]
  - trigger: 'press' | 'long-press'
```

- Portal-based dropdown
- Z-index: 9999
- Backdrop: semi-transparent black
- Items: 44px height minimum (touch target)

### 12.4 VideoPlayer
```
Props:
  - source: { type: 'direct' | 'hls', url: string }
  - autoPlay?: boolean
  - startPosition?: number
  - subtitles?: { language, url }[]
  - audioTracks?: { language, url }[]
  - onProgress: (position) => void
  - onEnded: () => void
```

- Full-screen, fixed position
- Lock/unlock gesture
- Double-tap sides to skip
- Pinch to zoom (when not in full-screen)
- Swipe up/down for volume (right side) / brightness (left side)

---

## 13. Navigation Structure

```
RootNavigator (Stack)
├── AuthStack
│   └── LoginScreen
│
└── MainTabs (Bottom Tab Navigator)
    ├── HomeStack
    │   ├── HomeScreen
    │   └── WatchScreen
    │
    ├── MoviesStack
    │   ├── MoviesScreen
    │   ├── MovieDetailScreen
    │   └── WatchScreen
    │
    ├── SeriesStack
    │   ├── SeriesScreen
    │   ├── SeriesDetailScreen
    │   └── WatchScreen
    │
    ├── BrowseStack
    │   ├── BrowseScreen
    │   ├── MoviesScreen
    │   └── SeriesScreen
    │
    └── FavoritesStack
        ├── FavoritesScreen
        └── MediaDetailScreen

Modal Stack (over MainTabs)
├── SearchScreen (full-screen modal)
└── MediaDetailModal (half-sheet on mobile)

Settings is accessed via user menu (not a tab/modal) - pushes onto stack with back navigation
```

---

## 14. Responsive Breakpoints

| Breakpoint | Width | Layout |
|------------|-------|--------|
| Mobile | < 640px | Single column, bottom tabs |
| Tablet Portrait | 640px - 1024px | Multi-column grid, bottom tabs |
| Tablet Landscape | > 1024px | Side navigation (future) |

---

## 15. Dark Mode Colors

Based on actual Tailwind 4 `@theme` values from `webapp/src/index.css` (Netflix-inspired palette):

| Token | Light | Dark (Actual) |
|-------|-------|---------------|
| background | #FFFFFF | #141414 (netflix-black) |
| surface | #F5F5F5 | #181818 (netflix-dark) |
| primary | #3B82F6 | #0071EB (netflix-blue) |
| text | #1F2937 | #FFFFFF (netflix-white) |
| textMuted | #6B7280 | #808080 (netflix-light-gray) |
| border | #E5E7EB | #2F2F2F (netflix-gray) |
| error | #EF4444 | #E50914 (netflix-red) |
| success | #22C55E | #22C55E |

---

## 16. API Error Handling

| Error | UI Response |
|-------|-------------|
| 401 Unauthorized | Redirect to login, clear tokens |
| 403 Forbidden | Show toast "Permission denied" |
| 404 Not Found | Show "Content not available" empty state |
| 500 Server Error | Show "Something went wrong" with retry button |
| Network Error | Show "No internet connection" banner |

### Token Refresh Flow
1. Request fails with 401
2. Queue this request (module-level singleton, prevent duplicate refresh)
3. Call `POST /auth/refresh` with `{ refresh_token }`
4. On success: update tokens via `setTokensCallback(data.access_token, data.refresh_token, data.expires_in)`, then retry queued request
5. On failure: call `onSessionExpiredCallback()`, clear auth, redirect to login

---

## 17. Offline Support

| Feature | Strategy |
|---------|----------|
| Watch progress | Persist locally, sync when online |
| Favorites | Persist locally, sync when online |
| Recently viewed | Persist locally (last 50) |
| Playback | Only works when online (no offline playback for v1) |
