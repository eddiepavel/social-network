import { useState } from 'react';
import type { CreatePostRequest, PostVisibility } from '../../types';
import { ErrorMessage } from '../common/ErrorMessage';

interface PostFormProps {
  onSubmit: (data: CreatePostRequest) => Promise<void>;
  isSubmitting?: boolean;
}

const visibilities: PostVisibility[] = ['public', 'semi-private', 'private'];

export function PostForm({ onSubmit, isSubmitting = false }: PostFormProps) {
  const [content, setContent] = useState('');
  const [visibility, setVisibility] = useState<PostVisibility>('public');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!content.trim()) {
      setError('Content is required');
      return;
    }
    try {
      await onSubmit({ content: content.trim(), visibility });
      setContent('');
      setVisibility('public');
    } catch (err: any) {
      setError(err.message || 'Failed to create post');
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft space-y-4"
    >
      <div>
        <h2 className="text-xl font-semibold text-white">Create Post</h2>
        <p className="text-sm text-muted">Share an update with your network.</p>
      </div>

      {error && <ErrorMessage message={error} />}

      <div className="space-y-2">
        <label className="text-sm text-muted">Content</label>
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="What's on your mind?"
          rows={4}
          disabled={isSubmitting}
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm text-muted">Visibility</label>
        <select
          value={visibility}
          onChange={(e) => setVisibility(e.target.value as PostVisibility)}
          disabled={isSubmitting}
          className="bg-white/5"
        >
          {visibilities.map((v) => (
            <option key={v} value={v}>
              {v === 'semi-private' ? 'Semi-private' : v.charAt(0).toUpperCase() + v.slice(1)}
            </option>
          ))}
        </select>
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full rounded-xl bg-primary-600 px-4 py-3 text-sm font-semibold text-white hover:bg-primary-500 transition shadow-sm shadow-primary-900/40 disabled:opacity-60 disabled:cursor-not-allowed"
      >
        {isSubmitting ? 'Posting...' : 'Post'}
      </button>
    </form>
  );
}
