# Phase 06: Admin UI — Storage Providers + Libraries (2 Screens)
Status: ⬜ Pending
Dependencies: Phase 05

## Objective

Build 2 admin screens:
1. **Storage Providers** (`Settings → Storage Providers`) — manage cloud accounts
2. **Libraries** (existing page, enhanced) — pick source type (Local / Cloud) + paste folder URL

Driver-agnostic forms: the same UI works for fshare MVP and for future GDrive/OneDrive additions without code changes (Registry drives the dropdown).

## Context

- Existing: [webapp/src/pages/admin/LibrariesPage.tsx](webapp/src/pages/admin/LibrariesPage.tsx) — local-only library create
- New: dedicated Storage Providers page
- API: backend endpoints from Phase 02 + 03 (StorageProviderRepo + cloudstorage registry)

## Backend API Endpoints (needed)

All admin-only. Add to [backend/internal/handler/storage_provider.go](backend/internal/handler/storage_provider.go):

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/admin/cloud/drivers` | GET | List registered drivers (type, display_name, auth_flow) |
| `/api/admin/cloud/providers` | GET | List user's providers with health status |
| `/api/admin/cloud/providers` | POST | Create provider — password body OR OAuth code |
| `/api/admin/cloud/providers/{id}` | GET | Single provider detail |
| `/api/admin/cloud/providers/{id}` | PATCH | Update display_name, credentials |
| `/api/admin/cloud/providers/{id}` | DELETE | Remove provider (cascade: libraries.storage_provider_id → NULL) |
| `/api/admin/cloud/providers/{id}/refresh` | POST | Force re-auth + update health |
| `/api/admin/cloud/providers/{id}/validate-url` | POST | body `{url}` → `{folder_id, item_count}` — for Add Library form |
| `/api/admin/cloud/oauth/start` | GET | (future OAuth drivers) return AuthURL |
| `/api/admin/cloud/oauth/callback` | GET | (future OAuth drivers) receive code → exchange |

## Implementation Steps

### 1. Backend — Provider CRUD + Validation Handler

- [ ] Tạo `backend/internal/handler/storage_provider.go`:
  ```go
  type StorageProviderHandler struct {
      repo      *repository.StorageProviderRepo
      registry  *cloudstorage.Registry
      cryptoKey []byte
      libRepo   *repository.LibraryRepo
  }

  // ListDrivers: GET /api/admin/cloud/drivers
  // Returns registered drivers for UI dropdown population.
  func (h *StorageProviderHandler) ListDrivers(w, r) {
      writeJSON(w, h.registry.List())
  }

  // CreateProvider: POST /api/admin/cloud/providers
  // Body: { provider_type, display_name, auth: { flow: "password"|"oauth", ... } }
  func (h *StorageProviderHandler) CreateProvider(w, r) {
      // 1. Lookup driver by provider_type
      // 2. Based on auth_flow:
      //    - password: call driver.AuthenticatePassword(email, password)
      //    - oauth: exchange code
      // 3. Call provider.GetAccountInfo to populate display + VIP status
      // 4. Validate account_type (fshare: must be Vip; else return 400)
      // 5. Encrypt credentials → save row
      // 6. Return StorageProvider (without credentials_encrypted)
  }

  // ValidateURL: POST /api/admin/cloud/providers/{id}/validate-url
  // Body: { url }. Returns folder_id + item count for UX feedback.
  func (h *StorageProviderHandler) ValidateURL(w, r) {
      // 1. Load provider + decrypt credentials
      // 2. Driver.NewProvider(creds)
      // 3. Provider.ParseFolderURL(url) → folder_id
      // 4. Provider.ListFolder(ctx, folder_id) → items
      // 5. Return { folder_id, item_count, sample_names: [first 3] }
  }
  ```
- [ ] Register routes in `cmd/server/server_app.go`
- [ ] Admin-only middleware on all routes

### 2. Webapp — API Client

- [ ] `webapp/src/api/cloudstorage.ts`:
  ```ts
  export interface DriverMeta {
    type: string;
    display_name: string;
    auth_flow: "password" | "oauth";
  }

  export interface StorageProvider {
    id: number;
    provider_type: string;
    display_name: string;
    account_email: string;
    account_type?: string;
    quota_bytes?: number;
    used_bytes?: number;
    account_expires_at?: string;
    token_expires_at?: string;
    last_refresh_at?: string;
    last_error?: string;
    health: "healthy" | "expiring_soon" | "expired" | "error" | "unknown";
  }

  export const cloudAPI = {
    listDrivers: () => fetchWithAuth<DriverMeta[]>("/api/admin/cloud/drivers"),
    listProviders: () => fetchWithAuth<StorageProvider[]>("/api/admin/cloud/providers"),
    createProviderPassword: (providerType: string, displayName: string, email: string, password: string) =>
      fetchWithAuth<StorageProvider>("/api/admin/cloud/providers", {
        method: "POST",
        body: JSON.stringify({ provider_type: providerType, display_name: displayName, auth: { flow: "password", email, password } }),
      }),
    refreshProvider: (id: number) =>
      fetchWithAuth<StorageProvider>(`/api/admin/cloud/providers/${id}/refresh`, { method: "POST" }),
    deleteProvider: (id: number) =>
      fetchWithAuth<void>(`/api/admin/cloud/providers/${id}`, { method: "DELETE" }),
    validateURL: (providerID: number, url: string) =>
      fetchWithAuth<{ folder_id: string; item_count: number; sample_names: string[] }>(
        `/api/admin/cloud/providers/${providerID}/validate-url`,
        { method: "POST", body: JSON.stringify({ url }) }
      ),
  };
  ```

### 3. Storage Providers Page — List + Actions

- [ ] `webapp/src/pages/admin/StorageProvidersPage.tsx`:
  ```tsx
  export default function StorageProvidersPage() {
      const { data: providers } = useQuery(["providers"], cloudAPI.listProviders);
      const [addOpen, setAddOpen] = useState(false);

      return (
          <AdminLayout>
              <PageHeader title="Storage Providers" action={<Button onClick={() => setAddOpen(true)}>+ Add Provider</Button>} />
              {providers?.length === 0 ? <EmptyState /> : (
                  <div className="space-y-3">
                      {providers.map(p => <ProviderCard key={p.id} provider={p} />)}
                  </div>
              )}
              {addOpen && <AddProviderDialog onClose={() => setAddOpen(false)} />}
          </AdminLayout>
      );
  }
  ```

- [ ] `ProviderCard` component:
  ```tsx
  function ProviderCard({ provider }) {
      return (
          <Card>
              <div className="flex items-center gap-3">
                  <ProviderIcon type={provider.provider_type} />
                  <div>
                      <div className="font-medium">{provider.display_name}</div>
                      <div className="text-sm text-gray-500">
                          {provider.account_email} · {provider.account_type ?? "—"}
                      </div>
                  </div>
                  <ProviderHealthBadge health={provider.health} />
                  <div className="ml-auto flex gap-2">
                      <Button variant="ghost" onClick={() => refresh(provider.id)}>Refresh</Button>
                      <Button variant="ghost" onClick={() => confirmDelete(provider.id)}>Delete</Button>
                  </div>
              </div>
              {provider.last_error && <ErrorBanner>{provider.last_error}</ErrorBanner>}
          </Card>
      );
  }
  ```

### 4. Add Provider Dialog — Driver-Agnostic Form

- [ ] `webapp/src/components/admin/AddProviderDialog.tsx`:
  ```tsx
  export function AddProviderDialog({ onClose }) {
      const { data: drivers } = useQuery(["drivers"], cloudAPI.listDrivers);
      const [selectedType, setSelectedType] = useState<string>("");
      const driver = drivers?.find(d => d.type === selectedType);

      return (
          <Dialog>
              <DialogHeader>Add Storage Provider</DialogHeader>
              <Select value={selectedType} onChange={setSelectedType}>
                  <option value="">Choose provider…</option>
                  {drivers?.map(d => <option key={d.type} value={d.type}>{d.display_name}</option>)}
              </Select>

              {driver?.auth_flow === "password" && <PasswordAuthForm driver={driver} onSuccess={onClose} />}
              {driver?.auth_flow === "oauth" && <OAuthFlowButton driver={driver} onSuccess={onClose} />}
          </Dialog>
      );
  }
  ```

- [ ] `PasswordAuthForm` (for fshare):
  ```tsx
  function PasswordAuthForm({ driver, onSuccess }) {
      const [displayName, setDisplayName] = useState(`My ${driver.display_name}`);
      const [email, setEmail] = useState("");
      const [password, setPassword] = useState("");
      const [testing, setTesting] = useState(false);
      const [error, setError] = useState<string | null>(null);

      const handleSave = async () => {
          setTesting(true); setError(null);
          try {
              await cloudAPI.createProviderPassword(driver.type, displayName, email, password);
              onSuccess();
          } catch (e: any) {
              setError(e.message);
          } finally { setTesting(false); }
      };

      return (
          <form>
              <Input label="Display Name" value={displayName} onChange={setDisplayName} />
              <Input label="Email" value={email} onChange={setEmail} type="email" />
              <Input label="Password" value={password} onChange={setPassword} type="password" />
              {driver.type === "fshare" && (
                  <Alert type="info">VIP account required. Free accounts cannot stream.</Alert>
              )}
              {error && <Alert type="error">{error}</Alert>}
              <Button onClick={handleSave} loading={testing}>Test & Save</Button>
          </form>
      );
  }
  ```

- [ ] `OAuthFlowButton` (stub for future drivers):
  ```tsx
  function OAuthFlowButton({ driver, onSuccess }) {
      // Opens popup to driver.AuthURL, listens for callback via postMessage
      // Currently no OAuth driver registered — placeholder
      return <div>OAuth flow not yet implemented.</div>;
  }
  ```

### 5. Provider Health Badge (reusable)

- [ ] `webapp/src/components/admin/ProviderHealthBadge.tsx`:
  ```tsx
  export function ProviderHealthBadge({ health }: { health: StorageProvider["health"] }) {
      const map = {
          healthy:        { icon: "✅", label: "Healthy",        color: "green" },
          expiring_soon:  { icon: "⚠️", label: "Expiring soon",  color: "amber" },
          expired:        { icon: "❌", label: "Expired",        color: "red"   },
          error:          { icon: "❌", label: "Error",          color: "red"   },
          unknown:        { icon: "❓", label: "Unknown",        color: "gray"  },
      };
      const m = map[health];
      return <Badge color={m.color}>{m.icon} {m.label}</Badge>;
  }
  ```

### 6. Library Form — Source Picker + URL Input

- [ ] Modify [webapp/src/pages/admin/LibrariesPage.tsx](webapp/src/pages/admin/LibrariesPage.tsx) Add Library dialog:
  ```tsx
  <Dialog>
      <DialogHeader>Add Library</DialogHeader>
      <Input label="Name" value={name} onChange={setName} />

      <RadioGroup value={sourceType} onChange={setSourceType}>
          <Radio value="local">📁 Local Filesystem</Radio>
          <Radio value="cloud" disabled={!providers?.length}>
              ☁️ Cloud Storage
              {!providers?.length && <span className="text-sm text-gray-500 ml-2">(add a provider first)</span>}
          </Radio>
      </RadioGroup>

      {sourceType === "local" && <Input label="Path" value={path} onChange={setPath} placeholder="/mnt/data/movies" />}

      {sourceType === "cloud" && (
          <>
              <StorageProviderPicker value={providerId} onChange={setProviderId} providers={providers} />
              <Input
                  label="Folder URL"
                  value={sourceURL}
                  onChange={setSourceURL}
                  placeholder="https://www.fshare.vn/folder/XZWCPAZV3J71"
              />
              <Button variant="ghost" onClick={handleValidate}>Validate URL</Button>
              {validation && (
                  <Alert type={validation.ok ? "success" : "error"}>
                      {validation.ok
                          ? `✅ Folder OK — ${validation.item_count} items found`
                          : `❌ ${validation.error}`}
                  </Alert>
              )}
          </>
      )}

      <Button onClick={handleCreate} disabled={!canSubmit}>Create Library</Button>
  </Dialog>
  ```

- [ ] `StorageProviderPicker`:
  ```tsx
  function StorageProviderPicker({ value, onChange, providers }) {
      return (
          <Select value={value} onChange={onChange}>
              <option value="">Choose provider…</option>
              {providers.map(p => (
                  <option key={p.id} value={p.id} disabled={p.health === "expired" || p.health === "error"}>
                      {p.display_name} · {p.account_email} {p.health !== "healthy" && `(${p.health})`}
                  </option>
              ))}
          </Select>
      );
  }
  ```

### 7. Library List — Source Badge

- [ ] Modify library row rendering:
  ```tsx
  {library.storage_provider_id ? (
      <Badge>☁️ {providerByID(library.storage_provider_id)?.display_name}</Badge>
  ) : (
      <Badge>📁 Local</Badge>
  )}
  ```

### 8. Media Card Badge (Library Browse)

- [ ] Modify [webapp/src/components/MediaCard.tsx](webapp/src/components/MediaCard.tsx):
  - [ ] Add prop `cloudBadge?: boolean`
  - [ ] Show ☁️ icon overlay khi `cloudBadge` true (webapp preview only — direct play Android)

### 9. Sidebar Navigation

- [ ] Add Settings sub-item: "Storage Providers" → route `/admin/settings/storage`
- [ ] Icon: ☁️ or Cloud from LucideIcons

## Acceptance Criteria

- [ ] End-to-end user flow (dev env):
  1. Admin → Settings → Storage Providers → + Add Provider
  2. Select "Fshare" → form shows email + password + display name
  3. Submit → backend calls `driver.AuthenticatePassword` → validates VIP → saves provider
  4. Provider appears in list with ✅ Healthy + "Vip" account type
  5. Admin → Libraries → + Add Library
  6. Select "☁️ Cloud Storage" → picks provider → pastes fshare folder URL
  7. Clicks Validate → "✅ Folder OK — N items found"
  8. Submits → library created with storage_provider_id + source_url
  9. Scan triggers (Phase 04) → websocket progress
- [ ] Unit tests: `ProviderHealthBadge`, `StorageProviderPicker`, `PasswordAuthForm` với mock API
- [ ] Admin-only guard verified
- [ ] Error states: wrong password → inline error; network down → retry; non-VIP account → clear message
- [ ] Non-destructive delete: deleting provider with linked libraries → confirmation dialog listing affected libraries
- [ ] Zero regression: "Add Local Library" flow works unchanged

## UX Details

- **Non-VIP warning**: khi GetAccountInfo returns `account_type != "Vip"` cho fshare → backend returns 400 với clear message; UI shows "VIP account required. [Upgrade link]"
- **Expiry countdown**: Provider card shows "VIP expires in 60 days" nếu `account_expires_at` < 90 days away
- **Validate URL UX**: debounce 500ms, show spinner, "Validating…" → success/error inline below input
- **Health sort**: list sorts by health (errors first, then expiring, then healthy alphabetical)
- **Refresh action**: spinning icon during refresh; toast on success/failure

## Gotchas

- **Password field state**: clear after submit (security); don't persist in form state
- **Credentials transit**: password sent over HTTPS. Velox LAN usually http → document warning in env setup guide
- **Same-email same-type UPSERT**: DB UNIQUE constraint returns error → UI shows "Provider already exists. Refresh or delete to re-add."
- **Cascade delete**: when deleting provider, libraries.storage_provider_id → NULL (via ON DELETE SET NULL). But the library is now broken (source_url points nowhere). UX: warn before delete listing affected libraries + auto-disable scan for orphan libraries.
- **Dropdown driver list from backend**: fetched from Registry — won't show drivers not registered at startup. Avoids client-server mismatch.
- **OAuth flow (future)**: will need popup-based callback handler. Stub placeholder now to avoid scope creep.

## Out of Scope

- Edit credentials in-place (current: delete + re-create). UX enhancement phase sau.
- Multiple fshare accounts per user — UI supports (UNIQUE per-email), no extra work.
- Per-library credentials override (always uses linked provider).
- Provider usage statistics (scan bandwidth, most-played, etc.) — phase sau.
- OAuth redirect URI configuration via admin UI (requires restart for now).
