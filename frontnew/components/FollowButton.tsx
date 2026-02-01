"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import { followUser, unfollowUser, getFollowStatus } from "@/lib/api";
import type { FollowStatus } from "@/lib/types";

type FollowButtonProps = {
  userId: string;
  currentUserId?: string;
};

export default function FollowButton({ userId, currentUserId }: FollowButtonProps) {
  const queryClient = useQueryClient();

  const { data: statusData, isLoading } = useQuery({
    queryKey: ["follow-status", userId],
    queryFn: () => getFollowStatus(userId),
    enabled: !!currentUserId && currentUserId !== userId,
  });

  const status: FollowStatus = statusData?.status ?? "none";

  const follow = useMutation({
    mutationFn: () => followUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["follow-status", userId] });
      queryClient.invalidateQueries({ queryKey: ["followers"] });
      queryClient.invalidateQueries({ queryKey: ["following"] });
    },
  });

  const unfollow = useMutation({
    mutationFn: () => unfollowUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["follow-status", userId] });
      queryClient.invalidateQueries({ queryKey: ["followers"] });
      queryClient.invalidateQueries({ queryKey: ["following"] });
    },
  });

  if (!currentUserId || currentUserId === userId) {
    return null;
  }

  if (isLoading) {
    return <Button variant="ghost" disabled>...</Button>;
  }

  if (status === "following") {
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

  if (status === "requested") {
    return (
      <Button
        variant="ghost"
        onClick={() => unfollow.mutate()}
        disabled={unfollow.isPending}
        className="follow-btn requested"
      >
        Requested
      </Button>
    );
  }

  return (
    <Button
      onClick={() => follow.mutate()}
      disabled={follow.isPending}
      className="follow-btn"
    >
      Follow
    </Button>
  );
}
