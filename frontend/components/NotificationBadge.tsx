"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { getUnseenNotificationCount } from "@/lib/api";

export default function NotificationBadge() {
  const { data: unseenCount } = useQuery({
    queryKey: ["unseen-count"],
    queryFn: getUnseenNotificationCount,
    refetchInterval: 30000,
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
    </Link>
  );
}
