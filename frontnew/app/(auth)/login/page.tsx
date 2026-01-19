"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { loginUser } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const login = useMutation({
    mutationFn: loginUser,
    onSuccess: () => router.push("/feed"),
  });

  return (
    <div className="surface card" style={{ maxWidth: 520, margin: "0 auto" }}>
      <h2 style={{ margin: 0 }}>Welcome back</h2>
      <p style={{ color: "var(--muted)" }}>Sign in to reconnect with your circles.</p>
      <FormField
        label="Email"
        name="email"
        type="email"
        placeholder="you@email.com"
        value={email}
        onChange={(event) => setEmail(event.target.value)}
      />
      <FormField
        label="Password"
        name="password"
        type="password"
        value={password}
        onChange={(event) => setPassword(event.target.value)}
      />
      {login.isError ? <p style={{ color: "#b42318" }}>{login.error.message}</p> : null}
      <Button
        onClick={() => login.mutate({ email, password })}
        disabled={!email || !password || login.isPending}
      >
        Sign in
      </Button>
      <p style={{ marginTop: 12 }}>
        New here? <Link href="/register">Create an account</Link>
      </p>
    </div>
  );
}
