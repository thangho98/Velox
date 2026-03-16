# ============================================================
# Velox — Production Multi-stage Dockerfile
# Supports: x86_64 (Intel/AMD), ARM64 (Raspberry Pi, Synology)
# ============================================================

# ----- Stage 1: Frontend build -----
FROM node:22-alpine AS frontend

WORKDIR /build/webapp
COPY webapp/package.json webapp/package-lock.json* ./
RUN npm ci --ignore-scripts
COPY webapp/ ./
RUN npm run build


# ----- Stage 2: Backend build -----
FROM golang:1.24-alpine AS backend

# CGO required for mattn/go-sqlite3
RUN apk add --no-cache gcc musl-dev

# go.mod requires 1.26 — let Go auto-download the right toolchain
ENV GOTOOLCHAIN=auto

WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./

# Build static binary with CGO
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=docker" \
    -o /velox ./cmd/server


# ----- Stage 3: Production runtime -----
FROM alpine:3.21

LABEL maintainer="thawng"
LABEL org.opencontainers.image.title="Velox"
LABEL org.opencontainers.image.description="Self-hosted home media server"

# Runtime deps:
#   ffmpeg/ffprobe  — transcoding + media probe
#   python3 + pip   — Subscene subtitle scraper (DrissionPage)
#   chromium + xvfb — headless browser for Cloudflare bypass
#   font packages   — subtitle rendering (burn-in)
#   nginx           — serve frontend SPA + reverse proxy API
#   tzdata          — timezone support
#   su-exec         — run as non-root (lightweight gosu alternative)
RUN apk add --no-cache \
    ffmpeg \
    chromaprint \
    python3 \
    py3-pip \
    chromium \
    xvfb-run \
    font-noto \
    font-noto-cjk \
    nginx \
    tzdata \
    su-exec \
    curl \
    tini

# Create velox user (UID/GID configurable at runtime)
RUN addgroup -g 1000 velox && \
    adduser -D -u 1000 -G velox -h /app velox

WORKDIR /app

# Python venv for Subscene scraper
COPY backend/scripts/requirements.txt /app/scripts/requirements.txt
RUN python3 -m venv /app/scripts/.venv && \
    /app/scripts/.venv/bin/pip install --no-cache-dir -r /app/scripts/requirements.txt

# Copy Subscene scraper
COPY backend/scripts/subscene_search.py /app/scripts/subscene_search.py

# Copy backend binary
COPY --from=backend /velox /app/velox

# Copy frontend build
COPY --from=frontend /build/webapp/dist /app/webapp

# Nginx config — SPA + API reverse proxy
RUN cat > /etc/nginx/http.d/velox.conf <<'NGINX'
server {
    listen 80;
    server_name _;

    # Frontend SPA
    root /app/webapp;
    index index.html;

    # Gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 1000;

    # Cache static assets
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # API + streaming → backend
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
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

# Remove default nginx config
RUN rm -f /etc/nginx/http.d/default.conf

# Entrypoint script
RUN cat > /app/entrypoint.sh <<'ENTRYPOINT'
#!/bin/sh
set -e

# ---- UID/GID remapping (Synology NAS friendly) ----
# Synology DSM often uses UID=1026, GID=100 (users group)
# Set PUID/PGID env vars to match your NAS user
PUID=${PUID:-1000}
PGID=${PGID:-1000}

if [ "$PUID" != "1000" ] || [ "$PGID" != "1000" ]; then
    echo "Remapping velox user to UID=$PUID GID=$PGID"
    deluser velox 2>/dev/null || true
    delgroup velox 2>/dev/null || true
    addgroup -g "$PGID" velox
    adduser -D -u "$PUID" -G velox -h /app velox
fi

# ---- Create data directories ----
VELOX_DATA_DIR=${VELOX_DATA_DIR:-/data}
export VELOX_DATA_DIR

mkdir -p "$VELOX_DATA_DIR" \
         "$VELOX_DATA_DIR/subtitles" \
         "$VELOX_DATA_DIR/transcode" \
         "$VELOX_DATA_DIR/trickplay"

chown -R "$PUID:$PGID" "$VELOX_DATA_DIR"
chown -R "$PUID:$PGID" /app/scripts

# ---- Set Chromium path for DrissionPage ----
export CHROME_PATH=/usr/bin/chromium-browser

# ---- Start Xvfb for Subscene scraper (virtual display) ----
if [ "${SUBSCENE_ENABLED:-true}" = "true" ]; then
    Xvfb :99 -screen 0 1280x720x16 -nolisten tcp &
    export DISPLAY=:99
fi

# ---- Start nginx (frontend) ----
nginx

# ---- Run backend as velox user ----
echo "Starting Velox (UID=$PUID, GID=$PGID, DATA=$VELOX_DATA_DIR)"
exec su-exec "$PUID:$PGID" /app/velox
ENTRYPOINT
RUN chmod +x /app/entrypoint.sh

# ---- Volumes ----
VOLUME ["/data", "/media"]

# ---- Ports ----
# 80: nginx (frontend + API proxy)
EXPOSE 80

# ---- Health check ----
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

# ---- Use tini as PID 1 (proper signal handling) ----
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/entrypoint.sh"]
