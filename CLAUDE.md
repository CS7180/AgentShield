# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## AgentShield Engineering Guide

This file is the authoritative operating guide for Claude Code working in this repository.
Full product requirements: @project_memory/AgentShield_PRD.md
React UI mockups (dark-themed, purple accent):
- `project_memory/AgentShield_Dashboard_Final.jsx`
- `project_memory/AgentShield_NewScan.jsx`
- `project_memory/AgentShield_ScanMonitor.jsx`
- `project_memory/AgentShield_Report.jsx`

---

## 1. Project Overview

**AgentShield** is a multi-agent red-blue teaming platform that automatically discovers security vulnerabilities in LLM-powered applications. It simulates adversarial attacks and defenses running simultaneously against a target LLM to identify vulnerabilities before deployment.

**Three scan modes:**
- **Red Team** — attacks only, no defense, baseline measurement
- **Blue Team** — measures false-positive rate on normal traffic
- **Adversarial** — red + blue simultaneous, comparative analysis

---

## 2. Tech Stack

| Layer | Technology |
|-------|-----------|
| API Gateway | Go 1.22, Gin, pgx/v5, go-redis/v9, Sarama, gorilla/websocket |
| Orchestrator | Go 1.22, gRPC server |
| Agent frameworks | Python 3.11+, CrewAI |
| Inter-service | gRPC + Protobuf |
| Message queue | Kafka (topics: `attack.results`, `defense.results`, `judge.evaluations`, `agent.status`) |
| Auth | Supabase Auth (frontend → Supabase directly; gateway validates Supabase-issued JWT) |
| Database | Supabase PostgreSQL (pgx/v5, Transaction Pooler port 6543) |
| Cache | Redis 7 |
| LLM models | Gemini 2.5 Flash (generation), Gemini 2.5 Pro (Judge) |
| Frontend | React 18, TypeScript, Vite, Recharts, TailwindCSS |
| Containers | Docker, Kubernetes, Helm |
| CI/CD | Jenkins + GitLab CI |
| Infrastructure | Terraform + AWS EKS |
| Monitoring | Prometheus + Grafana |

---

## 3. System Architecture

```
React Dashboard  ──── Supabase Auth (login/register/refresh — frontend only)
       │
       │  JWT (Authorization: Bearer)          WS ?token=<jwt>
       ▼
Go API Gateway  (JWT validation, rate limiting, SSRF guard, Prometheus)
       │
       │  gRPC
       ▼
Go Orchestrator  (task scheduling, agent lifecycle, result aggregation)
       │  gRPC task distribution
  ┌────┴───────────────────────────┐
  ▼                                ▼
Red Team Agents (Python/CrewAI)   Blue Team Agents (Python/CrewAI)
  Prompt Injection                  Input Guard
  Jailbreak                         Output Filter
  Data Leakage                      Behavior Monitor
  Constraint Drift                  Constraint Persistence
  └──────────────┬─────────────────┘
                 │  Kafka
                 ▼
         LLM-as-Judge (Python)
                 │
         Result Aggregator (Go)
                 │
  Report Generator (JSON + PDF, OWASP LLM Top 10)
                 │
          PostgreSQL + Redis
```

**Auth flow:** Frontend authenticates directly with Supabase SDK. The gateway has **no auth endpoints**. Every API call carries `Authorization: Bearer <supabase_jwt>`. The gateway validates the HS256 JWT using `SUPABASE_JWT_SECRET`. WebSocket upgrades use `?token=<jwt>` because browsers cannot set headers on WS connections.

---

## 4. Repository Structure & Ownership

```
AgentShield/
├── api-gateway/          # Go — Shanshou Li
│   ├── cmd/server/       # Binary entry point
│   ├── internal/         # All internal packages (not importable externally)
│   ├── proto/            # .proto source + generated stubs
│   └── migrations/       # SQL migrations (applied on startup)
├── orchestrator/         # Go — Shanshou Li  [Sprint 1, not yet created]
├── agents/               # Python/CrewAI — Shanshou Li  [Sprint 1, not yet created]
│   ├── red_team/
│   └── blue_team/
├── judge/                # Python — Shanshou Li  [Sprint 1, not yet created]
├── frontend/             # React/TS — Yachen Wang  [Sprint 1, not yet created]
├── infra/                # Terraform, Helm — Shanshou Li  [Sprint 2]
├── project_memory/       # PRD, UI mockups, reference docs (read-only)
└── CLAUDE.md
```

**Ownership rules:**
- Do not modify `project_memory/` files — they are reference artifacts.
- Do not create files outside the service directory being worked on unless explicitly asked.
- The `api-gateway/internal/` boundary is strict: no package inside it may be imported by another service.

---

