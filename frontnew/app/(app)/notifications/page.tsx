"use client";

import { useQuery } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import Tabs from "@/components/Tabs";
import FollowRequestsList from "@/components/FollowRequestsList";
import NotificationItem from "@/components/NotificationItem";
import EmptyState from "@/components/EmptyState";
import { getFollowRequests, getChatList } from "@/lib/api";
import { useState } from "react";

export default function NotificationsPage() {
  const [activeTab, setActiveTab] = useState("all");

  const { data: followRequests } = useQuery({
    queryKey: ["follow-requests"],
    queryFn: getFollowRequests,
  });

  const { data: chatList } = useQuery({
    queryKey: ["chat-list"],
    queryFn: getChatList,
  });

  const unreadChats = chatList?.filter((t) => t.unread_count > 0) ?? [];
  const followRequestCount = followRequests?.length ?? 0;
  const totalCount = followRequestCount + unreadChats.length;

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader title="Notifications" />
        <Tabs
          tabs={[
            { id: "all", label: `All (${totalCount})` },
            { id: "follow", label: `Follow Requests (${followRequestCount})` },
            { id: "messages", label: `Messages (${unreadChats.length})` },
          ]}
          activeTab={activeTab}
          onChange={setActiveTab}
        />

        {activeTab === "all" && (
          <div className="notifications-list">
            {followRequests?.map((request) => (
              <NotificationItem
                key={`follow-${request.request_id}`}
                type="follow_request"
                fromUserId={request.user_id}
                fromUserName={`${request.first_name} ${request.last_name}`}
                fromUserAvatar={request.avatar}
                createdAt={request.created_at}
              />
            ))}

            {unreadChats.map((chat) => (
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
              <EmptyState
                title="All caught up!"
                body="You have no new notifications."
              />
            )}
          </div>
        )}

        {activeTab === "follow" && <FollowRequestsList />}

        {activeTab === "messages" && (
          <div className="notifications-list">
            {unreadChats.length === 0 ? (
              <EmptyState
                title="No unread messages"
                body="You've read all your messages."
              />
            ) : (
              unreadChats.map((chat) => (
                <NotificationItem
                  key={`chat-${chat.room_id}`}
                  type="unread_message"
                  fromUserName={chat.room_name || "Chat"}
                  roomId={chat.room_id}
                  message={chat.last_message_content}
                  createdAt={chat.last_message_time || ""}
                  count={chat.unread_count}
                />
              ))
            )}
          </div>
        )}
      </section>
    </div>
  );
}
