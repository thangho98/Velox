# Phase 02: Database Schema + Encryption
Status: ⬜ Pending
Dependencies: Phase 01

## Objective

Thêm `storage_providers` table (1 row = 1 cloud account, reusable across multiple libraries). Extend `libraries` với foreign key tới provider + `source_url`. Viết AES-256-GCM encryption helper.

## Context

- Phase 01 đã có `pkg/fshare` hoạt động live với VIP account thật
- Phase 02 setup persistence layer cho credentials + library ↔ provider linkage
- Architecture đã chốt: storage provider = 1 cloud account; 1 provider → N libraries
- Encryption needed: fshare stores password (no refresh_token grant); GDrive/OneDrive store refresh_token (future)

## Implementation Steps

### 1. Encryption Helper
- [ ] Tạo `backend/pkg/crypto/aesgcm.go`:
  ```go
  package crypto

  // Encrypt returns nonce (12B) || ciphertext || tag.
  func Encrypt(key, plaintext []byte) ([]byte, error)
  func Decrypt(key, ciphertext []byte) ([]byte, error)

  // LoadKeyFromEnv reads hex-encoded 32-byte key from VELOX_CLOUD_SECRET.
  // Returns ErrNoKey if env unset, ErrInvalidKey if not 32 bytes hex.
  func LoadKeyFromEnv() ([]byte, error)
  ```
- [ ] Tests: `aesgcm_test.go`:
  - [ ] Round-trip (plaintext → encrypt → decrypt → same plaintext)
  - [ ] Tamper detection (modify ciphertext → decrypt fails)
  - [ ] Short key rejection (< 32 bytes → error)
  - [ ] Wrong key rejection (different key → decrypt fails)
  - [ ] Empty plaintext handling
- [ ] README note: generate key via `openssl rand -hex 32`; rotation requires re-login all providers

### 2. Migration 036 — `storage_providers` table

- [ ] File `backend/internal/database/migrate/036_storage_providers.go`:
  ```go
  package migrate

  var Migration036 = Migration{
      Version: 36,
      Name:    "storage_providers",
      Up: `
          CREATE TABLE storage_providers (
              id INTEGER PRIMARY KEY AUTOINCREMENT,
              user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
              provider_type TEXT NOT NULL,
              display_name TEXT NOT NULL,
              account_email TEXT NOT NULL,
              credentials_encrypted BLOB NOT NULL,
              account_type TEXT,
              quota_bytes INTEGER,
              used_bytes INTEGER,
              account_expires_at DATETIME,
              token_expires_at DATETIME,
              last_refresh_at DATETIME,
              last_check_at DATETIME,
              last_error TEXT,
              created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
              UNIQUE(user_id, provider_type, account_email)
          );
          CREATE INDEX idx_storage_providers_user ON storage_providers(user_id);
          CREATE INDEX idx_storage_providers_type ON storage_providers(provider_type);
      `,
  }
  ```
- [ ] Register trong `registry.go`

### 3. Migration 037 — extend `libraries`

- [ ] File `backend/internal/database/migrate/037_libraries_provider_link.go`:
  ```go
  var Migration037 = Migration{
      Version: 37,
      Name:    "libraries_provider_link",
      Up: `
          ALTER TABLE libraries ADD COLUMN storage_provider_id INTEGER REFERENCES storage_providers(id) ON DELETE SET NULL;
          ALTER TABLE libraries ADD COLUMN source_url TEXT;
          CREATE INDEX idx_libraries_provider ON libraries(storage_provider_id);
      `,
  }
  ```
- [ ] Register trong `registry.go`

**Semantics:**
- `storage_provider_id IS NULL` → local filesystem library (use existing `path` column)
- `storage_provider_id IS NOT NULL` → cloud library (use `source_url` + `storage_provider_id`)
- Zero migration side-effects: existing libraries auto-get NULL values

### 4. Models

- [ ] `backend/internal/model/storage_provider.go`:
  ```go
  type StorageProvider struct {
      ID                    int64      `json:"id"`
      UserID                int64      `json:"user_id"`
      ProviderType          string     `json:"provider_type"`  // "fshare"
      DisplayName           string     `json:"display_name"`
      AccountEmail          string     `json:"account_email"`
      CredentialsEncrypted  []byte     `json:"-"`              // never expose to API
      AccountType           *string    `json:"account_type,omitempty"`
      QuotaBytes            *int64     `json:"quota_bytes,omitempty"`
      UsedBytes             *int64     `json:"used_bytes,omitempty"`
      AccountExpiresAt      *time.Time `json:"account_expires_at,omitempty"`
      TokenExpiresAt        *time.Time `json:"token_expires_at,omitempty"`
      LastRefreshAt         *time.Time `json:"last_refresh_at,omitempty"`
      LastCheckAt           *time.Time `json:"last_check_at,omitempty"`
      LastError             *string    `json:"last_error,omitempty"`
      CreatedAt             time.Time  `json:"created_at"`
  }

  // HealthStatus derives from token_expires_at, last_error, last_check_at.
  type HealthStatus string
  const (
      HealthHealthy      HealthStatus = "healthy"
      HealthExpiringSoon HealthStatus = "expiring_soon"
      HealthExpired      HealthStatus = "expired"
      HealthError        HealthStatus = "error"
      HealthUnknown      HealthStatus = "unknown"
  )

  func (p *StorageProvider) ComputeHealth(now time.Time) HealthStatus { /* ... */ }
  ```

