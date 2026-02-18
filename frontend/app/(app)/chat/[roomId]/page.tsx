"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import ChatRoom from "@/components/ChatRoom";
import { getChatList } from "@/lib/api";
import useSession from "@/hooks/useSession";

export default function ChatRoomPage() {
  const params = useParams();
  const roomId = Array.isArray(params.roomId) ? params.roomId[0] : (params.roomId as string);
  const { data: session, isLoading: sessionLoading } = useSession();

  const { data: chatList } = useQuery({
    queryKey: ["chat-list"],
    queryFn: getChatList,
  });

  const room = chatList?.find((thread) => thread.room_id === roomId);

  if (sessionLoading) return <p>Loading...</p>;
  if (!session?.user_id) {
    return (
      <div className="surface card">
        <p>Please log in to view this chat.</p>
        <Link href="/login">Log in</Link>
      </div>
    );
  }

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <div className="chat-room-header">
        <Link href="/chat" className="back-link">
          ← Back to Inbox
        </Link>
      </div>
      <section className="surface card chat-room-container">
        <ChatRoom
          roomId={roomId}
          isGroup={room?.group_id !== ''}
          currentUserId={session.user_id}
          canEdit={room?.can_edit_room_name}
          roomName={room?.room_name || (room?.group_id !== '' ? "Group chat" : `${room?.other_user?.first_name} ${room?.other_user?.last_name}`)}
        />
      </section>
    </div>
  );
}
