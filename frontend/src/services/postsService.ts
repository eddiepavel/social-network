import { apiRequest } from './api';
import type {
  CreatePostRequest,
  FeedPost,
  Post,
  PostDetail,
  PostVisibility,
  SharePostRequest,
  UpdatePostRequest,
} from '../types';

export async function createPost(data: CreatePostRequest): Promise<Post> {
  return apiRequest<Post>('/api/posts/create', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getFeed(limit = 20, offset = 0): Promise<FeedPost[]> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  return apiRequest<FeedPost[]>(`/api/posts/feed?${params.toString()}`, {
    method: 'GET',
  });
}

export async function getPost(postId: string): Promise<PostDetail> {
  return apiRequest<PostDetail>(`/api/posts/id/${postId}`, { method: 'GET' });
}

export async function updatePost(postId: string, data: UpdatePostRequest): Promise<Post> {
  return apiRequest<Post>(`/api/posts/id/${postId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function deletePost(postId: string): Promise<string> {
  return apiRequest<string>(`/api/posts/id/${postId}`, { method: 'DELETE' });
}

export async function updatePostVisibility(
  postId: string,
  visibility: PostVisibility
): Promise<string> {
  return apiRequest<string>(`/api/posts/id/${postId}/privacy`, {
    method: 'PUT',
    body: JSON.stringify({ visibility }),
  });
}

export async function addPostViewer(postId: string, data: SharePostRequest): Promise<string> {
  return apiRequest<string>(`/api/posts/id/${postId}/privacy`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function removePostViewer(postId: string, data: SharePostRequest): Promise<string> {
  return apiRequest<string>(`/api/posts/id/${postId}/privacy`, {
    method: 'DELETE',
    body: JSON.stringify(data),
  });
}
