import { useState } from 'react';

interface InviteUserModalProps {
  onInvite: (userId: string) => Promise<void>;
  onClose: () => void;
  isLoading: boolean;
}

export function InviteUserModal({
  onInvite,
  onClose,
  isLoading,
}: InviteUserModalProps) {
  const [userId, setUserId] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null);

    if (!userId.trim()) {
      setLocalError('User ID is required');
      return;
    }

    try {
      await onInvite(userId.trim());
      setUserId('');
      onClose();
    } catch (error) {
      if (error instanceof Error) {
        setLocalError(error.message);
      }
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUserId(e.target.value);
    if (localError) setLocalError(null);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Invite User to Group</h3>
          <button className="close-btn" onClick={onClose} disabled={isLoading}>
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          {localError && <div className="error-message">{localError}</div>}

          <div className="form-group">
            <label htmlFor="userId">User ID</label>
            <input
              type="text"
              id="userId"
              value={userId}
              onChange={handleChange}
              placeholder="Enter user ID"
              disabled={isLoading}
              required
            />
            <p className="help-text">
              Enter the user ID of the person you want to invite
            </p>
          </div>

          <div className="form-actions">
            <button
              type="button"
              className="cancel-btn"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancel
            </button>
            <button type="submit" className="submit-btn" disabled={isLoading}>
              {isLoading ? 'Inviting...' : 'Send Invite'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
