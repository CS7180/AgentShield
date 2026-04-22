# Sprint 2 Retrospective — 2026-05-04

**Attendees:** Shanshou Li, Yachen Wang
**Format:** async written retrospective

---

## What went well

- **Playwright E2E tests shipped and pass in CI.** All 7 tests (3 auth + 4 scan-creation) run fully offline using fake JWT injection and `page.route()` mocking. The CI `frontend-e2e` job installs Chromium, runs tests, and uploads an HTML report as an artifact. This was the biggest gap from Sprint 1 and it's fully closed.

- **CI pipeline is now production-grade.** The pipeline grew from 1 job (Go tests) to a full 5-job chain: `secrets-scan` (Gitleaks) → `api-gateway` + `backend-services` → `frontend-build` (lint, unit tests, npm audit, build) → `frontend-e2e`. Every meaningful quality gate runs on every PR.

- **Gitleaks secret scanning works without a license.** Swapped `gitleaks/gitleaks-action@v2` (requires paid org license) for the open-source CLI downloaded directly from GitHub releases. Same scanning coverage, no cost, no OIDC token issues.

- **CLAUDE.md @imports refactor improved maintainability.** Splitting the 392-line monolith into 5 modular files under `docs/claude/` made it easier for both partners to update conventions and architecture notes independently. Claude Code sessions load more relevant context per token budget.

- **Claude Code extensibility fully wired up.** All required features are now in place: `@imports` in CLAUDE.md, 3 skills (with v1→v2 iteration), 2 hooks (PostToolUse lint + Stop test), GitHub MCP server, `security-reviewer` sub-agent, PR template with C.L.E.A.R. built in, and issue templates with Definition of Done.

- **Cross-layer debugging improved.** The `||` vs `??` E2E CI failure (GitHub Actions passing `""` for unset secrets) was diagnosed and fixed quickly by reasoning across three layers simultaneously: GitHub Actions interpolation, Vite env var handling, and Supabase `isSupabaseConfigured` logic. Writing the fix as a memory file means we won't repeat it.

- **Go backend test coverage raised significantly.** Systematically adding table-driven tests to `internal/auth`, `internal/middleware`, `internal/handler`, and `internal/validation` brought coverage from 24% to well above the Sprint 2 target. The CI coverage gate was raised accordingly.

---

## What could be improved

- **Vercel deployment was removed rather than replaced.** We removed `vercel.json` and the CI deploy jobs after hitting org-level friction, but didn't immediately substitute an alternative deployment platform. A public URL was a rubric requirement and should have been unblocked earlier in the sprint rather than at the end.

- **C.L.E.A.R. wasn't applied systematically until late.** The framework was documented in Sprint 1 planning but never appeared in actual PR descriptions until we added the PR template in Sprint 2. The right fix was the template — but we should have added it at the start of Sprint 1, not the end of Sprint 2.

- **Blue team agents and adversarial mode remain unimplemented.** The Sprint 2 goal included all 3 scan modes. Red team only (the baseline mode) is the state we shipped. Blue team agents, adversarial orchestration, and the comparative report view were descoped under time pressure. These represent the platform's core differentiator and remain the most important unfinished work.

- **Sprint docs were sometimes written in batches rather than in real time.** Standup notes captured the right information but were occasionally written at the end of a work session rather than during. This made the cadence feel less live than intended.

---

## Action items for final push

| Action | Owner | Priority |
|--------|-------|----------|
| Deploy frontend to a public URL (GitHub Pages or Netlify) | Yachen | HIGH |
| Publish technical blog post | Yachen | HIGH |
| Record 5-10 min video demo | Both | HIGH |
| Submit individual reflections | Both | HIGH |
| Submit peer evaluations | Both | HIGH |
| Write Sprint 2 retro | Both | ✅ Done |
| Blue team agents skeleton (at minimum) | Shanshou | MEDIUM |
