"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { useAuth } from "@/contexts/AuthContext";

// ── Custom SVG Icons ────────────────────────────────────────────────────

function EmailIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" strokeWidth="1.5" />
      <path d="M2 7l8.84 5.26a3 3 0 002.32 0L22 7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M7 16h2" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M15 16h2" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function LockIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="5" y="11" width="14" height="10" rx="2" stroke="currentColor" strokeWidth="1.5" />
      <path d="M8 11V7a4 4 0 018 0v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <circle cx="12" cy="16" r="1" fill="currentColor" />
      <path d="M12 16v2" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function EyeIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" stroke="currentColor" strokeWidth="1.5" />
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  );
}

function EyeOffIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M14.12 14.12a3 3 0 11-4.24-4.24" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M1 1l22 22" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function ArrowRightIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M5 12h14" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M13 5l7 7-7 7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SpinnerIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" opacity="0.25" />
      <path d="M12 2a10 10 0 019.95 9" stroke="currentColor" strokeWidth="4" strokeLinecap="round" />
    </svg>
  );
}

function AlertIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="1.5" />
      <path d="M12 8v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <circle cx="12" cy="16" r="0.75" fill="currentColor" />
    </svg>
  );
}

// ── Input Field ────────────────────────────────────────────────────────

function InputField({
  label,
  type,
  value,
  onChange,
  icon: Icon,
  placeholder,
  autoComplete,
  autoFocus,
  isFocused,
  onFocus,
  onBlur,
  rightElement,
}: {
  label: string;
  type: string;
  value: string;
  onChange: (v: string) => void;
  icon: React.ComponentType<{ className?: string }>;
  placeholder: string;
  autoComplete: string;
  autoFocus?: boolean;
  isFocused: boolean;
  onFocus: () => void;
  onBlur: () => void;
  rightElement?: React.ReactNode;
}) {
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="space-y-1.5">
      {/* Label always above */}
      <label className={`block text-xs font-medium tracking-wide transition-colors duration-300 ${
        isFocused ? "text-[var(--primary)]" : "text-[var(--text-dim)]"
      }`}>
        {label}
      </label>

      <div
        onClick={() => inputRef.current?.focus()}
        className={`group relative rounded-xl border transition-all duration-300 cursor-text ${
          isFocused
            ? "border-[var(--primary)] shadow-[0_0_0_3px_var(--primary-glow)]"
            : "border-[var(--border)] hover:border-[var(--border-light)]"
        }`}
      >
        {/* Animated gradient background on focus */}
        <motion.div
          className="absolute inset-0 rounded-xl pointer-events-none"
          style={{
            background: "linear-gradient(135deg, rgba(99,102,241,0.08), rgba(167,139,250,0.04))",
          }}
          initial={false}
          animate={{ opacity: isFocused ? 1 : 0 }}
          transition={{ duration: 0.3 }}
        />

        {/* Icon */}
        <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none z-10">
          <Icon
            className={`w-4 h-4 transition-all duration-300 ${
              isFocused ? "text-[var(--primary)]" : "text-[var(--text-muted)]"
            }`}
          />
        </div>

        {/* Input */}
        <input
          ref={inputRef}
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onFocus={onFocus}
          onBlur={onBlur}
          placeholder={placeholder}
          required
          autoComplete={autoComplete}
          autoFocus={autoFocus}
          className="w-full bg-transparent text-sm text-[var(--foreground)] placeholder:text-[var(--text-muted)] outline-none py-3 pl-10 pr-10 rounded-xl relative z-10"
        />

        {/* Right element */}
        {rightElement && (
          <div className="absolute inset-y-0 right-0 pr-2.5 flex items-center z-20">
            {rightElement}
          </div>
        )}

        {/* Bottom accent line */}
        <motion.div
          className="absolute bottom-0 left-3 right-3 h-[2px] bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] rounded-full"
          initial={false}
          animate={{
            scaleX: isFocused ? 1 : 0,
            opacity: isFocused ? 0.6 : 0,
          }}
          transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
          style={{ transformOrigin: "center" }}
        />
      </div>
    </div>
  );
}

