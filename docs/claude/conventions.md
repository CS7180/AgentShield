## Coding Conventions

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

### React (`frontend/`)
- Use `eslint` + `prettier`. Run `npm run lint` before committing.
- All API calls go through a single `src/api/client.js` module. Never call `fetch` directly in components.
- Supabase Auth: use `@supabase/supabase-js` client. JWT is extracted from the Supabase session and passed as `Authorization: Bearer` to the gateway. Never store raw JWTs in `localStorage` directly — use the Supabase session object.
- Dark theme with purple accent (`#7C3AED`). Follow the design in `project_memory/AgentShield_*.jsx`.
- WebSocket connection: pass token as `?token=<jwt>` query param.

---

## Commit Message Conventions

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
