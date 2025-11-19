AUTHENTICATION
    POST /api/auth/register [PUBLIC]
        body: email, password, first_name, last_name, dob, avatar?, nickname?, about_me?
        returns: user object, session

    POST /api/auth/login [PUBLIC]
        body: email, password
        returns: user object, sets session cookie

    POST /api/auth/logout [PROTECTED]
        clears session

    GET /api/auth/session [PROTECTED]
        returns: current user if logged in


USERS & PROFILE
    GET /api/users/:id [PROTECTED]
        returns: user profile (respects privacy settings)

    PUT /api/users/profile [PROTECTED - own profile only]
        body: first_name?, last_name?, nickname?, about_me?, avatar?
        returns: updated user

    PUT /api/users/privacy [PROTECTED - own profile only]
        body: is_public (boolean)
        returns: updated user

    GET /api/users/:id/posts [PROTECTED]
        returns: user's posts (respects privacy)


FOLLOWERS
    POST /api/follow/:userId [PROTECTED]
        sends follow request or auto-follows if public
        returns: follower relationship

    DELETE /api/follow/:userId [PROTECTED]
        unfollows user

    POST /api/follow/accept/:followerId [PROTECTED - must be followee]
        accepts pending follow request

    POST /api/follow/reject/:followerId [PROTECTED - must be followee]
        rejects follow request

    GET /api/followers/:userId [PROTECTED]
        returns: list of followers

    GET /api/following/:userId [PROTECTED]
        returns: list of users being followed

    GET /api/follow/requests [PROTECTED]
        returns: pending follow requests for current user


POSTS
    POST /api/posts [PROTECTED]
        body: content, image_id?, visibility ('public'|'semi-private'|'private'), allowed_users? (array of user_ids for private posts)
        returns: created post

    GET /api/posts [PROTECTED]
        returns: feed of visible posts for current user

    GET /api/posts/:id [PROTECTED]
        returns: post with comments (if allowed to view)

    DELETE /api/posts/:id [PROTECTED - author only]
        deletes own post

    POST /api/posts/:id/comments [PROTECTED]
        body: content, image_id?, parent_comment_id?
        returns: created comment

    GET /api/posts/:id/comments [PROTECTED]
        returns: comments for post


REACTIONS
    POST /api/reactions [PROTECTED]
        body: target_type ('post'|'comment'), target_id, reaction_type
        returns: created reaction

    DELETE /api/reactions/:id [PROTECTED - own reaction only]
        removes reaction

    PUT /api/reactions/:id [PROTECTED - own reaction only]
        updates reaction type

    GET /api/reactions/:targetType/:targetId [PROTECTED]
        returns: all reactions for post or comment


GROUPS [IMPLEMENTED]
    POST /api/groups [PROTECTED]
        body: group_name (required), description (required), image?
        returns: created group object
        notes: creator automatically added as member with status='joined'

    GET /api/groups [PROTECTED]
        returns: all groups with member counts
        includes: group_id, group_name, description, image, creator_id, created_at, member_count

    GET /api/groups/:id [PROTECTED - members only]
        returns: group details with members list
        access: only members with status='joined' can view
        includes: full group info + array of members with user details

    POST /api/groups/:id/invite [PROTECTED - members only]
        body: user_id (required)
        action: invites user to group, sets status='requested', records invited_by
        access: only group members can invite
        returns: success message
        notes: invited user must be accepted by creator to join

    POST /api/groups/:id/request [PROTECTED]
        body: none
        action: current user requests to join group, sets status='requested'
        returns: success message
        notes: requires creator approval to join

    POST /api/groups/:id/accept/:userId [PROTECTED - creator only]
        body: action ('accept' or 'reject')
        action: creator accepts or rejects pending join requests
        access: only group creator can accept/reject
        returns: success message
        notes: sets status to 'joined' or 'rejected'


GROUP POSTS
    POST /api/groups/:id/posts [PROTECTED - members only]
        body: content, image_id?
        returns: created group post

    GET /api/groups/:id/posts [PROTECTED - members only]
        returns: posts in group (members only)

    POST /api/groups/posts/:postId/comments [PROTECTED - members only]
        body: content, image_id?, parent_comment_id?
        returns: created comment

    GET /api/groups/posts/:postId/comments [PROTECTED - members only]
        returns: comments for group post


GROUP EVENTS
    POST /api/groups/:id/events [PROTECTED - members only]
        body: title, description, event_timestamp
        returns: created event

    GET /api/groups/:id/events [PROTECTED - members only]
        returns: events for group

    GET /api/events/:id [PROTECTED - group members only]
        returns: event details with RSVPs

    POST /api/events/:id/rsvp [PROTECTED - group members only]
        body: status ('going'|'maybe'|'not going')
        returns: updated rsvp

    GET /api/events/:id/rsvps [PROTECTED - group members only]
        returns: list of rsvps for event


CHAT (REST endpoints, actual chat via WebSocket)
    GET /api/chat/private/:userId [PROTECTED - must be following or followed by user]
        returns: chat history with specific user

    GET /api/chat/group/:groupId [PROTECTED - group members only]
        returns: group chat history

    POST /api/chat/private [PROTECTED - must be following or followed by receiver]
        body: receiver_id, content
        returns: created message (also sends via WebSocket)

    POST /api/chat/group [PROTECTED - group members only]
        body: group_id, content
        returns: created message (also broadcasts via WebSocket)


WEBSOCKET
    WS /ws/chat [PROTECTED - authenticated via session cookie]
        connects to chat hub
        receives real-time messages
        can send messages through socket
        message format: {type: 'private'|'group', to: userId|groupId, content: string}


NOTIFICATIONS
    GET /api/notifications [PROTECTED]
        returns: notifications for current user

    PUT /api/notifications/:id/seen [PROTECTED - own notifications only]
        marks notification as seen

    GET /api/notifications/unread [PROTECTED]
        returns: count of unread notifications

    DELETE /api/notifications/:id [PROTECTED - own notifications only]
        deletes notification


IMAGES
    POST /api/images/upload [PROTECTED]
        multipart/form-data: image file
        validates: JPEG, PNG, GIF
        returns: image_id, path

    GET /api/images/:filename [PROTECTED - conditional access]
        serves image file
        validates access permissions for private content
