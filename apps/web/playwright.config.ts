import { defineConfig, devices } from "@playwright/test";

// E2E config. The suite runs against an already-running stack (see `make up`),
// not a server Playwright starts itself — in CI/dev the app comes up via
// docker compose. Override the target with E2E_BASE_URL.
//
// NOTE: the `@playwright/test` version in package.json must match the
// `mcr.microsoft.com/playwright` image tag used by `make test-e2e`, or the
// baked-in browsers won't match the test runner. Keep them in sync.
const baseURL = process.env.E2E_BASE_URL ?? "http://localhost:3000";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "line" : "list",
  use: {
    baseURL,
    trace: "on-first-retry",
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
  ],
});
