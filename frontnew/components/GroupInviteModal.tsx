"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Modal from "@/components/Modal";
import Button from "@/components/Button";
import Avatar from "@/components/Avatar";
import { searchUsers, inviteToGroup, ApiError } from "@/lib/api";
import type { SearchUser } from "@/lib/types";

type GroupInviteModalProps = {
  isOpen: boolean;
  onClose: () => void;
  groupId: string;
  existingMemberIds: string[];
};

export default function GroupInviteModal({
  isOpen,
  onClose,
  groupId,
  existingMemberIds,
}: GroupInviteModalProps) {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchUser[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [invitedIds, setInvitedIds] = useState<Set<string>>(new Set());

  const [errorMsg, setErrorMsg] = useState("");

  const invite = useMutation({
    mutationFn: (userId: string) => inviteToGroup(groupId, userId),
    onSuccess: (_, userId) => {
      setErrorMsg("");
      setInvitedIds((prev) => new Set(prev).add(userId));
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
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
      setSearchResults(results.filter((u) => !existingMemberIds.includes(u.user_id)));
    } catch (error) {
      console.error("Search failed:", error);
    } finally {
      setIsSearching(false);
    }
  };

  const handleClose = () => {
    setSearchQuery("");
    setSearchResults([]);
    setInvitedIds(new Set());
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Invite to Group">
      <div className="group-invite-search">
        <input
          type="text"
          className="group-invite-input"
          placeholder="Search for users..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
        />
        <Button onClick={handleSearch} disabled={isSearching}>
          Search
        </Button>
      </div>

      {errorMsg && <p style={{ color: "#b42318", fontSize: "0.85rem", padding: "0 4px" }}>{errorMsg}</p>}
      <div className="group-invite-results">
        {isSearching ? (
          <p>Searching...</p>
        ) : searchResults.length === 0 ? (
          <p className="group-invite-hint">Search for users to invite them to the group</p>
        ) : (
          searchResults.map((user) => {
            const isInvited = invitedIds.has(user.user_id);
            return (
              <div key={user.user_id} className="group-invite-user">
                <Avatar
                  name={`${user.first_name} ${user.last_name}`}
                  size={40}
                />
                <div className="group-invite-user-info">
                  <span className="group-invite-user-name">
                    {user.first_name} {user.last_name}
                  </span>
                  {user.nickname && (
                    <span className="group-invite-user-nickname">@{user.nickname}</span>
                  )}
                </div>
                <Button
                  variant={isInvited ? "ghost" : "solid"}
                  onClick={() => invite.mutate(user.user_id)}
                  disabled={invite.isPending || isInvited}
                >
                  {isInvited ? "Invited" : "Invite"}
                </Button>
              </div>
            );
          })
        )}
      </div>
    </Modal>
  );
}
