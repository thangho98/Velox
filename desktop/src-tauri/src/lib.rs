mod hls_prewarm;
mod player;
mod secure_storage;

#[cfg(target_os = "macos")]
mod gl_view;

#[cfg(target_os = "macos")]
mod mac_embed;

use parking_lot::Mutex;
use serde::Serialize;
use std::sync::Arc;
use tauri::{
    menu::{Menu, MenuItem, PredefinedMenuItem, Submenu},
    AppHandle, Emitter, Manager, State, WindowEvent,
};

pub struct AppState {
    player: Arc<Mutex<player::Player>>,
}

#[derive(Serialize)]
struct VersionInfo {
    mpv: String,
    ffmpeg: String,
    placebo: Option<String>,
    hwdec_codecs: String,
}

#[tauri::command]
fn get_versions(state: State<AppState>) -> Result<VersionInfo, String> {
    let p = state.player.lock();
    Ok(VersionInfo {
        mpv: p.get_str("mpv-version").unwrap_or_default(),
        ffmpeg: p.get_str("ffmpeg-version").unwrap_or_default(),
        placebo: p.get_str("libplacebo-version").ok(),
        hwdec_codecs: p.get_str("hwdec-codecs").unwrap_or_default(),
    })
}

#[tauri::command]
fn player_load(
    url: String,
    start_time: Option<f64>,
    state: State<AppState>,
) -> Result<(), String> {
    if hls_prewarm::looks_like_hls(&url) {
        eprintln!("[player_load] pre-warming HLS playlist for {}", url);
        if let Err(e) =
            hls_prewarm::wait_for_first_segment(&url, std::time::Duration::from_secs(60))
        {
            return Err(format!("hls prewarm: {}", e));
        }
        eprintln!("[player_load] pre-warm done — handing off to mpv");
    }
    state
        .player
        .lock()
        .load(&url, start_time)
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn player_play(state: State<AppState>) -> Result<(), String> {
    state.player.lock().resume().map_err(|e| e.to_string())
}

#[tauri::command]
fn player_pause(state: State<AppState>) -> Result<(), String> {
    state.player.lock().pause().map_err(|e| e.to_string())
}

#[tauri::command]
fn player_stop(state: State<AppState>) -> Result<(), String> {
    state.player.lock().stop().map_err(|e| e.to_string())
}

#[tauri::command]
fn player_seek(seconds: f64, state: State<AppState>) -> Result<(), String> {
    state.player.lock().seek(seconds).map_err(|e| e.to_string())
}

#[tauri::command]
fn player_set_volume(volume: f64, state: State<AppState>) -> Result<(), String> {
    state
        .player
        .lock()
        .set_volume(volume)
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn player_set_muted(muted: bool, state: State<AppState>) -> Result<(), String> {
    state
        .player
        .lock()
        .set_muted(muted)
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn player_set_rate(rate: f64, state: State<AppState>) -> Result<(), String> {
    state.player.lock().set_rate(rate).map_err(|e| e.to_string())
}

#[tauri::command]
fn player_sub_add(
    url: String,
    lang: String,
    label: String,
    state: State<AppState>,
) -> Result<(), String> {
    state
        .player
        .lock()
        .sub_add(&url, &lang, &label)
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn player_sub_remove(state: State<AppState>) -> Result<(), String> {
    state.player.lock().sub_remove().map_err(|e| e.to_string())
}

#[tauri::command]
fn player_sub_delay(delay_ms: f64, state: State<AppState>) -> Result<(), String> {
    state
        .player
        .lock()
        .set_sub_delay(delay_ms)
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn player_sub_visible(visible: bool, state: State<AppState>) -> Result<(), String> {
    state
        .player
        .lock()
        .set_sub_visible(visible)
        .map_err(|e| e.to_string())
}

#[tauri::command]
fn player_audio_set_lang(
    lang: String,
    name: Option<String>,
    state: State<AppState>,
) -> Result<(), String> {
    state
        .player
        .lock()
        .set_audio_track_by_lang(&lang, name.as_deref())
        .map_err(|e| e.to_string())
}

#[derive(Serialize)]
struct PositionInfo {
    position: f64,
    duration: f64,
    paused: bool,
}

#[tauri::command]
fn player_position(state: State<AppState>) -> Result<PositionInfo, String> {
    let p = state.player.lock();
    Ok(PositionInfo {
        position: p.position().unwrap_or(0.0),
        duration: p.duration().unwrap_or(0.0),
        paused: p.paused().unwrap_or(true),
    })
}

#[derive(Serialize)]
struct ColorInfo {
    colormatrix: Option<String>,
    primaries: Option<String>,
    gamma: Option<String>,
    sig_peak: Option<String>,
    is_dolby_vision: bool,
    width: Option<String>,
    height: Option<String>,
    current_vo: Option<String>,
}

#[tauri::command]
fn color_info(state: State<AppState>) -> Result<ColorInfo, String> {
    let p = state.player.lock();
    let cm = p.get_str("video-params/colormatrix").ok();
    let is_dv = cm.as_deref() == Some("dolbyvision");
    Ok(ColorInfo {
        is_dolby_vision: is_dv,
        colormatrix: cm,
        primaries: p.get_str("video-params/primaries").ok(),
        gamma: p.get_str("video-params/gamma").ok(),
        sig_peak: p.get_str("video-params/sig-peak").ok(),
        width: p.get_str("video-params/w").ok(),
        height: p.get_str("video-params/h").ok(),
        current_vo: p.get_str("current-vo").ok(),
    })
}

#[tauri::command]
fn set_fullscreen(window: tauri::Window, fullscreen: bool) -> Result<(), String> {
    window.set_fullscreen(fullscreen).map_err(|e| e.to_string())
}

#[tauri::command]
fn is_fullscreen(window: tauri::Window) -> Result<bool, String> {
    window.is_fullscreen().map_err(|e| e.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let player = player::Player::new().expect("failed to init libmpv");

    tauri::Builder::default()
        .manage(AppState {
            player: Arc::new(Mutex::new(player)),
        })
        .plugin(tauri_plugin_single_instance::init(|app, argv, _cwd| {
            // Forward velox:// args from the second instance to the running app.
            // The deep-link plugin emits a deep-link://new-url event we listen to below.
            let _ = app.emit("single-instance", argv);
            if let Some(window) = app.get_webview_window("main") {
                let _ = window.set_focus();
            }
        }))
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .plugin(tauri_plugin_window_state::Builder::default().build())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_deep_link::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .invoke_handler(tauri::generate_handler![
            get_versions,
            player_load,
            player_play,
            player_pause,
            player_stop,
            player_seek,
            player_set_volume,
            player_set_muted,
            player_set_rate,
            player_sub_add,
            player_sub_remove,
            player_sub_delay,
            player_sub_visible,
            player_audio_set_lang,
            player_position,
            color_info,
            set_fullscreen,
            is_fullscreen,
            secure_storage::secure_get,
            secure_storage::secure_set,
            secure_storage::secure_remove,
        ])
        .setup(|app| {
            let handle = app.handle().clone();
            let state = app.state::<AppState>();
            let player = state.player.clone();

            // Install native video view + render context (macOS).
            #[cfg(target_os = "macos")]
            {
                if let Some(window) = app.get_webview_window("main") {
                    let mpv_handle = state.player.lock().raw_handle();
                    if let Err(e) = mac_embed::install_video_layer(&window, mpv_handle) {
                        eprintln!("[velox-desktop] failed to install video layer: {e}");
                    } else {
                        eprintln!("[velox-desktop] video layer installed");
                    }
                }
            }

            // Mac native menu bar (File / View / Playback / Window / Help).
            #[cfg(target_os = "macos")]
            {
                if let Err(e) = install_menu(&handle) {
                    eprintln!("[velox-desktop] failed to install menu: {e}");
                }
            }

            // Listen for deep-link events (velox://...) and forward to webview.
            // tauri-plugin-deep-link emits "deep-link://new-url" with a Vec<Url>.
            {
                let handle_clone = handle.clone();
                use tauri_plugin_deep_link::DeepLinkExt;
                let _ = handle.deep_link().on_open_url(move |event| {
                    let urls: Vec<String> =
                        event.urls().iter().map(|u| u.to_string()).collect();
                    eprintln!("[velox-desktop] deep-link: {:?}", urls);
                    let _ = handle_clone.emit("velox-deep-link", urls);
                });
            }

            // Stop mpv cleanly when the main window is about to close.
            if let Some(window) = app.get_webview_window("main") {
                let player_for_close = player.clone();
                window.on_window_event(move |ev| {
                    if let WindowEvent::CloseRequested { .. } = ev {
                        let _ = player_for_close.lock().stop();
                    }
                });
            }

            std::thread::spawn(move || {
                player::run_event_loop(player, handle);
            });
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(target_os = "macos")]
fn install_menu(handle: &AppHandle) -> tauri::Result<()> {
    // App submenu (Mac convention: app name first).
    let app_menu = Submenu::with_items(
        handle,
        "Velox",
        true,
        &[
            &PredefinedMenuItem::about(handle, Some("About Velox"), None)?,
            &PredefinedMenuItem::separator(handle)?,
            &PredefinedMenuItem::services(handle, None)?,
            &PredefinedMenuItem::separator(handle)?,
            &PredefinedMenuItem::hide(handle, None)?,
            &PredefinedMenuItem::hide_others(handle, None)?,
            &PredefinedMenuItem::show_all(handle, None)?,
            &PredefinedMenuItem::separator(handle)?,
            &PredefinedMenuItem::quit(handle, None)?,
        ],
    )?;

    let file_menu = Submenu::with_items(
        handle,
        "File",
        true,
        &[
            &MenuItem::with_id(handle, "open-file", "Open File…", true, Some("Cmd+O"))?,
            &MenuItem::with_id(handle, "open-url", "Open URL…", true, Some("Cmd+Shift+O"))?,
            &PredefinedMenuItem::separator(handle)?,
            &PredefinedMenuItem::close_window(handle, None)?,
        ],
    )?;

    let edit_menu = Submenu::with_items(
        handle,
        "Edit",
        true,
        &[
            &PredefinedMenuItem::undo(handle, None)?,
            &PredefinedMenuItem::redo(handle, None)?,
            &PredefinedMenuItem::separator(handle)?,
            &PredefinedMenuItem::cut(handle, None)?,
            &PredefinedMenuItem::copy(handle, None)?,
            &PredefinedMenuItem::paste(handle, None)?,
            &PredefinedMenuItem::select_all(handle, None)?,
        ],
    )?;

    let view_menu = Submenu::with_items(
        handle,
        "View",
        true,
        &[
            &MenuItem::with_id(handle, "toggle-fullscreen", "Enter Full Screen", true, Some("Cmd+Ctrl+F"))?,
            &MenuItem::with_id(handle, "reload", "Reload", true, Some("Cmd+R"))?,
        ],
    )?;

    let playback_menu = Submenu::with_items(
        handle,
        "Playback",
        true,
        &[
            &MenuItem::with_id(handle, "playback-play-pause", "Play / Pause", true, Some("Space"))?,
            &PredefinedMenuItem::separator(handle)?,
            &MenuItem::with_id(handle, "playback-seek-back", "Seek -10s", true, Some("Left"))?,
            &MenuItem::with_id(handle, "playback-seek-fwd", "Seek +10s", true, Some("Right"))?,
        ],
    )?;

    let window_menu = Submenu::with_items(
        handle,
        "Window",
        true,
        &[
            &PredefinedMenuItem::minimize(handle, None)?,
            &PredefinedMenuItem::maximize(handle, None)?,
        ],
    )?;

    let menu = Menu::with_items(
        handle,
        &[&app_menu, &file_menu, &edit_menu, &view_menu, &playback_menu, &window_menu],
    )?;
    handle.set_menu(menu)?;

    // Menu actions are forwarded to the webview as "velox-menu" events.
    let app_handle = handle.clone();
    handle.on_menu_event(move |_app, event| {
        let id = event.id().0.as_str();
        let _ = app_handle.emit("velox-menu", id);
    });
    Ok(())
}
