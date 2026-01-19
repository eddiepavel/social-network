"use client";

import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import useSession from "@/hooks/useSession";
import { logoutUser } from "@/lib/api";
import Button from "@/components/Button";
import SearchBar from "@/components/SearchBar";

export default function Nav() {
  const { data: session } = useSession();
  const queryClient = useQueryClient();
  const router = useRouter();

  const logout = useMutation({
    mutationFn: logoutUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["session"] });
      router.push("/");
    },
  });

  const homeHref = session?.user_id ? "/feed" : "/";

  return (
    <nav className="navbar">
      <Link href={homeHref}>
        <strong>Pulse</strong>
      </Link>
      <SearchBar />
      <div className="nav-links">
        <Link href="/feed">Feed</Link>
        <Link href="/groups">Groups</Link>
        <Link href="/chat">Chat</Link>
        {session?.user_id ? (
          <Link href={`/profile/${session.user_id}`}>Profile</Link>
        ) : (
          <Link href="/login">Login</Link>
        )}
      </div>
      <div>
        {session?.user_id ? (
          <Button
            variant="ghost"
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
          >
            Sign out
          </Button>
        ) : (
          <Link className="button" href="/register">
            Join
          </Link>
        )}
      </div>
    </nav>
  );
}
