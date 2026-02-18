"use client";

import { useState, useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import MessageBubble from "@/components/MessageBubble";
import EmojiPicker from "@/components/EmojiPicker";
import Button from "@/components/Button";
import {getRoomMessages, sendMessage, ApiError, updateGroup, updateRoomName} from "@/lib/api";
import { useWebSocket } from "@/hooks/useWebSocket";
import type { ChatMessage } from "@/lib/types";
import FormField from "@/components/FormField";

type ChatRoomProps = {
  roomId: string;
  currentUserId: string;
  isGroup: boolean
  roomName?: string;
  canEdit: boolean;
};

export default function ChatRoom({ roomId, currentUserId, roomName, isGroup, canEdit }: ChatRoomProps) {
  const queryClient = useQueryClient();
  const [newMessage, setNewMessage] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const { isConnected, enterChat, leaveChat } = useWebSocket();
  const [editOpen, setEditOpen] = useState(false);
  const [editValue, setEditValue] = useState(roomName || "");
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

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

  const update = useMutation({
    mutationFn: () => updateRoomName(roomId, editValue),
    onSuccess: () => {
      setValidationErrors({});
      queryClient.invalidateQueries({ queryKey: ["chat-list"] });
      setEditOpen(false)
    },
    onError: (error) => {
      setValidationErrors({});
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      }
    },
  });

  const closeEditModal = () => {
    setEditOpen(false)
    setValidationErrors({})
  }

  return (
    <div className="chat-room">
      {roomName && (
          <div className="chat-room-title">
            <div className="chat-room-name">
              {roomName}
              {isConnected && <span style={{ fontSize: "0.75rem", color: "green", marginLeft: "8px" }}>●</span>}
            </div>
            {canEdit && (
                <div className="chat-room-title-edit">
                  <button
                      aria-label="Edit room name"
                      className="chat-room-title-edit-btn"
                      type="button"
                      onClick={() => {
                        setEditValue(roomName ?? "");
                        setEditOpen(true);
                      }}
                  >
                    <span style={{ fontSize: "1.25rem" }}>✏️</span>
                  </button>
                </div>
            )}

            {editOpen && (
                <div className="edit-roomname-drawer">
                  <form
                      onSubmit={e => {
                        e.preventDefault();
                        update.mutate()
                      }}
                      className="edit-roomname-form"
                  >
                    <div className="edit-roomname-field-wrapper">
                      <FormField
                          label="Room name"
                          value={editValue}
                          name="room_name"
                          autoFocus
                          onChange={e => setEditValue(e.target.value)}
                          placeholder="Edit room name"
                          className="edit-roomname-input"
                          error={validationErrors.room_name?.replace('_', ' ')}
                      />
                    </div>
                    <div className="edit-roomname-button-wrapper">
                      <button
                          type="submit"
                          disabled={update.isPending || roomName === editValue}
                          className="edit-roomname-submit"
                      >
                        Save
                      </button>
                      <button
                          type="button"
                          className="edit-roomname-cancel"
                          onClick={closeEditModal}
                      >
                        Cancel
                      </button>
                    </div>
                  </form>
                </div>
            )}
          </div>
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
              isGroup={isGroup}
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
