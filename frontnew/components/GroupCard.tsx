import Link from "next/link";
import type { Group } from "@/lib/types";

export default function GroupCard({ group }: { group: Group }) {
  return (
    <Link href={`/groups/${group.group_id}`} className="surface card">
      <div className="badge">Group</div>
      <h3 style={{ margin: 0 }}>{group.group_name}</h3>
      <p style={{ color: "var(--muted)", margin: 0 }}>{group.description}</p>
      <div className="post-meta">
        <span>{group.member_count ?? 0} members</span>
      </div>
    </Link>
  );
}
