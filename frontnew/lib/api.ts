import type {
  ApiEnvelope,
  ChatThread,
  FeedPost,
  Group,
  GroupDetails,
  Post,
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
