import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { MembersList } from '../components/groups/MembersList';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { useAuth } from '../hooks/useAuth';
import { getGroupDetails } from '../services/groupsService';
import type { GroupDetails } from '../types';

export function GroupDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [group, setGroup] = useState<GroupDetails | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  if (isLoading) {
    return (
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <LoadingSpinner />
      </div>
    );
  }

  if (!group) {
    return (
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <div className="mx-auto max-w-5xl rounded-2xl border border-white/10 bg-card/70 p-8 text-center shadow-soft">
          <h2 className="text-2xl font-semibold">Group not found</h2>
          <button
            className="mt-4 rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition"
            onClick={() => navigate('/groups')}
          >
            Back to Groups
          </button>
        </div>
      </div>
    );
  }

  const isMember = group.members.some(
    (m) => m.user_id === user?.user_id && m.status === 'joined'
  );

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto flex max-w-5xl flex-col gap-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <button
              className="mb-3 inline-flex items-center gap-2 text-sm text-muted hover:text-white"
              onClick={() => navigate('/groups')}
            >
              ← Back to Groups
            </button>
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold">{group.group_name}</h1>
              <p className="text-slate-200">{group.description}</p>
              <p className="text-sm text-muted">
                Created {new Date(group.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        </div>

        {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

        {isMember && (
          <>
            <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft">
              <MembersList
                members={group.members}
                creatorId={group.creator_id}
                currentUserId={user?.user_id || null}
                onAcceptRequest={undefined}
                onRejectRequest={undefined}
                isProcessing={false}
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
