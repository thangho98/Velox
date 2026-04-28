---
name: release
description: Cut a Velox release — bump version files, add app_versions seed migration, update CHANGELOG, run scripts/release.sh (Docker multi-arch + APK + tag + GitHub Release), then overwrite the auto-generated GitHub Release body with the real CHANGELOG section. Use when the user asks to "release", "cắt release", "build release vX.Y.Z", or invokes /release.
---

# Release Velox

Single source of truth for cutting a Velox release. Don't improvise — every release must touch the same set of files in the same order so older clients keep upgrading cleanly.

## Inputs

- `patch` (default if no arg) — bumps `vX.Y.Z` patch. Example: `v0.1.8 → v0.1.9`.
- `minor` — bumps minor, resets patch. Example: `v0.1.9 → v0.2.0`.
- `major` — bumps major, resets minor + patch. Example: `v0.2.0 → v1.0.0`.
- `vX.Y.Z` — explicit version (must start with `v`).

## Pre-flight (do not skip)

Run these in parallel and stop on any failure:

1. `git rev-parse --abbrev-ref HEAD` — must be `main` (refuse if branch).
2. `git status --short` — uncommitted **tracked** changes? Refuse. Untracked junk like `.scratch/`, `*.m3u`, `filter_m3u*.py` is OK.
3. `git tag -l 'v*' --sort=-v:refname | head -1` — confirm latest tag matches what's in source.
4. `command -v docker && command -v gh && command -v git` — all must be installed.
5. `gh auth status` — must be authenticated.
6. `docker buildx ls` — confirm a builder exists for `linux/amd64,linux/arm64`.

If any check fails, surface the exact problem and the fix command. Don't proceed.

## Compute target version

```
LATEST=$(git tag -l 'v*' --sort=-v:refname | head -1)   # e.g. v0.1.8
# parse MAJOR.MINOR.PATCH from ${LATEST#v}, apply bump rule, build NEW="vMAJOR.MINOR.PATCH"
```

Echo a one-line plan and **wait for user confirmation** before any file edits:
```
Plan: v0.1.8 → v0.1.9 (patch). Will edit 5 files, create 1 commit, run scripts/release.sh.
Proceed? (y/N)
```

## File changes (apply in this exact order)

Let `NEW = vX.Y.Z`, `NUM = X.Y.Z`, `CODE = X*10000 + Y*100 + Z` (i.e. `0.1.9 → 109`, `0.2.0 → 200`, `1.0.0 → 10000`). Migration file uses zero-padded `NN_seed_version_XYZ` where `XYZ` is `MAJOR*100 + MINOR*10 + PATCH` for ≤9 (e.g. `018` for `0.1.8`, `019` for `0.1.9`); for ≥10 in any slot, expand naturally (`0.1.10 → 0110`, `1.0.0 → 100`).

### 1. Backend version constant

`backend/cmd/server/main.go`:
```go
const version = "velox vX.Y.Z"   // e.g. "velox v0.1.9"
```

### 2. Android version

`android/app/build.gradle.kts`:
```kotlin
versionCode = CODE       // e.g. 109
versionName = "X.Y.Z"    // e.g. "0.1.9"
```

### 3. Migration file (NEW)

Find next migration number: `ls backend/internal/database/migrate/*.go | grep -E '^[0-9]+_' | sort | tail -1` → bump.

Create `backend/internal/database/migrate/NNN_seed_version_XYZ.go`:
```go
package migrate

import "database/sql"

func upNNN(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO app_versions (platform, version_name, version_code, is_mandatory, release_notes)
		VALUES ('android', 'X.Y.Z', CODE, IS_MANDATORY, 'ONE-LINE SUMMARY')
		ON CONFLICT DO NOTHING;
	`)
	return err
}

