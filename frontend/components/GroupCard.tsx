import Link from "next/link";
import type { Group } from "@/lib/types";

export default function GroupCard({ group, userId }: { group: Group, userId?: string }) {
  return (
    <Link href={`/groups/${group.group_id}`} className="surface card">
      <div className="badge">
          <div>Group</div>
          <div>{group.creator_id === userId && ("Owner")}</div>
      </div>
        <h3 style={{ margin: 0, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
            {group.group_name}
        </h3>
        <p style={{
            color: "var(--muted)",
            margin: 0,
            overflow: "hidden",
            textOverflow: "ellipsis",
            display: "-webkit-box",
            WebkitLineClamp: 2,
            WebkitBoxOrient: "vertical",
            lineHeight: "1.4em",
            maxHeight: "2.8em"
        }}>
            {group.description}
        </p>
      <div className="post-meta">
        <span>{group.member_count ?? 0} members</span>
      </div>
    </Link>
  );
}
