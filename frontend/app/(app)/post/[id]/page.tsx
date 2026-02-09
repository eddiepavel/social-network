"use client";

import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import PostCard from "@/components/PostCard";
import { getPost } from "@/lib/api";
import useSession from "@/hooks/useSession";

export default function PostDetailPage() {
  const params = useParams();
  const postId = Array.isArray(params.id) ? params.id[0] : (params.id as string);
  const { data: session } = useSession();

  const { data: post, isLoading, isError, error } = useQuery({
    queryKey: ["post", postId],
    queryFn: () => getPost(postId),
    enabled: !!postId,
  });

  if (isLoading) return <p>Loading post...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!post) return <p>Post not found.</p>;

  return (
    <div className="grid" style={{ paddingBottom: 64, maxWidth: 700 }}>
      <PostCard
        post={post}
        currentUserId={session?.user_id}
        showFullComments
      />
    </div>
  );
}
