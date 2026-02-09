import type { ReactNode } from "react";
import AuthGuard from "@/components/AuthGuard";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <AuthGuard requireAuth={false}>
      <div className="page">
        <main className="container" style={{ padding: "72px 0" }}>
          {children}
        </main>
      </div>
    </AuthGuard>
  );
}
