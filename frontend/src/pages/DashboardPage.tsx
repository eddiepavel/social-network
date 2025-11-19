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
    <div className="page-container">
      <div className="dashboard">
        <header className="dashboard-header">
          <h1>Welcome to Social Network</h1>
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <Link
              to={`/profile/${user.user_id}`}
              style={{ color: '#4f46e5', textDecoration: 'none', fontWeight: 500 }}
            >
              View Profile
            </Link>
            <Link
              to="/profile/edit"
              style={{ color: '#4f46e5', textDecoration: 'none', fontWeight: 500 }}
            >
              Edit Profile
            </Link>
            <LogoutButton />
          </div>
        </header>

        <div className="dashboard-nav">
          <button className="nav-btn" onClick={() => navigate('/groups')}>
            Browse Groups
          </button>
        </div>

        <div className="user-info">
          <h2>Your Profile</h2>
          <div className="info-grid">
            <div className="info-item">
              <strong>Name:</strong> {user.first_name} {user.last_name}
            </div>
            <div className="info-item">
              <strong>Email:</strong> {user.email}
            </div>
            {user.nickname && (
              <div className="info-item">
                <strong>Nickname:</strong> {user.nickname}
              </div>
            )}
            <div className="info-item">
              <strong>Date of Birth:</strong> {user.dob}
            </div>
            <div className="info-item">
              <strong>Profile Type:</strong> {user.is_public ? 'Public' : 'Private'}
            </div>
            {user.about_me && (
              <div className="info-item">
                <strong>About:</strong> {user.about_me}
              </div>
            )}
            <div className="info-item">
              <strong>Member Since:</strong> {new Date(user.created_at).toLocaleDateString()}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
