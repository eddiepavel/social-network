export type ApiEnvelope<T> = {
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: unknown;
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
