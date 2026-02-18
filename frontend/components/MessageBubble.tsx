import Avatar from "@/components/Avatar";
import { formatDate } from "@/lib/utils";
import type { ChatMessage } from "@/lib/types";

type MessageBubbleProps = {
  message: ChatMessage;
  isGroup: boolean;
  isOwn: boolean;
};

export default function MessageBubble({ message, isOwn, isGroup }: MessageBubbleProps) {
  const senderName = message.sender_first_name && message.sender_last_name
    ? `${message.sender_first_name} ${message.sender_last_name}`
    : "User";

  return (
    <div className={`message-bubble ${isOwn ? "own" : "other"}`}>
      {!isOwn && (
        <Avatar
          src={message.sender_avatar}
          name={senderName}
          size={32}
        />
      )}
      <div className="message-content">
        {!isOwn && isGroup && <span className="message-sender">{senderName}</span>}
        <p className="message-text">{message.content}</p>
        <span className="message-time">{formatDate(message.created_at)}</span>
      </div>
    </div>
  );
}
