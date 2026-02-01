package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/services"
	db_chat "social-network/pkg/db/queries/chat"
	"social-network/pkg/db/sqlite"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, you should check the origin
		return true
	},
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(app *app.App, hub *services.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionID, err := helpers.GenerateFromString(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := sqlite.NewQuery(app.DB).Sessions.ValidateSession(r.Context(), sessionID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			app.Logger.Error("Failed to upgrade WebSocket", "error", err)
			return
		}

		userIDHex, _ := helpers.GenerateFromBytes(session.UserID)

		client := &services.Client{
			Hub:       hub,
			Conn:      conn,
			UserID:    session.UserID,
			UserIDHex: userIDHex,
			Send:      make(chan []byte, 256),
			Rooms:     make(map[string]bool),
		}

		hub.Register() <- client

		// Start goroutines for reading and writing
		go writePump(client)
		go readPump(client, app)
	}
}

// readPump handles incoming messages from the WebSocket
func readPump(client *services.Client, app *app.App) {
	defer func() {
		client.Hub.Unregister() <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(services.MaxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(services.PongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(services.PongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.Logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		var wsMsg services.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			app.Logger.Error("Failed to unmarshal WebSocket message", "error", err)
			continue
		}

		handleMessage(client, app, &wsMsg)
	}
}

// writePump handles outgoing messages to the WebSocket
func writePump(client *services.Client) {
	ticker := time.NewTicker(services.PingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(services.WriteWait))
			if !ok {
				// Hub closed the channel
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(services.WriteWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func handleMessage(client *services.Client, app *app.App, msg *services.WSMessage) {
	switch msg.Type {
	case "subscribe":
		// Subscribe to a chat room
		if msg.RoomID != "" {
			// Verify user is a participant in this room
			roomID, err := helpers.GenerateFromString(msg.RoomID)
			if err != nil {
				return
			}

			isParticipant, err := sqlite.NewQuery(app.DB).Chat.CheckUserIsParticipant(nil, db_chat.CheckUserIsParticipantParams{
				UserID: client.UserID,
				RoomID: roomID,
			})
			if err != nil || isParticipant == 0 {
				return
			}

			client.Hub.Subscribe(client, msg.RoomID)
		}

	case "unsubscribe":
		if msg.RoomID != "" {
			client.Hub.Unsubscribe(client, msg.RoomID)
		}

	case "message":
		// Handle incoming chat message
		if msg.RoomID == "" || msg.Content == "" {
			return
		}

		roomID, err := helpers.GenerateFromString(msg.RoomID)
		if err != nil {
			return
		}

		// Verify user is a participant
		isParticipant, err := sqlite.NewQuery(app.DB).Chat.CheckUserIsParticipant(nil, db_chat.CheckUserIsParticipantParams{
			UserID: client.UserID,
			RoomID: roomID,
		})
		if err != nil || isParticipant == 0 {
			return
		}

		// Save message to database
		messageID, _ := uuid.New().MarshalBinary()
		err = sqlite.NewQuery(app.DB).Chat.CreateMessage(nil, db_chat.CreateMessageParams{
			MessageID: messageID,
			Content:   msg.Content,
			SenderID:  client.UserID,
			TargetID:  roomID,
		})
		if err != nil {
			app.Logger.Error("Failed to save chat message", "error", err)
			return
		}

		messageIDHex, _ := helpers.GenerateFromBytes(messageID)

		// Broadcast to room
		broadcastMsg := services.WSMessage{
			Type:      "message",
			RoomID:    msg.RoomID,
			Content:   msg.Content,
			SenderID:  client.UserIDHex,
			Timestamp: time.Now(),
		}

		// Add message ID to data
		data, _ := json.Marshal(map[string]string{"message_id": messageIDHex})
		broadcastMsg.Data = data

		msgBytes, _ := json.Marshal(broadcastMsg)
		client.Hub.Broadcast(msg.RoomID, msgBytes, "") // Don't exclude sender, they should see their own message

	case "typing":
		// Broadcast typing indicator to room
		if msg.RoomID == "" {
			return
		}

		broadcastMsg := services.WSMessage{
			Type:     "typing",
			RoomID:   msg.RoomID,
			SenderID: client.UserIDHex,
		}

		msgBytes, _ := json.Marshal(broadcastMsg)
		client.Hub.Broadcast(msg.RoomID, msgBytes, client.UserIDHex)

	case "read":
		// Mark messages as read
		if msg.RoomID == "" {
			return
		}

		roomID, err := helpers.GenerateFromString(msg.RoomID)
		if err != nil {
			return
		}

		_ = sqlite.NewQuery(app.DB).Chat.MarkRoomMessagesAsRead(nil, db_chat.MarkRoomMessagesAsReadParams{
			UserID: client.UserID,
			RoomID: roomID,
		})

		// Notify other participants that messages were read
		broadcastMsg := services.WSMessage{
			Type:     "read",
			RoomID:   msg.RoomID,
			SenderID: client.UserIDHex,
		}

		msgBytes, _ := json.Marshal(broadcastMsg)
		client.Hub.Broadcast(msg.RoomID, msgBytes, client.UserIDHex)
	}
}

// GetUserChatRooms fetches and subscribes user to their chat rooms
func SubscribeUserToRooms(client *services.Client, db *sql.DB) error {
	rooms, err := sqlite.NewQuery(db).Chat.GetUserChatList(nil, db_chat.GetUserChatListParams{
		SenderID: client.UserID,
		UserID:   client.UserID,
		UserID_2: client.UserID,
	})
	if err != nil {
		return err
	}

	for _, room := range rooms {
		roomIDHex, _ := helpers.GenerateFromBytes(room.RoomID)
		client.Hub.Subscribe(client, roomIDHex)
	}

	return nil
}