## 5. Coding Conventions

### Go (`api-gateway/`, `orchestrator/`)
- `gofmt` + `goimports` always. No unformatted code.
- Error handling: wrap with `fmt.Errorf("context: %w", err)` — never swallow errors silently.
- All exported functions need a doc comment only if they are part of a public interface or handler. Internal helpers do not need comments unless logic is non-obvious.
- Interfaces defined at the point of use (consumer), not at the point of implementation.
- No `init()` functions except in `metrics/prometheus.go` (registering collectors).
- Context propagation: every function that does I/O takes `ctx context.Context` as its first argument.
- No naked `panic()` except in `main.go` during startup before the server is serving.
- Struct field order: public fields first, then private, then `protoimpl` fields last in proto stubs.
- Gin handlers return early on error via `c.AbortWithStatusJSON` — no redundant `return` after abort.

### Python (`agents/`, `judge/`)
- Python 3.11+. Type hints on all function signatures.
- Use `ruff` for linting, `black` for formatting.
- CrewAI agent definitions: one agent per file under `agents/red_team/` or `agents/blue_team/`.
- All LLM calls go through a single `llm_client.py` wrapper — never instantiate Gemini clients inline.
- Never log raw prompts or LLM responses at INFO level in production — use DEBUG.

### React / TypeScript (`frontend/`)
- TypeScript strict mode. No `any` types.
- Use `eslint` + `prettier`. Run `npm run lint` before committing.
- All API calls go through a single `src/api/client.ts` module. Never call `fetch` directly in components.
- Supabase Auth: use `@supabase/supabase-js` client. JWT is extracted from the Supabase session and passed as `Authorization: Bearer` to the gateway. Never store raw JWTs in `localStorage` directly — use the Supabase session object.
- Dark theme with purple accent (`#7C3AED`). Follow the design in `project_memory/AgentShield_*.jsx`.
- WebSocket connection: pass token as `?token=<jwt>` query param.

---

## 6. Testing Strategy (Strict TDD)

Follow red-green-refactor for all new logic:
1. **Red** — write a failing test that specifies the expected behavior.
2. **Green** — write the minimum code to make it pass.
3. **Refactor** — clean up without breaking tests.

Never write implementation code before a test exists for that behavior.

### Go tests
```bash
cd api-gateway
go test -race -cover ./...                   # all tests
go test -race -cover ./internal/auth/...     # single package
go test -v -run TestParseSupabaseToken ./internal/auth/...  # single test
```
- Target: **80%+ coverage** on `auth`, `middleware`, `validation`, `handler`, `ws` packages.
- Use table-driven tests (`[]struct{ name, input, want }`) for validators and JWT cases.
- No mocking of PostgreSQL — use test helpers that skip DB tests when `DATABASE_URL` is unset.
- WebSocket hub tests must not use `time.Sleep` longer than 50ms.

### Python tests
```bash
cd agents   # or judge/
pytest -v --cov=. --cov-report=term-missing
```
- Golden set: 50+ labeled `(attack_prompt, target_response, expected_success)` tuples in `judge/golden_set/`.
- Judge calibration tests must run against the real Gemini API in CI (not mocked) — use `pytest -m integration` marker.

### React tests
```bash
cd frontend
npm test -- --coverage
```
- Unit test all custom hooks and utility functions.
- Integration-test the scan creation flow with MSW (Mock Service Worker) — never mock `fetch` directly.

---

## 7. REST API Reference

```
# No auth endpoints — handled entirely by Supabase frontend SDK

GET  /health
GET  /metrics                              (Prometheus)

POST /api/v1/scans                         (JWTAuth + ScanCreateRateLimit)
GET  /api/v1/scans                         (JWTAuth)
GET  /api/v1/scans/:id                     (JWTAuth + Ownership)
POST /api/v1/scans/:id/start               (JWTAuth + Ownership)
POST /api/v1/scans/:id/stop                (JWTAuth + Ownership)
GET  /api/v1/scans/:id/report              (JWTAuth + Ownership)
GET  /api/v1/scans/:id/report/pdf          (JWTAuth + Ownership) → 501 until Sprint 2
GET  /api/v1/scans/:id/compare/:other_id   (JWTAuth + Ownership) → 501 until Sprint 2
POST /api/v1/judge/calibrate               (JWTAuth) → 501 until Sprint 2

WS   /ws/scans/:id/status                  (?token=<supabase_jwt>)
```

**Error envelope** (all errors):
```json
{ "error": "...", "code": "SNAKE_CASE", "status_code": 400, "timestamp": "...", "request_id": "..." }
```

---

## 8. Database Schema

Tables: `scans`, `api_keys` (in Supabase PostgreSQL, `auth.users` provided by Supabase)

