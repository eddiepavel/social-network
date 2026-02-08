"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import { followUser, unfollowUser, getFollowing, ApiError } from "@/lib/api";

type FollowButtonProps = {
  userId: string;
  currentUserId?: string;
};

export default function FollowButton({ userId, currentUserId }: FollowButtonProps) {
  const queryClient = useQueryClient();

  // Check if current user is following this user by checking the following list
  const { data: following, isLoading } = useQuery({
    queryKey: ["following", currentUserId],
    queryFn: () => getFollowing(currentUserId!),
    enabled: !!currentUserId && currentUserId !== userId,
  });

  const isFollowing = following?.some(f => f.user_id === userId) ?? false;

  const [errorMsg, setErrorMsg] = useState("");

  const follow = useMutation({
    mutationFn: () => followUser(userId),
    onSuccess: () => {
      setErrorMsg("");
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
    },
    onError: (error) => {
      setErrorMsg(error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message);
    },
  });

  const unfollow = useMutation({
    mutationFn: () => unfollowUser(userId),
    onSuccess: () => {
      setErrorMsg("");
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
    },
    onError: (error) => {
      setErrorMsg(error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message);
    },
  });

  if (!currentUserId || currentUserId === userId) {
    return null;
  }

  if (isLoading) {
    return <Button variant="ghost" disabled>...</Button>;
  }

  if (isFollowing) {
    return (
      <Button
        variant="ghost"
        onClick={() => unfollow.mutate()}
        disabled={unfollow.isPending}
        className="follow-btn following"
      >
        Following
      </Button>
    );
  }

  return (
    <div>
      <Button
        onClick={() => follow.mutate()}
        disabled={follow.isPending}
        className="follow-btn"
      >
        Follow
      </Button>
      {errorMsg && <p style={{ color: "#b42318", fontSize: "0.85rem" }}>{errorMsg}</p>}
    </div>
  );
}
