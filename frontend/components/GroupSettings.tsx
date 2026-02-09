"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Modal from "@/components/Modal";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { updateGroup, deleteGroup, ApiError } from "@/lib/api";
import type { Group } from "@/lib/types";

type GroupSettingsProps = {
  group: Group;
  isOpen: boolean;
  onClose: () => void;
};

export default function GroupSettings({ group, isOpen, onClose }: GroupSettingsProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
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
      onClose();
    },
    onError: (error) => {
      setValidationErrors({});
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      }
    },
  });

  const remove = useMutation({
    mutationFn: () => deleteGroup(group.group_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      router.push("/groups");
    },
  });

  if (showDeleteConfirm) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} title="Delete Group">
        <p>
          Are you sure you want to delete <strong>{group.group_name}</strong>?
          This action cannot be undone.
        </p>
        <div className="modal-actions">
          <Button
            onClick={() => remove.mutate()}
            disabled={remove.isPending}
            className="danger"
          >
            Delete Group
          </Button>
          <Button variant="ghost" onClick={() => setShowDeleteConfirm(false)}>
            Cancel
          </Button>
        </div>
      </Modal>
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
