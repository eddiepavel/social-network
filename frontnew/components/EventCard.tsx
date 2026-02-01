import { formatDate } from "@/lib/utils";
import type { GroupEvent } from "@/lib/types";
import Button from "@/components/Button";
import { rsvpToEvent } from "@/lib/api";
import { useState } from "react";

type EventCardProps = {
  event: GroupEvent;
  onRsvpUpdate?: (eventId: string, status: string) => void;
};

export default function EventCard({ event, onRsvpUpdate }: EventCardProps) {
  const [currentRsvp, setCurrentRsvp] = useState(event.user_rsvp);
  const [goingCount, setGoingCount] = useState(event.going_count);
  const [notGoingCount, setNotGoingCount] = useState(event.not_going_count);
  const [isLoading, setIsLoading] = useState(false);

  const handleRsvp = async (status: "going" | "not going") => {
    setIsLoading(true);
    try {
      await rsvpToEvent(event.event_id, status);

      // Update counts based on previous and new status
      if (currentRsvp === "going") setGoingCount(c => c - 1);
      if (currentRsvp === "not going") setNotGoingCount(c => c - 1);

      if (status === "going") setGoingCount(c => c + 1);
      if (status === "not going") setNotGoingCount(c => c + 1);

      setCurrentRsvp(status);
      onRsvpUpdate?.(event.event_id, status);
    } catch (error) {
      console.error("Failed to RSVP:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="event-card surface card" style={{ boxShadow: "none" }}>
      <div className="event-header">
        <h4 className="event-title">{event.event_name}</h4>
        <span className="event-date">{formatDate(event.timestamp)}</span>
      </div>

      <p className="event-description">{event.description}</p>

      <div className="event-stats">
        <span className="event-stat">
          <span className="event-stat-icon">✅</span>
          {goingCount} going
        </span>
        <span className="event-stat">
          <span className="event-stat-icon">❌</span>
          {notGoingCount} not going
        </span>
      </div>

      <div className="event-actions">
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
  );
}
