use libmpv2::events::{Event, PropertyData};
use libmpv2::Mpv;
use parking_lot::Mutex;
use serde::Serialize;
use std::sync::Arc;
use std::time::Duration;
use tauri::{AppHandle, Emitter};

pub struct Player {
    pub inner: Mpv,
}

impl Player {
    pub fn new() -> Result<Self, libmpv2::Error> {
        let mpv = Mpv::new()?;
        // mpv noisy diagnostics on stderr while we debug the embed pipeline.
        let _ = mpv.set_property("terminal", "yes");
        let _ = mpv.set_property("msg-level", "all=v");
        // Render-context embedding: surface is provided by host (NSOpenGLView).
        mpv.set_property("vo", "libmpv")?;
        // -copy variant copies decoded frames back to CPU together with their
        // side data. HDR10/HLG mastering metadata travels in HEVC SEI, which
        // VideoToolbox preserves. DV P5 (IPT-PQ-C2) RPU side data also rides
        // here intact because libmpv links jellyfin-ffmpeg with `tonemapx`,
        // which consumes AV_FRAME_DATA_DOVI_METADATA and reshapes IPT→BT.709
        // on CPU. The reshape is wired below in run_event_loop, gated by the
        // observed `video-params/colormatrix` value.
        mpv.set_property("hwdec", "videotoolbox-copy")?;
        mpv.set_property("target-colorspace-hint", "yes")?;
        mpv.set_property("tone-mapping", "bt.2390")?;
        mpv.set_property("keep-open", "yes")?;
        // Velox HLS WaitForSegment may take 90-180s on cold transcode start.
        mpv.set_property("network-timeout", 300i64)?;
        mpv.set_property("cache", "yes")?;
        mpv.set_property("demuxer-max-bytes", 200_000_000i64)?;
        // Fshare CDN (and similar anti-leech CDNs) block default libmpv UA.
        mpv.set_property(
            "user-agent",
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
        )?;
        // Velox renders subtitles via DualSubtitleOverlay (React) so mpv's own
        // text render is hidden by default. Set per-load via player_sub_visible.
        let _ = mpv.set_property("sub-visibility", "yes");
        Ok(Self { inner: mpv })
    }

    pub fn load(&mut self, path: &str, start_time: Option<f64>) -> Result<(), libmpv2::Error> {
        if let Some(t) = start_time {
            // libmpv loadfile syntax: loadfile <url> [<flags>] [<options>]
            // options is a comma-separated key=value string.
            let opts = format!("start={}", t);
            self.inner.command("loadfile", &[path, "replace", "0", &opts])
        } else {
            self.inner.command("loadfile", &[path])
        }
    }

    pub fn pause(&mut self) -> Result<(), libmpv2::Error> {
        self.inner.set_property("pause", true)
    }

    pub fn resume(&mut self) -> Result<(), libmpv2::Error> {
        self.inner.set_property("pause", false)
    }

    pub fn stop(&mut self) -> Result<(), libmpv2::Error> {
        self.inner.command("stop", &[])
    }

    pub fn seek(&mut self, seconds: f64) -> Result<(), libmpv2::Error> {
        self.inner
            .command("seek", &[&seconds.to_string(), "absolute"])
    }

    pub fn set_volume(&mut self, v: f64) -> Result<(), libmpv2::Error> {
        // mpv volume is 0..100; IPlayer uses 0..1.
        let pct = (v * 100.0).clamp(0.0, 100.0);
        self.inner.set_property("volume", pct)
    }

    pub fn set_muted(&mut self, m: bool) -> Result<(), libmpv2::Error> {
        self.inner.set_property("mute", m)
    }

    pub fn set_rate(&mut self, r: f64) -> Result<(), libmpv2::Error> {
        self.inner.set_property("speed", r)
    }

    pub fn sub_add(&mut self, url: &str, lang: &str, label: &str) -> Result<(), libmpv2::Error> {
        // sub-add <url> [<flags> [<title> [<lang>]]]
        // flags=select makes it the active subtitle.
        self.inner.command("sub-add", &[url, "select", label, lang])
    }

    pub fn sub_remove(&mut self) -> Result<(), libmpv2::Error> {
        // sub-remove without ID removes the currently selected subtitle track.
        self.inner.command("sub-remove", &[])
    }

    pub fn set_sub_delay(&mut self, ms: f64) -> Result<(), libmpv2::Error> {
        self.inner.set_property("sub-delay", ms / 1000.0)
    }

