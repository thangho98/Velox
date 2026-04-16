# Phase 06: Admin UI — Add Fshare Library (Webapp)
Status: ⬜ Pending
Dependencies: Phase 05

## Objective

Extend webapp admin UI cho phép add library với `source_type = 'fshare'`. Flow: login fshare → test connection → browse folder tree → pick root → save library → trigger scan.

## Context

- Existing admin page: [webapp/src/pages/admin/LibrariesPage.tsx](webapp/src/pages/admin/LibrariesPage.tsx)
- Create library dialog hiện tại chỉ có: name + path (filesystem)
- Need: source type picker → conditional form (local: path input; fshare: email/pass + folder picker)
- API client pattern: [webapp/src/api/](webapp/src/api/)

## Backend API Additions (needed first)

- [ ] `POST /api/cloud/fshare/test-login` — body `{email, password}` → return `{session_id}` (test only, DON'T persist)
- [ ] `POST /api/cloud/fshare/sessions` — body `{email, password}` → login + persist session, return `{session_id}`
- [ ] `GET /api/cloud/fshare/sessions/{id}/folders?parent_id={code}` — list folders tại fshare path
- [ ] `DELETE /api/cloud/fshare/sessions/{id}` — logout + remove session
- [ ] New handler: `backend/internal/handler/cloud_auth.go`

## Implementation Steps

### 1. Backend Endpoints
- [ ] Tạo [backend/internal/handler/cloud_auth.go](backend/internal/handler/cloud_auth.go):
  ```go
  type CloudAuthHandler struct {
      factory   *source.ResolverFactory
      sessRepo  *repository.CloudSessionRepo
      fshareCli *fshare.Client
      cryptoKey []byte
  }

  func (h *CloudAuthHandler) TestLogin(w, r) { /* login, don't persist */ }
  func (h *CloudAuthHandler) CreateSession(w, r) { /* login + encrypt + save */ }
  func (h *CloudAuthHandler) ListFolders(w, r) { /* use session, call fshare ListFolder */ }
  func (h *CloudAuthHandler) DeleteSession(w, r) { /* delete row */ }
  ```
- [ ] Register routes trong `cmd/server/server_app.go`:
  - `POST /api/admin/cloud/fshare/test-login` (admin-only)
  - `POST /api/admin/cloud/fshare/sessions`
  - `GET /api/admin/cloud/fshare/sessions/{id}/folders`
  - `DELETE /api/admin/cloud/fshare/sessions/{id}`
- [ ] All endpoints: admin-only middleware guard

### 2. Webapp API Client
- [ ] Tạo [webapp/src/api/cloud.ts](webapp/src/api/cloud.ts):
  ```ts
  export const cloudAPI = {
    fshare: {
      testLogin: (email: string, password: string) =>
        fetchWithAuth('/api/admin/cloud/fshare/test-login', {
          method: 'POST',
          body: JSON.stringify({ email, password }),
        }),
      createSession: (email, password) => /* ... */,
      listFolders: (sessionId: number, parentId?: string) => /* ... */,
      deleteSession: (sessionId) => /* ... */,
    },
  };
  ```

### 3. Source Type Picker in Create Library Dialog
- [ ] Modify [webapp/src/pages/admin/LibrariesPage.tsx](webapp/src/pages/admin/LibrariesPage.tsx):
  ```tsx
  const [sourceType, setSourceType] = useState<'local' | 'fshare'>('local');

  <select value={sourceType} onChange={...}>
    <option value="local">📁 Local Filesystem</option>
    <option value="fshare">☁️ Fshare Cloud</option>
  </select>

  {sourceType === 'local' && <LocalLibraryForm />}
  {sourceType === 'fshare' && <FshareLibraryForm />}
  ```

### 4. Fshare Login Form Component
- [ ] Tạo `webapp/src/components/admin/FshareLoginForm.tsx`:
  ```tsx
  export function FshareLoginForm({ onSessionCreated }) {
      const [email, setEmail] = useState('');
      const [password, setPassword] = useState('');
      const [testing, setTesting] = useState(false);

      const handleTest = async () => {
          setTesting(true);
          try {
              await cloudAPI.fshare.testLogin(email, password);
              toast.success('Login OK!');
          } catch (e) {
              toast.error(e.message);
          } finally { setTesting(false); }
      };

      const handleSave = async () => {
          const { session_id } = await cloudAPI.fshare.createSession(email, password);
          onSessionCreated(session_id);
      };

      return ( /* form UI */ );
  }
  ```

### 5. Fshare Folder Picker Component
- [ ] Tạo `webapp/src/components/admin/FshareFolderPicker.tsx`:
  ```tsx
  // Tree browser: root → click folder → expand → pick
  export function FshareFolderPicker({ sessionId, onSelect }) {
      const [tree, setTree] = useState<TreeNode[]>([]);
      const [expanded, setExpanded] = useState<Set<string>>(new Set());

      const loadFolder = async (parentId: string | null) => {
          const folders = await cloudAPI.fshare.listFolders(sessionId, parentId);
          // update tree state
      };

      useEffect(() => { loadFolder(null); }, [sessionId]);

      return (
          <div className="fshare-tree">
              {tree.map(node => <TreeNode node={node} onSelect={onSelect} />)}
          </div>
      );
  }
  ```
- [ ] Breadcrumb navigation: show current path
- [ ] Pick button: confirm selection
- [ ] Loading states + error handling

### 6. FshareLibraryForm — Orchestration
- [ ] Tạo `webapp/src/components/admin/FshareLibraryForm.tsx`:
  ```tsx
  export function FshareLibraryForm({ onSave }) {
      const [step, setStep] = useState<'login' | 'pick' | 'confirm'>('login');
      const [sessionId, setSessionId] = useState<number | null>(null);
      const [rootFolder, setRootFolder] = useState<{id: string, name: string} | null>(null);

      return (
          <div>
              {step === 'login' && <FshareLoginForm onSessionCreated={(id) => { setSessionId(id); setStep('pick'); }} />}
              {step === 'pick' && <FshareFolderPicker sessionId={sessionId} onSelect={(f) => { setRootFolder(f); setStep('confirm'); }} />}
              {step === 'confirm' && (
                  <div>
                      <p>Library root: {rootFolder.name}</p>
                      <button onClick={() => onSave({ source_type: 'fshare', source_credentials_id: sessionId, source_root_id: rootFolder.id })}>
                          Create Library
                      </button>
                  </div>
              )}
          </div>
      );
  }
  ```

### 7. Library List UI — Badge + Session Health
- [ ] Modify library list trong LibrariesPage:
  ```tsx
  {library.source_type === 'fshare' && <span className="badge">☁️ Fshare</span>}
  {library.source_type === 'local' && <span className="badge">📁 Local</span>}
  ```
- [ ] Add session health indicator: green (OK) / yellow (expires soon) / red (expired)
- [ ] "Re-authenticate" button khi session expired → re-open FshareLoginForm (reuse email)

## MediaCard Badge (khác file)
- [ ] Modify [webapp/src/components/MediaCard.tsx](webapp/src/components/MediaCard.tsx):
  - [ ] Add prop `sourceType?: string`
  - [ ] Show ☁️ icon overlay nếu `sourceType === 'fshare'`
  - [ ] (Webapp preview only — không stream thẳng được)

## Acceptance Criteria

- [ ] User flow end-to-end trên dev env:
  1. Admin mở LibrariesPage → click "Add Library"
  2. Chọn "☁️ Fshare Cloud"
  3. Nhập email + password → "Test Connection" → toast success
  4. Folder picker load → pick folder
  5. Confirm → Library created với `source_type='fshare'`
  6. Scan trigger → websocket progress hiển thị
- [ ] Unit tests `FshareLoginForm`, `FshareFolderPicker` với mock API
- [ ] Admin-only guard verified: non-admin user gọi endpoint → 403
- [ ] Error states handled: wrong password → error message hiển thị; network error → retry button
- [ ] Zero regression trên "Add Local Library" flow

## Gotchas

- **Credentials transit**: password trong POST body → HTTPS required (Velox chạy http trên LAN → note trong docs)
- **Session ID reuse**: Nếu user đã có session fshare cho email này → reuse row (upsert pattern). UNIQUE constraint catches này.
- **Folder picker pagination**: Folders có thể >100 items → paginate trong tree view
- **Password trong form state**: clear sau submit (security)
- **Test Connection vs Create Session**: Test Connection KHÔNG persist (dev-friendly retry). Create Session persists.

## Out of Scope

- Edit fshare library (change credentials) — phase sau (delete + recreate OK cho MVP)
- Multiple fshare accounts per Velox user — schema supports, UI chưa build picker
- Cookie-based login fallback (captcha workaround) — Phase 08 nếu cần
- Non-admin user add own fshare library (current: admin-only libraries)
