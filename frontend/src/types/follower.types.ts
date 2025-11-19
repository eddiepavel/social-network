// Follower/Following user
export interface FollowerUser {
  user_id: string;
  first_name: string;
  last_name: string;
  avatar?: string;
  nickname?: string;
  status: 'pending' | 'accepted' | 'rejected';
  created_at: string;
}

// Follow relationship response
export interface FollowResponse {
  follower_id: string;
  followee_id: string;
  status: 'pending' | 'accepted' | 'rejected';
  message: string;
}
