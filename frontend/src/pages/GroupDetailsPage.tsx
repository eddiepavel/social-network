import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { MembersList } from '../components/groups/MembersList';
import { InviteUserModal } from '../components/groups/InviteUserModal';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { useAuth } from '../hooks/useAuth';
import {
  getGroupDetails,
  requestToJoin,
  inviteUser,
  handleJoinRequest,
} from '../services/groupsService';
import type { GroupDetails } from '../types';

export function GroupDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [group, setGroup] = useState<GroupDetails | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      loadGroupDetails();
    }
  }, [id]);

  const loadGroupDetails = async () => {
    if (!id) return;

    try {
      setIsLoading(true);
      setError(null);
      const data = await getGroupDetails(id);
      setGroup(data);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to load group details');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleRequestToJoin = async () => {
    if (!id) return;

    try {
      setIsProcessing(true);
      setError(null);
      await requestToJoin(id);
      setSuccessMessage('Join request sent successfully!');
      // Reload group details to see updated status
      await loadGroupDetails();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to send join request');
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleInviteUser = async (userId: string) => {
    if (!id) return;

    try {
      setIsProcessing(true);
      setError(null);
      await inviteUser(id, { user_id: userId });
      setSuccessMessage('User invited successfully!');
      // Reload to see updated members
      await loadGroupDetails();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to invite user');
      }
      throw err;
    } finally {
      setIsProcessing(false);
    }
  };

  const handleAcceptRequest = async (userId: string) => {
    if (!id) return;

    try {
      setIsProcessing(true);
      setError(null);
      await handleJoinRequest(id, userId, { action: 'accept' });
      setSuccessMessage('Join request accepted!');
      await loadGroupDetails();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to accept request');
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleRejectRequest = async (userId: string) => {
    if (!id) return;

    try {
      setIsProcessing(true);
      setError(null);
      await handleJoinRequest(id, userId, { action: 'reject' });
      setSuccessMessage('Join request rejected!');
      await loadGroupDetails();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to reject request');
      }
    } finally {
      setIsProcessing(false);
    }
  };

  if (isLoading) {
    return (
      <div className="page-container">
        <LoadingSpinner />
      </div>
    );
  }

  if (!group) {
    return (
      <div className="page-container">
        <div className="error-state">
          <h2>Group not found</h2>
          <button className="submit-btn" onClick={() => navigate('/groups')}>
            Back to Groups
          </button>
        </div>
      </div>
    );
  }

  const isMember = group.members.some(
    (m) => m.user_id === user?.user_id && m.status === 'joined'
  );
  const isCreator = group.creator_id === user?.user_id;
  const hasPendingRequest = group.members.some(
    (m) => m.user_id === user?.user_id && m.status === 'requested'
  );

  return (
    <div className="page-container">
      <div className="group-details-page">
        <div className="group-header">
          <button className="back-btn" onClick={() => navigate('/groups')}>
            ← Back to Groups
          </button>
          <div className="group-title-section">
            <h1>{group.group_name}</h1>
            <p className="group-description-large">{group.description}</p>
            <p className="group-meta">
              Created {new Date(group.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>

        {error && (
          <ErrorMessage message={error} onClose={() => setError(null)} />
        )}

        {successMessage && (
          <div className="success-message">
            {successMessage}
            <button
              className="close-btn"
              onClick={() => setSuccessMessage(null)}
            >
              ×
            </button>
          </div>
        )}

        {!isMember && !hasPendingRequest && (
          <div className="group-actions">
            <button
              className="submit-btn"
              onClick={handleRequestToJoin}
              disabled={isProcessing}
            >
              {isProcessing ? 'Sending...' : 'Request to Join'}
            </button>
          </div>
        )}

        {hasPendingRequest && !isMember && (
          <div className="pending-status">
            <p>Your join request is pending approval</p>
          </div>
        )}

        {isMember && (
          <>
            <div className="member-actions">
              <button
                className="submit-btn"
                onClick={() => setShowInviteModal(true)}
                disabled={isProcessing}
              >
                Invite User
              </button>
            </div>

            <MembersList
              members={group.members}
              creatorId={group.creator_id}
              currentUserId={user?.user_id || null}
              onAcceptRequest={isCreator ? handleAcceptRequest : undefined}
              onRejectRequest={isCreator ? handleRejectRequest : undefined}
              isProcessing={isProcessing}
            />
          </>
        )}

        {showInviteModal && (
          <InviteUserModal
            onInvite={handleInviteUser}
            onClose={() => setShowInviteModal(false)}
            isLoading={isProcessing}
          />
        )}
      </div>
    </div>
  );
}
