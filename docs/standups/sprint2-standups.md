# Sprint 2 Async Standups

Format: **Done / Doing / Blockers**

---

## Standup 1 — 2026-04-22

### Shanshou Li
**Done:** Sprint 2 planning doc committed. Reviewed Sprint 1 retro action items.
**Doing:** Setting up GitHub Actions stages for `npm audit` and Gitleaks. Adding `claude-code-action` for AI PR reviews.
**Blockers:** Need `ANTHROPIC_API_KEY` added as a GitHub Actions secret for the AI review stage.

### Yachen Wang
**Done:** `vercel.json` committed with SPA rewrite rule. Frontend deployed to Vercel — login flow working on production URL.
**Doing:** Playwright setup (`playwright.config.ts`, login E2E test).
**Blockers:** Supabase OAuth redirect URL needs to include the Vercel production domain — updating allowed redirect URLs in Supabase dashboard.

---

## Standup 2 — 2026-04-26

### Shanshou Li
**Done:** `npm audit` + Gitleaks stages live in CI. `claude-code-action` posts review comments on PRs. Coverage gate raised to 50% (intermediate step toward 70%). Added tests for `internal/middleware` — now at 78%.
**Doing:** `internal/handler` coverage push. Blue Team agent stubs.
**Blockers:** None.

### Yachen Wang
**Done:** Playwright login test passing (`auth.spec.ts`). Scan creation E2E test drafted. Vercel preview deploys working — URL posted as PR comment.
**Doing:** Scan creation E2E test (waiting for Supabase test account setup). OWASP scorecard UI in report view.
**Blockers:** Need a persistent test user in Supabase for E2E — will use `PLAYWRIGHT_TEST_EMAIL` / `PLAYWRIGHT_TEST_PASSWORD` env vars.

---

## Standup 3 — 2026-04-30

### Shanshou Li
**Done:** `internal/auth`, `internal/handler`, `internal/ws` all at 80%+. CI coverage gate at 70%. Blue team Input Guard + Output Filter agent stubs in place.
**Doing:** Adversarial mode orchestration. Production deploy on merge to main.
**Blockers:** None.

### Yachen Wang
**Done:** Both Playwright E2E tests passing in CI. OWASP scorecard component complete. PDF download link shown when artifact available. Report comparison view polished.
**Doing:** Final polish pass — empty states, error toasts, responsive layout check.
**Blockers:** None.
