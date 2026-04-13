# AgentShield Gap Check And Execution Plan

Date: 2026-04-13

## 1. Plan Gap Check (PRD Sprint Plan vs Current Repo)

Legend: `Done` = implemented in this repo; `Partial` = scaffolded/limited implementation; `Pending` = not implemented in this repo.

### Sprint 1

- Go API Gateway + Orchestrator skeleton + gRPC Protobuf: `Done` (current scope)
  Evidence: `api-gateway/` + standalone `orchestrator/` service with gRPC `StartScan/StopScan/ScanStatus`.
- Kafka setup + topic producer/consumer boilerplate: `Partial`
  Evidence: Kafka consumer + dispatcher exist in gateway; full topic lifecycle/bootstrap scripts not present.
- PostgreSQL schema + migrations + Redis setup: `Done` (gateway scope)
  Evidence: migrations `001`-`007`, Redis rate limit middleware, DB repos.
- Prompt Injection / Jailbreak / Data Leakage agents: `Partial`
  Evidence: standalone `agents/` service implemented as deterministic simulator; production-grade LLM agents pending.
- LLM-as-Judge basic + 30 golden pairs: `Partial`
  Evidence: standalone `judge/` service + calibration persistence exist; golden-set seed expansion is pending.
- React dashboard skeleton + report visualization: `Partial`
  Evidence: `frontend/` app and pages exist; end-to-end live data integration remains partial.
- GitLab CI + Docker Compose + Jenkins pipeline skeleton: `Partial`
  Evidence: Docker Compose + Makefiles + GitHub Actions CI workflow exist; Jenkins/GitLab pipeline configs are not complete in repo.
- End-to-end red-team integration + Sprint docs: `Partial`
  Evidence: orchestrator now calls agents + judge and persists attack results/reports; real model-driven execution pending.

### Sprint 2

- Constraint Drift + 4 blue-team agents: `Pending`
- Adversarial orchestration logic: `Partial`
  Evidence: mode passed through orchestrator pipeline; advanced red/blue co-simulation logic pending.
- Result aggregator and comparative report pipeline: `Done` (gateway scope)
  Evidence: automatic report generation endpoint + comparison endpoint + orchestrator DB persistence.
- Judge calibration pipeline and 50+ golden pairs: `Partial`
- OWASP scorecard and PDF generation pipeline: `Partial`
  Evidence: automated JSON scorecard and simple PDF generation exist; richer templated PDF/reporting still pending.
- WebSocket real-time dashboard + report comparison UI: `Partial`
  Evidence: gateway WS hub exists; frontend compare page now calls compare API, but live stream-driven report UX is still pending.
- Prometheus/Grafana + K8s/Helm/Terraform + Jenkins quality gate: `Pending`
- 80%+ coverage: `Pending`

## 2. Changes Completed In This Round

- Implemented `GET /api/v1/scans/:id/compare/:other_id` in gateway.
- Implemented judge calibration endpoints:
  - `POST /api/v1/judge/calibrate`
  - `GET /api/v1/judge/calibration-report`
- Persisted judge calibration reports to PostgreSQL (`judge_calibrations` table).
- Added attack result ingestion/list APIs (`POST/GET /api/v1/scans/:id/attack-results`).
- Added report generation API (`POST /api/v1/scans/:id/report/generate`) with JSON + optional PDF artifact upload.
- Added standalone `orchestrator/` Go gRPC service implementing `StartScan`, `StopScan`, and `ScanStatus`.
- Wired local Docker Compose stack to run orchestrator with API gateway.
- Added standalone `agents/` HTTP service (`/run-scan`) and `judge/` HTTP service (`/evaluate-batch`).
- Integrated orchestrator pipeline with agents + judge + PostgreSQL persistence (scans, attack_results, reports).
- Added ownership enforcement inside report comparison logic for both compared reports.
- Added/updated tests for report comparison and judge calibration.
- Added baseline coverage gate in `api-gateway/Makefile` (`coverage-check`, threshold configurable).
- Added GitHub Actions CI workflow for:
  - api-gateway tests + coverage gate
  - orchestrator/agents/judge tests
  - frontend production build
- Implemented frontend report comparison page wired to `GET /api/v1/scans/:id/compare/:other_id`.
- Updated README implementation status to match code.

## 3. Executable Next Plan (Doable In Current Repo)

### P0 (next)

1. Replace deterministic simulated agents/judge with model-backed implementations.
   Acceptance: `agents` and `judge` call configurable model providers and produce non-static outcomes.
2. Introduce async job queue + retries/dead-letter handling in orchestrator pipeline.
   Acceptance: failed runs are retried with bounded policy; dead-lettered failures are queryable.
3. Push scan progress events from orchestrator to gateway via Kafka/WebSocket path.
   Acceptance: running scan progress updates are visible in frontend without polling.

### P1

1. Add structured seed data for golden set and benchmark scripts.
2. Extend judge calibration pipeline to 50+ reproducible evaluation pairs.
3. Improve report PDF quality with richer templates and evidence sections.

### P2

1. Add Grafana dashboard templates and alert rule examples.
2. Add deployment assets for multi-service production rollout (k8s/helm/terraform).
3. Raise repository-wide coverage and define staged coverage targets per service.
