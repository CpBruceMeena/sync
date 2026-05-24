import { describe, it, expect, vi, beforeEach } from "vitest";
import type { AuthResponse } from "@/types";

// Mock global fetch
const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();
Object.defineProperty(globalThis, "localStorage", { value: localStorageMock });

describe("API client", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    localStorageMock.clear();
  });

  it("should make login request", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        user: { id: "1", username: "test", email: "test@test.com", display_name: "Test", avatar_url: "", status: "online" },
        token: { access_token: "access-123", refresh_token: "refresh-456", expires_in: 3600 },
      }),
    });

    const { api } = await import("@/lib/api");
    const res = await api.login({ email: "test@test.com", password: "password" });

    expect(res.user.username).toBe("test");
    expect(res.token.access_token).toBe("access-123");

    // Check fetch was called correctly
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/auth/login"),
      expect.objectContaining({
        method: "POST",
        body: expect.stringContaining("test@test.com"),
      })
    );
  });

  it("should handle API errors", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: "Unauthorized",
      json: async () => ({ error: "Invalid credentials" }),
    });

    const { api } = await import("@/lib/api");
    await expect(api.login({ email: "bad@test.com", password: "wrong" })).rejects.toThrow();
  });

  it("should include auth token in requests", async () => {
    localStorageMock.setItem("access_token", "my-token");

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    const { api } = await import("@/lib/api");
    await api.getUsers();

    // Check Authorization header
    const callArgs = mockFetch.mock.calls[0];
    const headers = callArgs[1].headers;
    expect(headers["Authorization"]).toBe("Bearer my-token");
  });

  it("should make register request", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        user: { id: "1", username: "newuser", email: "new@test.com", display_name: "newuser", avatar_url: "", status: "online" },
        token: { access_token: "access-123", refresh_token: "refresh-456", expires_in: 3600 },
      }),
    });

    const { api } = await import("@/lib/api");
    const res = await api.register({ username: "newuser", email: "new@test.com", password: "password" });

    expect(res.user.username).toBe("newuser");
  });
});
