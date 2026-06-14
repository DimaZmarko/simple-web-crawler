"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryOptions,
} from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { client } from "@/api/client";
import type { components } from "@/api/__generated__/schema";

// All shapes derive from the generated client — never hand-written.
export type Crawl = components["schemas"]["Crawl"];
export type CrawlSummary = components["schemas"]["CrawlSummary"];
export type CrawlList = components["schemas"]["CrawlList"];
export type CreateCrawlRequest = components["schemas"]["CreateCrawlRequest"];
export type ValidationError = components["schemas"]["ValidationError"];
export type NotFoundError = components["schemas"]["NotFoundError"];

export const crawlKeys = {
  all: ["crawls"] as const,
  list: (cursor?: string) => ["crawls", "list", cursor ?? null] as const,
  detail: (id: string) => ["crawls", "detail", id] as const,
};

// Raised by useCreateCrawl when the API rejects the body with a 400. Carries the
// per-field errors from ValidationError so the form can render them inline.
export class CrawlValidationError extends Error {
  readonly errors: ValidationError["errors"];
  constructor(validation: ValidationError) {
    super(validation.message);
    this.name = "CrawlValidationError";
    this.errors = validation.errors;
  }
}

/**
 * Mutation wrapping POST /crawls. On a 202 it invalidates the list cache,
 * seeds the detail cache, and navigates to the new crawl's detail page.
 * A 400 rejects with a CrawlValidationError carrying field-level messages.
 */
export function useCreateCrawl() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation<Crawl, Error, CreateCrawlRequest>({
    mutationFn: async (body) => {
      const { data, error, response } = await client.POST("/crawls", { body });
      if (data) {
        return data;
      }
      if (response.status === 400 && error) {
        throw new CrawlValidationError(error as ValidationError);
      }
      throw new Error(
        (error as { message?: string } | undefined)?.message ??
          "Failed to create crawl",
      );
    },
    onSuccess: (crawl) => {
      queryClient.setQueryData(crawlKeys.detail(crawl.id), crawl);
      queryClient.invalidateQueries({ queryKey: crawlKeys.all });
      router.push(`/crawls/${crawl.id}`);
    },
  });
}

/** Query wrapping GET /crawls?cursor — one page of crawl summaries. */
export function useCrawlList(cursor?: string) {
  return useQuery<CrawlList>({
    queryKey: crawlKeys.list(cursor),
    queryFn: async () => {
      const { data, error } = await client.GET("/crawls", {
        params: { query: cursor ? { cursor } : {} },
      });
      if (error) {
        throw new Error(
          (error as { message?: string } | undefined)?.message ??
            "Failed to load crawls",
        );
      }
      return data;
    },
  });
}

/** Query wrapping GET /crawls/{id}. A 404 surfaces as `notFound: true`. */
export function useCrawl(
  id: string,
  options?: Partial<UseQueryOptions<Crawl | null>>,
) {
  return useQuery<Crawl | null>({
    queryKey: crawlKeys.detail(id),
    queryFn: async () => {
      const { data, error, response } = await client.GET("/crawls/{id}", {
        params: { path: { id } },
      });
      if (data) {
        return data;
      }
      if (response.status === 404) {
        return null;
      }
      throw new Error(
        (error as { message?: string } | undefined)?.message ??
          "Failed to load crawl",
      );
    },
    ...options,
  });
}
