# Sprint 1 Planning — Weeks 1–2

**Sprint Goal:** Red team mode working end-to-end. User can create a scan, agents execute attacks, Judge evaluates results, and the dashboard displays the report.

**Dates:** 2026-04-07 to 2026-04-20

---

## Acceptance Criteria (Definition of Done)

A story is done when:
- Implementation code exists
- At least one failing test was committed before implementation (TDD red step visible in git)
- Tests pass with no regressions
- `gofmt -l .` outputs nothing (Go) / `npm run lint` exits 0 (frontend)
- No secrets committed; `.env.example` updated if new env vars added
- PR reviewed with C.L.E.A.R. framework before merge

---

## Stories

### [Gateway] Supabase JWT validation — Shanshou
**Acceptance criteria:**
- `POST /api/v1/scans` returns 401 for missing/expired/invalid JWT
- `POST /api/v1/scans` returns 201 for valid Supabase-issued JWT
- JWT validated using JWKS endpoint (not shared secret)
- Table-driven tests cover: missing header, malformed token, expired token, wrong audience, valid token
- Coverage on `internal/auth` ≥ 80%

### [Gateway] Scan CRUD endpoints — Shanshou
**Acceptance criteria:**
- `POST /api/v1/scans` creates a scan and returns 201 with scan ID
- `GET /api/v1/scans` lists only the authenticated user's scans
- `GET /api/v1/scans/:id` returns 404 for another user's scan (ownership enforced)
- `POST /api/v1/scans/:id/start` transitions status to `running` (or `queued` if orchestrator disabled)
- `POST /api/v1/scans/:id/stop` transitions status to `stopped`
- SSRF validator blocks RFC1918, loopback, non-HTTPS target endpoints

### [Gateway] WebSocket status feed — Shanshou
**Acceptance criteria:**
- `WS /ws/scans/:id/status?token=<jwt>` accepts valid tokens
- Rejects missing/invalid tokens with 401 before upgrade
- Ownership enforced: cannot subscribe to another user's scan
- Messages from Kafka `agent.status` topic forwarded to connected clients

### [Frontend] Auth + protected routes — Yachen
**Acceptance criteria:**
- Unauthenticated users redirected to `/login`
- Google OAuth login works end-to-end via Supabase
- Session persists across page reloads
- `useAuth` hook throws if used outside `AuthProvider`
- Coverage on `src/auth/` ≥ 70%

### [Frontend] Scan creation page — Yachen
**Acceptance criteria:**
- Form validates HTTPS URL before submitting
- Mode selection (Red Team / Blue Team / Adversarial) works
- At least one attack type must be selected
- Auto-start checkbox creates and immediately starts the scan
- Partial failure (scan created, start failed) shows error with scan ID
- 3 unit tests covering: invalid URL, success with auto-start, success without auto-start

### [Frontend] Dashboard + monitoring pages — Yachen
**Acceptance criteria:**
- Dashboard shows total scans, critical findings, avg defence score, Judge tau
- Scan monitor shows live agent status, attack results feed, report lifecycle
- WebSocket feed appends events in real time
- Start/Stop buttons call correct API endpoints
- Coverage on all page components ≥ 70%

---

## Capacity

| Partner | Available days | Focus |
|---------|---------------|-------|
| Shanshou Li | 8 | api-gateway, orchestrator stub, CI |
| Yachen Wang | 8 | frontend (all pages), frontend tests |
