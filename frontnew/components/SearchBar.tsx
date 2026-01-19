"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function SearchBar() {
  const router = useRouter();
  const [value, setValue] = useState("");

  return (
    <form
      className="search"
      onSubmit={(event) => {
        event.preventDefault();
        const term = value.trim();
        if (!term) return;
        router.push(`/search?name=${encodeURIComponent(term)}`);
      }}
    >
      <input
        type="search"
        name="name"
        placeholder="Search people"
        value={value}
        onChange={(event) => setValue(event.target.value)}
      />
      <button type="submit">Go</button>
    </form>
  );
}
