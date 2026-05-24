"use client";

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import { auth as authHelper } from "@/lib/auth";
import { api } from "@/lib/api";
import type { User, TokenPair } from "@/types";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  updateUser: (user: User) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (authHelper.isAuthenticated()) {
      api
        .getMe()
        .then((user) => setUser(user))
        .catch(() => authHelper.clearTokens())
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    setError(null);
    try {
      const res = await api.login({ email, password });
      authHelper.setTokens(res.token.access_token, res.token.refresh_token);
      setUser(res.user);
    } catch (err: any) {
      setError(err.message || "Login failed");
      throw err;
    }
  }, []);

  const register = useCallback(
    async (username: string, email: string, password: string) => {
      setError(null);
      try {
        const res = await api.register({ username, email, password });
        authHelper.setTokens(res.token.access_token, res.token.refresh_token);
        setUser(res.user);
      } catch (err: any) {
        setError(err.message || "Registration failed");
        throw err;
      }
    },
    []
  );

  const logout = useCallback(() => {
    api.logout().catch(() => {});
    authHelper.clearTokens();
    setUser(null);
  }, []);

  const updateUser = useCallback((updatedUser: User) => {
    setUser(updatedUser);
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, loading, error, login, register, logout, updateUser }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
