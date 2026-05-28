"use client";

import { useState, useRef, useEffect } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";
import type { ProfileUpdate } from "@/types";

interface ProfileDialogProps {
  onClose: () => void;
}

export function ProfileDialog({ onClose }: ProfileDialogProps) {
  const { user, updateUser } = useAuth();
  const dialogRef = useRef<HTMLDivElement>(null);
  const [displayName, setDisplayName] = useState(user?.display_name || "");
  const [bio, setBio] = useState(user?.bio || "");
  const [avatarUrl, setAvatarUrl] = useState(user?.avatar_url || "");
  const [status, setStatus] = useState(user?.status || "online");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dialogRef.current && !dialogRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [onClose]);

  const handleSave = async () => {
    setSaving(true);
    setMessage(null);
    try {
      const data: ProfileUpdate = {
        display_name: displayName || undefined,
        avatar_url: avatarUrl || undefined,
        status: status || undefined,
        bio: bio || undefined,
      };
      const updatedUser = await api.updateProfile(data);
      updateUser(updatedUser);
      setMessage({ type: "success", text: "Profile updated successfully!" });
      setTimeout(() => onClose(), 1000);
    } catch (err: any) {
      setMessage({ type: "error", text: err.message || "Failed to update profile" });
    } finally {
      setSaving(false);
    }
  };

  if (!user) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <motion.div
        ref={dialogRef}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.2 }}
        className="glass rounded-2xl w-full max-w-md mx-4 overflow-hidden"
      >
        {/* Header */}
        <div className="p-6 pb-4 border-b border-[var(--border)]">
          <h2 className="text-lg font-semibold text-[var(--foreground)]">
            Edit Profile
          </h2>
          <p className="text-sm text-[var(--text-muted)] mt-1">
            Update your avatar, display name, and bio
          </p>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          {/* Avatar preview */}
          <div className="flex flex-col items-center gap-3">
            <div className="relative">
              <div className="w-20 h-20 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-bold text-2xl overflow-hidden">
                {avatarUrl ? (
                  <img src={api.getFileUrl(avatarUrl)} alt="Avatar" className="w-full h-full object-cover" />
                ) : (
                  user.username.charAt(0).toUpperCase()
                )}
              </div>
              <div className={`absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-2 border-[var(--surface)] ${
                status === "online" ? "bg-[var(--online)]" : status === "away" ? "bg-yellow-500" : "bg-[var(--offline)]"
              }`} />
            </div>
            <p className="text-xs text-[var(--text-muted)]">@{user.username}</p>
          </div>

          {/* Avatar URL */}
          <div>
            <label className="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
              Avatar URL
            </label>
            <input
              type="text"
              value={avatarUrl}
              onChange={(e) => setAvatarUrl(e.target.value)}
              placeholder="https://example.com/avatar.jpg"
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors"
            />
          </div>

          {/* Display Name */}
          <div>
            <label className="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
              Display Name
            </label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Your display name"
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors"
            />
          </div>

          {/* Bio */}
          <div>
            <label className="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
              Bio
            </label>
            <textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Tell us about yourself..."
              rows={3}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors resize-none"
            />
          </div>

          {/* Status */}
          <div>
            <label className="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
              Status
            </label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] focus:outline-none focus:border-[var(--primary)] transition-colors"
            >
              <option value="online">Online</option>
              <option value="away">Away</option>
              <option value="busy">Busy</option>
              <option value="offline">Offline</option>
            </select>
          </div>

          {/* Message feedback */}
          {message && (
            <p className={`text-sm text-center ${message.type === "success" ? "text-green-400" : "text-[var(--error)]"}`}>
              {message.text}
            </p>
          )}
        </div>

        {/* Footer */}
        <div className="p-6 pt-4 border-t border-[var(--border)] flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-sm text-[var(--text-dim)] hover:bg-[var(--surface-2)] transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] text-white hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {saving ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </motion.div>
    </div>
  );
}
