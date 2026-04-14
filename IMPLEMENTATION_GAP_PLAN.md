# AgentShield Gap Check And Execution Plan

Date: 2026-04-13

## 1. Plan Gap Check (PRD Sprint Plan vs Current Repo)

Legend: `Done` = implemented in this repo; `Partial` = scaffolded/limited implementation; `Pending` = not implemented in this repo.

### Sprint 1

- Go API Gateway + Orchestrator skeleton + gRPC Protobuf: `Done` (current scope)
  Evidence: `api-gateway/` + standalone `orchestrator/` service with gRPC `StartScan/StopScan/ScanStatus`.
- Kafka setup + topic producer/consumer boilerplate: `Partial`
  Evidence: gateway Kafka consumer + WS dispatcher exist; orchestrator publishes scan status to Kafka; topic/bootstrap ops scripts are not complete.
- PostgreSQL schema + migrations + Redis setup: `Done` (gateway scope)
  Evidence: migrations `001`-`007`, Redis rate limit middleware, DB repositories.
- Prompt Injection / Jailbreak / Data Leakage agents: `Partial`
  Evidence: standalone `agents/` supports `simulate` and `target_http` modes; production-grade multi-agent/LLM workflows pending.
- LLM-as-Judge basic + 30 golden pairs: `Partial`
  Evidence: standalone `judge/` supports `rule` and `openai_compat`; calibration persistence exists; expanded golden-set benchmark still pending.
- React dashboard skeleton + report visualization: `Done` (current scope)
  Evidence: dashboard/new-scan/monitor/report-compare/judge/settings pages are wired to live APIs.
- GitLab CI + Docker Compose + Jenkins pipeline skeleton: `Partial`
  Evidence: Docker Compose + Makefiles + GitHub Actions CI exist; Jenkins/GitLab pipeline configs are not complete in repo.
- End-to-end red-team integration + Sprint docs: `Done` (current scope)
  Evidence: orchestrator calls agents + judge, persists results/reports, and emits status updates for gateway WS fanout.

### Sprint 2

- Constraint Drift + 4 blue-team agents: `Pending`
- Adversarial orchestration logic: `Partial`
  Evidence: mode is propagated through pipeline; advanced red/blue co-simulation strategy pending.
- Result aggregator and comparative report pipeline: `Done` (gateway scope)
  Evidence: report generation endpoint + comparison endpoint + orchestrator DB persistence.
- Judge calibration pipeline and 50+ golden pairs: `Partial`
- OWASP scorecard and PDF generation pipeline: `Partial`
  Evidence: JSON scorecard and basic PDF generation exist; richer templated reporting is pending.
- WebSocket real-time dashboard + report comparison UI: `Done` (current scope)
  Evidence: gateway WS hub active, orchestrator Kafka status publishing active, monitoring/report compare UI wired.
- Prometheus/Grafana + K8s/Helm/Terraform + Jenkins quality gate: `Pending`
- 80%+ coverage: `Pending`

## 2. Changes Completed In This Round

- Frontend pages switched from mock/placeholder states to API-driven data flows:
  - `DashboardContent.jsx`
  - `NewScanContent.jsx`
  - `ScanMonitorContent.jsx`
  - `JudgeContent.jsx` (new)
  - `SettingsContent.jsx` (new)
- Removed obsolete mock UI files:
  - `frontend/src/AgentShield_Dashboard_Final.jsx`
  - `frontend/src/AgentShield_NewScan.jsx`
  - `frontend/src/AgentShield_Report.jsx`
  - `frontend/src/AgentShield_ScanMonitor.jsx`
- Expanded frontend API client to cover scans, attack results, report generation, judge calibration, and exported `API_BASE`.
- Added orchestrator scan status event publisher abstraction + Kafka implementation (`publisher.go`).
- Published orchestrator lifecycle/progress events to Kafka in pipeline execution path.
- Wired orchestrator startup to use Kafka status publisher with noop fallback.
- Added orchestrator bounded retry execution with exponential backoff:
  - `ORCHESTRATOR_EXEC_MAX_ATTEMPTS`
  - `ORCHESTRATOR_EXEC_RETRY_BASE_MS`
  - `ORCHESTRATOR_EXEC_RETRY_MAX_MS`
- Added idempotent retry behavior by resetting per-scan `attack_results` and `reports` before each retry attempt.
- Added dead-letter persistence (`scan_dead_letters`) after retry exhaustion and gateway query endpoint:
  - `GET /api/v1/scans/:id/dead-letters`
- Added frontend monitoring linkage for report lifecycle:
  - auto-generate report artifacts on scan completion when needed
  - report readiness status and manual regenerate action
  - dead-letter visibility in monitoring panel
- Added runtime mode toggles and env wiring:
  - `AGENTS_EXECUTION_MODE=simulate|target_http`
  - `JUDGE_EVAL_MODE=rule|openai_compat`
  - `JUDGE_LLM_BASE_URL`, `JUDGE_LLM_API_KEY`, `JUDGE_LLM_MODEL`
  - `ORCHESTRATOR_SCAN_STATUS_TOPIC`
- Updated compose/env examples and service READMEs for new integration/runtime modes.
- Re-ran gateway/orchestrator/agents/judge tests and frontend lint/build (pass).

## 3. Executable Next Plan (Doable In Current Repo)

### P0 (next)

1. Introduce async queue worker lifecycle in orchestrator pipeline.
   Acceptance: execution can be decoupled from direct request lifecycle with persistent queue semantics.
2. Harden model-backed execution paths.
   Acceptance: `target_http` and `openai_compat` support configurable request/response schemas and deterministic fallback behavior.
3. Add orchestration-level report artifact generation strategy.
   Acceptance: artifact generation (JSON/PDF path lifecycle) is deterministic for all deployment modes.

### P1

1. Add reproducible golden-set fixtures and benchmark scripts.
2. Extend judge calibration pipeline to 50+ reproducible evaluation pairs.
3. Improve report PDF quality with richer templates and evidence sections.

### P2

1. Add Grafana dashboard templates and alert rule examples.
2. Add deployment assets for multi-service production rollout (k8s/helm/terraform).
3. Raise repository-wide coverage and define staged coverage targets per service.
