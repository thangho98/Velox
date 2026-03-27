<p align="center">
  <img src="https://img.shields.io/docker/v/doublefeel/velox?sort=semver&label=version&color=e50914" alt="Version">
  <img src="https://img.shields.io/docker/pulls/doublefeel/velox?color=e50914" alt="Docker Pulls">
  <img src="https://img.shields.io/docker/image-size/doublefeel/velox/latest?color=e50914" alt="Image Size">
  <img src="https://img.shields.io/github/license/doublefeel/velox?color=e50914" alt="License">
</p>

<h1 align="center">Velox</h1>
<p align="center">A lightweight, self-hosted media server for movies and TV shows.<br>Think Jellyfin/Emby — but faster, simpler, and built from scratch.</p>

---

## Features

**Playback**
- Direct Play with automatic codec detection
- HLS transcoding with adaptive bitrate (ABR)
- Hardware acceleration — Intel VAAPI, NVIDIA NVENC, AMD VAAPI, Apple VideoToolbox
- Pre-transcode — Netflix-style offline encoding for instant playback
- On-demand remux — realtime transcode output cached as pre-transcode for next time
- Netflix-style quality selector (Original / 1080p / 720p / 480p / Auto)
- Skip Intro & Credits detection (chapter regex, audio fingerprint, black frame)
- Chromecast support

**Library**
- Automatic media scanning with file watcher
- TMDb metadata with poster/backdrop/logo artwork
- Multi-provider enrichment (OMDb, TheTVDB, Fanart.tv, TVmaze)
- NFO file support (read & write)
- Series grouping with season/episode organization
- Metadata editor with image upload

**Subtitles**
- Multi-provider search (OpenSubtitles, Subdl, BSPlayer, Podnapisi, Subscene)
- Auto-download on library scan
- Translation via DeepL / Google Translate
- External subtitle file support (SRT, ASS, VTT)

**Users & Admin**
- Multi-user with JWT authentication
- Per-user watch progress, favorites, and preferences
- Admin dashboard with server stats
- Activity logging
- Scheduled tasks (library scan, transcode cleanup, session cleanup)
- WebSocket notifications (real-time)
- Webhook support

**UI**
- Netflix-inspired dark theme
- Cinema Mode — trailer autoplay before main feature
- Multi-language support (English, Vietnamese)
- Responsive design (desktop + mobile)

---

## Quick Start

### Docker Compose (recommended)

```yaml
services:
  velox:
    image: doublefeel/velox:latest
    container_name: velox
    restart: unless-stopped
    ports:
      - "8096:80"
    volumes:
      - velox-data:/data
      - /path/to/movies:/media/movies:ro
      - /path/to/tvshows:/media/tv:ro
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/New_York
      - VELOX_HW_ACCEL=auto
    devices:
      - /dev/dri:/dev/dri  # Intel/AMD GPU

volumes:
  velox-data:
```

### Docker Run

```bash
docker run -d \
  --name velox \
  -p 8096:80 \
  -v velox-data:/data \
  -v /path/to/media:/media:ro \
  -e PUID=1000 -e PGID=1000 \
  -e TZ=America/New_York \
  --device /dev/dri:/dev/dri \
  doublefeel/velox:latest
```

Open `http://localhost:8096` and create your admin account.

---

## Hardware Acceleration

Velox auto-detects your GPU. Set `VELOX_HW_ACCEL=auto` (default) or specify:

