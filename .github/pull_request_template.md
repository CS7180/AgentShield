## Summary

<!-- 1-3 bullet points describing what this PR does and why -->

## Changes

<!-- List the key files/components changed and what was done -->

## C.L.E.A.R. Review

**Context** — What problem does this solve? What was the state before?

**Logic** — Why was this approach chosen over alternatives?

**Effort** — What's the scope? Rough line count or complexity?

**Assumptions** — What did you assume about the codebase, requirements, or user behaviour?

**Results** — What are the expected outcomes? How was it tested?

## Test Plan

- [ ] Unit tests pass (`npm test` / `go test ./...`)
- [ ] E2E tests pass (`npx playwright test`)
- [ ] Lint clean (`npm run lint`)
- [ ] Manual smoke test on affected pages/endpoints

## Security Checklist

- [ ] No secrets or credentials committed
- [ ] New endpoints have auth + ownership checks
- [ ] User inputs validated / sanitized
- [ ] No new SSRF surface introduced

## AI Disclosure

- **% AI-generated:** <!-- e.g. 80% -->
- **Tool:** Claude Code (claude-sonnet-4-6)
- **Human review:** <!-- describe what you verified manually -->
