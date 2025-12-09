import { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';
import {
  deletePost,
  getPost,
  updatePost,
  updatePostVisibility,
  addPostViewer,
  removePostViewer,
} from '../services/postsService';
import type { PostDetail, PostVisibility } from '../types';

export function PostDetailsPage() {
  const { postId } = useParams<{ postId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [post, setPost] = useState<PostDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editContent, setEditContent] = useState('');
  const [editVisibility, setEditVisibility] = useState<PostVisibility>('public');
  const [shareUserId, setShareUserId] = useState('');
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    if (postId) {
      loadPost(postId);
    }
  }, [postId]);

  const loadPost = async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      const data = await getPost(id);
      setPost(data);
      setEditContent(data.content);
      setEditVisibility(data.visibility as PostVisibility);
    } catch (err: any) {
      setError(err.message || 'Failed to load post');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!postId) return;
    if (!confirm('Delete this post?')) return;
    try {
      await deletePost(postId);
      navigate('/feed');
    } catch (err: any) {
      alert(err.message || 'Failed to delete post');
    }
  };

  const handleUpdate = async () => {
    if (!postId) return;
    setActionLoading('update');
    try {
      await updatePost(postId, { content: editContent });
      await loadPost(postId);
    } catch (err: any) {
      alert(err.message || 'Failed to update post');
    } finally {
      setActionLoading(null);
    }
  };

  const handleVisibility = async () => {
    if (!postId) return;
    setActionLoading('visibility');
    try {
      await updatePostVisibility(postId, editVisibility);
      await loadPost(postId);
    } catch (err: any) {
      alert(err.message || 'Failed to update visibility');
    } finally {
      setActionLoading(null);
    }
  };

  const handleShare = async (action: 'add' | 'remove') => {
    if (!postId || !shareUserId.trim()) return;
    setActionLoading(action);
    try {
      if (action === 'add') {
        await addPostViewer(postId, { user_id: shareUserId.trim() });
      } else {
        await removePostViewer(postId, { user_id: shareUserId.trim() });
      }
      setShareUserId('');
    } catch (err: any) {
      alert(err.message || 'Failed to update access');
    } finally {
      setActionLoading(null);
    }
  };

  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage message={error} />;
  if (!post) return <p className="text-slate-50">Post not found.</p>;

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto max-w-4xl space-y-6">
        <nav className="flex items-center justify-between rounded-2xl border border-white/10 bg-card/70 px-4 py-3 shadow-soft">
          <div className="flex items-center gap-4 text-sm text-muted">
            <Link to="/dashboard" className="font-semibold text-primary-300 hover:text-primary-200">
              Dashboard
            </Link>
            <span className="text-white/30">/</span>
            <Link to="/feed" className="font-semibold text-primary-300 hover:text-primary-200">
              Feed
            </Link>
            <span className="text-white/30">/</span>
            <span>Post</span>
          </div>
          <LogoutButton />
        </nav>

        <div className="rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft space-y-4">
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <p className="text-sm text-muted">Author: {post.author_id}</p>
              {post.created_at && (
                <p className="text-xs text-muted">{new Date(post.created_at).toLocaleString()}</p>
              )}
            </div>
            <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-semibold text-slate-100">
              {post.visibility}
            </span>
          </div>

          <p className="text-slate-50 whitespace-pre-wrap">{post.content}</p>

          {user?.user_id === post.author_id && (
            <div className="flex justify-end">
              <button
                onClick={handleDelete}
                className="rounded-lg border border-white/10 px-3 py-2 text-xs font-semibold text-slate-100 hover:bg-white/10 transition"
              >
                Delete
              </button>
            </div>
          )}

          <div className="space-y-3">
            <h3 className="text-lg font-semibold">Reactions</h3>
            {post.reactions?.length ? (
              <div className="space-y-2">
                {post.reactions.map((r) => (
                  <div key={r.reaction_id} className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm">
                    <div className="text-slate-100 font-semibold">{r.reaction_type}</div>
                    <div className="text-xs text-muted">by {r.user_id}</div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted">No reactions yet.</p>
            )}
          </div>

          <div className="space-y-3">
            <h3 className="text-lg font-semibold">Comments</h3>
            {post.comments?.length ? (
              <div className="space-y-2">
                {post.comments.map((c) => (
                  <div key={c.comment_id} className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm space-y-1">
                    <div className="text-slate-100">{c.content}</div>
                    <div className="text-xs text-muted">by {c.user_id}</div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted">No comments yet.</p>
            )}
          </div>

          {user?.user_id === post.author_id && (
            <div className="space-y-4 rounded-xl border border-white/10 bg-white/5 p-4">
              <h3 className="text-lg font-semibold">Manage Post</h3>
              <div className="space-y-2">
                <label className="text-sm text-muted">Content</label>
                <textarea
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  rows={4}
                  className="bg-white/5"
                  disabled={actionLoading === 'update'}
                />
                <button
                  onClick={handleUpdate}
                  disabled={actionLoading === 'update'}
                  className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {actionLoading === 'update' ? 'Saving...' : 'Save Content'}
                </button>
              </div>

              <div className="space-y-2">
                <label className="text-sm text-muted">Visibility</label>
                <select
                  value={editVisibility}
                  onChange={(e) => setEditVisibility(e.target.value as PostVisibility)}
                  disabled={actionLoading === 'visibility'}
                  className="bg-white/5"
                >
                  <option value="public">Public</option>
                  <option value="semi-private">Semi-private</option>
                  <option value="private">Private</option>
                </select>
                <button
                  onClick={handleVisibility}
                  disabled={actionLoading === 'visibility'}
                  className="rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {actionLoading === 'visibility' ? 'Saving...' : 'Update Visibility'}
                </button>
              </div>

              <div className="space-y-2">
                <label className="text-sm text-muted">Share with user (private/semi-private)</label>
                <input
                  type="text"
                  value={shareUserId}
                  onChange={(e) => setShareUserId(e.target.value)}
                  placeholder="User ID"
                  disabled={actionLoading === 'add' || actionLoading === 'remove'}
                />
                <div className="flex gap-2">
                  <button
                    onClick={() => handleShare('add')}
                    disabled={actionLoading === 'add'}
                    className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-500 transition disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    {actionLoading === 'add' ? 'Adding...' : 'Add Viewer'}
                  </button>
                  <button
                    onClick={() => handleShare('remove')}
                    disabled={actionLoading === 'remove'}
                    className="rounded-lg border border-white/10 px-4 py-2 text-sm font-semibold text-slate-100 hover:bg-white/10 transition disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    {actionLoading === 'remove' ? 'Removing...' : 'Remove Viewer'}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
