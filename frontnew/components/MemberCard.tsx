"use client";

import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import { removeMember } from "@/lib/api";
import type { GroupMember } from "@/lib/types";

type MemberCardProps = {
  member: GroupMember;
  groupId: string;
  canRemove: boolean;
  isCurrentUser: boolean;
};

export default function MemberCard({ member, groupId, canRemove, isCurrentUser }: MemberCardProps) {
  const queryClient = useQueryClient();

  const remove = useMutation({
    mutationFn: () => removeMember(groupId, member.user_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
    },
  });

  const memberName = member.first_name && member.last_name
    ? `${member.first_name} ${member.last_name}`
    : "Unknown User";

  return (
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
          onClick={() => remove.mutate()}
          disabled={remove.isPending}
          className="member-remove"
        >
          Remove
        </Button>
      )}
    </div>
  );
}
