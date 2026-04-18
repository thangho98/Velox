# fshare — Go client for fshare.vn API

Internal Velox package. Targets `api.fshare.vn` (mobile/desktop endpoint).
Handles hybrid token+cookie auth, the `code:201` session-expiry pattern, and
fshare's lack of 429/Retry-After headers.

## Quickstart

```go
c, _ := fshare.NewClient(fshare.Options{}) // uses DefaultAppKey + iOS UA

sess, err := c.Login(ctx, "you@example.com", "password")
// sess.Token + sess.SessionID populated. Email/AccountType stay empty —
// fetch via GetUserInfo below.

info, err := c.GetUserInfo(ctx)
// info.AccountType == "Vip" → downloads available. "Fee" → rejected on download.
// Side effect: client's session.Email + AccountType get populated.

items, err := c.ListFolder(ctx, "")        // root
items, err := c.ListFolder(ctx, "xyz123")  // nested folder by linkcode
for _, item := range items {
    if item.IsFolder() { /* recurse */ }
    fmt.Printf("%s %d bytes\n", item.Name, item.SizeBytes())
}

url, err := c.GetDirectLink(ctx, "abc456") // short-lived CDN URL — DO NOT CACHE
```

## Session restore pattern (resolver layer)

Resolver decrypts `cloud_sessions` row and hydrates the client without logging in:

```go
c, _ := fshare.NewClient(fshare.Options{})
c.RestoreSession(tokenFromDB, sessionIDFromDB)
c.SetCredentials(emailFromDB, passwordFromDB) // enables auto-relogin on code:201

items, err := c.ListFolder(ctx, folderCode)
// If the restored token is stale, ListFolder auto-relogins once and retries.
```

After a call, compare `c.Session().Token` vs the pre-call token to detect a
silent relogin and persist the new session back to the DB.

## Endpoint reference

All responses return **HTTP 200** on logical errors. Always parse the body
`code` field first.

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/user/login` | POST | `{user_email, password, app_key}` → `{code, msg, token, session_id}` — only these 4 fields |
| `/api/user/get` | GET | Healthy: returns full `UserInfo` object (id, email, account_type, expire_vip, …). Expired: returns `{code:201, msg:…}` |
| `/api/fileops/list` | GET | Root folder. `?pageIndex=&dirOnly=0&limit=60` |
| `/api/fileops/getFolderList` | POST | Nested folder. `{token, url, pageIndex, limit}` |
| `/api/session/download` | POST | Get direct URL. `{token, url, password, zipflag}` → `{location}` |

Pagination: **0-based** `pageIndex`, `limit=60` max observed.
Folder URL format: `https://www.fshare.vn/folder/<linkcode>`.
File URL format: `https://www.fshare.vn/file/<linkcode>`.

### Response field quirks

- All numeric fields in `FolderItem` and `UserInfo` are wire-encoded as **strings**
  (even `size`, `id`, `expire_vip`). Use the provided helpers (`SizeBytes()`,
  `ModifiedTime()`) to parse.
- `/api/user/login` does **not** return `email`/`account_type`. Call
  `GetUserInfo` separately if you need VIP status.
- Folders have empty `mimetype`; files always have one. `IsFolder()` encapsulates.
- `/api/user/get` response shape differs between healthy (object with no
  top-level `code`) and expired (object with `code:201`). The client sniffs
  with `isJSONObjectWithCode`.

## AppKey

The `app_key` field is **required** by fshare's login endpoint.
`DefaultAppKey` is verified live as of 2026-04. Override via `Options.AppKey`
if it ever gets revoked.

| Key | Source | Status (2026-04) |
|---|---|---|
| `L2S7R6ZMagggC5wWkQhX2+aDi467PPuftWUMRFSn` | tudoanh/get_fshare (default) | ✅ works |
| `GUxft6Beh3Bf8qKP7GC2IplYJZz1A53JQfRwne0R` | giangvo/synology-fshare | ✅ works |
| `dMnqMMZMUnN5YpvKENaEhdQQ5jxDqddt` | duythongle/fshare2gdrive (iOS sniff) | ❌ revoked — returns `code:37 Invalid User Agent` |

### Default User-Agent

`defaultUserAgent = "Fshare/1 CFNetwork/1209 Darwin/20.2.0"` (iOS Fshare app).
`okhttp/3.6.0` and browser UAs are rejected (`code:37`). Must be iOS-pattern UA.

## Error handling

| Error | Source | Meaning |
|---|---|---|
| `ErrInvalidCredentials` | login code != 200/201 | Email/password wrong or account locked |
| `ErrSessionExpired` | any call body `code==201` | Token stale; auto-relogin if credentials stashed |
| `ErrRateLimit` | retry exhausted after 5 attempts of 400s | fshare is throttling or flaky |
| `ErrFilePassword` | download HTTP 403 | File is password-protected |
| `ErrLinkDead` | download `code != 200` with empty location | File deleted or unavailable |
| `ErrNotVIP` | account_type != "Vip" on download | VIP required for `/session/download` |
| `ErrNotLoggedIn` | called before Login/RestoreSession | Programmer error |

## Retry behavior

`requestWithBackoff` retries only transient failures (400/5xx, network errors).
Exponential delays: **2s → 4s → 8s → 16s** (max 5 attempts, ~30s cumulative).
HTTP 403 is never retried (handled as `ErrFilePassword`).
fshare has no Retry-After header; delays are fixed.

## Session TTL

Empirically ~30 minutes (undocumented). Strategy:

1. **Lazy detection:** `withSessionRetry` catches `code:201` mid-call, re-logins
   with stashed credentials, retries once.
2. **Proactive refresh:** Scheduler (Plan W Phase 08) re-logins every ~25 min
   before the server kicks in.

## Download URL TTL

Minutes-scale, possibly single-use. **Never cache** the returned string.
Fetch fresh on every playback request. Android's OkHttp interceptor handles
reactive 403/404 refresh (Plan W Phase 07).

## Security

Credentials (email + password + token + session_id) must be **AES-GCM
encrypted at rest** with a per-install key (`VELOX_CLOUD_SECRET`). See
[Plan W Phase 02](../../../plans/plan-w-fshare-integration/phase-02-database-schema.md).

## References

API shape cross-referenced against 5 open-source fshare clients:

- [duythongle/fshare2gdrive](https://github.com/duythongle/fshare2gdrive) (NodeJS, 2025-01) — primary reference
- [dhhiep/fshare_tool](https://github.com/dhhiep/fshare_tool) (Ruby, 2024-08)
- [tudoanh/get_fshare](https://github.com/tudoanh/get_fshare) (Python, 2021)
- [giangvo/synology-fshare](https://github.com/giangvo/synology-fshare) (PHP, 2018)
- [haindvn/FShareDownloader](https://github.com/haindvn/FShareDownloader) (Python, 2020)

`pyload/pyload-plugins/FshareVn.py` uses the web UI (`www.fshare.vn/login.php`)
instead of the API. Velox intentionally picks the API path for stable JSON
responses, at the cost of needing `app_key`.
