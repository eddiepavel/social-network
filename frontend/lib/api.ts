import {
  ApiEnvelope,
  ChatMessage,
  ChatMessagesResponse,
  ChatThread,
  Comment,
  CreateEventRequest,
  CreateGroupPostRequest, EventRsvp,
  FeedPost,
  Follower,
  FollowRequest,
  Group,
  GroupDetails,
  GroupEvent,
  GroupJoinRequest,
  Notification,
  Post,
  PostWithDetails,
  RSVPRequest,
  SearchUser,
  User,
  WSMessage,
} from "@/lib/types";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE?.replace(/\/$/, "") ||
  "http://localhost:8000";

type ApiOptions = RequestInit & { raw?: boolean };

export class ApiError extends Error {
  code: string;
  details?: string | Record<string, string>;

  constructor(message: string, code: string, details?: string | Record<string, string>) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.details = details;
  }
}

async function apiFetch<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const response = await fetch(`${BASE_URL}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string> || {}),
    },
    ...options,
  });

  if (options.raw) {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return (await response.json()) as T;
  }

  const payload = (await response.json()) as ApiEnvelope<T>;
  if (!response.ok || payload.error) {
    const message = payload.error?.message || "Request failed.";
    const code = payload.error?.code || String(response.status);
    const details = payload.error?.details;
    throw new ApiError(message, code, details);
  }

  if (payload.data === undefined) {
    throw new Error("Unexpected empty response.");
  }

  return payload.data;
}

export function registerUser(input: {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  dob: string;
  avatar?: string;
  nickname?: string;
  about_me?: string;
}) {
  return apiFetch<User>("/api/public/register", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function loginUser(input: { email: string; password: string }) {
  return apiFetch<User>("/api/public/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function getSession() {
  try {
    return await apiFetch<User>("/api/auth/session");
  } catch {
    return null;
  }
}

export function logoutUser() {
  return apiFetch<string>("/api/auth/logout", { method: "POST" });
}

export async function getFeed(page = 1, size = 10) {
  const payload = await apiFetch<
    ApiEnvelope<{ data: FeedPost[]; pagination?: ApiEnvelope<FeedPost[]>["pagination"] }>
  >(`/api/posts/feed?page=${page}&size=${size}`, { raw: true });

  if (!payload.data) {
    throw new Error("Unexpected empty response.");
  }

  return {
    data: payload.data.data ?? [],
    pagination: payload.data.pagination,
  };
}

export function createPost(input: { content: string; image_id?: string; visibility: string; allowed_users?: string[] }) {
  return apiFetch<Post>("/api/posts/create", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getUserProfile(userId: string) {
  return apiFetch<User>(`/api/users/profile/${userId}`);
}

export function updateProfile(input: {
  first_name?: string;
  last_name?: string;
  nickname?: string;
  about_me?: string;
  avatar_id?: string;
}) {
  return apiFetch<User>("/api/users/profile", {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function updatePrivacy(isPublic: boolean) {
  return apiFetch<User>("/api/users/privacy", {
    method: "PUT",
    body: JSON.stringify({ is_public: isPublic }),
  });
}

export async function searchUsers(name: string) {
  const params = new URLSearchParams({ name });
  const result = await apiFetch<SearchUser[]>(`/api/users/search?${params.toString()}`);
  return result ?? [];
}

export function getGroups() {
  return apiFetch<Group[]>("/api/groups/all");
}

export function getGroup(groupId: string) {
  return apiFetch<GroupDetails>(`/api/groups/group/${groupId}`);
}

export function requestJoinGroup(groupId: string, payload: any) {
  return apiFetch<string>(`/api/groups/members/request/${groupId}`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function createGroup(input: { group_name: string; description: string; image?: string }) {
  return apiFetch<Group>("/api/groups/create", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getChatList() {
  return apiFetch<ChatThread[]>("/api/chat/list");
}

export function postMessage(roomId: string, input: { content: string }) {
  return apiFetch<string>(`/api/chat/${roomId}`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

// ============================================
// FOLLOWERS API
// ============================================

// Follow endpoint is a toggle - calling it again will unfollow
export function followUser(userId: string) {
  return apiFetch<string>(`/api/followers/user/${userId}/follow`, {
    method: "POST",
  });
}

// Unfollow uses the same toggle endpoint
export function unfollowUser(userId: string) {
  return apiFetch<string>(`/api/followers/user/${userId}/follow`, {
    method: "POST",
  });
}

export function getFollowers(userId: string) {
  return apiFetch<Follower[]>(`/api/followers/user/${userId}/followers`);
}

export function getFollowing(userId: string) {
  return apiFetch<Follower[]>(`/api/followers/user/${userId}/following`);
}

export function getFollowRequests() {
  return apiFetch<FollowRequest[]>("/api/followers/requests");
}

export function respondToFollowRequest(requestId: string, status: "accepted" | "rejected") {
  return apiFetch<string>(`/api/followers/requests/${requestId}/respond`, {
    method: "POST",
    body: JSON.stringify({ status }),
  });
}

// Get follow status using the dedicated endpoint
export function getFollowStatus(userId: string) {
  return apiFetch<{ status: "none" | "following" | "requested" | "self" }>(`/api/followers/status/${userId}`);
}

// ============================================
// POST ENHANCEMENTS API
// ============================================

export function getPost(postId: string) {
  return apiFetch<PostWithDetails>(`/api/posts/id/${postId}`);
}

export function editPost(postId: string, input: { content?: string; visibility?: string; image_id?: string }) {
  return apiFetch<Post>(`/api/posts/id/${postId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deletePost(postId: string) {
  return apiFetch<string>(`/api/posts/id/${postId}`, {
    method: "DELETE",
  });
}

export function toggleReaction(postId: string) {
  return apiFetch<{ reacted: boolean; count: number }>(`/api/posts/id/${postId}/reaction`, {
    method: "POST",
  });
}

export function updatePostVisibility(postId: string, visibility: string) {
  return apiFetch<Post>(`/api/posts/id/${postId}`, {
    method: "PUT",
    body: JSON.stringify({ visibility }),
  });
}

export function addPrivateViewer(postId: string, userId: string) {
  return apiFetch<string>(`/api/posts/id/${postId}/viewers/${userId}`, {
    method: "POST",
  });
}

export function removePrivateViewer(postId: string, userId: string) {
  return apiFetch<string>(`/api/posts/id/${postId}/viewers/${userId}`, {
    method: "DELETE",
  });
}

// ============================================
// COMMENTS API
// ============================================

export function getComments(postId: string) {
  return apiFetch<Comment[]>(`/api/posts/id/${postId}/comment`);
}

export function createComment(postId: string, input: { content: string; image_id?: string, parent_id: string }) {
  return apiFetch<Comment>(`/api/posts/id/${postId}/comment`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

// Note: Backend requires postId in path for comment operations
export function editComment(postId: string, commentId: string, input: { content: string; image_id?: string }) {
  return apiFetch<Comment>(`/api/posts/id/${postId}/comment/${commentId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteComment(postId: string, commentId: string) {
  return apiFetch<string>(`/api/posts/id/${postId}/comment/${commentId}`, {
    method: "DELETE",
  });
}

export function toggleCommentReaction(postId: string, commentId: string) {
  return apiFetch<{ reacted: boolean; count: number }>(`/api/posts/id/${postId}/comment/${commentId}/reaction`, {
    method: "POST",
  });
}

// ============================================
// FILE UPLOAD API
// ============================================

export async function uploadFile(file: File): Promise<{ file_id: string; url: string; filename: string; }> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch(`${BASE_URL}/api/storage/upload`, {
    method: "POST",
    credentials: "include",
    body: formData,
  } as RequestInit);

  const payload = (await response.json()) as ApiEnvelope<{ uuid: string; url: string, filename: string; }>;
  if (!response.ok || payload.error) {
    throw new Error(payload.error?.details as string || "Upload failed");
  }
  if (!payload.data) {
    throw new Error("Unexpected empty response");
  }
  return { file_id: payload.data.uuid, url: payload.data.url, filename: payload.data.filename };
}

// ============================================
// CHAT API
// ============================================

export async function getRoomMessages(roomId: string) {
  const response = await apiFetch<ChatMessagesResponse>(`/api/chat/${roomId}`);
  return response.messages ?? [];
}

export async function updateRoomName(roomId: string, room_name: string) {
  return apiFetch(`/api/chat/${roomId}/name`, {
    method: "PUT",
    body: JSON.stringify({ room_name: room_name })
  })
}

export function sendMessage(roomId: string, content: string) {
  return apiFetch<ChatMessage>(`/api/chat/${roomId}`, {
    method: "POST",
    body: JSON.stringify({ content }),
  });
}

export function startNewChat(userId: string, content: string = "👋") {
  return apiFetch<{ room_id: string }>("/api/chat/new", {
    method: "POST",
    body: JSON.stringify({ target_id: userId, content }),
  });
}

// ============================================
// GROUPS MANAGEMENT API
// ============================================

export function inviteToGroup(groupId: string, userId: string[]) {
  return apiFetch<string>(`/api/groups/invite/${groupId}`, {
    method: "POST",
    body: JSON.stringify({ users: userId }),
  });
}

export function getGroupRequests(groupId: string) {
  return apiFetch<GroupJoinRequest[]>(`/api/groups/members/requests/${groupId}`);
}

export function respondToGroupRequest(groupId: string, user_id: string, response: string) {
  return apiFetch<string>(`/api/groups/members/respond/${groupId}`, {
    method: "POST",
    body: JSON.stringify({ user_id: user_id, response: response }),
  });
}

export function leaveGroup(groupId: string, user_id: string) {
  return apiFetch<string>(`/api/groups/members/remove/${groupId}`, {
    method: "POST",
    body: JSON.stringify({user_id: user_id})
  });
}

export function removeMember(groupId: string, userId: string) {
  return apiFetch<string>(`/api/groups/members/remove/${groupId}`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId }),
  });
}

export function updateGroup(groupId: string, input: { group_name?: string; description?: string; image?: string }) {
  return apiFetch<Group>(`/api/groups/update/${groupId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteGroup(groupId: string) {
  return apiFetch<string>(`/api/groups/delete/${groupId}`, {
    method: "DELETE",
  });
}

// ============================================
// USER POSTS API
// ============================================

export function getUserPosts(userId: string, page = 1, size = 10) {
  return apiFetch<FeedPost[]>(`/api/users/profile/${userId}/posts?page=${page}&size=${size}`);
}

// ============================================
// NOTIFICATIONS API
// ============================================

export function getNotifications() {
  return apiFetch<Notification[]>(`/api/notifications/details`);
}

export function getUnseenNotifications() {
  return apiFetch<Notification[]>(`/api/notifications/unseen`);
}

export function getUnseenNotificationCount() {
  return apiFetch<{ count: number }>(`/api/notifications/unseen/count`);
}

export function markNotificationAsSeen(notifId: string) {
  return apiFetch<{ message: string }>(`/api/notifications/${notifId}/seen`, {
    method: "PUT",
  });
}

export function markAllNotificationsAsSeen() {
  return apiFetch<{ message: string }>("/api/notifications/seen/all", {
    method: "PUT",
  });
}

export function deleteNotification(notifId: string) {
  return apiFetch<{ message: string }>(`/api/notifications/${notifId}`, {
    method: "DELETE",
  });
}

// ============================================
// GROUP EVENTS API
// ============================================

export function createGroupEvent(groupId: string, input: CreateEventRequest) {
  return apiFetch<GroupEvent>(`/api/events/${groupId}/create`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function rsvpToEvent(eventId: string, status: RSVPRequest["status"]) {
  return apiFetch<{ message: string; status: string }>(`/api/events/${eventId}/rsvp`, {
    method: "POST",
    body: JSON.stringify({ status }),
  });
}

export function getEventRSVP(eventId: string) {
  return apiFetch<EventRsvp[]>(`/api/events/${eventId}/rsvp`, {
    method: "GET",
  });
}

// ============================================
// GROUP POSTS API
// ============================================

export function getGroupPosts(groupId: string, page = 1, size = 10) {
  return apiFetch<ApiEnvelope<FeedPost[]>>(`/api/groups/group/${groupId}/posts?page=${page}&size=${size}`);
}

export function createGroupPost(groupId: string, input: CreateGroupPostRequest) {
  return apiFetch<Post>(`/api/groups/group/${groupId}/posts`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

// ============================================
// WEBSOCKET CHAT API
// ============================================

const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_BASE?.replace(/\/$/, "") ||
  "ws://localhost:8000";

export type WebSocketCallbacks = {
  onMessage?: (message: WSMessage) => void;
  onTyping?: (message: WSMessage) => void;
  onRead?: (message: WSMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
};

export class ChatWebSocket {
  private ws: WebSocket | null = null;
  private callbacks: WebSocketCallbacks = {};
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectTimeout: NodeJS.Timeout | null = null;

  connect(callbacks: WebSocketCallbacks = {}) {
    this.callbacks = callbacks;
    this.createConnection();
  }

  private createConnection() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.ws = new WebSocket(`${WS_BASE_URL}/ws/connect`);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.callbacks.onOpen?.();
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        // The backend sends { type, payload } where payload is the actual data
        const payload = message.payload
          ? typeof message.payload === 'string'
            ? JSON.parse(message.payload)
            : message.payload
          : message.data;

        switch (message.type) {
          case "chat_message":
            this.callbacks.onMessage?.({
              type: "chat_message",
              data: payload,
            });
            break;
          case "private_message":
            this.callbacks.onMessage?.({
              type: "private_message",
              data: payload,
            });
            break;
          case "message":
            this.callbacks.onMessage?.(message);
            break;
          case "typing":
            this.callbacks.onTyping?.(message);
            break;
          case "read":
            this.callbacks.onRead?.(message);
            break;
          case "notification":
            // Forward notification events
            this.callbacks.onMessage?.({
              type: "notification",
              data: payload,
            });
            break;
        }
      } catch (err) {
        console.error("Failed to parse WebSocket message:", err);
      }
    };

    this.ws.onclose = () => {
      this.callbacks.onClose?.();
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      this.callbacks.onError?.(error);
    };
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error("Max reconnect attempts reached");
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);

    this.reconnectTimeout = setTimeout(() => {
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      this.createConnection();
    }, delay);
  }

  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
    this.ws?.close();
    this.ws = null;
  }

  sendMessage(roomId: string, content: string) {
    // Send as a proper Event { type, payload } matching backend format
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: "send_message",
        payload: JSON.stringify({ room_id: roomId, content }),
      }));
    } else {
      console.warn("WebSocket is not connected");
    }
  }

  sendTyping(roomId: string) {
    this.send({ type: "typing", room_id: roomId });
  }

  markAsRead(roomId: string) {
    this.send({ type: "read", room_id: roomId });
  }

  private send(message: WSMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket is not connected");
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton instance for the app
let chatWsInstance: ChatWebSocket | null = null;

export function getChatWebSocket(): ChatWebSocket {
  if (!chatWsInstance) {
    chatWsInstance = new ChatWebSocket();
  }
  return chatWsInstance;
}
