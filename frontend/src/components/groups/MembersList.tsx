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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-white">
          Members ({activeMembers.length})
        </h3>
      </div>
      <div className="grid gap-4">
        {activeMembers.length === 0 ? (
          <p className="rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-muted">
            No members yet
          </p>
        ) : (
          activeMembers.map((member) => (
            <div
              key={member.user_id}
              className="flex items-center justify-between rounded-xl border border-white/10 bg-white/5 px-4 py-3"
            >
              <div className="flex items-center gap-3">
                <div className="h-11 w-11 overflow-hidden rounded-full bg-primary-500/30 flex items-center justify-center text-white font-semibold">
                  {member.avatar ? (
                    <img src={member.avatar} alt={member.first_name || 'Member'} />
                  ) : (
                    <span>{member.first_name?.[0] || 'U'}</span>
                  )}
                </div>
                <div>
                  <p className="font-semibold text-white">
                    {member.first_name} {member.last_name}
                  </p>
                  {member.user_id === creatorId && (
                    <span className="ml-2 rounded-full bg-primary-600/30 px-3 py-1 text-xs text-primary-100">
                      Creator
                    </span>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {isCreator && pendingRequests.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-white">
            Pending Requests ({pendingRequests.length})
          </h3>
          <div className="space-y-3">
            {pendingRequests.map((request) => (
              <div
                key={request.user_id}
                className="flex items-center justify-between rounded-xl border border-white/10 bg-white/5 px-4 py-3"
              >
                <div className="flex items-center gap-3">
                  <div className="h-11 w-11 overflow-hidden rounded-full bg-primary-500/30 flex items-center justify-center text-white font-semibold">
                    {request.avatar ? (
                      <img src={request.avatar} alt={request.first_name || 'User'} />
                    ) : (
                      <span>{request.first_name?.[0] || 'U'}</span>
                    )}
                  </div>
                  <div>
                    <p className="font-semibold text-white">
                      {request.first_name} {request.last_name}
                    </p>
                    <p className="text-xs text-muted">
                      Requested {new Date(request.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                {onAcceptRequest && onRejectRequest && (
                  <div className="flex gap-2">
                    <button
                      className="rounded-lg bg-primary-600 px-3 py-2 text-xs font-semibold text-white hover:bg-primary-500 transition disabled:opacity-60 disabled:cursor-not-allowed"
                      onClick={() => onAcceptRequest(request.user_id)}
                      disabled={isProcessing}
                    >
                      Accept
                    </button>
                    <button
                      className="rounded-lg border border-white/10 px-3 py-2 text-xs font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
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
