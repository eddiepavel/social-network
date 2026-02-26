"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Modal from "@/components/Modal";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { updateGroup, deleteGroup, ApiError } from "@/lib/api";
import type { Group } from "@/lib/types";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { useToastContext } from "../app/providers";

type GroupSettingsProps = {
  group: Group;
  isOpen: boolean;
  onClose: () => void;
};

export default function GroupSettings({ group, isOpen, onClose }: GroupSettingsProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const toast = useToastContext();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [form, setForm] = useState({
    group_name: group.group_name,
    description: group.description,
  });

  const update = useMutation({
    mutationFn: () => updateGroup(group.group_id, form),
    onSuccess: () => {
      setValidationErrors({});
      queryClient.invalidateQueries({ queryKey: ["group", group.group_id] });
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      toast.success("Group settings updated successfully");
      onClose();
    },
    onError: (error) => {
      setValidationErrors({});
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      } else {
        const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message;
        toast.error(msg);
      }
    },
  });

  const remove = useMutation({
    mutationFn: () => deleteGroup(group.group_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      toast.success("Group deleted successfully");
      router.push("/groups");
    },
    onError: (error) => {
      const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message;
      toast.error(msg);
    },
  });

  if (showDeleteConfirm) {
    return (
      <ConfirmDialog
        isOpen={isOpen}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={() => remove.mutate()}
        title="Delete Group"
        message={`Are you sure you want to delete ${group.group_name}? This action cannot be undone. All posts, events, and member data will be permanently deleted.`}
        confirmText="Delete Group"
        cancelText="Cancel"
        type="danger"
        isLoading={remove.isPending}
      />
    );
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Group Settings">
      <div className="group-settings-form">
        <FormField
          label="Group Name"
          name="group_name"
          value={form.group_name}
          onChange={(e) => {
            setForm((prev) => ({ ...prev, group_name: e.target.value }));
            if (validationErrors.group_name) setValidationErrors(prev => { const n = {...prev}; delete n.group_name; return n; });
          }}
          error={validationErrors.group_name}
        />
        <FormField
          label="Description"
          name="description"
          as="textarea"
          value={form.description}
          onChange={(e) => {
            setForm((prev) => ({ ...prev, description: e.target.value }));
            if (validationErrors.description) setValidationErrors(prev => { const n = {...prev}; delete n.description; return n; });
          }}
          error={validationErrors.description}
        />
        {update.isError ? (
          update.error instanceof ApiError && typeof update.error.details === 'object' ? null : (
            <p style={{ color: "#b42318" }}>
              {update.error instanceof ApiError && typeof update.error.details === 'string'
                ? update.error.details
                : update.error.message}
            </p>
          )
        ) : null}
        <div className="modal-actions">
          <Button
            onClick={() => update.mutate()}
            disabled={update.isPending || !form.group_name.trim()}
          >
            Save Changes
          </Button>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
        </div>

        <hr style={{ margin: "20px 0", border: "none", borderTop: "1px solid var(--border)" }} />

        <div className="danger-zone">
          <h4 style={{ color: "#b42318", marginBottom: 8 }}>Danger Zone</h4>
          <p style={{ color: "var(--muted)", marginBottom: 12 }}>
            Once you delete a group, there is no going back. Please be certain.
          </p>
          <Button
            variant="ghost"
            className="danger"
            onClick={() => setShowDeleteConfirm(true)}
          >
            Delete Group
          </Button>
        </div>
      </div>
    </Modal>
  );
}
