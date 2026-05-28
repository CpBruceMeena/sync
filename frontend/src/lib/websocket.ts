const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";

type MessageHandler = (data: any) => void;

export class WSClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private url: string;
  // Queue messages sent while connecting — they'll be flushed once the socket opens
  private sendQueue: Array<Record<string, any>> = [];

  constructor() {
    this.url = `${WS_BASE}/ws`;
  }

  /**
   * Returns the full WS URL with the current token from localStorage.
   * This ensures the token is always fresh, even after user switch.
   */
  private getUrl(): string {
    const token =
      typeof window !== "undefined"
        ? localStorage.getItem("access_token")
        : null;
    return `${this.url}?token=${token}`;
  }

  private flushQueue() {
    if (this.sendQueue.length === 0) return;
    const queue = this.sendQueue.slice();
    this.sendQueue = [];
    for (const msg of queue) {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify(msg));
      }
    }
  }

  connect() {
    // Close any existing connection first
    if (this.ws) {
      this.ws.onopen = null;
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
    }

    const url = this.getUrl();
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      console.log("[WS] Connected");
      this.dispatch("connect", {});
      this.flushQueue();
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
    if (this.ws) {
      this.ws.onopen = null;
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
      this.ws = null;
    }
  }

  send(data: Record<string, any>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      // Queue the message — it will be flushed once the socket opens
      this.sendQueue.push(data);
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
