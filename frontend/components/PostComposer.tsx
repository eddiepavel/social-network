"use client";

import { useState, useEffect } from "react";
import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import ImageUpload from "@/components/ImageUpload";
import Avatar from "@/components/Avatar";
import { createPost, uploadFile, getFollowers, ApiError } from "@/lib/api";
import useSession from "@/hooks/useSession";

export default function PostComposer() {
  const { data: session } = useSession();
  const [content, setContent] = useState("");
  const [visibility, setVisibility] = useState("");
  const [selectedFollowers, setSelectedFollowers] = useState<string[]>([]);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const queryClient = useQueryClient();

  // Fetch followers when visibility is private
  const { data: followers, isLoading: loadingFollowers } = useQuery({
    queryKey: ["my-followers", session?.user_id],
    queryFn: () => getFollowers(session!.user_id),
    enabled: visibility === "private" && !!session?.user_id,
  });

  // Reset selected followers when visibility changes
  useEffect(() => {
    if (visibility !== "private") {
      setSelectedFollowers([]);
    }
  }, [visibility]);

  // Auto-hide success message after 3 seconds
  useEffect(() => {
    if (successMessage) {
      const timer = setTimeout(() => setSuccessMessage(null), 3000);
      return () => clearTimeout(timer);
    }
  }, [successMessage]);

  const create = useMutation({
    mutationFn: async (input: { content: string; image_id?: string; visibility: string; allowed_users?: string[] }) => {
      return createPost(input);
    },
    onSuccess: () => {
      setContent("");
      setImageFile(null);
      setImagePreview(null);
      setSelectedFollowers([]);
      setVisibility("public");
      setValidationErrors({});
      setSuccessMessage("Post created successfully!");
      queryClient.invalidateQueries({ queryKey: ["feed"] });
    },
    onError: (error) => {
      setSuccessMessage(null);
      setValidationErrors({});
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      }
    },
  });

  const handleImageSelect = (file: File) => {
    setImageFile(file);
    const reader = new FileReader();
    reader.onload = (e) => setImagePreview(e.target?.result as string);
    reader.readAsDataURL(file);
  };

  const handleRemoveImage = () => {
    setImageFile(null);
    setImagePreview(null);
  };

  const handleFollowerToggle = (userId: string) => {
    setSelectedFollowers((prev) =>
      prev.includes(userId)
        ? prev.filter((id) => id !== userId)
        : [...prev, userId]
    );
  };

  const handleSelectAll = () => {
    if (followers) {
      setSelectedFollowers(followers.map((f) => f.user_id));
    }
  };

  const handleDeselectAll = () => {
    setSelectedFollowers([]);
  };

  const handlePost = async () => {
    setIsUploading(true);
    try {
      let imageId: string | undefined;
      if (imageFile) {
        const uploaded = await uploadFile(imageFile);
        imageId = uploaded.file_id;
      }
      let postVisibility = visibility;
      if (!visibility) {
        postVisibility = session?.is_public ? 'public' : 'semi-private'
      }
      create.mutate({
        content,
        image_id: imageId,
        visibility: postVisibility,
        allowed_users: visibility === "private" ? selectedFollowers : undefined,
      });
    } catch (error) {
      console.error("Upload failed:", error);
    } finally {
      setIsUploading(false);
    }
  };

  const isPosting = create.isPending || isUploading;
  const canPost = content.trim() && !isPosting && 
    (visibility !== "private" || selectedFollowers.length > 0);

  return (
    <div className="surface card">
      <h3>Share a moment</h3>
      <FormField
        label="What is on your mind?"
        name="content"
        as="textarea"
        placeholder="Start a conversation, drop a quick update, or share a thought."
        value={content}
        onChange={(event) => {
          setContent(event.target.value);
          if (validationErrors.content) setValidationErrors(prev => { const n = {...prev}; delete n.content; return n; });
        }}
        error={validationErrors.content}
      />

      {imagePreview && (
        <div className="post-image-preview">
          <img src={imagePreview} alt="Preview" />
          <button
            className="remove-preview"
            onClick={handleRemoveImage}
            type="button"
          >
            Remove
          </button>
        </div>
      )}

      <div className="composer-footer">
        <label className="form-field" style={{ flex: 1 }}>
          <span>Visibility</span>
          <select
            name="visibility"
            value={visibility}
            onChange={(event) => setVisibility(event.target.value)}
          >
            <option disabled={!session?.is_public} value="public">Public - Everyone can see</option>
            <option disabled={session?.is_public} value="semi-private">Almost Private - Only followers</option>
            <option value="private">Private - Select specific followers</option>
          </select>
        </label>

        <div className="composer-actions">
          {!imagePreview && (
            <ImageUpload
              onImageSelect={handleImageSelect}
              accept="image/*"
              maxSizeMB={5}
              label="Add photo"
              compact
            />
          )}
          <Button
            onClick={handlePost}
            disabled={!canPost}
          >
            {isPosting ? "Posting..." : "Post"}
          </Button>
        </div>
      </div>

      {/* Follower selection for private posts */}
      {visibility === "private" && (
        <div className="follower-selector">
          <div className="follower-selector-header">
            <span className="follower-selector-title">
              Select who can see this post
              {selectedFollowers.length > 0 && (
                <span className="selected-count"> ({selectedFollowers.length} selected)</span>
              )}
            </span>
            {followers && followers.length > 0 && (
              <div className="follower-selector-actions">
                <button
                  type="button"
                  className="text-button"
                  onClick={handleSelectAll}
                >
                  Select all
                </button>
                <span className="divider">|</span>
                <button
                  type="button"
                  className="text-button"
                  onClick={handleDeselectAll}
                >
                  Deselect all
                </button>
              </div>
            )}
          </div>
          
          {loadingFollowers ? (
            <p className="text-secondary">Loading followers...</p>
          ) : followers && followers.length > 0 ? (
            <div className="follower-list">
              {followers.map((follower) => (
                <label key={follower.user_id} className="follower-item">
                  <input
                    type="checkbox"
                    checked={selectedFollowers.includes(follower.user_id)}
                    onChange={() => handleFollowerToggle(follower.user_id)}
                  />
                  <Avatar
                    src={follower.avatar}
                    name={`${follower.first_name} ${follower.last_name}`}
                    size={32}
                  />
                  <span className="follower-name">
                    {follower.first_name} {follower.last_name}
                    {follower.nickname && (
                      <span className="follower-nickname">@{follower.nickname}</span>
                    )}
                  </span>
                </label>
              ))}
            </div>
          ) : (
            <p className="text-secondary empty-followers">
              You don&apos;t have any followers yet. Only your followers can be selected for private posts.
            </p>
          )}
          
          {visibility === "private" && selectedFollowers.length === 0 && followers && followers.length > 0 && (
            <p className="validation-hint">Please select at least one follower to post privately.</p>
          )}
        </div>
      )}

      {successMessage && (
        <p className="success-message">{successMessage}</p>
      )}

      {create.isError ? (
        create.error instanceof ApiError && typeof create.error.details === 'object' ? null : (
          <p style={{ color: "#b42318" }}>
            {create.error instanceof ApiError && typeof create.error.details === 'string'
              ? create.error.details
              : create.error.message}
          </p>
        )
      ) : null}
    </div>
  );
}
