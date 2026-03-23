# Plan O: WebSocket Notifications (Real-time Push)

Real-time notification system cho Velox. Push notifications khi:
- Scan library hoàn tất / thất bại
- Media mới được thêm (new movie/episode detected)
- Transcode hoàn tất / thất bại
- Subtitle search/download hoàn tất
- Metadata identify hoàn tất
- Library watcher phát hiện file mới

---

## Phase 01: Database & Models

### 1.1 Migration: notifications table
**File:** `backend/internal/database/migrate/022_notifications.go`

```go
Table: notifications
- id: INTEGER PK
- user_id: INTEGER FK → users(id) CASCADE (NULL = broadcast to all)
- type: TEXT NOT NULL ("scan_complete", "media_added", "transcode_complete", "subtitle_downloaded", etc.)
- title: TEXT NOT NULL (display title)
- message: TEXT (optional detail)
- data: TEXT (JSON payload, flexible)
- read: INTEGER DEFAULT 0 (boolean)
- created_at: DATETIME DEFAULT CURRENT_TIMESTAMP
- read_at: DATETIME NULL

Indexes:
- idx_notifications_user_read ON (user_id, read)
- idx_notifications_created ON (created_at DESC)
```

### 1.2 Model
**File:** `backend/internal/model/notification.go`

```go
type Notification struct {
    ID        int64           `json:"id"`
    UserID    *int64          `json:"user_id"` // nil = broadcast
    Type      string          `json:"type"`
    Title     string          `json:"title"`
    Message   string          `json:"message"`
    Data      json.RawMessage `json:"data"`
    Read      bool            `json:"read"`
    CreatedAt time.Time       `json:"created_at"`
    ReadAt    *time.Time      `json:"read_at"`
}

type NotificationData struct {
    // Common fields
    LibraryID   *int64 `json:"library_id,omitempty"`
    MediaID     *int64 `json:"media_id,omitempty"`
    SeriesID    *int64 `json:"series_id,omitempty"`
    EpisodeID   *int64 `json:"episode_id,omitempty"`

    // Scan-specific
    ScannedCount int    `json:"scanned_count,omitempty"`
    NewCount     int    `json:"new_count,omitempty"`

    // Transcode-specific
    Quality     string `json:"quality,omitempty"`
    Duration    int    `json:"duration_seconds,omitempty"`

    // Subtitle-specific
    Language    string `json:"language,omitempty"`
    Provider    string `json:"provider,omitempty"`
}
```

### 1.3 Repository
**File:** `backend/internal/repository/notification.go`

Methods:
- `Create(ctx, notification) → error`
- `GetByID(ctx, id) → (*Notification, error)`
- `GetByUser(ctx, userID, unreadOnly, limit, offset) → ([]Notification, error)`
- `MarkAsRead(ctx, userID, notificationIDs) → error`
- `MarkAllAsRead(ctx, userID) → error`
- `Delete(ctx, userID, notificationIDs) → error`
- `DeleteOld(ctx, before time.Time) → (int64, error)` // Cleanup job
- `CountUnread(ctx, userID) → (int64, error)`

**Deliverable:** Database layer sẵn sàng, test qua repository

---

## Phase 02: WebSocket Infrastructure

### 2.1 WebSocket Hub (In-memory)
**File:** `backend/internal/websocket/hub.go`

```go
// Hub quản lý tất cả connections
type Hub struct {
    clients    map[int64]*Client  // userID -> Client (authenticated)
    broadcast  chan Message       // Broadcast channel
    register   chan *Client       // New connections
    unregister chan *Client       // Disconnections
}

// Client = 1 WebSocket connection
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    userID   int64
    send     chan []byte
}

// Message để broadcast
type Message struct {
    Type      string          `json:"type"` // "notification", "ping"
    UserIDs   []int64         `json:"user_ids,omitempty"` // empty = all
    Payload   json.RawMessage `json:"payload"`
}
```

