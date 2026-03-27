#!/usr/bin/env bash
set -euo pipefail

# ── Velox Release Script ─────────────────────────────────────────────
# Usage:
#   ./scripts/release.sh              # auto-bump patch (v0.1.0 → v0.1.1)
#   ./scripts/release.sh patch        # same as above
#   ./scripts/release.sh minor        # v0.1.1 → v0.2.0
#   ./scripts/release.sh major        # v0.2.0 → v1.0.0
#   ./scripts/release.sh v0.3.0       # explicit version
# ─────────────────────────────────────────────────────────────────────

DOCKER_REPO="doublefeel/velox"
PLATFORMS="linux/amd64,linux/arm64"

# ── Colors ───────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${CYAN}▸${NC} $1"; }
ok()    { echo -e "${GREEN}✓${NC} $1"; }
warn()  { echo -e "${YELLOW}!${NC} $1"; }
fail()  { echo -e "${RED}✗${NC} $1"; exit 1; }

# ── Pre-flight checks ───────────────────────────────────────────────
command -v docker >/dev/null || fail "docker not found"
command -v git >/dev/null    || fail "git not found"

# Must be in repo root
cd "$(git rev-parse --show-toplevel)" || fail "not a git repository"

# Check for uncommitted changes
if [[ -n "$(git status --porcelain)" ]]; then
    warn "You have uncommitted changes:"
    git status --short
    echo ""
    read -p "Continue anyway? (y/N) " -r
    [[ $REPLY =~ ^[Yy]$ ]] || exit 1
fi

# ── Determine version ───────────────────────────────────────────────
LATEST_TAG=$(git tag -l 'v*' --sort=-v:refname | head -1)
if [[ -z "$LATEST_TAG" ]]; then
    LATEST_TAG="v0.0.0"
    info "No existing tags found, starting from v0.0.0"
else
    info "Latest tag: $LATEST_TAG"
fi

# Parse current version
IFS='.' read -r MAJOR MINOR PATCH <<< "${LATEST_TAG#v}"

BUMP="${1:-patch}"
case "$BUMP" in
    patch)
        PATCH=$((PATCH + 1))
        NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
        ;;
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
        ;;
    v*)
        NEW_VERSION="$BUMP"
        ;;
    *)
        fail "Invalid bump type: $BUMP (use patch|minor|major|vX.Y.Z)"
        ;;
esac

echo ""
echo -e "  ${CYAN}Release Plan${NC}"
echo -e "  ─────────────────────────────"
echo -e "  Version:    ${GREEN}${NEW_VERSION}${NC}"
echo -e "  Docker:     ${DOCKER_REPO}:${NEW_VERSION#v}"
echo -e "              ${DOCKER_REPO}:latest"
echo -e "  Platforms:  ${PLATFORMS}"
echo -e "  ─────────────────────────────"
echo ""
read -p "Proceed? (y/N) " -r
[[ $REPLY =~ ^[Yy]$ ]] || exit 0

# ── Step 1: Create git tag ───────────────────────────────────────────
info "Creating git tag ${NEW_VERSION}..."
git tag -a "$NEW_VERSION" -m "Release ${NEW_VERSION}"
ok "Tag ${NEW_VERSION} created"

# ── Step 2: Build & push multi-arch image ────────────────────────────
VERSION_NUM="${NEW_VERSION#v}"

info "Building multi-arch image (${PLATFORMS})..."
info "This may take a few minutes..."
echo ""

docker buildx build \
    --platform "$PLATFORMS" \
    --tag "${DOCKER_REPO}:${VERSION_NUM}" \
    --tag "${DOCKER_REPO}:latest" \
    --build-arg VERSION="${VERSION_NUM}" \
    --push \
    .

echo ""
ok "Pushed ${DOCKER_REPO}:${VERSION_NUM}"
ok "Pushed ${DOCKER_REPO}:latest"

# ── Step 3: Push git tag ─────────────────────────────────────────────
info "Pushing tag to remote..."
git push origin "$NEW_VERSION" 2>/dev/null || warn "Failed to push tag (no remote or auth issue)"

# ── Done ─────────────────────────────────────────────────────────────
echo ""
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  ${GREEN}  Release ${NEW_VERSION} complete!${NC}"
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Docker Hub: https://hub.docker.com/r/${DOCKER_REPO}"
echo ""
echo -e "  Users can pull with:"
echo -e "    ${CYAN}docker pull ${DOCKER_REPO}:${VERSION_NUM}${NC}"
echo -e "    ${CYAN}docker pull ${DOCKER_REPO}:latest${NC}"
echo ""
