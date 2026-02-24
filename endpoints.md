# 📡 Pulse API Documentation

Complete API reference for the Pulse social network backend. All endpoints use JSON for request and response bodies unless otherwise specified.

**Base URL:** `http://localhost:8000/api`

**WebSocket URL:** `ws://localhost:8000/ws`

---

## 📑 Table of Contents

- [Authentication](#authentication)
- [Users & Profiles](#users--profiles)
- [Followers](#followers)
- [Posts](#posts)
- [Comments](#comments)
- [Reactions](#reactions)
- [Groups](#groups)
- [Group Events](#group-events)
- [Group Posts](#group-posts)
- [Chat](#chat)
- [Notifications](#notifications)
- [File Upload](#file-upload)
- [WebSocket](#websocket)

---

## 🔐 Authentication

Session-based authentication using HTTP-only cookies. Sessions persist until logout.

### Register

Create a new user account.

**Endpoint:** `POST /api/register`

**Access:** Public

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "dob": "1995-06-15",
  "avatar": "image-id-from-upload",
  "nickname": "johndoe",
  "about_me": "Love connecting with people!"
}
```

**Required Fields:**
- `email` (unique, valid email format)
- `password` (minimum 8 characters)
- `firstName`
- `lastName`
- `dob` (format: YYYY-MM-DD)

**Optional Fields:**
- `avatar` (image URL or path)
- `nickname`
- `aboutMe`

**Response:** `201 Created`
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "nickname": "johndoe",
  "avatar": "https://example.com/avatar.jpg",
  "isPublic": false
}
```

---

### Login

Authenticate and create a session.

**Endpoint:** `POST /api/login`

**Access:** Public

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:** `200 OK`
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "isPublic": false
}
```

**Note:** Session cookie is set automatically in response headers.

---

### Get Session

Check current session status.

**Endpoint:** `GET /api/session`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "isPublic": false
}
```

**Response (Not Authenticated):** `401 Unauthorized`

---

### Logout

Destroy current session.

**Endpoint:** `POST /api/logout`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "message": "Successfully logged out"
}
```

---

## 👤 Users & Profiles

### Get User Profile

Retrieve user profile information. Privacy settings determine visible fields.

**Endpoint:** `GET /api/users/profile/{id}`

**Access:** Authenticated

**Path Parameters:**
- `id` - User ID (UUID format)

**Response (Public Profile):** `200 OK`
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "dob": "1995-06-15",
  "avatar": "/storage/image/avatar.jpg",
  "nickname": "johndoe",
  "aboutMe": "Love connecting with people!",
  "isPublic": true,
  "followerCount": 42,
  "followingCount": 38,
  "isFollowing": false,
  "isFollower": true
}
```

**Response (Private Profile - Not Following):** `200 OK`
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "firstName": "John",
  "lastName": "Doe",
  "avatar": "/storage/image/avatar.jpg",
  "isPublic": false,
  "message": "This profile is private"
}
```

---

### Get User Posts

Get posts created by a specific user.

**Endpoint:** `GET /api/users/profile/{id}/posts`

**Access:** Authenticated

**Query Parameters:**
- `limit` (optional, default: 20) - Number of posts to return
- `offset` (optional, default: 0) - Pagination offset

**Response:** `200 OK`
```json
{
  "posts": [
    {
      "post_id": "650e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "author": {
        "firstName": "John",
        "lastName": "Doe",
        "avatar": "/storage/image/avatar.jpg"
      },
      "content": "Just shared an amazing moment!",
      "image": "/storage/image/post123.jpg",
      "visibility": "public",
      "createdAt": "2026-02-24T10:30:00Z",
      "reactionCount": 15,
      "commentCount": 3
    }
  ],
  "total": 42
}
```

---

### Update Profile

Update current user's profile information.

**Endpoint:** `PUT /api/users/profile`

**Access:** Authenticated (own profile only)

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "nickname": "johndoe",
  "about_me": "Updated bio!",
  "avatar_id": "image-id-from-upload"
}
```

**Note:** All fields are optional. Only include fields you want to update.

**Response:** `200 OK`
```json
{
  "message": "Profile updated successfully",
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "firstName": "John",
    "lastName": "Doe",
    "nickname": "johndoe",
    "aboutMe": "Updated bio!",
    "avatar": "/storage/image/newavatar.jpg"
  }
}
```

---

### Update Privacy Setting

Toggle profile between public and private.

**Endpoint:** `PUT /api/users/privacy`

**Access:** Authenticated

**Request Body:**
```json
{
  "is_public": true
}
```

**Response:** `200 OK`
```json
{
  "message": "Privacy settings updated",
  "isPublic": true
}
```

**Note:** Changing to public profile automatically accepts all pending follow requests.

---

### Search Users

Search for users by name or email.

**Endpoint:** `GET /api/users/search`

**Access:** Authenticated

**Query Parameters:**
- `q` (required) - Search query (minimum 2 characters)
- `limit` (optional, default: 20)

**Response:** `200 OK`
```json
{
  "users": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "firstName": "John",
      "lastName": "Doe",
      "nickname": "johndoe",
      "avatar": "/storage/image/avatar.jpg",
      "isPublic": true
    },
    {
      "user_id": "650e8400-e29b-41d4-a716-446655440002",
      "firstName": "Jane",
      "lastName": "Smith",
      "avatar": "/storage/image/jane.jpg",
      "isPublic": false
    }
  ]
}
```

---

## 👥 Followers

### Follow User

Send a follow request or instantly follow if the target user has a public profile. This endpoint toggles follow status - if already following, it will unfollow.

**Endpoint:** `POST /api/followers/user/{userId}/follow`

**Access:** Authenticated

**Path Parameters:**
- `userId` - Target user ID

**Request Body:** None (this is a toggle endpoint)

**Response (Public Profile - Now Following):** `200 OK`
```json
{
  "message": "Successfully following user",
  "status": "following"
}
```

**Response (Private Profile):** `200 OK`
```json
{
  "message": "Follow request sent",
  "status": "pending"
}
```

**Response (Already Following):** `200 OK`
```json
{
  "message": "Successfully unfollowed user",
  "status": "unfollowed"
}
```

**Note:** This endpoint toggles follow status. If already following, it unfollows.

---

### Get Follow Requests

Get pending follow requests for the current user (when profile is private).

**Endpoint:** `GET /api/followers/requests`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "requests": [
    {
      "request_id": "750e8400-e29b-41d4-a716-446655440003",
      "follower_id": "850e8400-e29b-41d4-a716-446655440004",
      "follower": {
        "firstName": "Alice",
        "lastName": "Johnson",
        "avatar": "/storage/image/alice.jpg"
      },
      "createdAt": "2026-02-23T15:20:00Z"
    }
  ]
}
```

---

### Respond to Follow Request

Accept or decline a follow request.

**Endpoint:** `POST /api/followers/requests/{requestId}/respond`

**Access:** Authenticated (request recipient only)

**Path Parameters:**
- `requestId` - Follow request ID

**Request Body:**
```json
{
  "response": "accept"
}
```

**Valid Responses:** `"accept"` or `"decline"`

**Response:** `200 OK`
```json
{
  "message": "Follow request accepted"
}
```

---

### Get Follow Status

Check follow relationship with another user.

**Endpoint:** `GET /api/followers/status/{userId}`

**Access:** Authenticated

**Path Parameters:**
- `userId` - Target user ID

**Response:** `200 OK`
```json
{
  "isFollowing": true,
  "isFollower": false,
  "hasPendingRequest": false
}
```

---

### Get Followers

Get list of users following a specific user.

**Endpoint:** `GET /api/followers/user/{userId}/followers`

**Access:** Authenticated

**Path Parameters:**
- `userId` - User ID

**Response:** `200 OK`
```json
{
  "followers": [
    {
      "user_id": "950e8400-e29b-41d4-a716-446655440005",
      "firstName": "Bob",
      "lastName": "Williams",
      "avatar": "/storage/image/bob.jpg",
      "isFollowing": true
    }
  ],
  "count": 42
}
```

**Note:** For private profiles, only visible to the profile owner and their followers.

---

### Get Following

Get list of users that a specific user is following.

**Endpoint:** `GET /api/followers/user/{userId}/following`

**Access:** Authenticated

**Response:** Same format as Get Followers

---

## 📝 Posts

### Create Post

Create a new post with optional image.

**Endpoint:** `POST /api/posts/create`

**Access:** Authenticated

**Request Body:**
```json
{
  "content": "Just had an amazing day at the beach!",
  "image_id": "image-id-from-upload",
  "visibility": "public",
  "allowed_users": []
}
```

**Fields:**
- `content` (required) - Post text content
- `image_id` (optional) - Image ID from upload endpoint
- `visibility` (required) - `"public"`, `"semi-private"`, or `"private"`
- `allowed_users` (required for private) - Array of user IDs who can view

**Visibility Rules:**
- `public` - All users can see (only if profile is public)
- `semi-private` - Only followers can see
- `private` - Only selected followers can see

**Response:** `201 Created`
```json
{
  "post_id": "650e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Just had an amazing day at the beach!",
  "image": "/storage/image/beach.jpg",
  "visibility": "public",
  "createdAt": "2026-02-24T12:00:00Z"
}
```

---

### Get Feed Posts

Get personalized feed of posts from followed users and groups.

**Endpoint:** `GET /api/posts/feed`

**Access:** Authenticated

**Query Parameters:**
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "posts": [
    {
      "post_id": "650e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "author": {
        "firstName": "John",
        "lastName": "Doe",
        "avatar": "/storage/image/avatar.jpg"
      },
      "content": "Just had an amazing day at the beach!",
      "image": "/storage/image/beach.jpg",
      "visibility": "public",
      "group_id": null,
      "createdAt": "2026-02-24T12:00:00Z",
      "reactionCount": 15,
      "commentCount": 3,
      "userReaction": "like"
    }
  ],
  "hasMore": true
}
```

---

### Get Post Details

Get a specific post with comments and reactions.

**Endpoint:** `GET /api/posts/id/{postId}`

**Access:** Authenticated

**Path Parameters:**
- `postId` - Post ID

**Response:** `200 OK`
```json
{
  "post": {
    "post_id": "650e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "author": {
      "firstName": "John",
      "lastName": "Doe",
      "avatar": "/storage/image/avatar.jpg"
    },
    "content": "Just had an amazing day at the beach!",
    "image": "/storage/image/beach.jpg",
    "visibility": "public",
    "createdAt": "2026-02-24T12:00:00Z"
  },
  "comments": [],
  "reactions": {
    "like": 15,
    "love": 3
  },
  "userReaction": "like"
}
```

---

### Edit Post

Update post content and/or image.

**Endpoint:** `PUT /api/posts/id/{postId}`

**Access:** Authenticated (post author only)

**Request Body:**
```json
{
  "content": "Updated post content!",
  "image_id": "new-image-id"
}
```

**Note:** Both fields are optional. Only include fields you want to update.

**Response:** `200 OK`
```json
{
  "message": "Post updated successfully",
  "post": {
    "post_id": "650e8400-e29b-41d4-a716-446655440001",
    "content": "Updated post content!",
    "image": "/storage/image/newimage.jpg",
    "updatedAt": "2026-02-24T13:00:00Z"
  }
}
```

---

### Delete Post

Delete a post permanently.

**Endpoint:** `DELETE /api/posts/id/{postId}`

**Access:** Authenticated (post author only)

**Response:** `200 OK`
```json
{
  "message": "Post deleted successfully"
}
```

---

### Update Post Visibility

Change who can see a post.

**Endpoint:** `PUT /api/posts/id/{postId}/privacy`

**Access:** Authenticated (post author only)

**Request Body:**
```json
{
  "visibility": "semi-private"
}
```

**Response:** `200 OK`
```json
{
  "message": "Post visibility updated",
  "visibility": "semi-private"
}
```

---

### Manage Private Post Access

Add or remove specific users from private post viewing list.

**Add User:**

**Endpoint:** `POST /api/posts/id/{postId}/privacy`

**Request Body:**
```json
{
  "user_id": "850e8400-e29b-41d4-a716-446655440004"
}
```

**Remove User:**

**Endpoint:** `DELETE /api/posts/id/{postId}/privacy`

**Request Body:**
```json
{
  "user_id": "850e8400-e29b-41d4-a716-446655440004"
}
```

---

## 💬 Comments

### Get Comments

Get all comments for a specific post.

**Endpoint:** `GET /api/posts/id/{postId}/comment`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "comments": [
    {
      "comment_id": "750e8400-e29b-41d4-a716-446655440006",
      "post_id": "650e8400-e29b-41d4-a716-446655440001",
      "user_id": "850e8400-e29b-41d4-a716-446655440004",
      "author": {
        "firstName": "Alice",
        "lastName": "Johnson",
        "avatar": "/storage/image/alice.jpg"
      },
      "content": "Amazing photo!",
      "image": null,
      "createdAt": "2026-02-24T12:15:00Z",
      "reactionCount": 2,
      "userReaction": null
    }
  ]
}
```

---

### Create Comment

Add a comment to a post.

**Endpoint:** `POST /api/posts/id/{postId}/comment`

**Access:** Authenticated

**Request Body:**
```json
{
  "content": "Great post!",
  "image": "/storage/image/reaction.gif"
}
```

**Response:** `201 Created`
```json
{
  "comment_id": "750e8400-e29b-41d4-a716-446655440006",
  "post_id": "650e8400-e29b-41d4-a716-446655440001",
  "user_id": "850e8400-e29b-41d4-a716-446655440004",
  "content": "Great post!",
  "image": "/storage/image/reaction.gif",
  "createdAt": "2026-02-24T12:15:00Z"
}
```

---

### Edit Comment

Update comment content.

**Endpoint:** `PUT /api/posts/id/{postId}/comment/{commentId}`

**Access:** Authenticated (comment author only)

**Request Body:**
```json
{
  "content": "Updated comment text"
}
```

**Response:** `200 OK`

---

### Delete Comment

Remove a comment permanently.

**Endpoint:** `DELETE /api/posts/id/{postId}/comment/{commentId}`

**Access:** Authenticated (comment author only)

**Response:** `200 OK`

---

## 👍 Reactions

### Toggle Reaction on Post

Add or remove your reaction to a post. This is a simple like/unlike system.

**Endpoint:** `POST /api/posts/id/{postId}/reaction`

**Access:** Authenticated

**Request Body:** None

**Response (Added):** `200 OK`
```json
{
  "message": "Reaction added",
  "user_reacted": true
}
```

**Response (Removed):** `200 OK`
```json
{
  "message": "Reaction removed",
  "user_reacted": false
}
```

**Note:** This is a toggle endpoint. Calling it when you haven't reacted adds a reaction. Calling it when you have reacted removes your reaction.

---

### Get Post Reactions

Get reaction count and your reaction status for a post.

**Endpoint:** `GET /api/posts/id/{postId}/reaction`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "count": 15,
  "user_reacted": true
}
```

**Fields:**
- `count` - Total number of reactions on the post
- `user_reacted` - Whether the current user has reacted to this post

---

### Toggle Reaction on Comment

Same as post reactions - a simple like/unlike toggle for comments.

**Endpoint:** `POST /api/posts/id/{postId}/comment/{commentId}/reaction`

**Access:** Authenticated

**Request Body:** None

**Response:** Same format as post reactions

---

## 👥 Groups

### Get All Groups

Browse all available groups.

**Endpoint:** `GET /api/groups/all`

**Access:** Authenticated

**Query Parameters:**
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "groups": [
    {
      "group_id": "950e8400-e29b-41d4-a716-446655440007",
      "title": "Photography Enthusiasts",
      "description": "Share your best shots!",
      "image": "/storage/image/camera.jpg",
      "creator_id": "550e8400-e29b-41d4-a716-446655440000",
      "createdAt": "2026-02-20T10:00:00Z",
      "memberCount": 24,
      "userMembershipStatus": "joined"
    }
  ]
}
```

**User Membership Status:**
- `"not-member"` - Not part of the group
- `"pending"` - Join request pending
- `"invited"` - Has pending invitation
- `"joined"` - Active member

---

### Get Group Details

Get detailed information about a specific group.

**Endpoint:** `GET /api/groups/group/{groupId}`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "group_id": "950e8400-e29b-41d4-a716-446655440007",
  "title": "Photography Enthusiasts",
  "description": "Share your best shots!",
  "image": "/storage/image/camera.jpg",
  "creator_id": "550e8400-e29b-41d4-a716-446655440000",
  "creator": {
    "firstName": "John",
    "lastName": "Doe",
    "avatar": "/storage/image/avatar.jpg"
  },
  "createdAt": "2026-02-20T10:00:00Z",
  "members": [
    {
      "user_id": "850e8400-e29b-41d4-a716-446655440004",
      "firstName": "Alice",
      "lastName": "Johnson",
      "avatar": "/storage/image/alice.jpg",
      "status": "joined",
      "joinedAt": "2026-02-21T09:00:00Z"
    }
  ],
  "memberCount": 24,
  "isCreator": false,
  "isMember": true
}
```

---

### Create Group

Create a new group.

**Endpoint:** `POST /api/groups/create`

**Access:** Authenticated

**Request Body:**
```json
{
  "group_name": "Photography Enthusiasts",
  "description": "A place for photographers to share and discuss their work",
  "image": "image-id-from-upload"
}
```

**Response:** `201 Created`
```json
{
  "group_id": "950e8400-e29b-41d4-a716-446655440007",
  "title": "Photography Enthusiasts",
  "description": "A place for photographers to share and discuss their work",
  "image": "/storage/image/camera.jpg",
  "creator_id": "550e8400-e29b-41d4-a716-446655440000",
  "createdAt": "2026-02-24T14:00:00Z"
}
```

**Note:** Creator is automatically added as a member with "joined" status.

---

### Update Group

Update group details (creator only).

**Endpoint:** `PUT /api/groups/update/{groupId}`

**Access:** Authenticated (group creator only)

**Request Body:**
```json
{
  "group_name": "Advanced Photography",
  "description": "Updated description",
  "image": "new-image-id"
}
```

**Response:** `200 OK`

---

### Delete Group

Permanently delete a group (creator only).

**Endpoint:** `DELETE /api/groups/delete/{groupId}`

**Access:** Authenticated (group creator only)

**Response:** `200 OK`

---

### Invite to Group

Invite a user to join the group.

**Endpoint:** `POST /api/groups/invite/{groupId}`

**Access:** Authenticated (group members only)

**Request Body:**
```json
{
  "users": [
    "850e8400-e29b-41d4-a716-446655440004",
    "950e8400-e29b-41d4-a716-446655440005"
  ]
}
```

**Response:** `200 OK`
```json
{
  "message": "Invitation sent successfully"
}
```

**Note:** Creates a notification for the invited user.

---

### Request to Join Group

Request membership in a group or remove pending request.

**Endpoint:** `POST /api/groups/members/request/{groupId}`

**Access:** Authenticated

**Request Body:**
```json
{
  "action": "request"
}
```

**Valid Actions:**
- `"request"` - Send join request
- `"remove"` - Cancel pending request
- `"accept_invite"` - Accept group invitation
- `"decline_invite"` - Decline group invitation

**Response:** `200 OK`
```json
{
  "message": "created"
}
```

**Note:** Creates a notification for the group creator.

---

### Get Group Join Requests

Get pending join requests for a group (creator only).

**Endpoint:** `GET /api/groups/members/requests/{groupId}`

**Access:** Authenticated (group creator only)

**Response:** `200 OK`
```json
{
  "requests": [
    {
      "user_id": "850e8400-e29b-41d4-a716-446655440004",
      "firstName": "Alice",
      "lastName": "Johnson",
      "avatar": "/storage/image/alice.jpg",
      "requestedAt": "2026-02-23T16:00:00Z"
    }
  ]
}
```

---

### Respond to Join Request or Invitation

Accept or decline a group join request or invitation.

**Endpoint:** `POST /api/groups/members/respond/{groupId}`

**Access:** Authenticated

**Request Body (Creator responding to join request):**
```json
{
  "user_id": "850e8400-e29b-41d4-a716-446655440004",
  "response": "accept"
}
```

**Valid Responses:** `"accept"` or `"decline"`

**Response:** `200 OK`
```json
{
  "message": "Request accepted"
}
```

---

### Remove Member / Leave Group

Remove a member from the group or leave the group.

**Endpoint:** `POST /api/groups/members/remove/{groupId}`

**Access:** Authenticated

**Request Body (Creator removing member):**
```json
{
  "user_id": "850e8400-e29b-41d4-a716-446655440004"
}
```

**Request Body (Member leaving):**
```json
{
  "action": "leave"
}
```

**Response:** `200 OK`

---

## 📅 Group Events

### Get Group Events

Get all events for a specific group.

**Endpoint:** `GET /api/groups/group/{groupId}/events`

**Access:** Authenticated (group members only)

**Response:** `200 OK`
```json
{
  "events": [
    {
      "event_id": "a50e8400-e29b-41d4-a716-446655440008",
      "group_id": "950e8400-e29b-41d4-a716-446655440007",
      "creator_id": "550e8400-e29b-41d4-a716-446655440000",
      "creator": {
        "firstName": "John",
        "lastName": "Doe",
        "avatar": "/storage/image/avatar.jpg"
      },
      "title": "Photo Walk Downtown",
      "description": "Let's explore the city together!",
      "eventTimestamp": "2026-03-01T14:00:00Z",
      "createdAt": "2026-02-24T10:00:00Z",
      "goingCount": 12,
      "notGoingCount": 3,
      "userResponse": "going"
    }
  ]
}
```

---

### Create Group Event

Create a new event for a group.

**Endpoint:** `POST /api/groups/group/{groupId}/events`

**Access:** Authenticated (group members only)

**Request Body:**
```json
{
  "title": "Photo Walk Downtown",
  "description": "Let's explore the city together and capture urban beauty!",
  "timestamp": "2026-03-01T14:00:00Z"
}
```

**Response:** `201 Created`
```json
{
  "event_id": "a50e8400-e29b-41d4-a716-446655440008",
  "group_id": "950e8400-e29b-41d4-a716-446655440007",
  "creator_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Photo Walk Downtown",
  "description": "Let's explore the city together and capture urban beauty!",
  "eventTimestamp": "2026-03-01T14:00:00Z",
  "createdAt": "2026-02-24T15:00:00Z"
}
```

**Note:** All group members receive a notification about the new event.

---

### RSVP to Event

Respond to a group event invitation.

**Endpoint:** `POST /api/events/{eventId}/rsvp`

**Access:** Authenticated (group members only)

**Request Body:**
```json
{
  "response": "going"
}
```

**Valid Responses:** `"going"`, `"not-going"`

**Response:** `200 OK`
```json
{
  "message": "RSVP recorded",
  "response": "going"
}
```

**Note:** Users can change their RSVP by calling the endpoint again with a different response.

---

### Get Event Details

Get detailed information about an event including all RSVPs.

**Endpoint:** `GET /api/events/{eventId}/rsvp`

**Access:** Authenticated (group members only)

**Response:** `200 OK`
```json
{
  "event": {
    "event_id": "a50e8400-e29b-41d4-a716-446655440008",
    "title": "Photo Walk Downtown",
    "description": "Let's explore the city together!",
    "eventTimestamp": "2026-03-01T14:00:00Z"
  },
  "rsvps": {
    "going": [
      {
        "user_id": "850e8400-e29b-41d4-a716-446655440004",
        "firstName": "Alice",
        "lastName": "Johnson",
        "avatar": "/storage/image/alice.jpg"
      }
    ],
    "not-going": []
  },
  "counts": {
    "going": 12,
    "not-going": 3
  },
  "userResponse": "going"
}
```

---

## 📄 Group Posts

### Get Group Posts

Get all posts in a specific group.

**Endpoint:** `GET /api/groups/group/{groupId}/posts`

**Access:** Authenticated (group members only)

**Query Parameters:**
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "posts": [
    {
      "post_id": "650e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "group_id": "950e8400-e29b-41d4-a716-446655440007",
      "author": {
        "firstName": "John",
        "lastName": "Doe",
        "avatar": "/storage/image/avatar.jpg"
      },
      "content": "Check out this amazing sunset!",
      "image": "/storage/image/sunset.jpg",
      "createdAt": "2026-02-24T18:00:00Z",
      "reactionCount": 8,
      "commentCount": 2
    }
  ]
}
```

---

### Create Group Post

Create a post in a group.

**Endpoint:** `POST /api/groups/group/{groupId}/posts`

**Access:** Authenticated (group members only)

**Request Body:**
```json
{
  "content": "Check out this amazing sunset!",
  "image": "/storage/image/sunset.jpg"
}
```

**Response:** `201 Created`
```json
{
  "post_id": "650e8400-e29b-41d4-a716-446655440001",
  "group_id": "950e8400-e29b-41d4-a716-446655440007",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Check out this amazing sunset!",
  "image": "/storage/image/sunset.jpg",
  "createdAt": "2026-02-24T18:00:00Z"
}
```

**Note:** Group posts are only visible to group members and don't have visibility settings.

---

## 💬 Chat

### Get Chat List

Get all chat rooms for the current user.

**Endpoint:** `GET /api/chat/list`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "chats": [
    {
      "room_id": "b50e8400-e29b-41d4-a716-446655440009",
      "roomType": "direct",
      "roomName": null,
      "group_id": null,
      "participants": [
        {
          "user_id": "850e8400-e29b-41d4-a716-446655440004",
          "firstName": "Alice",
          "lastName": "Johnson",
          "avatar": "/storage/image/alice.jpg"
        }
      ],
      "lastMessage": {
        "content": "Hey! How are you?",
        "createdAt": "2026-02-24T16:30:00Z",
        "sender_id": "850e8400-e29b-41d4-a716-446655440004"
      },
      "unreadCount": 2
    },
    {
      "room_id": "c50e8400-e29b-41d4-a716-446655440010",
      "roomType": "group",
      "roomName": "Photography Chat",
      "group_id": "950e8400-e29b-41d4-a716-446655440007",
      "lastMessage": {
        "content": "Don't forget the event tomorrow!",
        "createdAt": "2026-02-24T17:00:00Z",
        "sender_id": "550e8400-e29b-41d4-a716-446655440000"
      },
      "unreadCount": 0
    }
  ]
}
```

---

### Get Room Messages

Get message history for a chat room.

**Endpoint:** `GET /api/chat/{roomId}`

**Access:** Authenticated (room participant only)

**Query Parameters:**
- `limit` (optional, default: 50)
- `before` (optional) - Message ID for pagination (get messages before this ID)

**Response:** `200 OK`
```json
{
  "messages": [
    {
      "message_id": "d50e8400-e29b-41d4-a716-446655440011",
      "room_id": "b50e8400-e29b-41d4-a716-446655440009",
      "sender_id": "850e8400-e29b-41d4-a716-446655440004",
      "sender": {
        "firstName": "Alice",
        "lastName": "Johnson",
        "avatar": "/storage/image/alice.jpg"
      },
      "content": "Hey! How are you? 😊",
      "createdAt": "2026-02-24T16:30:00Z"
    }
  ],
  "hasMore": false
}
```

---

### Send Message

Send a message in a chat room.

**Endpoint:** `POST /api/chat/{roomId}`

**Access:** Authenticated (room participant only)

**Request Body:**
```json
{
  "content": "I'm doing great! Thanks for asking 😄"
}
```

**Response:** `201 Created`
```json
{
  "message_id": "e50e8400-e29b-41d4-a716-446655440012",
  "room_id": "b50e8400-e29b-41d4-a716-446655440009",
  "sender_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "I'm doing great! Thanks for asking 😄",
  "createdAt": "2026-02-24T16:35:00Z"
}
```

**Note:** Message is also sent via WebSocket to all connected participants.

---

### Create Chat Room and Send Message

Create a new direct chat room with a user and send the first message.

**Endpoint:** `POST /api/chat/new`

**Access:** Authenticated

**Request Body:**
```json
{
  "target_id": "850e8400-e29b-41d4-a716-446655440004",
  "content": "Hi! Want to collaborate on a photo project?"
}
```

**Response:** `201 Created`
```json
{
  "room_id": "b50e8400-e29b-41d4-a716-446655440009",
  "message": {
    "message_id": "d50e8400-e29b-41d4-a716-446655440011",
    "content": "Hi! Want to collaborate on a photo project?",
    "createdAt": "2026-02-24T16:30:00Z"
  }
}
```

**Validation:** Can only message users you follow or who follow you (or users with public profiles).

---

### Edit Room Name

Update the name of a chat room (group chats only).

**Endpoint:** `PUT /api/chat/{roomId}/name`

**Access:** Authenticated (room participant only)

**Request Body:**
```json
{
  "room_name": "Photography Discussion"
}
```

**Response:** `200 OK`

---

## 🔔 Notifications

### Get Notifications

Get all notifications for the current user.

**Endpoint:** `GET /api/notifications/`

**Access:** Authenticated

**Query Parameters:**
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response:** `200 OK`
```json
{
  "notifications": [
    {
      "notif_id": "f50e8400-e29b-41d4-a716-446655440013",
      "receiver_id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "follow_request",
      "from_id": "850e8400-e29b-41d4-a716-446655440004",
      "is_seen": false,
      "createdAt": "2026-02-24T15:00:00Z"
    },
    {
      "notif_id": "g50e8400-e29b-41d4-a716-446655440014",
      "receiver_id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "group_invite",
      "from_id": "950e8400-e29b-41d4-a716-446655440007",
      "group_id": "950e8400-e29b-41d4-a716-446655440007",
      "is_seen": false,
      "createdAt": "2026-02-24T14:30:00Z"
    }
  ]
}
```

**Notification Types:**
- `follow_request` - Someone sent you a follow request
- `follow_accepted` - Your follow request was accepted
- `group_invite` - You've been invited to a group
- `group_request` - Someone requested to join your group
- `group_accepted` - Your group join request was accepted
- `group_event` - New event created in a group you're in

---

### Get Unseen Notifications

Get only unseen notifications.

**Endpoint:** `GET /api/notifications/unseen`

**Access:** Authenticated

**Response:** Same format as Get Notifications, filtered to unseen only

---

### Get Notifications with User Details

Get notifications with full user details included.

**Endpoint:** `GET /api/notifications/details`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "notifications": [
    {
      "notif_id": "f50e8400-e29b-41d4-a716-446655440013",
      "type": "follow_request",
      "from": {
        "user_id": "850e8400-e29b-41d4-a716-446655440004",
        "firstName": "Alice",
        "lastName": "Johnson",
        "avatar": "/storage/image/alice.jpg"
      },
      "is_seen": false,
      "createdAt": "2026-02-24T15:00:00Z"
    }
  ]
}
```

---

### Get Unseen Notification Count

Get count of unseen notifications.

**Endpoint:** `GET /api/notifications/unseen/count`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "count": 5
}
```

---

### Mark Notification as Seen

Mark a specific notification as seen.

**Endpoint:** `PUT /api/notifications/{notificationId}/seen`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "message": "Notification marked as seen"
}
```

---

### Mark All Notifications as Seen

Mark all notifications as seen.

**Endpoint:** `PUT /api/notifications/seen/all`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "message": "All notifications marked as seen",
  "count": 5
}
```

---

### Delete Notification

Delete a specific notification.

**Endpoint:** `DELETE /api/notifications/{notificationId}`

**Access:** Authenticated

**Response:** `200 OK`
```json
{
  "message": "Notification deleted"
}
```

---

## 📤 File Upload

### Upload File

Upload an image file (JPEG, PNG, or GIF).

**Endpoint:** `POST /api/storage/upload`

**Access:** Authenticated

**Content-Type:** `multipart/form-data`

**Form Data:**
- `file` - The image file to upload

**Response:** `200 OK`
```json
{
  "path": "/storage/image/abc123def456.jpg",
  "url": "http://localhost:8000/api/storage/image/abc123def456.jpg"
}
```

**Validation:**
- File must be JPEG, PNG, or GIF format
- Maximum file size: 10MB (configurable)
- Generates unique filename to prevent collisions

**Note:** Uploaded files not used within 30 minutes are automatically cleaned up.

---

### Get Image

Retrieve an uploaded image.

**Endpoint:** `GET /api/storage/image/{image}`

**Access:** Public (authenticated users)

**Path Parameters:**
- `image` - Filename of the image

**Response:** Image file with appropriate Content-Type header

---

## 🔌 WebSocket

Real-time bidirectional communication for chat messages and notifications.

### Connect

Establish WebSocket connection.

**Endpoint:** `ws://localhost:8000/ws/connect`

**Access:** Authenticated (session cookie required)

**Connection Process:**
1. Client sends WebSocket upgrade request with session cookie
2. Server validates session and establishes connection
3. Client is registered in the WebSocket manager
4. Connection stays open for bi-directional messaging

---

### Message Format

All WebSocket messages use JSON format with an `event` field to identify the message type.

#### Client → Server Messages

**Join Chat Room:**
```json
{
  "event": "join_room",
  "roomId": "b50e8400-e29b-41d4-a716-446655440009"
}
```

**Leave Chat Room:**
```json
{
  "event": "leave_room"
}
```

**Send Chat Message:**
```json
{
  "event": "chat_message",
  "roomId": "b50e8400-e29b-41d4-a716-446655440009",
  "content": "Hello! 👋"
}
```

---

#### Server → Client Messages

**New Chat Message:**
```json
{
  "event": "chat_message",
  "data": {
    "message_id": "d50e8400-e29b-41d4-a716-446655440011",
    "room_id": "b50e8400-e29b-41d4-a716-446655440009",
    "sender_id": "850e8400-e29b-41d4-a716-446655440004",
    "sender": {
      "firstName": "Alice",
      "lastName": "Johnson",
      "avatar": "/storage/image/alice.jpg"
    },
    "content": "Hello! 👋",
    "createdAt": "2026-02-24T16:30:00Z"
  }
}
```

**New Notification:**
```json
{
  "event": "notification",
  "data": {
    "notif_id": "f50e8400-e29b-41d4-a716-446655440013",
    "receiver_id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "follow_request",
    "from_id": "850e8400-e29b-41d4-a716-446655440004",
    "from_name": "Alice Johnson",
    "is_seen": false,
    "createdAt": "2026-02-24T15:00:00Z"
  }
}
```

**Notification Update:**
```json
{
  "event": "notification_update",
  "data": {
    "notif_id": "f50e8400-e29b-41d4-a716-446655440013",
    "is_seen": true
  }
}
```

**Error:**
```json
{
  "event": "error",
  "message": "Invalid room ID"
}
```

---

### WebSocket Connection Management

**Heartbeat/Ping:**
The WebSocket connection implements automatic ping/pong to keep the connection alive.

**Reconnection:**
Client should implement reconnection logic with exponential backoff if connection drops.

**Multiple Tabs:**
Each browser tab/window establishes its own WebSocket connection. All connections for the same user receive relevant messages.

**Clean Disconnection:**
When a user logs out, the WebSocket connection should be closed gracefully and not attempt to reconnect.

---

## 📊 Common Response Codes

### Success Codes
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Successful, no response body

### Client Error Codes
- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Not authenticated or session expired
- `403 Forbidden` - Authenticated but not authorized for this resource
- `404 Not Found` - Resource doesn't exist
- `422 Unprocessable Entity` - Validation error

### Server Error Codes
- `500 Internal Server Error` - Unexpected server error

---

## 🔐 Authentication Notes

### Session Cookies
- Cookie name: `session_id` (or configured name)
- HTTP-only flag: `true`
- SameSite: `Strict` or `Lax`
- Secure flag: `true` in production

### Session Duration
- Sessions persist until explicit logout
- No automatic timeout by default (configurable)
- Sessions stored in database

### CORS
API supports CORS for frontend communication. Allowed origins are configurable.

---

## 🎯 Rate Limiting

Rate limiting may be implemented on certain endpoints to prevent abuse. If rate limited, you'll receive:

**Response:** `429 Too Many Requests`
```json
{
  "error": "Rate limit exceeded",
  "retryAfter": 60
}
```

---

## 📝 Notes

- All timestamps are in ISO 8601 format (UTC)
- UUIDs are used for all entity IDs
- Image paths are relative to the storage endpoint
- All endpoints require authentication unless marked as "Public"
- Privacy settings are strictly enforced at the database query level
- WebSocket connection requires valid session cookie

---

## 🐛 Error Response Format

All error responses follow this structure:

```json
{
  "error": {
    "code": "404",
    "message": "Resource not found",
    "details": {
      "field": "Additional context if applicable"
    }
  }
}
```

---

## 📮 Support

For API issues or questions:
- Check the code in `/backend/internal/handlers/`
- Review Postman collections in `/backend/`
- Contact: **Giannis** or **Eddie** on Discord

---

**Last Updated:** February 24, 2026  
**API Version:** 1.0  
**Backend Version:** Go 1.24.6
