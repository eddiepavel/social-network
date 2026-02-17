package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"social-network/pkg/db/sqlite"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

type Manager struct {
	DB      *sql.DB
	Logger  *slog.Logger
	clients ClientsConnected
	sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	sentNotifs      map[string]bool
	sentNotifsMutex sync.RWMutex
	handlers        map[string]EventHandler
}

func NewManager(db *sql.DB, l *slog.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		DB:         db,
		clients:    make(ClientsConnected),
		Logger:     l,
		ctx:        ctx,
		cancel:     cancel,
		sentNotifs: make(map[string]bool),
		handlers:   make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// setupEventHandlers registers all event handlers
// Add your custom handlers here!
func (m *Manager) setupEventHandlers() {
	m.handlers[EventPrivateMessage] = PrivateMessageHandler
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventEnterChat] = EnterChatHandler
	m.handlers[EventLeaveChat] = LeaveChatHandler
}

// routeEvent routes incoming events to the appropriate handler
func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c, m.DB); err != nil {
			m.Logger.Warn("Handler error", "type", event.Type, "userID", c.userID, "err", err)
			return err
		}
		return nil
	}
	return fmt.Errorf("unknown event type: %s", event.Type)
}

func (m *Manager) Start() {
	go m.StartNotificationWorker()
}

func (m *Manager) Shutdown() {
	m.cancel()
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request, userID []byte) {
	// Convert userID bytes to UUID string
	uid, err := uuid.FromBytes(userID)
	if err != nil {
		m.Logger.Error("failed to convert user ID to UUID", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr := uid.String()

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.Logger.Error("failed to upgrade connection", "err", err)
		return
	}

	client := NewClient(conn, m, userIDStr)

	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	m.clients[client.userID] = client
	m.Unlock()

	m.Logger.Info("Client connected", "userID", client.userID, "total_clients", len(m.clients))
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client.userID]; ok {
		client.connection.Close()
		delete(m.clients, client.userID)
		m.Logger.Info("Client disconnected", "userID", client.userID, "total_clients", len(m.clients))
	}
}

// IsUserInRoom checks if a user is currently viewing a specific chat room
func (m *Manager) IsUserInRoom(userID string, roomID string) bool {
	m.RLock()
	defer m.RUnlock()

	if client, ok := m.clients[userID]; ok {
		return client.activeRoom == roomID
	}
	return false
}

func (m *Manager) StartNotificationWorker() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	m.Logger.Info("Notification worker started")

	for {
		select {
		case <-m.ctx.Done():
			m.Logger.Info("Notification worker shutting down")
			return
		case <-ticker.C:
			m.checkAndSendNotifications()
		}
	}
}

