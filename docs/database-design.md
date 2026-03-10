# Velox Database Design
Version: 1.1
Last updated: 2026-03-11

## Design Principles

1. **Logical vs Physical separation**: `media` (content identity) tách biệt `media_files` (file trên disk)
2. **Multi-version**: 1 movie có thể có nhiều files (720p, 1080p, 4K, REMUX)
3. **Unified per-user state**: progress, favorites, ratings gộp trong 1 bảng `user_data` (inspired by Emby/Jellyfin)
4. **Soft references**: series → seasons → episodes → media (FK chain)
5. **Migration-first**: Mọi thay đổi qua versioned migrations
6. **Multi-provider metadata**: `tmdb_id` indexed + `external_ids` table cho các provider khác (IMDb, TVDB, AniDB)

---

## Entity Relationship Diagram

```
┌──────────────┐     ┌──────────────────┐     ┌───────────────┐
│  libraries   │────<│     media        │────<│  media_files  │
│              │     │  (logical item)  │     │ (physical file)│
└──────────────┘     └────────┬─────────┘     └───────┬───────┘
                              │                       │
                    ┌─────────┴──────────┐    ┌───────┴────────┐
                    │                    │    │                 │
              ┌─────┴─────┐    ┌────────┴┐   │  ┌───────────┐ │
              │  episodes │    │ credits  │   ├─<│ subtitles │ │
              │           │    └────┬─────┘   │  └───────────┘ │
              └─────┬─────┘         │         │  ┌─────────────┤
                    │          ┌────┴─────┐   └─<│audio_tracks │
              ┌─────┴─────┐   │  people  │      └─────────────┘
              │  seasons   │   └──────────┘
              └─────┬─────┘
              ┌─────┴─────┐     ┌──────────┐     ┌──────────────┐
              │  series    │     │  genres  │────<│ media_genres  │
              └───────────┘     └──────────┘     └──────────────┘

┌──────────┐     ┌──────────────┐     ┌───────────────────┐
│  users   │────<│  user_data   │     │ user_library_access│
│          │────<│refresh_tokens │     └───────────────────┘
│          │────<│  sessions    │
└──────────┘     └──────────────┘

┌──────────────┐
│ external_ids │  (media/series → TMDb, IMDb, TVDB, AniDB)
└──────────────┘

┌──────────────┐     ┌───────────────┐
│  scan_jobs   │     │ activity_log  │
└──────────────┘     └───────────────┘

┌──────────────┐     ┌───────────────┐
│  task_history│     │   webhooks    │
└──────────────┘     └───────────────┘
```

---

## Tables Detail

### Migration 001: Initial Schema (DONE ✅)

#### `libraries`
Media library = 1 thư mục gốc trên disk.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| name | TEXT | NOT NULL | Display name ("Movies", "TV Shows") |
| path | TEXT | NOT NULL UNIQUE | Absolute path ("/mnt/media/movies") |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

---

### Migration 002: Core Media Model

#### `libraries` (ALTER)
Thêm library type để scanner biết parse kiểu nào.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| type | TEXT | DEFAULT 'mixed' | 'movies' \| 'tvshows' \| 'mixed' |