Key `attack_results` fields (Sprint 2): `attack_type`, `attack_prompt`, `target_response`, `attack_success`, `severity` (critical/high/medium/low), `owasp_category`, `defense_intercepted`, `judge_confidence`, `latency_ms`, `tokens_used`

Migrations live in `api-gateway/migrations/` and are applied automatically on gateway startup. Use `IF NOT EXISTS` and idempotent guards. Never use `DROP` or destructive DDL in migrations.

**RLS:** Every table that stores user data must have `ENABLE ROW LEVEL SECURITY` and a `user_id = auth.uid()` policy. The gateway also enforces ownership in the `Ownership` middleware as a defense-in-depth layer.

---

## 9. gRPC / Protobuf Contract

```protobuf
// proto/orchestrator/orchestrator.proto
service OrchestratorService {
  rpc StartScan  (StartScanRequest)  returns (StartScanResponse);
  rpc StopScan   (StopScanRequest)   returns (StopScanResponse);
  rpc ScanStatus (ScanStatusRequest) returns (ScanStatusResponse);
}
```

**Proto rules:**
- Never change field numbers on existing messages — add new fields only.
- Always regenerate Go stubs with `make proto` after editing `.proto` files. The manual stubs in `proto/orchestrator/*.pb.go` are temporary scaffolding and must be replaced before enabling `ORCHESTRATOR_ENABLED=true`.
- Python agents use `grpcio-tools` to generate stubs: `python -m grpc_tools.protoc ...`
- The `.proto` source is the single source of truth. Generated files are derived artifacts.

---

## 10. Project-Specific Safety Rules

### Auth / Ownership
- The gateway has **no `/api/v1/auth/*` routes**. Never add them. Auth is Supabase's responsibility.
- Every scan-scoped route must use both Supabase RLS (DB layer) and the `Ownership` middleware (gateway layer).
- JWT validation must always call `jwt.WithAudience("authenticated")` and `jwt.WithExpirationRequired()`. Never relax these constraints.
- Never log JWTs, API keys, or the `SUPABASE_JWT_SECRET` value at any log level.

### SSRF Prevention
- All `target_endpoint` values must pass the `https_endpoint` custom validator before being stored or used.
- The validator blocks: non-HTTPS schemes, RFC1918 ranges (10/8, 172.16/12, 192.168/16), loopback (127/8), link-local (169.254/16), and the string `localhost`. This is non-negotiable.
- Never DNS-resolve `target_endpoint` in the gateway — SSRF via DNS rebinding is the agent's concern at call time.

### Secrets Handling
- Secrets live only in `.env` (local) or Kubernetes Secrets / AWS Secrets Manager (prod). Never hardcode them.
- `.env` is in `.gitignore`. Only `.env.example` (no real values) is committed.
- The only secret the gateway needs: `SUPABASE_JWT_SECRET` and `DATABASE_URL`. Keep it that way.
- In tests, use a throwaway `testSecret = []byte("test-secret-key-for-unit-tests-only")` — never a real project secret.

### Protobuf Compatibility
- Proto field numbers are permanent. If a field must change semantics, add a new field and deprecate the old one.
- The `proto/orchestrator/*.pb.go` stubs return `nil` from `ProtoReflect()` — they will panic on actual gRPC serialization. Run `make proto` before setting `ORCHESTRATOR_ENABLED=true`.

### External Dependency Boundaries
- The gateway must not import any Python agent package or call agent code directly — only via gRPC to the Orchestrator.
- The Orchestrator must not call the Gemini API directly — all LLM calls go through the `judge/` service or agent code.
- The React frontend must not connect directly to Kafka, PostgreSQL, or Redis — only via the gateway API and WebSocket.
- New Go dependencies require justification. Run `go mod tidy` after any `go get` and commit the updated `go.sum`.

---

## 11. Workflow: Explore → Plan → Implement → Commit

For every non-trivial change, follow this sequence:

1. **Explore** — read the relevant files before touching anything. Understand the existing code, interfaces, and tests. Use `Grep` and `Glob` tools rather than guessing.
2. **Plan** — before writing code, describe the approach. For multi-file changes, use `/plan` or write out the steps explicitly. Identify which tests will be affected.
3. **Implement** — write the failing test first (TDD). Then write the implementation. Keep changes focused; do not refactor adjacent code unless it is directly blocking.
4. **Quality gates** — run these before committing:

```bash
# Go
cd api-gateway
go build ./...
go test -race -cover ./...
gofmt -l .          # must output nothing

# Python
cd agents   # or judge/
ruff check .
pytest -v --cov=. --cov-report=term-missing

# React
cd frontend
npm run lint
npm test -- --coverage
```

