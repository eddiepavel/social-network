"use client";

import { Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import EmptyState from "@/components/EmptyState";
import Avatar from "@/components/Avatar";
import { searchUsers } from "@/lib/api";

function SearchContent() {
  const params = useSearchParams();
  const name = params.get("name")?.trim() ?? "";

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["search-users", name],
    queryFn: () => searchUsers(name),
    enabled: name.length > 0,
  });

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader title="Find people" />
        {!name ? (
          <EmptyState title="Search for a name" body="Type a first or last name to get started." />
        ) : null}
        {isLoading ? <p>Searching...</p> : null}
        {isError ? <p style={{ color: "#b42318" }}>{(error as Error).message}</p> : null}
        {!isLoading && name && data?.length === 0 ? (
          <EmptyState title="No matches" body="Try a different name or check spelling." />
        ) : null}
        <div className="grid">
          {data?.map((user) => (
            <Link key={user.user_id} href={`/profile/${user.user_id}`} className="surface card search-result">
              <Avatar
                src={user.avatar}
                name={`${user.first_name} ${user.last_name}`}
                size={40}
              />
              <div className="search-result-info">
                <strong>
                  {user.first_name} {user.last_name}
                </strong>
                <span className="post-meta">{user.nickname ? `@${user.nickname}` : "No nickname"}</span>
              </div>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={<p>Loading search...</p>}>
      <SearchContent />
    </Suspense>
  );
}
