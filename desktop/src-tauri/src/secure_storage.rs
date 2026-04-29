// File-backed K/V store at $APP_LOCAL_DATA/secure_store.json.
// Originally backed by the macOS Keychain via the `keyring` crate, but the
// per-rebuild signature change forced a Keychain ACL prompt on every launch.
// The threat model (LAN NAS, trusted device, 7-day refresh token) doesn't
// justify that friction, so tokens now live in the user's app data dir at
// rest — same trust level the webapp already grants `localStorage`.

use serde_json::{Map, Value};
use std::fs;
use std::path::PathBuf;
use std::sync::Mutex;
use tauri::Manager;

const STORE_FILE: &str = "secure_store.json";

// Serialize concurrent commands so a racing set/remove pair can't lose writes.
static FILE_LOCK: Mutex<()> = Mutex::new(());

fn store_path(app: &tauri::AppHandle) -> Result<PathBuf, String> {
    let dir = app
        .path()
        .app_local_data_dir()
        .map_err(|e| format!("app_local_data_dir: {e}"))?;
    fs::create_dir_all(&dir).map_err(|e| format!("mkdir {}: {e}", dir.display()))?;
    Ok(dir.join(STORE_FILE))
}

fn read_store(path: &PathBuf) -> Result<Map<String, Value>, String> {
    if !path.exists() {
        return Ok(Map::new());
    }
    let body = fs::read_to_string(path).map_err(|e| format!("read store: {e}"))?;
    match serde_json::from_str::<Value>(&body) {
        Ok(Value::Object(m)) => Ok(m),
        _ => Ok(Map::new()),
    }
}

fn write_store(path: &PathBuf, store: &Map<String, Value>) -> Result<(), String> {
    let body = serde_json::to_vec_pretty(store).map_err(|e| format!("encode store: {e}"))?;
    let tmp = path.with_extension("json.tmp");
    fs::write(&tmp, &body).map_err(|e| format!("write tmp: {e}"))?;
    fs::rename(&tmp, path).map_err(|e| format!("rename tmp: {e}"))?;
    Ok(())
}

#[tauri::command]
pub fn secure_get(app: tauri::AppHandle, key: String) -> Result<Option<String>, String> {
    let _guard = FILE_LOCK.lock().unwrap_or_else(|e| e.into_inner());
    let path = store_path(&app)?;
    let store = read_store(&path)?;
    Ok(store.get(&key).and_then(|v| v.as_str().map(String::from)))
}

#[tauri::command]
pub fn secure_set(app: tauri::AppHandle, key: String, value: String) -> Result<(), String> {
    let _guard = FILE_LOCK.lock().unwrap_or_else(|e| e.into_inner());
    let path = store_path(&app)?;
    let mut store = read_store(&path)?;
    store.insert(key, Value::String(value));
    write_store(&path, &store)
}

#[tauri::command]
pub fn secure_remove(app: tauri::AppHandle, key: String) -> Result<(), String> {
    let _guard = FILE_LOCK.lock().unwrap_or_else(|e| e.into_inner());
    let path = store_path(&app)?;
    let mut store = read_store(&path)?;
    store.remove(&key);
    write_store(&path, &store)
}