Features:
- Thread-safe (goroutines + channels)
- Authentication: JWT token từ query param `?token=...`
- Heartbeat: Ping/pong 30s
- Graceful shutdown: Close connections khi server stop

### 2.2 WebSocket Handler
**File:** `backend/internal/handler/websocket.go`

```go
// GET /api/ws — WebSocket endpoint
func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Upgrade HTTP → WebSocket
    // 2. Extract JWT from query param
    // 3. Validate token, get userID
    // 4. Register client với Hub
    // 5. Start read/write pumps
}
```

### 2.3 Dependencies
**File:** `backend/go.mod` (thêm nếu chưa có)

```
go get github.com/gorilla/websocket
```

**Deliverable:** WebSocket chạy được, test qua browser console

---

## Phase 03: Notification Service

### 3.1 Service Layer
**File:** `backend/internal/service/notification.go`

```go
type NotificationService struct {
    repo     *repository.NotificationRepo
    hub      *websocket.Hub
}

// CreateAndSend: lưu DB + gửi WebSocket real-time
func (s *NotificationService) CreateAndSend(ctx context.Context, notif *model.Notification) error {
    // 1. Save to DB
    // 2. Broadcast qua WebSocket
    // 3. Return
}

// Các helper methods cho từng loại notification:
func (s *NotificationService) NotifyScanComplete(userID *int64, libraryID int64, scanned, new int)
func (s *NotificationService) NotifyMediaAdded(userID *int64, mediaID int64, title string)
func (s *NotificationService) NotifyTranscodeComplete(userID int64, mediaID int64, success bool)
func (s *NotificationService) NotifySubtitleDownloaded(userID int64, mediaID int64, lang string)
```

### 3.2 Integration Points

**Scanner (`internal/scanner/scanner.go`):**
- Gọi `notificationService.NotifyScanComplete()` sau khi scan xong
- Gọi `notificationService.NotifyMediaAdded()` mỗi khi có media mới

**Transcoder (`internal/transcoder/transcoder.go`):**
- Gọi `notificationService.NotifyTranscodeComplete()` khi transcode xong/thất bại

**Subtitle Search (`internal/service/subtitle_search.go`):**
- Gọi `notificationService.NotifySubtitleDownloaded()` khi download xong

**Library Watcher (`internal/watcher/watcher.go`):**
- Gọi `notificationService.NotifyMediaAdded()` khi phát hiện file mới

**Deliverable:** Notifications được tạo và gửi real-time từ các service

---

## Phase 04: HTTP API Endpoints

**File:** `backend/internal/handler/notification.go`

```go
// GET /api/notifications
// Query: ?unread_only=true&limit=20&offset=0
// Response: { notifications: [...], unread_count: 5 }
func ListNotifications(w http.ResponseWriter, r *http.Request)

// PATCH /api/notifications/read
// Body: { "ids": [1, 2, 3] }
func MarkAsRead(w http.ResponseWriter, r *http.Request)

// PATCH /api/notifications/read-all
func MarkAllAsRead(w http.ResponseWriter, r *http.Request)

// DELETE /api/notifications
// Body: { "ids": [1, 2, 3] }
func DeleteNotifications(w http.ResponseWriter, r *http.Request)

// GET /api/notifications/unread-count
// Response: { count: 5 }
func GetUnreadCount(w http.ResponseWriter, r *http.Request)
```

**Deliverable:** REST API đầy đủ, test qua curl/Postman

---

## Phase 05: Frontend WebSocket Client

### 5.1 WebSocket Hook
**File:** `webapp/src/hooks/useWebSocket.ts`

```typescript
export function useWebSocket() {
    const [connected, setConnected] = useState(false)
    const [notifications, setNotifications] = useState<Notification[]>([])

    useEffect(() => {
        // 1. Connect đến /api/ws?token=...
        // 2. Xử lý onopen, onmessage, onclose, onerror
        // 3. Tự động reconnect với exponential backoff
        // 4. Heartbeat ping/pong
    }, [])

    return { connected, notifications }
}
```

