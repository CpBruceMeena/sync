"use client";

import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";

const EMOJIS = ["👍", "❤️", "😂", "🎉", "😢", "🙌"];

interface ReactionPickerProps {
  messageId: string;
  onReact: (messageId: string, emoji: string) => void;
}

export function ReactionPicker({ messageId, onReact }: ReactionPickerProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div ref={ref} className="relative inline-flex">
      <button
        onClick={(e) => {
          e.stopPropagation();
          setOpen(!open);
        }}
        className="w-5 h-5 rounded-full bg-[var(--surface-3)] hover:bg-[var(--border-light)] flex items-center justify-center text-[10px] text-[var(--text-muted)] hover:text-[var(--foreground)] transition-colors"
        title="Add reaction"
      >
        <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 8 }}
            transition={{ duration: 0.15 }}
            className="absolute bottom-full left-0 mb-1 flex gap-0.5 bg-[var(--surface)] border border-[var(--border)] rounded-xl px-2 py-1.5 shadow-lg z-10"
          >
            {EMOJIS.map((emoji) => (
              <button
                key={emoji}
                onClick={(e) => {
                  e.stopPropagation();
                  onReact(messageId, emoji);
                  setOpen(false);
                }}
                className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-[var(--surface-3)] text-base transition-colors"
              >
                {emoji}
              </button>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
