import { useState, FormEvent } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { updateProfile, updatePrivacy } from '../services/userService';
import type { UpdateProfileRequest } from '../types/user.types';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';

export function EditProfilePage() {
  const navigate = useNavigate();
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
      <div className="center-content">
        <p>Please log in to edit your profile</p>
      </div>
    );
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
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
    <div className="edit-profile-container">
      {/* Navigation */}
      <nav className="profile-nav">
        <div className="profile-nav-content">
          <div className="profile-nav-links">
            <Link to="/dashboard" className="profile-nav-link">
              Dashboard
            </Link>
            <Link to={`/profile/${user.user_id}`} className="profile-nav-link">
              View Profile
            </Link>
          </div>
          <LogoutButton />
        </div>
      </nav>

      {/* Edit Profile Content */}
      <div className="edit-profile-content">
        <div className="edit-profile-card">
          <h1 className="edit-profile-title">Edit Profile</h1>

          {error && <ErrorMessage message={error} />}
          {success && <div className="success-message">{success}</div>}

          {/* Profile Update Form */}
          <form onSubmit={handleProfileUpdate} className="profile-section">
            <div className="form-group">
              <label htmlFor="first_name">First Name</label>
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

            <div className="form-group">
              <label htmlFor="last_name">Last Name</label>
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

            <div className="form-group">
              <label htmlFor="nickname">Nickname (Optional)</label>
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

            <div className="form-group">
              <label htmlFor="about_me">About Me (Optional)</label>
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

            <button type="submit" disabled={loading} className="submit-btn">
              {loading ? 'Saving...' : 'Save Changes'}
            </button>
          </form>

          {/* Privacy Settings */}
          <div className="profile-section">
            <h2 className="section-title">Privacy Settings</h2>
            <div className="privacy-toggle-container">
              <div className="privacy-info">
                <h3>Profile Visibility</h3>
                <p>
                  {isPublic
                    ? 'Your profile is visible to everyone'
                    : 'Your profile is only visible to your followers'}
                </p>
              </div>
              <button
                type="button"
                onClick={handlePrivacyToggle}
                disabled={loading}
                className={`toggle-switch ${isPublic ? 'active' : ''}`}
                aria-label="Toggle privacy"
              >
                <span className="toggle-switch-slider" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