### 5.2 Notification Store
**File:** `webapp/src/hooks/stores/useNotifications.ts`

```typescript
// Global state cho notifications
interface NotificationState {
    items: Notification[]
    unreadCount: number
    markAsRead: (ids: number[]) => Promise<void>
    markAllAsRead: () => Promise<void>
    deleteNotifications: (ids: number[]) => Promise<void>
    addNotification: (notif: Notification) => void // từ WebSocket
}
```

### 5.3 Notification Bell Component
**File:** `webapp/src/components/NotificationBell.tsx`

```typescript
// Bell icon với badge số unread
// Dropdown panel hiển thị notifications
// Click notification → navigate tới media/library
// Infinite scroll hoặc "Load more"
// Mark as read / Delete
```

**Deliverable:** UI có thể nhận và hiển thị notifications real-time

---

## Phase 06: UI/UX Polish & Cleanup

### 6.1 Toast Notifications
**File:** `webapp/src/components/ToastNotification.tsx`

- Khi nhận notification mới → hiển thị toast (auto-dismiss 5s)
- Click toast → navigate đến content tương ứng

### 6.2 Notification Preferences
**File:** `webapp/src/pages/SettingsPage.tsx` (tab Notifications)

```typescript
// User có thể toggle từng loại notification:
- Scan complete
- New media added
- Transcode complete
- Subtitle downloaded
- Browser notifications (Web Push API - optional)
```

### 6.3 Cleanup Job
**File:** `backend/internal/service/scheduler.go` (thêm job)

```go
// Daily cleanup: xóa notifications > 30 ngày và đã đọc
func cleanupOldNotifications() {
    // Delete where created_at < NOW() - 30 days AND read = 1
}
```

**Deliverable:** Feature hoàn chỉnh, user-friendly

---

## Implementation Order

```
Phase 01 → Phase 02 → Phase 03 → Phase 04 → Phase 05 → Phase 06
   ↓          ↓          ↓          ↓          ↓          ↓
  DB       WebSocket   Service     API      Frontend    Polish
```

Mỗi phase có thể test độc lập trước khi chuyển phase tiếp theo.

---

## Testing Strategy

### Manual Testing

1. **WebSocket Connection:**
   ```javascript
   // Browser console
   const ws = new WebSocket('ws://localhost:8080/api/ws?token=...')
   ws.onmessage = (e) => console.log('Received:', JSON.parse(e.data))
   ```

2. **Trigger Notifications:**
   - Scan library → check notification
   - Request transcode → check notification
   - Download subtitle → check notification

3. **UI Testing:**
   - Bell icon hiển thị số unread đúng
   - Click notification → navigate đúng
   - Mark as read → số unread giảm

### Edge Cases

- WebSocket disconnect → auto-reconnect
- Multiple tabs → mỗi tab 1 connection
- Server restart → client reconnect
- Anonymous user (no token) → không cho connect WS

---

## API Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/ws` | WS | WebSocket connection |
| `/api/notifications` | GET | List notifications |
| `/api/notifications/read` | PATCH | Mark as read |
| `/api/notifications/read-all` | PATCH | Mark all as read |
| `/api/notifications` | DELETE | Delete notifications |
| `/api/notifications/unread-count` | GET | Get unread count |

---

## Dependencies

**Backend:**
- `github.com/gorilla/websocket` (WebSocket library chuẩn cho Go)

**Frontend:**
- Native WebSocket API (không cần thư viện thêm)

---

## Notes

1. **Security:** WebSocket cũng cần JWT authentication. Token gửi qua query param (không thể gửi header khi upgrade WS).

2. **Scalability:** Hiện tại in-memory Hub (đủ cho single instance). Nếu scale multiple servers sau này → cần Redis Pub/Sub.

3. **Broadcast vs Targeted:**
   - Broadcast: scan complete (tất cả admin users)
   - Targeted: transcode complete (user request transcode)

4. **Persistence:** Notifications lưu DB để user có thể xem lại history, không chỉ real-time.
