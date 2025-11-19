import { useState } from 'react';
import { followUser, unfollowUser } from '../../services/followerService';

interface FollowButtonProps {
  userId: string;
  isFollowing: boolean;
  onFollowChange: () => void;
}

export function FollowButton({ userId, isFollowing, onFollowChange }: FollowButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFollow = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await followUser(userId);
      onFollowChange();

      // Show message based on response
      if (response.status === 'pending') {
        alert('Follow request sent!');
      } else {
        alert('Now following user!');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to follow user');
      console.error('Follow error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleUnfollow = async () => {
    if (!confirm('Are you sure you want to unfollow this user?')) {
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await unfollowUser(userId);
      onFollowChange();
      alert('Unfollowed successfully!');
    } catch (err: any) {
      setError(err.message || 'Failed to unfollow user');
      console.error('Unfollow error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <button
        onClick={isFollowing ? handleUnfollow : handleFollow}
        disabled={loading}
        className={isFollowing ? 'btn-unfollow' : 'btn-follow'}
      >
        {loading ? 'Loading...' : isFollowing ? 'Unfollow' : 'Follow'}
      </button>
      {error && <p className="error-text">{error}</p>}
    </div>
  );
}
