// Velox-style HLS sub-playlist may be served before ffmpeg writes its first
// segment (the deployed server's WaitForSegment only checks file existence).
// libavformat's HLS demuxer treats that empty initial playlist as fatal:
//
//   [ffmpeg/demuxer] hls: Empty segment [...]
//   [lavf] avformat_find_stream_info() finished after 0 bytes.
//   [cplayer] finished playback, no audio or video data played (reason 4)
//
// This module polls a Velox stream URL until at least one #EXTINF: appears in
// the variant playlist, then returns. The caller can safely invoke mpv
// loadfile after we return Ok.
//
// IMPORTANT: a Velox stream URL may also resolve to a DirectPlay file (raw
// MP4/MKV) instead of HLS. We only attempt to parse playlists when the
// response actually looks like one — otherwise we return Ok immediately and
// let mpv stream the file directly.

use std::io::Read;
use std::time::{Duration, Instant};
use url::Url;

const POLL_INTERVAL: Duration = Duration::from_millis(700);
// Cap on bytes we're willing to read for playlist sniffing. Real m3u8 files
// are well under 100KB; anything larger is definitely a media file we should
// hand off to mpv without further inspection.
const MAX_PLAYLIST_BYTES: usize = 256 * 1024;

pub fn looks_like_hls(url: &str) -> bool {
    url.contains(".m3u8") || url.contains("/api/stream/")
}

pub fn wait_for_first_segment(start_url: &str, total_timeout: Duration) -> Result<(), String> {
    let agent = ureq::AgentBuilder::new()
        .timeout_connect(Duration::from_secs(10))
        .timeout(Duration::from_secs(20))
        .build();

    let deadline = Instant::now() + total_timeout;
    let mut current = start_url.to_string();
    let mut hops = 0u8;

    loop {
        if Instant::now() >= deadline {
            return Err("timeout waiting for HLS first segment".into());
        }

        let resp = match agent.get(&current).call() {
            Ok(r) => r,
            Err(e) => {
                eprintln!("[hls_prewarm] fetch err {} — retrying", e);
                std::thread::sleep(POLL_INTERVAL);
                continue;
            }
        };

        // Use the URL ureq actually fetched (after following redirects) when
        // resolving relative variant URIs from a master playlist.
        let final_url = resp.get_url().to_string();

        // Decide whether to bother reading the body as a playlist. If the
        // server didn't say it's an m3u8 we still sniff the first few bytes,
        // but we hard-cap how much we're willing to slurp.
        let content_type = resp
            .header("content-type")
            .unwrap_or("")
            .to_ascii_lowercase();
        let claims_m3u8 = content_type.contains("mpegurl") || content_type.contains("m3u8");

        let mut reader = resp.into_reader().take(MAX_PLAYLIST_BYTES as u64 + 1);
        let mut buf = Vec::with_capacity(8 * 1024);
        if let Err(e) = reader.read_to_end(&mut buf) {
            eprintln!("[hls_prewarm] read err {} — retrying", e);
            std::thread::sleep(POLL_INTERVAL);
            continue;
        }

        // Body either too big to be a playlist or doesn't start with #EXTM3U:
        // hand off to mpv directly.
        if buf.len() > MAX_PLAYLIST_BYTES {
            eprintln!(
                "[hls_prewarm] response > {}KB — assuming direct media, skipping prewarm",
                MAX_PLAYLIST_BYTES / 1024
            );
            return Ok(());
        }
        let body = match std::str::from_utf8(&buf) {
            Ok(s) => s,
            Err(_) => {
                eprintln!("[hls_prewarm] non-utf8 body — assuming direct media");
                return Ok(());
            }
        };
        if !body.starts_with("#EXTM3U") {
            if claims_m3u8 {
                eprintln!("[hls_prewarm] content-type m3u8 but body lacks #EXTM3U — handing off");
            } else {
                eprintln!("[hls_prewarm] not an m3u8 ({}) — handing off", content_type);
            }
            return Ok(());
        }

        if body.contains("#EXT-X-STREAM-INF:") {
            hops += 1;
            if hops > 3 {
                eprintln!("[hls_prewarm] master chain too deep — handing off");
                return Ok(());
            }
            // Pick the variant URI listed right after the first STREAM-INF line.
            let mut next: Option<&str> = None;
            let mut iter = body.lines();
            while let Some(line) = iter.next() {
                if line.starts_with("#EXT-X-STREAM-INF:") {
                    for follow in iter.by_ref() {
                        let t = follow.trim();
                        if !t.is_empty() && !t.starts_with('#') {
                            next = Some(t);
                            break;
                        }
                    }
                    break;
                }
            }
            let next = match next {
                Some(s) => s,
                None => {
                    eprintln!("[hls_prewarm] master had no variant URI — handing off");
                    return Ok(());
                }
            };
            let base = Url::parse(&final_url).map_err(|e| format!("base url parse: {}", e))?;
            let resolved = base
                .join(next)
                .map_err(|e| format!("variant url resolve: {}", e))?;
            current = resolved.to_string();
            eprintln!("[hls_prewarm] master → variant {}", current);
            continue;
        }

        if body.contains("#EXTINF:") {
            eprintln!("[hls_prewarm] ready: variant has #EXTINF");
            return Ok(());
        }

        // Empty media playlist (only #EXT-X-MAP). Wait and refetch.
        eprintln!("[hls_prewarm] empty variant — waiting for first segment");
        std::thread::sleep(POLL_INTERVAL);
    }
}
