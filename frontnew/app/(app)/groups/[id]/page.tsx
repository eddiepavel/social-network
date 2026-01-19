"use client";

import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import Button from "@/components/Button";
import { getGroup, requestJoinGroup } from "@/lib/api";

export default function GroupDetailsPage() {
  const params = useParams();
  const groupId = Array.isArray(params.id) ? params.id[0] : (params.id as string);
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["group", groupId],
    queryFn: () => getGroup(groupId),
    enabled: !!groupId,
  });

  const requestJoin = useMutation({
    mutationFn: () => requestJoinGroup(groupId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["group", groupId] }),
  });

  if (isLoading) return <p>Loading group...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data) return <p>Group not found.</p>;

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader
          title={data.group.group_name}
          action={
            <Button
              variant="ghost"
              onClick={() => requestJoin.mutate()}
              disabled={requestJoin.isPending}
            >
              Request to join
            </Button>
          }
        />
        <p style={{ color: "var(--muted)" }}>{data.group.description}</p>
        <div className="post-meta">
          <span>{data.group.member_count ?? data.members.length} members</span>
          <span>{data.group.is_owner ? "You are the owner" : ""}</span>
        </div>
      </section>
      <section className="surface card">
        <SectionHeader title="Members" />
        <div className="grid two">
          {data.members.map((member) => (
            <div key={member.user_id} className="surface card" style={{ boxShadow: "none" }}>
              <strong>
                {member.first_name} {member.last_name}
              </strong>
              <span className="post-meta">{member.status}</span>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
