import type { GroupMember } from '../../types';

interface MembersListProps {
  members: GroupMember[];
  creatorId: string;
  currentUserId: string | null;
  onAcceptRequest?: (userId: string) => void;
  onRejectRequest?: (userId: string) => void;
  isProcessing?: boolean;
}

export function MembersList({
  members,
  creatorId,
  currentUserId,
  onAcceptRequest,
  onRejectRequest,
  isProcessing = false,
}: MembersListProps) {
  const isCreator = currentUserId === creatorId;

  // Separate members by status
  const activeMembers = members.filter((m) => m.status === 'joined');
  const pendingRequests = members.filter((m) => m.status === 'requested');

  return (
    <div className="members-section">
      <h3>Members ({activeMembers.length})</h3>
      <div className="members-list">
        {activeMembers.length === 0 ? (
          <p className="no-members">No members yet</p>
        ) : (
          activeMembers.map((member) => (
            <div key={member.user_id} className="member-card">
              <div className="member-info">
                <div className="member-avatar">
                  {member.avatar ? (
                    <img src={member.avatar} alt={member.first_name || 'Member'} />
                  ) : (
                    <div className="avatar-placeholder">
                      {member.first_name?.[0] || 'U'}
                    </div>
                  )}
                </div>
                <div className="member-details">
                  <p className="member-name">
                    {member.first_name} {member.last_name}
                  </p>
                  {member.user_id === creatorId && (
                    <span className="creator-badge">Creator</span>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {isCreator && pendingRequests.length > 0 && (
        <div className="pending-requests-section">
          <h3>Pending Requests ({pendingRequests.length})</h3>
          <div className="requests-list">
            {pendingRequests.map((request) => (
              <div key={request.user_id} className="request-card">
                <div className="request-info">
                  <div className="member-avatar">
                    {request.avatar ? (
                      <img src={request.avatar} alt={request.first_name || 'User'} />
                    ) : (
                      <div className="avatar-placeholder">
                        {request.first_name?.[0] || 'U'}
                      </div>
                    )}
                  </div>
                  <div className="member-details">
                    <p className="member-name">
                      {request.first_name} {request.last_name}
                    </p>
                    <p className="request-date">
                      Requested {new Date(request.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                {onAcceptRequest && onRejectRequest && (
                  <div className="request-actions">
                    <button
                      className="accept-btn"
                      onClick={() => onAcceptRequest(request.user_id)}
                      disabled={isProcessing}
                    >
                      Accept
                    </button>
                    <button
                      className="reject-btn"
                      onClick={() => onRejectRequest(request.user_id)}
                      disabled={isProcessing}
                    >
                      Reject
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
