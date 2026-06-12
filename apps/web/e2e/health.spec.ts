import { test, expect } from "@playwright/test";

// Minimal end-to-end smoke test for spec 001 (Foundation & health):
// the web app boots, renders the shell, and surfaces the API's readiness
// status fetched from the backend through the generated client.
//
// Unlike the Vitest component test (which mocks the client), this exercises
// the real transport: browser → Next.js → /readyz on the API.

test("home page renders the app shell", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Simple Web Crawler" })).toBeVisible();
});

test("home page surfaces the live API status", async ({ page }) => {
  await page.goto("/");

  // The status starts as "Checking…" then resolves to ok/degraded once the
  // readiness query settles. Either resolved state proves the round-trip ran.
  const status = page.getByText(/^API: (ok|degraded)$/);
  await expect(status).toBeVisible({ timeout: 15_000 });
});
