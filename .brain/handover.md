━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 HANDOVER DOCUMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Đang làm: Cloud-backed playback hardening + exact audio-track HLS sessions
🔢 Đến bước: Investigation / stabilization after NAS deploy

✅ ĐÃ XONG:
   - Fixed HLS V2 filename parser for `_at{audioTrackId}`
   - Added parser tests for playlist + segment round-trip
   - Deployed parser fix lên NAS
   - Verified old `2x1 scale_vaapi` bug không còn là lỗi chính
   - Added lazy cloud URL resolver for FFmpeg session start/restart
   - Aligned Web + Android audio switching toward exact track ID
   - Added pull-to-refresh state on Android Home / Media Detail / Series Detail

⏳ CÒN LẠI:
   - Fix selected-audio directstream fallback when AC3 remux to MP4 fails
   - Investigate / harden against FShare HTTP 503 during FFmpeg input open
   - Re-test media `1371` end-to-end on Android Lenovo after fallback/retry improvements
   - Decide next implementation step for `docs/specs/shoko_integration_spec.md`

🔧 QUYẾT ĐỊNH QUAN TRỌNG:
   - Audio selection for HLS must use exact track ID, not only language
   - Cloud playback URL must be resolved lazily per FFmpeg start/restart
   - Keep single chosen audio track on the multi-output HLS path to preserve stream index
   - Throttle `cloud-media-probe` to avoid FShare rate limiting

⚠️ LƯU Ý CHO SESSION SAU:
   - Media `1371` now fails mainly because provider URL returns HTTP 503 from FShare
   - Directstream selected-audio path still throws `Cannot write moov atom before AC3 packets`
   - Current staged changes include backend playback fixes plus Android/Web playback UX work
   - HLS parser bug is fixed already; do not chase the old `scale 2x1` issue again

📁 FILES QUAN TRỌNG:
   - .brain/brain.json
   - .brain/session.json
   - .brain/session_log.txt
   - .brain/handover.md
   - backend/internal/hls/naming.go
   - backend/internal/handler/stream.go
   - backend/internal/transcoder/stream_session.go
   - docs/specs/shoko_integration_spec.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📍 Đã lưu! Để tiếp tục: Gõ /recap
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