#### `media`
**Logical identity** - đại diện cho 1 "content". Tách biệt khỏi file vật lý.
1 media = 1 movie HOẶC 1 episode (linked qua episodes table).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| library_id | INTEGER | FK → libraries(id) CASCADE | Library chứa media này |
| media_type | TEXT | NOT NULL DEFAULT 'movie' | 'movie' \| 'episode' |
| title | TEXT | NOT NULL | Tên hiển thị |
| sort_title | TEXT | DEFAULT '' | Tên sort (bỏ "The ", "A ") |
| tmdb_id | INTEGER | DEFAULT NULL | TMDb ID để match metadata |
| imdb_id | TEXT | DEFAULT NULL | IMDb ID (tt1234567) - hay dùng cho cross-reference |
| overview | TEXT | DEFAULT '' | Mô tả/synopsis |
| release_date | TEXT | DEFAULT '' | YYYY-MM-DD |
| rating | REAL | DEFAULT 0 | TMDb rating (0-10) |
| poster_path | TEXT | DEFAULT '' | Local path to poster image |
| backdrop_path | TEXT | DEFAULT '' | Local path to backdrop image |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_media_library` ON (library_id)
- `idx_media_tmdb` ON (tmdb_id) WHERE tmdb_id IS NOT NULL
- `idx_media_imdb` ON (imdb_id) WHERE imdb_id IS NOT NULL
- `idx_media_type` ON (media_type)
- `idx_media_title` ON (sort_title)

**Note:** Không còn lưu file_path, video_codec, etc ở đây. Chúng nằm ở `media_files`.
**Note:** TMDb + IMDb là 2 provider chính, indexed trực tiếp. Các provider khác (TVDB, AniDB...) lưu trong bảng `external_ids`.

#### `media_files`
**Physical file** - 1 file video trên disk. 1 media có thể có nhiều files (multi-version).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| media_id | INTEGER | FK → media(id) CASCADE | Media item sở hữu file này |
| file_path | TEXT | NOT NULL UNIQUE | Absolute path tới file |
| file_size | INTEGER | DEFAULT 0 | Size in bytes |
| duration | REAL | DEFAULT 0 | Duration in seconds |
| width | INTEGER | DEFAULT 0 | Video width (px) |
| height | INTEGER | DEFAULT 0 | Video height (px) |
| video_codec | TEXT | DEFAULT '' | h264, hevc, vp9, av1 |
| audio_codec | TEXT | DEFAULT '' | aac, dts, ac3, truehd (default track) |
| container | TEXT | DEFAULT '' | matroska, mov,mp4 |
| bitrate | INTEGER | DEFAULT 0 | Overall bitrate (bps) |
| fingerprint | TEXT | DEFAULT '' | "{file_size}:{header_hash}" — path-independent, xem rationale |
| is_primary | INTEGER | DEFAULT 1 | 1 = default version for playback |
| added_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |
| last_verified_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | NULL = file missing |

**Indexes:**
- `idx_mf_media` ON (media_id)
- `idx_mf_fingerprint` ON (fingerprint)
- `idx_mf_path` ON (file_path) — UNIQUE constraint covers this

**Design Rationale:**
- `fingerprint` = `"{file_size}:{xxhash64_of_first_64KB}"` — KHÔNG chứa file_path
  - Khi file rename/move: path thay đổi nhưng fingerprint giữ nguyên → match được file cũ
  - xxHash64 trên 64KB đầu: cực nhanh (~2GB/s), đủ unique cho media files khác nhau
  - Collision risk: 2 file cùng size + cùng 64KB đầu là rất hiếm, nhưng vẫn phải coi là soft match chứ không phải hard identity
  - Nếu fingerprint match, scanner nên verify thêm metadata nhẹ (`duration`, `container`, `video_codec`) trước khi quyết định đây là cùng một file đã rename
- `is_primary` cho phép user chọn version mặc định (ví dụ: prefer 1080p over 4K)
- `last_verified_at = NULL` nghĩa là file missing trên disk (drive unmounted, deleted)

---

### Migration 003: Series Model

#### `series`
TV Show container. Sở hữu seasons và episodes.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| library_id | INTEGER | FK → libraries(id) CASCADE | |
| title | TEXT | NOT NULL | "Breaking Bad" |
| sort_title | TEXT | DEFAULT '' | |
| tmdb_id | INTEGER | UNIQUE, DEFAULT NULL | |
| imdb_id | TEXT | DEFAULT NULL | IMDb ID |
| overview | TEXT | DEFAULT '' | |
| status | TEXT | DEFAULT '' | 'Returning Series' \| 'Ended' \| 'Canceled' |
| first_air_date | TEXT | DEFAULT '' | YYYY-MM-DD |
| poster_path | TEXT | DEFAULT '' | |
| backdrop_path | TEXT | DEFAULT '' | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_series_library` ON (library_id)
- `idx_series_tmdb` ON (tmdb_id) WHERE tmdb_id IS NOT NULL

