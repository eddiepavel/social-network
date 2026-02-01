"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import { followUser, unfollowUser, getFollowing } from "@/lib/api";

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

  const follow = useMutation({
    mutationFn: () => followUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
    },
  });

  const unfollow = useMutation({
    mutationFn: () => unfollowUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
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
    <Button
      onClick={() => follow.mutate()}
      disabled={follow.isPending}
      className="follow-btn"
    >
      Follow
    </Button>
  );
}
