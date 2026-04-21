## Testing Strategy (Strict TDD)

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
- Unit test all custom hooks and utility functions with Vitest + React Testing Library.
- Integration-test the scan creation flow end-to-end.
- Coverage thresholds: 70% lines, 70% functions, 70% statements, 60% branches (enforced in `vitest.config.js`).
- Current coverage: ~95% lines, ~80% branches (as of Sprint 1).

### E2E tests (Playwright)
```bash
cd frontend
npx playwright test
```
- At minimum: login flow, scan creation flow, scan monitor page load.
- Run in CI after unit tests pass.
