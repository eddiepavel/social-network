"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import CommentCard from "@/components/CommentCard";
import Button from "@/components/Button";
import ImageUpload from "@/components/ImageUpload";
import { getComments, createComment, uploadFile, ApiError } from "@/lib/api";

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
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const { data: comments, isLoading } = useQuery({
    queryKey: ["comments", postId],
    queryFn: () => getComments(postId),
    enabled: isExpanded,
  });

  const handleImageSelect = (file: File) => {
    setImageFile(file);
    setImagePreview(URL.createObjectURL(file));
  };

  const handleRemoveImage = () => {
    setImageFile(null);
    setImagePreview(null);
  };

  const addComment = useMutation({
    mutationFn: async () => {
      let image_id: string | undefined = undefined;

      if (imageFile) {
        setIsUploading(true);
        try {
          const uploadedImage = await uploadFile(imageFile);
          image_id = uploadedImage.file_id;
        } finally {
          setIsUploading(false);
        }
      }

      return createComment(postId, {
        content: newComment,
        ...(image_id && { image_id })
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      setNewComment("");
      setImageFile(null);
      setImagePreview(null);
    },
    onError: () => {
      setIsUploading(false);
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
        <div className="comment-form">
          <div className="comment-form-container">
            <div className="comment-form-input-area">
              <div className="comment-form-content">
                <textarea
                  className="comment-input"
                  placeholder="Write a comment..."
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  rows={2}
                />

                {imagePreview && (
                  <div className="comment-image-preview">
                    <img src={imagePreview} alt="Preview" />
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

              {!newComment.trim() && addComment.isError && (
                <div className="comment-form-error">
                  Please enter a comment
                </div>
              )}

              <div className="comment-form-actions">
                {!imagePreview && (
                  <ImageUpload
                    onImageSelect={handleImageSelect}
                    label="📎 Add Image"
                  />
                )}
              </div>
            </div>

            <div className="comment-form-button">
              <Button
                onClick={() => addComment.mutate()}
                disabled={addComment.isPending || isUploading || !newComment.trim()}
              >
                {isUploading ? "Uploading..." : addComment.isPending ? "Posting..." : "Post"}
              </Button>
            </div>
          </div>

          {addComment.isError && (
            <div className="comment-form-error">
              {addComment.error instanceof ApiError && addComment.error.details
                ? typeof addComment.error.details === 'object' && !Array.isArray(addComment.error.details)
                  ? Object.values(addComment.error.details)[0]
                  : typeof addComment.error.details === 'string'
                  ? addComment.error.details
                  : addComment.error.message
                : addComment.error.message}
            </div>
          )}
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
