"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import EmptyState from "@/components/EmptyState";
import { respondToGroupRequest, ApiError } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import {GroupJoinRequest} from "@/lib/types";
import { useToastContext } from "../app/providers";

type GroupRequestsListProps = {
  groupId: string;
  data: GroupJoinRequest[] | undefined
  isLoading?: boolean
  isError?: boolean
  error?: Error | null
};

export default function GroupRequestsList({ groupId, isError, isLoading, error, data }: GroupRequestsListProps) {
  const queryClient = useQueryClient();
  const toast = useToastContext();

  const respond = useMutation({
    mutationFn: ({ user_id, response }: { user_id: string; response: string }) =>
      respondToGroupRequest(groupId, user_id, response),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["group-requests", groupId] });
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
      toast.success(variables.response === "approve" ? "Member request accepted" : "Member request rejected");
    },
    onError: (error) => {
      const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message;
      toast.error(msg);
    },
  });

  if (isLoading) return <p>Loading requests...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data || data.length === 0) {
    return (
      <EmptyState
        title="No pending requests"
        body="When users request to join this group, they'll appear here."
      />
    );
  }
  
  return (
    <div className="group-requests-list">
      {data.map((request) => (
        <div key={request.user_id} className="group-request-item">
          <Link href={`/profile/${request.user_id}`} className="request-user">
            <Avatar
              src={request.avatar}
              name={`${request.first_name} ${request.last_name}`}
              size={40}
            />
            <div className="request-info">
              <span className="request-name">
                {request.first_name} {request.last_name}
              </span>
              <span className="request-time">{formatDate(request.created_at)}</span>
            </div>
          </Link>
          <div className="request-actions">
            <Button
              onClick={() => respond.mutate({ user_id: request.user_id, response: "approve" })}
              disabled={respond.isPending}
            >
              Accept
            </Button>
            <Button
              variant="ghost"
              onClick={() => respond.mutate({ user_id: request.user_id, response: "reject"})}
              disabled={respond.isPending}
            >
              Decline
            </Button>
            {respond.isError ? (
              <p style={{ color: "#b42318", fontSize: "0.85rem" }}>
                {respond.error instanceof ApiError && typeof respond.error.details === 'string'
                  ? respond.error.details
                  : respond.error.message}
              </p>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}
