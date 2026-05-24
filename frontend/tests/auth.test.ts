import { describe, it, expect, beforeEach } from "vitest";

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

describe("auth helpers", () => {
  beforeEach(() => {
    localStorageMock.clear();
  });

  it("should return null initially", async () => {
    const { auth } = await import("@/lib/auth");
    expect(auth.getAccessToken()).toBeNull();
    expect(auth.getRefreshToken()).toBeNull();
  });

  it("should store tokens", async () => {
    const { auth } = await import("@/lib/auth");
    auth.setTokens("access-123", "refresh-456");
    expect(auth.getAccessToken()).toBe("access-123");
    expect(auth.getRefreshToken()).toBe("refresh-456");
  });

  it("should clear tokens", async () => {
    const { auth } = await import("@/lib/auth");
    auth.setTokens("access-123", "refresh-456");
    auth.clearTokens();
    expect(auth.getAccessToken()).toBeNull();
    expect(auth.getRefreshToken()).toBeNull();
  });

  it("should check authentication", async () => {
    const { auth } = await import("@/lib/auth");
    expect(auth.isAuthenticated()).toBe(false);
    auth.setTokens("access-123", "refresh-456");
    expect(auth.isAuthenticated()).toBe(true);
  });
});
