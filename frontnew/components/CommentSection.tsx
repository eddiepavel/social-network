"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import CommentCard from "@/components/CommentCard";
import Button from "@/components/Button";
import { getComments, createComment } from "@/lib/api";

type CommentSectionProps = {
  postId: string;
  commentCount: number;
  currentUserId?: string;
  initiallyExpanded?: boolean;
};

export default function CommentSection({
  postId,
  commentCount,
  currentUserId,
  initiallyExpanded = false,
}: CommentSectionProps) {
  const queryClient = useQueryClient();
  const [isExpanded, setIsExpanded] = useState(initiallyExpanded);
  const [newComment, setNewComment] = useState("");

  const { data: comments, isLoading } = useQuery({
    queryKey: ["comments", postId],
    queryFn: () => getComments(postId),
    enabled: isExpanded,
  });

  const addComment = useMutation({
    mutationFn: () => createComment(postId, { content: newComment }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      setNewComment("");
    },
  });

  if (!isExpanded) {
    return (
      <button className="comment-toggle" onClick={() => setIsExpanded(true)}>
        {commentCount > 0 ? `View ${commentCount} comments` : "Add a comment"}
      </button>
    );
  }

  return (
    <div className="comment-section">
      <button className="comment-toggle" onClick={() => setIsExpanded(false)}>
        Hide comments
      </button>

      {currentUserId && (
        <div className="comment-input-wrapper">
          <textarea
            className="comment-input"
            placeholder="Write a comment..."
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            rows={2}
          />
          <Button
            onClick={() => addComment.mutate()}
            disabled={addComment.isPending || !newComment.trim()}
          >
            Post
          </Button>
        </div>
      )}

      {isLoading ? (
        <p className="comment-loading">Loading comments...</p>
      ) : (
        <div className="comments-list">
          {comments?.map((comment) => (
            <CommentCard
              key={comment.comment_id}
              comment={comment}
              postId={postId}
              currentUserId={currentUserId}
            />
          ))}
          {comments?.length === 0 && (
            <p className="no-comments">No comments yet. Be the first!</p>
          )}
        </div>
      )}
    </div>
  );
}
