import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import type { components } from "@/api/__generated__/schema";

const postMock = vi.fn();
vi.mock("@/api/client", () => ({
  client: {
    POST: (...args: unknown[]) => postMock(...args),
  },
}));

const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

import { CrawlForm } from "./crawl-form";

type Crawl = components["schemas"]["Crawl"];
type ValidationError = components["schemas"]["ValidationError"];

const queuedCrawl: Crawl = {
  id: "7d8f3b2a-1c4e-4a9b-8f6d-2e5c9a1b3d4f",
  seedUrl: "https://example.com",
  maxDepth: 2,
  maxPages: 500,
  status: "queued",
  createdAt: "2026-06-14T10:00:00Z",
  updatedAt: "2026-06-14T10:00:00Z",
};

function renderWithClient(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
  );
}

function setSeedUrl(value: string) {
  fireEvent.change(screen.getByLabelText(/seed url/i), {
    target: { value },
  });
}

function submit() {
  fireEvent.click(screen.getByRole("button", { name: /submit crawl/i }));
}

describe("CrawlForm", () => {
  beforeEach(() => {
    postMock.mockReset();
    pushMock.mockReset();
  });

  it("shows a client-side field error for an invalid URL and does not call the API", async () => {
    renderWithClient(<CrawlForm />);

    setSeedUrl("not-a-url");
    submit();

    expect(
      await screen.findByText(/must be an absolute http or https url/i),
    ).toBeInTheDocument();
    expect(postMock).not.toHaveBeenCalled();
  });

  it("submits a valid request, sends the typed body, and navigates to the detail page", async () => {
    postMock.mockResolvedValue({ data: queuedCrawl, response: { status: 202 } });

    renderWithClient(<CrawlForm />);

    // maxDepth (2) and maxPages (500) keep their default values.
    setSeedUrl("https://example.com");
    submit();

    await waitFor(() => expect(postMock).toHaveBeenCalledTimes(1));
    expect(postMock).toHaveBeenCalledWith("/crawls", {
      body: { seedUrl: "https://example.com", maxDepth: 2, maxPages: 500 },
    });
    await waitFor(() =>
      expect(pushMock).toHaveBeenCalledWith(`/crawls/${queuedCrawl.id}`),
    );
  });

  it("renders server-side 400 field errors next to the relevant field", async () => {
    const validationError: ValidationError = {
      message: "Request validation failed",
      errors: [{ field: "seedUrl", message: "URL is not reachable" }],
    };
    postMock.mockResolvedValue({
      error: validationError,
      response: { status: 400 },
    });

    renderWithClient(<CrawlForm />);

    setSeedUrl("https://example.com");
    submit();

    expect(
      await screen.findByText(/url is not reachable/i),
    ).toBeInTheDocument();
    expect(pushMock).not.toHaveBeenCalled();
  });
});
