# Phase 02: Database Schema + Encryption
Status: ⬜ Pending
Dependencies: Phase 01

## Objective

Extend `libraries` table để support multiple source types. Thêm `cloud_sessions` table để persist encrypted fshare credentials. Viết AES-GCM encryption helper.

## Context

- Current [backend/internal/model/library.go](backend/internal/model/library.go) chỉ có local filesystem path
- Migration runner: [backend/internal/database/migrate/registry.go](backend/internal/database/migrate/registry.go) — đã có 35 migrations
- Encryption cần cho fshare token (full account access) — PHẢI mã hoá at rest
- Decision: AES-GCM với key từ env `VELOX_CLOUD_SECRET` (32-byte hex, generate 1 lần)
- **Store cả email + password encrypted** (KHÔNG chỉ token) — fshare session ~30min TTL, cần re-login autonomous từ scheduler (Phase 08). Password-only persistence là điều kiện tiên quyết.

## Implementation Steps

### 1. Encryption Helper Package
- [ ] Tạo `backend/pkg/crypto/aesgcm.go`:
  ```go
  package crypto

  // Encrypt using AES-256-GCM. Key must be 32 bytes.
  // Returns: nonce (12B) || ciphertext || tag.
  func Encrypt(key, plaintext []byte) ([]byte, error)
  func Decrypt(key, ciphertext []byte) ([]byte, error)

  // LoadKeyFromEnv reads hex-encoded 32-byte key from VELOX_CLOUD_SECRET.
  // Returns error nếu env empty (cloud features disabled) hoặc key invalid.
  func LoadKeyFromEnv() ([]byte, error)
  ```
- [ ] Tests: `aesgcm_test.go` — round-trip, tamper detection, short key rejection
- [ ] Document cách generate: `openssl rand -hex 32`

### 2. Migration 036 — Extend `libraries`
- [ ] File `backend/internal/database/migrate/036_libraries_source_type.go`:
  ```go
  package migrate

  var Migration036 = Migration{
      Version: 36,
      Name:    "libraries_source_type",
      Up: `
          ALTER TABLE libraries ADD COLUMN source_type TEXT NOT NULL DEFAULT 'local';
          ALTER TABLE libraries ADD COLUMN source_credentials_id INTEGER REFERENCES cloud_sessions(id) ON DELETE SET NULL;
          ALTER TABLE libraries ADD COLUMN source_root_id TEXT; -- fshare folder linkcode
          CREATE INDEX idx_libraries_source_type ON libraries(source_type);
      `,
  }
  ```
- [ ] `source_type` CHECK constraint: `IN ('local', 'fshare')` (SQLite: add qua trigger hoặc rely on application-level enum)
- [ ] Register trong `registry.go`

### 3. Migration 037 — `cloud_sessions` table
- [ ] File `backend/internal/database/migrate/037_cloud_sessions.go`:
  ```go
  var Migration037 = Migration{
      Version: 37,
      Name:    "cloud_sessions",
      Up: `
          CREATE TABLE cloud_sessions (
              id INTEGER PRIMARY KEY AUTOINCREMENT,
              user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
              provider TEXT NOT NULL,
              account_email TEXT NOT NULL,
              credentials_encrypted BLOB NOT NULL, -- AES-GCM(email + password) — needed for auto re-login
              token_encrypted BLOB,                -- AES-GCM(active token from last login) — nullable before first login
              session_id_encrypted BLOB,           -- AES-GCM(cookie session_id) — nullable
              token_expires_at DATETIME,           -- estimated expiry (last_login + 30min). Soft value, confirm via CheckSession.
              last_refresh_at DATETIME,            -- time of last successful login/refresh
              last_error TEXT,
              account_type TEXT,                   -- "Vip" | "Fee" — from login response
              created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
              UNIQUE(user_id, provider, account_email)
          );
          CREATE INDEX idx_cloud_sessions_provider ON cloud_sessions(provider);
          CREATE INDEX idx_cloud_sessions_user ON cloud_sessions(user_id);
      `,
  }
  ```
- [ ] Register trong `registry.go`
- [ ] Migration 037 PHẢI chạy SAU 036 (foreign key dependency)

### 4. Models + Repository
- [ ] Extend `backend/internal/model/library.go`:
  ```go
  type Library struct {
      // ... existing fields
      SourceType          string     `json:"source_type"` // "local" | "fshare"
      SourceCredentialsID *int64     `json:"source_credentials_id,omitempty"`
      SourceRootID        *string    `json:"source_root_id,omitempty"`
  }

  const (
      SourceTypeLocal  = "local"
      SourceTypeFshare = "fshare"
  )
  ```

