"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Dropdown from "@/components/Dropdown";
import Modal from "@/components/Modal";
import Button from "@/components/Button";
import { editPost, deletePost } from "@/lib/api";

type PostActionsProps = {
  postId: string;
  content: string;
  visibility: string;
  isOwner: boolean;
};

export default function PostActions({ postId, content, visibility, isOwner }: PostActionsProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [editContent, setEditContent] = useState(content);
  const [editVisibility, setEditVisibility] = useState(visibility);

  const update = useMutation({
    mutationFn: () => editPost(postId, { content: editContent, visibility: editVisibility }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      setIsEditOpen(false);
    },
  });

  const remove = useMutation({
    mutationFn: () => deletePost(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      setIsDeleteOpen(false);
      router.push("/feed");
    },
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
          <label className="form-field">
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
          <div className="modal-actions">
            <Button
              onClick={() => update.mutate()}
              disabled={update.isPending || !editContent.trim()}
            >
              Save changes
            </Button>
            <Button variant="ghost" onClick={() => setIsEditOpen(false)}>
              Cancel
            </Button>
          </div>
        </div>
      </Modal>

      <Modal isOpen={isDeleteOpen} onClose={() => setIsDeleteOpen(false)} title="Delete Post">
        <p>Are you sure you want to delete this post? This action cannot be undone.</p>
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
