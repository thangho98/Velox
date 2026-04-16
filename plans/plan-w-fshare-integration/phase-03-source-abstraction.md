# Phase 03: CloudStorage Abstraction + Fshare Driver
Status: ⬜ Pending
Dependencies: Phase 02

## Objective

Tạo `internal/cloudstorage/` package với Provider + Driver interfaces. Implement Fshare driver wrapping `pkg/fshare`. Registry cho phép thêm Google Drive / OneDrive drivers sau này without touching scanner / stream handler.

## Context

Architecture đã chốt trong [plan.md](plan.md):
- **Provider** = operates on authenticated cloud session (list, download URL, health)
- **Driver** = factory + auth flow (build Provider from credentials)
- **Registry** = startup singleton; dispatch by provider_type string

Phase 01 `pkg/fshare` là raw API client (portable). Phase 03 là Velox-specific abstraction layer that adapts pkg/fshare to the Provider interface.

## Package Structure

```
backend/internal/cloudstorage/
├── provider.go       # Provider + Driver + PasswordAuthDriver + OAuthDriver interfaces
├── registry.go       # Registry singleton
├── types.go          # Item, AccountInfo, Credentials, AuthFlow
├── errors.go         # Shared errors (ErrSessionExpired, ErrRateLimit, ErrInvalidCredentials)
├── credentials.go    # Serialize/deserialize + AES-GCM integration
├── provider_test.go  # Interface contract tests
└── drivers/
    └── fshare/
        ├── driver.go        # implements Driver + PasswordAuthDriver
        ├── provider.go      # implements Provider (wraps pkg/fshare.Client)
        ├── url.go           # URL parsing: /folder/{code}, /file/{code}
        ├── driver_test.go
        └── url_test.go
```

## Implementation Steps

### 1. Core Interfaces + Types

- [ ] `backend/internal/cloudstorage/provider.go`:
  ```go
  package cloudstorage

  import "context"

  // Provider operates on an authenticated cloud session.
  // One Provider instance = one cloud account.
  type Provider interface {
      Type() string

      // URL parsing — user-pasted URL → native ID.
      ParseFolderURL(url string) (folderID string, err error)
      ParseFileURL(url string) (fileID string, err error)

      // Operations.
      ListFolder(ctx context.Context, folderID string) ([]Item, error)
      GetDownloadURL(ctx context.Context, fileID string) (string, error)
      GetAccountInfo(ctx context.Context) (*AccountInfo, error)
      CheckHealth(ctx context.Context) error
  }

  // Driver is the factory + auth flow for a provider type.
  type Driver interface {
      Type() string
      DisplayName() string
      AuthFlow() AuthFlow
      NewProvider(creds *Credentials) (Provider, error)
  }

  // PasswordAuthDriver is implemented by drivers using email/password auth.
  type PasswordAuthDriver interface {
      Driver
      AuthenticatePassword(ctx context.Context, email, password string) (*Credentials, error)
  }

  // OAuthDriver is implemented by drivers using OAuth 2.0 code flow.
  type OAuthDriver interface {
      Driver
      AuthURL(state, redirectURI string) string
      ExchangeCode(ctx context.Context, code, redirectURI string) (*Credentials, error)
      RefreshToken(ctx context.Context, creds *Credentials) (*Credentials, error)
  }
  ```

### 2. Value Types

- [ ] `backend/internal/cloudstorage/types.go`:
  ```go
  type AuthFlow int
  const (
      AuthFlowPassword AuthFlow = iota
      AuthFlowOAuth
  )

  // Item is the uniform cloud file/folder representation.
  type Item struct {
      ID       string
      Name     string
      IsFolder bool
      Size     int64
      Mimetype string
      Modified time.Time
      URL      string // canonical URL (for display / paste-back)
  }

  // AccountInfo normalizes provider-specific account metadata.
  type AccountInfo struct {
      Email       string
      DisplayName string
      AccountType string    // "Vip" (fshare), "Paid" (gdrive), etc.
      QuotaBytes  int64
      UsedBytes   int64
      ExpiresAt   time.Time // VIP expiry (zero = no subscription)
  }

  // Credentials is the encrypted-at-rest auth state. `Meta` holds
  // provider-specific fields that don't fit standard slots.
  type Credentials struct {
      ProviderType string            `json:"provider_type"`
      Email        string            `json:"email"`
      Password     string            `json:"password,omitempty"`      // password-auth providers
      AccessToken  string            `json:"access_token,omitempty"`  // both
      SessionID    string            `json:"session_id,omitempty"`    // password-auth (cookie)
      RefreshToken string            `json:"refresh_token,omitempty"` // OAuth only
      ExpiresAt    time.Time         `json:"expires_at"`
      Meta         map[string]string `json:"meta,omitempty"`
  }
  ```

