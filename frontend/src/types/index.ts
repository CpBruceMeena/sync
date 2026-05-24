export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
  status: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface AuthResponse {
  user: User;
  token: TokenPair;
}

export interface Conversation {
  id: string;
  type: "private" | "group";
  name: string;
  admin_id: string | null;
  created_at: string;
  updated_at: string;
  members?: ConversationMember[];
  last_message_content?: string;
  last_message_at?: string;
}

export interface ConversationMember {
  user_id: string;
  username: string;
  role: "admin" | "member";
  joined_at: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  sender_username: string;
  content: string;
  type: string;
  created_at: string;
  reactions?: MessageReaction[];
}

export interface MessageReaction {
  user_id: string;
  username: string;
  emoji: string;
  created_at: string;
}

export interface WSMessage {
  type: string;
  conversation_id?: string;
  sender_id?: string;
  sender_username?: string;
  content?: string;
  message_id?: string;
  user_id?: string;
  username?: string;
  status?: string;
  is_typing?: boolean;
  error?: string;
  data?: any;
}
