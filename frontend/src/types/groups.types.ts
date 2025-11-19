// Group-related type definitions

export interface Group {
  group_id: string;
  group_name: string;
  description: string;
  image?: string | null;
  creator_id: string;
  created_at: string;
  member_count?: number;
}

export interface GroupMember {
  user_id: string;
  group_id: string;
  status: 'joined' | 'requested' | 'rejected';
  invited_by?: string | null;
  created_at: string;
  // User details from JOIN
  first_name?: string;
  last_name?: string;
  avatar?: string | null;
}

export interface GroupDetails extends Group {
  members: GroupMember[];
}

export interface CreateGroupRequest {
  group_name: string;
  description: string;
  image?: string;
}

export interface InviteUserRequest {
  user_id: string;
}

export interface HandleJoinRequestRequest {
  action: 'accept' | 'reject';
}

export interface GroupsState {
  groups: Group[];
  currentGroup: GroupDetails | null;
  isLoading: boolean;
  error: string | null;
}
