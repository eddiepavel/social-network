"use client";

import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toggleReaction, ApiError } from "@/lib/api";

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
  const [localReacted, setLocalReacted] = useState(userReacted);
  const [localCount, setLocalCount] = useState(reactionCount);

  useEffect(() => {
    setLocalReacted(userReacted);
    setLocalCount(reactionCount);
  }, [userReacted, reactionCount]);

  const toggle = useMutation({
    mutationFn: () => toggleReaction(postId),
    onMutate: () => {
      setLocalReacted((prev) => !prev);
      setLocalCount((prev) => (localReacted ? prev - 1 : prev + 1));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
    onError: () => {
      setLocalReacted(userReacted);
      setLocalCount(reactionCount);
    },
  });

  return (
      <button
          className={`reaction-btn ${localReacted ? "reacted" : ""}`}
          onClick={() => toggle.mutate()}
          disabled={toggle.isPending}
      >
        <span className="reaction-icon">{localReacted ? "❤️" : "🤍"}</span>
        <span className="reaction-count">{localCount}</span>
      </button>
  );
}