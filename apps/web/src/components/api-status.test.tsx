import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import type { components } from "@/api/__generated__/schema";

// Mock the typed client so the component never hits the network.
const getMock = vi.fn();
vi.mock("@/api/client", () => ({
  client: {
    GET: (...args: unknown[]) => getMock(...args),
  },
}));

import { ApiStatus } from "./api-status";

// Typed against the generated schema: if the contract's ReadinessStatus
// shape changes, these fixtures fail to compile.
type Readiness = components["schemas"]["ReadinessStatus"];

const okBody: Readiness = { status: "ok", checks: { db: "ok" } };
const degradedBody: Readiness = { status: "degraded", checks: { db: "down" } };

function renderWithClient(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
  );
}

describe("ApiStatus", () => {
  beforeEach(() => {
    getMock.mockReset();
  });

  it("renders 'API: ok' when readiness status is ok", async () => {
    getMock.mockResolvedValue({ data: okBody });

    renderWithClient(<ApiStatus />);

    expect(await screen.findByText("API: ok")).toBeInTheDocument();
  });

  it("renders 'API: degraded' when readiness status is degraded", async () => {
    // openapi-fetch puts a 503 body in `error`, not `data`.
    getMock.mockResolvedValue({ error: degradedBody });

    renderWithClient(<ApiStatus />);

    expect(await screen.findByText("API: degraded")).toBeInTheDocument();
  });
});
