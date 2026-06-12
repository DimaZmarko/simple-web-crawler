"use client";

import { useQuery } from "@tanstack/react-query";
import { client } from "@/api/client";

async function fetchReadiness() {
  const { data, error } = await client.GET("/readyz", {});
  // openapi-fetch delivers non-2xx bodies in `error`, so a 503 carries the
  // degraded ReadinessStatus there. Surface whichever the response produced;
  // a genuine transport failure rejects the promise before this point.
  return data ?? error;
}

export function ApiStatus() {
  const { data, isPending, isError } = useQuery({
    queryKey: ["readyz"],
    queryFn: fetchReadiness,
  });

  if (isPending) {
    return <p>Checking API status…</p>;
  }

  const isOk = !isError && data?.status === "ok";

  return <p>API: {isOk ? "ok" : "degraded"}</p>;
}
