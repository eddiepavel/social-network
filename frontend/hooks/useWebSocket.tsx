"use client";

import { createContext, useContext, useEffect, useRef, useState, useCallback, ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import useSession from "@/hooks/useSession";
import type { Notification, ChatMessage } from "@/lib/types";

const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_BASE?.replace(/\/$/, "") || "ws://localhost:8000";

type WebSocketContextType = {
  isConnected: boolean;
  sendMessage: (roomId: string, content: string) => void;
  enterChat: (roomId: string) => void;
  leaveChat: () => void;
};

const WebSocketContext = createContext<WebSocketContextType>({
  isConnected: false,
  sendMessage: () => {},
  enterChat: () => {},
  leaveChat: () => {},
});

export function useWebSocket() {
  return useContext(WebSocketContext);
}

type WebSocketProviderProps = {
  children: ReactNode;
};

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const { data: session } = useSession();
  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleNotification = useCallback((notifData: any) => {
    // Create notification object matching frontend type
    const notification: Notification = {
      notif_id: notifData.notif_id,
      receiver_id: notifData.receiver_id,
      type: notifData.type,
      is_seen: notifData.is_seen || false,
      from_id: notifData.from_id,
      from_name: notifData.from_name || "",
      from_avatar: notifData.from_avatar,
      from_nickname: notifData.from_nickname,
      group_id: notifData.group_id,
      event_id: notifData.event_id,
      created_at: notifData.created_at,
    };

    // Update notifications cache - add to beginning
    queryClient.setQueryData(["notifications"], (old: Notification[] | undefined) => {
      if (!old) return [notification];
      // Avoid duplicates
      const exists = old.some((n) => n.notif_id === notification.notif_id);
      if (exists) return old;
      return [notification, ...old];
    });

    // Update unseen count
    queryClient.setQueryData(["unseen-count"], (old: { count: number } | undefined) => {
      const currentCount = old?.count ?? 0;
      return { count: currentCount + 1 };
    });

    // Invalidate follow-requests if it's a follow request notification
    if (notification.type === "follow_request") {
      queryClient.invalidateQueries({ queryKey: ["follow-requests"] });
    }

    // Invalidate relevant group queries for group notifications
    if (
      notification.type === "group_invitation" ||
      notification.type === "group_request" ||
      notification.type === "group_join_approved" ||
      notification.type === "group_join_rejected" ||
      notification.type === "group_event"
    ) {
      queryClient.invalidateQueries({ queryKey: ["groups"] });
      if (notification.group_id) {
        queryClient.invalidateQueries({ queryKey: ["group", notification.group_id] });
      }
    }
  }, [queryClient]);

  const handleChatMessage = useCallback((msgData: any) => {
    const roomId = msgData.room_id;
    if (!roomId) return; // Safety check
    
    const newMsg: ChatMessage = {
      message_id: msgData.message_id,
      room_id: roomId,
      content: msgData.content,
      sender_id: msgData.sender_id,
      sender_first_name: msgData.sender_first_name,
      sender_last_name: msgData.sender_last_name,
      sender_avatar: msgData.sender_avatar,
      created_at: msgData.created_at || new Date().toISOString(),
    };

    // Update chat messages cache for the room
    queryClient.setQueryData(["chat-messages", roomId], (old: ChatMessage[] | undefined) => {
      if (!old) return [newMsg];
      // Avoid duplicates
      const exists = old.some((m) => m.message_id === newMsg.message_id);
      if (exists) return old;
      return [newMsg, ...old];
    });

    // Also refresh the chat list
    queryClient.invalidateQueries({ queryKey: ["chat-list"] });
  }, [queryClient]);

  const createConnection = useCallback(() => {
    if (!session?.user_id) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(`${WS_BASE_URL}/ws/connect`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("[WebSocket] Connected");
      setIsConnected(true);
      reconnectAttemptsRef.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        const payload = message.payload
          ? typeof message.payload === "string"
            ? JSON.parse(message.payload)
            : message.payload
          : message.data;

        switch (message.type) {
          case "notification":
            handleNotification(payload);
            break;
          case "chat_message":
            handleChatMessage(payload);
            break;
          case "private_message":
            handleChatMessage(payload);
            break;
        }
      } catch (err) {
        console.error("[WebSocket] Failed to parse message:", err);
      }
    };

    ws.onclose = () => {
      console.log("[WebSocket] Disconnected");
      setIsConnected(false);
      // Use inline reconnect logic to avoid circular dependency
      if (reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log(`[WebSocket] Attempting to reconnect (${reconnectAttemptsRef.current}/${maxReconnectAttempts})...`);
          // Re-call connection function - use ref to get latest version
          if (session?.user_id && (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN)) {
            const newWs = new WebSocket(`${WS_BASE_URL}/ws/connect`);
            wsRef.current = newWs;
            // Copy the handlers
            newWs.onopen = ws.onopen;
            newWs.onmessage = ws.onmessage;
            newWs.onclose = ws.onclose;
            newWs.onerror = ws.onerror;
          }
        }, delay);
      } else {
        console.error("[WebSocket] Max reconnect attempts reached");
      }
    };

    ws.onerror = (error) => {
      console.error("[WebSocket] Error:", error);
    };
  }, [session?.user_id, handleNotification, handleChatMessage]);

  const sendMessage = useCallback((roomId: string, content: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: "send_message",
        payload: JSON.stringify({ room_id: roomId, content }),
      }));
    } else {
      console.warn("[WebSocket] Not connected, cannot send message");
    }
  }, []);

  const enterChat = useCallback((roomId: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: "enter_chat",
        payload: JSON.stringify({ room_id: roomId }),
      }));
      console.log("[WebSocket] Entered chat room:", roomId);
    }
  }, []);

  const leaveChat = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: "leave_chat",
        payload: "{}",
      }));
      console.log("[WebSocket] Left chat room");
    }
  }, []);

  // Connect when user is logged in
  useEffect(() => {
    if (session?.user_id) {
      createConnection();
    }

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        reconnectAttemptsRef.current = maxReconnectAttempts; // Prevent reconnection
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [session?.user_id, createConnection]);

  return (
    <WebSocketContext.Provider value={{ isConnected, sendMessage, enterChat, leaveChat }}>
      {children}
    </WebSocketContext.Provider>
  );
}