func (m *Manager) checkAndSendNotifications() {
	m.RLock()
	clientUserIDs := make([]string, 0, len(m.clients))
	for userID := range m.clients {
		clientUserIDs = append(clientUserIDs, userID)
	}
	m.RUnlock()

	for _, userIDStr := range clientUserIDs {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			m.Logger.Error("failed to parse user ID", "userID", userIDStr, "err", err)
			continue
		}

		userIDBytes, err := uid.MarshalBinary()
		if err != nil {
			m.Logger.Error("failed to convert user ID to bytes", "userID", userIDStr, "err", err)
			continue
		}

		// Use the new query that includes user details
		notifications, err := sqlite.NewQuery(m.DB).Notifications.GetUnseenNotificationsWithUserDetails(context.Background(), userIDBytes)
		if err != nil {
			m.Logger.Error("failed to get unseen notifications", "userID", userIDStr, "err", err)
			continue
		}

		for _, notif := range notifications {
			m.sentNotifsMutex.RLock()
			alreadySent := m.sentNotifs[notif.NotifID]
			m.sentNotifsMutex.RUnlock()

			if alreadySent {
				continue
			}

			fromID, err := uuid.FromBytes(notif.FromID)
			if err != nil {
				m.Logger.Error("failed to convert from_id", "notif_id", notif.NotifID, "err", err)
				continue
			}
			fromIDStr := fromID.String()

			groupIDStr := ""
			if len(notif.GroupID) > 0 {
				groupID, err := uuid.FromBytes(notif.GroupID)
				if err != nil {
					m.Logger.Error("failed to convert group_id", "notif_id", notif.NotifID, "err", err)
				} else {
					groupIDStr = groupID.String()
				}
			}

			eventIDStr := ""
			if len(notif.EventID) > 0 {
				eventID, err := uuid.FromBytes(notif.EventID)
				if err != nil {
					m.Logger.Error("failed to convert event_id", "notif_id", notif.NotifID, "err", err)
				} else {
					eventIDStr = eventID.String()
				}
			}

			// Build from_name from first_name and last_name
			fromName := notif.FromFirstName + " " + notif.FromLastName

			// Get optional fields
			fromAvatar := ""
			if notif.FromAvatar.Valid {
				fromAvatar = notif.FromAvatar.String
			}
			fromNickname := ""
			if notif.FromNickname.Valid {
				fromNickname = notif.FromNickname.String
			}

			notifEvent := NotificationEvent{
				NotifID:      notif.NotifID,
				ReceiverID:   userIDStr,
				Type:         notif.Type,
				FromID:       fromIDStr,
				FromName:     fromName,
				FromAvatar:   fromAvatar,
				FromNickname: fromNickname,
				GroupID:      groupIDStr,
				EventID:      eventIDStr,
				CreatedAt:    notif.CreatedAt.Time,
				IsSeen:       notif.IsSeen.Bool,
			}

			payload, err := json.Marshal(notifEvent)
			if err != nil {
				m.Logger.Error("failed to marshal notification", "notif_id", notif.NotifID, "err", err)
				continue
			}

			event := Event{
				Type:    EventNotification,
				Payload: payload,
			}

			m.RLock()
			client, ok := m.clients[userIDStr]
			m.RUnlock()

			if ok {
				select {
				case client.egress <- event:
					m.Logger.Info("Notification sent to client", "userID", userIDStr, "notif_id", notif.NotifID)
					m.sentNotifsMutex.Lock()
					m.sentNotifs[notif.NotifID] = true
					m.sentNotifsMutex.Unlock()
				default:
					m.Logger.Warn("Client egress channel full", "userID", userIDStr, "notif_id", notif.NotifID)
				}
			}
		}
	}

	m.cleanupSentNotifications()
}

func (m *Manager) cleanupSentNotifications() {
	m.sentNotifsMutex.Lock()
	defer m.sentNotifsMutex.Unlock()

	if len(m.sentNotifs) > 10000 {
		m.sentNotifs = make(map[string]bool)
		m.Logger.Info("Cleaned up sent notifications cache")
	}
}

// BroadcastChatMessage sends a chat message to all participants in a room
func (m *Manager) BroadcastChatMessage(roomID []byte, msg ChatMessageEvent) {
	participants, err := sqlite.NewQuery(m.DB).Chat.GetRoomParticipants(context.Background(), roomID)
	if err != nil {
		m.Logger.Error("failed to get room participants for broadcast", "err", err)
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		m.Logger.Error("failed to marshal chat message event", "err", err)
		return
	}

	event := Event{
		Type:    EventChatMessage,
		Payload: payload,
	}

	for _, participantIDBytes := range participants {
		uid, err := uuid.FromBytes(participantIDBytes)
		if err != nil {
			continue
		}
		userIDStr := uid.String()

		m.RLock()
		client, ok := m.clients[userIDStr]
		m.RUnlock()

		if ok {
			select {
			case client.egress <- event:
				m.Logger.Info("Chat message broadcast to participant", "userID", userIDStr, "room", msg.RoomID)
			default:
				m.Logger.Warn("Client egress channel full during broadcast", "userID", userIDStr)
			}
		}
	}
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Allow empty origin (for tools like Postman)
	if origin == "" {
		return true
	}

	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8000",
		"http://localhost:8080",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}
