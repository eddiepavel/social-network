import { formatDate } from "@/lib/utils";
import type { GroupEvent } from "@/lib/types";
import Button from "@/components/Button";
import Modal from "@/components/Modal";
import Tabs from "@/components/Tabs";
import {rsvpToEvent, ApiError, getEventRSVP} from "@/lib/api";
import { useState } from "react";
import {useQuery, useQueryClient} from "@tanstack/react-query";
import Avatar from "@/components/Avatar";
import { useToastContext } from "../app/providers";
import Link from "next/link";

type EventCardProps = {
  event: GroupEvent;
  onRsvpUpdate?: (eventId: string, status: string) => void;
};

export default function EventCard({ event, onRsvpUpdate }: EventCardProps) {
  const queryClient = useQueryClient();
  const toast = useToastContext();

  const [currentRsvp, setCurrentRsvp] = useState(event.user_rsvp);
  const [goingCount, setGoingCount] = useState(event.going_count);
  const [notGoingCount, setNotGoingCount] = useState(event.not_going_count);
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [activeTab, setActiveTab] = useState("going");

  const { data, isLoading: eventsLoading, isError, error } = useQuery({
    queryKey: ["group_event", event.event_id],
    queryFn: () => getEventRSVP(event.event_id),
  });

  const goingUsers = data?.filter(r => r.status === "going") || [];
  const notGoingUsers = data?.filter(r => r.status === "not going") || [];

  const tabs = [
    { id: "going", label: `Going (${goingUsers.length})` },
    { id: "not-going", label: `Not Going (${notGoingUsers.length})` }
  ];

  const handleRsvp = async (status: "going" | "not going") => {
    setIsLoading(true);
    setErrorMsg("");
    try {
      await rsvpToEvent(event.event_id, status);

      // Update counts based on previous and new status
      if (currentRsvp === "going") setGoingCount(c => c - 1);
      if (currentRsvp === "not going") setNotGoingCount(c => c - 1);

      if (status === "going") setGoingCount(c => c + 1);
      if (status === "not going") setNotGoingCount(c => c + 1);

      setCurrentRsvp(status);
      onRsvpUpdate?.(event.event_id, status);
      toast.success(`You are ${status === "going" ? "going" : "not going"} to this event`);

      queryClient.invalidateQueries({ queryKey: ["group_event", event.event_id] });
    } catch (error) {
      const msg = error instanceof ApiError && typeof error.details === 'string' ? error.details : error instanceof Error ? error.message : "Failed to update RSVP";
      setErrorMsg(msg);
      toast.error(msg);
    } finally {
      setIsLoading(false);
    }
  };

  if (eventsLoading) return <p>Loading group event...</p>;


  return (
    <div className="event-card surface card" style={{ 
      boxShadow: "0 2px 8px rgba(0,0,0,0.08)", 
      borderRadius: "12px",
      padding: "1.25rem",
      display: "flex",
      flexDirection: "row",
      gap: "1.5rem",
      alignItems: "stretch"
    }}>
      {/* Left Section - Event Details */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: "0.75rem" }}>
        <h3 style={{ 
          margin: 0, 
          fontSize: "1.25rem", 
          fontWeight: 600,
          color: "var(--text-primary)"
        }}>
          {event.event_name}
        </h3>
        
        <p style={{ 
          margin: 0, 
          color: "var(--text-secondary)", 
          fontSize: "0.95rem",
          lineHeight: 1.5,
          flex: 1
        }}>
          {event.description}
        </p>
        
        <div style={{ 
          display: "flex", 
          alignItems: "center", 
          gap: "0.5rem",
          color: "var(--text-tertiary)",
          fontSize: "0.85rem"
        }}>
          <span>👤</span>
          <span>Created by <strong style={{ color: "var(--text-secondary)" }}>{event.creator.first_name} {event.creator.last_name}</strong></span>
        </div>

        <Button
          variant="ghost"
          onClick={() => setIsModalOpen(true)}
          style={{ alignSelf: "flex-start", marginTop: "0.5rem" }}
        >
          👥 View Attendees
        </Button>
      </div>

      {/* Right Section - Time & RSVP */}
      <div style={{ 
        display: "flex", 
        flexDirection: "column", 
        alignItems: "flex-end",
        justifyContent: "space-between",
        minWidth: "160px",
        gap: "0.75rem"
      }}>
        <div style={{ 
          background: "var(--bg-tertiary)", 
          padding: "0.5rem 0.75rem", 
          borderRadius: "8px",
          textAlign: "end"
        }}>
          <div style={{ fontSize: "0.75rem", color: "var(--text-tertiary)", textTransform: "uppercase" }}>
            Event Time
          </div>
          <div style={{ fontSize: "0.95rem", fontWeight: 600, color: "var(--text-primary)" }}>
            {formatDate(event.timestamp)}
          </div>
        </div>

        <div style={{ 
          display: "flex", 
          gap: "1rem",
          padding: "0 0.75rem",
          fontSize: "1.1rem",
          color: "var(--text-secondary)"
        }}>
          <span style={{ display: "flex", alignItems: "center", gap: "0.25rem" }}>
            <span style={{ color: "#22c55e" }}>✓</span> {goingCount}
          </span>
          <span style={{ display: "flex", alignItems: "center", gap: "0.25rem" }}>
            <span style={{ color: "#ef4444" }}>✗</span> {notGoingCount}
          </span>
        </div>

        {errorMsg && <p style={{ color: "#b42318", fontSize: "0.8rem", margin: 0 }}>{errorMsg}</p>}
        
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <Button
            variant={currentRsvp === "going" ? "solid" : "ghost"}
            onClick={() => handleRsvp("going")}
            disabled={isLoading}
          >
            Going
          </Button>
          <Button
            variant={currentRsvp === "not going" ? "solid" : "ghost"}
            onClick={() => handleRsvp("not going")}
            disabled={isLoading}
          >
            Not Going
          </Button>
        </div>
      </div>

      {/* Attendees Modal */}
      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} title="Event Attendees">
        <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />
        
        <div style={{ marginTop: "1rem", minHeight: "150px" }}>
          {activeTab === "going" && (
            <div style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
              {goingUsers.length === 0 ? (
                <p style={{ color: "var(--text-tertiary)", textAlign: "center" }}>No one is going yet</p>
              ) : (
                goingUsers.map(user => (
                  <div key={user.user_id} style={{ 
                    display: "flex", 
                    alignItems: "center", 
                    gap: "0.75rem",
                    padding: "0.5rem",
                    borderRadius: "8px",
                    background: "var(--bg-tertiary)"
                  }}>
                    <Avatar
                        src={user.avatar}
                        name={user.first_name + " " + user.last_name}
                        size={40}
                    />
                    <span style={{ fontWeight: 500 }}>{user.first_name} {user.last_name}</span>
                  </div>
                ))
              )}
            </div>
          )}
          
          {activeTab === "not-going" && (
            <div style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
              {notGoingUsers.length === 0 ? (
                <p style={{ color: "var(--text-tertiary)", textAlign: "center" }}>No one has declined yet</p>
              ) : (
                notGoingUsers.map(user => (
                  <div key={user.user_id} style={{ 
                    display: "flex", 
                    alignItems: "center", 
                    gap: "0.75rem",
                    padding: "0.5rem",
                    borderRadius: "8px",
                    background: "var(--bg-tertiary)"
                  }}>
                    <Avatar
                      src={user.avatar}
                      name={user.first_name + " " + user.last_name}
                      size={40}
                    />
                    <span style={{ fontWeight: 500 }}>{user.first_name} {user.last_name}</span>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
}
