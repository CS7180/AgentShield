import { test, expect } from '@playwright/test';

/**
 * Authentication flow tests.
 *
 * These tests run against the live Vite dev server.  They do NOT require a
 * real Supabase project — the playwright config injects stub env vars so that
 * `isSupabaseConfigured` is true and `ProtectedRoute` behaves normally.
 *
 * No session is injected here, so the app acts as an unauthenticated visitor.
 */

test.describe('Authentication', () => {
  test('login page renders Sign in heading and Continue with Google button', async ({
    page,
  }) => {
    await page.goto('/login');

    // The LoginPage renders a <div> styled like a heading — use exact match to
    // avoid the strict-mode violation against the "Sign in with your Google
    // account…" subtitle that also contains the words "Sign in".
    await expect(page.getByText('Sign in', { exact: true })).toBeVisible();
    await expect(
      page.getByRole('button', { name: /continue with google/i }),
    ).toBeVisible();
  });

  test('unauthenticated visit to /dashboard redirects to /login', async ({
    page,
  }) => {
    await page.goto('/dashboard');

    // ProtectedRoute navigates to /login when session is null.
    await expect(page).toHaveURL(/\/login/);
    // The login page content should be visible at the redirect target.
    await expect(
      page.getByRole('button', { name: /continue with google/i }),
    ).toBeVisible();
  });

  test('unauthenticated visit to /scans redirects to /login', async ({
    page,
  }) => {
    await page.goto('/scans');

    await expect(page).toHaveURL(/\/login/);
  });
});
