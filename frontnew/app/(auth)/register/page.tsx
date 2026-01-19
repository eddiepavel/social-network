"use client";

import type { ChangeEvent } from "react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { registerUser } from "@/lib/api";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    first_name: "",
    last_name: "",
    email: "",
    password: "",
    dob: "",
    nickname: "",
  });

  const register = useMutation({
    mutationFn: registerUser,
    onSuccess: () => router.push("/feed"),
  });

  const update =
    (key: keyof typeof form) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm((prev) => ({ ...prev, [key]: event.target.value }));
  };

  return (
    <div className="surface card" style={{ maxWidth: 620, margin: "0 auto" }}>
      <h2 style={{ margin: 0 }}>Create your account</h2>
      <p style={{ color: "var(--muted)" }}>Start shaping your circle.</p>
      <div className="grid two">
        <FormField label="First name" name="first_name" value={form.first_name} onChange={update("first_name")} />
        <FormField label="Last name" name="last_name" value={form.last_name} onChange={update("last_name")} />
      </div>
      <FormField
        label="Email"
        name="email"
        type="email"
        placeholder="you@email.com"
        value={form.email}
        onChange={update("email")}
      />
      <FormField
        label="Password"
        name="password"
        type="password"
        value={form.password}
        onChange={update("password")}
      />
      <FormField label="Date of birth" name="dob" type="date" value={form.dob} onChange={update("dob")} />
      <FormField label="Nickname" name="nickname" value={form.nickname} onChange={update("nickname")} />
      {register.isError ? (
        <p style={{ color: "#b42318" }}>{register.error.message}</p>
      ) : null}
      <Button
        onClick={() =>
          register.mutate({
            ...form,
          })
        }
        disabled={!form.email || !form.password || !form.first_name || !form.last_name || !form.dob}
      >
        Create account
      </Button>
      <p style={{ marginTop: 12 }}>
        Already registered? <Link href="/login">Sign in</Link>
      </p>
    </div>
  );
}
