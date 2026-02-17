import type { ReactNode } from "react";
import AppShell from "@/components/AppShell";
import AuthGuard from "@/components/AuthGuard";

export default function AppLayout({ children }: { children: ReactNode }) {
  return <AuthGuard requireAuth={true}><AppShell>{children}</AppShell></AuthGuard>
}
