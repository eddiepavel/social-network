"use client";

import { useQuery } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import Tabs from "@/components/Tabs";
import FollowRequestsList from "@/components/FollowRequestsList";
import NotificationItem from "@/components/NotificationItem";
import EmptyState from "@/components/EmptyState";
import { getFollowRequests, getChatList, getNotifications, getUnseenNotificationCount } from "@/lib/api";
import { useState, useMemo } from "react";

export default function NotificationsPage() {
  const [activeTab, setActiveTab] = useState("all");

  const { data: notifications } = useQuery({
    queryKey: ["notifications"],
    queryFn: getNotifications,
  });

  const { data: followRequests } = useQuery({
    queryKey: ["follow-requests"],
    queryFn: getFollowRequests,
  });

  const { data: chatList } = useQuery({
    queryKey: ["chat-list"],
    queryFn: getChatList,
  });

  const { data: unseenCount } = useQuery({
    queryKey: ["unseen-count"],
    queryFn: getUnseenNotificationCount,
  });

  // Filter notifications by type
  const followNotifications = useMemo(() => {
    return notifications?.filter(n => 
      n.type === "follow_request" || n.type === "follow_accepted"
    ) ?? [];
  }, [notifications]);

  const groupNotifications = useMemo(() => {
    return notifications?.filter(n => 
      n.type === "group_invitation" || 
      n.type === "group_request" || 
      n.type === "group_join_approved" || 
      n.type === "group_join_rejected" ||
      n.type === "group_event"
    ) ?? [];
  }, [notifications]);

  const postNotifications = useMemo(() => {
    return notifications?.filter(n => 
      n.type === "post_comment" || 
      n.type === "comment_reply" || 
      n.type === "post_reaction" ||
      n.type === "comment_reaction"
    ) ?? [];
  }, [notifications]);

  // Create a map of from_id to request_id for follow requests
  const followRequestMap = useMemo(() => {
    const map = new Map<string, string>();
    followRequests?.forEach(req => {
      map.set(req.user_id, req.request_id);
    });
    return map;
  }, [followRequests]);

  const unreadChats = chatList?.filter((t) => t.unread_count > 0) ?? [];
  const totalUnseenCount = unseenCount?.count ?? 0;

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader title="Notifications" />
        <Tabs
          tabs={[
            { id: "all", label: `All ${totalUnseenCount > 0 ? `(${totalUnseenCount})` : ""}` },
            { id: "follow", label: "Follow" },
            { id: "groups", label: "Groups" },
            { id: "posts", label: "Posts" },
            { id: "messages", label: `Messages (${unreadChats.length})` },
          ]}
          activeTab={activeTab}
          onChange={setActiveTab}
        />

        {activeTab === "all" && (
          <div className="notifications-list">
            {notifications && notifications.length > 0 ? (
              notifications.map((notification) => (
                <NotificationItem
                  key={notification.notif_id}
                  notification={notification}
                  requestId={followRequestMap.get(notification.from_id)}
                />
              ))
            ) : (
              <EmptyState
                title="All caught up!"
                body="You have no new notifications."
              />
            )}
          </div>
        )}

        {activeTab === "follow" && (
          <div className="notifications-list">
            {followNotifications.length === 0 ? (
              <EmptyState
                title="No follow notifications"
                body="Follow notifications will appear here."
              />
            ) : (
              followNotifications.map((notification) => (
                <NotificationItem
                  key={notification.notif_id}
                  notification={notification}
                  requestId={followRequestMap.get(notification.from_id)}
                />
              ))
            )}
          </div>
        )}

        {activeTab === "groups" && (
          <div className="notifications-list">
            {groupNotifications.length === 0 ? (
              <EmptyState
                title="No group notifications"
                body="Group notifications will appear here."
              />
            ) : (
              groupNotifications.map((notification) => (
                <NotificationItem
                  key={notification.notif_id}
                  notification={notification}
                />
              ))
            )}
          </div>
        )}

        {activeTab === "posts" && (
          <div className="notifications-list">
            {postNotifications.length === 0 ? (
              <EmptyState
                title="No post notifications"
                body="Post notifications will appear here."
              />
            ) : (
              postNotifications.map((notification) => (
                <NotificationItem
                  key={notification.notif_id}
                  notification={notification}
                />
              ))
            )}
          </div>
        )}

        {activeTab === "messages" && (
          <div className="notifications-list">
            {unreadChats.length === 0 ? (
              <EmptyState
                title="No unread messages"
                body="You've read all your messages."
              />
            ) : (
              unreadChats.map((chat) => {
                // Create a notification-like object for unread messages
                const messageNotification = {
                  notif_id: `chat-${chat.room_id}`,
                  receiver_id: "",
                  type: "message" as const,
                  is_seen: false,
                  from_id: chat.room_id,
                  from_name: chat.room_name || "Chat",
                  from_avatar: undefined,
                  created_at: chat.last_message_time || "",
                };
                return (
                  <NotificationItem
                    key={`chat-${chat.room_id}`}
                    notification={messageNotification}
                  />
                );
              })
            )}
          </div>
        )}
      </section>
    </div>
  );
}