### 3. Registry

- [ ] `backend/internal/cloudstorage/registry.go`:
  ```go
  type Registry struct {
      mu      sync.RWMutex
      drivers map[string]Driver
  }

  func NewRegistry() *Registry {
      return &Registry{drivers: map[string]Driver{}}
  }

  func (r *Registry) Register(d Driver) error {
      r.mu.Lock()
      defer r.mu.Unlock()
      if _, exists := r.drivers[d.Type()]; exists {
          return fmt.Errorf("cloudstorage: driver %q already registered", d.Type())
      }
      r.drivers[d.Type()] = d
      return nil
  }

  func (r *Registry) Get(providerType string) (Driver, error) {
      r.mu.RLock()
      defer r.mu.RUnlock()
      d, ok := r.drivers[providerType]
      if !ok {
          return nil, fmt.Errorf("%w: %s", ErrUnknownProvider, providerType)
      }
      return d, nil
  }

  type DriverMeta struct {
      Type        string   `json:"type"`
      DisplayName string   `json:"display_name"`
      AuthFlow    AuthFlow `json:"auth_flow"`
  }

  // List returns driver metadata for admin UI dropdown.
  func (r *Registry) List() []DriverMeta { /* ... */ }
  ```

### 4. Typed Errors

- [ ] `backend/internal/cloudstorage/errors.go`:
  ```go
  var (
      ErrUnknownProvider    = errors.New("cloudstorage: unknown provider type")
      ErrSessionExpired     = errors.New("cloudstorage: session expired")
      ErrInvalidCredentials = errors.New("cloudstorage: invalid credentials")
      ErrRateLimit          = errors.New("cloudstorage: rate limited")
      ErrFileNotFound       = errors.New("cloudstorage: file not found")
      ErrInvalidURL         = errors.New("cloudstorage: url not recognized for this provider")
      ErrNotSupported       = errors.New("cloudstorage: operation not supported")
      ErrAccountNotPremium  = errors.New("cloudstorage: account tier does not support downloads")
  )
  ```

### 5. Credentials Encryption Helper

- [ ] `backend/internal/cloudstorage/credentials.go`:
  ```go
  // EncryptCredentials serializes Credentials to JSON then AES-GCM encrypts.
  func EncryptCredentials(key []byte, creds *Credentials) ([]byte, error) {
      b, err := json.Marshal(creds)
      if err != nil { return nil, err }
      return crypto.Encrypt(key, b)
  }

  func DecryptCredentials(key, ciphertext []byte) (*Credentials, error) {
      plain, err := crypto.Decrypt(key, ciphertext)
      if err != nil { return nil, err }
      var creds Credentials
      if err := json.Unmarshal(plain, &creds); err != nil { return nil, err }
      return &creds, nil
  }
  ```

### 6. Fshare Driver

- [ ] `backend/internal/cloudstorage/drivers/fshare/driver.go`:
  ```go
  package fshare

  import (
      "context"
      "time"

      "github.com/thawng/velox/internal/cloudstorage"
      pkgfshare "github.com/thawng/velox/pkg/fshare"
  )

  type Config struct {
      AppKey    string // optional — falls back to pkgfshare.DefaultAppKey
      UserAgent string // optional
  }

  type Driver struct {
      cfg Config
  }

  func NewDriver(cfg Config) *Driver {
      return &Driver{cfg: cfg}
  }

  // --- Driver interface ---

  func (d *Driver) Type() string        { return "fshare" }
  func (d *Driver) DisplayName() string { return "Fshare" }
  func (d *Driver) AuthFlow() cloudstorage.AuthFlow {
      return cloudstorage.AuthFlowPassword
  }

  func (d *Driver) NewProvider(creds *cloudstorage.Credentials) (cloudstorage.Provider, error) {
      client, err := d.newClient()
      if err != nil { return nil, err }
      client.RestoreSession(creds.AccessToken, creds.SessionID)
      client.SetCredentials(creds.Email, creds.Password)
      return &provider{client: client}, nil
  }

  // --- PasswordAuthDriver interface ---

  func (d *Driver) AuthenticatePassword(ctx context.Context, email, password string) (*cloudstorage.Credentials, error) {
      client, err := d.newClient()
      if err != nil { return nil, err }
      sess, err := client.Login(ctx, email, password)
      if err != nil {
          return nil, mapFshareError(err)
      }
      return &cloudstorage.Credentials{
          ProviderType: "fshare",
          Email:        email,
          Password:     password, // stored encrypted at rest
          AccessToken:  sess.Token,
          SessionID:    sess.SessionID,
          ExpiresAt:    time.Now().Add(25 * time.Minute),
      }, nil
  }

  func (d *Driver) newClient() (*pkgfshare.Client, error) {
      opts := pkgfshare.Options{
          AppKey:    d.cfg.AppKey,
          UserAgent: d.cfg.UserAgent,
      }
      return pkgfshare.NewClient(opts)
  }
  ```

