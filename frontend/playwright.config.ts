import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "html",
  timeout: 60000,

  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],

  webServer: [
    {
      command: process.env.CI
        ? "./server"
        : "go run ./cmd/server/",
      cwd: "../backend",
      port: 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 120000,
      env: {
        SERVER_PORT: "8080",
        JWT_SECRET: "playwright-e2e-secret",
        ACCESS_TTL: "60",
        REFRESH_TTL: "7",
        DB_HOST: "localhost",
        DB_PORT: "5432",
        DB_USER: "postgres",
        DB_PASSWORD: "postgres",
        DB_NAME: "sync_test",
        DB_SSLMODE: "disable",
      },
    },
    {
      command: "npm run dev",
      cwd: ".",
      port: 3000,
      reuseExistingServer: !process.env.CI,
      timeout: 120000,
    },
  ],
});
