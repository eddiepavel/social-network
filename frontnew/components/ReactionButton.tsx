"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toggleReaction } from "@/lib/api";

type ReactionButtonProps = {
  postId: string;
  reactionCount: number;
  userReacted: boolean;
};

export default function ReactionButton({
  postId,
  reactionCount,
  userReacted,
}: ReactionButtonProps) {
  const queryClient = useQueryClient();

  const toggle = useMutation({
    mutationFn: () => toggleReaction(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
    },
  });

  return (
    <button
      className={`reaction-btn ${userReacted ? "reacted" : ""}`}
      onClick={() => toggle.mutate()}
      disabled={toggle.isPending}
    >
      <span className="reaction-icon">{userReacted ? "❤️" : "🤍"}</span>
      <span className="reaction-count">{reactionCount}</span>
    </button>
  );
}
