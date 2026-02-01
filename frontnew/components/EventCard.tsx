import { formatDate } from "@/lib/utils";
import type { GroupEvent } from "@/lib/types";
import Button from "@/components/Button";

type EventCardProps = {
  event: GroupEvent;
};

export default function EventCard({ event }: EventCardProps) {
  return (
    <div className="event-card surface card" style={{ boxShadow: "none" }}>
      <div className="event-header">
        <h4 className="event-title">{event.title}</h4>
        <span className="event-date">{formatDate(event.event_date)}</span>
      </div>

      <p className="event-description">{event.description}</p>

      {event.location && (
        <p className="event-location">
          <span className="event-location-icon">📍</span>
          {event.location}
        </p>
      )}

      <div className="event-stats">
        <span className="event-stat">
          <span className="event-stat-icon">✅</span>
          {event.going_count ?? 0} going
        </span>
        <span className="event-stat">
          <span className="event-stat-icon">❌</span>
          {event.not_going_count ?? 0} not going
        </span>
      </div>

      <div className="event-actions">
        <Button variant={event.user_rsvp === "going" ? "solid" : "ghost"} disabled>
          Going
        </Button>
        <Button variant={event.user_rsvp === "not_going" ? "solid" : "ghost"} disabled>
          Not Going
        </Button>
      </div>

      <p className="event-note" style={{ fontSize: "0.8rem", color: "var(--muted)", marginTop: 8 }}>
        RSVP functionality coming soon
      </p>
    </div>
  );
}
