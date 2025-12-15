import { useState, useEffect } from 'react';
import { GroupCard } from '../components/groups/GroupCard';
import { GroupForm } from '../components/groups/GroupForm';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { listGroups, createGroup } from '../services/groupsService';
import type { Group, CreateGroupRequest } from '../types';

export function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await listGroups();
      setGroups(data);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to load groups');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateGroup = async (data: CreateGroupRequest) => {
    try {
      setIsCreating(true);
      setError(null);
      const newGroup = await createGroup(data);
      setGroups((prev) => [newGroup, ...prev]);
      setShowCreateForm(false);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to create group');
      }
      throw err; // Re-throw so GroupForm can handle it
    } finally {
      setIsCreating(false);
    }
  };

  const handleCancelCreate = () => {
    setShowCreateForm(false);
    setError(null);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto flex max-w-6xl flex-col gap-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm uppercase tracking-wide text-muted">Discover</p>
            <h1 className="text-3xl font-semibold">Groups</h1>
          </div>
          {!showCreateForm && (
            <button
              className="rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40"
              onClick={() => setShowCreateForm(true)}
            >
              Create Group
            </button>
          )}
        </div>

        {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

        {showCreateForm && (
          <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft">
            <GroupForm
              onSubmit={handleCreateGroup}
              onCancel={handleCancelCreate}
              isLoading={isCreating}
            />
          </div>
        )}

        <div>
          {!groups ? (
            <div className="rounded-2xl border border-white/10 bg-card/70 p-6 text-center text-muted">
              <p>No groups found.</p>
              {!showCreateForm && (
                <button
                  className="mt-4 rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition"
                  onClick={() => setShowCreateForm(true)}
                >
                  Create your first group
                </button>
              )}
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {groups.map((group) => (
                <GroupCard key={group.group_id} group={group} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
