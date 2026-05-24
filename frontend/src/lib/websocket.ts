const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";

type MessageHandler = (data: any) => void;

export class WSClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private url: string;

  constructor() {
    const token =
      typeof window !== "undefined"
        ? localStorage.getItem("access_token")
        : null;
    this.url = `${WS_BASE}/ws?token=${token}`;
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log("[WS] Connected");
      this.dispatch("connect", {});
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    };

    this.ws.onclose = () => {
      console.log("[WS] Disconnected, reconnecting in 3s...");
      this.dispatch("disconnect", {});
      this.reconnectTimer = setTimeout(() => this.connect(), 3000);
    };

    this.ws.onerror = (err) => {
      console.error("[WS] Error:", err);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.dispatch(data.type, data);
      } catch (e) {
        console.error("[WS] Failed to parse message:", e);
      }
    };
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
  }

  send(data: Record<string, any>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.warn("[WS] Cannot send, not connected");
    }
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => this.off(type, handler);
  }

  off(type: string, handler: MessageHandler) {
    this.handlers.get(type)?.delete(handler);
  }

  private dispatch(type: string, data: any) {
    this.handlers.get(type)?.forEach((handler) => handler(data));
  }
}

let wsInstance: WSClient | null = null;

export function getWSClient(): WSClient {
  if (!wsInstance) {
    wsInstance = new WSClient();
  }
  return wsInstance;
}
