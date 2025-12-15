export type PostVisibility = 'public' | 'private' | 'semi-private';

export interface Post {
  post_id: string;
  author_id: string;
  content: string;
  image_id?: string | null;
  visibility: PostVisibility;
  created_at?: string;
}

export interface FeedPost extends Post {
  reaction_count: number;
  comment_count: number;
}

export interface PostReaction {
  reaction_id: string;
  user_id: string;
  reaction_type: string;
}

export interface PostComment {
  comment_id: string;
  user_id: string;
  content: string;
  reactions: PostReaction[];
}

export interface PostDetail extends Post {
  reactions: PostReaction[];
  comments: PostComment[];
}

export interface CreatePostRequest {
  content: string;
  visibility: PostVisibility;
  image_id?: string;
}

export interface UpdatePostRequest {
  content?: string;
  image_id?: string;
}

export interface SharePostRequest {
  user_id: string;
}
