"use client";

import type { ChangeEvent } from "react";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import SectionHeader from "@/components/SectionHeader";
import { getUserProfile, updatePrivacy, updateProfile } from "@/lib/api";
import useSession from "@/hooks/useSession";

export default function ProfilePage() {
  const params = useParams();
  const userId = Array.isArray(params.id) ? params.id[0] : (params.id as string);
  const { data: session } = useSession();
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["profile", userId],
    queryFn: () => getUserProfile(userId),
    enabled: !!userId,
  });

  const isSelf = useMemo(() => session?.user_id === userId, [session?.user_id, userId]);
  const [form, setForm] = useState({
    first_name: "",
    last_name: "",
    nickname: "",
    about_me: "",
  });

  useEffect(() => {
    if (!data) return;
    setForm({
      first_name: data.first_name || "",
      last_name: data.last_name || "",
      nickname: data.nickname || "",
      about_me: data.about_me || "",
    });
  }, [data]);

  const saveProfile = useMutation({
    mutationFn: updateProfile,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["profile", userId] }),
  });

  const updatePrivacyMutation = useMutation({
    mutationFn: updatePrivacy,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["profile", userId] }),
  });

  const updateField =
    (key: keyof typeof form) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm((prev) => ({ ...prev, [key]: event.target.value }));
    };

  if (isLoading) return <p>Loading profile...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data) return <p>No profile found.</p>;

  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <SectionHeader title="Profile" />
        <p style={{ marginTop: 0, color: "var(--muted)" }}>
          {data.first_name} {data.last_name} · {data.email}
        </p>
        <p>{data.about_me || "No bio yet."}</p>
        <div className="post-meta">
          <span>Nickname: {data.nickname || "—"}</span>
          <span>Privacy: {data.is_public ? "Public" : "Private"}</span>
        </div>
      </section>

      {isSelf ? (
        <section className="surface card">
          <SectionHeader title="Edit profile" />
          <FormField label="First name" name="first_name" value={form.first_name} onChange={updateField("first_name")} />
          <FormField label="Last name" name="last_name" value={form.last_name} onChange={updateField("last_name")} />
          <FormField label="Nickname" name="nickname" value={form.nickname} onChange={updateField("nickname")} />
          <FormField label="About" name="about_me" as="textarea" value={form.about_me} onChange={updateField("about_me")} />
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
            <Button
              onClick={() => saveProfile.mutate(form)}
              disabled={saveProfile.isPending}
            >
              Save profile
            </Button>
            <Button
              variant="ghost"
              onClick={() => updatePrivacyMutation.mutate(!(data.is_public ?? false))}
              disabled={updatePrivacyMutation.isPending}
            >
              Toggle privacy
            </Button>
          </div>
          {saveProfile.isError ? (
            <p style={{ color: "#b42318" }}>{saveProfile.error.message}</p>
          ) : null}
        </section>
      ) : null}
    </div>
  );
}
