const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("access_token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(body.error || res.statusText, res.status);
  }

  return res.json();
}

export const api = {
  // Auth
  register: (data: { username: string; email: string; password: string }) =>
    request<import("../types").AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  login: (data: { email: string; password: string }) =>
    request<import("../types").AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  refresh: (refreshToken: string) =>
    request<{ token: import("../types").TokenPair }>("/api/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),

  logout: () =>
    request<void>("/api/auth/logout", { method: "POST" }),

  getMe: () =>
    request<import("../types").User>("/api/auth/me"),

  // Users
  getUsers: () =>
    request<import("../types").User[]>("/api/users"),

  getUser: (id: string) =>
    request<import("../types").User>(`/api/users/${id}`),

  updateProfile: (data: { display_name?: string; avatar_url?: string; status?: string }) =>
    request<import("../types").User>("/api/users/me", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  // Conversations
  getConversations: () =>
    request<import("../types").Conversation[]>("/api/conversations"),

  createConversation: (data: { type: string; name?: string; members: string[] }) =>
    request<import("../types").Conversation>("/api/conversations", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getConversation: (id: string) =>
    request<import("../types").Conversation>(`/api/conversations/${id}`),

  addMember: (conversationId: string, username: string) =>
    request<void>(`/api/conversations/${conversationId}/members`, {
      method: "POST",
      body: JSON.stringify({ username }),
    }),

  removeMember: (conversationId: string, userId: string) =>
    request<void>(`/api/conversations/${conversationId}/members/${userId}`, {
      method: "DELETE",
    }),

  // Messages
  getMessages: (conversationId: string, cursor?: string, limit = 50) =>
    request<import("../types").Message[]>(
      `/api/conversations/${conversationId}/messages?limit=${limit}${cursor ? `&cursor=${cursor}` : ""
      }`
    ),

  sendMessage: (conversationId: string, content: string, type = "text") =>
    request<import("../types").Message>(`/api/conversations/${conversationId}/messages`, {
      method: "POST",
      body: JSON.stringify({ content, type }),
    }),

  deleteMessage: (id: string) =>
    request<void>(`/api/messages/${id}`, { method: "DELETE" }),

  // Reactions
  toggleReaction: (messageId: string, emoji: string) =>
    request<{ reactions: import("../types").MessageReaction[] }>(
      `/api/messages/${messageId}/reactions`,
      {
        method: "POST",
        body: JSON.stringify({ emoji }),
      }
    ),

  // Notifications
  getNotifications: (limit = 50) =>
    request<import("../types").Notification[]>(`/api/notifications?limit=${limit}`),

  getUnreadCount: () =>
    request<{ count: number }>("/api/notifications/unread-count"),

  markNotificationRead: (id: string) =>
    request<void>(`/api/notifications/${id}/read`, { method: "PUT" }),

  markAllNotificationsRead: () =>
    request<void>("/api/notifications/read-all", { method: "PUT" }),

  // Files
  uploadFile: async (conversationId: string, file: File) => {
    const token = localStorage.getItem("access_token");
    const formData = new FormData();
    formData.append("file", file);

    const res = await fetch(`${API_BASE}/api/files/upload`, {
      method: "POST",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: formData,
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new ApiError(body.error || res.statusText, res.status);
    }

    return res.json() as Promise<import("../types").Attachment>;
  },

  getFileUrl: (filename: string) => `${API_BASE}/api/files/${filename}`,
};
