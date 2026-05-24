"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import { Sidebar } from "@/components/Sidebar";
import { SelectedConvProvider } from "@/contexts/SelectedConvContext";

export default function ChatLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { user, loading } = useAuth();

  useEffect(() => {
    if (!loading && !user) {
      router.push("/login");
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-[var(--background)]">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 rounded-full border-2 border-[var(--primary)] border-t-transparent animate-spin" />
          <p className="text-sm text-[var(--text-muted)]">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) return null;

  return (
    <div className="flex h-screen bg-[var(--background)]">
      <SelectedConvProvider>
        <Sidebar />
        <main className="flex-1 flex flex-col min-w-0">{children}</main>
      </SelectedConvProvider>
    </div>
  );
}
