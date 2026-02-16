"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import { followUser, unfollowUser, getFollowStatus, ApiError } from "@/lib/api";

type FollowButtonProps = {
  userId: string;
  currentUserId?: string;
};

export default function FollowButton({ userId, currentUserId }: FollowButtonProps) {
  const queryClient = useQueryClient();

  // Get the follow status (none, following, requested, or self)
  const { data: statusData, isLoading } = useQuery({
    queryKey: ["followStatus", userId, currentUserId],
    queryFn: () => getFollowStatus(userId),
    enabled: !!currentUserId && currentUserId !== userId,
  });

  const status = statusData?.status;

  const [errorMsg, setErrorMsg] = useState("");

  const toggleFollow = useMutation({
    mutationFn: () => followUser(userId),
    onSuccess: () => {
      setErrorMsg("");
      queryClient.invalidateQueries({ queryKey: ["followStatus", userId, currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
      queryClient.invalidateQueries({ queryKey: ["userPosts", userId] });
    },
    onError: (error) => {
      setErrorMsg(error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message);
    },
  });

  if (!currentUserId || currentUserId === userId || status === "self") {
    return null;
  }

  if (isLoading) {
    return <Button variant="ghost" disabled>...</Button>;
  }

  if (status === "following") {
    return (
      <Button
        variant="ghost"
        onClick={() => toggleFollow.mutate()}
        disabled={toggleFollow.isPending}
        className="follow-btn following"
      >
        {toggleFollow.isPending ? "Unfollowing..." : "Following"}
      </Button>
    );
  }

  if (status === "requested") {
    return (
      <Button
        variant="ghost"
        onClick={() => toggleFollow.mutate()}
        disabled={toggleFollow.isPending}
        className="follow-btn requested"
      >
        {toggleFollow.isPending ? "Canceling..." : "Requested"}
      </Button>
    );
  }

  return (
    <div>
      <Button
        onClick={() => toggleFollow.mutate()}
        disabled={toggleFollow.isPending}
        className="follow-btn"
      >
        {toggleFollow.isPending ? "Following..." : "Follow"}
      </Button>
      {errorMsg && <p style={{ color: "#b42318", fontSize: "0.85rem", marginTop: 4 }}>{errorMsg}</p>}
    </div>
  );
}
