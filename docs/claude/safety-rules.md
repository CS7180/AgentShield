## Project-Specific Safety Rules

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
- In tests, use a throwaway `testSecret = []byte("test-secret-key-for-unit-tests-only")` — never a real project secret.

### Protobuf Compatibility
- Proto field numbers are permanent. If a field must change semantics, add a new field and deprecate the old one.
- The `proto/orchestrator/*.pb.go` stubs return `nil` from `ProtoReflect()` — they will panic on actual gRPC serialisation. Run `make proto` before setting `ORCHESTRATOR_ENABLED=true`.

### External Dependency Boundaries
- The gateway must not import any Python agent package or call agent code directly — only via gRPC to the Orchestrator.
- The Orchestrator must not call the Gemini API directly — all LLM calls go through the `judge/` service or agent code.
- The React frontend must not connect directly to Kafka, PostgreSQL, or Redis — only via the gateway API and WebSocket.
- New Go dependencies require justification. Run `go mod tidy` after any `go get` and commit the updated `go.sum`.

### OWASP LLM Top 10 Awareness
AgentShield tests target systems against these categories. The platform itself must not be vulnerable to the same issues:

| # | Category | Platform defence |
|---|----------|-----------------|
| LLM01 | Prompt Injection | Attack prompts stored as data, never interpolated into gateway logic |
| LLM02 | Insecure Output Handling | Judge reasoning treated as untrusted; not reflected in API errors |
| LLM06 | Sensitive Information Disclosure | Scan results and attack prompts gated behind ownership checks + RLS |
| LLM09 | Misinformation | Judge calibration golden set maintains ground truth for evaluation quality |
