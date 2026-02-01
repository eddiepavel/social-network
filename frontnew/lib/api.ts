import type {
  ApiEnvelope,
  ChatMessage,
  ChatThread,
  Comment,
  FeedPost,
  Follower,
  FollowRequest,
  FollowStatus,
  Group,
  GroupDetails,
  GroupJoinRequest,
  Post,
  PostWithDetails,
  SearchUser,
  User,
} from "@/lib/types";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE?.replace(/\/$/, "") ||
  "http://localhost:8000";

type ApiOptions = RequestInit & { raw?: boolean };

async function apiFetch<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const response = await fetch(`${BASE_URL}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
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
    throw new Error(message);
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

export function createPost(input: { content: string; image_id?: string; visibility: string }) {
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
  avatar?: string;
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

export function searchUsers(name: string) {
  const params = new URLSearchParams({ name });
  return apiFetch<SearchUser[]>(`/api/users/search?${params.toString()}`);
}

export function getGroups() {
  return apiFetch<Group[]>("/api/groups/all");
}

export function getGroup(groupId: string) {
  return apiFetch<GroupDetails>(`/api/groups/group/${groupId}`);
}

export function requestJoinGroup(groupId: string) {
  return apiFetch<string>(`/api/groups/members/request/${groupId}`, {
    method: "POST",
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

export function followUser(userId: string) {
  return apiFetch<string>(`/api/followers/follow/${userId}`, {
    method: "POST",
  });
}

export function unfollowUser(userId: string) {
  return apiFetch<string>(`/api/followers/unfollow/${userId}`, {
    method: "DELETE",
  });
}

export function getFollowers(userId: string) {
  return apiFetch<Follower[]>(`/api/followers/${userId}/followers`);
}

export function getFollowing(userId: string) {
  return apiFetch<Follower[]>(`/api/followers/${userId}/following`);
}

export function getFollowRequests() {
  return apiFetch<FollowRequest[]>("/api/followers/requests");
}

export function respondToFollowRequest(requestId: string, accept: boolean) {
  return apiFetch<string>(`/api/followers/requests/${requestId}`, {
    method: "PUT",
    body: JSON.stringify({ accept }),
  });
}

export function getFollowStatus(userId: string) {
  return apiFetch<{ status: FollowStatus }>(`/api/followers/status/${userId}`);
}

// ============================================
// POST ENHANCEMENTS API
// ============================================

export function getPost(postId: string) {
  return apiFetch<PostWithDetails>(`/api/posts/${postId}`);
}

export function editPost(postId: string, input: { content?: string; visibility?: string }) {
  return apiFetch<Post>(`/api/posts/${postId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deletePost(postId: string) {
  return apiFetch<string>(`/api/posts/${postId}`, {
    method: "DELETE",
  });
}

export function toggleReaction(postId: string) {
  return apiFetch<{ reacted: boolean; count: number }>(`/api/posts/${postId}/react`, {
    method: "POST",
  });
}

export function updatePostVisibility(postId: string, visibility: string) {
  return apiFetch<Post>(`/api/posts/${postId}/visibility`, {
    method: "PUT",
    body: JSON.stringify({ visibility }),
  });
}

export function addPrivateViewer(postId: string, userId: string) {
  return apiFetch<string>(`/api/posts/${postId}/viewers/${userId}`, {
    method: "POST",
  });
}

export function removePrivateViewer(postId: string, userId: string) {
  return apiFetch<string>(`/api/posts/${postId}/viewers/${userId}`, {
    method: "DELETE",
  });
}

// ============================================
// COMMENTS API
// ============================================

export function getComments(postId: string) {
  return apiFetch<Comment[]>(`/api/posts/${postId}/comments`);
}

export function createComment(postId: string, input: { content: string }) {
  return apiFetch<Comment>(`/api/posts/${postId}/comments`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function editComment(commentId: string, input: { content: string }) {
  return apiFetch<Comment>(`/api/comments/${commentId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteComment(commentId: string) {
  return apiFetch<string>(`/api/comments/${commentId}`, {
    method: "DELETE",
  });
}

export function toggleCommentReaction(commentId: string) {
  return apiFetch<{ reacted: boolean; count: number }>(`/api/comments/${commentId}/react`, {
    method: "POST",
  });
}

// ============================================
// FILE UPLOAD API
// ============================================

export async function uploadFile(file: File): Promise<{ file_id: string; url: string }> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch(`${BASE_URL}/api/files/upload`, {
    method: "POST",
    credentials: "include",
    body: formData,
  });

  const payload = (await response.json()) as ApiEnvelope<{ file_id: string; url: string }>;
  if (!response.ok || payload.error) {
    throw new Error(payload.error?.message || "Upload failed");
  }
  if (!payload.data) {
    throw new Error("Unexpected empty response");
  }
  return payload.data;
}

// ============================================
// CHAT API
// ============================================

export function getRoomMessages(roomId: string) {
  return apiFetch<ChatMessage[]>(`/api/chat/${roomId}/messages`);
}

export function sendMessage(roomId: string, content: string) {
  return apiFetch<ChatMessage>(`/api/chat/${roomId}`, {
    method: "POST",
    body: JSON.stringify({ content }),
  });
}

export function startNewChat(userId: string) {
  return apiFetch<{ room_id: string }>(`/api/chat/start/${userId}`, {
    method: "POST",
  });
}

// ============================================
// GROUPS MANAGEMENT API
// ============================================

export function inviteToGroup(groupId: string, userId: string) {
  return apiFetch<string>(`/api/groups/members/invite/${groupId}/${userId}`, {
    method: "POST",
  });
}

export function getGroupRequests(groupId: string) {
  return apiFetch<GroupJoinRequest[]>(`/api/groups/members/requests/${groupId}`);
}

export function respondToGroupRequest(groupId: string, userId: string, accept: boolean) {
  return apiFetch<string>(`/api/groups/members/requests/${groupId}/${userId}`, {
    method: "PUT",
    body: JSON.stringify({ accept }),
  });
}

export function leaveGroup(groupId: string) {
  return apiFetch<string>(`/api/groups/members/leave/${groupId}`, {
    method: "DELETE",
  });
}

export function removeMember(groupId: string, userId: string) {
  return apiFetch<string>(`/api/groups/members/${groupId}/${userId}`, {
    method: "DELETE",
  });
}

export function updateGroup(groupId: string, input: { group_name?: string; description?: string; image?: string }) {
  return apiFetch<Group>(`/api/groups/${groupId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteGroup(groupId: string) {
  return apiFetch<string>(`/api/groups/${groupId}`, {
    method: "DELETE",
  });
}
