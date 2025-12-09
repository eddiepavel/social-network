import { useState } from 'react';
import type { ChangeEvent, FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { updateProfile, updatePrivacy } from '../services/userService';
import type { UpdateProfileRequest } from '../types/user.types';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';

export function EditProfilePage() {
  const { user, setUser } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    nickname: user?.nickname || '',
    about_me: user?.about_me || '',
  });

  const [isPublic, setIsPublic] = useState(user?.is_public || false);

  if (!user) {
    return (
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <p>Please log in to edit your profile</p>
      </div>
    );
  }

  const handleInputChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleProfileUpdate = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // Only send fields that have changed
      const updates: UpdateProfileRequest = {};
      if (formData.first_name !== user.first_name) updates.first_name = formData.first_name;
      if (formData.last_name !== user.last_name) updates.last_name = formData.last_name;
      if (formData.nickname !== (user.nickname || '')) updates.nickname = formData.nickname;
      if (formData.about_me !== (user.about_me || '')) updates.about_me = formData.about_me;

      if (Object.keys(updates).length === 0) {
        setError('No changes to save');
        setLoading(false);
        return;
      }

      const updatedUser = await updateProfile(updates);
      setUser(updatedUser);
      setSuccess('Profile updated successfully!');
    } catch (err: any) {
      setError(err.message || 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  const handlePrivacyToggle = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const updatedUser = await updatePrivacy({ is_public: !isPublic });
      setUser(updatedUser);
      setIsPublic(updatedUser.is_public);
      setSuccess(`Profile is now ${updatedUser.is_public ? 'public' : 'private'}!`);
    } catch (err: any) {
      setError(err.message || 'Failed to update privacy setting');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto max-w-4xl space-y-6">
        <nav className="flex items-center justify-between rounded-2xl border border-white/10 bg-card/70 px-4 py-3 shadow-soft">
          <div className="flex items-center gap-4 text-sm text-muted">
            <Link to="/dashboard" className="font-semibold text-primary-300 hover:text-primary-200">
              Dashboard
            </Link>
            <span className="text-white/30">/</span>
            <Link to={`/profile/${user.user_id}`} className="font-semibold text-primary-300 hover:text-primary-200">
              Profile
            </Link>
            <span className="text-white/30">/</span>
            <span>Edit</span>
          </div>
          <LogoutButton />
        </nav>

        <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft space-y-6">
          <h1 className="text-3xl font-semibold">Edit Profile</h1>

          {error && <ErrorMessage message={error} />}
          {success && (
            <div className="rounded-xl border border-primary-500/40 bg-primary-500/10 px-4 py-3 text-sm text-primary-100">
              {success}
            </div>
          )}

          <form onSubmit={handleProfileUpdate} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label htmlFor="first_name" className="text-sm text-muted">
                  First Name
                </label>
                <input
                  type="text"
                  id="first_name"
                  name="first_name"
                  value={formData.first_name}
                  onChange={handleInputChange}
                  disabled={loading}
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="last_name" className="text-sm text-muted">
                  Last Name
                </label>
                <input
                  type="text"
                  id="last_name"
                  name="last_name"
                  value={formData.last_name}
                  onChange={handleInputChange}
                  disabled={loading}
                  required
                />
              </div>
            </div>

            <div className="space-y-2">
              <label htmlFor="nickname" className="text-sm text-muted">
                Nickname (Optional)
              </label>
              <input
                type="text"
                id="nickname"
                name="nickname"
                value={formData.nickname}
                onChange={handleInputChange}
                disabled={loading}
                placeholder="Your nickname"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="about_me" className="text-sm text-muted">
                About Me (Optional)
              </label>
              <textarea
                id="about_me"
                name="about_me"
                value={formData.about_me}
                onChange={handleInputChange}
                disabled={loading}
                rows={4}
                placeholder="Tell us about yourself..."
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-xl bg-primary-600 px-4 py-3 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40 disabled:opacity-60 disabled:cursor-not-allowed"
            >
              {loading ? 'Saving...' : 'Save Changes'}
            </button>
          </form>

          <div className="space-y-3 rounded-xl border border-white/10 bg-white/5 px-4 py-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Privacy Settings</h2>
                <p className="text-sm text-muted">
                  {isPublic
                    ? 'Your profile is visible to everyone'
                    : 'Your profile is only visible to your followers'}
                </p>
              </div>
              <button
                type="button"
                onClick={handlePrivacyToggle}
                disabled={loading}
                className={`relative inline-flex h-8 w-14 items-center rounded-full transition ${
                  isPublic ? 'bg-primary-600' : 'bg-white/10'
                } disabled:opacity-60 disabled:cursor-not-allowed`}
                aria-label="Toggle privacy"
              >
                <span
                  className={`absolute left-1 h-6 w-6 rounded-full bg-white shadow transition-transform ${
                    isPublic ? 'translate-x-6' : 'translate-x-0'
                  }`}
                />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
