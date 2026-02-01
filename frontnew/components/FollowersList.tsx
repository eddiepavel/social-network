"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { getFollowers, getFollowing } from "@/lib/api";
import Avatar from "@/components/Avatar";
import EmptyState from "@/components/EmptyState";

type FollowersListProps = {
  userId: string;
  type: "followers" | "following";
};

export default function FollowersList({ userId, type }: FollowersListProps) {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: [type, userId],
    queryFn: () => (type === "followers" ? getFollowers(userId) : getFollowing(userId)),
    enabled: !!userId,
  });

  if (isLoading) return <p>Loading {type}...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data || data.length === 0) {
    return (
      <EmptyState
        title={type === "followers" ? "No followers yet" : "Not following anyone"}
        body={type === "followers" ? "When people follow this user, they'll appear here." : "Follow people to see them here."}
      />
    );
  }

  return (
    <div className="followers-list">
      {data.map((user) => (
        <Link
          key={user.user_id}
          href={`/profile/${user.user_id}`}
          className="follower-item"
        >
          <Avatar
            src={user.avatar}
            name={`${user.first_name} ${user.last_name}`}
            size={40}
          />
          <div className="follower-info">
            <span className="follower-name">
              {user.first_name} {user.last_name}
            </span>
            {user.nickname && (
              <span className="follower-nickname">@{user.nickname}</span>
            )}
          </div>
        </Link>
      ))}
    </div>
  );
}
