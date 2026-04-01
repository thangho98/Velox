# Plan Q: Refactor God Files

Created: 2026-03-31
Status: ⬜ Pending

## Overview

5 Go files exceed 800 lines, violating single-responsibility principle.
Goal: split each into focused files under 500 lines. Pure structural refactor — no behavior changes, no new logic.

## Approach

Split by **logical concern grouping**. Each new file keeps the same package, same struct receiver, just moves functions to a separate file. Go allows methods on a type to span multiple files in the same package — no interface changes needed.

## Rules
1. **No behavior changes** — only move functions between files
2. **Same package** — all split files stay in same Go package
3. **Shared state** — struct fields, mutexes, constants stay in the main file
4. **Imports** — each new file gets only the imports it needs
5. **Build + vet** after each phase before moving to next

---

## Phase 1: `repository/media.go` (857 lines → 4 files)

Already contains 2 distinct repo types (`MediaRepo` + `MediaFileRepo`) plus browse types. Cleanest split.

| New File | Contents | ~Lines |
|----------|----------|--------|
| `media.go` (keep) | `MediaRepo`: constructor, `scanMedia`, CRUD, `UpdateMetadata`, `UpdateImagePath`, `SetMetadataLocked`, `UpdateOMDbRatings`, `UpdateTitle`, `Delete`, `FirstPosterBy*` | ~300 |
| `media_query.go` | `MediaRepo`: `List`, `Search`, `GetByTmdbID`, `GetByImdbID`, `ListWithIMDbID`, `ListWithGenres`, `ListFiltered` | ~200 |
| `media_file.go` | `MediaFileRepo`: constructor, `scanMediaFile`, all CRUD + queries | ~250 |
| `media_browse.go` | `MediaFileRepo.BrowseFolders` + `BrowseResult`/`BrowseFolderItem` types | ~150 |

**Key:** `mediaColumns` const stays in `media.go`. Type definitions move with their primary repo.

---

## Phase 2: `service/pretranscode.go` (934 lines → 3 files)

Clean 3-way split: scheduler loop / job worker / admin operations.

| New File | Contents | ~Lines |
|----------|----------|--------|
| `pretranscode.go` (keep) | Struct, constructor, setters, `Start`, `Stop`, `Pause`, `Resume`, `schedulerLoop`, `isInSchedule`, `recoverInterruptedJobs` | ~210 |
| `pretranscode_worker.go` | `processJob`, `pickAudioRemuxJob`, `EnqueueAudioRemux`, `shouldSkipEncode`, `runAudioRemux`, `runUniversalTranscode*`, `encodeFile`, `runFFmpeg`, `estimateBitrateForHeight` | ~360 |
| `pretranscode_admin.go` | `EnqueueLibrary`, `EnqueueAllLibraries`, `CancelAll`, `GetStatus*`, `EstimateStorage`, `CleanupAll`, `CleanupByMediaFile`, `List*`, `Get/SetProfile*`, `RemuxFromHLS`, `FindBestPretranscode` | ~370 |

---

## Phase 3: `transcoder/transcoder.go` (1,148 lines → 4 files)

Split by: HLS generation / ABR / encoding helpers / utilities.

| New File | Contents | ~Lines |
|----------|----------|--------|
| `transcoder.go` (keep) | Struct, `New`, `ActiveCount`, `TryActiveCount`, slot management, `isHLSComplete`, `waitForFirstSegment`, `SegmentPath`, `WaitForSegment`, `hasActiveJobInDir`, `MasterPlaylistPath`, `ABRMasterPath`, `ABRCached`, `Clean`, `CleanupOlderThan`, `RemuxToWriter`, path/offset helpers | ~300 |
| `transcoder_hls.go` | `GenerateHLS`, `GenerateHLSWithAudio`, `startHLSBackground`, `startHLSAndWaitComplete`, `runHLSFFmpeg`, `runMultiOutputHLS`, `writeMasterPlaylistWithAudio` | ~350 |
| `transcoder_abr.go` | `GenerateABRHLS`, `generateABRVariants`, `generateABRVariant`, `writeABRMasterPlaylist` | ~170 |
| `transcoder_encoding.go` | All `hw*` functions, `build*Args`, `buildImageSubtitleBurnIn*`, `escapeFFmpegSubtitlePath`, `isHDRFile`, `hdrToneMapFilter`, `detectSubtitlesFilter`, `SupportsSubtitleBurnIn` | ~250 |

---

## Phase 4: `handler/playback.go` (810 lines → 3 files)

Keep main handler + types in place; extract helpers to separate files within handler package.

| New File | Contents | ~Lines |
|----------|----------|--------|
| `playback.go` (keep) | Struct, constructor, types, `GetPlaybackInfo`, `applyAdminPlaybackPolicy`, `cloneValues`, `buildURLWithQuery` | ~480 |
| `playback_quality.go` | `buildQualityOptions`, `resolveSelectedAudioTrackID`, `GetClientCapabilities`, `applyClientCapabilityOverrides` | ~140 |
| `playback_tracks.go` | `normalizeCapabilityValues`, `normalizeLanguageCode`, `normalizeContainerValue`, `languageMatches`, `findSubtitleByLanguage`, `findSubtitleByID`, `filterPlayableSubtitles`, `playbackModeQuery` | ~190 |

---

## Phase 5: `service/metadata.go` (1,477 lines → 4 files)

Largest file, most functions. Split by domain concern.

| New File | Contents | ~Lines |
|----------|----------|--------|
| `metadata.go` (keep) | Struct, constructor, setters, `MatchAndPersistMovie`, `MatchAndPersistEpisode`, `findOrCreateSeries`, `ensureSeries`, `findOrCreateSeason`, `linkEpisode`, `updateEpisodeLink`, `GetSeries` | ~430 |
| `metadata_enrichment.go` | `enrichOMDbRatings`, `enrichTVDBSeries`, `enrichFanartMovie`, `enrichFanartShow`, `enrichTVmazeSeries` | ~210 |
| `metadata_credits.go` | All `sync*Genres`, `sync*Credits`, `ensure*Genre`, `ensure*Person` functions (both TMDb and input variants) | ~340 |
| `metadata_editor.go` | `IdentifyByTmdbID`, `RefreshMetadata`, `refreshEpisodeMetadata`, `AutoMatchAndRefresh`, `BulkRefreshAllMetadata`, `Edit*Metadata`, `Unlock*`, `UpdateImage*`, `Write*NFO`, `build*NFOData` | ~500 |

---

## Verification

After each phase:
```sh
cd backend && go build ./... && go vet ./...
```

After all phases:
```sh
cd backend && make test
```

## Summary

| Phase | File | Before | After | New Files |
|-------|------|--------|-------|-----------|
| 1 | `repository/media.go` | 857 | 4 files (150-300 each) | +3 |
| 2 | `service/pretranscode.go` | 934 | 3 files (210-370 each) | +2 |
| 3 | `transcoder/transcoder.go` | 1,148 | 4 files (170-350 each) | +3 |
| 4 | `handler/playback.go` | 810 | 3 files (140-480 each) | +2 |
| 5 | `service/metadata.go` | 1,477 | 4 files (210-500 each) | +3 |
| **Total** | | **5,226** | **18 files** | **+13** |
