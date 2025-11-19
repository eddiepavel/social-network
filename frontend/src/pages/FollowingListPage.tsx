import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { getFollowing } from '../services/followerService';
import type { FollowerUser } from '../types/follower.types';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';

export function FollowingListPage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [following, setFollowing] = useState<FollowerUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!userId) {
      setError('User ID is required');
      setLoading(false);
      return;
    }

    fetchFollowing();
  }, [userId]);

  const fetchFollowing = async () => {
    if (!userId) return;

    try {
      setLoading(true);
      setError(null);
      const data = await getFollowing(userId);
      setFollowing(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load following list');
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
          <h1 className="profile-name">Following</h1>

          {following.length === 0 ? (
            <p className="empty-message">Not following anyone yet</p>
          ) : (
            <div className="users-list">
              {following.map((user) => (
                <div key={user.user_id} className="user-list-item">
                  <Link to={`/profile/${user.user_id}`} className="user-item-link">
                    <div className="user-avatar">
                      {user.first_name[0]}{user.last_name[0]}
                    </div>
                    <div className="user-info">
                      <div className="user-name">
                        {user.first_name} {user.last_name}
                      </div>
                      {user.nickname && (
                        <div className="user-nickname">"{user.nickname}"</div>
                      )}
                      <div className="user-meta">
                        Following since {new Date(user.created_at).toLocaleDateString()}
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
