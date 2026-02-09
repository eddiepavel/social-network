export type ApiEnvelope<T> = {
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: string | Record<string, string>;
  };
  pagination?: {
    page: number;
    size: number;
    current: number;
    total_items: number;
    total_pages: number;
  };
};

export type User = {
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  dob: string;
  avatar?: string;
  nickname?: string;
  about_me?: string;
  is_public?: boolean;
  created_at?: string;
};

export type SearchUser = {
  user_id: string;
  first_name: string;
  last_name: string;
  nickname?: string;
};

export type FeedPost = {
  post_id: string;
  content: string;
  image_id?: string | null;
  image_url?: string;
  visibility: string;
  author_id: string;
  author_name?: string;
  author_avatar?: string;
  created_at: string;
  reaction_count: number;
  user_reacted: boolean;
  comment_count: number;
};

export type Post = {
  post_id: string;
  content: string;
  image_id?: string | null;
  visibility: string;
  author_id: string;
  created_at: string;
};

export type Group = {
  group_id: string;
  group_name: string;
  description: string;
  image?: string;
  image_url?: string;
  creator_id?: string;
  created_at?: string;
  member_count?: number;
  is_owner?: boolean;
};

export type GroupMember = {
  user_id: string;
  status: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  can_remove_member?: boolean;
};

export type GroupDetails = {
  group: Group;
  members: GroupMember[];
};

export type ChatThread = {
  room_id: string;
  room_name?: string;
  is_group: boolean;
  last_message_id?: string;
  last_message_content?: string;
  last_message_time?: string;
  last_message_sender_id?: string;
  unread_count: number;
};

export type ChatMessage = {
  message_id: string;
  room_id: string;
  sender_id: string;
  content: string;
  created_at: string;
  sender_first_name?: string;
  sender_last_name?: string;
  sender_avatar?: string;
};

export type ChatMessagesResponse = {
  messages: ChatMessage[];
  has_more: boolean;
  next_cursor?: {
    cursor_timestamp: string;
    cursor_id: string;
  };
};

export type Follower = {
  user_id: string;
  first_name: string;
  last_name: string;
  nickname?: string;
  avatar?: string;
  followed_at?: string;
};

export type FollowRequest = {
  request_id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  nickname?: string;
  avatar?: string;
  created_at: string;
};

export type FollowStatus = "none" | "following" | "requested" | "self";

export type Comment = {
  comment_id: string;
  post_id: string;
  author_id: string;
  content: string;
  created_at: string;
  author_first_name?: string;
  author_last_name?: string;
  author_avatar?: string;
  reaction_count?: number;
  user_reacted?: boolean;
};

export type PostWithDetails = FeedPost & {
  author_first_name?: string;
  author_last_name?: string;
  author_avatar?: string;
  comments?: Comment[];
  allowed_viewers?: string[];
};

export type GroupEvent = {
  event_id: string;
  event_name: string;
  description: string;
  timestamp: string;
  created_at: string;
  going_count: number;
  not_going_count: number;
  user_rsvp?: string | null;
};

export type Notification = {
  notif_id: string;
  type: "follow_request" | "group_invitation" | "group_request" | "group_event" | "message";
  is_seen: boolean;
  from_id: string;
  from_name?: string;
  group_id?: string;
  group_name?: string;
  event_id?: string;
  event_title?: string;
  created_at: string;
};

export type NotificationsResponse = {
  notifications: Notification[];
  unread_count: number;
};

export type GroupJoinRequest = {
  request_id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  nickname?: string;
  avatar?: string;
  created_at: string;
};

// WebSocket message types
export type WSMessageType = "message" | "typing" | "read" | "subscribe" | "unsubscribe";

export type WSMessage = {
  type: WSMessageType;
  room_id?: string;
  content?: string;
  sender_id?: string;
  timestamp?: string;
  data?: Record<string, unknown>;
};

// Group post creation request
export type CreateGroupPostRequest = {
  content: string;
  image_id?: string;
};

// Event creation request
export type CreateEventRequest = {
  title: string;
  description: string;
  timestamp: string; // ISO 8601 format
};

// RSVP request
export type RSVPRequest = {
  status: "going" | "not going" | "maybe";
};