#### `seasons`
Season of a series.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| series_id | INTEGER | FK → series(id) CASCADE | |
| season_number | INTEGER | NOT NULL | 0 = Specials, 1+ = regular |
| title | TEXT | DEFAULT '' | "Season 1" or custom title |
| overview | TEXT | DEFAULT '' | |
| poster_path | TEXT | DEFAULT '' | |
| episode_count | INTEGER | DEFAULT 0 | Total episodes (from TMDb) |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_seasons_series` ON (series_id)
- UNIQUE (series_id, season_number)

#### `episodes`
Episode links a season to a media item.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| series_id | INTEGER | FK → series(id) CASCADE | |
| season_id | INTEGER | FK → seasons(id) CASCADE | |
| media_id | INTEGER | FK → media(id) CASCADE, UNIQUE | Link to media item (playable) |
| episode_number | INTEGER | NOT NULL | |
| title | TEXT | DEFAULT '' | Episode title |
| overview | TEXT | DEFAULT '' | |
| still_path | TEXT | DEFAULT '' | Episode screenshot |
| air_date | TEXT | DEFAULT '' | YYYY-MM-DD |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_ep_series` ON (series_id)
- `idx_ep_season` ON (season_id)
- UNIQUE (season_id, episode_number)

**Relationship flow:**
```
series (1) → seasons (N) → episodes (N) → media (1) → media_files (N)
                                            ↑
                                     episode.media_id = media.id
                                     media.media_type = 'episode'
```

---

### Migration 004: Genres & People

#### `genres`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| name | TEXT | NOT NULL UNIQUE | "Action", "Comedy" |
| tmdb_id | INTEGER | UNIQUE, DEFAULT NULL | TMDb genre ID |

#### `media_genres`
Many-to-many: media ↔ genre. Cũng dùng cho series.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| media_id | INTEGER | DEFAULT NULL | FK → media(id) CASCADE (for movies) |
| series_id | INTEGER | DEFAULT NULL | FK → series(id) CASCADE (for series) |
| genre_id | INTEGER | NOT NULL | FK → genres(id) CASCADE |

**Constraints:**
- CHECK ((media_id IS NOT NULL AND series_id IS NULL) OR (media_id IS NULL AND series_id IS NOT NULL)) — exactly one owner
- UNIQUE (media_id, genre_id) WHERE media_id IS NOT NULL
- UNIQUE (series_id, genre_id) WHERE series_id IS NOT NULL

#### `people`
Actors, directors, writers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| name | TEXT | NOT NULL | |
| tmdb_id | INTEGER | UNIQUE, DEFAULT NULL | |
| profile_path | TEXT | DEFAULT '' | Local path to headshot |

#### `credits`
Who worked on what, in what role.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| media_id | INTEGER | DEFAULT NULL | FK → media(id) CASCADE (for movies) |
| series_id | INTEGER | DEFAULT NULL | FK → series(id) CASCADE (for series) |
| person_id | INTEGER | NOT NULL | FK → people(id) CASCADE |
| character | TEXT | DEFAULT '' | "Walter White" |
| role | TEXT | NOT NULL | 'cast' \| 'director' \| 'writer' |
| display_order | INTEGER | DEFAULT 0 | Sort order (billing) |

**Constraints:**
- CHECK ((media_id IS NOT NULL AND series_id IS NULL) OR (media_id IS NULL AND series_id IS NOT NULL)) — exactly one owner

**Indexes:**
- `idx_credits_media` ON (media_id) WHERE media_id IS NOT NULL
- `idx_credits_series` ON (series_id) WHERE series_id IS NOT NULL
- `idx_credits_person` ON (person_id)

---

### Migration 005: Scan Jobs

