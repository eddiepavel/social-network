# WebSocket Implementation Documentation

## Overview

This WebSocket implementation provides real-time communication between the server and clients. It's built using the `gorilla/websocket` library and follows a clean, extensible architecture.

## Architecture

### Core Components

1. **Manager** (`pkg/ws/manager.go`)
   - Manages all connected clients
   - Routes incoming events to handlers
   - Runs background workers (e.g., notification polling)
   - Handles client connections/disconnections

2. **Client** (`pkg/ws/client.go`)
   - Represents a single WebSocket connection
   - Manages read/write goroutines
   - Has a buffered egress channel for outgoing messages

3. **Event System** (`pkg/ws/event.go`)
   - Defines all event types (notifications, messages, etc.)
   - Provides event structures with proper JSON serialization

4. **Event Handlers** (`pkg/ws/handlers.go`)
   - Contains handler functions for incoming events
   - Easy to extend with new handlers

## How It Works

### Connection Flow

1. Client connects to `/api/ws/connect` (requires authentication)
2. Server upgrades HTTP connection to WebSocket
3. Client is added to the manager's client map
4. Two goroutines start:
   - `readMessages()`: Reads incoming messages from client
   - `writeMessages()`: Writes outgoing messages to client

### Message Flow

#### Server → Client (Push)
```
Background Worker → Manager → Client.egress channel → writeMessages() → WebSocket → Client
```

#### Client → Server
```
Client → WebSocket → readMessages() → Manager.routeEvent() → Event Handler
```

### Background Notification Worker

- Runs every **3 seconds**
- For each connected client:
  1. Queries database for unseen notifications
  2. Converts notification data to JSON
  3. Pushes to client's egress channel
  4. Tracks sent notifications to avoid duplicates

## Event Types

### Server → Client Events (Outgoing)

| Event Type | Description | Payload Structure |
|------------|-------------|-------------------|
| `notification` | Push notifications to client | `NotificationEvent` |
| `error` | Error messages | `ErrorEvent` |
| `private_message` | Private messages | `PrivateMessageEvent` |

### Client → Server Events (Incoming)

| Event Type | Description | Handler Function |
|------------|-------------|------------------|
| `private_message` | Send private message | `PrivateMessageHandler` |
| `typing_indicator` | Typing status | `TypingIndicatorHandler` |

## Adding New Features

### Step 1: Define Event Type

In `pkg/ws/event.go`:

```go
const (
    EventGroupMessage = "group_message"
)

type GroupMessageEvent struct {
    GroupID string    `json:"group_id"`
    From    string    `json:"from_user"`
    Message string    `json:"message"`
    Sent    time.Time `json:"sent"`
}
```

### Step 2: Create Handler

In `pkg/ws/handlers.go`:

```go
func GroupMessageHandler(event Event, c *Client, db *sql.DB) error {
    var msg GroupMessageEvent
    if err := json.Unmarshal(event.Payload, &msg); err != nil {
        return fmt.Errorf("failed to unmarshal message: %w", err)
    }

    // Your logic here
    msg.From = c.userID
    msg.Sent = time.Now()

    // Broadcast to all group members
    // ...

    return nil
}
```

### Step 3: Register Handler

In `pkg/ws/manager.go`, add to `setupEventHandlers()`:

```go
func (m *Manager) setupEventHandlers() {
    m.handlers[EventPrivateMessage] = PrivateMessageHandler
    m.handlers[EventTypingIndicator] = TypingIndicatorHandler
    m.handlers[EventGroupMessage] = GroupMessageHandler  // Add this
}
```

That's it! Your new event type is now ready to use.

## Event Structure

All WebSocket messages follow this JSON format:

```json
{
    "type": "event_type_here",
    "payload": {
        // Event-specific data
    }
}
```

### Example: Notification Event
```json
{
    "type": "notification",
    "payload": {
        "notif_id": "123e4567-e89b-12d3-a456-426614174000",
        "receiver_id": "user-uuid",
        "type": "follow_request",
        "from_id": "sender-uuid",
        "created_at": "2025-12-16T10:30:00Z",
        "is_seen": false
    }
}
```

