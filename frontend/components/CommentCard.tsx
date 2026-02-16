"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import ImageUpload from "@/components/ImageUpload";
import { editComment, deleteComment, uploadFile, ApiError } from "@/lib/api";
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
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [shouldRemoveImage, setShouldRemoveImage] = useState(false);

  const isOwner = currentUserId === comment.author_id;

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

  const handleStartEdit = () => {
    setIsEditing(true);
    setEditContent(comment.content);
    setImageFile(null);
    setImagePreview(null);
    setShouldRemoveImage(false);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditContent(comment.content);
    setImageFile(null);
    setImagePreview(null);
    setShouldRemoveImage(false);
  };

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
            {isOwner && (
              <div className="comment-actions">
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
              </div>
            )}
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
      </div>
    </div>
  );
}