#### `scan_jobs`
Track scan progress per library.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| library_id | INTEGER | FK → libraries(id) CASCADE | |
| status | TEXT | NOT NULL DEFAULT 'queued' | 'queued' \| 'scanning' \| 'completed' \| 'failed' |
| total_files | INTEGER | DEFAULT 0 | |
| scanned_files | INTEGER | DEFAULT 0 | |
| new_files | INTEGER | DEFAULT 0 | Files added this scan |
| errors | INTEGER | DEFAULT 0 | |
| error_log | TEXT | DEFAULT '' | Last few error messages |
| started_at | DATETIME | DEFAULT NULL | |
| finished_at | DATETIME | DEFAULT NULL | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_scanjob_library` ON (library_id)
- `idx_scanjob_status` ON (status)

---

### Migration 006: Subtitles & Audio Tracks

#### `subtitles`
All known subtitle tracks (embedded + external sidecar files).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| media_file_id | INTEGER | FK → media_files(id) CASCADE | |
| language | TEXT | DEFAULT '' | ISO 639-1 code: 'en', 'vi', 'ja' |
| codec | TEXT | DEFAULT '' | subrip, ass, hdmv_pgs_subtitle, dvd_subtitle, webvtt |
| title | TEXT | DEFAULT '' | Track title from file ("English SDH") |
| is_embedded | INTEGER | DEFAULT 1 | 1 = in video file, 0 = external .srt/.vtt |
| stream_index | INTEGER | DEFAULT -1 | FFmpeg stream index (embedded only) |
| file_path | TEXT | DEFAULT '' | Path to external subtitle file |
| is_forced | INTEGER | DEFAULT 0 | Forced subtitle (foreign language parts) |
| is_default | INTEGER | DEFAULT 0 | Default track flag |
| is_sdh | INTEGER | DEFAULT 0 | Subtitles for Deaf/Hard of Hearing |

**Constraints:**
- UNIQUE (media_file_id, stream_index) WHERE is_embedded = 1 — idempotent rescan for embedded subs
- UNIQUE (media_file_id, file_path) WHERE is_embedded = 0 — idempotent rescan for external subs

**Indexes:**
- `idx_sub_mediafile` ON (media_file_id)
- `idx_sub_lang` ON (language)

**Subtitle type detection:**
- `codec` in ('subrip', 'ass', 'ssa', 'webvtt', 'mov_text') → **text-based** → extract to VTT, render in browser
- `codec` in ('hdmv_pgs_subtitle', 'dvd_subtitle') → **image-based** → must burn-in via FFmpeg transcode

#### `audio_tracks`
All audio tracks in a media file. Enables multi-audio switching.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| media_file_id | INTEGER | FK → media_files(id) CASCADE | |
| stream_index | INTEGER | NOT NULL | FFmpeg stream index |
| codec | TEXT | DEFAULT '' | aac, ac3, eac3, dts, truehd, flac, opus, mp3 |
| language | TEXT | DEFAULT '' | ISO 639-1: 'en', 'vi', 'ja' |
| channels | INTEGER | DEFAULT 2 | 2=stereo, 6=5.1, 8=7.1 |
| channel_layout | TEXT | DEFAULT '' | 'stereo', '5.1', '7.1' |
| bitrate | INTEGER | DEFAULT 0 | Audio bitrate (bps) |
| title | TEXT | DEFAULT '' | "English DTS-HD" |
| is_default | INTEGER | DEFAULT 0 | Default track in file |

**Constraints:**
- UNIQUE (media_file_id, stream_index) — idempotent rescan, mỗi stream_index chỉ có 1 row

**Indexes:**
- `idx_audio_mediafile` ON (media_file_id)

---

### Migration 007: Users & Auth

#### `users`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| username | TEXT | NOT NULL UNIQUE | Login username |
| display_name | TEXT | NOT NULL | Shown in UI |
| password_hash | TEXT | NOT NULL | bcrypt hash, cost 12 |
| is_admin | INTEGER | DEFAULT 0 | |
| avatar_path | TEXT | DEFAULT '' | Local path to avatar |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

#### `user_preferences`
Per-user settings.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| user_id | INTEGER | PK, FK → users(id) CASCADE | |
| subtitle_language | TEXT | DEFAULT '' | Preferred subtitle: 'vi', 'en' |
| audio_language | TEXT | DEFAULT '' | Preferred audio: 'ja', 'en' |
| max_streaming_quality | TEXT | DEFAULT 'auto' | 'auto' \| '2160p' \| '1080p' \| '720p' \| '480p' |
| theme | TEXT | DEFAULT 'dark' | 'dark' \| 'light' \| 'auto' |

#### `user_library_access`
Whitelist: which libraries a user can see.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| user_id | INTEGER | FK → users(id) CASCADE | |
| library_id | INTEGER | FK → libraries(id) CASCADE | |

**Constraints:** PRIMARY KEY (user_id, library_id)

---

### Migration 008: Auth Tokens & Sessions

#### `refresh_tokens`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| user_id | INTEGER | FK → users(id) CASCADE | |
| token_hash | TEXT | NOT NULL UNIQUE | SHA256 of refresh token |
| device_name | TEXT | DEFAULT '' | "Chrome on macOS" |
| ip_address | TEXT | DEFAULT '' | |
| expires_at | DATETIME | NOT NULL | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_rt_user` ON (user_id)
- `idx_rt_expires` ON (expires_at)

