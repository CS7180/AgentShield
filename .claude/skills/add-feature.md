---
name: add-feature
description: >
  Automates the AgentShield Explore → Plan → Implement → Commit workflow
  for any new feature. Enforces TDD, security rules, quality gates, and
  commit format. Auto-detects service (Go vs React) and applies the correct
  conventions. Includes a mandatory security checklist for any code touching
  auth, endpoints, or database queries.
usage: /add-feature <feature description>
---

You are executing the AgentShield feature-addition workflow for:

**Feature:** $ARGUMENTS

---

## SERVICE DETECTION

Before anything else, determine the service scope from the feature description and existing file structure:

- **Go (api-gateway/):** New handler, middleware, domain type, migration, gRPC change, Kafka producer/consumer, WebSocket change, validation rule.
- **React (frontend/):** New page, component, hook, auth flow change, WebSocket client change.
- **Both:** Feature that adds a gateway API endpoint AND a React UI that calls it.

State your determination and which root directories are in scope. All subsequent steps apply only to the detected scope(s).

---

## SECURITY CLASSIFICATION

Classify the feature against these three tiers before writing any code:

- **Tier 1 — Standard:** No auth, no external URLs, no DB writes.
- **Tier 2 — Data path:** Reads/writes DB or cache; no new auth logic.
- **Tier 3 — Security-sensitive:** Any of: new auth mechanism, new `target_endpoint`-style input, new JWT claim access, new DB table with user data, new middleware, cross-user data access.

State your tier. **Tier 3 makes the security checklist in STEP 3b mandatory.**

---

## STEP 1 — EXPLORE

Read every file you will touch. Use `Read`, `Grep`, `Glob` — never guess at file contents.

Checklist:
- [ ] Read all files to be modified
- [ ] Read the test files for each affected package
- [ ] Read the interface definitions the new code must satisfy (defined at the consumer)
- [ ] Identify an existing similar feature as a reference pattern
- [ ] If Go: check `api-gateway/migrations/` for the next migration number
- [ ] If React: confirm whether `src/api/client.ts` exists yet

Do not proceed until you can state: "I have read all relevant files and understand the existing patterns: [list them]."

---

## STEP 2 — PLAN

Write the complete plan before writing any code.

**Go features:**
- List files to create/modify with a one-line description each
- New exported types and function signatures
- Table-driven test cases (inputs, expected outputs, error conditions)
- Migration filename if DB schema changes: `NNN_<snake_case>.sql` with `IF NOT EXISTS` guards — no `DROP`, `DELETE`, `TRUNCATE`, `ALTER ... DROP COLUMN`

**React features:**
- Components, hooks, and API client functions to add in `src/api/client.ts`
- Confirm no direct `fetch` calls in components, no raw JWT from `localStorage`

State the commit sequence (`test → feat → refactor`).

---

## STEP 3 — IMPLEMENT (TDD: Red → Green → Refactor)

### 3a. Red — Write the Failing Test First

Write the test file before any implementation. Confirm it fails to compile or fails at runtime.

**Go:**
- Package: `package <pkg>_test` (external test package)
- Table-driven: `[]struct{ name, input, want }{...}` for validators, JWT, state transitions
- JWT test secret: `var testSecret = []byte("test-secret-key-for-unit-tests-only")`
- Gin tests: `gin.SetMode(gin.TestMode)` + `httptest.NewRecorder()`
- Skip DB tests when `DATABASE_URL` is unset
- Confirm the test fails before proceeding

**React:** Unit tests for custom hooks and utilities; MSW for API integration tests.

### 3b. Green — Minimum Implementation

Write only the code needed to make the tests pass.

**Go rules:**
- `gofmt` + `goimports` on every file you create or modify
- `fmt.Errorf("context: %w", err)` — never swallow errors silently
- `c.AbortWithStatusJSON(...)` on Gin error paths; no `return` after abort
- Every I/O function: `ctx context.Context` as first argument
- No `init()` except `metrics/prometheus.go`
- No naked `panic()` except in `main.go` before server starts

**React rules:**
- All API calls through `src/api/client.ts` — never call `fetch` directly in components
- Auth: `const { session } = useAuth()` → `session.access_token` for Bearer header
- WS connections: `?token=<jwt>` query param (browsers cannot set WS headers)
- Styling: TailwindCSS, dark theme, `#7C3AED` purple accent; follow patterns in existing components

**Security Checklist (mandatory for Tier 3 — skip for Tier 1/2):**

SSRF:
- [ ] Any field accepting a URL uses the `https_endpoint` custom validator in `validation/validator.go`
- [ ] The validator blocks: non-HTTPS schemes, RFC1918 (10/8, 172.16/12, 192.168/16), loopback (127/8), link-local (169.254/16), `localhost`
- [ ] No DNS resolution of `target_endpoint` inside the gateway

JWT / Auth:
- [ ] `jwt.WithAudience("authenticated")` and `jwt.WithExpirationRequired()` are always set — never relaxed
- [ ] No `/api/v1/auth/*` routes added — auth is Supabase's responsibility
- [ ] No JWT, API key, or `SUPABASE_JWT_SECRET` value appears in any log statement

RLS / Ownership:
- [ ] New user-data tables have `ENABLE ROW LEVEL SECURITY` + `user_id = auth.uid()` policy
- [ ] New scan-scoped routes use both `JWTAuth` middleware and `Ownership` middleware

Secrets:
- [ ] No secrets hardcoded — `.env` only (local) or K8s Secrets / AWS Secrets Manager (prod)
- [ ] Only `.env.example` (no real values) committed; `.env` stays gitignored

External Boundaries:
- [ ] Gateway does not import Python agent packages — only calls Orchestrator via gRPC
- [ ] React frontend does not connect directly to Kafka, PostgreSQL, or Redis

### 3c. Refactor

Clean up after green without changing behavior. Re-run tests after each change.

---

## STEP 4 — QUALITY GATES (all are blockers)

**Go (`api-gateway/`):**
```bash
go build ./...
go test -race -cover ./...
gofmt -l $(git diff --name-only HEAD -- '*.go')
```

Use `git diff --name-only HEAD -- '*.go'` to scope `gofmt` to **files you changed** — not pre-existing issues or generated proto stubs. Any listed file that you created or modified is a blocker. Pre-existing unformatted files and `proto/**/*.pb.go` are exempt.

Coverage target: 80%+ on `auth`, `middleware`, `validation`, `handler`, `ws` packages. If a new package falls below 80%, add test cases before committing.

**React (`frontend/`):**
```bash
npx eslint $(git diff --name-only HEAD -- 'frontend/src/**/*.{js,jsx,ts,tsx}' | tr '\n' ' ')
```

Scope ESLint to **files you created or modified** — not pre-existing issues in unchanged files. If no React files were changed, skip this gate.

Run both gate sets if the feature touches both services.

---

## STEP 5 — COMMIT

Format: `<type>(<scope>): <imperative summary>`

Types: `feat`, `fix`, `test`, `refactor`, `proto`, `migrate`, `chore`, `docs`
Scopes: `gateway`, `orchestrator`, `agents`, `judge`, `frontend`, `infra`, `kafka`, `ws`, `auth`

TDD commit sequence:
```
test(scope): add failing tests for <feature>
feat(scope): implement <feature>
refactor(scope): <what was cleaned up>   ← only if refactoring done
```

**Hard stops — never commit if:**
- Any quality gate is failing (build, test, lint)
- `gofmt` reports any file you created or modified
- A JWT, API key, or `SUPABASE_JWT_SECRET` value appears in staged files
- A new `/api/v1/auth/*` route is staged
- A Tier 3 feature skipped the security checklist
