import { apiRequest } from './api';
import { API_ENDPOINTS } from '../utils/constants';
import type {
  Group,
  GroupDetails,
  CreateGroupRequest,
  InviteUserRequest,
  HandleJoinRequestRequest,
} from '../types';

// Get all groups
export async function listGroups(): Promise<Group[]> {
  return apiRequest<Group[]>(API_ENDPOINTS.GROUPS, {
    method: 'GET',
  });
}

// Create a new group
export async function createGroup(data: CreateGroupRequest): Promise<Group> {
  return apiRequest<Group>(API_ENDPOINTS.GROUPS, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// Get group details with members
export async function getGroupDetails(groupId: string): Promise<GroupDetails> {
  return apiRequest<GroupDetails>(API_ENDPOINTS.GROUP_DETAILS(groupId), {
    method: 'GET',
  });
}

// Invite a user to a group
export async function inviteUser(
  groupId: string,
  data: InviteUserRequest
): Promise<{ message: string }> {
  return apiRequest<{ message: string }>(API_ENDPOINTS.GROUP_INVITE(groupId), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// Request to join a group
export async function requestToJoin(
  groupId: string
): Promise<{ message: string }> {
  return apiRequest<{ message: string }>(API_ENDPOINTS.GROUP_REQUEST(groupId), {
    method: 'POST',
  });
}

// Accept or reject a join request (creator only)
export async function handleJoinRequest(
  groupId: string,
  userId: string,
  data: HandleJoinRequestRequest
): Promise<{ message: string }> {
  return apiRequest<{ message: string }>(
    API_ENDPOINTS.GROUP_ACCEPT(groupId, userId),
    {
      method: 'POST',
      body: JSON.stringify(data),
    }
  );
}
