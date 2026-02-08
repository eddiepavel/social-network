"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import ImageUpload from "@/components/ImageUpload";
import { createPost, uploadFile, ApiError } from "@/lib/api";

export default function PostComposer() {
  const [content, setContent] = useState("");
  const [visibility, setVisibility] = useState("public");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const queryClient = useQueryClient();

  const create = useMutation({
    mutationFn: async (input: { content: string; image_id?: string; visibility: string }) => {
      return createPost(input);
    },
    onSuccess: () => {
      setContent("");
      setImageFile(null);
      setImagePreview(null);
      setValidationErrors({});
      queryClient.invalidateQueries({ queryKey: ["feed"] });
    },
    onError: (error) => {
      setValidationErrors({});
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      }
    },
  });

  const handleImageSelect = (file: File | null) => {
    setImageFile(file);
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => setImagePreview(e.target?.result as string);
      reader.readAsDataURL(file);
    } else {
      setImagePreview(null);
    }
  };

  const handlePost = async () => {
    setIsUploading(true);
    try {
      let imageId: string | undefined;
      if (imageFile) {
        const uploaded = await uploadFile(imageFile);
        imageId = uploaded.file_id;
      }
      create.mutate({
        content,
        image_id: imageId,
        visibility,
      });
    } catch (error) {
      console.error("Upload failed:", error);
    } finally {
      setIsUploading(false);
    }
  };

  const isPosting = create.isPending || isUploading;

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
            onClick={() => handleImageSelect(null)}
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
            <option value="public">Public</option>
            <option value="semi-private">Semi-private</option>
            <option value="private">Private</option>
          </select>
        </label>

        <div className="composer-actions">
          <ImageUpload
            onFileSelect={handleImageSelect}
            accept="image/*"
            maxSizeMB={5}
            label="Add photo"
            compact
          />
          <Button
            onClick={handlePost}
            disabled={!content.trim() || isPosting}
          >
            {isPosting ? "Posting..." : "Post"}
          </Button>
        </div>
      </div>

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
