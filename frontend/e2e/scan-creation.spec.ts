import { test, expect, type Page } from '@playwright/test';
import { Buffer } from 'node:buffer';
import { readFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';

// ─── Env helpers ──────────────────────────────────────────────────────────────

/**
 * Resolve the actual VITE_SUPABASE_URL in priority order:
 *  1. process.env (GitHub Actions secrets / any external injection)
 *  2. .env.local  (local developer machine)
 *  3. Placeholder fallback (CI without secrets — stubs still make auth work)
 *
 * Playwright worker processes do NOT inherit env mutations made in
 * playwright.config.ts, so we read .env.local directly here.
 */
function resolveSupabaseUrl(): string {
  if (process.env.VITE_SUPABASE_URL) return process.env.VITE_SUPABASE_URL;
  const envPath = resolve(process.cwd(), '.env.local');
  if (existsSync(envPath)) {
    const m = readFileSync(envPath, 'utf8').match(/^VITE_SUPABASE_URL=(.+)$/m);
    if (m) return m[1].trim();
  }
  return 'https://placeholder.supabase.co';
}

// ─── JWT / session helpers ────────────────────────────────────────────────────

/**
 * Build a syntactically valid (but unsigned) JWT whose payload is a proper
 * base64-encoded JSON blob.  Supabase's client-side JWT decoder calls
 * `atob(parts[1])` — using standard base64 (not base64url) for the payload
 * means the `replace(/-/g, '+')` conversion is a no-op, and atob succeeds.
 */
function makeFakeJwt(userId: string, email: string): string {
  const b64 = (o: object) => Buffer.from(JSON.stringify(o)).toString('base64');
  const exp = Math.floor(Date.now() / 1000) + 3600;
  const header = b64({ alg: 'HS256', typ: 'JWT' });
  const payload = b64({
    sub: userId,
    email,
    role: 'authenticated',
    aud: 'authenticated',
    exp,
    iat: Math.floor(Date.now() / 1000),
    iss: 'supabase',
  });
  return `${header}.${payload}.FAKESIG`;
}

// ─── Test constants ───────────────────────────────────────────────────────────

const SUPABASE_URL = resolveSupabaseUrl();
/**
 * Supabase v2 stores the session under `sb-<ref>-auth-token` where `ref` is
 * the first subdomain segment of the Supabase project URL — NOT the full
 * hostname.  e.g. `https://abc123.supabase.co` → `sb-abc123-auth-token`.
 * We inject this key into localStorage before each page load so that
 * `supabase.auth.getSession()` finds a live session without any network call.
 */
const STORAGE_KEY = `sb-${new URL(SUPABASE_URL).hostname.split('.')[0]}-auth-token`;

function resolveApiBase(): string {
  if (process.env.VITE_API_URL) return process.env.VITE_API_URL;
  const envPath = resolve(process.cwd(), '.env.local');
  if (existsSync(envPath)) {
    const m = readFileSync(envPath, 'utf8').match(/^VITE_API_URL=(.+)$/m);
    if (m) return m[1].trim();
  }
  return 'http://localhost:8080';
}

const API_BASE = resolveApiBase();

const MOCK_USER = {
  id: 'user-e2e-test-123',
  aud: 'authenticated',
  role: 'authenticated',
  email: 'e2e-test@example.com',
  email_confirmed_at: '2024-01-01T00:00:00Z',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  app_metadata: { provider: 'google', providers: ['google'] },
  user_metadata: { email: 'e2e-test@example.com', full_name: 'E2E Test User' },
};

function buildMockSession() {
  return {
    access_token: makeFakeJwt(MOCK_USER.id, MOCK_USER.email),
    refresh_token: 'mock-refresh-token-e2e',
    expires_in: 3600,
    expires_at: Math.floor(Date.now() / 1000) + 3600,
    token_type: 'bearer',
    user: MOCK_USER,
  };
}

const CREATED_SCAN = {
  id: 'scan-e2e-abc123',
  target_endpoint: 'https://api.example.com/v1/chat',
  mode: 'red_team',
  attack_types: ['prompt_injection', 'jailbreak', 'data_leakage'],
  status: 'pending',
  created_at: new Date().toISOString(),
};

// ─── Per-test setup ───────────────────────────────────────────────────────────

/**
 * Inject a fake Supabase session into localStorage before the page loads.
 * This makes AuthProvider.getSession() return a valid session immediately
 * (reads from localStorage — no network call) so ProtectedRoute lets us in.
 */
async function injectMockAuth(page: Page) {
  const session = buildMockSession();
  await page.addInitScript(
    ({ key, sess }) => {
      localStorage.setItem(key, JSON.stringify(sess));
    },
    { key: STORAGE_KEY, sess: session },
  );
}

/**
 * Route all backend API calls and Supabase auth calls to mock responses so
 * tests run fully offline.
 */
async function mockApiRoutes(page: Page) {
  const supabaseOrigin = new URL(SUPABASE_URL).origin;

  // Suppress any background Supabase auth network calls.
  await page.route(`${supabaseOrigin}/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ user: MOCK_USER }),
    }),
  );

  // Mock all gateway API calls.
  await page.route(`${API_BASE}/api/v1/**`, async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    // POST /api/v1/scans  →  201 Created
    if (url === `${API_BASE}/api/v1/scans` && method === 'POST') {
      return route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(CREATED_SCAN),
      });
    }

    // POST /api/v1/scans/:id/start  →  200 running
    if (url.includes('/api/v1/scans/') && url.endsWith('/start') && method === 'POST') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          ...CREATED_SCAN,
          status: 'running',
          message: 'Scan started.',
        }),
      });
    }

    return route.continue();
  });
}

// ─── Tests ────────────────────────────────────────────────────────────────────

test.describe('Scan creation', () => {
  test.beforeEach(async ({ page }) => {
    await injectMockAuth(page);
    await mockApiRoutes(page);
  });

  test('Create Scan page renders target URL input and submit button', async ({
    page,
  }) => {
    await page.goto('/scans');

    await expect(page.getByText('Create Scan')).toBeVisible();
    await expect(
      page.getByPlaceholder(/https:\/\/example\.com/i),
    ).toBeVisible();
    await expect(
      page.getByRole('button', { name: /create.*scan/i }),
    ).toBeVisible();
  });

  test('non-HTTPS URL shows validation error on Validate URL click', async ({
    page,
  }) => {
    await page.goto('/scans');

    await page.getByPlaceholder(/https:\/\/example\.com/i).fill('http://insecure.example.com');
    await page.getByRole('button', { name: /validate url/i }).click();

    await expect(page.getByText(/must be a valid https url/i)).toBeVisible();
  });

  test('valid HTTPS URL passes Validate URL check', async ({ page }) => {
    await page.goto('/scans');

    await page
      .getByPlaceholder(/https:\/\/example\.com/i)
      .fill('https://api.example.com/v1/chat');
    await page.getByRole('button', { name: /validate url/i }).click();

    await expect(page.getByText(/looks valid/i)).toBeVisible();
  });

  test('submitting with a valid URL creates a scan and shows result panel', async ({
    page,
  }) => {
    await page.goto('/scans');

    await page
      .getByPlaceholder(/https:\/\/example\.com/i)
      .fill('https://api.example.com/v1/chat');

    // autoStart is on by default — button label is "Create + Start Scan".
    await page.getByRole('button', { name: /create.*scan/i }).click();

    // After the API call resolves, the "Latest Scan" panel appears with the
    // scan ID returned by the mock.
    await expect(page.getByText(CREATED_SCAN.id)).toBeVisible({
      timeout: 10_000,
    });
    await expect(page.getByText(CREATED_SCAN.target_endpoint)).toBeVisible();
  });
});