- [ ] Extend `backend/internal/model/library.go`:
  ```go
  type Library struct {
      // ... existing fields
      StorageProviderID *int64  `json:"storage_provider_id,omitempty"`
      SourceURL         *string `json:"source_url,omitempty"`
  }

  func (l *Library) IsCloudLibrary() bool {
      return l.StorageProviderID != nil
  }
  ```

### 5. Repository

- [ ] Tạo `backend/internal/repository/storage_provider.go`:
  ```go
  type StorageProviderRepo struct { db DBTX }

  func NewStorageProviderRepo(db DBTX) *StorageProviderRepo

  // CRUD
  func (r *StorageProviderRepo) Create(ctx context.Context, p *StorageProvider) error
  func (r *StorageProviderRepo) GetByID(ctx context.Context, id int64) (*StorageProvider, error) // wraps sql.ErrNoRows → ErrNotFound
  func (r *StorageProviderRepo) GetByAccount(ctx context.Context, userID int64, providerType, email string) (*StorageProvider, error)
  func (r *StorageProviderRepo) Update(ctx context.Context, p *StorageProvider) error // RowsAffected check
  func (r *StorageProviderRepo) Delete(ctx context.Context, id int64) error
  func (r *StorageProviderRepo) ListByUser(ctx context.Context, userID int64) ([]*StorageProvider, error)
  func (r *StorageProviderRepo) ListAll(ctx context.Context) ([]*StorageProvider, error) // for admin dashboard

  // For scheduler: find providers needing refresh
  func (r *StorageProviderRepo) ListExpiringSoon(ctx context.Context, beforeTime time.Time) ([]*StorageProvider, error)

  // Atomic credential update (called by scheduler + auto-refresh)
  func (r *StorageProviderRepo) UpdateCredentials(ctx context.Context, id int64, credsEncrypted []byte, tokenExpiresAt time.Time) error
  ```

- [ ] Extend `backend/internal/repository/library.go`:
  - [ ] Update `Create`, `Get`, `List`, `Update` queries để bao gồm `storage_provider_id` + `source_url`
  - [ ] Add `ListByProvider(ctx, providerID)` — list libraries sharing a provider (for delete confirmation)

### 6. Config + Startup Wiring

- [ ] Extend `backend/internal/config/config.go`:
  ```go
  type CloudConfig struct {
      Enabled      bool   // VELOX_CLOUD_SOURCES_ENABLED
      SecretKey    []byte // decoded from VELOX_CLOUD_SECRET (hex)
      FshareAppKey string // VELOX_FSHARE_APP_KEY (optional — uses pkg/fshare.DefaultAppKey)
  }
  ```
- [ ] `Load()` behavior:
  - `Enabled=false` → skip all cloud config validation (backwards compat)
  - `Enabled=true` → require valid `SecretKey` (fail fast)
  - `FshareAppKey` empty OK → driver uses `DefaultAppKey`
- [ ] Wire `StorageProviderRepo` vào `ServerApp` trong `cmd/server/server_app.go`

## Acceptance Criteria

- [ ] `make migrate` runs clean: `cd backend && go run ./cmd/server migrate up`
- [ ] Existing libraries auto-get `storage_provider_id = NULL` (backwards compat)
- [ ] `storage_providers` table exists với đầy đủ columns + indexes
- [ ] Unit tests pass:
  - [ ] `crypto.Encrypt/Decrypt` round-trip + tamper detection
  - [ ] `StorageProviderRepo` CRUD happy path
  - [ ] `StorageProviderRepo.GetByID` returns `repository.ErrNotFound` khi không tồn tại
  - [ ] `StorageProviderRepo.Update` returns ErrNotFound khi RowsAffected=0
- [ ] `config.Load()` fail với error rõ ràng khi `VELOX_CLOUD_SOURCES_ENABLED=true` nhưng `VELOX_CLOUD_SECRET` missing/invalid
- [ ] `go vet` + `golangci-lint` clean

## Gotchas

- SQLite ADD COLUMN với foreign key: OK, không validate existing rows. Default NULL acceptable.
- `credentials_encrypted BLOB`: Go `[]byte` binding OK với `database/sql`.
- Encryption key rotation: NOT in scope phase này. Documented as manual process (stop server → change key → re-register all providers).
- DBTX interface: reuse existing pattern.
- **Storing password at rest** (fshare): larger attack surface vs OAuth refresh tokens. Mitigate: per-install AES key, BLOB column, zero plaintext logs, warn admin UI.
- **Session TTL là local bookkeeping** — authoritative check is API body `code == 201`. `token_expires_at` = estimated (last_refresh + 25min).

## Out of Scope

- Key rotation tool (manual for MVP)
- Multi-user provider sharing (1 provider = 1 user; UNIQUE constraint enforces)
- Provider ACL (per-library access control)
- Credentials audit log