#### `sessions`
Active session = 1 logged-in device. Linked to refresh_token for revocation.
DELETE row = logout/revoke device.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| user_id | INTEGER | FK → users(id) CASCADE | |
| refresh_token_id | INTEGER | FK → refresh_tokens(id) SET NULL | Link to auth credential |
| device_name | TEXT | DEFAULT '' | "Chrome on macOS" |
| ip_address | TEXT | DEFAULT '' | |
| user_agent | TEXT | DEFAULT '' | |
| expires_at | DATETIME | NOT NULL | Session expiry (= refresh token expiry) |
| last_active_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | Updated on each API call |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_sessions_user` ON (user_id)
- `idx_sessions_expires` ON (expires_at) — cleanup expired sessions

**Design Rationale:**
- Revoke device = DELETE session + DELETE linked refresh_token → immediate logout
- `expires_at` mirrors refresh_token expiry — cleanup job deletes expired rows
- No `revoked_at`: revoke = delete, simpler than soft-delete for sessions

---

### Migration 009: Per-User State

#### `user_data`
Unified per-user-per-item state. Gộp progress + favorite + rating vào 1 bảng.
Inspired by Emby/Jellyfin's `UserDatas` pattern — 1 JOIN duy nhất khi render poster + progress bar + favorite icon.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| user_id | INTEGER | NOT NULL | FK → users(id) CASCADE |
| media_id | INTEGER | NOT NULL | FK → media(id) CASCADE |
| position | REAL | DEFAULT 0 | Watch position in seconds |
| completed | INTEGER | DEFAULT 0 | 1 = finished watching |
| is_favorite | INTEGER | DEFAULT 0 | 1 = user favorited |
| rating | REAL | DEFAULT NULL | User rating 1.0 - 10.0 (NULL = not rated) |
| play_count | INTEGER | DEFAULT 0 | Times fully watched |
| last_played_at | DATETIME | DEFAULT NULL | Last play timestamp |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Constraints:** PRIMARY KEY (user_id, media_id) — replaces old PK(media_id)

**Indexes:**
- `idx_ud_user` ON (user_id)
- `idx_ud_continue` ON (user_id, completed, position) WHERE completed = 0 AND position > 0 — "Continue Watching" query
- `idx_ud_favorites` ON (user_id, is_favorite) WHERE is_favorite = 1

**Design Rationale:**
- Home screen render poster: cần progress (position) + favorite (heart icon) + rating → 1 JOIN thay vì 3
- `play_count` hỗ trợ "Most Watched" dashboard
- `last_played_at` cho "Recently Played" sort
- `rating = NULL` phân biệt "chưa rate" vs "rate 0"
- Series favorite: dùng `user_series_data` (xem bên dưới)

#### `user_series_data`
Per-user state cho series (favorite, rating). Tách riêng vì series không có progress/position.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| user_id | INTEGER | NOT NULL | FK → users(id) CASCADE |
| series_id | INTEGER | NOT NULL | FK → series(id) CASCADE |
| is_favorite | INTEGER | DEFAULT 0 | |
| rating | REAL | DEFAULT NULL | 1.0 - 10.0 |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Constraints:** PRIMARY KEY (user_id, series_id)

---

### Migration 010: Activity & Webhooks (Plan F)

#### `activity_log`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| user_id | INTEGER | DEFAULT NULL | FK → users(id) SET NULL |
| action | TEXT | NOT NULL | 'play_start' \| 'play_stop' \| 'login' \| 'library_scan' \| 'media_added' |
| media_id | INTEGER | DEFAULT NULL | |
| details | TEXT | DEFAULT '' | JSON string with extra data |
| ip_address | TEXT | DEFAULT '' | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

**Indexes:**
- `idx_activity_user` ON (user_id)
- `idx_activity_action` ON (action)
- `idx_activity_created` ON (created_at)

#### `task_history`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| task_name | TEXT | NOT NULL | 'library_scan', 'transcode_cleanup', etc |
| status | TEXT | NOT NULL | 'running' \| 'completed' \| 'failed' |
| details | TEXT | DEFAULT '' | JSON |
| started_at | DATETIME | NOT NULL | |
| finished_at | DATETIME | DEFAULT NULL | |

**Indexes:**
- `idx_task_name` ON (task_name)

#### `webhooks`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| url | TEXT | NOT NULL | Webhook URL |
| events | TEXT | NOT NULL | Comma-separated: 'media_added,playback_start' |
| active | INTEGER | DEFAULT 1 | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | |

---

### Migration 011: External IDs (Plan A - TMDb Integration)

#### `external_ids`
Flexible provider ID storage. TMDb + IMDb đã có indexed columns trên `media`/`series`.
Bảng này cho các provider phụ (TVDB, AniDB, MusicBrainz...) và cho future-proofing.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PK AUTOINCREMENT | |
| item_type | TEXT | NOT NULL | 'media' \| 'series' \| 'season' \| 'episode' \| 'person' |
| item_id | INTEGER | NOT NULL | ID trong bảng tương ứng |
| provider | TEXT | NOT NULL | 'tvdb' \| 'anidb' \| 'anilist' \| 'musicbrainz' |
| provider_id | TEXT | NOT NULL | ID string từ provider |

**Constraints:**
- UNIQUE (item_type, item_id, provider)

**Indexes:**
- `idx_extid_item` ON (item_type, item_id)
- `idx_extid_provider` ON (provider, provider_id) — lookup by provider ID

**Design Rationale:**
- `tmdb_id` + `imdb_id` trên `media`/`series` = primary, indexed trực tiếp → fast lookup
- `external_ids` = secondary providers, ít query hơn nên EAV pattern OK
- `item_type` + `item_id` thay vì nhiều FK nullable → đơn giản, extensible
- Vì không có FK thật theo `item_type`, mọi delete trên `media/series/season/episode/person` phải cleanup `external_ids` ở application layer hoặc trong migration-safe cleanup job
- Nên có periodic integrity check: tìm `external_ids` orphaned và xóa/rebuild từ source of truth
- Khi cần lookup: `SELECT item_id FROM external_ids WHERE provider = 'tvdb' AND provider_id = '12345' AND item_type = 'series'`

**⚠️ Cleanup Strategy (no FK → must handle orphans):**
- Application-level: khi DELETE media/series/person → cũng DELETE FROM external_ids WHERE item_type=? AND item_id=?
- Implement trong repository layer (Go code), không dùng DB trigger
- Periodic cleanup job (Plan F scheduled tasks): `DELETE FROM external_ids WHERE item_type='media' AND item_id NOT IN (SELECT id FROM media)` — chạy weekly
- Acceptable tradeoff: orphan rows chỉ waste storage, không affect correctness

---

## Key Query Patterns

### 1. Home Screen - "Continue Watching"
```sql
SELECT m.id, m.title, m.poster_path, m.media_type,
       ud.position, ud.is_favorite, ud.updated_at,
       mf.duration
