import { useState } from 'react';
import type { ChangeEvent, FormEvent } from 'react';

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

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
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

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setUserId(e.target.value);
    if (localError) setLocalError(null);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-2xl border border-white/10 bg-card/90 p-6 shadow-soft"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-start justify-between">
          <div>
            <p className="text-xs uppercase tracking-wide text-muted">Invite</p>
            <h3 className="text-xl font-semibold text-white">Invite User to Group</h3>
          </div>
          <button
            className="text-lg text-muted hover:text-white transition"
            onClick={onClose}
            disabled={isLoading}
          >
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          {localError && (
            <div className="mb-3 rounded-lg border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-100">
              {localError}
            </div>
          )}

          <div className="space-y-2">
            <label htmlFor="userId" className="text-sm text-muted">
              User ID
            </label>
            <input
              type="text"
              id="userId"
              value={userId}
              onChange={handleChange}
              placeholder="Enter user ID"
              disabled={isLoading}
              required
            />
            <p className="text-xs text-muted">
              Enter the user ID of the person you want to invite
            </p>
          </div>

          <div className="mt-6 flex justify-end gap-3">
            <button
              type="button"
              className="rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40 disabled:opacity-60 disabled:cursor-not-allowed"
              disabled={isLoading}
            >
              {isLoading ? 'Inviting...' : 'Send Invite'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
