"use client";

import type { ChangeEvent } from "react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import Button from "@/components/Button";
import FormField from "@/components/FormField";
import { registerUser, ApiError } from "@/lib/api";

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
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  const register = useMutation({
    mutationFn: registerUser,
    onSuccess: () => router.push("/feed"),
    onError: (error) => {
      setValidationErrors({});
    
      if (error instanceof ApiError && error.details && typeof error.details === 'object') {
        setValidationErrors(error.details);
      }
 
      if (error instanceof ApiError && error.code == "419"){
        console.log("never refresh")
        location.reload()
      }
    },
  });

  const update =
    (key: keyof typeof form) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm((prev) => ({ ...prev, [key]: event.target.value }));

    if (validationErrors[key]) {
      setValidationErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[key];
        return newErrors;
      });
    }
  };

  return (
    <div className="surface card" style={{ maxWidth: 620, margin: "0 auto" }}>
      <h2 style={{ margin: 0 }}>Create your account</h2>
      <p style={{ color: "var(--muted)" }}>Start shaping your circle.</p>
      <div className="grid two">
        <FormField 
          label="First name" 
          name="first_name" 
          value={form.first_name} 
          onChange={update("first_name")} 
          error={validationErrors.first_name}
        />
        <FormField 
          label="Last name" 
          name="last_name" 
          value={form.last_name} 
          onChange={update("last_name")} 
          error={validationErrors.last_name}
        />
      </div>
      <FormField
        label="Email"
        name="email"
        type="email"
        placeholder="you@email.com"
        value={form.email}
        onChange={update("email")}
        error={validationErrors.email}
      />
      <FormField
        label="Password"
        name="password"
        type="password"
        value={form.password}
        onChange={update("password")}
        error={validationErrors.password}
      />
      <FormField 
        label="Date of birth" 
        name="dob" 
        type="date" 
        value={form.dob} 
        onChange={update("dob")} 
        error={validationErrors.dob}
      />
      <FormField 
        label="Nickname" 
        name="nickname" 
        value={form.nickname} 
        onChange={update("nickname")} 
        error={validationErrors.nickname}
      />
      {register.isError ? (
        register.error instanceof ApiError && typeof register.error.details === 'object' ? null : (
          <p style={{ color: "#b42318" }}>
            {register.error instanceof ApiError && typeof register.error.details === 'string'
              ? register.error.details
              : register.error.message}
          </p>
        )
      ) : null}
      <Button
        onClick={() =>
          register.mutate({
            ...form,
          })
        }
      >
        Create account
      </Button>
      <p style={{ marginTop: 12 }}>
        Already registered? <Link href="/login">Sign in</Link>
      </p>
    </div>
  );
}