FROM user_data ud
JOIN media m ON m.id = ud.media_id
JOIN media_files mf ON mf.media_id = m.id AND mf.is_primary = 1
WHERE ud.user_id = ? AND ud.completed = 0 AND ud.position > 0
ORDER BY ud.updated_at DESC
LIMIT 20
```
**Note:** 1 JOIN lấy được cả progress + favorite status — không cần LEFT JOIN 3 bảng.

### 2. Home Screen - "Next Up" (next unwatched episode)
```sql
SELECT e.id, e.title, e.episode_number, e.still_path,
       s.title AS series_title, s.poster_path,
       sea.season_number
FROM episodes e
JOIN seasons sea ON sea.id = e.season_id
JOIN series s ON s.id = e.series_id
JOIN media m ON m.id = e.media_id
LEFT JOIN user_data ud ON ud.media_id = m.id AND ud.user_id = ?
WHERE s.id IN (
    -- Series the user has watched at least 1 episode of
    SELECT DISTINCT e2.series_id FROM episodes e2
    JOIN user_data ud2 ON ud2.media_id = e2.media_id AND ud2.user_id = ?
    WHERE ud2.position > 0 OR ud2.completed = 1
)
AND (ud.completed IS NULL OR ud.completed = 0)
AND (ud.position IS NULL OR ud.position = 0)  -- not started yet
ORDER BY sea.season_number, e.episode_number
LIMIT 1
```

### 3. Movie List with Genres
```sql
SELECT m.*, GROUP_CONCAT(g.name) AS genre_names
FROM media m
LEFT JOIN media_genres mg ON mg.media_id = m.id
LEFT JOIN genres g ON g.id = mg.genre_id
WHERE m.media_type = 'movie' AND m.library_id IN (
    SELECT library_id FROM user_library_access WHERE user_id = ?
)
GROUP BY m.id
ORDER BY m.sort_title
LIMIT ? OFFSET ?
```

### 4. Series Detail with Season/Episode Tree
```sql
-- Series info + user favorite/rating
SELECT s.*, usd.is_favorite, usd.rating AS user_rating
FROM series s
LEFT JOIN user_series_data usd ON usd.series_id = s.id AND usd.user_id = ?
WHERE s.id = ?;

