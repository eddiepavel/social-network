"use client";

import { useState, useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import MessageBubble from "@/components/MessageBubble";
import EmojiPicker from "@/components/EmojiPicker";
import Button from "@/components/Button";
import { getRoomMessages, sendMessage, getChatWebSocket, ApiError } from "@/lib/api";
import type { ChatMessage, WSMessage } from "@/lib/types";

type ChatRoomProps = {
  roomId: string;
  currentUserId: string;
  roomName?: string;
};

export default function ChatRoom({ roomId, currentUserId, roomName }: ChatRoomProps) {
  const queryClient = useQueryClient();
  const [newMessage, setNewMessage] = useState("");
  const [wsConnected, setWsConnected] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<ReturnType<typeof getChatWebSocket> | null>(null);

  const { data: messages, isLoading } = useQuery({
    queryKey: ["chat-messages", roomId],
    queryFn: () => getRoomMessages(roomId),
    enabled: !!roomId,
  });

  // Initialize WebSocket connection
  useEffect(() => {
    const ws = getChatWebSocket();
    wsRef.current = ws;

    ws.connect({
      onOpen: () => {
        setWsConnected(true);
      },
      onClose: () => {
        setWsConnected(false);
      },
      onMessage: (wsMessage: WSMessage) => {
        // Handle incoming private messages for this room
        if (wsMessage.type === "private_message") {
          const msgData = wsMessage.data as any;
          // Only add message if it's for this room
          if (msgData && msgData.room_id === roomId) {
            queryClient.setQueryData(["chat-messages", roomId], (old: ChatMessage[] | undefined) => {
              if (!old) return old;
              const newMsg: ChatMessage = {
                message_id: msgData.message_id,
                room_id: roomId,
                content: msgData.content,
                sender_id: msgData.sender_id,
                created_at: msgData.created_at || new Date().toISOString(),
              };
              // Avoid duplicates
              const exists = old.some((m) => m.message_id === newMsg.message_id);
              if (exists) return old;
              return [newMsg, ...old];
            });
          }
        }
      },
    });

    return () => {
      // Cleanup on unmount
    };
  }, [roomId, queryClient]);

  const send = useMutation({
    mutationFn: () => sendMessage(roomId, newMessage),
    onSuccess: () => {
      setNewMessage("");
      // Invalidate to refetch and ensure consistency
      queryClient.invalidateQueries({ queryKey: ["chat-messages", roomId] });
      queryClient.invalidateQueries({ queryKey: ["chat-list"] });
    },
    onError: () => {},
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
      {roomName && (
        <h3 className="chat-room-title">
          {roomName}
          {wsConnected && <span style={{ fontSize: "0.75rem", color: "green", marginLeft: "8px" }}>●</span>}
        </h3>
      )}

      <div className="chat-messages">
        {isLoading ? (
          <p className="chat-loading">Loading messages...</p>
        ) : !messages || messages.length === 0 ? (
          <p className="chat-empty">No messages yet. Start the conversation!</p>
        ) : (
          [...messages].reverse().map((message) => (
            <MessageBubble
              key={message.message_id}
              message={message}
              isOwn={message.sender_id === currentUserId}
            />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {send.isError ? (
        <p style={{ color: "#b42318", fontSize: "0.85rem", padding: "0 12px" }}>
          {send.error instanceof ApiError && typeof send.error.details === 'string'
            ? send.error.details
            : send.error.message}
        </p>
      ) : null}
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
