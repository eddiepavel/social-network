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
      setGroups(data.data);
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
      <div className="page-container">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="groups-page">
        <div className="page-header">
          <h1>Groups</h1>
          {!showCreateForm && (
            <button
              className="create-group-btn"
              onClick={() => setShowCreateForm(true)}
            >
              Create Group
            </button>
          )}
        </div>

        {error && (
          <ErrorMessage message={error} onClose={() => setError(null)} />
        )}

        {showCreateForm && (
          <div className="create-group-section">
            <GroupForm
              onSubmit={handleCreateGroup}
              onCancel={handleCancelCreate}
              isLoading={isCreating}
            />
          </div>
        )}

        <div className="groups-container">
          {!groups ? (
            <div className="no-groups">
              <p>No groups found.</p>
              {!showCreateForm && (
                <button
                  className="submit-btn"
                  onClick={() => setShowCreateForm(true)}
                >
                  Create your first group
                </button>
              )}
            </div>
          ) : (
            <div className="groups-grid">
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
