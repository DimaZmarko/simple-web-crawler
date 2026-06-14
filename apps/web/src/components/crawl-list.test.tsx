import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
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

import { CrawlList } from "./crawl-list";

type CrawlList = components["schemas"]["CrawlList"];

const page1: CrawlList = {
  items: [
    {
      id: "7d8f3b2a-1c4e-4a9b-8f6d-2e5c9a1b3d4f",
      seedUrl: "https://example.com",
      status: "queued",
      createdAt: "2026-06-14T10:00:00Z",
    },
    {
      id: "1a2b3c4d-5e6f-4071-8293-a4b5c6d7e8f9",
      seedUrl: "https://another.example.org",
      status: "queued",
      createdAt: "2026-06-14T09:45:00Z",
    },
  ],
  nextCursor: "cursor-page-2",
};

const emptyList: CrawlList = { items: [], nextCursor: null };

function renderWithClient(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
  );
}

describe("CrawlList", () => {
  beforeEach(() => {
    getMock.mockReset();
  });

  it("renders items from the list response with status and link to detail", async () => {
    getMock.mockResolvedValue({ data: page1 });

    renderWithClient(<CrawlList />);

    expect(await screen.findByText("https://example.com")).toBeInTheDocument();
    expect(
      screen.getByText("https://another.example.org"),
    ).toBeInTheDocument();

    const link = screen.getByRole("link", { name: "https://example.com" });
    expect(link).toHaveAttribute(
      "href",
      "/crawls/7d8f3b2a-1c4e-4a9b-8f6d-2e5c9a1b3d4f",
    );
    expect(screen.getAllByText("queued")).toHaveLength(2);
  });

  it("passes nextCursor to the API when 'Load more' is clicked", async () => {
    getMock.mockResolvedValue({ data: page1 });

    renderWithClient(<CrawlList />);
    await screen.findByText("https://example.com");

    // First call is the initial page (no cursor).
    expect(getMock).toHaveBeenNthCalledWith(1, "/crawls", {
      params: { query: {} },
    });

    fireEvent.click(screen.getByRole("button", { name: /load more/i }));

    await waitFor(() =>
      expect(getMock).toHaveBeenCalledWith("/crawls", {
        params: { query: { cursor: "cursor-page-2" } },
      }),
    );
  });

  it("renders an empty state when there are no crawls", async () => {
    getMock.mockResolvedValue({ data: emptyList });

    renderWithClient(<CrawlList />);

    expect(await screen.findByText(/no crawls yet/i)).toBeInTheDocument();
  });
});
