import { defineConfig, devices } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

/**
 * Load .env.local into process.env so that tests can derive the correct
 * Supabase storage key.  Vite 8 gives .env.local precedence over
 * webServer.env, so the running app will always use the .env.local values
 * in local development.  We need the test runner (Node.js side) to agree
 * on the same VITE_SUPABASE_URL so the localStorage injection uses the
 * right key (`sb-<ref>-auth-token`).
 *
 * In CI, .env.local does not exist — values come from GitHub Actions secrets
 * that are already in process.env.
 */
function loadDotEnvLocal() {
  try {
    const raw = readFileSync(resolve(__dirname, '.env.local'), 'utf8');
    for (const line of raw.split('\n')) {
      const m = line.match(/^([A-Z_][A-Z0-9_]*)=(.*)$/);
      if (m) {
        const [, key, val] = m;
        // Don't overwrite vars that the CI environment already provided.
        if (!process.env[key]) process.env[key] = val.trim();
      }
    }
  } catch {
    // .env.local absent (CI) — fall through; GHA secrets set process.env directly.
  }
}

loadDotEnvLocal();

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI
    ? [['github'], ['html', { open: 'never' }]]
    : [['html', { open: 'never' }]],
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    // Vite 8 gives .env.local file values precedence over process.env, so in
    // local dev these stubs only take effect when .env.local is absent.
    // In CI (no .env.local) they ensure isSupabaseConfigured === true.
    env: {
      VITE_SUPABASE_URL:
        process.env.VITE_SUPABASE_URL ?? 'https://placeholder.supabase.co',
      VITE_SUPABASE_ANON_KEY:
        process.env.VITE_SUPABASE_ANON_KEY ?? 'placeholder-anon-key',
      VITE_API_URL: process.env.VITE_API_URL ?? 'http://localhost:8080',
    },
  },
});
