"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import { followUser, unfollowUser, getFollowStatus, ApiError } from "@/lib/api";
import { useToastContext } from "../app/providers";
import { ConfirmDialog } from "@/components/ConfirmDialog";

type FollowButtonProps = {
  userId: string;
  currentUserId?: string;
};

export default function FollowButton({ userId, currentUserId }: FollowButtonProps) {
  const toast = useToastContext();
  const queryClient = useQueryClient();

  // Get the follow status (none, following, requested, or self)
  const { data: statusData, isLoading: statusLoading } = useQuery({
    queryKey: ["followStatus", userId, currentUserId],
    queryFn: () => getFollowStatus(userId),
    enabled: !!currentUserId && currentUserId !== userId,
  });

  const status = statusData?.status;

  const [errorMsg, setErrorMsg] = useState("");
  const [showUnfollowConfirm, setShowUnfollowConfirm] = useState(false);

  const followMutation = useMutation({
    mutationFn: () => followUser(userId),
    onSuccess: () => {
      setErrorMsg("");
      toast.success("Follow request sent!");
      queryClient.invalidateQueries({ queryKey: ["followStatus", userId, currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
      queryClient.invalidateQueries({ queryKey: ["userPosts", userId] });
    },
    onError: (error) => {
      const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message;
      setErrorMsg(msg);
      toast.error(msg);
    },
  });

  const unfollowMutation = useMutation({
    mutationFn: () => unfollowUser(userId),
    onSuccess: () => {
      setErrorMsg("");
      setShowUnfollowConfirm(false);
      toast.success("Successfully unfollowed user");
      queryClient.invalidateQueries({ queryKey: ["followStatus", userId, currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["following", currentUserId] });
      queryClient.invalidateQueries({ queryKey: ["followers", userId] });
      queryClient.invalidateQueries({ queryKey: ["userPosts", userId] });
    },
    onError: (error) => {
      const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message;
      setErrorMsg(msg);
      toast.error(msg);
    },
  });

  const handleFollowClick = () => {
    if (status === "following") {
      setShowUnfollowConfirm(true);
    } else if (status === "requested") {
      // Cancel request
      unfollowMutation.mutate();
    } else {
      followMutation.mutate();
    }
  };

  if (!currentUserId || currentUserId === userId || status === "self") {
    return null;
  }

  if (statusLoading) {
    return <Button variant="ghost" disabled>...</Button>;
  }

  const isLoading = followMutation.isPending || unfollowMutation.isPending;

  let buttonText = "Follow";
  if (isLoading) buttonText = "Processing...";
  else if (status === "following") buttonText = "Following";
  else if (status === "requested") buttonText = "Requested";

  return (
    <>
      <div>
        <Button
          variant={status === "following" || status === "requested" ? "ghost" : "solid"}
          onClick={handleFollowClick}
          disabled={isLoading}
          className={`follow-btn ${status === "following" ? "following" : ""} ${status === "requested" ? "requested" : ""}`}
        >
          {buttonText}
        </Button>
        {errorMsg && <p style={{ color: "#b42318", fontSize: "0.85rem", marginTop: 4 }}>{errorMsg}</p>}
      </div>

      <ConfirmDialog
        isOpen={showUnfollowConfirm}
        onClose={() => setShowUnfollowConfirm(false)}
        onConfirm={() => unfollowMutation.mutate()}
        title="Unfollow User"
        message="Are you sure you want to unfollow this user? You will need to send a new follow request to follow them again."
        confirmText="Unfollow"
        cancelText="Cancel"
        type="warning"
        isLoading={unfollowMutation.isPending}
      />
    </>
  );
}
