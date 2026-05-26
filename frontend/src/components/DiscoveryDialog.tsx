"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";
import { useSelectedConv } from "@/contexts/SelectedConvContext";
import type { UserResult, GroupDetail } from "@/types";

interface DiscoveryDialogProps {
  onClose: () => void;
}

type Tab = "users" | "groups";

export function DiscoveryDialog({ onClose }: DiscoveryDialogProps) {
  const { user } = useAuth();
  const { setSelectedConv } = useSelectedConv();
  const dialogRef = useRef<HTMLDivElement>(null);
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const [activeTab, setActiveTab] = useState<Tab>("users");
  const [searchTerm, setSearchTerm] = useState("");
  const [users, setUsers] = useState<UserResult[]>([]);
  const [groups, setGroups] = useState<GroupDetail[]>([]);
  const [selectedGroup, setSelectedGroup] = useState<GroupDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

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

  // Load initial data and handle search
  const performSearch = useCallback(
    async (term: string, tab: Tab) => {
      if (!term.trim()) {
        if (tab === "groups") {
          setLoading(true);
          try {
            const g = await api.listPublicGroups(20, 0);
            setGroups(g);
          } catch (err) {
            console.error("Failed to load public groups:", err);
            setGroups([]);
          } finally {
            setLoading(false);
          }
        } else {
          setUsers([]);
        }
        return;
      }

      setLoading(true);
      setError("");
      try {
        if (tab === "users") {
          const u = await api.searchUsers(term, 20);
          setUsers(u);
        } else {
          const g = await api.searchPublicGroups(term, 20);
          setGroups(g);
        }
      } catch (err) {
        console.error(`Failed to search ${tab}:`, err);
        setError(`Failed to search ${tab}`);
      } finally {
        setLoading(false);
      }
    },
    []
  );

  // Load public groups on mount for groups tab
  useEffect(() => {
    if (activeTab === "groups") {
      performSearch(searchTerm, "groups");
    }
  }, [activeTab, performSearch, searchTerm]);

  const handleSearchChange = (value: string) => {
    setSearchTerm(value);
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
      searchTimeoutRef.current = undefined;
    }
    searchTimeoutRef.current = setTimeout(() => {
      performSearch(value, activeTab);
    }, 300);
  };

  const handleTabChange = (tab: Tab) => {
    setActiveTab(tab);
    setSelectedGroup(null);
    setError("");
    if (!searchTerm.trim() && tab === "groups") {
      performSearch("", "groups");
    }
    if (!searchTerm.trim() && tab === "users") {
      setUsers([]);
    }
  };

  const handleUserClick = async (targetUser: UserResult) => {
    if (!user) return;
    try {
      const conv = await api.createConversation({
        type: "private",
        members: [targetUser.username],
      });
      setSelectedConv(conv);
      onClose();
    } catch (err) {
      console.error("Failed to create conversation:", err);
    }
  };

  const handleGroupClick = async (group: GroupDetail) => {
    try {
      const detail = await api.getGroupDetails(group.id);
      setSelectedGroup(detail);
    } catch (err) {
      console.error("Failed to load group details:", err);
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
        className="glass rounded-2xl w-full max-w-lg mx-4 overflow-hidden"
      >
        {/* Header */}
        <div className="p-6 pb-4 border-b border-[var(--border)]">
          <h2 className="text-lg font-semibold text-[var(--foreground)]">
            Discover
          </h2>
          <p className="text-sm text-[var(--text-muted)] mt-1">
            Find users and public groups
          </p>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-[var(--border)] px-6">
          <button
            onClick={() => handleTabChange("users")}
            className={`py-3 px-4 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "users"
                ? "border-[var(--primary)] text-[var(--foreground)]"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--foreground)]"
            }`}
          >
            Users
          </button>
          <button
            onClick={() => handleTabChange("groups")}
            className={`py-3 px-4 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "groups"
                ? "border-[var(--primary)] text-[var(--foreground)]"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--foreground)]"
            }`}
          >
            Public Groups
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          {/* Search */}
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder={
              activeTab === "users"
                ? "Search by username, display name, or email..."
                : "Search public groups by name..."
            }
            className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors"
          />

          {/* Group Detail View */}
          {selectedGroup ? (
            <div className="space-y-4">
              <button
                onClick={() => setSelectedGroup(null)}
                className="text-sm text-[var(--primary)] hover:underline"
              >
                &larr; Back to list
              </button>
              <div className="p-4 rounded-xl bg-[var(--surface-2)] border border-[var(--border)]">
                <h3 className="text-lg font-semibold text-[var(--foreground)]">
                  {selectedGroup.name}
                </h3>
                <p className="text-sm text-[var(--text-muted)] mt-1">
                  {selectedGroup.member_count} members
                </p>
              </div>
              {selectedGroup.members && selectedGroup.members.length > 0 && (
                <div>
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2">
                    Members
                  </h4>
                  <div className="space-y-1 max-h-48 overflow-y-auto">
                    {selectedGroup.members.map((m) => (
                      <div
                        key={m.user_id}
                        className="flex items-center gap-3 px-3 py-2 rounded-lg"
                      >
                        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-semibold text-xs">
                          {m.username.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <p className="text-sm font-medium text-[var(--foreground)]">
                            {m.username}
                          </p>
                          <p className="text-xs text-[var(--text-muted)] capitalize">
                            {m.role}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            /* List view */
            <div className="max-h-64 overflow-y-auto space-y-1">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="w-5 h-5 rounded-full border-2 border-[var(--primary)] border-t-transparent animate-spin" />
                </div>
              ) : error ? (
                <p className="text-center text-sm text-[var(--error)] py-4">
                  {error}
                </p>
              ) : activeTab === "users" ? (
                users.length === 0 ? (
                  <p className="text-center text-sm text-[var(--text-muted)] py-4">
                    {searchTerm
                      ? "No users found"
                      : "Start typing to search for users"}
                  </p>
                ) : (
                  users.map((u) => (
                    <button
                      key={u.id}
                      onClick={() => handleUserClick(u)}
                      className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-[var(--surface-2)] transition-colors text-left"
                    >
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-semibold text-xs">
                        {u.username.charAt(0).toUpperCase()}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-[var(--foreground)] truncate">
                          {u.display_name || u.username}
                        </p>
                        <p className="text-xs text-[var(--text-muted)] truncate">
                          @{u.username}
                          {u.status !== "offline" && (
                            <span className="ml-2 text-[var(--online)]">
                              {u.status}
                            </span>
                          )}
                        </p>
                      </div>
                    </button>
                  ))
                )
              ) : (
                groups.length === 0 ? (
                  <p className="text-center text-sm text-[var(--text-muted)] py-4">
                    {searchTerm
                      ? "No public groups found"
                      : "No public groups available"}
                  </p>
                ) : (
                  groups.map((g) => (
                    <button
                      key={g.id}
                      onClick={() => handleGroupClick(g)}
                      className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-[var(--surface-2)] transition-colors text-left"
                    >
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--accent)] to-[var(--primary)] flex items-center justify-center text-white font-semibold text-xs">
                        {g.name.charAt(0).toUpperCase()}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-[var(--foreground)]">
                          {g.name}
                        </p>
                        <p className="text-xs text-[var(--text-muted)]">
                          {g.member_count} members
                        </p>
                      </div>
                      <svg
                        className="w-4 h-4 text-[var(--text-muted)] flex-shrink-0"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                    </button>
                  ))
                )
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-6 pt-4 border-t border-[var(--border)] flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-sm text-[var(--text-dim)] hover:bg-[var(--surface-2)] transition-colors"
          >
            Close
          </button>
        </div>
      </motion.div>
    </div>
  );
}
