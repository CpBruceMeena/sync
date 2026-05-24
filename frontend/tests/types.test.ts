import { describe, it, expect } from "vitest";
import type { User, WSMessage, AuthResponse } from "@/types";

describe("Type definitions", () => {
  it("should define all required types", () => {
    const user: User = {
      id: "1",
      username: "test",
      email: "test@test.com",
      display_name: "Test User",
      avatar_url: "",
      status: "online",
    };

    expect(user.id).toBe("1");
    expect(user.username).toBe("test");
    expect(user.status).toBe("online");
  });

  it("should define WSMessage type", () => {
    const msg: WSMessage = {
      type: "new_message",
      content: "Hello",
      sender_id: "user-1",
    };
    expect(msg.type).toBe("new_message");
    expect(msg.content).toBe("Hello");
  });

  it("should define AuthResponse type", () => {
    const response: AuthResponse = {
      user: {
        id: "1",
        username: "test",
        email: "test@test.com",
        display_name: "Test",
        avatar_url: "",
        status: "online",
      },
      token: {
        access_token: "abc",
        refresh_token: "def",
        expires_in: 3600,
      },
    };
    expect(response.user.username).toBe("test");
    expect(response.token.access_token).toBeDefined();
  });
});
