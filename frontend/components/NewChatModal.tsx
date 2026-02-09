"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Modal from "@/components/Modal";
import Button from "@/components/Button";
import Avatar from "@/components/Avatar";
import { searchUsers, startNewChat, ApiError } from "@/lib/api";
import type { SearchUser } from "@/lib/types";

type NewChatModalProps = {
  isOpen: boolean;
  onClose: () => void;
  currentUserId: string;
};

export default function NewChatModal({ isOpen, onClose, currentUserId }: NewChatModalProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchUser[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const [errorMsg, setErrorMsg] = useState("");

  const startChat = useMutation({
    mutationFn: (userId: string) => startNewChat(userId),
    onSuccess: (data) => {
      setErrorMsg("");
      queryClient.invalidateQueries({ queryKey: ["chat-list"] });
      onClose();
      router.push(`/chat/${data.room_id}`);
    },
    onError: (error) => {
      setErrorMsg(error instanceof ApiError && typeof error.details === 'string' ? error.details : error.message);
    },
  });

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    setIsSearching(true);
    try {
      const results = await searchUsers(searchQuery);
      setSearchResults(results.filter((u) => u.user_id !== currentUserId));
    } catch (error) {
      console.error("Search failed:", error);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="New Chat">
      <div className="new-chat-search">
        <input
          type="text"
          className="new-chat-input"
          placeholder="Search for a user..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
        />
        <Button onClick={handleSearch} disabled={isSearching}>
          Search
        </Button>
      </div>

      {errorMsg && <p style={{ color: "#b42318", fontSize: "0.85rem", padding: "0 4px" }}>{errorMsg}</p>}
      <div className="new-chat-results">
        {isSearching ? (
          <p>Searching...</p>
        ) : searchResults.length === 0 ? (
          <p className="new-chat-hint">Search for users to start a conversation</p>
        ) : (
          searchResults.map((user) => (
            <button
              key={user.user_id}
              className="new-chat-user"
              onClick={() => startChat.mutate(user.user_id)}
              disabled={startChat.isPending}
            >
              <Avatar
                name={`${user.first_name} ${user.last_name}`}
                size={40}
              />
              <div className="new-chat-user-info">
                <span className="new-chat-user-name">
                  {user.first_name} {user.last_name}
                </span>
                {user.nickname && (
                  <span className="new-chat-user-nickname">@{user.nickname}</span>
                )}
              </div>
            </button>
          ))
        )}
      </div>
    </Modal>
  );
}
