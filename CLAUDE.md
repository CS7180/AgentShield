# CLAUDE.md

This file is the authoritative operating guide for Claude Code working in this repository.

Full product requirements: @project_memory/AgentShield_PRD.md

React UI mockups (dark-themed, purple accent):
- `project_memory/AgentShield_Dashboard_Final.jsx`
- `project_memory/AgentShield_NewScan.jsx`
- `project_memory/AgentShield_ScanMonitor.jsx`
- `project_memory/AgentShield_Report.jsx`

---

## 1. Project Overview

**AgentShield** is a multi-agent red-blue teaming platform that automatically discovers security vulnerabilities in LLM-powered applications. It simulates adversarial attacks and defences running simultaneously against a target LLM to identify vulnerabilities before deployment.

**Three scan modes:**
- **Red Team** — attacks only, no defence, baseline measurement
- **Blue Team** — measures false-positive rate on normal traffic
- **Adversarial** — red + blue simultaneous, comparative analysis

---

## 2. Architecture, Tech Stack & Repository Structure

@docs/claude/architecture.md

---

## 3. Coding Conventions & Commit Standards

@docs/claude/conventions.md

---

## 4. Testing Strategy

@docs/claude/testing.md

---

## 5. REST API, Database & gRPC Reference

@docs/claude/api-reference.md

---

## 6. Security Rules & OWASP Awareness

@docs/claude/safety-rules.md

---

## 7. Workflow: Explore → Plan → Implement → Commit

For every non-trivial change, follow this sequence:

1. **Explore** — read the relevant files before touching anything. Use `Grep` and `Glob` tools rather than guessing.
2. **Plan** — describe the approach before writing code. For multi-file changes, use `/plan` or write out the steps explicitly.
3. **Implement** — write the failing test first (TDD). Then write the implementation. Keep changes focused.
4. **Quality gates** — run before committing:

```bash
# Go
cd api-gateway && go build ./... && go test -race -cover ./... && gofmt -l .

# Python
cd agents && ruff check . && pytest -v --cov=. --cov-report=term-missing

# React
cd frontend && npm run lint && npm test -- --coverage
```

5. **Commit** — follow the conventions in `docs/claude/conventions.md`.

---

## 8. Claude Code Context Management

- **Start a new task:** Run `/clear` before switching to a different service or unrelated feature.
- **Long sessions:** Use `/compact` when context is filling up mid-task.
- **Resume a session:** Use `claude --continue` to pick up exactly where you left off.
- **Per-service work:** Open Claude Code from within the service directory (`cd api-gateway && claude`).
- **Reading before editing:** Always read a file with the `Read` tool before editing, even if you wrote it earlier in the session.

---

## 9. Permissions & Allowlists

Claude Code uses a two-layer permission system:

**Layer 1 — Machine-enforced allowlist** (`.claude/settings.json`):
Auto-approves listed `Bash(...)` patterns without prompting. Also configures two hooks:
- **PostToolUse** — runs `npm run lint` after any frontend JS/JSX file is edited
- **Stop** — runs `npm test` as a quality gate when Claude finishes responding

**Layer 2 — Policy rules** (enforced by judgment):

Always ask before running:
- `git push` or `git push --force` (any remote write)
- `docker compose up` with `api-gateway` service (connects to live Supabase DB)
- Any `kubectl apply`, `helm upgrade`, or `terraform apply`
- Any SQL with `DROP`, `DELETE`, `TRUNCATE`, or `ALTER TABLE ... DROP COLUMN`
- Any command that writes to `.env` or touches secret values

---

## 10. Sprint Plan

Full sprint docs, retrospectives, and standup logs: `docs/sprints/` and `docs/standups/`

- **Sprint 1 (Weeks 1–2):** Go API Gateway ✅ · JWKS JWT validation ✅ · Scan CRUD + WebSocket ✅ · Frontend dashboard (all pages) ✅ · 95% frontend test coverage ✅
- **Sprint 2 (Weeks 3–4):** Playwright E2E · Vercel deploy · CI/CD hardening · Go coverage 80%+ · Blue team agents · Adversarial mode · OWASP PDF report · Judge calibration 50+ pairs

---

## 11. Team

- **Shanshou Li:** Go API Gateway, Orchestrator, gRPC, Kafka, red/blue team agents, LLM-as-Judge, CI/CD, K8s deployment
- **Yachen Wang:** React dashboard, frontend visualisation, WebSocket integration, PDF report generation, Playwright E2E, Vercel deployment