    pub fn set_sub_visible(&mut self, visible: bool) -> Result<(), libmpv2::Error> {
        self.inner
            .set_property("sub-visibility", if visible { "yes" } else { "no" })
    }

    /// Switch audio track by language. Walks the `track-list` and finds the
    /// audio track whose `lang` (or `title`) matches; activates by setting `aid`.
    pub fn set_audio_track_by_lang(
        &mut self,
        lang: &str,
        name: Option<&str>,
    ) -> Result<(), libmpv2::Error> {
        let count: i64 = self.inner.get_property("track-list/count").unwrap_or(0);
        for i in 0..count {
            let track_type: String = self
                .inner
                .get_property(&format!("track-list/{}/type", i))
                .unwrap_or_default();
            if track_type != "audio" {
                continue;
            }
            let track_lang: String = self
                .inner
                .get_property(&format!("track-list/{}/lang", i))
                .unwrap_or_default();
            let track_title: String = self
                .inner
                .get_property(&format!("track-list/{}/title", i))
                .unwrap_or_default();
            let matches = track_lang.eq_ignore_ascii_case(lang)
                || name
                    .map(|n| track_title.eq_ignore_ascii_case(n))
                    .unwrap_or(false)
                || track_title.eq_ignore_ascii_case(lang);
            if matches {
                let id: i64 = self
                    .inner
                    .get_property(&format!("track-list/{}/id", i))
                    .unwrap_or(-1);
                if id >= 0 {
                    return self.inner.set_property("aid", id);
                }
            }
        }
        Ok(())
    }

    pub fn position(&self) -> Result<f64, libmpv2::Error> {
        self.inner.get_property("time-pos")
    }

    pub fn duration(&self) -> Result<f64, libmpv2::Error> {
        self.inner.get_property("duration")
    }

    pub fn paused(&self) -> Result<bool, libmpv2::Error> {
        self.inner.get_property("pause")
    }

    pub fn get_str(&self, key: &str) -> Result<String, libmpv2::Error> {
        self.inner.get_property(key)
    }

    pub fn raw_handle(&self) -> *mut libmpv2_sys::mpv_handle {
        self.inner.ctx.as_ptr()
    }

    /// Subscribe to property changes that the front-end cares about. Called
    /// once at startup; the event loop translates these into `player-event`.
    pub fn observe_properties(&mut self) -> Result<(), libmpv2::Error> {
        // Reply IDs (arbitrary, used to identify which property changed).
        const ID_TIME_POS: u64 = 1;
        const ID_PAUSE: u64 = 2;
        const ID_DURATION: u64 = 3;
        const ID_BUFFERING: u64 = 4;
        const ID_COLORMATRIX: u64 = 5;
        self.inner
            .observe_property("time-pos", libmpv2::Format::Double, ID_TIME_POS)?;
        self.inner
            .observe_property("pause", libmpv2::Format::Flag, ID_PAUSE)?;
        self.inner
            .observe_property("duration", libmpv2::Format::Double, ID_DURATION)?;
        self.inner.observe_property(
            "paused-for-cache",
            libmpv2::Format::Flag,
            ID_BUFFERING,
        )?;
        // Used to gate the DV P5 reshape filter chain — see set_dv_reshape.
        self.inner.observe_property(
            "video-params/colormatrix",
            libmpv2::Format::String,
            ID_COLORMATRIX,
        )?;
        Ok(())
    }

    /// Toggle the libavfilter `tonemapx` IPT-PQ-C2 reshape chain on or off.
    ///
    /// macOS's libmpv embed surface is OpenGL 4.1 (no compute shaders), so
    /// libplacebo cannot run the DV reshape itself. Instead we route the
    /// frame through jellyfin-ffmpeg's CPU SIMD `tonemapx` filter, which
    /// reads `AV_FRAME_DATA_DOVI_METADATA` from each frame, reshapes
    /// IPT-PQ-C2 → BT.2020/PQ, then tonemaps to BT.709 SDR (matching the
    /// server-side fallback that used to handle DV P5 for this client).
    /// Toggle is observed in run_event_loop via `video-params/colormatrix`.
    pub fn set_dv_reshape(&mut self, enable: bool) -> Result<(), libmpv2::Error> {
        // tonemapx asserts on input pixfmt yuv420p10 for the DV reshape path.
        // VideoToolbox-copy hands us p010le (NV12-style 10-bit), so prepend a
        // swscale conversion before tonemapx; the trailing format=yuv420p caps
        // the output at 8-bit SDR for the OpenGL renderer.
        const DV_RESHAPE_VF: &str = "lavfi=[setparams=color_primaries=bt2020:color_trc=smpte2084:colorspace=bt2020nc,format=yuv420p10,tonemapx=tonemap=bt2390:desat=0:peak=100:t=bt709:m=bt709:p=bt709,format=yuv420p]";
        if enable {
            self.inner.set_property("vf", DV_RESHAPE_VF)
        } else {
            self.inner.set_property("vf", "")
        }
    }
}

