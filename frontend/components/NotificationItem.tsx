"use client";

import Link from "next/link";
import Avatar from "@/components/Avatar";
import { formatDate } from "@/lib/utils";

type NotificationItemProps = {
  type: "follow_request" | "unread_message";
  fromUserId?: string;
  fromUserName: string;
  fromUserAvatar?: string;
  roomId?: string;
  message?: string;
  createdAt: string;
  count?: number;
};

export default function NotificationItem({
  type,
  fromUserId,
  fromUserName,
  fromUserAvatar,
  roomId,
  message,
  createdAt,
  count,
}: NotificationItemProps) {
  const getNotificationContent = () => {
    switch (type) {
      case "follow_request":
        return {
          href: `/profile/${fromUserId}`,
          text: "wants to follow you",
          icon: "👤",
        };
      case "unread_message":
        return {
          href: `/chat/${roomId}`,
          text: count && count > 1 ? `${count} unread messages` : message || "sent you a message",
          icon: "💬",
        };
      default:
        return {
          href: "#",
          text: "New notification",
          icon: "🔔",
        };
    }
  };

  const content = getNotificationContent();

  return (
    <Link href={content.href} className="notification-item">
      <Avatar src={fromUserAvatar} name={fromUserName} size={40} />
      <div className="notification-content">
        <p className="notification-text">
          <strong>{fromUserName}</strong> {content.text}
        </p>
        <span className="notification-time">{formatDate(createdAt)}</span>
      </div>
      <span className="notification-icon">{content.icon}</span>
    </Link>
  );
}
