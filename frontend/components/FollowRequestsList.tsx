"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { getFollowRequests, respondToFollowRequest, ApiError } from "@/lib/api";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import EmptyState from "@/components/EmptyState";
import { formatDate } from "@/lib/utils";

export default function FollowRequestsList() {
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["follow-requests"],
    queryFn: getFollowRequests,
  });

  const respond = useMutation({
    mutationFn: ({ requestId, accept }: { requestId: string; accept: boolean }) =>
      respondToFollowRequest(requestId, accept),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["follow-requests"] });
      queryClient.invalidateQueries({ queryKey: ["followers"] });
    },
    onError: () => {},
  });

  if (isLoading) return <p>Loading requests...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data || data.length === 0) {
    return (
      <EmptyState
        title="No pending requests"
        body="When someone wants to follow you, their request will appear here."
      />
    );
  }

  return (
    <div className="follow-requests-list">
      {data.map((request) => (
        <div key={request.request_id} className="follow-request-item">
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
              onClick={() => respond.mutate({ requestId: request.request_id, accept: true })}
              disabled={respond.isPending}
            >
              Accept
            </Button>
            <Button
              variant="ghost"
              onClick={() => respond.mutate({ requestId: request.request_id, accept: false })}
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
