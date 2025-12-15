import { apiRequest } from './api';
import { API_ENDPOINTS } from '../utils/constants';
import type { Group, GroupDetails, CreateGroupRequest } from '../types';

// Get all groups
export async function listGroups(): Promise<Group[]> {
  return apiRequest<Group[]>(API_ENDPOINTS.GROUPS, {
    method: 'GET',
          credentials: 'include',
  });
}

// Create a new group
export async function createGroup(data: CreateGroupRequest): Promise<Group> {
  return apiRequest<Group>(API_ENDPOINTS.GROUP_CREATE, {
    method: 'POST',
    body: JSON.stringify(data),
          credentials: 'include',
  });
}

// Get group details with members
export async function getGroupDetails(groupId: string): Promise<GroupDetails> {
  return apiRequest<GroupDetails>(API_ENDPOINTS.GROUP_DETAILS(groupId), {
    method: 'GET',
          credentials: 'include',
  });
}

// The current backend only supports listing, creating, and fetching group details.
// Invite/join request flows are not implemented server-side yet.
export async function inviteUser(): Promise<never> {
  throw new Error('Group invitations are not supported by the backend yet.');
}

export async function requestToJoin(): Promise<never> {
  throw new Error('Join requests are not supported by the backend yet.');
}

export async function handleJoinRequest(): Promise<never> {
  throw new Error('Managing join requests is not supported by the backend yet.');
}
