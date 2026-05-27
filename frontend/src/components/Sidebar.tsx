"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";
import { useWebSocket } from "@/contexts/WebSocketContext";
import { useSelectedConv } from "@/contexts/SelectedConvContext";
import { CreateGroupDialog } from "./CreateGroupDialog";
import { NotificationBadge } from "./NotificationBadge";
import { ProfileDialog } from "./ProfileDialog";
import { UserProfileDialog } from "./UserProfileDialog";
import type { Conversation } from "@/types";
import { DiscoveryDialog } from "./DiscoveryDialog";

export function Sidebar() {
  const { user, logout } = useAuth();
  const { onlineUsers, isConnected } = useWebSocket();
  const { setSelectedConv } = useSelectedConv();
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateGroup, setShowCreateGroup] = useState(false);
  const [showDiscovery, setShowDiscovery] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const [showUserProfile, setShowUserProfile] = useState(false);
  const [viewUser, setViewUser] = useState<import("@/types").User | null>(null);
  const [activeConvId, setActiveConvId] = useState<string | null>(null);
  const [convTab, setConvTab] = useState<"all" | "private" | "group">("all");

  useEffect(() => {
    api
      .getConversations()
      .then(setConversations)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const handleCreateGroup = async (
    name: string,
    members: string[],
    isPublic: boolean
  ) => {
    try {
      const conv = await api.createConversation({
        type: "group",
        name,
        members,
        is_public: isPublic,
      });
      setConversations((prev) => [conv, ...prev]);
      setShowCreateGroup(false);
    } catch (err) {
      console.error("Failed to create group:", err);
    }
  };

  const handleSelect = (conv: Conversation) => {
    setActiveConvId(conv.id);
    setSelectedConv(conv);
  };

  const isUserOnline = (userId: string) => onlineUsers.some((u) => u.user_id === userId);

  return (
    <>
      <aside className="w-72 flex-shrink-0 glass border-r border-[var(--border)] flex flex-col h-full">
        {/* Header */}
        <div className="p-4 border-b border-[var(--border)]">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center overflow-hidden">
                <img src="/sync_logo.png" alt="sync" className="w-full h-full object-contain p-1" />
              </div>
              <span className="text-sm font-semibold text-[var(--foreground)]">
                sync
              </span>
            </div>
            <div className="flex items-center gap-1">
              <button
                onClick={() => setShowDiscovery(true)}
                className="p-1.5 rounded-lg hover:bg-[var(--surface-3)] text-[var(--text-muted)] hover:text-[var(--primary)] transition-colors"
                title="Discover users and groups"
              >
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </button>
              <NotificationBadge />
              <span
                className={`w-2 h-2 rounded-full ${
                  isConnected
                    ? "bg-[var(--online)]"
                    : "bg-[var(--offline)]"
                }`}
              />
              <span className="text-[10px] text-[var(--text-muted)]">
                {isConnected ? "Connected" : "Offline"}
              </span>
            </div>
          </div>

          {/* Search */}
          <input
            type="text"
            placeholder="Search conversations..."
            className="w-full px-3 py-2 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] transition-colors"
          />
        </div>

        {/* Conversation tabs */}
        <div className="flex border-b border-[var(--border)] px-4 pt-2">
          {(["all", "private", "group"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setConvTab(tab)}
              className={`py-2 px-3 text-xs font-medium border-b-2 transition-colors capitalize ${
                convTab === tab
                  ? "border-[var(--primary)] text-[var(--foreground)]"
                  : "border-transparent text-[var(--text-muted)] hover:text-[var(--foreground)]"
              }`}
            >
              {tab === "all" ? "All" : tab === "private" ? "1-1 Chats" : "Groups"}
            </button>
          ))}
        </div>

        {/* Conversation list */}
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="w-6 h-6 rounded-full border-2 border-[var(--primary)] border-t-transparent animate-spin" />
            </div>
          ) : conversations.length === 0 ? (
            <div className="text-center py-8 px-4">
              <p className="text-sm text-[var(--text-muted)]">
                No conversations yet
              </p>
              <p className="text-xs text-[var(--text-muted)] mt-1">
                Create a group or send a message to get started
              </p>
            </div>
          ) : (
            <AnimatePresence>
              {conversations
                .filter((conv) => convTab === "all" || conv.type === convTab)
                .map((conv) => {
                const otherMember = conv.members?.find(
                  (m) => m.user_id !== user?.id
                );
                const displayName = conv.name || otherMember?.username || "Unknown";
                const isOnline = otherMember
                  ? isUserOnline(otherMember.user_id)
                  : false;

                return (
                  <motion.button
                    key={conv.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    onClick={() => handleSelect(conv)}
                    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-left transition-colors ${
                      activeConvId === conv.id
                        ? "bg-[var(--surface-3)] border border-[var(--border-light)]"
                        : "hover:bg-[var(--surface-2)] border border-transparent"
                    }`}
                  >
                    <div className="relative flex-shrink-0">
                      <div
                        onClick={(e) => {
                          e.stopPropagation();
                          if (otherMember) {
                            api.getUser(otherMember.user_id).then((u) => {
                              setViewUser(u);
                              setShowUserProfile(true);
                            }).catch(() => {});
                          }
                        }}
                        className="w-10 h-10 rounded-full bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] flex items-center justify-center text-white font-semibold text-sm cursor-pointer hover:opacity-80 transition-opacity"
                      >
                        {displayName.charAt(0).toUpperCase()}
                      </div>
                      {conv.type === "private" && (
                        <div
                          className={`absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-2 border-[var(--surface)] ${
                            isOnline
                              ? "bg-[var(--online)]"
                              : "bg-[var(--offline)]"
                          }`}
                        />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-[var(--foreground)] truncate">
                          {displayName}
                        </span>
                        {conv.last_message_at && (
                          <span className="text-[10px] text-[var(--text-muted)] flex-shrink-0">
                            {new Date(conv.last_message_at).toLocaleTimeString(
                              [],
                              { hour: "2-digit", minute: "2-digit" }
                            )}
                          </span>
                        )}
                      </div>
                      {conv.last_message_content && (
                        <span className="text-xs text-[var(--text-muted)] truncate block">
                          {conv.last_message_content}
                        </span>
                      )}
                    </div>
                  </motion.button>
                );
              })}
            </AnimatePresence>
          )}
        </div>

        {/* User footer */}
        <div className="p-3 border-t border-[var(--border)]">
          <div className="flex items-center gap-3 px-2">              <button onClick={() => setShowProfile(true)} className="flex items-center gap-3 flex-1 min-w-0 text-left">
              <div className="relative flex-shrink-0">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--accent)] to-[var(--primary)] flex items-center justify-center text-white font-semibold text-xs overflow-hidden">
                  {user?.avatar_url ? (
                    <img src={api.getFileUrl(user.avatar_url)} alt="Avatar" className="w-full h-full object-cover" />
                  ) : (
                    user?.username?.charAt(0).toUpperCase() || "U"
                  )}
                </div>
                <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-[var(--online)] border-2 border-[var(--surface)]" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-[var(--foreground)] truncate">
                  {user?.display_name || user?.username}
                </p>
                <p className="text-[10px] text-[var(--text-muted)]">{user?.status || "Online"}</p>
              </div>
            </button>
            <div className="flex gap-1">
              <button
                onClick={() => setShowCreateGroup(true)}
                className="p-1.5 rounded-lg hover:bg-[var(--surface-3)] text-[var(--text-muted)] hover:text-[var(--primary)] transition-colors"
                title="Create Group"
              >
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </button>
              <button
                onClick={logout}
                className="p-1.5 rounded-lg hover:bg-[var(--surface-3)] text-[var(--text-muted)] hover:text-[var(--error)] transition-colors"
                title="Logout"
              >
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </aside>

      {showCreateGroup && (
        <CreateGroupDialog
          onClose={() => setShowCreateGroup(false)}
          onCreate={handleCreateGroup}
        />
      )}

      {showDiscovery && (
        <DiscoveryDialog
          onClose={() => setShowDiscovery(false)}
        />
      )}

      {showProfile && (
        <ProfileDialog
          onClose={() => setShowProfile(false)}
        />
      )}

      {showUserProfile && viewUser && (
        <UserProfileDialog
          user={viewUser}
          onClose={() => { setShowUserProfile(false); setViewUser(null); }}
        />
      )}
    </>
  );
}
