"use client";

import { useState } from "react";
import Modal from "@/components/Modal";
import FormField from "@/components/FormField";
import Button from "@/components/Button";
import { createGroupEvent, ApiError } from "@/lib/api";
import type { CreateEventRequest, GroupEvent } from "@/lib/types";

type CreateEventModalProps = {
  isOpen: boolean;
  onClose: () => void;
  groupId: string;
  onEventCreated?: (event: GroupEvent) => void;
};

export default function CreateEventModal({
  isOpen,
  onClose,
  groupId,
  onEventCreated,
}: CreateEventModalProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [time, setTime] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const resetForm = () => {
    setTitle("");
    setDescription("");
    setDate("");
    setTime("");
    setError("");
    setFieldErrors({});
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setFieldErrors({});

    // Validate fields
    const errors: Record<string, string> = {};
    if (!title.trim()) errors.title = "Title is required";
    if (!description.trim()) errors.description = "Description is required";
    if (!date) errors.date = "Date is required";
    if (!time) errors.time = "Time is required";

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }

    // Combine date and time into ISO 8601 timestamp
    const timestamp = new Date(`${date}T${time}`).toISOString();

    const payload: CreateEventRequest = {
      title: title.trim(),
      description: description.trim(),
      timestamp,
    };

    setIsSubmitting(true);
    try {
      const event = await createGroupEvent(groupId, payload);
      onEventCreated?.(event);
      handleClose();
    } catch (err) {
      if (err instanceof ApiError) {
        if (typeof err.details === "object") {
          setFieldErrors(err.details);
        } else {
          setError(err.details || err.message);
        }
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to create event");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  // Get minimum date (today)
  const today = new Date().toISOString().split("T")[0];

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Create Event">
      <form onSubmit={handleSubmit} className="form-stack">
        <FormField
          label="Event Title"
          name="title"
          type="text"
          placeholder="Enter event title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          error={fieldErrors.title}
          required
        />

        <FormField
          label="Description"
          name="description"
          as="textarea"
          placeholder="Describe your event..."
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          error={fieldErrors.description}
          rows={4}
          required
        />

        <FormField
          label="Date"
          name="date"
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          error={fieldErrors.date}
          min={today}
          required
        />

        <FormField
          label="Time"
          name="time"
          type="time"
          value={time}
          onChange={(e) => setTime(e.target.value)}
          error={fieldErrors.time}
          required
        />

        {error && (
          <p style={{ color: "#b42318", fontSize: "0.85rem", margin: 0 }}>
            {error}
          </p>
        )}

        <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
          <Button type="button" variant="ghost" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Creating..." : "Create Event"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
