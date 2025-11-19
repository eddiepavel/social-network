import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

export function LogoutButton() {
  const navigate = useNavigate();
  const { logout, isLoading } = useAuth();

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout error:', error);
      // Still navigate to login even if API fails
      navigate('/login');
    }
  };

  return (
    <button onClick={handleLogout} disabled={isLoading} className="logout-btn">
      {isLoading ? 'Logging out...' : 'Logout'}
    </button>
  );
}
