import type { FeedPost } from "@/lib/types";
import { formatDate } from "@/lib/utils";

export default function PostCard({ post }: { post: FeedPost }) {
  return (
    <article className="surface card">
      <div className="post-meta">
        <span className="tag">{post.visibility}</span>
        <span>{formatDate(post.created_at)}</span>
      </div>
      <p style={{ fontSize: "1.05rem" }}>{post.content}</p>
      {post.image_url ? (
        <img
          src={post.image_url}
          alt="Post visual"
          style={{ width: "100%", borderRadius: 14, border: "1px solid #e2d6bf" }}
        />
      ) : null}
      <div className="post-meta">
        <span>{post.reaction_count} reactions</span>
        <span>{post.comment_count} comments</span>
      </div>
    </article>
  );
}
