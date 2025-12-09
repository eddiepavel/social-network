import { apiRequest } from './api';
import type { FollowerUser, FollowRequest } from '../types/follower.types';

// Follow a user
export async function followUser(userId: string): Promise<string> {
  return apiRequest<string>(`/api/followers/${userId}/follow`, {
    method: 'POST',
  });
}

// Unfollow a user
export async function unfollowUser(userId: string): Promise<string> {
  return apiRequest<string>(`/api/followers/${userId}/follow`, {
    method: 'POST',
  });
}

// Accept follow request
export async function acceptFollowRequest(requestId: number): Promise<string> {
  return apiRequest<string>(`/api/followers/requests/${requestId}/respond`, {
    method: 'POST',
    body: JSON.stringify({ status: 'accepted' }),
  });
}

// Reject follow request
export async function rejectFollowRequest(requestId: number): Promise<string> {
  return apiRequest<string>(`/api/followers/requests/${requestId}/respond`, {
    method: 'POST',
    body: JSON.stringify({ status: 'rejected' }),
  });
}

// Get followers list
export async function getFollowers(userId: string): Promise<FollowerUser[]> {
  return apiRequest<FollowerUser[]>(`/api/followers/${userId}/followers`, {
    method: 'GET',
  });
}

// Get following list
export async function getFollowing(userId: string): Promise<FollowerUser[]> {
  return apiRequest<FollowerUser[]>(`/api/followers/${userId}/following`, {
    method: 'GET',
  });
}

// Get pending follow requests
export async function getFollowRequests(): Promise<FollowRequest[]> {
  return apiRequest<FollowRequest[]>('/api/followers/requests', {
    method: 'GET',
  });
}
