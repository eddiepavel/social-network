"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import GroupCard from "@/components/GroupCard";
import SectionHeader from "@/components/SectionHeader";
import EmptyState from "@/components/EmptyState";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { createGroup, getGroups } from "@/lib/api";
import { useState } from "react";

export default function GroupsPage() {
  const queryClient = useQueryClient();
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["groups"],
    queryFn: getGroups,
  });

  const [form, setForm] = useState({ group_name: "", description: "" });
  const create = useMutation({
    mutationFn: createGroup,
    onSuccess: () => {
      setForm({ group_name: "", description: "" });
      queryClient.invalidateQueries({ queryKey: ["groups"] });
    },
  });

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader title="Groups" />
        {isLoading ? <p>Loading groups...</p> : null}
        {isError ? <p style={{ color: "#b42318" }}>{(error as Error).message}</p> : null}
        {!isLoading && data?.length === 0 ? (
          <EmptyState
            title="No groups yet"
            body="Create the first group to start building your community."
          />
        ) : null}
        <div className="grid two">
          {data?.map((group) => (
            <GroupCard key={group.group_id} group={group} />
          ))}
        </div>
      </section>
      <section className="surface card">
        <SectionHeader title="Start a new group" />
        <FormField
          label="Group name"
          name="group_name"
          value={form.group_name}
          onChange={(event) => setForm((prev) => ({ ...prev, group_name: event.target.value }))}
        />
        <FormField
          label="Description"
          name="description"
          as="textarea"
          value={form.description}
          onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
        />
        {create.isError ? (
          <p style={{ color: "#b42318" }}>{create.error.message}</p>
        ) : null}
        <Button
          onClick={() => create.mutate(form)}
          disabled={!form.group_name || !form.description || create.isPending}
        >
          Create group
        </Button>
      </section>
    </div>
  );
}
