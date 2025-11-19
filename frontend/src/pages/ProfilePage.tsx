import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { getUserProfile } from '../services/userService';
import { getFollowers, getFollowing } from '../services/followerService';
import type { User } from '../types/auth.types';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';
import { FollowButton } from '../components/follow/FollowButton';

export function ProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const [profile, setProfile] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [followersCount, setFollowersCount] = useState(0);
  const [followingCount, setFollowingCount] = useState(0);
  const [isFollowing, setIsFollowing] = useState(false);

  const isOwnProfile = currentUser?.user_id === userId;

  useEffect(() => {
    if (!userId) {
      setError('User ID is required');
      setLoading(false);
      return;
    }

    const fetchProfileData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch profile
        const profileData = await getUserProfile(userId);
        setProfile(profileData);

        // Fetch followers and following counts
        try {
          const [followersData, followingData] = await Promise.all([
            getFollowers(userId),
            getFollowing(userId),
          ]);
          setFollowersCount(followersData.length);
          setFollowingCount(followingData.length);

          // Check if current user is following this profile
          // We need to check if the profile's userId is in the current user's following list
          if (currentUser && !isOwnProfile) {
            const currentUserFollowing = await getFollowing(currentUser.user_id);
            const isCurrentlyFollowing = currentUserFollowing.some(
              (f) => f.user_id === userId
            );
            setIsFollowing(isCurrentlyFollowing);
          }
        } catch (err) {
          // If can't access followers/following, just set to 0
          console.log('Cannot access follower data');
        }
      } catch (err: any) {
        setError(err.message || 'Failed to load profile');
      } finally {
        setLoading(false);
      }
    };

    fetchProfileData();
  }, [userId, currentUser, isOwnProfile]);

  const handleFollowChange = () => {
    // Refresh the page to update follower status
    window.location.reload();
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="center-content">
        <ErrorMessage message={error} />
        <button onClick={() => navigate('/dashboard')} className="btn-secondary">
          Back to Dashboard
        </button>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="center-content">
        <p>Profile not found</p>
      </div>
    );
  }

  return (
    <div className="profile-container">
      {/* Navigation */}
      <nav className="profile-nav">
        <div className="profile-nav-content">
          <div className="profile-nav-links">
            <Link to="/dashboard" className="profile-nav-link">
              Dashboard
            </Link>
            {isOwnProfile && (
              <>
                <Link to="/profile/edit" className="profile-nav-link">
                  Edit Profile
                </Link>
                <Link to="/follow-requests" className="profile-nav-link">
                  Follow Requests
                </Link>
              </>
            )}
          </div>
          <LogoutButton />
        </div>
      </nav>

      {/* Profile Content */}
      <div className="profile-content">
        <div className="profile-card">
          {/* Profile Header */}
          <div className="profile-header">
            <div className="profile-avatar">
              {profile.first_name[0]}{profile.last_name[0]}
            </div>
            <div className="profile-info">
              <h1 className="profile-name">
                {profile.first_name} {profile.last_name}
              </h1>
              {profile.nickname && (
                <p className="profile-nickname">"{profile.nickname}"</p>
              )}
              <div className="profile-meta">
                <span className={`profile-badge ${profile.is_public ? 'badge-public' : 'badge-private'}`}>
                  {profile.is_public ? '🌍 Public Profile' : '🔒 Private Profile'}
                </span>
              </div>

              {/* Follower Stats */}
              <div className="follower-stats">
                <Link to={`/profile/${userId}/followers`} className="stat-item">
                  <span className="stat-count">{followersCount}</span>
                  <span className="stat-label">Followers</span>
                </Link>
                <Link to={`/profile/${userId}/following`} className="stat-item">
                  <span className="stat-count">{followingCount}</span>
                  <span className="stat-label">Following</span>
                </Link>
              </div>

              {/* Follow Button for other users */}
              {!isOwnProfile && (
                <div className="follow-actions">
                  <FollowButton
                    userId={userId}
                    isFollowing={isFollowing}
                    onFollowChange={handleFollowChange}
                  />
                </div>
              )}
            </div>
          </div>

          {/* Profile Details */}
          <div className="profile-details">
            <div className="profile-detail">
              <div className="profile-detail-label">Email</div>
              <div className="profile-detail-value">{profile.email}</div>
            </div>

            <div className="profile-detail">
              <div className="profile-detail-label">Date of Birth</div>
              <div className="profile-detail-value">{profile.dob}</div>
            </div>

            {profile.about_me && (
              <div className="profile-detail">
                <div className="profile-detail-label">About Me</div>
                <div className="profile-detail-value">{profile.about_me}</div>
              </div>
            )}

            <div className="profile-detail">
              <div className="profile-detail-label">Member Since</div>
              <div className="profile-detail-value">
                {new Date(profile.created_at).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </div>
            </div>
          </div>

          {/* Edit Button for Own Profile */}
          {isOwnProfile && (
            <div className="profile-actions">
              <button
                onClick={() => navigate('/profile/edit')}
                className="btn-edit-profile"
              >
                Edit Profile
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
