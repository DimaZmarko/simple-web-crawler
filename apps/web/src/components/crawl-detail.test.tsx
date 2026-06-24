import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import type { components } from "@/api/__generated__/schema";

const getMock = vi.fn();
vi.mock("@/api/client", () => ({
  client: {
    GET: (...args: unknown[]) => getMock(...args),
  },
}));

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
  }: {
    children: ReactNode;
    href: string;
  }) => <a href={href}>{children}</a>,
}));

import { CrawlDetail } from "./crawl-detail";

type Crawl = components["schemas"]["Crawl"];
type NotFoundError = components["schemas"]["NotFoundError"];

const queuedCrawl: Crawl = {
  id: "7d8f3b2a-1c4e-4a9b-8f6d-2e5c9a1b3d4f",
  seedUrl: "https://example.com",
  maxDepth: 3,
  maxPages: 750,
  status: "queued",
  createdAt: "2026-06-14T10:00:00Z",
  updatedAt: "2026-06-14T10:00:00Z",
};

function renderWithClient(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
  );
}

describe("CrawlDetail", () => {
  beforeEach(() => {
    getMock.mockReset();
  });

  it("renders the crawl config and queued status", async () => {
    getMock.mockResolvedValue({ data: queuedCrawl, response: { status: 200 } });

    renderWithClient(<CrawlDetail id={queuedCrawl.id} />);

    expect(await screen.findByText(queuedCrawl.id)).toBeInTheDocument();
    expect(screen.getAllByText("https://example.com").length).toBeGreaterThan(0);
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("750")).toBeInTheDocument();
    expect(screen.getAllByText("queued").length).toBeGreaterThan(0);
  });

  it("renders a not-found state on a 404", async () => {
    const notFound: NotFoundError = { message: "crawl not found" };
    getMock.mockResolvedValue({ error: notFound, response: { status: 404 } });

    renderWithClient(<CrawlDetail id="missing" />);

    expect(await screen.findByText(/crawl not found/i)).toBeInTheDocument();
  });
});
