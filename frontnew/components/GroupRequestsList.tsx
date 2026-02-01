"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import EmptyState from "@/components/EmptyState";
import { getGroupRequests, respondToGroupRequest } from "@/lib/api";
import { formatDate } from "@/lib/utils";

type GroupRequestsListProps = {
  groupId: string;
};

export default function GroupRequestsList({ groupId }: GroupRequestsListProps) {
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["group-requests", groupId],
    queryFn: () => getGroupRequests(groupId),
    enabled: !!groupId,
  });

  const respond = useMutation({
    mutationFn: ({ userId, accept }: { userId: string; accept: boolean }) =>
      respondToGroupRequest(groupId, userId, accept),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group-requests", groupId] });
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
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
        <div key={request.request_id} className="group-request-item">
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
              onClick={() => respond.mutate({ userId: request.user_id, accept: true })}
              disabled={respond.isPending}
            >
              Accept
            </Button>
            <Button
              variant="ghost"
              onClick={() => respond.mutate({ userId: request.user_id, accept: false })}
              disabled={respond.isPending}
            >
              Decline
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