func downNNN(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DELETE FROM app_versions WHERE platform = 'android' AND version_code = CODE;
	`)
	return err
}
```

**Rules for `is_mandatory`**:
- `0` — additive features only (new endpoints, new fields, new migrations that older clients ignore). **Default.**
- `1` — breaking change. Older Android clients must upgrade or features will break (renamed endpoints, removed fields, incompatible auth, schema requires fresh client).

Ask the user explicitly which one if it's not obvious from the commits — never assume `1`.

### 4. Register migration

`backend/internal/database/migrate/registry.go` — add the entry, in numeric order:
```go
{
    Version: NNN,
    Name:    "seed_version_XYZ",
    Up:      upNNN,
    Down:    downNNN,
},
```

### 5. CHANGELOG

`CHANGELOG.md` — prepend a new section under `# Changelog`:
```markdown
## vX.Y.Z [YYYY-MM-DD]

<one-paragraph summary: N commits since vPREV, file count, LOC, top 3 highlights>

### Added
…

### Changed
…

### Fixed
…

### Notes
…
```

Generate by running `git log vPREV..HEAD --oneline` and grouping commits by their Power Prefix (`Add`/`Enhance`/`Fix`/`Make`/`Refactor`/`Document`). For each commit, expand into bullet points matching the codebase changes — read the actual diff (`git show <sha>`) for accuracy, don't rely on commit messages alone.

## Commit

```
git add backend/cmd/server/main.go \
        android/app/build.gradle.kts \
        backend/internal/database/migrate/registry.go \
        backend/internal/database/migrate/NNN_seed_version_XYZ.go \
        CHANGELOG.md

git commit -m "$(cat <<'EOF'
Make(release): bump to vX.Y.Z + seed app_versions migration NNN

<one-line WHY: what this release ships, why mandatory/non-mandatory>

- backend/cmd/server: bump version string vPREV -> vNEW
- android: bump versionCode CODE_PREV -> CODE_NEW, versionName PREV -> NEW
- migrate NNN_seed_version_XYZ: insert android app_versions row
  (is_mandatory=0 — additive features) / (is_mandatory=1 — breaking)
- CHANGELOG: detailed vX.Y.Z [DATE] section
EOF
)"
```

## Run scripts/release.sh

```
printf "y\ny\n" | bash scripts/release.sh NEW 2>&1 | tee /tmp/velox-release.log
```

Run **in background** (`run_in_background: true`) — Docker multi-arch + APK assembleRelease takes 15-30 min. Two `y`s skip the "uncommitted untracked" prompt and the "Proceed?" prompt.

The script does:
1. Create git tag `vX.Y.Z`
2. `docker buildx build --push` per platform (amd64, arm64)
3. `docker buildx imagetools create` to merge multi-arch manifest → push `:X.Y.Z` and `:latest`
4. `git push origin vX.Y.Z`
5. `cd android && ./gradlew assembleRelease` → copy APK to root
6. `gh release create vX.Y.Z velox-vX.Y.Z.apk --generate-notes`

## Post-release fixes (do these every time)

### A. Verify the APK isn't stale

The build script's APK-copy step (`find ... -name "*.apk" | head -1`) returns whatever's in `android/app/build/outputs/apk/release/` even if `assembleRelease` failed. Always check:

```
ls -la android/app/build/outputs/apk/release/app-release.apk velox-vX.Y.Z.apk
cat android/app/build/outputs/apk/release/output-metadata.json | grep -E "versionCode|versionName"
```

`output-metadata.json` MUST show `versionCode: CODE_NEW` and `versionName: "X.Y.Z"`. If it doesn't, the build silently failed:
1. Check log for `BUILD FAILED`.
2. Most common cause used to be Android Lint detector crash (`NonNullableMutableLiveDataDetector` / `IncompatibleClassChangeError`). That's now disabled in `android/app/build.gradle.kts` via `lint { disable += "NullSafeMutableLiveData" }` — if it returns, investigate root cause, don't blanket `-x lintVitalAnalyzeRelease`.
3. After fix, rebuild: `cd android && ./gradlew assembleRelease`.
4. Copy + replace upload: `cp android/app/build/outputs/apk/release/app-release.apk velox-vX.Y.Z.apk && gh release upload vX.Y.Z velox-vX.Y.Z.apk --clobber`.

### B. Replace auto-generated release notes

`gh release create --generate-notes` produces a one-liner: `**Full Changelog**: https://github.com/.../compare/vPREV...vNEW`. Replace with the real CHANGELOG section:

```
# Extract the new section from CHANGELOG.md (lines from "## vX.Y.Z" until next "## v" or EOF)
awk '/^## vX\.Y\.Z /,/^## v[^X]/' CHANGELOG.md | sed '$d' > /tmp/velox-vX.Y.Z-notes.md

# Strip the leading "## vX.Y.Z [DATE]" header (GitHub already shows the title)
sed -i '' '1d' /tmp/velox-vX.Y.Z-notes.md

# Push to GitHub
gh release edit vX.Y.Z --notes-file /tmp/velox-vX.Y.Z-notes.md
```

Verify: `gh release view vX.Y.Z --json body -q '.body[0:200]'` — first 200 chars must match the CHANGELOG section, not the one-line auto note.

## Final report

Print the user a short status block:
```
✅ Tag:      vX.Y.Z pushed
✅ Docker:   doublefeel/velox:X.Y.Z + :latest (linux/amd64, linux/arm64)
✅ APK:      velox-vX.Y.Z.apk  <SIZE>  versionCode=CODE versionName=X.Y.Z
✅ Notes:    CHANGELOG vX.Y.Z section pushed to GitHub Release body
🔗 https://github.com/thangho98/Velox/releases/tag/vX.Y.Z
```

Then offer one concrete follow-up if applicable: deploy to NAS, smoke-test the new APK on a device, monitor docker hub for pull errors. Don't stack offers.

## Things to never do

- Don't run `release.sh` if `git status --short` shows tracked-file changes (untracked junk is OK).
- Don't bump `is_mandatory=1` without explicit user confirmation — it forces every old Android client to update.
- Don't skip the post-release notes step. The `--generate-notes` placeholder is unhelpful for users.
- Don't commit `velox-vX.Y.Z.apk` to git. It belongs only on the GitHub Release.
- Don't `--no-verify` the commit. Pre-commit hooks run `gofmt`, `prettier`, `ktlintFormat`, `detekt` — let them run.
- Don't tag without pushing the commit first. The `release.sh` script tags `HEAD`, so make sure `HEAD` is the version-bump commit, not work-in-progress.
