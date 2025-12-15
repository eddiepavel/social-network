import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  getFollowRequests,
  acceptFollowRequest,
  rejectFollowRequest,
} from '../services/followerService';
import type { FollowRequest } from '../types/follower.types';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';

export function FollowRequestsPage() {
  const [requests, setRequests] = useState<FollowRequest[]>([]);
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
      setRequests(Array.isArray(data) ? data : []);
    } catch (err: any) {
      setError(err.message || 'Failed to load follow requests');
    } finally {
      setLoading(false);
    }
  };

  const handleAccept = async (requestId: number) => {
    setActionLoading(String(requestId));
    try {
      await acceptFollowRequest(requestId);
      // Remove from list
      setRequests(requests.filter((r) => r.id !== requestId));
      alert('Follow request accepted!');
    } catch (err: any) {
      alert(err.message || 'Failed to accept request');
    } finally {
      setActionLoading(null);
    }
  };

  const handleReject = async (requestId: number) => {
    setActionLoading(String(requestId));
    try {
      await rejectFollowRequest(requestId);
      // Remove from list
      setRequests(requests.filter((r) => r.id !== requestId));
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
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto max-w-4xl space-y-6">
        <nav className="flex items-center justify-between rounded-2xl border border-white/10 bg-card/70 px-4 py-3 shadow-soft">
          <div className="flex items-center gap-4 text-sm text-muted">
            <Link to="/dashboard" className="font-semibold text-primary-300 hover:text-primary-200">
              Dashboard
            </Link>
            <span className="text-white/30">/</span>
            <span>Follow Requests</span>
          </div>
          <LogoutButton />
        </nav>

        <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft">
          <h1 className="text-2xl font-semibold mb-4">Follow Requests</h1>

          {error && <ErrorMessage message={error} />}

          {requests.length === 0 ? (
            <p className="rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-muted">
              No pending follow requests
            </p>
          ) : (
            <div className="space-y-3">
              {requests.map((request) => (
                <div
                  key={request.id}
                  className="flex items-center justify-between rounded-xl border border-white/10 bg-white/5 px-4 py-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="h-11 w-11 rounded-full bg-primary-500/30 flex items-center justify-center text-white font-semibold">
                      {request.follower_name?.[0] ?? '?'}
                    </div>
                    <div>
                      <Link
                        to={`/profile/${request.follower_id}`}
                        className="font-semibold text-white hover:text-primary-200"
                      >
                        {request.follower_name || request.follower_id}
                      </Link>
                      <p className="text-xs text-muted">
                        {new Date(request.created_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleAccept(request.id)}
                      disabled={actionLoading === String(request.id)}
                      className="rounded-lg bg-primary-600 px-3 py-2 text-xs font-semibold text-white hover:bg-primary-500 transition disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                      {actionLoading === String(request.id) ? 'Loading...' : 'Accept'}
                    </button>
                    <button
                      onClick={() => handleReject(request.id)}
                      disabled={actionLoading === String(request.id)}
                      className="rounded-lg border border-white/10 px-3 py-2 text-xs font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                      {actionLoading === String(request.id) ? 'Loading...' : 'Reject'}
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