| GPU | Value | Docker Config |
|-----|-------|---------------|
| Intel (6th gen+) | `vaapi` | `devices: /dev/dri:/dev/dri` |
| NVIDIA | `nvenc` | See [NVIDIA setup](#nvidia-gpu) |
| AMD | `vaapi` | `devices: /dev/dri:/dev/dri` |
| Apple Silicon | `videotoolbox` | Native only (not Docker) |
| None | `none` | Software encoding (libx264) |

### NVIDIA GPU

Install [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html), then:

```yaml
services:
  velox:
    image: doublefeel/velox:latest
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PUID` | `1000` | User ID for file permissions |
| `PGID` | `1000` | Group ID for file permissions |
| `TZ` | `UTC` | Timezone |
| `VELOX_DATA_DIR` | `/data` | Database, cache, and config storage |
| `VELOX_HW_ACCEL` | `auto` | Hardware acceleration (`auto`/`vaapi`/`nvenc`/`qsv`/`videotoolbox`/`none`) |
| `VELOX_MAX_TRANSCODES` | `2` | Maximum concurrent transcode sessions |
| `VELOX_FILE_WATCHER` | `true` | Auto-detect new/changed media files |
| `VELOX_TRICKPLAY_ENABLED` | `false` | Generate thumbnail preview strips |
| `VELOX_TMDB_API_KEY` | | TMDb API key (optional, built-in available) |
| `VELOX_OMDB_API_KEY` | | OMDb API key for ratings enrichment |
| `VELOX_TVDB_API_KEY` | | TheTVDB API key for TV metadata |
| `VELOX_FANART_API_KEY` | | Fanart.tv API key for artwork |
| `VELOX_SUBDL_API_KEY` | | Subdl API key for subtitles |
| `SUBSCENE_ENABLED` | `false` | Enable Subscene scraper (requires Chrome) |

---

## Architecture

```
                    ┌─────────┐
                    │  nginx  │ :80
                    └────┬────┘
              ┌──────────┴──────────┐
              │                     │
        ┌─────┴─────┐        ┌─────┴─────┐
        │  React SPA │        │  Go API   │ :8080
        │  (Vite)    │        │  (stdlib) │
        └────────────┘        └─────┬─────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
              ┌─────┴─────┐  ┌─────┴─────┐  ┌─────┴─────┐
              │  SQLite    │  │  FFmpeg   │  │ WebSocket │
              │  (WAL)     │  │ transcode │  │   hub     │
              └────────────┘  └───────────┘  └───────────┘
```

| Layer | Technology |
|-------|-----------|
| Frontend | React 19, TypeScript, Vite 8, TailwindCSS 4 |
| Backend | Go 1.26, stdlib `net/http` (Go 1.22+ routing) |
| Database | SQLite (WAL mode, 27 migrations) |
| Transcoding | FFmpeg 8.0 with HW acceleration |
| Container | Alpine 3.21, nginx, tini, multi-stage Docker build |

---

## Pre-transcode (Offline Encoding)

Encode your library in advance for instant playback — zero buffering:

1. Go to **Settings > Pre-transcode**
2. Enable and select quality profiles (480p / 720p / 1080p / 1440p / 4K)
3. Click **Start Encoding**

Pre-encoded files are served instantly (`<100ms`) via `http.ServeContent`. When a user watches at a quality that was previously transcoded in realtime, the output is automatically remuxed into a pre-transcode file for next time.

Quality profiles:

| Profile | Resolution | Video Bitrate | Audio |
|---------|-----------|---------------|-------|
| 480p | 854x480 | 1.5 Mbps | 128k AAC |
| 720p | 1280x720 | 4 Mbps | 128k AAC |
| 1080p | 1920x1080 | 8 Mbps | 192k AAC |
| 1440p | 2560x1440 | 16 Mbps | 192k AAC |
| 4K | 3840x2160 | 40 Mbps | 256k AAC |

---

## Skip Intro / Credits

Velox detects intro and credits segments using three methods:

1. **Chapter markers** — Extracted from video metadata during scan
2. **Audio fingerprint** — Compares episodes via chromaprint to find recurring segments
3. **Black frame detection** — Detects credits by finding black frames + silence

Run detection from **Settings > Skip Intro / Credits** with real-time WebSocket progress.

---

## Synology / NAS Setup

Velox runs great on Synology NAS (DS920+, DS923+) and Xpenology:

```yaml
environment:
  - PUID=1026        # Synology default user ID
  - PGID=100         # Synology users group
  - VELOX_HW_ACCEL=vaapi
  - VELOX_MAX_TRANSCODES=2
devices:
  - /dev/dri:/dev/dri
```

---

## Development

### Prerequisites
- Go 1.26+
- Node.js 22+
- FFmpeg 8.0+
- SQLite3

### Backend
```bash
cd backend
make dev          # go run ./cmd/server
make build        # go build -o bin/velox
make test         # go test ./... -v
make lint         # golangci-lint
```

### Frontend
```bash
cd webapp
npm run dev       # Vite dev server (port 3000)
npm run build     # Production build
npm run lint      # ESLint
```

### Docker
```bash
docker build -t velox:latest .
docker compose up -d
```

### Release
```bash
./scripts/release.sh patch    # v0.1.0 → v0.1.1
./scripts/release.sh minor    # v0.1.1 → v0.2.0
./scripts/release.sh major    # v0.2.0 → v1.0.0
./scripts/release.sh v1.0.0   # explicit version
```

---

## API

All responses follow the format:
```json
{ "data": { ... } }        // success
{ "error": "message" }     // error
```

Authentication via JWT Bearer token or API key (for streaming).

---

## Contributing

Contributions are welcome! Please open an issue first to discuss what you'd like to change.

## License

[MIT](LICENSE)

---

<p align="center">
  Built with Go + React + FFmpeg
</p>
