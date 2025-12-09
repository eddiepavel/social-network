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
      const message = await followUser(userId);
      onFollowChange();
      alert(message || 'Follow action completed');
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
      const message = await unfollowUser(userId);
      onFollowChange();
      alert(message || 'Unfollowed successfully!');
    } catch (err: any) {
      setError(err.message || 'Failed to unfollow user');
      console.error('Unfollow error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-2">
      <button
        onClick={isFollowing ? handleUnfollow : handleFollow}
        disabled={loading}
        className={`w-full rounded-lg border border-white/10 px-4 py-2 font-semibold transition ${
          isFollowing
            ? 'bg-white/5 text-slate-100 hover:bg-white/10'
            : 'bg-primary-600 text-white hover:bg-primary-500 shadow-sm shadow-primary-900/40'
        } disabled:opacity-60 disabled:cursor-not-allowed`}
      >
        {loading ? 'Loading...' : isFollowing ? 'Unfollow' : 'Follow'}
      </button>
      {error && <p className="text-sm text-red-200">{error}</p>}
    </div>
  );
}
