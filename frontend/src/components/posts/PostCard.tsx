import { Link } from 'react-router-dom';
import type { FeedPost, Post } from '../../types';

interface PostCardProps {
  post: FeedPost | Post;
  isOwner: boolean;
  onDelete?: (id: string) => Promise<void>;
}

export function PostCard({ post, isOwner, onDelete }: PostCardProps) {
  const created = post.created_at ? new Date(post.created_at).toLocaleString() : '';

  const handleDelete = async () => {
    if (!onDelete) return;
    if (!confirm('Delete this post?')) return;
    await onDelete(post.post_id);
  };

  const reactionCount = 'reaction_count' in post ? post.reaction_count : undefined;
  const commentCount = 'comment_count' in post ? post.comment_count : undefined;

  return (
    <div className="rounded-2xl border border-white/10 bg-card/70 p-5 shadow-soft space-y-3">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1">
          <p className="text-sm text-muted">Author: {post.author_id}</p>
          <p className="text-xs text-muted">{created}</p>
        </div>
        <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-semibold text-slate-100">
          {post.visibility}
        </span>
      </div>

      <p className="text-slate-50 whitespace-pre-wrap">{post.content}</p>

      <div className="flex items-center justify-between text-sm text-muted">
        <div className="flex gap-3">
          {typeof reactionCount === 'number' && (
            <span>Reactions: {reactionCount}</span>
          )}
          {typeof commentCount === 'number' && (
            <span>Comments: {commentCount}</span>
          )}
        </div>
        <Link
          to={`/posts/${post.post_id}`}
          className="text-primary-300 hover:text-primary-200 font-semibold"
        >
          View
        </Link>
      </div>

      {isOwner && onDelete && (
        <div className="flex justify-end">
          <button
            onClick={handleDelete}
            className="rounded-lg border border-white/10 px-3 py-2 text-xs font-semibold text-slate-100 hover:bg-white/10 transition"
          >
            Delete
          </button>
        </div>
      )}
    </div>
  );
}
