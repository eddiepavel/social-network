import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { getFollowers } from '../services/followerService';
import type { FollowerUser } from '../types/follower.types';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';

export function FollowersListPage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [followers, setFollowers] = useState<FollowerUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!userId) {
      setError('User ID is required');
      setLoading(false);
      return;
    }

    fetchFollowers();
  }, [userId]);

  const fetchFollowers = async () => {
    if (!userId) return;

    try {
      setLoading(true);
      setError(null);
      const data = await getFollowers(userId);
      setFollowers(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load followers');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="center-content">
        <ErrorMessage message={error} />
        <button onClick={() => navigate(`/profile/${userId}`)} className="btn-secondary">
          Back to Profile
        </button>
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
            <Link to={`/profile/${userId}`} className="profile-nav-link">
              Back to Profile
            </Link>
          </div>
          <LogoutButton />
        </div>
      </nav>

      {/* Content */}
      <div className="profile-content">
        <div className="profile-card">
          <h1 className="profile-name">Followers</h1>

          {followers.length === 0 ? (
            <p className="empty-message">No followers yet</p>
          ) : (
            <div className="users-list">
              {followers.map((follower) => (
                <div key={follower.user_id} className="user-list-item">
                  <Link to={`/profile/${follower.user_id}`} className="user-item-link">
                    <div className="user-avatar">
                      {follower.first_name[0]}{follower.last_name[0]}
                    </div>
                    <div className="user-info">
                      <div className="user-name">
                        {follower.first_name} {follower.last_name}
                      </div>
                      {follower.nickname && (
                        <div className="user-nickname">"{follower.nickname}"</div>
                      )}
                      <div className="user-meta">
                        Following since {new Date(follower.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  </Link>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
