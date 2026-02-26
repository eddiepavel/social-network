"use client";

import {useMutation, useQueryClient} from "@tanstack/react-query";
import Link from "next/link";
import { respondToFollowRequest, ApiError} from "@/lib/api";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import EmptyState from "@/components/EmptyState";
import {formatDate} from "@/lib/utils";
import {FollowRequest} from "@/lib/types";
import { useToastContext } from "../app/providers";

type FollowRequestsListProps = {
    data?: FollowRequest[] | undefined
    isLoading?: boolean
    isError?: boolean
    error?: Error | null
};

export default function FollowRequestsList({data, isLoading, isError, error}: FollowRequestsListProps) {
    const queryClient = useQueryClient();
    const toast = useToastContext();

    const respond = useMutation({
        mutationFn: ({requestId, status}: { requestId: string; status: "accepted" | "rejected" }) =>
            respondToFollowRequest(requestId, status),
        onSuccess: (_, variables) => {
            queryClient.invalidateQueries({queryKey: ["follow-requests"]});
            queryClient.invalidateQueries({queryKey: ["followers"]});
            toast.success(variables.status === "accepted" ? "Follow request accepted" : "Follow request rejected");
        },
        onError: (error) => {
            const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message;
            toast.error(msg);
        },
    });

    if (isLoading) return <p>Loading requests...</p>;
    if (isError) return <p style={{color: "#b42318"}}>{(error as Error).message}</p>;
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
                            onClick={() => respond.mutate({requestId: request.request_id, status: "accepted"})}
                            disabled={respond.isPending}
                        >
                            Accept
                        </Button>
                        <Button
                            variant="ghost"
                            onClick={() => respond.mutate({requestId: request.request_id, status: "rejected"})}
                            disabled={respond.isPending}
                        >
                            Decline
                        </Button>
                        {respond.isError ? (
                            <p style={{color: "#b42318", fontSize: "0.85rem"}}>
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
