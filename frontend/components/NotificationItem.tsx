"use client";

import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Avatar from "@/components/Avatar";
import Button from "@/components/Button";
import { formatDate } from "@/lib/utils";
import { respondToFollowRequest, markNotificationAsSeen, ApiError } from "@/lib/api";
import type { Notification } from "@/lib/types";
import { useRouter } from "next/navigation";

type NotificationItemProps = {
  notification: Notification;
  compact?: boolean;
  requestId?: string; // For follow_request notifications
};

export default function NotificationItem({
  notification,
  compact = false,
  requestId,
}: NotificationItemProps) {
  const queryClient = useQueryClient();
  const router = useRouter();

  const respond = useMutation({
    mutationFn: ({ reqId, status }: { reqId: string; status: "accepted" | "rejected" }) =>
      respondToFollowRequest(reqId, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["follow-requests"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["followers"] });
    },
  });

  const markAsSeen = useMutation({
    mutationFn: () => markNotificationAsSeen(notification.notif_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["unseen-count"] });
    },
  });

  const getNotificationContent = () => {
    switch (notification.type) {
      case "follow_request":
        return {
          href: `/profile/${notification.from_id}`,
          text: "wants to follow you",
          icon: "👤",
          showActions: true,
        };
      case "follow_accepted":
        return {
          href: `/profile/${notification.from_id}`,
          text: "accepted your follow request",
          icon: "✅",
          showActions: false,
        };
      case "group_invitation":
        return {
          href: `/groups/${notification.group_id}`,
          text: "invited you to a group",
          icon: "👥",
          showActions: false,
        };
      case "group_request":
        return {
          href: `/groups/${notification.group_id}`,
          text: "requested to join your group",
          icon: "📩",
          showActions: false,
        };
      case "group_join_approved":
        return {
          href: `/groups/${notification.group_id}`,
          text: "approved your group join request",
          icon: "✅",
          showActions: false,
        };
      case "group_join_rejected":
        return {
          href: `/groups/${notification.group_id}`,
          text: "rejected your group join request",
          icon: "❌",
          showActions: false,
        };
      case "group_event":
        return {
          href: `/groups/${notification.group_id}`,
          text: "created a new event",
          icon: "📅",
          showActions: false,
        };
      case "post_comment":
        return {
          href: `/post/${notification.from_id}`,
          text: "commented on your post",
          icon: "💬",
          showActions: false,
        };
      case "comment_reply":
        return {
          href: `/post/${notification.from_id}`,
          text: "replied to your comment",
          icon: "↩️",
          showActions: false,
        };
      case "post_reaction":
        return {
          href: `/post/${notification.from_id}`,
          text: "reacted to your post",
          icon: "❤️",
          showActions: false,
        };
      case "comment_reaction":
        return {
          href: `/post/${notification.from_id}`,
          text: "reacted to your comment",
          icon: "❤️",
          showActions: false,
        };
      case "message":
        return {
          href: `/chat`,
          text: "sent you a message",
          icon: "💬",
          showActions: false,
        };
      default:
        return {
          href: "#",
          text: "new notification",
          icon: "🔔",
          showActions: false,
        };
    }
  };

  const content = getNotificationContent();

  const handleClick = () => {
    if (!notification.is_seen) {
      markAsSeen.mutate();
    }
  };

  return (
    <div
      className={`notification-item ${!notification.is_seen ? "unread" : ""}`}
      style={{ opacity: notification.is_seen ? 0.7 : 1 }}
    >
      <Avatar
        src={notification.from_avatar}
        name={notification.from_name}
        size={40}
      />
      <Link href={content.href} onClick={handleClick} className="notification-content-link">
        <div className="notification-content">
          <p className="notification-text">
            <strong>{notification.from_nickname || notification.from_name}</strong> {content.text}
          </p>
          <span className="notification-time">{formatDate(notification.created_at)}</span>
        </div>
      </Link>
      <span className="notification-icon">{content.icon}</span>
      
      {content.showActions && requestId && notification.type === "follow_request" && (
        <div className="notification-actions">
          <Button
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              respond.mutate({ reqId: requestId, status: "accepted" });
            }}
            disabled={respond.isPending}
          >
            Accept
          </Button>
          <Button
            variant="ghost"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              respond.mutate({ reqId: requestId, status: "rejected" });
            }}
            disabled={respond.isPending}
          >
            Decline
          </Button>
        </div>
      )}
      {respond.isError && (
        <p style={{ color: "#b42318", fontSize: "0.85rem", marginTop: "4px", width: "100%" }}>
          {respond.error instanceof ApiError && typeof respond.error.details === 'string'
            ? respond.error.details
            : respond.error.message}
        </p>
      )}
    </div>
  );
}
