package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thawng/velox/internal/model"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = 30 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients: userID -> map of clients (multiple connections per user)
	clients map[int64]map[*Client]bool

	// Broadcast channel for notifications
	broadcast chan *model.Notification

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex

	// Logger
	logger *slog.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		broadcast:  make(chan *model.Notification, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Register returns the register channel
func (h *Hub) Register() chan<- *Client {
	return h.register
}

// Run starts the hub event loop
func (h *Hub) Run() {
	h.logger.Info("WebSocket hub started")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
			h.logger.Info("client registered", "user_id", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if userClients, ok := h.clients[client.userID]; ok {
				if _, ok := userClients[client]; ok {
					delete(userClients, client)
					close(client.send)
					if len(userClients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			h.logger.Info("client unregistered", "user_id", client.userID)

		case notification := <-h.broadcast:
			h.broadcastNotification(notification)
		}
	}
}

// broadcastNotification sends notification to target users
func (h *Hub) broadcastNotification(n *model.Notification) {
	// Always wrap in { notification: ... } so frontend can reliably detect notification messages
	payloadBytes, err := json.Marshal(map[string]any{"notification": n})
	if err != nil {
		h.logger.Error("failed to marshal notification payload", "error", err)
		return
	}
	message := &model.WebSocketMessage{
		Type:    "notification",
		Payload: payloadBytes,
	}

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("failed to marshal message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// If userID is specified, send only to that user (and broadcast if nil)
	if n.UserID != nil {
		if userClients, ok := h.clients[*n.UserID]; ok {
			for client := range userClients {
				select {
				case client.send <- data:
				default:
					// Client too slow — drop message, it will reconnect
				}
			}
		}
	} else {
		// Broadcast to all connected clients
		for _, userClients := range h.clients {
			for client := range userClients {
				select {
				case client.send <- data:
				default:
					// Client too slow — drop message, it will reconnect
				}
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userClients, ok := h.clients[userID]; ok {
		for client := range userClients {
			select {
			case client.send <- message:
			default:
				// Client too slow — drop message
			}
		}
	}
}

// Broadcast sends a notification to all connected clients
func (h *Hub) Broadcast(notification *model.Notification) {
	h.broadcast <- notification
}

// BroadcastToAdmins sends a typed message to all connected admin clients.
// Used for transient progress updates that don't need persistence.
func (h *Hub) BroadcastToAdmins(msgType string, payload any) {
	data, err := json.Marshal(&model.WebSocketMessage{
		Type:    msgType,
		Payload: mustMarshal(payload),
	})
	if err != nil {
		h.logger.Error("failed to marshal admin broadcast", "type", msgType, "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userClients := range h.clients {
		for client := range userClients {
			if !client.isAdmin {
				continue
			}
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// GetConnectedCount returns the number of connected clients
func (h *Hub) GetConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, userClients := range h.clients {
		count += len(userClients)
	}
	return count
}

// Client represents a single WebSocket connection
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	userID  int64
	isAdmin bool
	send    chan []byte
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID int64, isAdmin bool) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		userID:  userID,
		isAdmin: isAdmin,
		send:    make(chan []byte, 256),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("websocket read error", "error", err)
			}
			break
		}
		// We don't process incoming messages - client only receives notifications
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
