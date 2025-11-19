import { useState } from 'react';
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

  const handleSubmit = async (e: React.FormEvent) => {
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
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    // Clear errors when user types
    if (localError) setLocalError(null);
  };

  return (
    <form className="group-form" onSubmit={handleSubmit}>
      <h3>Create New Group</h3>

      {localError && <div className="error-message">{localError}</div>}

      <div className="form-group">
        <label htmlFor="group_name">Group Name</label>
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

      <div className="form-group">
        <label htmlFor="description">Description</label>
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

      <div className="form-actions">
        <button
          type="button"
          className="cancel-btn"
          onClick={onCancel}
          disabled={isLoading}
        >
          Cancel
        </button>
        <button type="submit" className="submit-btn" disabled={isLoading}>
          {isLoading ? 'Creating...' : 'Create Group'}
        </button>
      </div>
    </form>
  );
}
