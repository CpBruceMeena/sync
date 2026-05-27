"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
  type ReactNode,
} from "react";
import { getWSClient, WSClient } from "@/lib/websocket";
import { useAuth } from "./AuthContext";
import type { WSMessage, PresenceInfo } from "@/types";

interface WebSocketContextType {
  isConnected: boolean;
  sendMessage: (msg: Record<string, any>) => void;
  subscribe: (type: string, handler: (data: WSMessage) => void) => () => void;
  onlineUsers: PresenceInfo[];
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(
  undefined
);

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [onlineUsers, setOnlineUsers] = useState<PresenceInfo[]>([]);
  const clientRef = useRef<WSClient | null>(null);

  useEffect(() => {
    if (!user) return;

    const client = getWSClient();
    clientRef.current = client;

    const cleanupConnect = client.on("connect", () => setIsConnected(true));
    const cleanupDisconnect = client.on("disconnect", () =>
      setIsConnected(false)
    );
    const cleanupOnline = client.on("online_users", (data: WSMessage) => {
      if (data.data) setOnlineUsers(data.data);
    });

    client.connect();

    return () => {
      cleanupConnect();
      cleanupDisconnect();
      cleanupOnline();
      client.disconnect();
    };
  }, [user]);

  const sendMessage = useCallback((msg: Record<string, any>) => {
    clientRef.current?.send(msg);
  }, []);

  const subscribe = useCallback(
    (type: string, handler: (data: WSMessage) => void) => {
      if (!clientRef.current) return () => {};
      return clientRef.current.on(type, handler);
    },
    []
  );

  return (
    <WebSocketContext.Provider
      value={{ isConnected, sendMessage, subscribe, onlineUsers }}
    >
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  const ctx = useContext(WebSocketContext);
  if (!ctx)
    throw new Error("useWebSocket must be used within WebSocketProvider");
  return ctx;
}