#[derive(Serialize, Clone)]
struct PlayerEvent {
    kind: &'static str,
    detail: Option<String>,
    f: Option<f64>,
    b: Option<bool>,
}

impl PlayerEvent {
    fn simple(kind: &'static str) -> Self {
        Self {
            kind,
            detail: None,
            f: None,
            b: None,
        }
    }
    fn float(kind: &'static str, v: f64) -> Self {
        Self {
            kind,
            detail: None,
            f: Some(v),
            b: None,
        }
    }
    fn flag(kind: &'static str, v: bool) -> Self {
        Self {
            kind,
            detail: None,
            f: None,
            b: Some(v),
        }
    }
}

pub fn run_event_loop(player: Arc<Mutex<Player>>, handle: AppHandle) {
    // Subscribe to property changes once.
    if let Err(e) = player.lock().observe_properties() {
        eprintln!("[velox-desktop] observe_properties failed: {e}");
    }

    loop {
        let (events, shutdown) = {
            let mut p = player.lock();
            let mut batch: Vec<PlayerEvent> = Vec::new();
            let mut shutdown = false;
            // Drain events without blocking too long; 0.1s wait is the upper bound.
            for _ in 0..20 {
                // colormatrix flips between yuv709/bt2020nc/dolbyvision as files
                // load. The Event borrows from p.inner, so we capture the new
                // value into a flag and defer set_dv_reshape until after the
                // match drops the borrow.
                let mut dv_reshape: Option<bool> = None;
                let mut drained = false;
                let payload = match p.inner.wait_event(0.0) {
                    Some(Ok(event)) => match event {
                        Event::PropertyChange {
                            name: "video-params/colormatrix",
                            change: PropertyData::Str(s),
                            ..
                        } => {
                            dv_reshape = Some(s == "dolbyvision");
                            None
                        }
                        other => translate_event(other),
                    },
                    Some(Err(_)) => None,
                    None => {
                        drained = true;
                        None
                    }
                };
                if drained {
                    break;
                }
                if let Some(enable) = dv_reshape {
                    if let Err(e) = p.set_dv_reshape(enable) {
                        eprintln!("[velox-desktop] set_dv_reshape failed: {e}");
                    }
                }
                if let Some(payload) = payload {
                    if payload.kind == "shutdown" {
                        shutdown = true;
                    }
                    batch.push(payload);
                }
            }
            (batch, shutdown)
        };

        for payload in events {
            let _ = handle.emit("player-event", payload);
        }

        if shutdown {
            return;
        }
        std::thread::sleep(Duration::from_millis(50));
    }
}

fn translate_event(event: Event) -> Option<PlayerEvent> {
    match event {
        Event::StartFile => Some(PlayerEvent::simple("start-file")),
        Event::FileLoaded => Some(PlayerEvent::simple("file-loaded")),
        Event::PlaybackRestart => Some(PlayerEvent::simple("playback-restart")),
        Event::EndFile(_) => Some(PlayerEvent::simple("end-file")),
        Event::Shutdown => Some(PlayerEvent::simple("shutdown")),
        Event::Seek => Some(PlayerEvent::simple("seek")),
        Event::PropertyChange { name, change, .. } => match (name, change) {
            ("time-pos", PropertyData::Double(v)) => Some(PlayerEvent::float("time-pos", v)),
            ("duration", PropertyData::Double(v)) => Some(PlayerEvent::float("duration", v)),
            ("pause", PropertyData::Flag(v)) => Some(PlayerEvent::flag("pause", v)),
            ("paused-for-cache", PropertyData::Flag(v)) => {
                Some(PlayerEvent::flag("buffering", v))
            }
            _ => None,
        },
        _ => None,
    }
}
