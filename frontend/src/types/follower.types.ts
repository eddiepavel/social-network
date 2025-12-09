// Follower/Following user
export interface FollowerUser {
  user_id: string;
  first_name: string;
  last_name: string;
  avatar?: string;
  nickname?: string;
  created_at: string;
}

// Pending follow request
export interface FollowRequest {
  id: number;
  follower_id: string;
  follower_name: string;
  created_at: string;
}