### Example: Private Message Event
```json
{
    "type": "private_message",
    "payload": {
        "from_user": "sender-uuid",
        "to_user": "recipient-uuid",
        "message": "Hello!",
        "sent": "2025-12-16T10:30:00Z"
    }
}
```

## Security

- **Authentication Required**: All WebSocket connections require valid session authentication via auth middleware
- **Origin Checking**: CORS is enforced via `checkOrigin()` function (allows empty origin for Postman testing)
- **User Validation**: User ID is extracted from authenticated session context using `middleware.GetUserIDFromContext()`

### Allowed Origins (Configurable)

The `checkOrigin` function in `pkg/ws/manager.go` controls which origins can connect:

```go
func checkOrigin(r *http.Request) bool {
    origin := r.Header.Get("Origin")

    // Allow empty origin (for tools like Postman)
    if origin == "" {
        return true
    }

    allowedOrigins := []string{
        "http://localhost:3000",  // React dev server
        "http://localhost:8000",  // Backend server
        "http://localhost:8080",  // Alternative frontend port
    }

    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }

    return false
}
```

**Important for Production**: Remove the empty origin check and add your production frontend URL to `allowedOrigins`.

## Configuration

### Manager Settings

- **Notification Poll Interval**: 3 seconds (configurable in `StartNotificationWorker`)
- **Client Buffer Size**: 16 messages per client
- **Read Limit**: 512 bytes per message

### WebSocket Upgrader

```go
websocketUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     checkOrigin,
}
```

## Error Handling

- Client disconnections are handled gracefully
- Failed message deliveries are logged
- Full egress channels are detected (logs warning)
- Malformed JSON is caught and logged

## Performance Considerations

1. **Non-blocking Sends**: Uses `select` with `default` to avoid blocking on full channels
2. **Buffered Channels**: Each client has a 16-message buffer
3. **Concurrent Safe**: Uses RWMutex for client map access
4. **Memory Management**: Sent notification cache is cleaned up periodically

## Testing

See the [Testing with Postman](#testing-with-postman) section below.

## Comparison with ws_old

### Similarities
- Manager/Client architecture ✓
- Event handler system ✓
- Read/write goroutines ✓
- Event routing ✓

### Improvements
- No import cycles ✓
- Uses google/uuid (consistent with codebase) ✓
- Background notification worker ✓
- Better error handling ✓
- Cleaner separation of concerns ✓
- Extensible event handler registration ✓

## Troubleshooting

### Client Can't Connect

**401 Unauthorized**
- Check authentication (valid session cookie required)
- Verify session is valid: test with `GET /api/auth/session` first
- Ensure Cookie header is set: `Cookie: session_id=your-session-id`

**403 Forbidden / Origin Not Allowed**
- Check CORS origin settings in `checkOrigin()` function
- For Postman testing: empty origin is allowed by default
- For browser clients: add your frontend URL to `allowedOrigins`

**404 Not Found**
- Verify WebSocket URL: `ws://localhost:PORT/api/ws/connect`
- Check server is running on the correct port

### "No User in Context" Error

If you get this error, it means the context key types don't match. **Solution**:
- Always use `middleware.GetUserIDFromContext(r.Context())` to get the user ID
- Don't define your own `contextKey` type - use the middleware's helper function
- This is already implemented correctly in `internal/routes/routes.go`

### Messages Not Received
- Check client's egress channel isn't full (buffer size: 16)
- Verify event handler is registered in `setupEventHandlers()`
- Check server logs for handler errors

### High Memory Usage
- Sent notification cache grows over time (auto-cleaned at 10k entries)
- Consider adjusting client buffer size
- Monitor connected client count

## API Endpoint

**WebSocket Connection**
```
WS /api/ws/connect
```

**Authentication**: Required (session cookie)

**Headers**:
```
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: <generated>
Sec-WebSocket-Version: 13
Cookie: session_id=<your-session-id>
```

## Future Enhancements

Potential features to add:
- [ ] Group message broadcasting
- [ ] Read receipts
- [ ] Online user list broadcasting
- [ ] Reconnection logic with message replay
- [ ] Message persistence for offline users
- [ ] Rate limiting
- [ ] Heartbeat/ping-pong for connection health
