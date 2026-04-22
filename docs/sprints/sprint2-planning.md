# Sprint 2 Planning — Weeks 3–4

**Sprint Goal:** Full platform with all 3 scan modes. Deployed on Vercel. CI/CD pipeline complete with security gates. OWASP reports with comparison view. 80%+ test coverage.

**Dates:** 2026-04-21 to 2026-05-04

---

## Acceptance Criteria (Definition of Done)

Same as Sprint 1, plus:
- Feature deployed to Vercel preview URL before merge
- CI pipeline passes all stages (lint, typecheck, tests, E2E, security, AI review)
- No new `npm audit` HIGH or CRITICAL vulnerabilities introduced

---

## Stories

### [Frontend] Vercel deployment — Yachen
**Acceptance criteria:**
- `vercel.json` present with SPA fallback rewrite rule
- `VITE_API_BASE_URL` and `VITE_SUPABASE_*` set as Vercel env vars
- Production URL accessible and login works end-to-end
- Preview deploys auto-created on every PR

### [Frontend] Playwright E2E tests — Yachen
**Acceptance criteria:**
- `playwright.config.ts` committed
- Test 1: login flow — unauthenticated user redirected to login, Google OAuth completes, dashboard loads
- Test 2: scan creation — authenticated user fills form, submits, sees scan ID in monitor
- Both tests run in CI (`npx playwright test --reporter=github`)

### [CI/CD] Pipeline hardening — Shanshou
**Acceptance criteria:**
- `npm audit --audit-level=high` stage in CI; fails build on HIGH/CRITICAL
- Gitleaks action scans every push for committed secrets
- `claude-code-action` posts AI review comment on every PR
- Vercel preview deploy URL posted as PR comment
- Production deploy triggers on merge to main

### [Gateway] Go test coverage push — Shanshou
**Acceptance criteria:**
- `internal/auth` ≥ 80% coverage
- `internal/middleware` ≥ 80% coverage
- `internal/handler` ≥ 80% coverage
- `internal/validation` ≥ 80% coverage
- CI coverage gate raised from 25% to 70%

### [Frontend] Report comparison view — Yachen
**Acceptance criteria:**
- `ReportCompareContent.jsx` wired to `compareScans` API
- Shows score delta, severity delta table, trend indicator
- Handles case where fewer than 2 completed scans exist (warning state)

### [Frontend] OWASP scorecard in report — Yachen
**Acceptance criteria:**
- Report view shows per-category pass/fail/partial for OWASP LLM Top 10
- Severity distribution chart using Recharts
- PDF download link shown when `report_pdf_path` is available

### [CLAUDE.md] @imports refactor — Yachen
**Acceptance criteria:**
- CLAUDE.md split into modular files under `docs/claude/`
- CLAUDE.md uses `@docs/claude/architecture.md` etc. syntax
- All content preserved, git history shows evolution

---

## Capacity

| Partner | Available days | Focus |
|---------|---------------|-------|
| Shanshou Li | 8 | CI/CD hardening, Go coverage, Blue team agents, adversarial mode |
| Yachen Wang | 8 | Vercel deploy, Playwright, report views, CLAUDE.md refactor |