-- Seasons
SELECT * FROM seasons WHERE series_id = ? ORDER BY season_number;

-- Episodes for a season (with watch status + favorite)
SELECT e.*, m.id AS media_id,
       ud.position, ud.completed, ud.is_favorite,
       mf.duration
FROM episodes e
JOIN media m ON m.id = e.media_id
LEFT JOIN media_files mf ON mf.media_id = m.id AND mf.is_primary = 1
LEFT JOIN user_data ud ON ud.media_id = m.id AND ud.user_id = ?
WHERE e.season_id = ?
ORDER BY e.episode_number
```

### 5. Playback Decision - Get Media File + Tracks
```sql
-- Primary file for a media item
SELECT * FROM media_files WHERE media_id = ? AND is_primary = 1;

-- All audio tracks
SELECT * FROM audio_tracks WHERE media_file_id = ? ORDER BY is_default DESC, stream_index;

-- All subtitles
SELECT * FROM subtitles WHERE media_file_id = ? ORDER BY is_default DESC, language;
```

### 6. User Favorites (movies + series combined)
```sql
-- Favorite movies
SELECT m.id, m.title, m.poster_path, 'movie' AS type, ud.updated_at
FROM user_data ud
JOIN media m ON m.id = ud.media_id
WHERE ud.user_id = ? AND ud.is_favorite = 1 AND m.media_type = 'movie'

UNION ALL

-- Favorite series
SELECT s.id, s.title, s.poster_path, 'series' AS type, usd.updated_at
FROM user_series_data usd
JOIN series s ON s.id = usd.series_id
WHERE usd.user_id = ? AND usd.is_favorite = 1

ORDER BY updated_at DESC
```

### 7. Scanner - Find by Fingerprint (rename detection)
```sql
-- File at new_path not in DB, compute fingerprint = "{size}:{xxhash64_first_64kb}"
-- Check if same file exists under different path (= renamed/moved)
SELECT * FROM media_files WHERE fingerprint = ? AND file_path != ?
-- If found: UPDATE file_path = new_path (file was renamed, not a new file)
```

### 8. Full-Text Search (requires FTS5 - Plan E)
```sql
-- Create virtual table
CREATE VIRTUAL TABLE media_fts USING fts5(title, overview, content=media, content_rowid=id);

-- Search
SELECT m.* FROM media m
JOIN media_fts ON media_fts.rowid = m.id
WHERE media_fts MATCH ?
ORDER BY rank
LIMIT 20
```

---

## Migration Mapping to Plans

| Migration | Plan | Phase | Tables |
|-----------|------|-------|--------|
| 001 | Setup | Phase 1 (done) | libraries, media (old), progress (old) |
| 002 | A | Phase 2 | ALTER libraries, RECREATE media (+imdb_id), CREATE media_files |
| 003 | A | Phase 2 | series (+imdb_id), seasons, episodes |
| 004 | A | Phase 2 | genres, media_genres, people, credits |
| 005 | A | Phase 3 | scan_jobs |
| 006 | A | Phase 5 | subtitles, audio_tracks |
| 007 | B | Phase 1 | users, user_preferences, user_library_access |
| 008 | B | Phase 2 | refresh_tokens, sessions |
| 009 | B | Phase 3 | DROP progress, CREATE user_data, CREATE user_series_data |
| 010 | F | Phase 1 | activity_log, task_history, webhooks |
| 011 | A | Phase 4 | external_ids |

---

## SQLite-Specific Notes

1. **WAL mode**: Enabled on connection. Allows concurrent reads during writes.
2. **MaxOpenConns(1)**: SQLite handles 1 writer at a time. Multiple readers OK with WAL.
3. **Foreign keys**: Enabled via `_foreign_keys=on` in DSN. Not default in SQLite.
4. **ALTER TABLE limitations**: SQLite < 3.35 can't DROP COLUMN. Migration 002 will recreate `media` table.
5. **No ENUM type**: Use TEXT + CHECK constraints for type columns.
6. **DATETIME**: Stored as TEXT in ISO 8601 format. Use `CURRENT_TIMESTAMP` for defaults.
7. **Boolean**: SQLite has no bool. Use INTEGER 0/1.
