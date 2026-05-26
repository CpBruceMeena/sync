import { test, expect } from "@playwright/test";

// Helper: generate unique test user to avoid conflicts
function testUser(prefix = "e2e") {
  const suffix = Date.now();
  return {
    username: `${prefix}_user_${suffix}`,
    email: `${prefix}_${suffix}@test.com`,
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

    // Should show error message (or redirect back to login page)
    // Check we're still on the login page
    await expect(page.locator("h1")).toContainText("Welcome Back", {
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

test.describe("Comprehensive E2E: Messaging & Groups", () => {
  test("should register two users, create group, and send messages", async ({
    page,
  }) => {
    // Register Alice (the group creator)
    const alice = testUser("alice");
    await page.goto("/register");
    await page.fill('input[type="text"]', alice.username);
    await page.fill('input[type="email"]', alice.email);
    let pwInputs = page.locator('input[type="password"]');
    await pwInputs.nth(0).fill(alice.password);
    await pwInputs.nth(1).fill(alice.password);
    await page.click('button[type="submit"]');
    await page.waitForURL("**/chat", { timeout: 15000 });

    // Verify Alice sees her username in the sidebar
    await expect(page.locator(`text=${alice.username}`)).toBeVisible({
      timeout: 5000,
    });

    // Alice registers Bob via the API (need another browser context for full E2E)
    // Since we only have one page, we'll use the API directly to create Bob
    const bob = testUser("bob");
    const registerBobRes = await fetch(
      "http://localhost:8080/api/auth/register",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username: bob.username,
          email: bob.email,
          password: bob.password,
        }),
      }
    );
    expect(registerBobRes.status).toBe(201);

    // Also register Charlie for group test
    const charlie = testUser("charlie");
    const registerCharlieRes = await fetch(
      "http://localhost:8080/api/auth/register",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username: charlie.username,
          email: charlie.email,
          password: charlie.password,
        }),
      }
    );
    expect(registerCharlieRes.status).toBe(201);

    // Now create a group conversation via the UI
    // Click the Create Group button
    await page.click('button[title="Create Group"]');

    // Wait for the Create Group dialog to appear
    await expect(page.locator("h2:has-text('Create Group')")).toBeVisible({
      timeout: 5000,
    });

    // Fill in group name
    await page.fill('input[placeholder="Group name"]', "Test Group E2E");

    // Search for Bob in the user search
    await page.fill(
      'input[placeholder="Search users..."]',
      bob.username.substring(0, 5)
    );

    // Wait for search results and click on Bob
    await page.click(`button:has-text('${bob.username}')`, { timeout: 5000 });

    // Also add Charlie - clear search first
    await page.fill('input[placeholder="Search users..."]', "");
    await page.fill(
      'input[placeholder="Search users..."]',
      charlie.username.substring(0, 5)
    );
    await page.click(`button:has-text('${charlie.username}')`, {
      timeout: 5000,
    });

    // Click Create Group button
    await page.click("button:has-text('Create Group')");

    // The dialog should close and the group should appear in the sidebar
    // Wait for the dialog to close
    await expect(page.locator("h2:has-text('Create Group')")).not.toBeVisible({
      timeout: 5000,
    });

    // The group should now appear in the conversation list
    await expect(page.locator("text=Test Group E2E")).toBeVisible({
      timeout: 5000,
    });

    // Click on the group to open it
    await page.click("text=Test Group E2E");

    // Wait for messages area to load (should show "Select a Conversation" disappears)
    // Send a message in the group
    const messageInput = page.locator('input[placeholder="Type a message..."], textarea[placeholder="Type a message..."]');
    if (await messageInput.isVisible()) {
      await messageInput.fill("Hello group! This is Alice.");
      await messageInput.press("Enter");
    }

    // Verify the message appears in the chat (may have optimistic + confirmed duplicates)
    const sentMessage = page.locator("text=Hello group! This is Alice.").first();
    await expect(sentMessage).toBeVisible({ timeout: 5000 });
  });

  test("should login with test credentials (alice) and verify conversation list", async ({
    page,
  }) => {
    // Login with existing alice test account
    await page.goto("/login");
    await page.fill('input[type="email"]', "alice@test.com");
    await page.fill('input[type="password"]', "password123");
    await page.click('button[type="submit"]');

    // Should redirect to /chat
    await page.waitForURL("**/chat", { timeout: 15000 });

    // Verify user info is visible
    await expect(page.locator("text=alice")).toBeVisible({ timeout: 5000 });

    // Verify the conversation sidebar is loaded (should show "Conversations" header)
    await expect(page.locator("text=Conversations")).toBeVisible({
      timeout: 5000,
    });

    // Verify the WebSocket connection status
    await expect(page.locator("text=Connected")).toBeVisible({ timeout: 5000 });

    // Logout
    await page.click('button[title="Logout"]');
    await page.waitForURL("**/login", { timeout: 10000 });
  });
});

// API-level E2E verification (not UI-dependent, tests backend contract)
test.describe("API Contract Tests", () => {
  test("health endpoint returns ok", async () => {
    const res = await fetch("http://localhost:8080/health");
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.status).toBe("ok");
  });

  test("register and login flow via API", async () => {
    const suffix = Date.now();
    const email = `api_test_${suffix}@test.com`;

    // Register
    const regRes = await fetch("http://localhost:8080/api/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username: `api_user_${suffix}`,
        email,
        password: "password123",
      }),
    });
    expect(regRes.status).toBe(201);
    const regBody = await regRes.json();
    expect(regBody.user).toBeDefined();
    expect(regBody.token).toBeDefined();
    const accessToken = regBody.token.access_token;

    // Get current user
    const meRes = await fetch("http://localhost:8080/api/auth/me", {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    expect(meRes.status).toBe(200);
    const meBody = await meRes.json();
    expect(meBody.email).toBe(email);

    // Refresh token
    const refreshRes = await fetch("http://localhost:8080/api/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: regBody.token.refresh_token }),
    });
    expect(refreshRes.status).toBe(200);

    // Logout
    const logoutRes = await fetch("http://localhost:8080/api/auth/logout", {
      method: "POST",
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    expect(logoutRes.status).toBe(200);

    // List users
    const usersRes = await fetch("http://localhost:8080/api/users", {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    expect(usersRes.status).toBe(200);
    const users = await usersRes.json();
    expect(Array.isArray(users)).toBe(true);
  });
});
