# Sprint 1 Async Standups

Format: **Done / Doing / Blockers**

---

## Standup 1 — 2026-04-08

### Shanshou Li
**Done:** Set up `api-gateway` skeleton — Gin router, config loading, health endpoint. Added initial DB migration for `scans` table.
**Doing:** Implementing Supabase JWT validation middleware. Writing table-driven tests for token edge cases first (TDD red step committed).
**Blockers:** None.

### Yachen Wang
**Done:** Scaffolded `frontend/` with Vite + React Router. Set up `AuthProvider` and `ProtectedRoute`. Google OAuth login working locally.
**Doing:** Building `NewScanContent.jsx` — form validation, mode selection, attack type toggles.
**Blockers:** Need gateway `POST /api/v1/scans` to be available to test the full create flow. Will mock via `vi.fn()` in tests for now.

---

## Standup 2 — 2026-04-12

### Shanshou Li
**Done:** JWT middleware complete — `WithAudience("authenticated")` + `WithExpirationRequired()` enforced. SSRF validator blocks RFC1918/loopback/non-HTTPS. Scan CRUD endpoints (`POST`, `GET`, `GET/:id`, `start`, `stop`) with ownership middleware. Coverage on `internal/validation` at 89%, `internal/auth` still low.
**Doing:** WebSocket hub and Kafka consumer wiring. gRPC stub to orchestrator.
**Blockers:** Kafka requires Docker to be running — CI will need `docker compose up` step.

### Yachen Wang
**Done:** `NewScanContent.jsx` complete — HTTPS validation, auto-start flow, partial-failure handling. `DashboardContent.jsx` wired to `listScans` and `getScanReport`. All 12 API client functions implemented with real HTTP calls.
**Doing:** `ScanMonitorContent.jsx` — WebSocket feed, polling, report lifecycle, start/stop actions.
**Blockers:** `useNavigate` causing test failures — need to wrap renders in `MemoryRouter`. Will fix before pushing.

---

## Standup 3 — 2026-04-18

### Shanshou Li
**Done:** WebSocket hub complete and tested. Kafka consumer dispatches `agent.status` events to connected WS clients. gRPC orchestrator integration with `ORCHESTRATOR_ENABLED` flag. Replaced HS256 JWT secret with ES256 JWKS endpoint validation — stronger security posture. CI `ci.yml` added (Go build/test + frontend build).
**Doing:** Final PR review, coverage check, merge prep.
**Blockers:** Proto stubs need `make proto` before `ORCHESTRATOR_ENABLED=true` — documented in CLAUDE.md.

### Yachen Wang
**Done:** `ScanMonitorContent.jsx` complete — WebSocket feed, polling, auto-report generation, dead-letter display. `ReportCompareContent.jsx` and `JudgeContent.jsx` wired. Fixed `NewScanContent` test router context bug. Added 24 tests for `ScanMonitorContent` + `useAuth` tests — coverage now 95%.
**Doing:** Opening Sprint 1 PR, writing C.L.E.A.R. review notes.
**Blockers:** None — frontend Sprint 1 scope complete.
