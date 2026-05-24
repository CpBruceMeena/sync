"use client";

import { useState, useRef, useEffect } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import type { User } from "@/types";

interface CreateGroupDialogProps {
  onClose: () => void;
  onCreate: (name: string, members: string[]) => void;
}

export function CreateGroupDialog({
  onClose,
  onCreate,
}: CreateGroupDialogProps) {
  const [groupName, setGroupName] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [users, setUsers] = useState<User[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const dialogRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    api
      .getUsers()
      .then(setUsers)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dialogRef.current &&
        !dialogRef.current.contains(e.target as Node)
      ) {
        onClose();
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [onClose]);

  const toggleUser = (userId: string) => {
    setSelectedUsers((prev) => {
      const next = new Set(prev);
      if (next.has(userId)) {
        next.delete(userId);
      } else {
        next.add(userId);
      }
      return next;
    });
  };

  const handleCreate = () => {
    if (!groupName.trim() || selectedUsers.size === 0) return;
    onCreate(groupName.trim(), Array.from(selectedUsers));
  };

  const filteredUsers = users.filter(
    (u) =>
      u.username.toLowerCase().includes(searchTerm.toLowerCase()) &&
      !selectedUsers.has(u.id)
  );

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
            Create Group
          </h2>
          <p className="text-sm text-[var(--text-muted)] mt-1">
            Give your group a name and add members
          </p>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          <input
            type="text"
            value={groupName}
            onChange={(e) => setGroupName(e.target.value)}
            placeholder="Group name"
            className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors"
          />

          {/* Selected users */}
          {selectedUsers.size > 0 && (
            <div className="flex flex-wrap gap-2">
              {Array.from(selectedUsers).map((userId) => {
                const u = users.find((u) => u.id === userId);
                return (
                  <span
                    key={userId}
                    className="inline-flex items-center gap-1 px-2 py-1 rounded-full bg-[var(--surface-3)] border border-[var(--border)] text-xs text-[var(--foreground)]"
                  >
                    {u?.username || userId}
                    <button
                      onClick={() => toggleUser(userId)}
                      className="text-[var(--text-muted)] hover:text-[var(--error)]"
                    >
                      ×
                    </button>
                  </span>
                );
              })}
            </div>
          )}

          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search users..."
            className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors"
          />

          {/* User list */}
          <div className="max-h-48 overflow-y-auto space-y-1">
            {loading ? (
              <div className="flex items-center justify-center py-4">
                <div className="w-5 h-5 rounded-full border-2 border-[var(--primary)] border-t-transparent animate-spin" />
              </div>
            ) : filteredUsers.length === 0 ? (
              <p className="text-center text-sm text-[var(--text-muted)] py-4">
                {searchTerm
                  ? "No users found"
                  : "No users available"}
              </p>
            ) : (
              filteredUsers.map((u) => (
                <button
                  key={u.id}
                  onClick={() => toggleUser(u.id)}
                  className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-[var(--surface-2)] transition-colors text-left"
                >
                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-semibold text-xs">
                    {u.username.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-[var(--foreground)]">
                      {u.username}
                    </p>
                    <p className="text-xs text-[var(--text-muted)]">
                      {u.email}
                    </p>
                  </div>
                </button>
              ))
            )}
          </div>
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
            onClick={handleCreate}
            disabled={!groupName.trim() || selectedUsers.size === 0}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] text-white hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            Create Group
          </button>
        </div>
      </motion.div>
    </div>
  );
}
