# Notification System

## How it works

Notifications are created from the backend when users do certain actions. We dont expose an endpoint for the frontend to create notifications directly because that would be insecure (users could spoof notifications, spam, etc)

## Notification types

These are the valid notification types (defined in the notifications table CHECK constraint):
- `follow_request` - When someone sends you a follow request
- `group_invitation` - When someone invites you to a group
- `group_request` - When someone requests to join your group
- `group_event` - When theres a new group event
- `message` - When someone sends you a message

## How to add notifications to your feature

1. Import the helper at the top of your handler file:
```go
"social-network/internal/helpers"
```

2. Call `helpers.CreateNotification()` after the action completes:
```go
err = helpers.CreateNotification(app, receiverID, "follow_request", currentUserID, nil, nil)
if err != nil {
    app.Logger.Error("failed to create notification", "err", err)
    // Don't fail the request if notification fails
}
```

### Parameters:
- `app` - The app instance
- `receiverID` - The user who should receive the notification (as []byte)
- `notifType` - One of the valid types above (as string)
- `fromID` - The user who triggered the notification (as []byte)
- `groupID` - Optional group ID if relevant (as []byte, use nil if not needed)
- `eventID` - Optional event ID if relevant (as []byte, use nil if not needed)

### Example from followers.go:

```go
// Create follow request
err := sqlite.NewQuery(app.DB).Followers.CreateFollowRequest(r.Context(),
    db_followers.CreateFollowRequestParams{
        FollowerID: currentUserID,
        FolloweeID: user.UserID,
    })
if err != nil {
    utils.Internal(w, err)
    return
}

// Notify the user they got a follow request
err = helpers.CreateNotification(app, user.UserID, "follow_request", currentUserID, nil, nil)
if err != nil {
    app.Logger.Error("failed to create follow request notification", "err", err)
    // Don't fail the request if notification fails
}
```

## Important notes

- Always use `[]byte` for IDs when calling the helper (not string/hex)
- The helper uses `context.Background()` internally so it wont block your request
- If notification creation fails, we log the error but dont fail the user's request
- The notification gets a unique UUID automatically
- Timestamps are set automatically
- Notifications start as unseen by default

## Where to add notifications

Add notification creation in these handlers when:
- Someone sends a follow request (done in `FollowUser`)
- Someone accepts your follow request (todo in `UpdateFollowRequest`)
- Someone reacts to your post
- Someone comments on your post
- Someone invites you to a group
- Someone requests to join your group
- Someone creates a group event
- Someone sends you a message

## Frontend endpoints

The frontend can use these endpoints to work with notifications:
- `GET /api/notifications/` - Get all notifications
- `GET /api/notifications/unseen` - Get only unseen ones
- `GET /api/notifications/details` - Get notifications with user details (JOIN with users table)
- `GET /api/notifications/unseen/count` - Get count of unseen
- `PUT /api/notifications/{notificationId}/seen` - Mark one as seen
- `PUT /api/notifications/seen/all` - Mark all as seen
- `DELETE /api/notifications/{notificationId}` - Delete one

All endpoints require authentication