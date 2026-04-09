# FFmpeg HLS Transcoding Reference

Comprehensive reference for FFmpeg parameters used in HLS transcoding for media server applications (Velox, Jellyfin, Emby, Plex).

---

## Table of Contents

1. [Input Options (before -i)](#1-input-options-before--i)
2. [Timestamp/Sync Options](#2-timestampsync-options)
3. [Video Copy Options](#3-video-copy-options)
4. [Audio Transcode Options](#4-audio-transcode-options)
5. [HLS Muxer Options (-f hls)](#5-hls-muxer-options--f-hls)
6. [Muxer/General Options](#6-muxergeneral-options)
7. [Common Use Cases](#7-common-use-cases-with-exact-commands)
8. [Known Pitfalls & Conflicts](#8-known-pitfalls--conflicts)
9. [Jellyfin vs Emby vs Plex](#9-what-jellyfin-vs-emby-vs-plex-use)

---

## 1. Input Options (before -i)

### `-ss position` (Seek)

Seeks to position in the input file before reading. Behavior differs dramatically based on placement relative to `-i`:

**Before `-i` (Input Seeking) -- FAST:**
```
ffmpeg -ss 01:23:45 -i input.mkv ...
```
- Seeks by keyframe in the demuxer -- extremely fast even for multi-GB files.
- Since FFmpeg 2.1, when **transcoding** with `-accurate_seek` (default), the extra segment between the nearest keyframe and the exact seek position is decoded and discarded. Result is frame-accurate.
- When **stream copying** (`-c:v copy`), FFmpeg can only cut at keyframes. If `-ss 01:30:00` is specified but the nearest keyframe is at `01:29:58`, the output starts there. No frame-accurate seeking is possible with copy mode.
- When `-noaccurate_seek` is used, the extra segment between keyframe and position is preserved (included in output).

**After `-i` (Output Seeking) -- SLOW:**
```
ffmpeg -i input.mkv -ss 01:23:45 ...
```
- FFmpeg decodes the entire input from the beginning up to the seek position, then discards everything before it. Very slow for long files but guarantees frame-accuracy.
- Rarely used in practice because input seeking with `-accurate_seek` achieves the same accuracy much faster.

**Velox approach:** `-ss` is placed before `-i` for both copy and transcode modes. This is the standard approach used by all major media servers.

### `-t duration` and `-to position`

| Option | As Input Option | As Output Option |
|--------|----------------|-----------------|
| `-t duration` | Limits how much data is read from input | Stops writing after output reaches this duration |
| `-to position` | Stops reading input at this position | Stops writing output at this position |

`-to` and `-t` are mutually exclusive; `-t` takes priority. When `-ss` is also used, `-to` is relative to the start of the file, not relative to `-ss`.

### `-fflags` (Format Flags)

Format-level flags that control demuxer/muxer behavior. Multiple flags are combined with `+`:
```
-fflags +genpts+igndts
```

#### Input Flags

| Flag | Description | When to Use |
|------|-------------|-------------|
| `+genpts` | **Generate missing PTS** from DTS when PTS is absent. | Files with missing presentation timestamps (some MPEG-TS captures, broken recordings). **WARNING:** Conflicts with `-avoid_negative_ts make_zero` -- see Pitfalls section. |
| `+igndts` | **Ignore DTS** when PTS is set (sets DTS to NOPTS). | Files where DTS values are corrupt or inconsistent but PTS is valid. |
| `+discardcorrupt` | **Discard corrupted packets** instead of passing them through. | Damaged recordings, incomplete downloads, network captures with errors. |
| `+fastseek` | **Enable fast but inaccurate seeks** for some formats. | When approximate seeking is acceptable (e.g., thumbnail generation). Trades accuracy for speed. |
| `+nofillin` | **Do not fill missing calculable values** in packet fields. | Advanced debugging. Rarely used in production. |
| `+noparse` | **Disable AVParsers** (requires `+nofillin`). | Advanced debugging only. |
| `+nobuffer` | **Reduce buffering** during initial stream analysis. Reduces latency but may miss some stream info. | Live/realtime input where latency matters more than completeness. |
| `+sortdts` | **Interleave output by DTS** (only works for AVI with index). | AVI files with out-of-order packets. |
| `+ignidx` | **Ignore the index** in container. | Corrupt index in AVI/MKV files. Forces sequential reading. |

#### Output Flags

| Flag | Description | When to Use |
|------|-------------|-------------|
| `+bitexact` | Write only platform/build/time-independent data. | Reproducible builds, test suites. |
| `+flush_packets` | Write/flush packets immediately (no buffering). | Low-latency live streaming. |
| `+shortest` | Stop muxing when the shortest stream ends. | When audio and video have different durations. |
| `+autobsf` | Automatically apply required bitstream filters. | Enabled by default. Auto-inserts `h264_mp4toannexb` for MPEG-TS, etc. |

### `-probesize bytes`

Number of bytes to analyze for stream detection. Default: 5,000,000 (5 MB).

- Higher values find streams that are interleaved far apart (e.g., subtitle tracks in large MKV files).
- Jellyfin uses `200M` (200 MB). Velox uses `50000000` (50 MB) for subtitle burn-in, `5000000` (5 MB) for simple copy.
- Trade-off: higher values increase startup latency.

### `-analyzeduration microseconds`

Maximum duration (in microseconds) to analyze for stream detection. Default: 5,000,000 (5 seconds).

- Higher values improve detection of codec parameters for streams with sparse packets.
- Jellyfin uses `200000000` (200 seconds). Velox uses `100000000` (100 seconds) for subtitle burn-in.
- Both `-probesize` and `-analyzeduration` should be set together -- whichever limit is hit first stops analysis.

### `-re` (Realtime)

Read input at native frame rate, simulating a live source. Without this, FFmpeg reads as fast as possible.

- **Use case:** Simulating live streaming from a file (e.g., feeding RTMP). Never used for VOD HLS transcoding -- you want FFmpeg to go as fast as possible.
- Not used by Velox, Jellyfin, Emby, or Plex for VOD HLS.

### `-readrate speed`

Read input at `speed` times realtime. `-readrate 1.0` is equivalent to `-re`. `-readrate 2.0` reads at 2x speed.

- Not commonly used for HLS VOD transcoding.

### `-readrate_initial_burst seconds`

How many seconds of input to read at full speed before `-readrate` pacing kicks in.

- Useful for live streaming where you want to buffer ahead initially. Not used for VOD.

### `-noaccurate_seek`

When `-ss` is used before `-i`, disables frame-accurate seeking. FFmpeg seeks to the nearest preceding keyframe and starts output there, without decoding/discarding the gap.

- **Use case:** Jellyfin uses this with `-c:v copy` for fast seeking without decoding the gap.
- Result: output may start a few seconds before the requested position.

### `-hwaccel` Options

Hardware-accelerated decoding. Placed before `-i`.

| Option | Platform | Command |
|--------|----------|---------|
| VideoToolbox | macOS | `-hwaccel videotoolbox` |
| CUDA/NVDEC | NVIDIA GPU | `-hwaccel cuda` or `-hwaccel cuda -hwaccel_output_format cuda` |
| VAAPI | Linux (Intel/AMD) | `-hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128` |
| QSV | Intel (cross-platform) | `-hwaccel qsv` |

**Key detail:** `-hwaccel_output_format` keeps decoded frames in GPU memory, avoiding a GPU-to-CPU-to-GPU roundtrip when using a matching HW encoder (e.g., `cuda` decode + `h264_nvenc` encode).

**Performance:** Intel QSV H.265-to-H.264 transcoding reduces CPU load from ~90% (libx264) to ~20% (h264_qsv). With HW decode enabled, further reduces to ~4%.

---

## 2. Timestamp/Sync Options

### `-copyts`

**Do not process input timestamps.** Keep original timestamp values without sanitizing them.

- Without `-copyts`: FFmpeg normalizes timestamps so output starts near 0.
- With `-copyts`: original PTS/DTS values from the input are preserved in the output.

**Critical for HLS:** When using with HLS, this interacts differently with MPEG-TS vs fMP4 segments:
- **fMP4:** Timestamp offsets are handled in MOOF/TRAF atoms. Players read the base decode time from the fragment and compute correct presentation times regardless of absolute values. `-copyts` works fine.
- **MPEG-TS:** PTS values are embedded directly in PES packet headers. If the source MKV starts at PTS 107s (common for BluRay REMUX), the output MPEG-TS also starts at 107s, which breaks player seek calculations.

### `-start_at_zero`

**Only meaningful when `-copyts` is also set.** Shifts output timestamps so the first frame starts at 0, while preserving relative timing.

- Without: output retains the input's absolute timestamp values.
- With: subtracts the input file's start time from all timestamps.

**The combination `-copyts -start_at_zero`** is what Jellyfin uses with fMP4 segments. This preserves relative timing (no drift) while ensuring the output timeline starts at 0.

### `-avoid_negative_ts`

Controls how FFmpeg handles negative timestamps in the output. This is applied at the muxer level.

| Value | Behavior |
|-------|----------|
| `auto` (default) | Enable shifting when required by the target format. For MPEG-TS output, this typically enables `make_non_negative`. |
| `make_non_negative` | Shift timestamps so that leading negative timestamps become non-negative. Only fixes initial negative timestamps -- does not normalize the entire timeline. |
| `make_zero` | Shift ALL packet timestamps so the first timestamp is 0. Stronger guarantee than `make_non_negative`. Rewrites timestamps at the muxer level. |
| `disabled` | Do not modify timestamps at all. Output preserves whatever timestamp values arrive from the encoder/decoder. |

**Critical property:** When shifting is enabled, ALL output streams (audio, video, subtitle) are shifted by the SAME amount. Relative A/V sync is preserved.

**Velox approach:** Uses `-avoid_negative_ts make_zero` for MPEG-TS output with `-c:v copy`. This ensures output starts at 0 even with BluRay REMUX files that have non-zero start PTS.

**Jellyfin approach:** Uses `-avoid_negative_ts disabled` with `-copyts -start_at_zero` for fMP4 output. Timestamp normalization is handled by `-start_at_zero` instead.

### `-output_ts_offset offset`

Adds a fixed time offset to all output timestamps. Positive values delay streams.

- Default: 0 (no offset).
- Use case: aligning segments in multi-file HLS workflows, or creating a time gap at the start.
- Rarely needed in standard HLS transcoding.

### Timestamp Option Interaction Matrix

| Combination | MPEG-TS Output | fMP4 Output | Notes |
|-------------|---------------|-------------|-------|
| (none) | FFmpeg normalizes to ~0 | FFmpeg normalizes to ~0 | Default behavior, usually fine |
| `-copyts` alone | **DANGEROUS**: preserves original PTS (e.g., 107s for BluRay) | Works -- fMP4 atoms handle offsets | Breaks MPEG-TS player seek |
| `-copyts -start_at_zero` | PTS shifted to 0 | PTS shifted to 0 | **Jellyfin's approach for fMP4** |
| `-avoid_negative_ts make_zero` | PTS shifted to 0 at muxer | PTS shifted to 0 at muxer | **Velox's approach for MPEG-TS** |
| `-copyts -avoid_negative_ts disabled` | Preserves original PTS | Preserves original PTS | Jellyfin uses this -- relies on `-start_at_zero` |
| `-copyts -start_at_zero -avoid_negative_ts disabled` | Correct (start at 0) | Correct (start at 0) | **Jellyfin's full combination** |

---

## 3. Video Copy Options

### `-c:v copy`

Copies the video bitstream unchanged -- no decoding or encoding. Extremely fast (limited by I/O, not CPU).

**Container compatibility:**

| Source | Target | Issues |
|--------|--------|--------|
| MKV (H.264) | MPEG-TS | Works. `h264_mp4toannexb` BSF auto-inserted. |
| MKV (H.265) | MPEG-TS | Works. `hevc_mp4toannexb` BSF auto-inserted. |
| MKV (H.264) | fMP4 | Works. No BSF needed. |
| MP4 (H.264) | MPEG-TS | Works. BSF auto-inserted. |
| MPEG-TS (H.264) | fMP4 | Works. May need `-bsf:v h264_mp4toannexb` explicitly in some versions. |

### `-bsf:v h264_mp4toannexb`

Converts H.264 bitstream from length-prefixed mode (MP4/MKV style) to start-code-prefixed mode (Annex B, required by MPEG-TS).

- **Auto-inserted** by FFmpeg when outputting to MPEG-TS format (`-f hls` with `mpegts` segments or `-f mpegts`).
- Explicitly specifying it is harmless but unnecessary for modern FFmpeg.
- There is also `hevc_mp4toannexb` for H.265/HEVC streams.

### `-tag:v codec_tag`

Override the codec tag written to the output container. Rarely needed for HLS.

- Use case: forcing `hvc1` tag for HEVC in fMP4 (`-tag:v hvc1`) for Apple compatibility.

### Keyframe Alignment with Copy Mode

**This is the biggest limitation of `-c:v copy` for HLS:**

- HLS segments can only be split at keyframes (I-frames).
- `-hls_time 6` means "split at the next keyframe after 6 seconds," not "split at exactly 6 seconds."
- If the source has keyframes every 10 seconds, segments will be ~10 seconds each regardless of `hls_time`.
- If keyframe interval varies, segment durations will be irregular.
- B-frames in H.264 High profile can cause the start PTS of segments to differ from the expected position.

**For consistent segment duration, re-encode with forced keyframes:**
```
-force_key_frames "expr:gte(t,n_forced*6)" -sc_threshold:v 0
```

---

## 4. Audio Transcode Options

### Codec Selection

| Codec | FFmpeg Encoder | Quality | Compatibility | Notes |
|-------|---------------|---------|---------------|-------|
| AAC | `aac` (native) | Good | Universal | Default FFmpeg AAC encoder. Good enough for streaming. |
| AAC | `libfdk_aac` | Better | Universal | Higher quality but requires non-free license. Jellyfin uses this when available. |
| Opus | `libopus` | Best | Limited HLS | Excellent quality/bitrate ratio but not supported in all HLS clients. |
| MP3 | `libmp3lame` | Adequate | Universal | Legacy. No reason to use over AAC. |

### Common Audio Options

```
-c:a aac -b:a 192k -ac 2 -ar 48000
```

| Option | Description | Typical Values |
|--------|-------------|---------------|
| `-c:a aac` | Audio codec | `aac`, `libfdk_aac`, `copy` |
| `-b:a 192k` | Audio bitrate | `128k` (stereo), `192k` (high quality stereo), `384k` (5.1 surround) |
| `-ac 2` | Audio channel count | `2` (stereo downmix), `6` (5.1 passthrough) |
| `-ar 48000` | Audio sample rate in Hz | `44100`, `48000` (standard) |

### `-af` Audio Filters

| Filter | Usage | Description |
|--------|-------|-------------|
| `volume=2` | `-af "volume=2"` | Amplify audio by 2x (Jellyfin uses this for some transcodes) |
| `aresample=async=1` | `-af "aresample=async=1"` | Fix A/V sync by resampling audio timestamps |
| `loudnorm` | `-af loudnorm` | Normalize loudness (EBU R128) |

### Audio/Video Sync Issues (Copy Video + Transcode Audio)

When video is copied (`-c:v copy`) and audio is transcoded (`-c:a aac`), sync issues can arise because:

1. **Video packets have original timing** (keyframe-aligned, possibly with B-frame reordering).
2. **Audio packets are re-timed** by the encoder based on sample count.
3. The muxer must interleave these properly.

Mitigations:
- `-max_muxing_queue_size 4096` -- prevents muxer overflow when A/V packet timing diverges.
- `-avoid_negative_ts make_zero` -- ensures both streams start at 0.
- `-max_delay 5000000` -- allows up to 5 seconds of interleave delay.

---

## 5. HLS Muxer Options (-f hls)

### `-hls_time seconds`

Target segment duration in seconds. Default: 2.

- FFmpeg cuts at the **next keyframe after** this duration has elapsed.
- With `-c:v copy`, actual duration depends on source keyframe interval.
- With encoding + `-force_key_frames`, duration is consistent.
- Apple recommends 6 seconds. Velox and Jellyfin both use 6.

### `-hls_list_size count`

Maximum number of entries in the playlist. Default: 5.

- `0` = include ALL segments (no sliding window). Required for VOD/Event.
- `> 0` = sliding window playlist (live streaming). Old segments are removed from the playlist.

### `-hls_playlist_type type`

| Type | Behavior |
|------|----------|
| `event` | Segments are appended but never removed. Playlist grows. `#EXT-X-PLAYLIST-TYPE:EVENT` is written. Clients know the stream is ongoing but past segments are stable. `#EXT-X-ENDLIST` is added when encoding finishes. |
| `vod` | Same as `event` but `#EXT-X-PLAYLIST-TYPE:VOD` is written. Forces `hls_list_size` to 0. Implies the entire playlist is available upfront. |
| (omitted) | Standard live mode. Old segments can be removed based on `hls_list_size`. |

**Velox uses `event`** -- segments are added as FFmpeg produces them; `#EXT-X-ENDLIST` marks completion. This allows the client to start playback before encoding finishes.

**Jellyfin uses `vod`** -- the playlist declares all segments upfront. With `-start_number` set to the seek segment, this works for their seek model.

### `-hls_segment_type type`

| Type | Extension | Pros | Cons |
|------|-----------|------|------|
| `mpegts` (default) | `.ts` | Universal compatibility. Each segment is self-contained (includes codec params). No init segment needed. | Larger overhead per segment. Does not support AV1 codec well. |
| `fmp4` | `.m4s` + `init.mp4` | Smaller segments. Supports modern codecs (AV1, HEVC). Compatible with both HLS and DASH from same files. Better caching (shared with DASH CDN). | Requires `#EXT-X-MAP` in playlist. Some older players/browsers have issues. Chrome/Edge may need hls.js. Needs HLS version 7+. |

**Key difference:** MPEG-TS includes parsing info in every segment. fMP4 separates initialization data into a single init segment referenced by `#EXT-X-MAP`.

**Browser compatibility:** Safari natively supports both. Chrome/Edge/Firefox require MSE-based players (hls.js, video.js) for either, but fMP4 has had more edge-case bugs in some player versions.

### `-hls_segment_filename pattern`

Output pattern for segment files. Supports `printf`-style formatting.

```
-hls_segment_filename "/path/to/segments/seg_%04d.ts"
```

- `%d` = segment number
- `%v` = variant index (multi-variant streams)
- When `strftime` is enabled, supports time format codes.

### `-hls_fmp4_init_filename filename`

Sets the filename for the fMP4 initialization segment. Default: `init.mp4`.

```
-hls_fmp4_init_filename "stream-init.mp4"
```

### `-hls_segment_options options`

Passes format options to the segment muxer as colon-separated key=value pairs.

```
-hls_segment_options "movflags=+frag_keyframe+empty_moov"
```

Internally, FFmpeg sets `movflags` to `+frag_custom+dash+delay_moov` for fMP4 segments.

### `-hls_flags flags`

Multiple flags can be combined with `+`:

| Flag | Description |
|------|-------------|
| `single_file` | Store all segments in one file, using byte ranges in the playlist. Requires HLS version 4. |
| `delete_segments` | Delete segments removed from the playlist after `hls_delete_threshold` duration. For live streaming. |
| `append_list` | Append new segments to existing playlist (don't overwrite). Remove `#EXT-X-ENDLIST` from old list. |
| `round_durations` | Round segment durations to integers in the playlist instead of float. |
| `discont_start` | Add `#EXT-X-DISCONTINUITY` tag at the start of the playlist. |
| `omit_endlist` | Don't write `#EXT-X-ENDLIST` when encoding finishes. For ongoing live streams. |
| `split_by_time` | Allow segments to start on non-keyframes. Can fix timing on sources with irregular keyframes but may cause artifacts. |
| `independent_segments` | Add `#EXT-X-INDEPENDENT-SEGMENTS` when all segments start with keyframes. Informational tag for clients. |
| `iframes_only` | Generate an I-frames-only playlist. For trick play (fast forward/rewind). |
| `program_date_time` | Add `#EXT-X-PROGRAM-DATE-TIME` tags with wall-clock timestamps. |
| `temp_file` | Write segments to a temp file first, then atomically rename. Prevents serving incomplete segments (HTTP 416 errors). Important for live streaming. |
| `periodic_rekey` | Re-read the key info file before each segment for key rotation. |
| `second_level_segment_index` | Use `%%d` in filename for segment index when `strftime` is on. |
| `second_level_segment_size` | Use `%%s` in filename for segment size when `strftime` is on. |
| `second_level_segment_duration` | Use `%%t` in filename for segment duration when `strftime` is on. |

### `-start_number number`

Set the sequence number of the first segment. Default: 0.

- Jellyfin uses this for seek: when seeking to segment 24 (position / hls_time), sets `-start_number 24` so the playlist starts there.

### `-master_pl_name filename`

Create a master playlist (multi-variant) with this filename. Used for ABR (Adaptive Bitrate) with `var_stream_map`.

### `-var_stream_map mapping`

Define variant streams for multi-bitrate HLS. Example:
```
-var_stream_map "v:0,a:0 v:1,a:1"
```

---

## 6. Muxer/General Options

### `-max_delay microseconds`

Maximum muxing or demuxing delay in microseconds.

- Controls how long the muxer will buffer packets for interleaving.
- Velox and Jellyfin both use `5000000` (5 seconds).
- Higher values allow better interleaving at the cost of initial latency.

### `-max_muxing_queue_size packets`

Maximum number of packets buffered per stream in the muxer queue.

- FFmpeg waits for at least one packet from each stream before writing. When streams have very different packet rates (e.g., video at 24fps vs subtitle at 1 per minute), the queue can overflow.
- Default varies by version. Common error: "Too many packets buffered for output stream."
- Velox uses `4096`. Jellyfin uses `2048`.
- Set to `0` to disable the limit (unlimited buffering -- can consume lots of memory).

### `-map_metadata input_index`

Copy metadata from input. `-1` strips all metadata.

- `-map_metadata -1` -- used by Velox and Jellyfin to produce clean HLS output without source metadata.

### `-map_chapters input_index`

Copy chapter information. `-1` strips all chapters.

- `-map_chapters -1` -- chapters are irrelevant in HLS segments.

### `-movflags flags`

Flags for the MP4/fMP4 muxer (applies to fMP4 HLS segments):

| Flag | Description |
|------|-------------|
| `frag_keyframe` | Start a new fragment at each keyframe. |
| `empty_moov` | Write an initial moov atom without sample descriptions. Essential for streaming -- allows header to be sent before media data. |
| `default_base_moof` | Use `default-base-is-moof` flag in tfhd instead of absolute offsets. Makes fragments position-independent. |
| `frag_custom` | Allow custom fragment boundaries. Used internally by HLS muxer. |
| `dash` | Write sidx (segment index) for DASH compatibility. |
| `delay_moov` | Delay writing moov until the first frame. Used internally by HLS fMP4. |
| `negative_cts_offsets` | Use negative CTS offsets for B-frames. Reduces edit list usage. |

**For MSE (Media Source Extensions) compatibility:**
```
-movflags frag_keyframe+empty_moov+default_base_moof
```

### `-sn` (Disable Subtitles)

Disable subtitle output. Used when subtitles are handled separately (WebVTT sidecar) or burned in via filters.

---

## 7. Common Use Cases with Exact Commands

### a) MKV (H.264 + AC3) -> HLS with Video Copy + AAC Audio (MPEG-TS)

```bash
ffmpeg -hide_banner -loglevel warning \
  -probesize 5000000 -analyzeduration 5000000 \
  -i input.mkv \
  -map_metadata -1 -map_chapters -1 \
  -map 0:v:0 -map 0:a:0 \
  -sn \
  -c:v copy \
  -avoid_negative_ts make_zero \
  -max_muxing_queue_size 4096 \
  -c:a aac -b:a 192k -ac 2 \
  -f hls \
  -max_delay 5000000 \
  -hls_time 6 \
  -hls_list_size 0 \
  -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

**Explanation:**
- `-c:v copy` -- no video re-encoding (fast).
- `-avoid_negative_ts make_zero` -- ensures MPEG-TS output starts at PTS 0 (critical for BluRay REMUX files).
- `-c:a aac -b:a 192k -ac 2` -- transcode AC3 to stereo AAC for universal compatibility.
- `-hls_playlist_type event` -- playlist grows as segments are produced; `#EXT-X-ENDLIST` added at finish.
- `h264_mp4toannexb` BSF is auto-inserted for MKV->TS.

### b) MKV (H.264 + AC3) -> HLS with Video Copy + AAC Audio (fMP4)

```bash
ffmpeg -hide_banner -loglevel warning \
  -analyzeduration 200000000 -probesize 1000000000 \
  -fflags +genpts \
  -i input.mkv \
  -map_metadata -1 -map_chapters -1 \
  -map 0:v:0 -map 0:a:0 \
  -sn \
  -c:v copy \
  -bsf:v h264_mp4toannexb \
  -start_at_zero \
  -c:a aac -b:a 192k -ac 2 \
  -copyts -avoid_negative_ts disabled \
  -max_muxing_queue_size 2048 \
  -f hls \
  -max_delay 5000000 \
  -hls_time 6 \
  -hls_segment_type fmp4 \
  -hls_fmp4_init_filename "init.mp4" \
  -start_number 0 \
  -hls_segment_filename "seg_%d.mp4" \
  -hls_playlist_type vod \
  -hls_list_size 0 \
  master.m3u8
```

**Explanation (Jellyfin-style):**
- `-fflags +genpts` -- generate missing PTS (safety for some MKV files).
- `-copyts -start_at_zero -avoid_negative_ts disabled` -- Jellyfin's timestamp strategy: preserve original timing but shift start to 0. Timestamp normalization is done by `-start_at_zero`, not `avoid_negative_ts`.
- `-hls_segment_type fmp4` -- fragmented MP4 segments (fMP4 handles timestamp offsets in MOOF atoms).
- `-hls_playlist_type vod` -- complete playlist declared upfront.

### c) MKV (H.265/HEVC) -> HLS with Full Transcode to H.264

```bash
ffmpeg -hide_banner -loglevel warning \
  -probesize 50000000 -analyzeduration 100000000 \
  -hwaccel videotoolbox \
  -i input.mkv \
  -map 0:v:0 -map 0:a:0 \
  -sn \
  -vf "scale=-2:1080" \
  -c:v libx264 -preset veryfast -crf 23 \
  -threads 0 -pix_fmt yuv420p \
  -c:a aac -b:a 128k -ac 2 \
  -f hls \
  -hls_time 6 \
  -hls_list_size 0 \
  -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

**With NVIDIA hardware encoding:**
```bash
ffmpeg -hide_banner -loglevel warning \
  -hwaccel cuda -hwaccel_output_format cuda \
  -i input.mkv \
  -map 0:v:0 -map 0:a:0 \
  -sn \
  -c:v h264_nvenc -preset p4 -b:v 4M -maxrate 6M -bufsize 12M \
  -c:a aac -b:a 128k -ac 2 \
  -f hls \
  -hls_time 6 \
  -hls_list_size 0 \
  -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

### d) MKV with Multiple Audio Tracks -> HLS with Multi-Audio

Uses FFmpeg's multi-output syntax (single input read, multiple output files):

```bash
ffmpeg -hide_banner -loglevel warning \
  -probesize 50000000 -analyzeduration 100000000 \
  -i input.mkv \
  -map_metadata -1 -map_chapters -1 \
  -avoid_negative_ts make_zero \
  -max_muxing_queue_size 4096 \
  # Video output (copy)
  -map 0:v:0 -c:v copy \
  -an \
  -f hls -max_delay 5000000 \
  -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  -hls_segment_filename "video_%04d.ts" \
  video.m3u8 \
  # Audio track 1 (English, stream index 1)
  -map 0:1 \
  -c:a aac -b:a 128k -ac 2 \
  -f hls -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  -hls_segment_filename "audio_1_%04d.ts" \
  audio_1.m3u8 \
  # Audio track 2 (Japanese, stream index 2)
  -map 0:2 \
  -c:a aac -b:a 128k -ac 2 \
  -f hls -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  -hls_segment_filename "audio_2_%04d.ts" \
  audio_2.m3u8
```

**Master playlist (written manually):**
```m3u8
#EXTM3U
#EXT-X-VERSION:4
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",LANGUAGE="eng",NAME="English",DEFAULT=YES,AUTOSELECT=YES,URI="audio_1.m3u8"
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",LANGUAGE="jpn",NAME="Japanese",DEFAULT=NO,AUTOSELECT=NO,URI="audio_2.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=4000000,AUDIO="audio"
video.m3u8
```

**Key insight:** Single FFmpeg process reads input once and produces N+1 outputs. Video segments start immediately without waiting for audio.

### e) Seeking (Resume from Position) with Video Copy

```bash
ffmpeg -hide_banner -loglevel warning \
  -probesize 5000000 -analyzeduration 5000000 \
  -ss 3480.000 \
  -i input.mkv \
  -map_metadata -1 -map_chapters -1 \
  -map 0:v:0 -map 0:a:0 \
  -sn \
  -c:v copy \
  -avoid_negative_ts make_zero \
  -max_muxing_queue_size 4096 \
  -c:a aac -b:a 192k -ac 2 \
  -f hls \
  -max_delay 5000000 \
  -hls_time 6 \
  -hls_list_size 0 \
  -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

**Key points:**
- `-ss` before `-i` seeks in the demuxer (fast keyframe seek).
- With `-c:v copy`, output starts at the nearest preceding keyframe.
- `-avoid_negative_ts make_zero` resets output timestamps to 0.
- Output segment 0 corresponds to the seek position, not the file start.

### f) Seeking with Full Transcode

```bash
ffmpeg -hide_banner -loglevel warning \
  -probesize 50000000 -analyzeduration 100000000 \
  -hwaccel videotoolbox \
  -ss 3480.000 \
  -i input.mkv \
  -map 0:v:0 -map 0:a:0 \
  -sn \
  -c:v libx264 -preset veryfast -crf 23 -threads 0 -pix_fmt yuv420p \
  -c:a aac -b:a 128k -ac 2 \
  -f hls \
  -hls_time 6 \
  -hls_list_size 0 \
  -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

**Key points:**
- `-ss` before `-i` with transcoding is frame-accurate (since FFmpeg 2.1).
- The gap between the keyframe seek point and the exact position is decoded and discarded.
- No need for `-avoid_negative_ts make_zero` when transcoding -- encoder generates clean timestamps.

### g) HDR -> SDR Tone Mapping

**Software (zscale + hable):**
```bash
ffmpeg -hide_banner -loglevel warning \
  -i hdr_input.mkv \
  -map 0:v:0 -map 0:a:0 \
  -vf "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv,format=yuv420p" \
  -c:v libx264 -preset veryfast -crf 23 -pix_fmt yuv420p \
  -c:a aac -b:a 128k -ac 2 \
  -f hls \
  -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

**VAAPI hardware tonemap (Linux Intel/AMD):**
```bash
ffmpeg -hide_banner -loglevel warning \
  -hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128 \
  -i hdr_input.mkv \
  -map 0:v:0 -map 0:a:0 \
  -vf "tonemap_vaapi=format=nv12" \
  -c:v h264_vaapi -profile:v main -qp 23 \
  -c:a aac -b:a 128k -ac 2 \
  -f hls \
  -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  -hls_segment_filename "seg_%04d.ts" \
  master.m3u8
```

**Filter chain breakdown (software):**

| Step | Filter | Purpose |
|------|--------|---------|
| 1 | `zscale=t=linear:npl=100` | Convert transfer function to linear light (npl=100 nits nominal peak luminance) |
| 2 | `format=gbrpf32le` | Convert to 32-bit float planar RGB for high precision tonemapping |
| 3 | `zscale=p=bt709` | Convert color primaries from BT.2020 to BT.709 |
| 4 | `tonemap=tonemap=hable:desat=0` | Apply Hable (Uncharted 2) tone mapping curve. `desat=0` preserves color saturation. |
| 5 | `zscale=t=bt709:m=bt709:r=tv` | Set transfer/matrix to BT.709, TV range |
| 6 | `format=yuv420p` | Convert to 8-bit 4:2:0 for H.264 compatibility |

**Tone mapping algorithms:**

| Algorithm | Character | Notes |
|-----------|-----------|-------|
| `hable` | Balanced, natural look | Most popular for media servers. Based on Uncharted 2 filmic curve. |
| `reinhard` | Brighter highlights | Jellyfin's default. Can look washed out without desat tuning. |
| `mobius` | Similar to reinhard with smoother rolloff | Good alternative to reinhard. |
| `bt2390` | ITU standard | Closest to reference HDR->SDR conversion. |

---

## 8. Known Pitfalls & Conflicts

### Pitfall 1: `-fflags +genpts` vs `-avoid_negative_ts make_zero`

**DO NOT combine these.** `+genpts` regenerates PTS values from DTS, creating non-zero timestamps. `make_zero` then cannot detect "negative" timestamps to shift because `genpts` has already made them positive but non-zero. The result: output PTS does not start at 0.

**Fix:** Use one or the other:
- MPEG-TS output: use `-avoid_negative_ts make_zero` without `+genpts`.
- fMP4 output: use `-fflags +genpts` with `-copyts -start_at_zero -avoid_negative_ts disabled` (Jellyfin approach).

### Pitfall 2: `-copyts -start_at_zero` with MPEG-TS vs fMP4

- **fMP4:** Works correctly. Timestamp offsets are encoded in MOOF/TRAF atoms; players read `baseMediaDecodeTime` to compute presentation.
- **MPEG-TS:** `-copyts` alone preserves original PTS in PES headers. Even with `-start_at_zero`, some edge cases with BluRay REMUX files (start PTS > 0) can cause player seek miscalculation if the source has a non-standard start time that `-start_at_zero` doesn't fully compensate.
- **Recommendation:** For MPEG-TS, prefer `-avoid_negative_ts make_zero` over `-copyts -start_at_zero`.

### Pitfall 3: B-frame Issues with `-c:v copy`

H.264 High profile uses B-frames for better compression. With `-c:v copy`:

- B-frames have PTS that is out of DTS order (display order differs from decode order).
- When segmenting at keyframes, the first few frames of a segment may have PTS values that appear to go backward relative to the previous segment's last frames.
- Some players handle this correctly; others show brief glitches at segment boundaries.
- No workaround with copy mode -- this is inherent to the bitstream.

### Pitfall 4: BluRay REMUX Files with Non-Zero Start PTS

BluRay REMUX MKV files often have a non-zero start PTS (e.g., PTS starts at 107 seconds) inherited from the disc's BDAV structure.

- Without timestamp correction, HLS segments will have PTS starting at 107s, causing seek to overshoot by 107 seconds.
- **Solution for MPEG-TS:** `-avoid_negative_ts make_zero`
- **Solution for fMP4:** `-copyts -start_at_zero -avoid_negative_ts disabled`
- **Do NOT use:** `-fflags +genpts` with `make_zero` (see Pitfall 1).

### Pitfall 5: SourceBuffer Quota Issues with High Bitrate Content

Browser MSE (Media Source Extensions) has limited SourceBuffer capacity:

- **Chrome:** ~150 MB per SourceBuffer. At 30 Mbps (4K REMUX with copy), each 6-second segment is ~22 MB. Chrome runs out of buffer space after ~40 seconds.
- **Firefox/Safari:** Different limits but same fundamental issue.

**Mitigations:**
- Reduce segment duration (e.g., `-hls_time 4` instead of 6) to produce smaller segments.
- Client-side: actively evict old segments from the SourceBuffer (`SourceBuffer.remove()`).
- hls.js: configure `maxBufferLength` and `maxMaxBufferLength` appropriately.
- Transcode to lower bitrate when source is very high bitrate.

### Pitfall 6: Segment Duration Irregularity with Copy Mode

With `-c:v copy`, segment boundaries align to source keyframes, NOT to `hls_time`:

- Source with keyframes every 2s + `hls_time 6`: segments are ~6s (3 keyframes).
- Source with keyframes every 10s + `hls_time 6`: segments are ~10s (1 keyframe per segment).
- Source with variable keyframe interval: unpredictable segment durations.

**Impact:** ABR (Adaptive Bitrate) switching requires aligned segment boundaries. With copy mode and different sources, segments won't align, causing buffering during quality switches.

**Fix for ABR:** Always re-encode with forced keyframes:
```
-force_key_frames "expr:gte(t,n_forced*6)" -sc_threshold:v 0
```

### Pitfall 7: `hls_time` is Approximate, Not Exact

Even when transcoding, `hls_time` targets an average, not an exact guarantee. The actual split happens at the next keyframe after the target time. With `-force_key_frames`, the keyframe and segment boundary align, but the segment still includes all frames up to (and including) the keyframe.

### Pitfall 8: Multi-Output + Copy Mode Interleaving

When using FFmpeg's multi-output syntax (video + N audio outputs from single input), with `-c:v copy`:
- Video and audio packets are read from a single demuxer but dispatched to different muxers.
- If audio transcoding is slow (e.g., high channel count), video muxer may stall waiting for interleaving.
- `-max_muxing_queue_size 4096` prevents this from becoming an error.

### Pitfall 9: `hls_playlist_type vod` with Live Transcoding

Using `-hls_playlist_type vod` while FFmpeg is still encoding creates a playlist that claims to be complete VOD. Clients may try to seek to segments that don't exist yet.

**Jellyfin's approach:** Uses `vod` type but handles this at the application level by intercepting segment requests and waiting for FFmpeg to produce them.

**Velox's approach:** Uses `event` type, which correctly communicates "stream is live, don't seek past the end." `#EXT-X-ENDLIST` is only added when encoding finishes.

### Pitfall 10: HEVC (H.265) in MPEG-TS

While technically supported, HEVC in MPEG-TS has compatibility issues:
- Not supported by Apple's native HLS player (Safari requires HEVC in fMP4 with `hvc1` tag).
- Some set-top boxes and smart TVs may not support HEVC in TS.
- If you need to stream HEVC via HLS, use fMP4 segments.

---

## 9. What Jellyfin vs Emby vs Plex Use

### Jellyfin (Open Source -- Full Command Visibility)

Based on actual FFmpeg command logs from Jellyfin 10.10+:

**Input Arguments:**
```
-analyzeduration 200M -probesize 1G
-fflags +genpts
-noaccurate_seek          # (when seeking with copy)
-ss HH:MM:SS.SSS          # (before -i)
```

**Timestamp Handling:**
```
-copyts
-start_at_zero
-avoid_negative_ts disabled
```

**HLS Output (fMP4 mode -- default since Jellyfin 10.9+):**
```
-f hls
-max_delay 5000000
-hls_time 6
-hls_segment_type fmp4
-hls_fmp4_init_filename "{hash}-1.mp4"
-hls_segment_filename "/cache/transcodes/{hash}%d.mp4"
-hls_playlist_type vod
-hls_list_size 0
-start_number {seek_segment}
-max_muxing_queue_size 2048
```

**HLS Output (MPEG-TS mode -- for subtitle burn-in / compatibility):**
```
-f hls
-max_delay 5000000
-hls_time 3               # shorter segments for transcode
-hls_segment_type mpegts
-hls_segment_filename "/cache/transcodes/{hash}%d.ts"
-hls_playlist_type vod
-hls_list_size 0
-start_number {seek_segment}
-max_muxing_queue_size 2048
```

**Video (Copy):**
```
-codec:v:0 copy
-bsf:v h264_mp4toannexb   # explicit, even though auto-inserted
```

**Video (Transcode):**
```
-codec:v:0 libx264
-preset veryfast
-crf 23
-maxrate 9570100
-bufsize 19140200
-profile:v:0 high
-level 51
-x264opts:0 subme=0:me_range=4:rc_lookahead=10:me=dia:no_chroma_me:8x8dct=0:partitions=none
-force_key_frames:0 "expr:gte(t,{offset}+n_forced*3)"
-sc_threshold:v:0 0
```

**Audio:**
```
-codec:a:0 copy            # (when compatible)
-codec:a:0 libfdk_aac      # (when transcoding, if available)
-ac 2 -ab 384000 -ar 48000
```

**Key Jellyfin Design Decisions:**
- fMP4 by default (better for HEVC, smaller overhead).
- `-copyts -start_at_zero -avoid_negative_ts disabled` -- timestamp handling delegated to `-start_at_zero`.
- `-fflags +genpts` -- safety net for missing PTS (works because they don't use `make_zero`).
- `-noaccurate_seek` with copy mode -- fast seeking without decoding the gap.
- `-hls_playlist_type vod` -- declares complete playlist upfront.
- `-start_number` set to the seek segment index.
- Explicit `-bsf:v h264_mp4toannexb` even for fMP4 output.
- Very large probesize/analyzeduration (1G / 200M) for maximum stream detection reliability.
- `-x264opts` with aggressive speed optimizations (subme=0, partitions=none).
- `-force_key_frames` at 3-second intervals for consistent seeking.

### Emby (Closed Source -- Limited Visibility)

Based on community reports and forum logs:

- Uses its own modified FFmpeg build.
- General HLS parameters similar to Jellyfin (shared heritage -- Emby forked from original MediaBrowser, Jellyfin forked from Emby).
- `-hls_time` typically 3 or 6 seconds.
- Supports both MPEG-TS and fMP4 segment types.
- Hardware acceleration via VAAPI, NVENC, QSV similar to Jellyfin.
- Detailed command lines are not publicly documented due to closed source.

### Plex (Closed Source -- Modified FFmpeg)

Based on log analysis and community reports:

- Uses a heavily customized FFmpeg fork ("Plex Transcoder").
- **DASH for modern clients** (not pure HLS) -- Plex switched to DASH/fMP4 for web and most apps.
- Falls back to HLS MPEG-TS for legacy clients (Roku, older smart TVs).
- Segment duration configurable in settings (default ~4-6 seconds).
- Throttled transcoding by default (transcodes slightly ahead of playback position to save CPU).
- Exact FFmpeg arguments visible in Plex Media Server logs (`/var/lib/plexmediaserver/Logs/`).
- Key difference from Jellyfin: Plex's transcoder binary is not standard FFmpeg -- it includes custom muxers and demuxers.

### Velox (This Project)

**Input:**
```
-probesize 5000000 -analyzeduration 5000000       # simple copy
-probesize 50000000 -analyzeduration 100000000    # subtitle burn-in / transcode
-hwaccel {videotoolbox|vaapi|cuda|qsv}            # when available
-ss {offset}                                       # before -i
```

**Timestamp Handling (MPEG-TS -- different from Jellyfin):**
```
-avoid_negative_ts make_zero
# NO -copyts, NO -start_at_zero, NO -fflags +genpts
```

**HLS Output:**
```
-f hls
-max_delay 5000000
-hls_time 6
-hls_list_size 0
-hls_playlist_type event           # (not vod -- correct for live transcoding)
-hls_segment_filename "seg_%04d.ts"
# MPEG-TS segments (no fMP4)
```

**Video (Copy):**
```
-c:v copy
# no explicit -bsf:v (auto-inserted)
```

**Video (Transcode):**
```
-c:v {libx264|h264_nvenc|h264_vaapi|h264_qsv|h264_videotoolbox|h264_amf}
-preset veryfast -crf 23 -threads 0 -pix_fmt yuv420p   # software
-profile:v main -qp 23                                   # VAAPI
```

**Audio:**
```
-c:a aac -b:a 128k -ac 2     # transcode mode
-c:a aac -b:a 192k -ac 2     # copy mode (higher quality)
```

**Key Velox Design Decisions:**
- MPEG-TS segments (universal compatibility, simpler implementation).
- `-avoid_negative_ts make_zero` instead of `-copyts -start_at_zero` (simpler, more reliable for MPEG-TS).
- `-hls_playlist_type event` (correct semantics for live transcoding).
- No `-fflags +genpts` (conflicts with `make_zero`).
- Multi-output FFmpeg for multi-audio (single input read, video + N audio outputs).
- Semaphore-based concurrency limiting (video copy skips semaphore).
- Hardware fallback: try HW encoder first, fall back to software on failure.
- Master playlist written BEFORE encoding starts for multi-audio.

---

## Appendix: Quick Reference Cheat Sheet

### Minimal Video Copy HLS (MPEG-TS)
```bash
ffmpeg -ss {offset} -i input.mkv \
  -c:v copy -c:a aac -b:a 192k -ac 2 \
  -avoid_negative_ts make_zero \
  -f hls -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  output.m3u8
```

### Minimal Full Transcode HLS (MPEG-TS)
```bash
ffmpeg -ss {offset} -i input.mkv \
  -c:v libx264 -preset veryfast -crf 23 \
  -c:a aac -b:a 128k -ac 2 \
  -f hls -hls_time 6 -hls_list_size 0 -hls_playlist_type event \
  output.m3u8
```

### Minimal fMP4 HLS (Jellyfin-style)
```bash
ffmpeg -fflags +genpts -i input.mkv \
  -c:v copy -c:a aac -b:a 192k -ac 2 \
  -copyts -start_at_zero -avoid_negative_ts disabled \
  -f hls -hls_time 6 -hls_segment_type fmp4 \
  -hls_list_size 0 -hls_playlist_type vod \
  output.m3u8
```

---

## Sources

- [FFmpeg Formats Documentation](https://ffmpeg.org/ffmpeg-formats.html)
- [FFmpeg Main Documentation](https://www.ffmpeg.org/ffmpeg.html)
- [FFmpeg Bitstream Filters Documentation](https://ffmpeg.org/ffmpeg-bitstream-filters.html)
- [FFmpeg Wiki: Seeking](https://fftrac-bg.ffmpeg.org/wiki/Seeking)
- [FFmpeg HLS Muxer Source (muxers.texi)](https://github.com/FFmpeg/FFmpeg/blob/master/doc/muxers.texi)
- [FFmpeg HLS Encoder Source (hlsenc.c)](https://github.com/FFmpeg/FFmpeg/blob/master/libavformat/hlsenc.c)
- [HLS Packaging using FFmpeg - OTTVerse](https://ottverse.com/hls-packaging-using-ffmpeg-live-vod/)
- [HLS and Fragmented MP4 - hlsbook.net](https://hlsbook.net/hls-fragmented-mp4/)
- [Using FFmpeg as a HLS Streaming Server (Parts 1-8) - Martin Riedl](https://www.martin-riedl.de/2018/08/24/using-ffmpeg-as-a-hls-streaming-server-part-1/)
- [Jellyfin GitHub Repository](https://github.com/jellyfin/jellyfin)
- [Jellyfin FFmpeg Repository](https://github.com/jellyfin/jellyfin-ffmpeg)
- [Jellyfin Media Streaming - DeepWiki](https://deepwiki.com/jellyfin/jellyfin/3-media-streaming)
- [Jellyfin Issue #13039 - Buffering without Transcoding](https://github.com/jellyfin/jellyfin/issues/13039)
- [Jellyfin Issue #11450 - HLS H264 Buffer Stalling](https://github.com/jellyfin/jellyfin/issues/11450)
- [Jellyfin Issue #3960 - max_muxing_queue_size](https://github.com/jellyfin/jellyfin/issues/3960)
- [Plex Transcoder Fork](https://github.com/comio/plex-ffmpeg)
- [NVIDIA FFmpeg Transcoding Guide](https://developer.nvidia.com/blog/nvidia-ffmpeg-transcoding-guide/)
- [NVIDIA FFmpeg with GPU Hardware Acceleration](https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/ffmpeg-with-nvidia-gpu/index.html)
- [VAAPI Hardware Accelerated Encoding Guide](https://gist.github.com/Brainiarc7/95c9338a737aa36d9bb2931bed379219)
- [Chrome SourceBuffer Quota Exceeded](https://developer.chrome.com/blog/quotaexceedederror)
- [hls.js High Bitrate Configuration](https://github.com/video-dev/hls.js/issues/5031)
- [FFmpeg HDR to SDR Conversion Guide](https://ericswpark.com/blog/2022/2022-12-14-ffmpeg-convert-hdr-to-sdr/)
- [Jellyfin HDR Tonemapping Issue #415](https://github.com/jellyfin/jellyfin/issues/415)
- [MSE Transcoding Assets - MDN](https://developer.mozilla.org/en-US/docs/Web/API/Media_Source_Extensions_API/Transcoding_assets_for_MSE)
- [LosslessCut avoid_negative_ts Discussion](https://github.com/mifi/lossless-cut/discussions/1874)
- [fMP4 Playback Issue on Chrome/Edge - Bitmovin](https://community.bitmovin.com/t/playback-issue-with-ffmpeg-generated-hls-stream-using-fmp4-segments-on-chrome-edge/3053)
- [Fragmented MP4 (fMP4) Support for HLS - GlobalDots](https://www.globaldots.com/resources/blog/fragmented-mp4-fmp4-support-for-hls-v4-lets-shed-some-light-on-it/)
