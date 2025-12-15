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
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <ErrorMessage message={error} />
        <button
          onClick={() => navigate(`/profile/${userId}`)}
          className="mt-4 rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition"
        >
          Back to Profile
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto max-w-4xl space-y-6">
        <nav className="flex items-center justify-between rounded-2xl border border-white/10 bg-card/70 px-4 py-3 shadow-soft">
          <div className="flex items-center gap-4 text-sm text-muted">
            <Link to="/dashboard" className="font-semibold text-primary-300 hover:text-primary-200">
              Dashboard
            </Link>
            <span className="text-white/30">/</span>
            <Link to={`/profile/${userId}`} className="font-semibold text-primary-300 hover:text-primary-200">
              Profile
            </Link>
            <span className="text-white/30">/</span>
            <span>Followers</span>
          </div>
          <LogoutButton />
        </nav>

        <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft">
          <h1 className="text-2xl font-semibold mb-4">Followers</h1>

          {followers.length === 0 ? (
            <p className="rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-muted">
              No followers yet
            </p>
          ) : (
            <div className="space-y-3">
              {followers.map((follower) => (
                <div
                  key={follower.user_id}
                  className="flex items-center gap-3 rounded-xl border border-white/10 bg-white/5 px-4 py-3"
                >
                  <Link
                    to={`/profile/${follower.user_id}`}
                    className="flex items-center gap-3 flex-1"
                  >
                    <div className="h-11 w-11 rounded-full bg-primary-500/30 flex items-center justify-center text-white font-semibold">
                      {follower.first_name[0]}
                      {follower.last_name[0]}
                    </div>
                    <div>
                      <div className="font-semibold text-white">
                        {follower.first_name} {follower.last_name}
                      </div>
                      {follower.nickname && (
                        <div className="text-xs text-muted">"{follower.nickname}"</div>
                      )}
                      <div className="text-xs text-muted">
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
