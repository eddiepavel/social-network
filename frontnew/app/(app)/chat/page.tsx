"use client";

import { useQuery } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import EmptyState from "@/components/EmptyState";
import { getChatList } from "@/lib/api";
import { formatDate } from "@/lib/utils";

export default function ChatPage() {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["chat-list"],
    queryFn: getChatList,
  });

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader title="Inbox" />
        {isLoading ? <p>Loading chats...</p> : null}
        {isError ? <p style={{ color: "#b42318" }}>{(error as Error).message}</p> : null}
        {!isLoading && data?.length === 0 ? (
          <EmptyState
            title="No conversations yet"
            body="Start following people or join groups to unlock messaging."
          />
        ) : null}
        <div className="grid">
          {data?.map((thread) => (
            <div key={thread.room_id} className="surface card" style={{ boxShadow: "none" }}>
              <strong>{thread.room_name || (thread.is_group ? "Group chat" : "Direct chat")}</strong>
              <p style={{ color: "var(--muted)", margin: 0 }}>
                {thread.last_message_content || "No messages yet"}
              </p>
              <div className="post-meta">
                <span>{thread.unread_count} unread</span>
                <span>{thread.last_message_time ? formatDate(thread.last_message_time) : ""}</span>
              </div>
            </div>
          ))}
        </div>
      </section>
      <section className="surface card">
        <SectionHeader title="Live chat" />
        <p style={{ color: "var(--muted)" }}>
          WebSocket chat is available at <code>/ws/chat</code> once the backend is running.
          Hook a client to receive real-time messages.
        </p>
      </section>
    </div>
  );
}
