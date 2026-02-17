"use client";

import { useState, useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import MessageBubble from "@/components/MessageBubble";
import EmojiPicker from "@/components/EmojiPicker";
import Button from "@/components/Button";
import { getRoomMessages, sendMessage, ApiError } from "@/lib/api";
import { useWebSocket } from "@/hooks/useWebSocket";
import type { ChatMessage } from "@/lib/types";

type ChatRoomProps = {
  roomId: string;
  currentUserId: string;
  roomName?: string;
};

export default function ChatRoom({ roomId, currentUserId, roomName }: ChatRoomProps) {
  const queryClient = useQueryClient();
  const [newMessage, setNewMessage] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const { isConnected, enterChat, leaveChat } = useWebSocket();

  // Notify backend when entering/leaving chat room
  useEffect(() => {
    if (isConnected && roomId) {
      enterChat(roomId);
    }
    return () => {
      if (isConnected) {
        leaveChat();
      }
    };
  }, [isConnected, roomId, enterChat, leaveChat]);

  const { data: messages, isLoading } = useQuery({
    queryKey: ["chat-messages", roomId],
    queryFn: () => getRoomMessages(roomId),
    enabled: !!roomId,
  });

  // The WebSocket provider now handles chat_message events globally
  // No need for local WebSocket setup - messages are updated via query cache

  const send = useMutation({
    mutationFn: async () => {
      // Always use REST API to send - it's more reliable and the backend
      // will broadcast the message to all participants via WebSocket
      return sendMessage(roomId, newMessage);
    },
    onMutate: async () => {
      // Optimistically add the message to the cache
      const optimisticMsg: ChatMessage = {
        message_id: `temp-${Date.now()}`,
        room_id: roomId,
        content: newMessage,
        sender_id: currentUserId,
        created_at: new Date().toISOString(),
      };

      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["chat-messages", roomId] });

      // Snapshot previous value
      const previousMessages = queryClient.getQueryData<ChatMessage[]>(["chat-messages", roomId]);

      // Optimistically update
      queryClient.setQueryData(["chat-messages", roomId], (old: ChatMessage[] | undefined) => {
        if (!old) return [optimisticMsg];
        return [optimisticMsg, ...old];
      });

      return { previousMessages };
    },
    onSuccess: () => {
      setNewMessage("");
      // Invalidate chat-list for the preview update
      queryClient.invalidateQueries({ queryKey: ["chat-list"] });
      // Also invalidate chat-messages to get the real message with proper ID
      queryClient.invalidateQueries({ queryKey: ["chat-messages", roomId] });
    },
    onError: (_err, _vars, context) => {
      // Rollback on error
      if (context?.previousMessages) {
        queryClient.setQueryData(["chat-messages", roomId], context.previousMessages);
      }
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
      {roomName && (
        <h3 className="chat-room-title">
          {roomName}
          {isConnected && <span style={{ fontSize: "0.75rem", color: "green", marginLeft: "8px" }}>●</span>}
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
