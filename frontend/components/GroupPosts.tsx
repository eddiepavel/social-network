"use client";

import PostCard from "@/components/PostCard";
import {useState} from "react";
import {useInfiniteQuery} from "@tanstack/react-query";
import type {ApiEnvelope, FeedPost, User} from "@/lib/types";
import { getGroupPosts } from "@/lib/api";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import SectionHeader from "@/components/SectionHeader";
import PostComposer from "@/components/PostComposer";
import EmptyState from "@/components/EmptyState";

type GroupPostsProps = {
    groupName: string;
    groupId: string;
    session: User | null | undefined;
};

export default function GroupPosts({ groupName, groupId, session }: GroupPostsProps) {
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
        queryKey: ["groupPosts", groupId],
        queryFn: (context) => {
            const pageParam = typeof context.pageParam === "number" ? context.pageParam : 1;
            return getGroupPosts(groupId, pageParam, pageSize);
        },
        getNextPageParam: (lastPage) => {
            const pagination = lastPage?.pagination;
            if (!pagination) return undefined;
            if (pagination.page < pagination.total_pages) {
                return pagination.page + 1;
            }
            return undefined;
        },
        initialPageParam: 1,
    });

    const posts = data?.pages.flatMap((page) => page?.data ?? []) ?? [];
    const lastElementRef = useInfiniteScroll(
        () => {
            if (hasNextPage && !isFetchingNextPage) fetchNextPage();
        },
        !!hasNextPage,
        isFetchingNextPage
    );

    return (
        <div className="split" style={{ paddingBottom: 24 }}>
            <div className="feed">
                <SectionHeader title={`${groupName}'s feed`} />
                {session?.user_id && <PostComposer groupId={groupId} />}
                {isLoading ? <p>Loading posts...</p> : null}
                {isError ? <p style={{ color: "#b42318" }}>{(error as Error).message}</p> : null}
                {posts.length === 0 && !isLoading ? (
                    <EmptyState
                        title="No posts yet"
                        body="Invite more people or drop the first update to kick off the feed."
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