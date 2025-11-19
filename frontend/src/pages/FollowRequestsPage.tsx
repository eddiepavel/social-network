import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  getFollowRequests,
  acceptFollowRequest,
  rejectFollowRequest,
} from '../services/followerService';
import type { FollowerUser } from '../types/follower.types';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';

export function FollowRequestsPage() {
  const [requests, setRequests] = useState<FollowerUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    fetchRequests();
  }, []);

  const fetchRequests = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getFollowRequests();
      setRequests(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load follow requests');
    } finally {
      setLoading(false);
    }
  };

  const handleAccept = async (userId: string) => {
    setActionLoading(userId);
    try {
      await acceptFollowRequest(userId);
      // Remove from list
      setRequests(requests.filter((r) => r.user_id !== userId));
      alert('Follow request accepted!');
    } catch (err: any) {
      alert(err.message || 'Failed to accept request');
    } finally {
      setActionLoading(null);
    }
  };

  const handleReject = async (userId: string) => {
    setActionLoading(userId);
    try {
      await rejectFollowRequest(userId);
      // Remove from list
      setRequests(requests.filter((r) => r.user_id !== userId));
      alert('Follow request rejected');
    } catch (err: any) {
      alert(err.message || 'Failed to reject request');
    } finally {
      setActionLoading(null);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="profile-container">
      {/* Navigation */}
      <nav className="profile-nav">
        <div className="profile-nav-content">
          <div className="profile-nav-links">
            <Link to="/dashboard" className="profile-nav-link">
              Dashboard
            </Link>
          </div>
          <LogoutButton />
        </div>
      </nav>

      {/* Content */}
      <div className="profile-content">
        <div className="profile-card">
          <h1 className="profile-name">Follow Requests</h1>

          {error && <ErrorMessage message={error} />}

          {requests.length === 0 ? (
            <p className="empty-message">No pending follow requests</p>
          ) : (
            <div className="follow-requests-list">
              {requests.map((request) => (
                <div key={request.user_id} className="follow-request-item">
                  <div className="request-user-info">
                    <div className="request-avatar">
                      {request.first_name[0]}{request.last_name[0]}
                    </div>
                    <div>
                      <Link
                        to={`/profile/${request.user_id}`}
                        className="request-user-name"
                      >
                        {request.first_name} {request.last_name}
                      </Link>
                      {request.nickname && (
                        <p className="request-nickname">"{request.nickname}"</p>
                      )}
                      <p className="request-time">
                        {new Date(request.created_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <div className="request-actions">
                    <button
                      onClick={() => handleAccept(request.user_id)}
                      disabled={actionLoading === request.user_id}
                      className="btn-accept"
                    >
                      {actionLoading === request.user_id ? 'Loading...' : 'Accept'}
                    </button>
                    <button
                      onClick={() => handleReject(request.user_id)}
                      disabled={actionLoading === request.user_id}
                      className="btn-reject"
                    >
                      {actionLoading === request.user_id ? 'Loading...' : 'Reject'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
