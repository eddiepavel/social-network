"use client";

import { useState } from "react";

type EmojiPickerProps = {
  onSelect: (emoji: string) => void;
};

const EMOJI_CATEGORIES = {
  smileys: ["😀", "😃", "😄", "😁", "😅", "😂", "🤣", "😊", "😇", "🙂", "😉", "😌", "😍", "🥰", "😘"],
  gestures: ["👍", "👎", "👌", "✌️", "🤞", "🤟", "🤙", "👋", "🤚", "🖐️", "✋", "👏", "🙌", "🤝", "🙏"],
  hearts: ["❤️", "🧡", "💛", "💚", "💙", "💜", "🖤", "🤍", "💔", "❣️", "💕", "💞", "💓", "💗", "💖"],
  objects: ["🎉", "🎊", "🎁", "🎈", "✨", "🔥", "⭐", "🌟", "💫", "💥", "💯", "🏆", "🥇", "🎯", "🎪"],
};

export default function EmojiPicker({ onSelect }: EmojiPickerProps) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="emoji-picker">
      <button
        type="button"
        className="emoji-trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-label="Add emoji"
      >
        😊
      </button>
      {isOpen && (
        <div className="emoji-dropdown">
          {Object.entries(EMOJI_CATEGORIES).map(([category, emojis]) => (
            <div key={category} className="emoji-category">
              {emojis.map((emoji) => (
                <button
                  key={emoji}
                  type="button"
                  className="emoji-btn"
                  onClick={() => {
                    onSelect(emoji);
                    setIsOpen(false);
                  }}
                >
                  {emoji}
                </button>
              ))}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
