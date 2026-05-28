"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { motion } from "framer-motion";

interface MessageInputProps {
  onSend: (content: string) => void;
  onSendFile?: (file: File) => void;
}

export function MessageInput({ onSend, onSendFile }: MessageInputProps) {
  const [content, setContent] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [sending, setSending] = useState(false);
  const sendingRef = useRef(false);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const handleSend = useCallback(async () => {
    // Synchronous guard: prevents double-submit before React re-renders
    if (sendingRef.current) return;

    if (selectedFile && onSendFile) {
      sendingRef.current = true;
      setSending(true);
      try {
        await onSendFile(selectedFile);
      } finally {
        setSelectedFile(null);
        sendingRef.current = false;
        setSending(false);
      }
      return;
    }

    if (!content.trim()) return;

    sendingRef.current = true;
    setSending(true);
    const text = content.trim();
    setContent("");
    if (inputRef.current) {
      inputRef.current.style.height = "auto";
    }

    try {
      await onSend(text);
    } finally {
      sendingRef.current = false;
      setSending(false);
    }
  }, [content, selectedFile, onSend, onSendFile]);

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

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
    // Reset input so the same file can be re-selected
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const clearSelectedFile = () => {
    setSelectedFile(null);
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const isImageFile = (file: File): boolean => {
    return file.type.startsWith("image/");
  };

  const canSend = content.trim() || selectedFile;

  return (
    <div className="glass px-4 py-3 border-t border-[var(--border)]">
      {/* Selected file preview */}
      {selectedFile && (
        <div className="mb-2 flex items-center gap-2 px-3 py-2 rounded-lg bg-[var(--surface-2)] border border-[var(--border)]">
          {isImageFile(selectedFile) ? (
            <div className="w-10 h-10 rounded-lg overflow-hidden flex-shrink-0 bg-[var(--surface-3)]">
              <img
                src={URL.createObjectURL(selectedFile)}
                alt={selectedFile.name}
                className="w-full h-full object-cover"
              />
            </div>
          ) : (
            <div className="w-10 h-10 rounded-lg flex-shrink-0 bg-[var(--surface-3)] flex items-center justify-center">
              <svg className="w-5 h-5 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            </div>
          )}
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium text-[var(--foreground)] truncate">
              {selectedFile.name}
            </p>
            <p className="text-[10px] text-[var(--text-muted)]">
              {formatFileSize(selectedFile.size)}
            </p>
          </div>
          <button
            onClick={clearSelectedFile}
            className="flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--foreground)] hover:bg-[var(--surface-3)] transition-colors"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      <div className="flex items-end gap-3">
        <div className="flex-1 relative">
          <textarea
            ref={inputRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            onInput={handleInput}
            placeholder={selectedFile ? "Add a caption..." : "Type a message..."}
            rows={1}
            className="w-full px-4 py-2.5 rounded-xl bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] resize-none transition-colors"
            style={{ minHeight: "40px", maxHeight: "120px" }}
          />
        </div>

        {/* Attachment button */}
        <motion.button
          onClick={() => fileInputRef.current?.click()}
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          className={`flex-shrink-0 w-10 h-10 rounded-xl flex items-center justify-center transition-colors ${
            selectedFile
              ? "bg-[var(--primary)] text-white"
              : "bg-[var(--surface-2)] border border-[var(--border)] text-[var(--text-muted)] hover:text-[var(--foreground)]"
          }`}
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
          </svg>
        </motion.button>
        <input
          ref={fileInputRef}
          type="file"
          accept=".jpg,.jpeg,.png,.gif,.webp,.svg,.pdf,.doc,.docx,.txt,.csv,.json,.xml,.mp3,.mp4,.mov,.zip,.tar,.gz"
          className="hidden"
          onChange={handleFileSelect}
        />

        {/* Send button */}
        <motion.button
          onClick={handleSend}
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          disabled={!canSend || sending}
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
