"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import ImageUpload from "@/components/ImageUpload";
import { editComment, deleteComment, uploadFile, ApiError, createComment } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import type { CommentWithReplies } from "@/lib/types";

type CommentCardProps = {
  comment: CommentWithReplies;
  postId: string;
  currentUserId?: string;
  replies?: CommentWithReplies[];
  depth?: number;
};

export default function CommentCard({
                                      comment,
                                      postId,
                                      currentUserId,
                                      replies = [],
                                      depth = 0
                                    }: CommentCardProps) {
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [shouldRemoveImage, setShouldRemoveImage] = useState(false);

  // Reply states
  const [isReplying, setIsReplying] = useState(false);
  const [replyContent, setReplyContent] = useState("");
  const [replyImageFile, setReplyImageFile] = useState<File | null>(null);
  const [replyImagePreview, setReplyImagePreview] = useState<string | null>(null);
  const [isUploadingReply, setIsUploadingReply] = useState(false);

  const [isCollapsed, setIsCollapsed] = useState(depth > 0);
  const [isAnimating, setIsAnimating] = useState(false);

  const isOwner = currentUserId === comment.author_id;
  const hasReplies = replies?.length > 0;

  // Count total nested replies
  const countAllReplies = (comments: CommentWithReplies[] | undefined): number => {
    if (comments === undefined) return;
    return comments.reduce((count, comment) => {
      return count + 1 + countAllReplies(comment.replies);
    }, 0);
  };

  const totalReplies = hasReplies ? countAllReplies(replies) : 0;

  const handleImageSelect = (file: File) => {
    setImageFile(file);
    setImagePreview(URL.createObjectURL(file));
    setShouldRemoveImage(false);
  };

  const handleRemoveImage = () => {
    setImageFile(null);
    setImagePreview(null);
    setShouldRemoveImage(true);
  };

  const handleReplyImageSelect = (file: File) => {
    setReplyImageFile(file);
    setReplyImagePreview(URL.createObjectURL(file));
  };

  const handleRemoveReplyImage = () => {
    setReplyImageFile(null);
    setReplyImagePreview(null);
  };

  const handleStartEdit = () => {
    setIsEditing(true);
    setEditContent(comment.content);
    setImageFile(null);
    setImagePreview(null);
    setShouldRemoveImage(false);
  };

  const handleStartReply = () => {
    setIsReplying(true);
    setReplyContent("");
    setReplyImageFile(null);
    setReplyImagePreview(null);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditContent(comment.content);
    setImageFile(null);
    setImagePreview(null);
    setShouldRemoveImage(false);
  };

  const handleCancelReply = () => {
    setIsReplying(false);
    setReplyContent("");
    setReplyImageFile(null);
    setReplyImagePreview(null);
  };

  const addCommentReply = useMutation({
    mutationFn: async () => {
      let image_id: string | undefined = undefined;

      if (replyImageFile) {
        setIsUploadingReply(true);
        try {
          const uploadedImage = await uploadFile(replyImageFile);
          image_id = uploadedImage.file_id;
        } finally {
          setIsUploadingReply(false);
        }
      }

      return createComment(postId, {
        content: replyContent,
        parent_id: comment.comment_id,
        ...(image_id && { image_id })
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      setReplyContent("");
      setReplyImageFile(null);
      setReplyImagePreview(null);
      setIsReplying(false);
    },
    onError: () => {
      setIsUploadingReply(false);
    },
  });

  const updateComment = useMutation({
    mutationFn: async () => {
      let image_id: string | undefined = undefined;

      if (shouldRemoveImage) {
        image_id = "";
      } else if (imageFile) {
        setIsUploading(true);
        try {
          const uploadedImage = await uploadFile(imageFile);
          image_id = uploadedImage.file_id;
        } finally {
          setIsUploading(false);
        }
      }

      return editComment(postId, comment.comment_id, {
        content: editContent,
        ...(image_id !== undefined && { image_id })
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      setIsEditing(false);
      setImageFile(null);
      setImagePreview(null);
      setShouldRemoveImage(false);
    },
    onError: () => {
      setIsUploading(false);
    },
  });

  const removeComment = useMutation({
    mutationFn: () => deleteComment(postId, comment.comment_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
    },
    onError: () => {},
  });

  const authorName = comment.author_nickname
      ? comment.author_nickname
      : comment.author_first_name && comment.author_last_name
          ? `${comment.author_first_name} ${comment.author_last_name}`
          : "Unknown User";

  const displayImageUrl = imagePreview || (!shouldRemoveImage && comment.image_url) || null;

  const handleToggleCollapse = () => {
    if (!isCollapsed) {
      setIsAnimating(true);
      setIsCollapsed(true);
      setTimeout(() => {
        setIsAnimating(false);
      }, 200);
    } else {
      setIsCollapsed(false);
    }
  };

  return (
      <div className={`comment-thread ${depth > 0 ? 'comment-reply' : ''}`}>
        <div className="comment-card">
          {depth > 0 && <div className="comment-thread-line" />}
          <Link href={`/profile/${comment.author_id}`} className="comment-avatar-link">
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
                  <div className="comment-edit-content">
                <textarea
                    value={editContent}
                    onChange={(e) => setEditContent(e.target.value)}
                    className="comment-edit-input"
                    rows={3}
                />

                    {displayImageUrl && (
                        <div className="comment-image-preview">
                          <img src={displayImageUrl} alt="Comment" />
                          <button
                              type="button"
                              className="remove-image-btn"
                              onClick={handleRemoveImage}
                              aria-label="Remove image"
                          >
                            ✕
                          </button>
                        </div>
                    )}
                  </div>

                  <div className="comment-edit-actions">
                    {!displayImageUrl && (
                        <ImageUpload
                            onImageSelect={handleImageSelect}
                            label="📎 Add Image"
                        />
                    )}
                    <Button
                        onClick={() => updateComment.mutate()}
                        disabled={updateComment.isPending || isUploading || !editContent.trim()}
                    >
                      {isUploading ? "Uploading..." : updateComment.isPending ? "Saving..." : "Save"}
                    </Button>
                    <Button variant="ghost" onClick={handleCancelEdit}>
                      Cancel
                    </Button>
                  </div>

                  {updateComment.isError && (
                      <div className="comment-form-error">
                        {updateComment.error instanceof ApiError && updateComment.error.details
                            ? typeof updateComment.error.details === 'object' && !Array.isArray(updateComment.error.details)
                                ? Object.values(updateComment.error.details)[0]
                                : typeof updateComment.error.details === 'string'
                                    ? updateComment.error.details
                                    : updateComment.error.message
                            : updateComment.error.message}
                      </div>
                  )}
                </div>
            ) : (
                <>
                  <p className="comment-text">{comment.content}</p>
                  {comment.image_url && comment.image_url.trim() !== "" && (
                      <img
                          src={comment.image_url}
                          alt="Comment"
                          className="comment-image"
                      />
                  )}
                  <div className="comment-actions">
                    {isOwner && (
                        <>
                          <button
                              className="comment-action"
                              onClick={handleStartEdit}
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
                        </>
                    )}
                    <button
                        className="comment-action"
                        onClick={handleStartReply}
                    >
                      Reply
                    </button>
                    {hasReplies && (
                        <button
                            className="comment-action collapse-toggle"
                            onClick={handleToggleCollapse}
                        >
                          {isCollapsed ? (
                              <>
                                <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
                                  <path d="M3 4.5L6 7.5L9 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                                </svg>
                                Show {totalReplies} {totalReplies === 1 ? 'reply' : 'replies'}
                              </>
                          ) : (
                              <>
                                <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
                                  <path d="M9 7.5L6 4.5L3 7.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                                </svg>
                                Hide {totalReplies} {totalReplies === 1 ? 'reply' : 'replies'}
                              </>
                          )}
                        </button>
                    )}
                  </div>
                  {removeComment.isError && (
                      <div className="comment-form-error" style={{ marginTop: "8px" }}>
                        {removeComment.error instanceof ApiError && removeComment.error.details
                            ? typeof removeComment.error.details === 'object' && !Array.isArray(removeComment.error.details)
                                ? Object.values(removeComment.error.details)[0]
                                : typeof removeComment.error.details === 'string'
                                    ? removeComment.error.details
                                    : removeComment.error.message
                            : removeComment.error.message}
                      </div>
                  )}
                </>
            )}

            {/* Reply Form */}
            {isReplying && currentUserId && (
                <div className="comment-form comment-reply-form">
                  <div className="comment-form-container">
                    <div className="comment-form-input-area">
                      <div className="comment-form-content">
                    <textarea
                        className="comment-input"
                        placeholder="Write a reply..."
                        value={replyContent}
                        onChange={(e) => setReplyContent(e.target.value)}
                        rows={2}
                    />

                        {replyImagePreview && (
                            <div className="comment-image-preview">
                              <img src={replyImagePreview} alt="Preview" />
                              <button
                                  type="button"
                                  className="remove-image-btn"
                                  onClick={handleRemoveReplyImage}
                                  aria-label="Remove image"
                              >
                                ✕
                              </button>
                            </div>
                        )}
                      </div>

                      <div className="comment-form-actions">
                        {!replyImagePreview && (
                            <ImageUpload
                                onImageSelect={handleReplyImageSelect}
                                label="📎 Add Image"
                            />
                        )}
                      </div>
                    </div>

                    <div className="comment-form-button">
                      <Button
                          onClick={() => addCommentReply.mutate()}
                          disabled={addCommentReply.isPending || isUploadingReply || !replyContent.trim()}
                      >
                        {isUploadingReply ? "Uploading..." : addCommentReply.isPending ? "Posting..." : "Reply"}
                      </Button>
                      <Button variant="ghost" onClick={handleCancelReply}>
                        Cancel
                      </Button>
                    </div>
                  </div>

                  {addCommentReply.isError && (
                      <div className="comment-form-error">
                        {addCommentReply.error instanceof ApiError && addCommentReply.error.details
                            ? typeof addCommentReply.error.details === 'object' && !Array.isArray(addCommentReply.error.details)
                                ? Object.values(addCommentReply.error.details)[0]
                                : typeof addCommentReply.error.details === 'string'
                                    ? addCommentReply.error.details
                                    : addCommentReply.error.message
                            : addCommentReply.error.message}
                      </div>
                  )}
                </div>
            )}
          </div>
        </div>

        {/* Render nested replies */}
        {hasReplies && (isAnimating || !isCollapsed) && (
            <div className={`comment-replies ${isCollapsed && isAnimating ? 'collapsing' : ''}`}>
              {replies?.map((reply) => (
                  <CommentCard
                      key={reply.comment_id}
                      comment={reply}
                      postId={postId}
                      currentUserId={currentUserId}
                      replies={reply.replies}
                      depth={depth + 1}
                  />
              ))}
            </div>
        )}
      </div>
  );
}