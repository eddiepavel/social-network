// API configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// API endpoints
export const API_ENDPOINTS = {
  REGISTER: '/api/auth/register',
  LOGIN: '/api/auth/login',
  LOGOUT: '/api/auth/logout',
  SESSION: '/api/auth/session',

  // Groups
  GROUPS: '/api/groups',
  GROUP_DETAILS: (id: string) => `/api/groups/${id}`,
  GROUP_INVITE: (id: string) => `/api/groups/${id}/invite`,
  GROUP_REQUEST: (id: string) => `/api/groups/${id}/request`,
  GROUP_ACCEPT: (groupId: string, userId: string) => `/api/groups/${groupId}/accept/${userId}`,
} as const;
