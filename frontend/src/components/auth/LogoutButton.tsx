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
    <button
      onClick={handleLogout}
      disabled={isLoading}
      className="rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
    >
      {isLoading ? 'Logging out...' : 'Logout'}
    </button>
  );
}
