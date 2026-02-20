"use client";

import type { ChangeEvent } from "react";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import SectionHeader from "@/components/SectionHeader";
import Avatar from "@/components/Avatar";
import FollowButton from "@/components/FollowButton";
import FollowersList from "@/components/FollowersList";
import FollowRequestsList from "@/components/FollowRequestsList";
import Tabs from "@/components/Tabs";
import ImageUpload from "@/components/ImageUpload";
import PostCard from "@/components/PostCard";
import { getUserProfile, updatePrivacy, updateProfile, getFollowers, getFollowing, getUserPosts, ApiError } from "@/lib/api";
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

  const { data: userPosts, isLoading: postsLoading, isError: postsError } = useQuery({
    queryKey: ["userPosts", userId],
    queryFn: () => getUserPosts(userId, 1, 20),
    retry: false, // Don't retry if unauthorized
  });

  const [activeTab, setActiveTab] = useState("posts");
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [form, setForm] = useState({
    first_name: "",
    last_name: "",
    nickname: "",
    about_me: "",
    avatar_id: "",
  });

  useEffect(() => {
    if (!data) return;
    setForm({
      first_name: data.first_name || "",
      last_name: data.last_name || "",
      nickname: data.nickname || "",
      about_me: data.about_me || "",
      avatar_id: data.avatar_id || ""
    });
  }, [data]);

  const saveProfile = useMutation({
    mutationFn: updateProfile,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["profile", userId] }),
    onError: (error) => {
      setValidationErrors({});
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      }
    },
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

  const handleAvatarSelect = (file: File) => {
    setAvatarFile(file);
    const reader = new FileReader();
    reader.onload = (e) => setAvatarPreview(e.target?.result as string);
    reader.readAsDataURL(file);
  };

  const handleSaveProfile = async () => {
    let avatarId = form.avatar_id;
    if (avatarFile) {
      const { uploadFile } = await import("@/lib/api");
      const uploaded = await uploadFile(avatarFile);
      avatarId = uploaded.filename;
    }
    saveProfile.mutate({ ...form, avatar_id: avatarId });
    setAvatarFile(null);
    setAvatarPreview(null);
  };

  if (isLoading) return <p>Loading profile...</p>;
  if (isError) return <p style={{ color: "#b42318" }}>{(error as Error).message}</p>;
  if (!data) return <p>No profile found.</p>;


  return (
    <div className="grid" style={{ paddingBottom: 64 }}>
      <section className="surface card">
        <div className="profile-header">
          <Avatar
            src={data.avatar}
            name={`${data.first_name} ${data.last_name}`}
            size={80}
          />
          <div className="profile-header-info">
            <div className="profile-header-top">
              <h1 style={{ margin: 0, fontSize: "1.5rem" }}>
                {data.first_name} {data.last_name}
              </h1>
              <FollowButton userId={userId} currentUserId={session?.user_id} />
            </div>
            <p style={{ margin: "4px 0", color: "var(--muted)" }}>{data.email}</p>
            <div className="profile-stats">
              <span><strong>{data.followers}</strong> followers</span>
              <span><strong>{data.following}</strong> following</span>
            </div>
          </div>
        </div>
        <p>{data.about_me || "No bio yet."}</p>
        <div className="post-meta">
          {data.nickname && <span>@{data.nickname}</span>}
          <span className="tag">{data.is_public ? "Public" : "Private"}</span>
        </div>
      </section>
      {data.can_view && (
        <section className="surface card">
          <Tabs
            tabs={[
              { id: "posts", label: `Posts (${userPosts?.length ?? 0})` },
              { id: "followers", label: `Followers (${data.followers})` },
              { id: "following", label: `Following (${data.following})` },
              ...(isSelf ? [{ id: "requests", label: "Requests" }] : []),
            ]}
            activeTab={activeTab}
            onChange={setActiveTab}
          />
          {activeTab === "posts" && (
            <div style={{ marginTop: 16 }}>
              {postsLoading ? (
                <p>Loading posts...</p>
              ) : userPosts && userPosts.length > 0 ? (
                userPosts.map((post) => (
                  <PostCard key={post.post_id} post={post} currentUserId={session?.user_id} />
                ))
              ) : (
                <p style={{ textAlign: "center", color: "var(--muted)", padding: "2rem" }}>
                  {data?.is_public === false && !isSelf
                    ? "This user's profile is private. Follow them to see their posts."
                    : "No posts yet."}
                </p>
              )}
            </div>
          )}
          {activeTab === "followers" && <FollowersList userId={userId} type="followers" />}
          {activeTab === "following" && <FollowersList userId={userId} type="following" />}
          {activeTab === "requests" && isSelf && <FollowRequestsList />}
        </section>
      )}


      {isSelf ? (
        <section className="surface card">
          <SectionHeader title="Edit profile" />
          <div className="edit-avatar-section">
            <Avatar
              src={avatarPreview || data.avatar}
              name={`${data.first_name} ${data.last_name}`}
              size={80}
            />
            <ImageUpload
              onImageSelect={handleAvatarSelect}
              accept="image/*"
              maxSizeMB={5}
              label="Change avatar"
              compact
            />
          </div>
          <FormField label="First name" name="first_name" value={form.first_name} onChange={(e) => { updateField("first_name")(e); if (validationErrors.first_name) setValidationErrors(prev => { const n = { ...prev }; delete n.first_name; return n; }); }} error={validationErrors.first_name} />
          <FormField label="Last name" name="last_name" value={form.last_name} onChange={(e) => { updateField("last_name")(e); if (validationErrors.last_name) setValidationErrors(prev => { const n = { ...prev }; delete n.last_name; return n; }); }} error={validationErrors.last_name} />
          <FormField label="Nickname" name="nickname" value={form.nickname} onChange={(e) => { updateField("nickname")(e); if (validationErrors.nickname) setValidationErrors(prev => { const n = { ...prev }; delete n.nickname; return n; }); }} error={validationErrors.nickname} />
          <FormField label="About" name="about_me" as="textarea" value={form.about_me} onChange={(e) => { updateField("about_me")(e); if (validationErrors.about_me) setValidationErrors(prev => { const n = { ...prev }; delete n.about_me; return n; }); }} error={validationErrors.about_me} />
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
            <Button
              onClick={handleSaveProfile}
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
            saveProfile.error instanceof ApiError && typeof saveProfile.error.details === 'object' ? null : (
              <p style={{ color: "#b42318" }}>
                {saveProfile.error instanceof ApiError && typeof saveProfile.error.details === 'string'
                  ? saveProfile.error.details
                  : saveProfile.error.message}
              </p>
            )
          ) : null}
        </section>
      ) : null}
    </div>
  );
}
