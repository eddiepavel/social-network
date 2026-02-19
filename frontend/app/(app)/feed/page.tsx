"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import PostComposer from "@/components/PostComposer";
import PostCard from "@/components/PostCard";
import SectionHeader from "@/components/SectionHeader";
import EmptyState from "@/components/EmptyState";
import { getFeed } from "@/lib/api";
import type { FeedPost, ApiEnvelope } from "@/lib/types";
import useSession from "@/hooks/useSession";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { useState } from "react";

export default function FeedPage() {
  const { data: session } = useSession();
  const [pageSize] = useState(10);
  const {
    data,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery<ApiEnvelope<FeedPost[]>, Error>({
    queryKey: ["feed"],
    queryFn: (context) => {
      const pageParam = typeof context.pageParam === "number" ? context.pageParam : 1;
      return getFeed(pageParam, pageSize);
    },
    getNextPageParam: (lastPage) => {
      const pagination = lastPage.pagination;
      if (!pagination) return undefined;
      if (pagination.page < pagination.total_pages) {
        return pagination.page + 1;
      }
      return undefined;
    },
    initialPageParam: 1,
  });

  const posts = data?.pages.flatMap((page) => page.data ?? []) ?? [];
  const lastElementRef = useInfiniteScroll(
    () => {
      if (hasNextPage && !isFetchingNextPage) fetchNextPage();
    },
    !!hasNextPage,
    isFetchingNextPage
  );

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
        {posts.map((post, idx) => {
          if (idx === posts.length - 1) {
            return (
              <div ref={lastElementRef} key={post.post_id}>
                <PostCard post={post as FeedPost} currentUserId={session?.user_id} />
              </div>
            );
          }
          return <PostCard key={post.post_id} post={post as FeedPost} currentUserId={session?.user_id} />;
        })}
        {isFetchingNextPage && <p>Loading more...</p>}
      </div>
    </div>
  );
}