### 7. Fshare Provider + URL Parsing

- [ ] `backend/internal/cloudstorage/drivers/fshare/url.go`:
  ```go
  package fshare

  import (
      "fmt"
      "net/url"
      "regexp"
      "strings"

      "github.com/thawng/velox/internal/cloudstorage"
  )

  // Supported URL shapes:
  //   https://www.fshare.vn/folder/<linkcode>
  //   https://www.fshare.vn/folder/<linkcode>/anything
  //   https://fshare.vn/folder/<linkcode>
  //   https://www.fshare.vn/file/<linkcode>
  //   fshare://<linkcode>  (internal scheme — not exposed to users)

  var (
      folderPathRe = regexp.MustCompile(`^/folder/([A-Za-z0-9]+)`)
      filePathRe   = regexp.MustCompile(`^/file/([A-Za-z0-9]+)`)
  )

  func parseFolderURL(rawURL string) (string, error) {
      u, err := url.Parse(strings.TrimSpace(rawURL))
      if err != nil { return "", fmt.Errorf("%w: %v", cloudstorage.ErrInvalidURL, err) }
      if !strings.Contains(u.Host, "fshare.vn") {
          return "", fmt.Errorf("%w: host %q", cloudstorage.ErrInvalidURL, u.Host)
      }
      m := folderPathRe.FindStringSubmatch(u.Path)
      if m == nil {
          return "", fmt.Errorf("%w: path %q does not match /folder/<code>", cloudstorage.ErrInvalidURL, u.Path)
      }
      return m[1], nil
  }

  func parseFileURL(rawURL string) (string, error) { /* symmetric with folder */ }
  ```

- [ ] `backend/internal/cloudstorage/drivers/fshare/provider.go`:
  ```go
  type provider struct {
      client *pkgfshare.Client
  }

  func (p *provider) Type() string { return "fshare" }

  func (p *provider) ParseFolderURL(url string) (string, error) { return parseFolderURL(url) }
  func (p *provider) ParseFileURL(url string) (string, error)   { return parseFileURL(url) }

  func (p *provider) ListFolder(ctx context.Context, folderID string) ([]cloudstorage.Item, error) {
      items, err := p.client.ListFolder(ctx, folderID)
      if err != nil { return nil, mapFshareError(err) }
      return convertItems(items), nil
  }

  func (p *provider) GetDownloadURL(ctx context.Context, fileID string) (string, error) {
      url, err := p.client.GetDirectLink(ctx, fileID)
      if err != nil { return "", mapFshareError(err) }
      return url, nil
  }

  func (p *provider) GetAccountInfo(ctx context.Context) (*cloudstorage.AccountInfo, error) {
      info, err := p.client.GetUserInfo(ctx)
      if err != nil { return nil, mapFshareError(err) }
      expires, _ := strconv.ParseInt(info.ExpireVIP, 10, 64)
      quota, _ := strconv.ParseInt(info.Webspace, 10, 64)
      return &cloudstorage.AccountInfo{
          Email:       info.Email,
          DisplayName: info.Name,
          AccountType: info.AccountType,
          QuotaBytes:  quota,
          ExpiresAt:   time.Unix(expires, 0),
      }, nil
  }

  func (p *provider) CheckHealth(ctx context.Context) error {
      return mapFshareError(p.client.CheckSession(ctx))
  }

  // Convert fshare items → cloudstorage.Item (uniform shape).
  func convertItems(src []pkgfshare.FolderItem) []cloudstorage.Item {
      out := make([]cloudstorage.Item, 0, len(src))
      for _, it := range src {
          canonicalURL := ""
          if it.IsFolder() {
              canonicalURL = "https://www.fshare.vn/folder/" + it.Linkcode
          } else {
              canonicalURL = "https://www.fshare.vn/file/" + it.Linkcode
          }
          out = append(out, cloudstorage.Item{
              ID:       it.Linkcode,
              Name:     it.Name,
              IsFolder: it.IsFolder(),
              Size:     it.SizeBytes(),
              Mimetype: it.Mimetype,
              Modified: it.ModifiedTime(),
              URL:      canonicalURL,
          })
      }
      return out
  }

  // mapFshareError translates pkg/fshare errors → cloudstorage errors.
  func mapFshareError(err error) error {
      if err == nil { return nil }
      switch {
      case errors.Is(err, pkgfshare.ErrSessionExpired):
          return cloudstorage.ErrSessionExpired
      case errors.Is(err, pkgfshare.ErrInvalidCredentials):
          return cloudstorage.ErrInvalidCredentials
      case errors.Is(err, pkgfshare.ErrRateLimit):
          return cloudstorage.ErrRateLimit
      case errors.Is(err, pkgfshare.ErrLinkDead), errors.Is(err, pkgfshare.ErrNotLoggedIn):
          return cloudstorage.ErrFileNotFound
      case errors.Is(err, pkgfshare.ErrNotVIP):
          return cloudstorage.ErrAccountNotPremium
      default:
          return err
      }
  }
  ```

