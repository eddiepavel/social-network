import type { ReactNode } from "react";
import Nav from "@/components/Nav";

export default function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="page">
      <div className="container">
        <Nav />
      </div>
      <main className="container">{children}</main>
    </div>
  );
}
