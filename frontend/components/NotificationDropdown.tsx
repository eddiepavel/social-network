"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import NotificationItem from "@/components/NotificationItem";
import { getFollowRequests, getNotifications, getUnseenNotificationCount, markAllNotificationsAsSeen } from "@/lib/api";
import { useWebSocket } from "@/hooks/useWebSocket";
import Button from "@/components/Button";

export default function NotificationDropdown() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();
  const { isConnected } = useWebSocket();

  // Fetch notifications only on mount and when manually triggered
  // Real-time updates come via WebSocket
  const { data: notifications, isLoading: notificationsLoading } = useQuery({
    queryKey: ["notifications"],
    queryFn: getNotifications,
    staleTime: Infinity, // Don't refetch automatically
  });

  const { data: followRequests } = useQuery({
    queryKey: ["follow-requests"],
    queryFn: getFollowRequests,
    staleTime: Infinity, // Don't refetch automatically
  });

  const { data: unseenCount } = useQuery({
    queryKey: ["unseen-count"],
    queryFn: getUnseenNotificationCount,
    staleTime: Infinity, // Don't refetch automatically
  });

  const markAllSeen = useMutation({
    mutationFn: markAllNotificationsAsSeen,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.setQueryData(["unseen-count"], { count: 0 });
    },
  });

  // Create a map of from_id to request_id for follow requests
  const followRequestMap = new Map<string, string>();
  followRequests?.forEach(req => {
    followRequestMap.set(req.user_id, req.request_id);
  });

  // Get unseen notifications
  const unseenNotifications = notifications?.filter(n => !n.is_seen) ?? [];
  const totalCount = unseenCount?.count ?? 0;

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  return (
    <div className="notification-dropdown" ref={dropdownRef}>
      <button
        className="notification-trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-label="Notifications"
      >
        <span className="notification-bell">🔔</span>
        {totalCount > 0 && (
          <span className="notification-count">
            {totalCount > 99 ? "99+" : totalCount}
          </span>
        )}
        {isConnected && (
          <span className="ws-status connected" title="Real-time connected" />
        )}
      </button>

      {isOpen && (
        <div className="notification-menu">
          <div className="notification-header">
            <h4>Notifications</h4>
            <div className="notification-header-actions">
              {totalCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => markAllSeen.mutate()}
                  disabled={markAllSeen.isPending}
                >
                  Mark all read
                </Button>
              )}
              <Link href="/notifications" onClick={() => setIsOpen(false)}>
                See all
              </Link>
            </div>
          </div>

          <div className="notification-list">
            {notificationsLoading ? (
              <p className="notification-empty">Loading...</p>
            ) : unseenNotifications.length === 0 ? (
              <p className="notification-empty">No new notifications</p>
            ) : (
              unseenNotifications.slice(0, 5).map((notification) => (
                <NotificationItem
                  key={notification.notif_id}
                  notification={notification}
                  requestId={followRequestMap.get(notification.from_id)}
                  compact={true}
                  onAction={() => setIsOpen(false)}
                />
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
