"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import useSession from "@/hooks/useSession";

type AuthGuardProps = {
  children: React.ReactNode;
  requireAuth?: boolean;
};

export default function AuthGuard({ children, requireAuth = false }: AuthGuardProps) {
  const { data: session, isLoading } = useSession();
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;

    // For auth pages (login/register), redirect if authenticated
    if (!requireAuth && session) {
      router.push("/feed");
    }

    // For protected pages, redirect if not authenticated
    if (requireAuth && !session) {
      router.push("/login");
    }
  }, [session, isLoading, requireAuth, router]);

  // Show loading state while checking authentication
  if (isLoading) {
    return (
      <div className="page">
        <main className="container" style={{ padding: "72px 0", textAlign: "center" }}>
          <p>Loading...</p>
        </main>
      </div>
    );
  }

  // For auth pages, don't render if authenticated (redirecting)
  if (!requireAuth && session) {
    return null;
  }

  // For protected pages, don't render if not authenticated (redirecting)
  if (requireAuth && !session) {
    return null;
  }

  return <>{children}</>;
}
