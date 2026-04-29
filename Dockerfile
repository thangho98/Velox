# ============================================================
# Velox — Production Multi-stage Dockerfile
# Supports: x86_64 (Intel/AMD), ARM64 (Raspberry Pi/Mac M-series)
# ============================================================

# ----- Stage 1: Frontend build -----
FROM node:22-bullseye-slim AS frontend

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /build
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY webapp/package.json webapp/
COPY packages/shared/package.json packages/shared/
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --ignore-scripts
COPY packages/shared/ packages/shared/
COPY webapp/ webapp/
ARG VITE_DEBUG=false
ENV VITE_DEBUG=$VITE_DEBUG
RUN cd webapp && pnpm run build


# ----- Stage 2: Backend build -----
FROM golang:1.24-bookworm AS backend

# CGO required for mattn/go-sqlite3
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc-dev

# go.mod requires 1.26 — let Go auto-download the right toolchain
ENV GOTOOLCHAIN=auto

WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY backend/ ./

# Build static binary with CGO
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=docker" \
    -o /velox ./cmd/server


# ----- Stage 3: Production runtime -----
FROM debian:bookworm-slim

LABEL maintainer="thawng"
LABEL org.opencontainers.image.title="Velox"
LABEL org.opencontainers.image.description="Self-hosted home media server"

# Runtime deps:
#   jellyfin-ffmpeg — transcoding + media probe (via official Jellyfin APT repo)
#   python3 + pip   — Subscene subtitle scraper (DrissionPage)
#   chromium + xvfb — headless browser for Cloudflare bypass
#   font packages   — subtitle rendering (burn-in)
#   nginx           — serve frontend SPA + reverse proxy API
#   tzdata          — timezone support
#   gosu            — run as non-root (su-exec equivalent)
RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends \
    curl \
    gnupg \
    ca-certificates \
    python3 \
    python3-pip \
    python3-venv \
    chromium \
    xvfb \
    fonts-noto \
    fonts-noto-cjk \
    nginx \
    tzdata \
    gosu \
    tini \
    sqlite3

# Install jellyfin-ffmpeg7 directly from Jellyfin's official repository
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    mkdir -p /etc/apt/keyrings && \
    curl -fsSL https://repo.jellyfin.org/jellyfin_team.gpg.key | gpg --dearmor -o /etc/apt/keyrings/jellyfin.gpg && \
    echo "deb [arch=$( dpkg --print-architecture ) signed-by=/etc/apt/keyrings/jellyfin.gpg] https://repo.jellyfin.org/debian bookworm main" | tee /etc/apt/sources.list.d/jellyfin.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends jellyfin-ffmpeg7 && \
    apt-get remove -y gnupg && \
    apt-get autoremove -y

# GPU drivers — x86_64 only (Intel VAAPI, AMD VAAPI)
# Note: intel-media-va-driver-non-free is the recommended iHD driver
ARG TARGETARCH
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    if [ "$TARGETARCH" = "amd64" ]; then \
        sed -i 's/Components: main/Components: main non-free non-free-firmware/' /etc/apt/sources.list.d/debian.sources && \
        apt-get update && apt-get install -y --no-install-recommends \
        intel-media-va-driver-non-free \
        va-driver-all \
        mesa-va-drivers; \
    fi

# Create velox user (UID/GID configurable at runtime)
RUN groupadd -g 1000 velox && \
    useradd -r -u 1000 -g velox -d /app velox

WORKDIR /app

# Python venv for Subscene scraper
COPY backend/scripts/requirements.txt /app/scripts/requirements.txt
RUN python3 -m venv /app/scripts/.venv
RUN --mount=type=cache,target=/root/.cache/pip \
    /app/scripts/.venv/bin/pip install -r /app/scripts/requirements.txt

# Copy Subscene scraper
COPY backend/scripts/subscene_search.py /app/scripts/subscene_search.py

# Copy backend binary
COPY --from=backend /velox /app/velox

# Copy frontend build
COPY --from=frontend /build/webapp/dist /app/webapp

