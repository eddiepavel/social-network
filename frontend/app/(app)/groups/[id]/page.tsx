"use client";

import { useState, useMemo } from "react";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import SectionHeader from "@/components/SectionHeader";
import Button from "@/components/Button";
import Tabs from "@/components/Tabs";
import MemberCard from "@/components/MemberCard";
import GroupInviteModal from "@/components/GroupInviteModal";
import GroupRequestsList from "@/components/GroupRequestsList";
import GroupSettings from "@/components/GroupSettings";
import EventsList from "@/components/EventsList";
import CreateEventModal from "@/components/CreateEventModal";
import {getGroup, requestJoinGroup, leaveGroup, getGroupRequests} from "@/lib/api";
import useSession from "@/hooks/useSession";
import GroupPosts from "@/components/GroupPosts";

export default function GroupDetailsPage() {
  const params = useParams();
  const groupId = Array.isArray(params.id) ? params.id[0] : (params.id as string);
  const queryClient = useQueryClient();
  const { data: session } = useSession();

  const [activeTab, setActiveTab] = useState("members");
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isCreateEventOpen, setIsCreateEventOpen] = useState(false);

  const { data, isLoading, isError, error, refetch: refetchGroup } = useQuery({
    queryKey: ["group", groupId],
    queryFn: () => getGroup(groupId),
    enabled: !!groupId,
  });

  const requestJoin = useMutation({
    mutationFn: (payload: any) => requestJoinGroup(groupId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["group", groupId] }),
  });

  const leave = useMutation({
    mutationFn: () => leaveGroup(groupId, session?.user_id ?? ""),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["group", groupId] });
      queryClient.invalidateQueries({ queryKey: ["groups"] });
    },
  });

  const isMember = useMemo(() => {
    if (!session?.user_id || !data?.members) return false;
    return data.members.some(
      (m) => m.user_id === session.user_id && m.status === "joined"
    );
  }, [session?.user_id, data?.members]);

  const isOwner = data?.group.is_owner ?? false;

  const {
    data: requests,
    isLoading: requestsIsLoading,
    isError: requestsIsError,
    error: requestsError
  } = useQuery({
    queryKey: ["group-requests", groupId],
    queryFn: () => getGroupRequests(groupId),
    enabled: !!groupId && isOwner,
  });

  const acceptedMembers = useMemo(() => {
    return data ? data.members?.filter((m) => m.status === "joined") : [];
  }, [data?.members]);

  if (isLoading) return <p>Loading group...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data) return <p>Group not found.</p>;

  const memberIds = data?.members?.map((m) => m.user_id);

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader
          title={data.group.group_name}
          action={
            <div style={{ display: "flex", gap: 8 }}>
              {(isOwner || data.group.user_status == "joined") && (
                <>
                  <Button variant="ghost" onClick={() => setIsInviteOpen(true)}>
                    Invite
                  </Button>
                  {isOwner && (<Button variant="ghost" onClick={() => setIsSettingsOpen(true)}>
                    Settings
                  </Button>)}

                </>
              )}
              {data.group.user_status == "not_member" && (
                <Button
                  onClick={() => requestJoin.mutate({ action: "request" })}
                  disabled={requestJoin.isPending}
                >
                  Request to join
                </Button>

              )}
              {data.group.user_status == "invited" && (
                <>
                  <Button
                    onClick={() => requestJoin.mutate({ action: "accept_invite" })}
                    disabled={requestJoin.isPending}
                  >
                    Accept
                  </Button>
                  <Button
                    variant="ghost"
                    onClick={() => requestJoin.mutate({ action: "decline_invite" })}
                    disabled={requestJoin.isPending}
                  >
                    Decline
                  </Button>
                </>
              )}
              {data.group.user_status == "requested" && (
                <Button variant="ghost"
                  onClick={() => requestJoin.mutate({ action: "remove" })}
                  disabled={requestJoin.isPending}
                >
                  Request pending
                </Button>
              )}
              {data.group.user_status == "joined" && !isOwner && (
                <Button
                  variant="ghost"
                  onClick={() => leave.mutate()}
                  disabled={leave.isPending}
                >
                  Leave group
                </Button>
              )}
            </div>
          }
        />
        <p style={{ color: "var(--muted)" }}>{data.group.description}</p>
        <div className="post-meta" style={{ display: "flex", alignItems: "center" }}>
          <span>{data.group.total_members ?? acceptedMembers?.length} members</span>
          {isOwner && <span className="tag">Owner</span>}
          {isMember && !isOwner && <span className="tag">Member</span>}
        </div>
      </section>

      {acceptedMembers?.length != 0 && data.members && (<section className="surface card">
        <Tabs
          tabs={[
            { id: "members", label: `Members (${acceptedMembers?.length})` },
            ...(isMember? [{ id: "posts", label: "Posts"}] : []),
            ...(isMember ? [{ id: "events", label: `Events (${data.events?.length ?? 0})` }] : []),
            ...(isOwner ? [{ id: "requests", label: `Join Requests (${requests?.length ?? 0})` }] : []),
          ]}
          activeTab={activeTab}
          onChange={setActiveTab}
        />

        {activeTab === "members" && (
          <div className="members-grid">
            {acceptedMembers?.map((member) => (
              <MemberCard
                key={member.user_id}
                member={member}
                groupId={groupId}
                canRemove={isOwner}
                isCurrentUser={member.user_id === session?.user_id}
              />
            ))}
          </div>
        )}

        {activeTab === "posts" && isMember && (
            <GroupPosts groupName={data.group.group_name} groupId={groupId} session={session}/>
        )}


        {activeTab === "events" && isMember && (
          <div>
            <div style={{ display: "flex", justifyContent: "flex-end", marginBottom: 16 }}>
              <Button onClick={() => setIsCreateEventOpen(true)}>
                Create Event
              </Button>
            </div>
            <EventsList events={data.events ?? []} />
          </div>
        )}

        {activeTab === "requests" && isOwner && (
          <GroupRequestsList groupId={groupId} data={requests} isLoading={requestsIsLoading} isError={requestsIsError} error={requestsError}/>
        )}
      </section>)}


      {(isOwner || data.group.user_status == "joined") && (
        <>
          <GroupInviteModal
            isOpen={isInviteOpen}
            onClose={() => setIsInviteOpen(false)}
            groupId={groupId}
            existingMemberIds={memberIds ?? []}
          />
          <GroupSettings
            group={data.group}
            isOpen={isSettingsOpen}
            onClose={() => setIsSettingsOpen(false)}
          />
          <CreateEventModal
            isOpen={isCreateEventOpen}
            onClose={() => setIsCreateEventOpen(false)}
            groupId={groupId}
            onEventCreated={() => refetchGroup()}
          />
        </>
      )}
    </div>
  );
}
