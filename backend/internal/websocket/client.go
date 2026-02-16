package websocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type ClientsConnected map[string]*Client

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	userID     string
	egress     chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager, userID string) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		userID:     userID,
		egress:     make(chan Event, 16),
	}
}

func (c *Client) readMessages() {
	defer func() {
		if c.manager != nil && c.manager.Logger != nil {
			c.manager.Logger.Info("Closing client (readMessages)", "userID", c.userID)
		}
		close(c.egress)
	}()

	c.connection.SetReadLimit(512)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if c.manager != nil && c.manager.Logger != nil {
					c.manager.Logger.Warn("could not read connection", "userID", c.userID, "err", err)
				}
			}
			break
		}

		var event Event
		if err := json.Unmarshal(payload, &event); err != nil {
			if c.manager != nil && c.manager.Logger != nil {
				c.manager.Logger.Warn("failed to unmarshal event", "userID", c.userID, "err", err)
			}
			continue
		}

		if err := c.manager.routeEvent(event, c); err != nil {
			if c.manager != nil && c.manager.Logger != nil {
				c.manager.Logger.Warn("routeEvent error", "userID", c.userID, "type", event.Type, "err", err)
			}
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		if c.manager != nil && c.manager.Logger != nil {
			c.manager.Logger.Info("Closing client (writeMessages)", "userID", c.userID)
		}
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					if c.manager != nil && c.manager.Logger != nil {
						c.manager.Logger.Warn("could not write close message", "userID", c.userID, "err", err)
					}
				}
				if c.manager != nil && c.manager.Logger != nil {
					c.manager.Logger.Info("egress closed (writeMessages)", "userID", c.userID)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				if c.manager != nil && c.manager.Logger != nil {
					c.manager.Logger.Warn("marshal error (writeMessages)", "userID", c.userID, "err", err)
				}
				return
			}

			if c.manager != nil && c.manager.Logger != nil {
				c.manager.Logger.Info("Sending message to client", "userID", c.userID, "type", message.Type)
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				if c.manager != nil && c.manager.Logger != nil {
					c.manager.Logger.Warn("could not write message", "userID", c.userID, "err", err)
				}
				return
			}
		}
	}
}
