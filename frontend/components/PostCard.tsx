"use client";

import Link from "next/link";
import type { FeedPost } from "@/lib/types";
import { formatDate } from "@/lib/utils";
import ReactionButton from "@/components/ReactionButton";
import CommentSection from "@/components/CommentSection";
import PostActions from "@/components/PostActions";
import Avatar from "@/components/Avatar";

type PostCardProps = {
  post: FeedPost;
  currentUserId?: string;
  showFullComments?: boolean;
};

function getVisibilityLabel(visibility: string): string {
  switch (visibility) {
    case "public":
      return "Public";
    case "semi-private":
      return "Almost Private";
    case "private":
      return "Private";
    case "group":
      return "Group";
    default:
      return visibility;
  }
}

export default function PostCard({ post, currentUserId, showFullComments = false }: PostCardProps) {
  const isOwner = currentUserId === post.author_id;
  const authorName = post.author_nickname
    ? post.author_nickname
    : post.author_first_name && post.author_last_name
    ? `${post.author_first_name} ${post.author_last_name}`
    : "User";

  return (
    <article className="surface card post-card">
      <div className="post-header">
        <Link href={`/profile/${post.author_id}`} className="post-author">
          <Avatar
            src={post.author_avatar}
            name={authorName}
            size={40}
          />
          <div className="post-author-info">
            <span className="post-author-name">{authorName}</span>
            <span className="post-time">{formatDate(post.created_at)}</span>
          </div>
        </Link>
        <div className="post-header-right">
          <span className="tag">{getVisibilityLabel(post.visibility)}</span>
          <PostActions
            postId={post.post_id}
            content={post.content}
            visibility={post.visibility}
            isOwner={isOwner}
            imageUrl={post.image_url}
          />
        </div>
      </div>

      <Link href={`/post/${post.post_id}`} className="post-content-link">
        <p className="post-text">{post.content}</p>
      </Link>

      {post.image_url ? (
        <Link href={`/post/${post.post_id}`}>
          <img
            src={post.image_url}
            alt="Post visual"
            className="post-image"
          />
        </Link>
      ) : null}

      <div className="post-interactions">
        <ReactionButton
          postId={post.post_id}
          reactionCount={post.reaction_count}
          userReacted={post.user_reacted}
        />
        <Link href={`/post/${post.post_id}`} className="post-comment-link">
          💬 {post.comment_count}
        </Link>
      </div>

      <CommentSection
        postId={post.post_id}
        commentCount={post.comment_count}
        currentUserId={currentUserId}
        initiallyExpanded={showFullComments}
      />
    </article>
  );
}
