
# HANDOVER DOCUMENT

**Date:** 2026-03-28

## Current State: Pretranscode separate volume + pending v0.1.3 release

### Done this session:
- Diagnosed Docker `exec format error` on NAS (arm64 image on amd64 host)
- Rebuilt v0.1.2 multi-arch (amd64+arm64) from Linux, pushed to Docker Hub
- Added `VELOX_PRETRANSCODE_DIR` env var — pretranscode output mountable to separate volume
  - config.go: `PretranscodePath` field, reads `VELOX_PRETRANSCODE_DIR` (default: `$VELOX_DATA_DIR/pretranscode`)
  - pretranscode.go: `outputBaseDir` replaces `dataDir`, `OutputDir()` returns it directly
  - main.go: passes `cfg.PretranscodePath` instead of `cfg.DataDir`
  - Dockerfile entrypoint: creates `$VELOX_PRETRANSCODE_DIR` directory
- Created git tag v0.1.3 (not yet built/pushed)

### Pending:
1. **Build & push v0.1.3** — `docker buildx build --platform linux/amd64,linux/arm64 --tag doublefeel/velox:0.1.3 --tag doublefeel/velox:latest --build-arg VERSION="0.1.3" --push .`
2. **NAS docker-compose.yml** — add `platform: linux/amd64` to prevent wrong arch pull
3. Plan M: Search, Filter & Folder Browser
4. Plan N: i18n remaining pages
5. Test on-demand remux end-to-end

### Important notes for next session:
- Tag v0.1.3 already created locally, just need buildx push
- NAS Docker pulls arm64 from multi-arch manifest — always use `platform: linux/amd64` in compose
- VELOX_PRETRANSCODE_DIR defaults to $VELOX_DATA_DIR/pretranscode (backwards compatible)
- Docker Hub: doublefeel/velox:0.1.2 + latest (rebuilt from Linux)
- marker_test.go has pre-existing compile error (missing wsHub arg) — unrelated to our changes