# Nginx config — SPA + API reverse proxy
RUN cat > /etc/nginx/sites-available/velox.conf <<'NGINX'
server {
    listen 80;
    server_name _;

    # Trust X-Forwarded-For from private networks (Docker bridges, LAN, loopback)
    # so $remote_addr reflects the original client IP, not the upstream proxy.
    # Public-internet clients hitting the container directly cannot spoof these
    # headers because their source IP wouldn't match these ranges.
    set_real_ip_from 10.0.0.0/8;
    set_real_ip_from 172.16.0.0/12;
    set_real_ip_from 192.168.0.0/16;
    set_real_ip_from 127.0.0.1;
    set_real_ip_from ::1;
    real_ip_header X-Forwarded-For;
    real_ip_recursive on;

    # Frontend SPA
    root /app/webapp;
    index index.html;

    # Gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 1000;

    # Cache static assets (Vite content-hashed filenames)
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Never cache index.html — ensures browser loads latest JS/CSS bundles
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
    }

    # WebSocket endpoint
    location /api/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        # $http_host preserves the original Host header including port, so
        # backend-built URLs (POST /api/stream/{id}/url) keep :8098 etc.
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }

    # API + streaming → backend
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Streaming support
        proxy_buffering off;
        proxy_request_buffering off;
        proxy_http_version 1.1;
        proxy_set_header Connection "";

        # Long timeout for transcoding
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        client_max_body_size 0;
    }

    # SPA fallback — all non-file routes → index.html
    location / {
        try_files $uri $uri/ /index.html;
    }
}
NGINX

# Enable nginx site and disable default
RUN rm -f /etc/nginx/sites-enabled/default && \
    ln -s /etc/nginx/sites-available/velox.conf /etc/nginx/sites-enabled/velox.conf

# Entrypoint script
RUN cat > /app/entrypoint.sh <<'ENTRYPOINT'
#!/bin/sh
set -e

# ---- Create data directories ----
VELOX_DATA_DIR=${VELOX_DATA_DIR:-/data}
export VELOX_DATA_DIR

VELOX_PRETRANSCODE_DIR=${VELOX_PRETRANSCODE_DIR:-$VELOX_DATA_DIR/pretranscode}
export VELOX_PRETRANSCODE_DIR

mkdir -p "$VELOX_DATA_DIR" \
         "$VELOX_DATA_DIR/subtitles" \
         "$VELOX_DATA_DIR/transcode" \
         "$VELOX_DATA_DIR/trickplay" \
         "$VELOX_PRETRANSCODE_DIR"

# ---- Clear realtime transcode cache on startup ----
# HLS transcode segments are temporary — stale cache from prior versions
# can cause playback failures. Pretranscode files are NOT affected.
rm -rf "$VELOX_DATA_DIR/transcode/"*

# ---- VAAPI: default to iHD driver (Intel 6th gen+) ----
# Override via LIBVA_DRIVER_NAME env var in docker-compose if needed
export LIBVA_DRIVER_NAME=${LIBVA_DRIVER_NAME:-iHD}

# ---- Set Chromium path for DrissionPage ----
export CHROME_PATH=/usr/bin/chromium

# ---- Start Xvfb for Subscene scraper (virtual display) ----
if [ "${SUBSCENE_ENABLED:-true}" = "true" ]; then
    Xvfb :99 -screen 0 1280x720x16 -nolisten tcp &
    export DISPLAY=:99
fi

# ---- Start nginx (frontend) ----
nginx

# ---- Run backend ----
echo "Starting Velox (DATA=$VELOX_DATA_DIR)"
exec /app/velox
ENTRYPOINT
RUN chmod +x /app/entrypoint.sh

# ---- Volumes ----
VOLUME ["/data"]

# ---- Ports ----
# 80: nginx (frontend + API proxy)
EXPOSE 80

# ---- Health check ----
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

# ---- Use tini as PID 1 (proper signal handling) ----
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/app/entrypoint.sh"]
