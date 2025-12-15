import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { PostForm } from '../components/posts/PostForm';
import { PostCard } from '../components/posts/PostCard';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { ErrorMessage } from '../components/common/ErrorMessage';
import { LogoutButton } from '../components/auth/LogoutButton';
import {
  createPost,
  deletePost,
  getFeed,
} from '../services/postsService';
import type { FeedPost, CreatePostRequest } from '../types';

export function FeedPage() {
  const { user } = useAuth();
  const [feed, setFeed] = useState<FeedPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [posting, setPosting] = useState(false);

  useEffect(() => {
    loadFeed();
  }, []);

  const loadFeed = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getFeed();
      setFeed(Array.isArray(data) ? data : []);
    } catch (err: any) {
      setError(err.message || 'Failed to load feed');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (data: CreatePostRequest) => {
    setPosting(true);
    try {
      const created = await createPost(data);
      setFeed((prev) => [
        {
          ...created,
          reaction_count: 0,
          comment_count: 0,
        },
        ...prev,
      ]);
    } finally {
      setPosting(false);
    }
  };

  const handleDelete = async (postId: string) => {
    try {
      await deletePost(postId);
      setFeed((prev) => prev.filter((p) => p.post_id !== postId));
    } catch (err: any) {
      alert(err.message || 'Failed to delete post');
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="min-h-screen bg-surface text-slate-50 px-4 py-10">
      <div className="mx-auto max-w-5xl space-y-6">
        <nav className="flex items-center justify-between rounded-2xl border border-white/10 bg-card/70 px-4 py-3 shadow-soft">
          <div className="flex items-center gap-4 text-sm text-muted">
            <Link to="/dashboard" className="font-semibold text-primary-300 hover:text-primary-200">
              Dashboard
            </Link>
            <span className="text-white/30">/</span>
            <span>Feed</span>
          </div>
          <LogoutButton />
        </nav>

        <PostForm onSubmit={handleCreate} isSubmitting={posting} />

        {error && <ErrorMessage message={error} />}

        <div className="space-y-4">
          {feed.length === 0 ? (
            <p className="rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-muted">
              No posts yet.
            </p>
          ) : (
            feed.map((post) => (
              <PostCard
                key={post.post_id}
                post={post}
                isOwner={user?.user_id === post.author_id}
                onDelete={user?.user_id === post.author_id ? handleDelete : undefined}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
}
