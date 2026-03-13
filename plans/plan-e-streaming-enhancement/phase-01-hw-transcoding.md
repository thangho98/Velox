# Phase 01: Hardware Transcoding
Status: ✅ Done
Plan: E - Streaming Enhancement

## Tasks

### 1. Hardware Capability Detection
- [ ] `DetectHWAccel() []string` - probe available accelerators via `ffmpeg -hwaccels`
- [ ] Detect: VideoToolbox (macOS), VAAPI (Linux Intel/AMD), NVENC (NVIDIA), QSV (Intel)
- **File:** `internal/playback/hwdetect.go` - NEW

### 2. HW Accel FFmpeg Args
- [ ] VideoToolbox: `-hwaccel videotoolbox -c:v h264_videotoolbox`
- [ ] VAAPI: `-hwaccel vaapi -hwaccel_device /dev/dri/renderD128 -c:v h264_vaapi`
- [ ] NVENC: `-hwaccel cuda -c:v h264_nvenc`
- [ ] QSV: `-hwaccel qsv -c:v h264_qsv`
- [ ] Integrate into `playback/ffmpeg.go` BuildFFmpegArgs
- **File:** `internal/playback/ffmpeg.go` - Update

### 3. Fallback on Failure
- [ ] If HW transcode fails → auto-retry with software encoder
- [ ] Log warning: "HW accel failed, falling back to software"
- [ ] Don't crash, don't leave broken session

### 4. HDR → SDR Tone Mapping
- [ ] Detect HDR content (color_transfer=smpte2084, color_primaries=bt2020nc)
- [ ] Apply tone mapping filter when transcoding for SDR client
- [ ] FFmpeg: `-vf "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable,zscale=t=bt709:m=bt709,format=yuv420p"`
- **File:** `internal/playback/ffmpeg.go`

### 5. Config
- [ ] `VELOX_HW_ACCEL=auto|videotoolbox|vaapi|nvenc|qsv|none` (default: auto)
- [ ] `VELOX_MAX_TRANSCODES=2` (concurrent transcode limit)
- **File:** `internal/config/config.go`

---
Next: phase-02-adaptive-bitrate.md
