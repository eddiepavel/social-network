package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	db_chat "social-network/pkg/db/queries/chat"
	"social-network/pkg/db/sqlite"

	"github.com/google/uuid"
)

// PrivateMessageHandler handles private messages between users
// TODO implement message storage logic
func PrivateMessageHandler(event Event, c *Client, db *sql.DB) error {
	var msg PrivateMessageEvent
	if err := json.Unmarshal(event.Payload, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Validate sender is the connected client
	if c == nil {
		return fmt.Errorf("unauthenticated client")
	}

	// Set the sender to the connected client's ID
	msg.From = c.userID
	msg.Sent = time.Now()

	// Find the recipient
	c.manager.RLock()
	recipient, ok := c.manager.clients[msg.To]
	c.manager.RUnlock()

	if !ok {
		c.manager.Logger.Warn("Recipient not online", "from", msg.From, "to", msg.To)
		return fmt.Errorf("recipient not online")
	}

	// Marshal and send to recipient
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	recipient.egress <- Event{
		Type:    EventPrivateMessage,
		Payload: payload,
	}

	c.manager.Logger.Info("Private message sent", "from", msg.From, "to", msg.To)
	return nil
}

// SendMessageHandler handles chat messages sent via WebSocket
// It stores the message in the DB and broadcasts to all room participants
func SendMessageHandler(event Event, c *Client, db *sql.DB) error {
	var msg SendMessageEvent
	if err := json.Unmarshal(event.Payload, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal send_message: %w", err)
	}

	if c == nil {
		return fmt.Errorf("unauthenticated client")
	}

	if msg.RoomID == "" || msg.Content == "" {
		return fmt.Errorf("room_id and content are required")
	}

	// Convert room ID from string to bytes
	roomUUID, err := uuid.Parse(msg.RoomID)
	if err != nil {
		return fmt.Errorf("invalid room_id: %w", err)
	}
	roomIDBytes, err := roomUUID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal room_id: %w", err)
	}

	// Convert sender ID from string to bytes
	senderUUID, err := uuid.Parse(c.userID)
	if err != nil {
		return fmt.Errorf("invalid sender user ID: %w", err)
	}
	senderIDBytes, err := senderUUID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal sender ID: %w", err)
	}

	// Verify sender is a participant in the room
	isParticipant, err := sqlite.NewQuery(db).Chat.CheckUserIsParticipant(context.Background(), db_chat.CheckUserIsParticipantParams{
		UserID: senderIDBytes,
		RoomID: roomIDBytes,
	})
	if err != nil || isParticipant == 0 {
		return fmt.Errorf("user is not a participant of this room")
	}

	// Create message ID
	messageID, err := uuid.New().MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to generate message ID: %w", err)
	}

	// Store message in DB
	err = sqlite.NewQuery(db).Chat.CreateMessage(context.Background(), db_chat.CreateMessageParams{
		MessageID: messageID,
		Content:   msg.Content,
		SenderID:  senderIDBytes,
		TargetID:  roomIDBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	messageUUIDStr, _ := uuid.FromBytes(messageID)
	now := time.Now()

	// Fetch sender info for broadcast
	senderFirstName := ""
	senderLastName := ""
	senderAvatar := ""
	senderInfo, err := sqlite.NewQuery(db).Chat.GetUserBasicInfo(context.Background(), senderIDBytes)
	if err == nil {
		senderFirstName = senderInfo.FirstName
		senderLastName = senderInfo.LastName
		if senderInfo.Avatar.Valid {
			senderAvatar = senderInfo.Avatar.String
		}
	}

	// Broadcast to all room participants
	c.manager.BroadcastChatMessage(roomIDBytes, ChatMessageEvent{
		MessageID:       messageUUIDStr.String(),
		RoomID:          msg.RoomID,
		SenderID:        c.userID,
		Content:         msg.Content,
		CreatedAt:       now,
		SenderFirstName: senderFirstName,
		SenderLastName:  senderLastName,
		SenderAvatar:    senderAvatar,
	})

	c.manager.Logger.Info("Chat message stored and broadcast", "from", c.userID, "room", msg.RoomID)
	return nil
}
