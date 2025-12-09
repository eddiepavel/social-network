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
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <ErrorMessage message={error} />
        <button
          onClick={() => navigate('/dashboard')}
          className="mt-4 rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition"
        >
          Back to Dashboard
        </button>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <p>Profile not found</p>
      </div>
    );
  }

  const targetUserId = profile.user_id;

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto max-w-5xl space-y-6">
        <nav className="flex items-center justify-between rounded-2xl border border-white/10 bg-card/70 px-4 py-3 shadow-soft">
          <div className="flex items-center gap-4 text-sm text-muted">
            <Link to="/dashboard" className="font-semibold text-primary-300 hover:text-primary-200">
              Dashboard
            </Link>
            <span className="text-white/30">/</span>
            <span>Profile</span>
            {isOwnProfile && (
              <>
                <span className="text-white/30">/</span>
                <Link to="/follow-requests" className="font-semibold text-primary-300 hover:text-primary-200">
                  Follow Requests
                </Link>
              </>
            )}
          </div>
          <LogoutButton />
        </nav>

        <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft space-y-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="h-16 w-16 rounded-2xl bg-primary-500/30 flex items-center justify-center text-xl font-bold text-white">
              {profile.first_name[0]}
              {profile.last_name[0]}
            </div>
            <div className="flex-1 space-y-2">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                <div>
                  <h1 className="text-3xl font-semibold">
                    {profile.first_name} {profile.last_name}
                  </h1>
                  {profile.nickname && (
                    <p className="text-sm text-muted">"{profile.nickname}"</p>
                  )}
                </div>
                <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-semibold text-slate-100">
                  {profile.is_public ? '🌍 Public Profile' : '🔒 Private Profile'}
                </span>
              </div>

              <div className="flex gap-4">
                <Link
                  to={`/profile/${targetUserId}/followers`}
                  className="flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm font-semibold text-white hover:border-primary-400/40"
                >
                  <span className="text-lg">{followersCount}</span>
                  <span className="text-muted">Followers</span>
                </Link>
                <Link
                  to={`/profile/${targetUserId}/following`}
                  className="flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm font-semibold text-white hover:border-primary-400/40"
                >
                  <span className="text-lg">{followingCount}</span>
                  <span className="text-muted">Following</span>
                </Link>
              </div>

              {!isOwnProfile && (
                <div className="max-w-xs">
                  <FollowButton
                    userId={targetUserId}
                    isFollowing={isFollowing}
                    onFollowChange={handleFollowChange}
                  />
                </div>
              )}
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <div className="text-xs uppercase tracking-wide text-muted">Email</div>
              <div className="mt-1 text-white">{profile.email}</div>
            </div>
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <div className="text-xs uppercase tracking-wide text-muted">Date of Birth</div>
              <div className="mt-1 text-white">{profile.dob}</div>
            </div>
            {profile.about_me && (
              <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3 sm:col-span-2">
                <div className="text-xs uppercase tracking-wide text-muted">About Me</div>
                <div className="mt-1 text-white">{profile.about_me}</div>
              </div>
            )}
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <div className="text-xs uppercase tracking-wide text-muted">Member Since</div>
              <div className="mt-1 text-white">
                {new Date(profile.created_at).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </div>
            </div>
          </div>

          {isOwnProfile && (
            <div className="flex">
              <button
                onClick={() => navigate('/profile/edit')}
                className="rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40"
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
