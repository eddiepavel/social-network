package websocket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
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

// TypingIndicatorHandler handles typing indicators
func TypingIndicatorHandler(event Event, c *Client, db *sql.DB) error {
	var typingEvent TypingIndicatorEvent
	if err := json.Unmarshal(event.Payload, &typingEvent); err != nil {
		return fmt.Errorf("failed to unmarshal typing indicator: %w", err)
	}

	// Validate sender
	if c == nil {
		return fmt.Errorf("unauthenticated client")
	}

	// Set the sender info
	typingEvent.From = c.userID

	// Find the recipient
	c.manager.RLock()
	recipient, ok := c.manager.clients[typingEvent.To]
	c.manager.RUnlock()

	if !ok {
		// Recipient not online, ignore silently
		return nil
	}

	// Marshal and send
	payload, err := json.Marshal(typingEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal typing indicator: %w", err)
	}

	recipient.egress <- Event{
		Type:    EventTypingIndicator,
		Payload: payload,
	}

	return nil
}
