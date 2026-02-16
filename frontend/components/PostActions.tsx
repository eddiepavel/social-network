"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Dropdown from "@/components/Dropdown";
import Modal from "@/components/Modal";
import Button from "@/components/Button";
import ImageUpload from "@/components/ImageUpload";
import { editPost, deletePost, uploadFile, ApiError } from "@/lib/api";

type PostActionsProps = {
  postId: string;
  content: string;
  visibility: string;
  isOwner: boolean;
  imageUrl?: string;
};

export default function PostActions({ postId, content, visibility, isOwner, imageUrl }: PostActionsProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [editContent, setEditContent] = useState(content);
  const [editVisibility, setEditVisibility] = useState(visibility);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [shouldRemoveImage, setShouldRemoveImage] = useState(false);

  const update = useMutation({
    mutationFn: async (input: { content: string; visibility: string; image_id?: string }) => {
      return editPost(postId, input);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      setIsEditOpen(false);
      setImageFile(null);
      setImagePreview(null);
    },
    onError: () => {},
  });

  const handleImageSelect = (file: File) => {
    setImageFile(file);
    const reader = new FileReader();
    reader.onload = (e) => setImagePreview(e.target?.result as string);
    reader.readAsDataURL(file);
    setShouldRemoveImage(false);
  };

  const handleRemoveImage = () => {
    setImageFile(null);
    setImagePreview(null);
    setShouldRemoveImage(true);
  };

  const handleUpdate = async () => {
    setIsUploading(true);
    try {
      let imageId: string | undefined;
      if (imageFile) {
        const uploaded = await uploadFile(imageFile);
        imageId = uploaded.file_id;
      } else if (shouldRemoveImage) {
        imageId = ""; // Empty string to remove image
      }
      // If imageId is undefined, backend will keep existing image
      update.mutate({
        content: editContent,
        visibility: editVisibility,
        image_id: imageId,
      });
    } catch (error) {
      console.error("Upload failed:", error);
    } finally {
      setIsUploading(false);
    }
  };

  const remove = useMutation({
    mutationFn: () => deletePost(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      setIsDeleteOpen(false);
      router.push("/feed");
    },
    onError: () => {},
  });

  if (!isOwner) return null;

  return (
    <>
      <Dropdown
        trigger={
          <button className="post-actions-trigger" aria-label="Post options">
            ⋮
          </button>
        }
        items={[
          { label: "Edit", onClick: () => setIsEditOpen(true) },
          { label: "Delete", onClick: () => setIsDeleteOpen(true), danger: true },
        ]}
      />

      <Modal isOpen={isEditOpen} onClose={() => setIsEditOpen(false)} title="Edit Post">
        <div className="edit-post-form">
          <textarea
            className="edit-post-input"
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            rows={4}
          />
          
          {(imagePreview || (imageUrl && !shouldRemoveImage)) && (
            <div className="post-image-preview">
              <img src={imagePreview || imageUrl} alt="Preview" />
              <button
                className="remove-preview"
                onClick={handleRemoveImage}
                type="button"
              >
                Remove
              </button>
            </div>
          )}

          <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
            <label className="form-field" style={{ flex: 1 }}>
              <span>Visibility</span>
              <select
                value={editVisibility}
                onChange={(e) => setEditVisibility(e.target.value)}
              >
                <option value="public">Public</option>
                <option value="semi-private">Semi-private</option>
                <option value="private">Private</option>
              </select>
            </label>
            {!(imagePreview || (imageUrl && !shouldRemoveImage)) && (
              <div style={{ paddingTop: "1.5rem" }}>
                <ImageUpload
                  onImageSelect={handleImageSelect}
                  accept="image/*"
                  maxSizeMB={5}
                  label="Add photo"
                  compact
                />
              </div>
            )}
          </div>

          {update.isError ? (
            <p style={{ color: "#b42318", fontSize: "0.85rem" }}>
              {update.error instanceof ApiError && typeof update.error.details === 'string'
                ? update.error.details
                : update.error.message}
            </p>
          ) : null}
          <div className="modal-actions">
            <Button
              onClick={handleUpdate}
              disabled={update.isPending || isUploading || !editContent.trim()}
            >
              {isUploading ? "Uploading..." : "Save changes"}
            </Button>
            <Button variant="ghost" onClick={() => setIsEditOpen(false)}>
              Cancel
            </Button>
          </div>
        </div>
      </Modal>

      <Modal isOpen={isDeleteOpen} onClose={() => setIsDeleteOpen(false)} title="Delete Post">
        <p>Are you sure you want to delete this post? This action cannot be undone.</p>
        {remove.isError ? (
          <p style={{ color: "#b42318", fontSize: "0.85rem" }}>
            {remove.error instanceof ApiError && typeof remove.error.details === 'string'
              ? remove.error.details
              : remove.error.message}
          </p>
        ) : null}
        <div className="modal-actions">
          <Button
            onClick={() => remove.mutate()}
            disabled={remove.isPending}
            className="danger"
          >
            Delete
          </Button>
          <Button variant="ghost" onClick={() => setIsDeleteOpen(false)}>
            Cancel
          </Button>
        </div>
      </Modal>
    </>
  );
}
