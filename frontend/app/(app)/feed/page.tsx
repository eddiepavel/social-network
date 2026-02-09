"use client";

import { useQuery } from "@tanstack/react-query";
import PostComposer from "@/components/PostComposer";
import PostCard from "@/components/PostCard";
import SectionHeader from "@/components/SectionHeader";
import EmptyState from "@/components/EmptyState";
import { getFeed } from "@/lib/api";
import useSession from "@/hooks/useSession";

export default function FeedPage() {
  const { data: session } = useSession();
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["feed", 1],
    queryFn: () => getFeed(1, 10),
  });

  const posts = data?.data ?? [];

  return (
    <div className="split" style={{ paddingBottom: 64 }}>
      <div className="feed">
        <SectionHeader title="Your feed" />
        {session?.user_id && <PostComposer />}
        {isLoading ? <p>Loading posts...</p> : null}
        {isError ? <p style={{ color: "#b42318" }}>{(error as Error).message}</p> : null}
        {posts.length === 0 && !isLoading ? (
          <EmptyState
            title="No posts yet"
            body="Start following people or drop the first update to kick off the feed."
          />
        ) : null}
        {posts.map((post) => (
          <PostCard key={post.post_id} post={post} currentUserId={session?.user_id} />
        ))}
      </div>
    </div>
  );
}
