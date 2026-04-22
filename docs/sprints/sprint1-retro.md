# Sprint 1 Retrospective — 2026-04-20

**Attendees:** Shanshou Li, Yachen Wang
**Format:** async written retrospective

---

## What went well

- **JWT JWKS migration landed cleanly.** Switching from HS256 shared-secret to ES256 JWKS public-key validation was a significant security improvement and went in without breaking existing tests. The table-driven test structure made it easy to cover all the token edge cases.

- **Frontend coverage reached 95%.** After fixing the missing `MemoryRouter` context in `NewScanContent.test.jsx` and adding 24 tests for `ScanMonitorContent`, coverage jumped from ~50% to 95%. The new `ScanMonitorContent` tests cover WebSocket, polling, report lifecycle, and all action handlers.

- **Scan creation flow is solid.** The local flag fix for the partial-start failure case (scan created but start failed) was caught by TDD — we wrote the failing test first and the fix was minimal. The success message no longer shows when auto-start fails.

- **Claude Code productivity was high.** The `add-feature` skill's Tier 1/2/3 security checklist caught two potential SSRF gaps during implementation. The PostToolUse lint hook would have caught the one formatting issue that slipped through in the ws_handler change.

---

## What could be improved

- **Go test coverage is still low (24%).** Most packages in `api-gateway/internal/` have 0% coverage. We set a 25% baseline gate in CI but the real target is 80%+ on auth, middleware, validation, handler, and ws. This is the biggest technical debt entering Sprint 2.

- **No E2E tests yet.** Playwright was planned for Sprint 1 but deprioritised. This affects both the Testing rubric and CI pipeline completeness. Must be done in Sprint 2 Week 1.

- **CI pipeline is incomplete.** Currently only runs Go tests + frontend build. Missing: `npm audit`, Gitleaks, Playwright E2E stage, AI PR review, and Vercel preview deploy. Each of these is a rubric item.

- **Sprint docs written retroactively.** Process docs should be written at sprint start, not after. For Sprint 2, standup notes will be written in real time.

---

## Action items for Sprint 2

| Action | Owner | Due |
|--------|-------|-----|
| Add Playwright + 2 E2E tests | Yachen | Week 3 Day 1 |
| Expand Go coverage to 80%+ on core packages | Shanshou | Week 3 |
| Add npm audit + Gitleaks + claude-code-action to CI | Shanshou | Week 3 Day 2 |
| Deploy frontend to Vercel | Yachen | Week 3 Day 1 |
| Write Sprint 2 planning doc before starting | Both | Week 3 Day 1 |
