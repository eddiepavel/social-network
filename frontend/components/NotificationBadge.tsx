"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { getUnseenNotificationCount } from "@/lib/api";
import { useWebSocket } from "@/hooks/useWebSocket";

export default function NotificationBadge() {
  const { isConnected } = useWebSocket();

  // Fetch count only on mount - real-time updates come via WebSocket
  const { data: unseenCount } = useQuery({
    queryKey: ["unseen-count"],
    queryFn: getUnseenNotificationCount,
    staleTime: Infinity, // Don't auto-refetch
  });

  const totalCount = unseenCount?.count ?? 0;

  return (
    <Link href="/notifications" className="notification-badge-link">
      <span className="notification-bell">🔔</span>
      {totalCount > 0 && (
        <span className="notification-count">
          {totalCount > 99 ? "99+" : totalCount}
        </span>
      )}
      {isConnected && (
        <span className="ws-status connected" title="Real-time connected" />
      )}
    </Link>
  );
}