### 8. Startup Wiring

- [ ] `backend/cmd/server/server_app.go`:
  ```go
  if cfg.Cloud.Enabled {
      registry := cloudstorage.NewRegistry()

      fshareDriver := fshareDriver.NewDriver(fshareDriver.Config{
          AppKey:    cfg.Cloud.FshareAppKey,
          UserAgent: cfg.Cloud.FshareUserAgent,
      })
      if err := registry.Register(fshareDriver); err != nil {
          log.Fatalf("register fshare driver: %v", err)
      }

      // Future: registry.Register(googledriveDriver.NewDriver(cfg.Cloud.GDrive))

      app.CloudRegistry = registry
  }
  ```

## Acceptance Criteria

- [ ] `cd backend && go build ./internal/cloudstorage/...` clean
- [ ] Unit tests pass (race detector):
  - [ ] `Registry` — register + lookup + duplicate detection
  - [ ] `EncryptCredentials` + `DecryptCredentials` round-trip
  - [ ] Fshare driver `AuthenticatePassword` (mock `pkg/fshare` client via interface)
  - [ ] Fshare URL parser — folder/file shapes, rejections, edge cases (trailing slash, query params)
  - [ ] `mapFshareError` — all pkg/fshare error types map correctly
  - [ ] `convertItems` — folder vs file detection, URL canonicalization
- [ ] Integration test with real `pkg/fshare.Client` (mocked HTTP server) — Provider interface end-to-end
- [ ] Smoke test (manual, dev env): AuthenticatePassword → NewProvider → ListFolder → GetDownloadURL — all work against real fshare VIP account
- [ ] Interface contract documented — adding new driver requires only: implement Provider + Driver, register at startup
- [ ] No regression: startup OK khi `VELOX_CLOUD_SOURCES_ENABLED=false`

## Design Notes

- **Credentials.Password stored encrypted**: fshare has no refresh_token grant; auto-refresh scheduler needs password. GDrive/OneDrive use refresh_token (password field empty).
- **Session restore pattern**: Driver.NewProvider takes decrypted Credentials, hydrates internal client. No per-call login.
- **Error translation boundary**: pkg/fshare errors stay inside driver; cloudstorage errors cross the boundary upward. Scanner/handler don't know fshare-specific errors.
- **URL parsing is driver responsibility**: Each driver owns its URL format. Scanner gives raw URL → driver parses → native ID.
- **Provider instances are disposable**: Created fresh per request (no pooling). Cookie jar + token in fresh client per request. Thread-safety via driver's NewProvider factory.
- **Registry is read-only after startup**: Drivers register once in main(); runtime only reads.

## Gotchas

- `pkg/fshare.Client` has internal state (cookie jar + session). Creating fresh Client per call isolates sessions. If pooling becomes necessary → cache by credentials.ID.
- URL parsing edge cases: trailing slash, query params (`?t=123`), fragments (`#section`), percent-encoded names. Tests cover all.
- `Credentials.Meta` map: use for extras that don't fit (e.g., fshare `account_type` cached). Keep thin — main fields first-class.
- Converting `pkg/fshare.FolderItem` → `cloudstorage.Item` loses provider-specific fields (PID, path). If scanner needs those → add to Item.Meta later.
- Fshare VIP expiry (`ExpireVIP`) is string unix ts. Convert to time.Time in `GetAccountInfo`.

## Out of Scope

- Google Drive driver (architecture supports, implementation phase sau)
- OneDrive driver (same)
- OAuth callback HTTP handler (needed when adding OAuth provider)
- Provider instance pooling / LRU cache
- Cross-provider file copy / sync
