"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import NotificationItem from "@/components/NotificationItem";
import { getFollowRequests, getChatList } from "@/lib/api";

export default function NotificationDropdown() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

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
  const unreadChats = chatList?.filter((t) => t.unread_count > 0) ?? [];
  const totalCount = followRequestCount + unreadChats.length;

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
      </button>

      {isOpen && (
        <div className="notification-menu">
          <div className="notification-header">
            <h4>Notifications</h4>
            <Link href="/notifications" onClick={() => setIsOpen(false)}>
              See all
            </Link>
          </div>

          <div className="notification-list">
            {followRequests?.slice(0, 3).map((request) => (
              <NotificationItem
                key={`follow-${request.request_id}`}
                type="follow_request"
                fromUserId={request.user_id}
                fromUserName={`${request.first_name} ${request.last_name}`}
                fromUserAvatar={request.avatar}
                createdAt={request.created_at}
              />
            ))}

            {unreadChats.slice(0, 3).map((chat) => (
              <NotificationItem
                key={`chat-${chat.room_id}`}
                type="unread_message"
                fromUserName={chat.room_name || "Chat"}
                roomId={chat.room_id}
                message={chat.last_message_content}
                createdAt={chat.last_message_time || ""}
                count={chat.unread_count}
              />
            ))}

            {totalCount === 0 && (
              <p className="notification-empty">No new notifications</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
