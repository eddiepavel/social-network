import EventCard from "@/components/EventCard";
import EmptyState from "@/components/EmptyState";
import type { GroupEvent } from "@/lib/types";

type EventsListProps = {
  events: GroupEvent[];
};

export default function EventsList({ events }: EventsListProps) {
  if (!events || events.length === 0) {
    return (
      <EmptyState
        title="No events yet"
        body="Events for this group will appear here once they are created."
      />
    );
  }

  return (
    <div className="events-list">
      {events.map((event) => (
        <EventCard key={event.event_id} event={event} />
      ))}
    </div>
  );
}
