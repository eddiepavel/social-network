// API configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000';

// API endpoints
export const API_ENDPOINTS = {
  REGISTER: '/api/public/register',
  LOGIN: '/api/public/login',
  LOGOUT: '/api/auth/logout',
  SESSION: '/api/auth/session',

  // Groups
  GROUPS: '/api/api/groups',
  GROUP_DETAILS: (id: string) => `/api/api/groups/${id}`,
  GROUP_INVITE: (id: string) => `/api/api/groups/${id}/invite`,
  GROUP_REQUEST: (id: string) => `/api/api/groups/${id}/request`,
  GROUP_ACCEPT: (groupId: string, userId: string) => `/api/api/groups/${groupId}/accept/${userId}`,
} as const;