- [ ] Tạo `backend/internal/model/cloud_session.go`:
  ```go
  type CloudSession struct {
      ID                    int64      `json:"id"`
      UserID                int64      `json:"user_id"`
      Provider              string     `json:"provider"` // "fshare"
      AccountEmail          string     `json:"account_email"`
      CredentialsEncrypted  []byte     `json:"-"` // AES-GCM blob {email, password}
      TokenEncrypted        []byte     `json:"-"` // AES-GCM blob — active session token
      SessionIDEncrypted    []byte     `json:"-"` // AES-GCM blob — cookie session_id
      TokenExpiresAt        *time.Time `json:"token_expires_at,omitempty"` // estimated, ~last_login + 30min
      LastRefreshAt         *time.Time `json:"last_refresh_at,omitempty"`
      LastError             *string    `json:"last_error,omitempty"`
      AccountType           *string    `json:"account_type,omitempty"` // "Vip" | "Fee"
      CreatedAt             time.Time  `json:"created_at"`
  }

  // CredentialsPayload is what gets encrypted in credentials_encrypted.
  type CredentialsPayload struct {
      Email    string `json:"email"`
      Password string `json:"password"`
  }
  ```

- [ ] Tạo `backend/internal/repository/cloud_session.go`:
  ```go
  type CloudSessionRepo struct { db DBTX }

  func (r *CloudSessionRepo) Create(ctx context.Context, s *CloudSession) error
  func (r *CloudSessionRepo) GetByID(ctx context.Context, id int64) (*CloudSession, error) // wraps sql.ErrNoRows → ErrNotFound
  func (r *CloudSessionRepo) GetByUserProvider(ctx context.Context, userID int64, provider, email string) (*CloudSession, error)
  func (r *CloudSessionRepo) Update(ctx context.Context, s *CloudSession) error // checks RowsAffected
  func (r *CloudSessionRepo) Delete(ctx context.Context, id int64) error
  func (r *CloudSessionRepo) ListByUser(ctx context.Context, userID int64) ([]*CloudSession, error)
  ```

- [ ] Extend `backend/internal/repository/library.go`:
  - [ ] Update `Create`, `Get`, `List`, `Update` queries để bao gồm 3 columns mới
  - [ ] Default `source_type = 'local'` ở application level khi legacy create

### 5. Config + Startup Wiring
- [ ] Extend `backend/internal/config/config.go`:
  ```go
  type CloudConfig struct {
      Enabled      bool   // VELOX_CLOUD_SOURCES_ENABLED
      SecretKey    []byte // decoded from VELOX_CLOUD_SECRET (hex)
      FshareAppKey string // VELOX_FSHARE_APP_KEY
  }
  ```
- [ ] Load trong `Load()`. Nếu `Enabled=true` nhưng `SecretKey` empty/invalid → return error (fail fast)
- [ ] Nếu `Enabled=false` → skip key validation (backwards compat)
- [ ] Wire `CloudSessionRepo` vào `ServerApp` trong `cmd/server/server_app.go`

## Acceptance Criteria

- [ ] `make migrate` runs clean: `cd backend && go run ./cmd/server migrate up`
- [ ] Existing libraries rows auto-get `source_type = 'local'` (default value)
- [ ] `cloud_sessions` table exists với đầy đủ columns + indexes
- [ ] Unit tests pass cho `crypto.Encrypt/Decrypt` (round-trip, tamper detect)
- [ ] Unit tests pass cho `CloudSessionRepo` (CRUD happy path + ErrNotFound)
- [ ] `config.Load()` fail với error rõ ràng khi `VELOX_CLOUD_SOURCES_ENABLED=true` nhưng secret missing/invalid
- [ ] `go vet` + `golangci-lint run` clean

## Gotchas

- SQLite `ALTER TABLE ADD COLUMN` với foreign key: SQLite cho phép, nhưng không validate existing rows. Default NULL OK cho `source_credentials_id` (legacy local libs).
- `*_encrypted BLOB`: Go `[]byte` binding OK với `database/sql`.
- Encryption key rotation: KHÔNG scope phase này. Document in README: "Rotating VELOX_CLOUD_SECRET requires re-login all cloud sessions".
- DBTX interface: reuse existing pattern (tx hoặc *sql.DB compatible).
- **Session TTL thực tế ~30 phút (không phải 24h)** — research từ 5 OSS repos xác nhận. `TokenExpiresAt` là estimated (last_login + 30min); authoritative check là body `code == 201` từ API call. Phase 08 scheduler refresh mỗi 25min preemptive.
- **Storing password at rest** = larger attack surface. Mitigate: AES-256-GCM với unique per-install key (`VELOX_CLOUD_SECRET`), BLOB column, no plaintext logging. Alternative (token refresh grant) không available trong fshare API.

## Out of Scope

- Key rotation tool (manual process for now)
- Multiple providers sharing same cloud_sessions row (each provider = own row)
- Audit log cho session modifications (phase sau nếu cần)
