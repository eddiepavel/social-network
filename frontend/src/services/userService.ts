import { apiRequest } from './api';
import type { User } from '../types/auth.types';
import type { UpdateProfileRequest, UpdatePrivacyRequest } from '../types/user.types';

// Get user profile by ID
export async function getUserProfile(userId: string): Promise<User> {
  return apiRequest<User>(`/api/users/${userId}`, {
    method: 'GET',
  });
}

// Update current user's profile
export async function updateProfile(data: UpdateProfileRequest): Promise<User> {
  return apiRequest<User>('/api/users/profile', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// Update current user's privacy setting
export async function updatePrivacy(data: UpdatePrivacyRequest): Promise<User> {
  return apiRequest<User>('/api/users/privacy', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}
