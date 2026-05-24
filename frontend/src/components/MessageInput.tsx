"use client";

import { useState, useRef, useEffect } from "react";
import { motion } from "framer-motion";

interface MessageInputProps {
  onSend: (content: string) => void;
}

export function MessageInput({ onSend }: MessageInputProps) {
  const [content, setContent] = useState("");
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const handleSend = () => {
    if (!content.trim()) return;
    onSend(content.trim());
    setContent("");
    if (inputRef.current) {
      inputRef.current.style.height = "auto";
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleInput = () => {
    if (inputRef.current) {
      inputRef.current.style.height = "auto";
      inputRef.current.style.height = Math.min(
        inputRef.current.scrollHeight,
        120
      ) + "px";
    }
  };

  return (
    <div className="glass px-4 py-3 border-t border-[var(--border)]">
      <div className="flex items-end gap-3">
        <div className="flex-1 relative">
          <textarea
            ref={inputRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            onInput={handleInput}
            placeholder="Type a message..."
            rows={1}
            className="w-full px-4 py-2.5 rounded-xl bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] resize-none transition-colors"
            style={{ minHeight: "40px", maxHeight: "120px" }}
          />
        </div>
        <motion.button
          onClick={handleSend}
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          disabled={!content.trim()}
          className="flex-shrink-0 w-10 h-10 rounded-xl bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white disabled:opacity-40 transition-opacity"
        >
          <svg
            className="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
            />
          </svg>
        </motion.button>
      </div>
    </div>
  );
}
