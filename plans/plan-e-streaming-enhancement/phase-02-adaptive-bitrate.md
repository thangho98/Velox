# Phase 02: Adaptive Bitrate HLS
Status: ✅ Done
Plan: E - Streaming Enhancement
Dependencies: Phase 01

## Tasks

### 1. Multi-quality HLS Variants
- [ ] Generate variants: 480p/1.5Mbps, 720p/4Mbps, 1080p/8Mbps
- [ ] Skip qualities higher than source resolution
- [ ] FFmpeg `-var_stream_map` for multi-output

### 2. Master Playlist with Variants
- [ ] Generate `master.m3u8` with `#EXT-X-STREAM-INF` per quality
- [ ] Include BANDWIDTH, RESOLUTION, CODECS
- [ ] hls.js auto-switches based on client bandwidth

### 3. Realtime Segment-by-Segment Transcoding
- [ ] Transcode on-demand (don't wait for full file)
- [ ] Start playback immediately
- [ ] Kill FFmpeg when user stops/disconnects

### 4. Quality Selection UI
- [ ] Quality picker in player controls (Auto/1080p/720p/480p/Original)
- [ ] "Auto" = adaptive (hls.js default)
- [ ] "Original" = direct play (no transcode)
- [ ] Show current quality indicator

### 5. Bandwidth Estimation
- [ ] hls.js reports estimated bandwidth
- [ ] Show to user: "Streaming at 4.2 Mbps"
- [ ] If bandwidth < lowest variant → show buffer warning

---
Next: phase-03-trickplay.md
