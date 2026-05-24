"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { useAuth } from "@/contexts/AuthContext";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(email, password);
      router.push("/chat");
    } catch (err: any) {
      setError(err.message || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-[var(--background)]">
      {/* Background gradient */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -left-40 w-80 h-80 bg-[var(--primary)] rounded-full opacity-[0.08] blur-[100px]" />
        <div className="absolute -bottom-40 -right-40 w-80 h-80 bg-[var(--accent)] rounded-full opacity-[0.06] blur-[100px]" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="relative w-full max-w-md"
      >
        <div className="glass rounded-2xl p-8 space-y-6">
          {/* Logo */}
          <div className="text-center space-y-2">
            <motion.div
              initial={{ scale: 0.8 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
              className="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] mb-3"
            >
              <img src="/logo.png" alt="sync" className="w-full h-full object-contain p-1" />
            </motion.div>
            <h1 className="text-2xl font-bold text-[var(--foreground)]">
              Welcome Back
            </h1>
            <p className="text-sm text-[var(--text-muted)]">
              Sign in to your account to continue
            </p>
          </div>

          {/* Error */}
          {error && (
            <motion.div
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              className="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-sm text-[var(--error)]"
            >
              {error}
            </motion.div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-[var(--text-dim)]">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                required
                className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] focus:ring-1 focus:ring-[var(--primary)] transition-colors"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-[var(--text-dim)]">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                className="w-full px-4 py-2.5 rounded-lg bg-[var(--surface-2)] border border-[var(--border)] text-[var(--foreground)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--primary)] focus:ring-1 focus:ring-[var(--primary)] transition-colors"
              />
            </div>
            <motion.button
              type="submit"
              disabled={loading}
              whileHover={{ scale: 1.01 }}
              whileTap={{ scale: 0.99 }}
              className="w-full py-2.5 rounded-lg bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] text-white font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
            >
              {loading ? "Signing in..." : "Sign In"}
            </motion.button>
          </form>

          {/* Footer */}
          <p className="text-center text-sm text-[var(--text-muted)]">
            Don&apos;t have an account?{" "}
            <Link
              href="/register"
              className="text-[var(--accent)] hover:underline"
            >
              Create one
            </Link>
          </p>
        </div>
      </motion.div>
    </div>
  );
}
