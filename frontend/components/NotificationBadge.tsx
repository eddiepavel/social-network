"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { getFollowRequests, getChatList } from "@/lib/api";

export default function NotificationBadge() {
  const { data: followRequests } = useQuery({
    queryKey: ["follow-requests"],
    queryFn: getFollowRequests,
    refetchInterval: 30000,
  });

  const { data: chatList } = useQuery({
    queryKey: ["chat-list"],
    queryFn: getChatList,
    refetchInterval: 10000,
  });

  const followRequestCount = followRequests?.length ?? 0;
  const unreadMessages = chatList?.reduce((sum, thread) => sum + thread.unread_count, 0) ?? 0;
  const totalCount = followRequestCount + unreadMessages;

  return (
    <Link href="/notifications" className="notification-badge-link">
      <span className="notification-bell">🔔</span>
      {totalCount > 0 && (
        <span className="notification-count">
          {totalCount > 99 ? "99+" : totalCount}
        </span>
      )}
    </Link>
  );
}
