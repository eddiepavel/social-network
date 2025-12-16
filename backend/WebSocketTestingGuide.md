# WebSocket Testing Guide with Postman

## Prerequisites

1. **Postman Desktop App** (WebSocket support requires desktop version, not web)
2. **Running Backend Server** (on port specified in .env)
3. **Valid User Account** (registered and can login)

## Step-by-Step Testing Guide

### Step 1: Start the Server

```bash
# Make sure you're in the backend directory
cd backend

# Start the server
go run cmd/server/main.go
```

You should see:
```
2025/12/16 10:30:00 database success
2025/12/16 10:30:00 Notification worker started
2025/12/16 10:30:00 Server starting on port 8000...
```

### Step 2: Register a User (If You Don't Have an Account)

#### 2.1 Create a Registration Request in Postman

**Method**: `POST`
**URL**: `http://localhost:8000/api/public/register`
**Headers**:
```
Content-Type: application/json
```
**Body** (JSON):
```json
{
    "email": "testuser@example.com",
    "password": "SecurePassword123",
    "first_name": "Test",
    "last_name": "User",
    "dob": "1990-01-01",
    "nickname": "testuser",
    "avatar": ""
}
```

**Field Descriptions**:
- `email`: Valid email address (required)
- `password`: User password (required)
- `first_name`: First name (required)
- `last_name`: Last name (required)
- `dob`: Date of birth in YYYY-MM-DD format (required)
- `nickname`: Optional username/nickname
- `avatar`: Optional base64 encoded image

#### 2.2 Send Registration Request

Click **Send**. You should get a `200 OK` or `201 Created` response with user details.

Example Response:
```json
{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "testuser@example.com",
    "first_name": "Test",
    "last_name": "User"
}
```

**Note**: Save the `user_id` - you'll need it for testing!

#### 2.3 Register a Second User (Optional - For Testing Messages)

Repeat the above with different details:
```json
{
    "email": "testuser2@example.com",
    "password": "SecurePassword456",
    "first_name": "Test",
    "last_name": "User2",
    "dob": "1992-05-15",
    "nickname": "testuser2"
}
```

Save this user_id too for testing private messages between users.

### Step 3: Login and Get Session Cookie

#### 3.1 Create a Login Request in Postman

**Method**: `POST`
**URL**: `http://localhost:8000/api/public/login`
**Headers**:
```
Content-Type: application/json
```
**Body** (JSON):
```json
{
    "email": "your-email@example.com",
    "password": "your-password"
}
```

#### 3.2 Send the Request

Click **Send**. You should get a `200 OK` response.

#### 3.3 Extract Session Cookie

1. Go to the **Headers** tab in the response
2. Look for `Set-Cookie` header
3. Copy the `session_id` value

Example:
```
Set-Cookie: session_id=550e8400-e29b-41d4-a716-446655440000; Path=/; Max-Age=604800; HttpOnly
```

Copy: `550e8400-e29b-41d4-a716-446655440000`

### Step 4: Connect to WebSocket

#### 4.1 Create WebSocket Request

1. Click **New** → **WebSocket Request**
2. Enter WebSocket URL:
   ```
   ws://localhost:8000/api/ws/connect
   ```

#### 4.2 Add Session Cookie

**Important**: Postman WebSocket doesn't have a separate Cookies tab. Instead:

**Option A - Use Headers Tab (Recommended)**:
1. Click on the **Headers** tab
2. Add a new header:
   - **Key**: `Cookie`
   - **Value**: `session_id=<paste-your-session-id-from-step-3.3>`

Example:
```
Cookie: session_id=550e8400-e29b-41d4-a716-446655440000
```

**Option B - Let Postman Auto-Send**:
1. If you logged in via Postman HTTP request in the same workspace, cookies should auto-send
2. Just make sure the login request was to the same domain (`localhost:8000`)

#### 4.3 Connect

1. Click **Connect**
2. You should see: `Connected to ws://localhost:8000/api/ws/connect`

Check your server logs - you should see:
```
INFO Client connected userID=<user-uuid> total_clients=1
```

### Step 5: Test Receiving Notifications (Server → Client)

