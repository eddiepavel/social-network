"use client";

import { useState, useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import MessageBubble from "@/components/MessageBubble";
import EmojiPicker from "@/components/EmojiPicker";
import Button from "@/components/Button";
import { getRoomMessages, sendMessage } from "@/lib/api";

type ChatRoomProps = {
  roomId: string;
  currentUserId: string;
  roomName?: string;
};

const POLL_INTERVAL = 3000;

export default function ChatRoom({ roomId, currentUserId, roomName }: ChatRoomProps) {
  const queryClient = useQueryClient();
  const [newMessage, setNewMessage] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { data: messages, isLoading } = useQuery({
    queryKey: ["chat-messages", roomId],
    queryFn: () => getRoomMessages(roomId),
    refetchInterval: POLL_INTERVAL,
    enabled: !!roomId,
  });

  const send = useMutation({
    mutationFn: () => sendMessage(roomId, newMessage),
    onSuccess: () => {
      setNewMessage("");
      queryClient.invalidateQueries({ queryKey: ["chat-messages", roomId] });
      queryClient.invalidateQueries({ queryKey: ["chat-list"] });
    },
  });

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (newMessage.trim()) {
      send.mutate();
    }
  };

  const handleEmojiSelect = (emoji: string) => {
    setNewMessage((prev) => prev + emoji);
  };

  return (
    <div className="chat-room">
      {roomName && <h3 className="chat-room-title">{roomName}</h3>}

      <div className="chat-messages">
        {isLoading ? (
          <p className="chat-loading">Loading messages...</p>
        ) : messages?.length === 0 ? (
          <p className="chat-empty">No messages yet. Start the conversation!</p>
        ) : (
          messages?.map((message) => (
            <MessageBubble
              key={message.message_id}
              message={message}
              isOwn={message.sender_id === currentUserId}
            />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      <form className="chat-input-form" onSubmit={handleSubmit}>
        <EmojiPicker onSelect={handleEmojiSelect} />
        <input
          type="text"
          className="chat-input"
          placeholder="Type a message..."
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
        />
        <Button type="submit" disabled={!newMessage.trim() || send.isPending}>
          Send
        </Button>
      </form>
    </div>
  );
}
