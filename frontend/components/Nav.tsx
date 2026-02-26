"use client";

import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import useSession from "@/hooks/useSession";
import { logoutUser } from "@/lib/api";
import Button from "@/components/Button";
import Avatar from "@/components/Avatar";
import SearchBar from "@/components/SearchBar";
import NotificationDropdown from "@/components/NotificationDropdown";
import Logo from "./Logo";

export default function Nav() {
  const { data: session } = useSession();
  const queryClient = useQueryClient();
  const router = useRouter();

  const logout = useMutation({
    mutationFn: logoutUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["session"] });
      router.push("/login");
    },
  });

  const homeHref = session?.user_id ? "/feed" : "/";

  return (
    <nav className="navbar">
      <Link href={homeHref}>
        <Logo size="large"/>
      </Link>
      <SearchBar />
      <div className="nav-links">
        <Link href="/feed">Feed</Link>
        <Link href="/groups">Groups</Link>
        <Link href="/chat">Chat</Link>
        {session?.user_id ? (
          <>
            <Link href={`/profile/${session.user_id}`}>Profile</Link>
            <NotificationDropdown />
          </>
        ) : (
          <Link href="/login">Login</Link>
        )}
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        {session?.user_id ? (
          <>
            <Link 
              href={`/profile/${session.user_id}`} 
              style={{ display: "flex", alignItems: "center", gap: 8, textDecoration: "none", color: "inherit" }}
            >
              <Avatar 
                src={session.avatar} 
                name={`${session.first_name} ${session.last_name}`}
                size={32}
              />
              <span style={{ fontWeight: 500 }}>
                {session.first_name} {session.last_name}
              </span>
            </Link>
            <Button
              variant="ghost"
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
            >
              Sign out
            </Button>
          </>
        ) : (
          <Link className="button" href="/register">
            Join
          </Link>
        )}
      </div>
    </nav>
  );
}