Notifications are created automatically through business logic (not a direct endpoint). To trigger a notification:

#### 5.1 Trigger a Notification via Follow Request

**Prerequisites**: User2 must have a **private profile** (is_public = false)

**Step 1**: Connect User2 to WebSocket
- Use User2's session: `Cookie: session_id=<user2-session-id>`
- Connect to: `ws://localhost:8000/api/ws/connect`
- User2 should see: `Connected`

**Step 2**: Make User1 follow User2 (this creates the notification)
- **Method**: `POST`
- **URL**: `http://localhost:8000/api/followers/<user2-id>/follow`
- **Headers**:
  ```
  Cookie: session_id=<user1-session-id>
  ```
- **No body needed**

**What happens**:
1. If User2 has a **private profile**, a follow request is created
2. A `follow_request` notification is automatically created for User2
3. Within 3 seconds, User2 receives the notification via WebSocket

**Note**:
- If User2 has a **public profile**, they'll be auto-followed without a notification
- The `/api/notifications/` endpoint is READ-ONLY for retrieving notifications, not creating them
- To make User2's profile private: `PUT /api/users/privacy` with `{"is_public": false}`

#### 5.2 Wait for Push Notification

Within 3 seconds, you should receive a notification in your WebSocket connection:

```json
{
    "type": "notification",
    "payload": {
        "notif_id": "...",
        "receiver_id": "your-user-id",
        "type": "follow_request",
        "from_id": "sender-user-id",
        "created_at": "2025-12-16T10:30:00Z",
        "is_seen": false
    }
}
```

### Step 6: Test Sending Messages (Client → Server)

#### 6.1 Send Private Message

In your WebSocket connection, click **New Message** and send:

```json
{
    "type": "private_message",
    "payload": {
        "to_user": "<recipient-user-id>",
        "message": "Hello from Postman!"
    }
}
```

Click **Send**.

#### 6.2 Check Server Logs

You should see:
```
INFO Private message sent from=<your-user-id> to=<recipient-user-id>
```

If the recipient is online, they'll receive:
```json
{
    "type": "private_message",
    "payload": {
        "from_user": "your-user-id",
        "to_user": "recipient-user-id",
        "message": "Hello from Postman!",
        "sent": "2025-12-16T10:30:00Z"
    }
}
```

#### 6.3 Send Typing Indicator

Send this message:

```json
{
    "type": "typing_indicator",
    "payload": {
        "to_user": "<recipient-user-id>",
        "is_typing": true
    }
}
```

The recipient will receive:
```json
{
    "type": "typing_indicator",
    "payload": {
        "from_user": "your-user-id",
        "to_user": "recipient-user-id",
        "is_typing": true
    }
}
```

## Testing with Multiple Clients

### Option 1: Multiple Postman Windows

1. Open Postman Desktop App twice (or use different browsers)
2. Login with different accounts in each
3. Connect to WebSocket with each account's session
4. Send messages between them

### Option 2: Chrome/Firefox WebSocket Extensions

1. Install "Simple WebSocket Client" extension
2. Connect to `ws://localhost:8000/api/ws/connect`
3. Manually add session cookie in browser dev tools first

### Option 3: JavaScript Console (Quick Test)

1. Login in browser first
2. Open browser console (F12)
3. Run:

```javascript
const ws = new WebSocket('ws://localhost:8000/api/ws/connect');

ws.onopen = () => {
    console.log('Connected!');

    // Send a private message
    ws.send(JSON.stringify({
        type: 'private_message',
        payload: {
            to_user: 'recipient-user-id',
            message: 'Hello from browser!'
        }
    }));
};

ws.onmessage = (event) => {
    console.log('Received:', JSON.parse(event.data));
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};
```

## Common Test Scenarios

### Test 1: Notification Push
1. Connect User2 to WebSocket
2. Use User1 to follow User2 (or other action that creates notification)
3. Wait 3 seconds max
4. Verify User2 receives notification via WebSocket

### Test 2: Private Messages
1. Connect two clients (User A and User B)
2. User A sends private message to User B
3. Verify User B receives the message
4. User B replies
5. Verify User A receives the reply

