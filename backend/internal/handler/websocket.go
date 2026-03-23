package handler

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/thawng/velox/internal/auth"
	ws "github.com/thawng/velox/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins - CORS is handled by middleware
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub        *ws.Hub
	jwtManager *auth.JWTManager
	logger     *slog.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub, jwtManager *auth.JWTManager, logger *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:        hub,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Handle upgrades HTTP connection to WebSocket
// GET /api/ws?token={access_token}
func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		h.logger.Warn("websocket connection attempt without token")
		respondError(w, http.StatusUnauthorized, "missing token")
		return
	}

	// Validate token
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		h.logger.Warn("websocket invalid token", "error", err)
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Create and register client
	client := ws.NewClient(h.hub, conn, claims.UserID, claims.IsAdmin)
	go client.WritePump()
	go client.ReadPump()

	h.hub.Register() <- client

	h.logger.Info("websocket client connected", "user_id", claims.UserID)
}

// Register returns the register channel for testing purposes
func (h *WebSocketHandler) Register() chan<- *ws.Client {
	return h.hub.Register()
}
