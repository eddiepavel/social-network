import type { ReactNode } from "react";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="page">
      <main className="container" style={{ padding: "72px 0" }}>
        {children}
      </main>
    </div>
  );
}
