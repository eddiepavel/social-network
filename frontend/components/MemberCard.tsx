"use client";

import { useState } from "react";
import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import { removeMember } from "@/lib/api";
import type { GroupMember } from "@/lib/types";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { useToastContext } from "../app/providers";

type MemberCardProps = {
  member: GroupMember;
  groupId: string;
  canRemove: boolean;
  isCurrentUser: boolean;
};

export default function MemberCard({ member, groupId, canRemove, isCurrentUser }: MemberCardProps) {
  const queryClient = useQueryClient();
  const toast = useToastContext();
  const [showRemoveConfirm, setShowRemoveConfirm] = useState(false);

  const remove = useMutation({
    mutationFn: () => removeMember(groupId, member.user_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
      setShowRemoveConfirm(false);
      toast.success("Member removed successfully");
    },
    onError: (error) => {
      const msg = error instanceof Error ? error.message : "Failed to remove member";
      toast.error(msg);
    },
  });

  const memberName = member.first_name && member.last_name
    ? `${member.first_name} ${member.last_name}`
    : "Unknown User";

  return (
    <>
      <div className="member-card surface card" style={{ boxShadow: "none" }}>
        <Link href={`/profile/${member.user_id}`} className="member-link">
          <Avatar
            src={member.avatar}
            name={memberName}
            size={40}
          />
          <div className="member-info">
            <strong className="member-name">{memberName}</strong>
            <span className="member-status">{member.status}</span>
            {isCurrentUser && <span className="member-you">(You)</span>}
          </div>
        </Link>
        {canRemove && member.can_remove_member && !isCurrentUser && (
          <Button
            variant="ghost"
            onClick={() => setShowRemoveConfirm(true)}
            className="member-remove"
          >
            Remove
          </Button>
        )}
      </div>

      <ConfirmDialog
        isOpen={showRemoveConfirm}
        onClose={() => setShowRemoveConfirm(false)}
        onConfirm={() => remove.mutate()}
        title="Remove Member"
        message={`Are you sure you want to remove ${memberName} from this group? They will need to request to join again.`}
        confirmText="Remove Member"
        cancelText="Cancel"
        type="warning"
        isLoading={remove.isPending}
      />
    </>
  );
}
