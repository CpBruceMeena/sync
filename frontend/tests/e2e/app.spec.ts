import { test, expect } from "@playwright/test";

// Helper: generate unique test user to avoid conflicts
function testUser() {
  const suffix = Date.now();
  return {
    username: `e2e_user_${suffix}`,
    email: `e2e_${suffix}@test.com`,
    password: "e2etest123",
  };
}

test.describe("Application E2E Flow", () => {
  test("should register a new user", async ({ page }) => {
    const user = testUser();

    await page.goto("/register");
    await expect(page.locator("h1")).toContainText("Create Account");

    // Fill in registration form
    await page.fill('input[type="text"]', user.username);
    await page.fill('input[type="email"]', user.email);
    await page.fill('[placeholder="••••••••"]', user.password);

    // Fill confirm password (second password field)
    const passwordInputs = page.locator('input[type="password"]');
    await passwordInputs.nth(1).fill(user.password);

    // Submit
    await page.click('button[type="submit"]');

    // Should redirect to /chat after successful registration
    await page.waitForURL("**/chat", { timeout: 15000 });
    await expect(page.locator("text=sync")).toBeVisible({ timeout: 5000 });
  });

  test("should login with existing user", async ({ page }) => {
    const user = testUser();

    // First register a user
    await page.goto("/register");
    await page.fill('input[type="text"]', user.username);
    await page.fill('input[type="email"]', user.email);
    const passwordInputs = page.locator('input[type="password"]');
    await passwordInputs.nth(0).fill(user.password);
    await passwordInputs.nth(1).fill(user.password);
    await page.click('button[type="submit"]');
    await page.waitForURL("**/chat", { timeout: 15000 });

    // Logout by clicking the logout button in the sidebar
    await page.click('button[title="Logout"]');
    await page.waitForURL("**/login", { timeout: 10000 });

    // Now login
    await page.fill('input[type="email"]', user.email);
    await page.fill('input[type="password"]', user.password);
    await page.click('button[type="submit"]');

    // Should redirect to /chat
    await page.waitForURL("**/chat", { timeout: 15000 });
    await expect(page.locator("text=sync")).toBeVisible({ timeout: 5000 });
  });

  test("should show user info after login", async ({ page }) => {
    const user = testUser();

    // Register a user
    await page.goto("/register");
    await page.fill('input[type="text"]', user.username);
    await page.fill('input[type="email"]', user.email);
    const passwordInputs = page.locator('input[type="password"]');
    await passwordInputs.nth(0).fill(user.password);
    await passwordInputs.nth(1).fill(user.password);
    await page.click('button[type="submit"]');
    await page.waitForURL("**/chat", { timeout: 15000 });

    // Check user info is visible in sidebar
    await expect(page.locator(`text=${user.username}`)).toBeVisible({
      timeout: 5000,
    });
  });

  test("should show validation errors on register", async ({ page }) => {
    await page.goto("/register");

    // Submit empty form
    await page.click('button[type="submit"]');

    // Browser validation should prevent submission of empty required fields
    // Check we're still on the register page
    await expect(page.locator("h1")).toContainText("Create Account");

    // Submit with short password
    await page.fill('input[type="text"]', "test");
    await page.fill('input[type="email"]', "test@test.com");
    const passwordInputs = page.locator('input[type="password"]');
    await passwordInputs.nth(0).fill("123");
    await passwordInputs.nth(1).fill("123");
    await page.click('button[type="submit"]');

    // Should show error about short password
    await expect(page.locator("text=at least 6 characters")).toBeVisible({
      timeout: 5000,
    });
  });

  test("should show login validation errors", async ({ page }) => {
    await page.goto("/login");

    // Submit with invalid credentials
    await page.fill('input[type="email"]', "nonexistent@test.com");
    await page.fill('input[type="password"]', "wrongpassword");
    await page.click('button[type="submit"]');

    // Should show error message
    await expect(page.locator("text=Login failed")).toBeVisible({
      timeout: 10000,
    });
  });

  test("should navigate between login and register", async ({ page }) => {
    await page.goto("/login");

    // Click "Create one" link to go to register
    await page.click("text=Create one");
    await page.waitForURL("**/register");
    await expect(page.locator("h1")).toContainText("Create Account");

    // Click "Sign in" link to go back to login
    await page.click("text=Sign in");
    await page.waitForURL("**/login");
    await expect(page.locator("h1")).toContainText("Welcome Back");
  });
});
