"use client";

import { useQuery } from "@tanstack/react-query";
import { getSession } from "@/lib/api";

export default function useSession() {
  return useQuery({
    queryKey: ["session"],
    queryFn: getSession,
  });
}
