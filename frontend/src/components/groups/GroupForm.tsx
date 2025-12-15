import { useState } from 'react';
import type { ChangeEvent, FormEvent } from 'react';
import type { CreateGroupRequest } from '../../types';

interface GroupFormProps {
  onSubmit: (data: CreateGroupRequest) => Promise<void>;
  onCancel: () => void;
  isLoading: boolean;
}

export function GroupForm({ onSubmit, onCancel, isLoading }: GroupFormProps) {
  const [formData, setFormData] = useState<CreateGroupRequest>({
    group_name: '',
    description: '',
  });
  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLocalError(null);

    // Validation
    if (!formData.group_name.trim() || !formData.description.trim()) {
      setLocalError('Group name and description are required');
      return;
    }

    try {
      await onSubmit(formData);
      // Reset form on success
      setFormData({ group_name: '', description: '' });
    } catch (error) {
      if (error instanceof Error) {
        setLocalError(error.message);
      }
    }
  };

  const handleChange = (
    e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear errors when user types
    if (localError) setLocalError(null);
  };

  return (
    <form
      className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft"
      onSubmit={handleSubmit}
    >
      <div className="mb-4">
        <h3 className="text-xl font-semibold text-white">Create New Group</h3>
        <p className="text-sm text-muted mt-1">
          Spin up a space for your community.
        </p>
      </div>

      {localError && (
        <div className="mb-4 rounded-lg border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-100">
          {localError}
        </div>
      )}

      <div className="space-y-2 mb-4">
        <label htmlFor="group_name" className="text-sm text-muted">
          Group Name
        </label>
        <input
          type="text"
          id="group_name"
          name="group_name"
          value={formData.group_name}
          onChange={handleChange}
          required
          disabled={isLoading}
          placeholder="Enter group name"
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="description" className="text-sm text-muted">
          Description
        </label>
        <textarea
          id="description"
          name="description"
          value={formData.description}
          onChange={handleChange}
          required
          disabled={isLoading}
          placeholder="Describe your group"
          rows={4}
        />
      </div>

      <div className="mt-6 flex justify-end gap-3">
        <button
          type="button"
          className="rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
          onClick={onCancel}
          disabled={isLoading}
        >
          Cancel
        </button>
        <button
          type="submit"
          className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40 disabled:opacity-60 disabled:cursor-not-allowed"
          disabled={isLoading}
        >
          {isLoading ? 'Creating...' : 'Create Group'}
        </button>
      </div>
    </form>
  );
}
