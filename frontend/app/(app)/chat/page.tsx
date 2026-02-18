"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import EmptyState from "@/components/EmptyState";
import Button from "@/components/Button";
import NewChatModal from "@/components/NewChatModal";
import { getChatList } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import useSession from "@/hooks/useSession";

export default function ChatPage() {
  const { data: session } = useSession();
  const [isNewChatOpen, setIsNewChatOpen] = useState(false);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["chat-list"],
    queryFn: getChatList,
    refetchInterval: 1000000,
  });

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader
          title="Inbox"
          action={
            session?.user_id && (
              <Button onClick={() => setIsNewChatOpen(true)}>New Chat</Button>
            )
          }
        />
        {isLoading ? <p>Loading chats...</p> : null}
        {isError ? <p style={{ color: "#b42318" }}>{(error as Error).message}</p> : null}
        {!isLoading && data?.length === 0 ? (
          <EmptyState
            title="No conversations yet"
            body="Start following people or join groups to unlock messaging."
          />
        ) : null}
        <div className="chat-threads">
          {data?.map((thread) => (
            <Link
              key={thread.room_id}
              href={`/chat/${thread.room_id}`}
              className="chat-thread"
            >
              <div className="chat-thread-info">
                <strong className="chat-thread-name">
                  {thread.room_name || (thread.is_group ? "Group chat" : `${thread.other_user?.first_name} ${thread.other_user?.last_name}`)}
                </strong>
                <p className="chat-thread-preview">
                  {session?.user_id === thread.last_message_sender?.user_id ? 'You' : thread.is_group ? `${thread.last_message_sender?.first_name} ${thread.last_message_sender?.last_name}` : 'Them'}: {thread.last_message_content || "No messages yet"}
                </p>
              </div>
              <div className="chat-thread-meta">
                {thread.unread_count > 0 && (
                  <span className="chat-unread-badge">{thread.unread_count}</span>
                )}
                <span className="chat-thread-time">
                  {thread.last_message_time ? formatDate(thread.last_message_time) : ""}
                </span>
              </div>
            </Link>
          ))}
        </div>
      </section>

      {session?.user_id && (
        <NewChatModal
          isOpen={isNewChatOpen}
          onClose={() => setIsNewChatOpen(false)}
          currentUserId={session.user_id}
        />
      )}
    </div>
  );
}
