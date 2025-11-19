import { apiRequest } from './api';
import { API_ENDPOINTS } from '../utils/constants';
import type { User, LoginFormData, RegisterFormData } from '../types';

/**
 * Authentication Service
 * Handles all authentication-related API calls using native Fetch
 */

// Register a new user
export async function register(formData: RegisterFormData): Promise<User> {
  return apiRequest<User>(API_ENDPOINTS.REGISTER, {
    method: 'POST',
    body: JSON.stringify(formData),
  });
}

// Login with email and password
export async function login(credentials: LoginFormData): Promise<User> {
  return apiRequest<User>(API_ENDPOINTS.LOGIN, {
    method: 'POST',
    body: JSON.stringify(credentials),
  });
}

// Logout current user
export async function logout(): Promise<void> {
  return apiRequest<void>(API_ENDPOINTS.LOGOUT, {
    method: 'POST',
  });
}

// Check current session and get user data
export async function getSession(): Promise<User> {
  return apiRequest<User>(API_ENDPOINTS.SESSION, {
    method: 'GET',
  });
}