### Test 3: Typing Indicators
1. Connect two clients
2. User A sends typing indicator to User B
3. Verify User B receives typing status
4. Send `is_typing: false` to stop

### Test 4: Offline Message Handling
1. Connect User A only
2. User A sends message to offline User B
3. Check server logs - should see "Recipient not online"
4. (Future enhancement: verify message stored in DB)

### Test 5: Reconnection
1. Connect WebSocket
2. Disconnect manually
3. Reconnect with same session
4. Verify it works

### Test 6: Invalid Event Type
1. Connect WebSocket
2. Send invalid event:
```json
{
    "type": "unknown_event",
    "payload": {}
}
```
3. Check server logs - should see error about unknown event type

## Troubleshooting

### "WebSocket connection failed"
- ✅ Check server is running
- ✅ Verify URL is correct: `ws://` not `wss://`
- ✅ Check port matches server port

### "Unauthorized" or immediate disconnect
- ✅ Verify session cookie is set correctly
- ✅ Check session hasn't expired (7 days default)
- ✅ Login again to get fresh session

### "Recipient not online"
- ✅ This is expected if recipient isn't connected
- ✅ Open another WebSocket connection for recipient
- ✅ Check recipient user_id is correct

### No notifications received
- ✅ Wait up to 3 seconds (polling interval)
- ✅ Check notification exists in database
- ✅ Verify `is_seen` is `false`
- ✅ Check `receiver_id` matches your user_id

### Messages in wrong format
- ✅ Ensure JSON is valid
- ✅ Check `type` field is correct
- ✅ Verify `payload` structure matches event type
- ✅ Don't include extra fields in root level

## Server Log Debugging

Enable detailed logging by checking these messages:

### Connection Events
```
INFO Client connected userID=<uuid> total_clients=1
INFO Client disconnected userID=<uuid> total_clients=0
```

### Message Events
```
INFO Sending message to client userID=<uuid> type=notification
INFO Private message sent from=<uuid> to=<uuid>
```

### Errors
```
WARN could not read connection userID=<uuid> err=...
WARN Handler error type=private_message userID=<uuid> err=...
WARN Client egress channel full userID=<uuid>
```

## Advanced: Automated Testing Script

Save as `test_websocket.js`:

```javascript
const WebSocket = require('ws');

const sessionId = 'your-session-id-here';
const ws = new WebSocket('ws://localhost:8000/api/ws/connect', {
    headers: {
        'Cookie': `session_id=${sessionId}`
    }
});

ws.on('open', () => {
    console.log('✅ Connected');

    // Test 1: Send private message
    setTimeout(() => {
        ws.send(JSON.stringify({
            type: 'private_message',
            payload: {
                to_user: 'recipient-uuid',
                message: 'Test message'
            }
        }));
        console.log('📤 Sent private message');
    }, 1000);

    // Test 2: Send typing indicator
    setTimeout(() => {
        ws.send(JSON.stringify({
            type: 'typing_indicator',
            payload: {
                to_user: 'recipient-uuid',
                is_typing: true
            }
        }));
        console.log('⌨️  Sent typing indicator');
    }, 2000);
});

ws.on('message', (data) => {
    const event = JSON.parse(data);
    console.log('📨 Received:', event.type, event.payload);
});

ws.on('error', (error) => {
    console.error('❌ Error:', error.message);
});

ws.on('close', () => {
    console.log('🔌 Disconnected');
});

// Keep alive for 30 seconds
setTimeout(() => ws.close(), 30000);
```

Run with:
```bash
node test_websocket.js
```

## Summary Checklist

Before reporting issues:

- [ ] Server is running
- [ ] Valid session cookie obtained via login
- [ ] WebSocket URL is correct (`ws://localhost:PORT/api/ws/connect`)
- [ ] Session cookie added to WebSocket request
- [ ] JSON messages are properly formatted
- [ ] User IDs are valid UUIDs
- [ ] Checked server logs for errors

## Next Steps

After basic testing works:
1. Integrate with your frontend
2. Add reconnection logic
3. Implement message persistence
4. Add more event types as needed
5. Set up production WebSocket server (wss://)

Happy testing! 🚀
