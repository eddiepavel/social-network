"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import { editComment, deleteComment } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import type { Comment } from "@/lib/types";

type CommentCardProps = {
  comment: Comment;
  postId: string;
  currentUserId?: string;
};

export default function CommentCard({ comment, postId, currentUserId }: CommentCardProps) {
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);

  const isOwner = currentUserId === comment.author_id;

  const updateComment = useMutation({
    mutationFn: () => editComment(postId, comment.comment_id, { content: editContent }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      setIsEditing(false);
    },
  });

  const removeComment = useMutation({
    mutationFn: () => deleteComment(postId, comment.comment_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
    },
  });

  const authorName = comment.author_first_name && comment.author_last_name
    ? `${comment.author_first_name} ${comment.author_last_name}`
    : "Unknown User";

  return (
    <div className="comment-card">
      <Link href={`/profile/${comment.author_id}`}>
        <Avatar
          src={comment.author_avatar}
          name={authorName}
          size={32}
        />
      </Link>
      <div className="comment-content">
        <div className="comment-header">
          <Link href={`/profile/${comment.author_id}`} className="comment-author">
            {authorName}
          </Link>
          <span className="comment-time">{formatDate(comment.created_at)}</span>
        </div>

        {isEditing ? (
          <div className="comment-edit">
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              className="comment-edit-input"
            />
            <div className="comment-edit-actions">
              <Button
                onClick={() => updateComment.mutate()}
                disabled={updateComment.isPending || !editContent.trim()}
              >
                Save
              </Button>
              <Button variant="ghost" onClick={() => setIsEditing(false)}>
                Cancel
              </Button>
            </div>
          </div>
        ) : (
          <>
            <p className="comment-text">{comment.content}</p>
            {isOwner && (
              <div className="comment-actions">
                <button
                  className="comment-action"
                  onClick={() => setIsEditing(true)}
                >
                  Edit
                </button>
                <button
                  className="comment-action danger"
                  onClick={() => removeComment.mutate()}
                  disabled={removeComment.isPending}
                >
                  Delete
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
