"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { createPost } from "@/lib/api";

export default function PostComposer() {
  const [content, setContent] = useState("");
  const [visibility, setVisibility] = useState("public");
  const queryClient = useQueryClient();

  const create = useMutation({
    mutationFn: createPost,
    onSuccess: () => {
      setContent("");
      queryClient.invalidateQueries({ queryKey: ["feed"] });
    },
  });

  return (
    <div className="surface card">
      <h3>Share a moment</h3>
      <FormField
        label="What is on your mind?"
        name="content"
        as="textarea"
        placeholder="Start a conversation, drop a quick update, or share a thought."
        value={content}
        onChange={(event) => setContent(event.target.value)}
      />
      <div className="split">
        <label className="form-field">
          <span>Visibility</span>
          <select
            name="visibility"
            value={visibility}
            onChange={(event) => setVisibility(event.target.value)}
          >
            <option value="public">Public</option>
            <option value="semi-private">Semi-private</option>
            <option value="private">Private</option>
          </select>
        </label>
        <div style={{ display: "flex", alignItems: "flex-end", justifyContent: "flex-end" }}>
          <Button
            onClick={() =>
              create.mutate({
                content,
                visibility,
              })
            }
            disabled={!content.trim() || create.isPending}
          >
            Post
          </Button>
        </div>
      </div>
      {create.isError ? (
        <p style={{ color: "#b42318" }}>{create.error.message}</p>
      ) : null}
    </div>
  );
}
