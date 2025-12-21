# WebSocket Implementation

## Architecture

The WebSocket implementation uses a manager-client pattern with event-based routing:

```
Client → HTTP Upgrade → WebSocket Manager → Event Routing → Event Handlers
```

## Key Files

- **`internal/websocket/manager.go`** - WebSocket connection manager, notification worker
- **`internal/websocket/client.go`** - Individual client connection handling
- **`internal/websocket/event.go`** - Event type definitions and constants
- **`internal/websocket/event_handlers.go`** - Event handler implementations
- **`internal/handlers/websocket.go`** - HTTP upgrade handler
- **`internal/routes/routes.go`** - WebSocket route registration (`wsRoutes()`)
- **`cmd/server/main.go`** - Manager initialization and startup

## How It Works

1. **Connection**: Client connects to `GET /api/ws/connect` (requires auth)
2. **Upgrade**: HTTP connection upgrades to WebSocket via `handlers.ConnectWebSocket()`
3. **Manager**: `websocket.Manager` maintains all active connections
4. **Events**: Client sends JSON events → Manager routes to appropriate handler
5. **Notifications**: Background worker polls database every 3 seconds, pushes to connected clients

## Event Flow

**Client → Server:**
```
Client sends: {"type": "private_message", "payload": {...}}
  ↓
Manager.routeEvent() looks up handler for "private_message"
  ↓
PrivateMessageHandler() processes the event
  ↓
Handler sends response to recipient's WebSocket
```

**Server → Client (Notifications):**
```
NotificationWorker runs every 3 seconds
  ↓
Queries database for unseen notifications
  ↓
Sends {"type": "notification", "payload": {...}} to client's WebSocket
```

## Adding New Features (Example: Chat Messages)

### 1. Define Event Type (`internal/websocket/event.go`)

```go
const (
    EventChatMessage = "chat_message"
)

type ChatMessageEvent struct {
    From      string    `json:"from"`
    To        string    `json:"to"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 2. Create Event Handler (`internal/websocket/event_handlers.go`)

```go
func ChatMessageHandler(event Event, c *Client, db *sql.DB) error {
    var msg ChatMessageEvent
    if err := json.Unmarshal(event.Payload, &msg); err != nil {
        return err
    }

    // Set sender
    msg.From = c.userID
    msg.Timestamp = time.Now()

    // TODO: Save message to database

    // Find recipient and send
    c.manager.RLock()
    recipient, ok := c.manager.clients[msg.To]
    c.manager.RUnlock()

    if ok {
        payload, _ := json.Marshal(msg)
        recipient.egress <- Event{
            Type:    EventChatMessage,
            Payload: payload,
        }
    }

    return nil
}
```

### 3. Register Handler (`internal/websocket/manager.go`)

In `setupEventHandlers()`:
```go
func (m *Manager) setupEventHandlers() {
    m.handlers[EventPrivateMessage] = PrivateMessageHandler
    m.handlers[EventTypingIndicator] = TypingIndicatorHandler
    m.handlers[EventChatMessage] = ChatMessageHandler  // Add this
}
```

That's it! Clients can now send:
```json
{"type": "chat_message", "payload": {"to": "user-id", "message": "Hello!"}}
```

## Client Connection Example (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8000/api/ws/connect');

// Send event
ws.send(JSON.stringify({
    type: 'private_message',
    payload: { to: 'user-id', message: 'Hi!' }
}));

// Receive events
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'notification') {
        console.log('Notification:', data.payload);
    }
};
```

## Notes

- WebSocket connections require authentication (session cookie)
- Manager automatically removes disconnected clients
- Notification worker runs in background goroutine
- All events are JSON with `{"type": "...", "payload": {...}}`