// Particle background component
function Particles() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animId: number;
    let particles: Array<{
      x: number; y: number; vx: number; vy: number; size: number; alpha: number;
    }> = [];

    const resize = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    resize();
    window.addEventListener("resize", resize);

    // Create particles
    const count = Math.min(50, Math.floor((canvas.width * canvas.height) / 40000));
    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        vx: (Math.random() - 0.5) * 0.3,
        vy: (Math.random() - 0.5) * 0.3,
        size: Math.random() * 2 + 0.5,
        alpha: Math.random() * 0.4 + 0.1,
      });
    }

    const draw = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      // Draw connections between nearby particles
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x;
          const dy = particles[i].y - particles[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 150) {
            ctx.beginPath();
            ctx.moveTo(particles[i].x, particles[i].y);
            ctx.lineTo(particles[j].x, particles[j].y);
            ctx.strokeStyle = `rgba(99, 102, 241, ${0.06 * (1 - dist / 150)})`;
            ctx.lineWidth = 0.5;
            ctx.stroke();
          }
        }
      }

      // Draw and update particles
      for (const p of particles) {
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < 0) p.x = canvas.width;
        if (p.x > canvas.width) p.x = 0;
        if (p.y < 0) p.y = canvas.height;
        if (p.y > canvas.height) p.y = 0;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(99, 102, 241, ${p.alpha})`;
        ctx.fill();
      }

      animId = requestAnimationFrame(draw);
    };
    draw();

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener("resize", resize);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 pointer-events-none z-0"
    />
  );
}

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [focusedField, setFocusedField] = useState<string | null>(null);

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

  const easeOut: [number, number, number, number] = [0.22, 1, 0.36, 1];

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: { staggerChildren: 0.08, delayChildren: 0.1 },
    },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.5, ease: easeOut },
    },
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-[var(--background)] relative overflow-hidden">
      {/* Particle background */}
      <Particles />

      {/* Animated gradient orbs */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none z-0">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 120, repeat: Infinity, ease: "linear" }}
          className="absolute -top-1/2 -left-1/2 w-full h-full"
        >
          <div className="absolute top-[25%] left-[35%] w-[700px] h-[700px] rounded-full bg-gradient-to-r from-[var(--primary)]/8 via-purple-500/4 to-transparent blur-[150px]" />
        </motion.div>
        <motion.div
          animate={{ rotate: -360 }}
          transition={{ duration: 90, repeat: Infinity, ease: "linear" }}
          className="absolute -bottom-1/2 -right-1/2 w-full h-full"
        >
          <div className="absolute bottom-[25%] right-[35%] w-[600px] h-[600px] rounded-full bg-gradient-to-l from-[var(--accent)]/8 via-violet-500/4 to-transparent blur-[150px]" />
        </motion.div>
        {/* Third subtle orb */}
        <motion.div
          animate={{ scale: [1, 1.1, 1], opacity: [0.3, 0.5, 0.3] }}
          transition={{ duration: 8, repeat: Infinity, ease: "easeInOut" }}
          className="absolute top-[50%] left-[50%] -translate-x-1/2 -translate-y-1/2"
        >
          <div className="w-[400px] h-[400px] rounded-full bg-gradient-to-b from-indigo-500/3 to-transparent blur-[100px]" />
        </motion.div>
      </div>

      {/* Grid pattern overlay */}
      <div
        className="fixed inset-0 pointer-events-none z-[1] opacity-[0.02]"
        style={{
          backgroundImage:
            "linear-gradient(var(--primary) 1px, transparent 1px), linear-gradient(90deg, var(--primary) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />

      <motion.div
        initial={{ opacity: 0, y: 30, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
        className="relative w-full max-w-md z-10"
      >
        {/* Glow behind card */}
        <div className="absolute -inset-4 bg-gradient-to-r from-[var(--primary)]/5 via-transparent to-[var(--accent)]/5 rounded-[32px] blur-[60px] opacity-50" />

        {/* Card */}
        <div className="relative rounded-3xl border border-[var(--border)] bg-[var(--surface)]/70 backdrop-blur-2xl shadow-2xl shadow-black/40 overflow-hidden">
          {/* Animated gradient border */}
          <motion.div
            className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-[var(--primary)] to-transparent"
            animate={{ opacity: [0.3, 0.7, 0.3] }}
            transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
          />

          <motion.div
            className="p-8 pt-10 space-y-7"
            variants={containerVariants}
            initial="hidden"
            animate="visible"
          >
            {/* Logo */}
            <motion.div variants={itemVariants} className="text-center space-y-3">
              <motion.div
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: 0.15, duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
                className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-[var(--primary)] to-[var(--accent)] shadow-lg shadow-[var(--primary)]/20 mb-4"
              >
                <img src="/sync_logo.png" alt="sync" className="w-full h-full object-contain p-2" />
              </motion.div>
              <motion.h1
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2, duration: 0.5 }}
                className="text-[28px] font-bold tracking-tight text-[var(--foreground)]"
              >
                Welcome back
              </motion.h1>
              <motion.p
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.3, duration: 0.5 }}
                className="text-sm text-[var(--text-muted)] leading-relaxed"
              >
                Sign in to your account to continue where you left off
              </motion.p>
            </motion.div>

            {/* Error */}
            <motion.div variants={itemVariants}>
              {error && (
                <motion.div
                  initial={{ opacity: 0, y: -8, height: 0 }}
                  animate={{ opacity: 1, y: 0, height: "auto" }}
                  className="p-3.5 rounded-xl bg-red-500/8 border border-red-500/15 flex items-center gap-2.5"
                >
                  <AlertIcon className="w-4 h-4 text-[var(--error)] flex-shrink-0 mt-0.5" />
                  <span className="text-sm text-[var(--error)]">{error}</span>
                </motion.div>
              )}
            </motion.div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-5">
              <motion.div variants={itemVariants}>
                <InputField
                  label="Email"
                  type="email"
                  value={email}
                  onChange={setEmail}
                  icon={EmailIcon}
                  placeholder="you@example.com"
                  autoComplete="email"
                  autoFocus
                  isFocused={focusedField === "email"}
                  onFocus={() => setFocusedField("email")}
                  onBlur={() => setFocusedField(null)}
                />
              </motion.div>

              <motion.div variants={itemVariants}>
                <InputField
                  label="Password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={setPassword}
                  icon={LockIcon}
                  placeholder="Enter your password"
                  autoComplete="current-password"
                  isFocused={focusedField === "password"}
                  onFocus={() => setFocusedField("password")}
                  onBlur={() => setFocusedField(null)}
                  rightElement={
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className={`p-1.5 rounded-lg transition-all duration-200 ${
                        showPassword
                          ? "text-[var(--primary)] bg-[var(--primary)]/8"
                          : "text-[var(--text-muted)] hover:text-[var(--text-dim)] hover:bg-[var(--surface-2)]"
                      }`}
                      tabIndex={-1}
                    >
                      {showPassword ? (
                        <EyeOffIcon className="w-4 h-4" />
                      ) : (
                        <EyeIcon className="w-4 h-4" />
                      )}
                    </button>
                  }
                />
              </motion.div>

              <motion.div variants={itemVariants}>
                <motion.button
                  type="submit"
                  disabled={loading}
                  whileHover={!loading ? { scale: 1.01 } : {}}
                  whileTap={!loading ? { scale: 0.99 } : {}}
                  className="relative w-full py-3 rounded-xl bg-gradient-to-r from-[var(--primary)] to-[var(--accent)] text-white font-semibold text-sm transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed overflow-hidden group"
                >
                  <span className="relative z-10 flex items-center justify-center gap-2">
                    {loading ? (
                      <>
                        <SpinnerIcon className="w-4 h-4 animate-spin" />
                        Signing in...
                      </>
                    ) : (
                      <>
                        <span>Sign In</span>
                        <ArrowRightIcon className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
                      </>
                    )}
                  </span>
                  <div className="absolute inset-0 -translate-x-full group-hover:translate-x-full transition-transform duration-700 bg-gradient-to-r from-transparent via-white/10 to-transparent" />
                </motion.button>
              </motion.div>
            </form>

            {/* Divider */}
            <motion.div variants={itemVariants}>
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-[var(--border)]" />
                </div>
                <div className="relative flex justify-center text-xs">
                  <span className="px-3 bg-[var(--surface)]/70 text-[var(--text-muted)]">
                    New here?
                  </span>
                </div>
              </div>
            </motion.div>

            {/* Register link */}
            <motion.div variants={itemVariants}>
              <Link
                href="/register"
                className="block w-full py-2.5 rounded-xl border border-[var(--border)] text-sm font-medium text-[var(--foreground)] text-center hover:bg-[var(--surface-2)] hover:border-[var(--border-light)] transition-all duration-200"
              >
                Create an account
              </Link>
            </motion.div>
          </motion.div>
        </div>

        {/* Footer */}
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.6, duration: 0.5 }}
          className="text-center text-xs text-[var(--text-muted)] mt-6"
        >
          Protected by end-to-end encryption &middot; sync v2.0
        </motion.p>
      </motion.div>
    </div>
  );
}
