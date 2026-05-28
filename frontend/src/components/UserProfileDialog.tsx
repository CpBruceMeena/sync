"use client";

import { useState, useEffect, useRef } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import type { User } from "@/types";

interface UserProfileDialogProps {
  user: User;
  onClose: () => void;
}

export function UserProfileDialog({ user: profileUser, onClose }: UserProfileDialogProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const [user, setUser] = useState<User>(profileUser);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dialogRef.current && !dialogRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [onClose]);

  // Try to fetch fresh data if only basic info was provided
  useEffect(() => {
    if (profileUser.id) {
      api.getUser(profileUser.id).then(setUser).catch(() => {});
    }
  }, [profileUser.id]);

  const statusColor = (status: string) => {
    switch (status) {
      case "online": return "bg-[var(--online)]";
      case "away": return "bg-yellow-500";
      case "busy": return "bg-[var(--error)]";
      default: return "bg-[var(--offline)]";
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <motion.div
        ref={dialogRef}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.2 }}
        className="glass rounded-2xl w-full max-w-sm mx-4 overflow-hidden"
      >
        {/* Header with avatar */}
        <div className="relative pt-8 pb-6 px-6 text-center bg-gradient-to-b from-[var(--primary)]/10 to-transparent">
          <div className="relative inline-block">
            <div className="w-20 h-20 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-bold text-2xl overflow-hidden mx-auto ring-4 ring-[var(--surface)]">
              {user.avatar_url ? (
                <img src={api.getFileUrl(user.avatar_url)} alt={user.display_name || user.username} className="w-full h-full object-cover" />
              ) : (
                user.username.charAt(0).toUpperCase()
              )}
            </div>
            <div className={`absolute -bottom-0.5 -right-0.5 w-5 h-5 rounded-full border-2 border-[var(--surface)] ${statusColor(user.status)}`} />
          </div>
          <h2 className="text-lg font-semibold text-[var(--foreground)] mt-3">
            {user.display_name || user.username}
          </h2>
          <p className="text-sm text-[var(--text-muted)]">@{user.username}</p>
        </div>

        {/* Info */}
        <div className="px-6 py-4 space-y-4">
          {user.bio && (
            <div>
              <p className="text-xs font-medium text-[var(--text-muted)] uppercase tracking-wider mb-1">Bio</p>
              <p className="text-sm text-[var(--foreground)]">{user.bio}</p>
            </div>
          )}
          <div className="flex items-center justify-between text-sm">
            <span className="text-[var(--text-muted)]">Status</span>
            <span className="flex items-center gap-1.5 text-[var(--foreground)] capitalize">
              <span className={`w-2 h-2 rounded-full ${statusColor(user.status)}`} />
              {user.status}
            </span>
          </div>
          {user.email && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-[var(--text-muted)]">Email</span>
              <span className="text-[var(--foreground)]">{user.email}</span>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 pb-4 flex justify-end">
          <button
            onClick={onClose}
            className="px-5 py-2 rounded-lg text-sm font-medium bg-[var(--surface-2)] text-[var(--foreground)] hover:bg-[var(--surface-3)] transition-colors"
          >
            Close
          </button>
        </div>
      </motion.div>
    </div>
  );
}
