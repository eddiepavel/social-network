// User profile update request
export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  nickname?: string;
  about_me?: string;
  avatar?: string;
}

// Privacy update request
export interface UpdatePrivacyRequest {
  is_public: boolean;
}
