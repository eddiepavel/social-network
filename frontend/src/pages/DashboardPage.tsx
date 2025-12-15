import { Link, useNavigate } from 'react-router-dom';
import { LogoutButton } from '../components/auth/LogoutButton';
import { useAuth } from '../hooks/useAuth';

export function DashboardPage() {
  const { user } = useAuth();
  const navigate = useNavigate();

  if (!user) {
    return <div>Loading...</div>;
  }

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto flex max-w-5xl flex-col gap-6">
        <header className="flex flex-col gap-4 rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-sm uppercase tracking-wide text-muted">Dashboard</p>
              <h1 className="text-3xl font-semibold">Welcome to Social Network</h1>
            </div>
            <div className="flex items-center gap-3">
              <Link
                to={`/profile/${user.user_id}`}
                className="text-sm font-semibold text-primary-300 hover:text-primary-200"
              >
                View Profile
              </Link>
              <Link
                to="/profile/edit"
                className="text-sm font-semibold text-primary-300 hover:text-primary-200"
              >
                Edit Profile
              </Link>
              <LogoutButton />
            </div>
          </div>
          <div>
            <button
              className="rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40"
              onClick={() => navigate('/groups')}
            >
              Browse Groups
            </button>
            <button
              className="ml-3 rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition"
              onClick={() => navigate('/feed')}
            >
              Open Feed
            </button>
          </div>
        </header>

        <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft">
          <h2 className="text-xl font-semibold">Your Profile</h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <p className="text-xs uppercase tracking-wide text-muted">Name</p>
              <p className="mt-1 font-semibold text-white">
                {user.first_name} {user.last_name}
              </p>
            </div>
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <p className="text-xs uppercase tracking-wide text-muted">Email</p>
              <p className="mt-1 font-semibold text-white">{user.email}</p>
            </div>
            {user.nickname && (
              <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
                <p className="text-xs uppercase tracking-wide text-muted">Nickname</p>
                <p className="mt-1 font-semibold text-white">{user.nickname}</p>
              </div>
            )}
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <p className="text-xs uppercase tracking-wide text-muted">Date of Birth</p>
              <p className="mt-1 font-semibold text-white">{user.dob}</p>
            </div>
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <p className="text-xs uppercase tracking-wide text-muted">Profile Type</p>
              <p className="mt-1 font-semibold text-white">
                {user.is_public ? 'Public' : 'Private'}
              </p>
            </div>
            {user.about_me && (
              <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3 sm:col-span-2">
                <p className="text-xs uppercase tracking-wide text-muted">About</p>
                <p className="mt-1 text-slate-100">{user.about_me}</p>
              </div>
            )}
            <div className="rounded-xl border border-white/5 bg-white/5 px-4 py-3">
              <p className="text-xs uppercase tracking-wide text-muted">Member Since</p>
              <p className="mt-1 font-semibold text-white">
                {new Date(user.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
