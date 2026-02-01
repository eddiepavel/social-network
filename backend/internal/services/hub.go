package services

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	UserID   []byte
	UserIDHex string
	Send     chan []byte
	Rooms    map[string]bool // roomId -> subscribed
	mu       sync.Mutex
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string          `json:"type"`      // "message", "typing", "read", "subscribe", "unsubscribe"
	RoomID    string          `json:"room_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	SenderID  string          `json:"sender_id,omitempty"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// Hub manages all WebSocket connections
type Hub struct {
	// Registered clients by user ID (hex string)
	clients    map[string]*Client
	clientsMu  sync.RWMutex

	// Room subscriptions: roomId -> map of userIdHex -> client
	rooms      map[string]map[string]*Client
	roomsMu    sync.RWMutex

	// Channels
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage

	logger *slog.Logger
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Exclude string // userIdHex to exclude from broadcast
}

// NewHub creates a new Hub instance
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		logger:     logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clientsMu.Lock()
			h.clients[client.UserIDHex] = client
			h.clientsMu.Unlock()
			h.logger.Info("WebSocket client registered", "userID", client.UserIDHex)

		case client := <-h.unregister:
			h.clientsMu.Lock()
			if _, ok := h.clients[client.UserIDHex]; ok {
				delete(h.clients, client.UserIDHex)
				close(client.Send)
			}
			h.clientsMu.Unlock()

			// Remove from all rooms
			h.roomsMu.Lock()
			for roomID := range client.Rooms {
				if room, ok := h.rooms[roomID]; ok {
					delete(room, client.UserIDHex)
					if len(room) == 0 {
						delete(h.rooms, roomID)
					}
				}
			}
			h.roomsMu.Unlock()
			h.logger.Info("WebSocket client unregistered", "userID", client.UserIDHex)

		case msg := <-h.broadcast:
			h.roomsMu.RLock()
			if room, ok := h.rooms[msg.RoomID]; ok {
				for userIDHex, client := range room {
					if userIDHex != msg.Exclude {
						select {
						case client.Send <- msg.Message:
						default:
							// Client's buffer is full, skip
							h.logger.Warn("Client buffer full, dropping message", "userID", userIDHex)
						}
					}
				}
			}
			h.roomsMu.RUnlock()
		}
	}
}

// Subscribe adds a client to a room
func (h *Hub) Subscribe(client *Client, roomID string) {
	h.roomsMu.Lock()
	defer h.roomsMu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[string]*Client)
	}
	h.rooms[roomID][client.UserIDHex] = client

	client.mu.Lock()
	client.Rooms[roomID] = true
	client.mu.Unlock()

	h.logger.Debug("Client subscribed to room", "userID", client.UserIDHex, "roomID", roomID)
}

// Unsubscribe removes a client from a room
func (h *Hub) Unsubscribe(client *Client, roomID string) {
	h.roomsMu.Lock()
	defer h.roomsMu.Unlock()

	if room, ok := h.rooms[roomID]; ok {
		delete(room, client.UserIDHex)
		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}

	client.mu.Lock()
	delete(client.Rooms, roomID)
	client.mu.Unlock()

	h.logger.Debug("Client unsubscribed from room", "userID", client.UserIDHex, "roomID", roomID)
}

// Broadcast sends a message to all clients in a room
func (h *Hub) Broadcast(roomID string, message []byte, excludeUserIDHex string) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Exclude: excludeUserIDHex,
	}
}

// SendToUser sends a message directly to a specific user
func (h *Hub) SendToUser(userIDHex string, message []byte) bool {
	h.clientsMu.RLock()
	client, ok := h.clients[userIDHex]
	h.clientsMu.RUnlock()

	if !ok {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// GetClient returns a client by user ID hex
func (h *Hub) GetClient(userIDHex string) (*Client, bool) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	client, ok := h.clients[userIDHex]
	return client, ok
}

// IsUserOnline checks if a user is connected
func (h *Hub) IsUserOnline(userIDHex string) bool {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	_, ok := h.clients[userIDHex]
	return ok
}

// Register channel getter
func (h *Hub) Register() chan<- *Client {
	return h.register
}

// Unregister channel getter
func (h *Hub) Unregister() chan<- *Client {
	return h.unregister
}

// Constants for WebSocket
const (
	// Time allowed to write a message to the peer
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	PongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than PongWait)
	PingPeriod = (PongWait * 9) / 10

	// Maximum message size allowed from peer
	MaxMessageSize = 8192
)
