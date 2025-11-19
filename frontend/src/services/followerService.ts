import { apiRequest } from './api';
import type { FollowerUser, FollowResponse } from '../types/follower.types';

// Follow a user
export async function followUser(userId: string): Promise<FollowResponse> {
  return apiRequest<FollowResponse>(`/api/follow/${userId}`, {
    method: 'POST',
  });
}

// Unfollow a user
export async function unfollowUser(userId: string): Promise<string> {
  return apiRequest<string>(`/api/follow/${userId}`, {
    method: 'DELETE',
  });
}

// Accept follow request
export async function acceptFollowRequest(followerId: string): Promise<string> {
  return apiRequest<string>(`/api/follow/accept/${followerId}`, {
    method: 'POST',
  });
}

// Reject follow request
export async function rejectFollowRequest(followerId: string): Promise<string> {
  return apiRequest<string>(`/api/follow/reject/${followerId}`, {
    method: 'POST',
  });
}

// Get followers list
export async function getFollowers(userId: string): Promise<FollowerUser[]> {
  return apiRequest<FollowerUser[]>(`/api/followers/${userId}`, {
    method: 'GET',
  });
}

// Get following list
export async function getFollowing(userId: string): Promise<FollowerUser[]> {
  return apiRequest<FollowerUser[]>(`/api/following/${userId}`, {
    method: 'GET',
  });
}

// Get pending follow requests
export async function getFollowRequests(): Promise<FollowerUser[]> {
  return apiRequest<FollowerUser[]>('/api/follow/requests', {
    method: 'GET',
  });
}