5. **Commit** — see conventions below.

> **First-time setup** (api-gateway):
> ```bash
> cd api-gateway
> cp .env.example .env          # fill in SUPABASE_JWT_SECRET + DATABASE_URL
> go mod tidy                   # generates go.sum (requires Go 1.22+)
> docker compose up -d redis kafka
> go run ./cmd/server
> ```

---

## 12. Commit Message Conventions

Format: `<type>(<scope>): <imperative summary>`

| Type | When to use |
|------|-------------|
| `feat` | New feature or endpoint |
| `fix` | Bug fix |
| `test` | Adding or fixing tests (red→green step) |
| `refactor` | Refactor after green (refactor step) |
| `proto` | Protobuf schema changes |
| `migrate` | SQL migration additions |
| `chore` | Deps, CI config, build tooling |
| `docs` | CLAUDE.md, README, comments |

**Scopes:** `gateway`, `orchestrator`, `agents`, `judge`, `frontend`, `infra`, `kafka`, `ws`, `auth`

**Examples:**
```
test(gateway): add table-driven tests for SSRF validator
feat(gateway): implement ownership middleware for scan routes
refactor(gateway): extract rate-limit key builder to helper
proto(orchestrator): add ScanStatus rpc to OrchestratorService
migrate(gateway): add attack_results table for Sprint 2
fix(agents): handle empty response from Gemini Flash gracefully
```

TDD commits typically appear in sequence: `test(...)` → `feat(...)` → `refactor(...)`.

---

## 13. Claude Code Context Management

- **Start a new task:** Run `/clear` to reset context before switching to a different service or unrelated feature.
- **Long sessions:** Use `/compact` when context is filling up mid-task to summarize and continue.
- **Resume a session:** Use `claude --continue` to pick up exactly where you left off.
- **Per-service work:** Open Claude Code from within the service directory (`cd api-gateway && claude`) to keep context focused on that service's files.
- **Reading before editing:** Always read a file with the `Read` tool before editing it, even if you wrote it earlier in the session. The file may have been modified externally (e.g., `go mod tidy` updating `go.mod`).

---

## 14. Permissions & Allowlists

Claude Code uses a two-layer permission system for this project:

**Layer 1 — Machine-enforced allowlist** (`.claude/settings.json`):
Auto-approves the listed `Bash(...)` patterns without prompting. Covers all safe read/build/test/lint/format commands for Go, Python, React, Docker (local infra only), and git read operations. This file is committed to the repo so the allowlist is consistent for all contributors.

**Layer 2 — Policy rules below** (enforced by judgment):

Auto-approved (covered by `.claude/settings.json`):
- All `go build`, `go test`, `go mod tidy`, `gofmt`, `go vet`, `go env`
- All `make` targets: `build`, `test`, `lint`, `proto`, `tidy`, `docker-up`, `docker-down`
- `docker compose up -d redis kafka zookeeper` and `docker compose down` / `build` / `ps` / `logs`
- All `pytest`, `ruff`, `black`, `pip install -r requirements.txt`
- All `npm install`, `npm run lint`, `npm test`, `npm run build`
- `git status`, `git diff`, `git log`, `git add`, `git commit`, `git branch`, `git checkout`, `git stash`

**Always ask before running:**
- `git push` or `git push --force` (any remote write)
- `docker compose up` with `api-gateway` service (connects to live Supabase DB)
- Any `kubectl apply`, `helm upgrade`, or `terraform apply`
- Any SQL with `DROP`, `DELETE`, `TRUNCATE`, or `ALTER TABLE ... DROP COLUMN`
- `go get <new-package>` or `pip install <new-package>` not already in `go.mod` / `requirements.txt`
- Any command that writes to `.env` or touches secret values

---

## 15. Sprint Plan

- **Sprint 1 (Weeks 1-2):** Go API Gateway ✅ · Go Orchestrator skeleton · gRPC/Protobuf · Kafka · PostgreSQL/Redis · 3 Red Team Agents · LLM-as-Judge (30 golden pairs) · React skeleton · Jenkins/Docker CI → **Red team mode end-to-end**
- **Sprint 2 (Weeks 3-4):** Constraint Drift Agent · 4 Blue Team Agents · Adversarial mode · Result Aggregator · OWASP scorecard + PDF reports · WebSocket dashboard · Judge calibration (50+ pairs) · K8s/Helm/Terraform → **Full platform, all 3 modes**

---

## 16. Team

- **Shanshou Li:** Go API Gateway, Orchestrator, gRPC, Kafka, red/blue team agents, LLM-as-Judge, Jenkins CI/CD, K8s deployment
- **Yachen Wang:** React dashboard, frontend visualization, WebSocket integration, PDF report generation
