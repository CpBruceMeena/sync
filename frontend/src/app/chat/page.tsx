"use client";

import { useState, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";
import { useWebSocket } from "@/contexts/WebSocketContext";
import { useSelectedConv } from "@/contexts/SelectedConvContext";
import { MessageInput } from "@/components/MessageInput";
import { ReactionPicker } from "@/components/ReactionPicker";
import type { Message, MessageReaction, WSMessage } from "@/types";

export default function ChatPage() {
  const { user } = useAuth();
  const { subscribe, onlineUsers } = useWebSocket();
  const { selectedConv } = useSelectedConv();
  const [messages, setMessages] = useState<Message[]>([]);
  const [loadingMessages, setLoadingMessages] = useState(false);

  // Load messages when conversation is selected
  useEffect(() => {
    if (!selectedConv) return;

    setLoadingMessages(true);
    api
      .getMessages(selectedConv.id)
      .then((msgs) => setMessages(msgs.reverse()))
      .catch(console.error)
      .finally(() => setLoadingMessages(false));
  }, [selectedConv]);

  // Subscribe to new messages via WebSocket
  useEffect(() => {
    if (!selectedConv) return;

    const unsub = subscribe("new_message", (data: WSMessage) => {
      if (data.conversation_id === selectedConv.id && data.content && data.message_id) {
        setMessages((prev) => {
          const msgId = data.message_id || "";
          // Deduplicate: skip if message already exists (avoids double-add from broadcast-to-all)
          if (prev.some((m) => m.id === msgId)) return prev;
          const newMsg: Message = {
            id: msgId,
            conversation_id: data.conversation_id || "",
            sender_id: data.sender_id || "",
            sender_username: data.sender_username || "",
            content: data.content || "",
            type: "text",
            created_at: new Date().toISOString(),
          };
          return [...prev, newMsg];
        });
      }
    });

    return unsub;
  }, [selectedConv, subscribe]);

  // Subscribe to reaction events via WebSocket
  useEffect(() => {
    if (!selectedConv) return;

    const handleReactionEvent = (data: WSMessage) => {
      if (data.conversation_id !== selectedConv.id || !data.message_id) return;

      setMessages((prev) =>
        prev.map((m) =>
          m.id === data.message_id && data.data
            ? { ...m, reactions: data.data as any }
            : m
        )
      );
    };

    const unsubAdded = subscribe("reaction_added", handleReactionEvent);
    const unsubRemoved = subscribe("reaction_removed", handleReactionEvent);

    return () => {
      unsubAdded();
      unsubRemoved();
    };
  }, [selectedConv, subscribe]);

  const handleReact = useCallback(
    async (messageId: string, emoji: string) => {
      try {
        const result = await api.toggleReaction(messageId, emoji);
        setMessages((prev) =>
          prev.map((m) =>
            m.id === messageId
              ? { ...m, reactions: result.reactions }
              : m
          )
        );
      } catch (err) {
        console.error("Failed to toggle reaction:", err);
      }
    },
    []
  );

  const handleSendMessage = useCallback(
    async (content: string) => {
      if (!selectedConv || !content.trim()) return;

      const tempId = `temp-${Date.now()}`;
      const optimisticMsg: Message = {
        id: tempId,
        conversation_id: selectedConv.id,
        sender_id: user?.id || "",
        sender_username: user?.username || "",
        content,
        type: "text",
        created_at: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, optimisticMsg]);

      try {
        const msg = await api.sendMessage(selectedConv.id, content);
        setMessages((prev) =>
          prev.map((m) => (m.id === tempId ? msg : m))
        );
      } catch (err) {
        console.error("Failed to send message:", err);
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
      }
    },
    [selectedConv, user]
  );

  const handleSendFile = useCallback(
    async (file: File) => {
      if (!selectedConv) return;

      const tempId = `temp-${Date.now()}`;
      const optimisticMsg: Message = {
        id: tempId,
        conversation_id: selectedConv.id,
        sender_id: user?.id || "",
        sender_username: user?.username || "",
        content: file.name,
        type: "file",
        created_at: new Date().toISOString(),
        attachments: [{
          id: "",
          file_url: "",
          file_type: file.type,
          file_name: file.name,
          file_size: file.size,
        }],
      };

      setMessages((prev) => [...prev, optimisticMsg]);

      try {
        const result = await api.uploadFile(selectedConv.id, file);

        // Send a file-type message with the attachment metadata
        const msg = await api.sendMessage(
          selectedConv.id,
          `Sent a file: ${file.name}`,
          "file",
          {
            id: result.id,
            file_url: result.file_url,
            file_type: result.file_type,
            file_name: result.file_name,
            file_size: result.file_size,
          }
        );

        const updatedMsg: Message = {
          ...msg,
          attachments: [{
            id: result.id,
            file_url: result.file_url,
            file_type: result.file_type,
            file_name: result.file_name,
            file_size: result.file_size,
          }],
        };

        setMessages((prev) =>
          prev.map((m) => (m.id === tempId ? updatedMsg : m))
        );
      } catch (err) {
        console.error("Failed to send file:", err);
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
      }
    },
    [selectedConv, user]
  );

  if (!selectedConv) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center space-y-4"
        >
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-[var(--surface-3)]">
            <svg
              className="w-10 h-10 text-[var(--text-muted)]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
          </div>
          <h2 className="text-xl font-semibold text-[var(--foreground)]">
            Select a Conversation
          </h2>
          <p className="text-sm text-[var(--text-muted)] max-w-sm">
            Choose a conversation from the sidebar or start a new one to begin
            chatting
          </p>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-w-0">
      {/* Chat header */}
      <div className="glass px-6 py-3 flex items-center gap-3 border-b border-[var(--border)]">
        <div className="relative">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-semibold text-sm">
            {selectedConv.name
              ? selectedConv.name.charAt(0).toUpperCase()
              : selectedConv.members?.find((m) => m.user_id !== user?.id)
                  ?.username?.charAt(0)
                  .toUpperCase() || "?"}
          </div>
          {onlineUsers.some(
            (u) =>
              u.user_id ===
              selectedConv.members?.find((m) => m.user_id !== user?.id)?.user_id
          ) && (
            <div className="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full bg-[var(--online)] border-2 border-[var(--surface)]" />
          )}
        </div>
        <div className="flex-1 min-w-0">
          <h2 className="text-sm font-semibold text-[var(--foreground)] truncate">
            {selectedConv.name ||
              selectedConv.members?.find((m) => m.user_id !== user?.id)
                ?.username ||
              "Unknown"}
          </h2>
          <p className="text-xs text-[var(--text-muted)]">
            {selectedConv.type === "group"
              ? `${selectedConv.members?.length || 0} members`
              : "Private conversation"}
          </p>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-1 smooth-scroll">
        {loadingMessages ? (
          <div className="flex items-center justify-center h-full">
            <div className="w-8 h-8 rounded-full border-2 border-[var(--primary)] border-t-transparent animate-spin" />
          </div>
        ) : (
          <AnimatePresence initial={false}>
            {messages.map((msg, i) => (
              <motion.div
                key={msg.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.15, delay: i * 0.01 }}
              >
                <MessageBubble
                  message={msg}
                  isOwn={msg.sender_id === user?.id}
                  isGroup={selectedConv.type === "group"}
                  userId={user?.id || ""}
                  onReact={handleReact}
                />
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>

      {/* Input */}
      <MessageInput onSend={handleSendMessage} onSendFile={handleSendFile} />
    </div>
  );
}

// Internal MessageBubble component
function MessageBubble({
  message,
  isOwn,
  isGroup,
  userId,
  onReact,
}: {
  message: Message;
  isOwn: boolean;
  isGroup: boolean;
  userId: string;
  onReact: (messageId: string, emoji: string) => void;
}) {
  // Group reactions by emoji
  const reactionSummary = message.reactions?.reduce<
    Record<string, { count: number; hasReacted: boolean }>
  >((acc, rxn) => {
    if (!acc[rxn.emoji]) {
      acc[rxn.emoji] = { count: 0, hasReacted: false };
    }
    acc[rxn.emoji].count++;
    if (rxn.user_id === userId) {
      acc[rxn.emoji].hasReacted = true;
    }
    return acc;
  }, {});

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const isImageMime = (mime: string): boolean => {
    return mime.startsWith("image/");
  };

  return (
    <div
      className={`flex ${isOwn ? "justify-start" : "justify-end"} mb-2 message-enter`}
    >
      <div className="max-w-[70%] flex flex-col gap-1">
        <div
          className={`rounded-2xl px-4 py-2 ${
            isOwn
              ? "bg-[var(--surface-2)] border border-[var(--border)] text-[var(--foreground)] rounded-tl-md"
              : "bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] text-white rounded-tr-md"
          }`}
        >
          {isGroup && isOwn && (
            <p className="text-xs font-medium text-[var(--text-muted)] mb-1">
              You
            </p>
          )}
          {isGroup && !isOwn && message.sender_username && (
            <p className="text-xs font-medium text-[var(--accent-light)] mb-1">
              {message.sender_username}
            </p>
          )}

          {/* Attachment previews */}
          {message.attachments && message.attachments.length > 0 && (
            <div className="mb-1.5 space-y-1">
              {message.attachments.map((att) => (
                <div key={att.id || att.file_name}>
                  {isImageMime(att.file_type) ? (
                    <a
                      href={api.getFileUrl(att.file_url)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="block"
                    >
                      <img
                        src={api.getFileUrl(att.file_url)}
                        alt={att.file_name}
                        className="max-w-full rounded-lg max-h-64 object-cover border border-[var(--border)]"
                      />
                    </a>
                  ) : (
                    <a
                      href={api.getFileUrl(att.file_url)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={`flex items-center gap-2 px-3 py-2 rounded-lg border transition-colors ${
                        isOwn
                          ? "bg-white/10 border-white/20 hover:bg-white/15"
                          : "bg-[var(--surface-3)] border-[var(--border)] hover:bg-[var(--surface-2)]"
                      }`}
                    >
                      <svg className="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                      </svg>
                      <div className="flex-1 min-w-0">
                        <p className="text-xs font-medium truncate">
                          {att.file_name}
                        </p>
                        <p className="text-[10px] opacity-60">
                          {formatFileSize(att.file_size)}
                        </p>
                      </div>
                      <svg className="w-4 h-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                      </svg>
                    </a>
                  )}
                </div>
              ))}
            </div>
          )}

          {message.content && (message.type !== "file" || !message.attachments?.length) && (
            <p className="text-sm whitespace-pre-wrap break-words">
              {message.content}
            </p>
          )}
          <p
            className={`text-[10px] mt-1 ${
              isOwn ? "text-[var(--text-muted)]" : "text-white/60"
            }`}
          >
            {new Date(message.created_at).toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            })}
          </p>
        </div>

        {/* Reactions */}
        {(reactionSummary && Object.keys(reactionSummary).length > 0) || (
          <div className="flex items-center gap-1 pl-1">
            <ReactionPicker messageId={message.id} onReact={onReact} />
          </div>
        )}
        {reactionSummary && Object.keys(reactionSummary).length > 0 && (
          <div className="flex items-center gap-1 pl-1">
            {Object.entries(reactionSummary).map(([emoji, { count, hasReacted }]) => (
              <button
                key={emoji}
                onClick={() => onReact(message.id, emoji)}
                className={`inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-xs transition-colors ${
                  hasReacted
                    ? "bg-[var(--primary)]/10 border border-[var(--primary)]/30"
                    : "bg-[var(--surface-3)] border border-[var(--border)] hover:bg-[var(--surface-2)]"
                }`}
              >
                <span>{emoji}</span>
                {count > 1 && (
                  <span className="text-[10px] text-[var(--text-muted)]">{count}</span>
                )}
              </button>
            ))}
            <ReactionPicker messageId={message.id} onReact={onReact} />
          </div>
        )}
      </div>
    </div>
  );
}
