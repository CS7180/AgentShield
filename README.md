# AgentShield

AgentShield is a multi-agent red-blue teaming platform for discovering security vulnerabilities in LLM-powered applications before deployment.

This repository currently contains the Go API gateway, standalone orchestrator/agents/judge services, the React dashboard client, and product reference artifacts.

## Product Vision

AgentShield is designed to let security and AI engineering teams point the system at a target LLM API and run repeatable security scans across three modes:

- `red_team`: attacks only, used to establish a vulnerability baseline
- `blue_team`: defenses only, used to measure false positives on normal traffic
- `adversarial`: red and blue agents run simultaneously to measure defense effectiveness under attack

The PRD defines four red-team agent categories and four blue-team agent categories:

- Red team: prompt injection, jailbreak, data leakage, and constraint drift
- Blue team: input guard, output filter, behavior monitor, and constraint persistence

The planned output is an OWASP LLM Top 10-aligned report available in both JSON and PDF formats, with real-time scan status surfaced in the dashboard.

## Architecture

At the product level, AgentShield is designed around the following flow:

```text
React Dashboard
      |
      v
Go API Gateway
      |
      v
Go Orchestrator
      |
      +--> Red Team Agents (Python / CrewAI)
      +--> Blue Team Agents (Python / CrewAI)
      +--> LLM-as-Judge
      |
      v
Kafka -> Result Aggregation -> Report Generation
      |
      v
PostgreSQL + Redis
```

The current repository state is earlier than that full target architecture:

- `api-gateway/` is present and functional as the current service entry point
- `frontend/` is present as the React dashboard client
- `project_memory/` contains the PRD and UI references
- `orchestrator/` is present as a standalone gRPC service used by the gateway
- `agents/` and `judge/` are present as standalone HTTP services for execution/evaluation

## Current Implementation Status

The Go API gateway already includes:

- Supabase JWT validation
- REST endpoints for scan creation, listing, retrieval, start, and stop
- ownership checks for per-user scan access
- Redis-backed rate limiting
- Kafka consumer wiring for real-time event dispatch
- WebSocket status streaming
- PostgreSQL persistence and startup migrations
- Prometheus health and metrics endpoints
- gRPC orchestration integration to standalone orchestrator service

The following API and integration capabilities are implemented:

- judge calibration calculation endpoint (`POST /api/v1/judge/calibrate`) and latest calibration report retrieval (`GET /api/v1/judge/calibration-report`) with PostgreSQL persistence
- attack result ingestion/list endpoints for each scan
- report JSON upsert/retrieval, report generation from attack results, and PDF artifact path retrieval
- scan report comparison (`GET /api/v1/scans/:id/compare/:other_id`)
- frontend pages wired to live APIs for dashboard, scan creation/start, monitoring, report compare, and judge calibration
- WebSocket scan status feed path wired end-to-end (`orchestrator -> Kafka topic -> gateway dispatcher -> /ws/scans/:id/status`)
- service runtime modes for local-to-real transition:
  - `AGENTS_EXECUTION_MODE=simulate|target_http`
  - `JUDGE_EVAL_MODE=rule|openai_compat` (+ `JUDGE_LLM_*`)
- repository CI workflow for api-gateway/orchestrator/agents/judge tests and frontend build
- baseline api-gateway coverage gate (`make coverage-check`, default threshold: 25%)

The following platform capabilities are still pending:

- production-grade Python/CrewAI multi-agent implementations (current `agents` service supports simulator or direct target HTTP execution)
- richer LLM-as-Judge calibration benchmark workflows and expanded golden set (current `judge` service supports rule or OpenAI-compatible inference mode)
- async queue workers with retry/dead-letter semantics and deployment automation for multi-service production rollout

When `ORCHESTRATOR_ENABLED=false`, the gateway uses a stub orchestrator and newly started scans are queued instead of being executed end-to-end.

## Repository Layout

```text
AgentShield/
├── api-gateway/      Go API gateway
├── orchestrator/     Go gRPC orchestrator service
├── agents/           Attack execution service
├── judge/            Evaluation service
├── frontend/         React dashboard
├── project_memory/   PRD and UI mockups
├── CLAUDE.md         repository engineering guide
└── README.md
```

Important reference files:

- `project_memory/AgentShield_PRD.md`: product requirements and target architecture
- `project_memory/AgentShield_*.jsx`: dashboard and report UI mockups
- `api-gateway/Makefile`: local build, run, test, and Docker commands

## Quick Start

### Prerequisites

- Go 1.22
- Docker and Docker Compose
- A Supabase project for PostgreSQL and Auth

### Environment Setup

Copy the example env file and fill in the required values:

```bash
cd api-gateway
cp .env.example .env
```

Required settings:

- `SUPABASE_JWT_SECRET`: Supabase JWT secret used to validate Bearer tokens
- `DATABASE_URL`: PostgreSQL connection string for your Supabase database
- `REDIS_URL`: local Redis URL for rate limiting and caching
- `KAFKA_BROKERS`: Kafka broker list for event consumption

Optional settings:

- `ORCHESTRATOR_ENABLED=false` keeps the gateway in stub mode for local development
- set `ORCHESTRATOR_ENABLED=true` and `ORCHESTRATOR_ADDR=orchestrator:50051` to use the standalone orchestrator service
- `ALLOWED_ORIGINS=http://localhost:3000` controls CORS

### Run Locally

Start local services (agents, judge, orchestrator, Redis, Kafka):

```bash
cd api-gateway
make docker-full
```

Run the gateway:

```bash
cd api-gateway
make run
```

Run frontend dashboard:

```bash
cd frontend
npm install
npm run dev
```

Run services manually in separate terminals (optional alternative to `make docker-full`):

```bash
cd orchestrator
make run
```

```bash
cd agents
make run
```

```bash
cd judge
make run
```

Build and test:

```bash
cd api-gateway
make build
make test
```

The gateway listens on `http://localhost:8080` by default.

## API Surface

Current gateway endpoints:

- `GET /health`
- `GET /metrics`
- `POST /api/v1/scans`
- `GET /api/v1/scans`
- `GET /api/v1/scans/:id`
- `POST /api/v1/scans/:id/start`
- `POST /api/v1/scans/:id/stop`
- `POST /api/v1/scans/:id/attack-results`
- `GET /api/v1/scans/:id/attack-results`
- `GET /api/v1/scans/:id/dead-letters`
- `PUT /api/v1/scans/:id/report`
- `GET /api/v1/scans/:id/report`
- `GET /api/v1/scans/:id/report/pdf`
- `POST /api/v1/scans/:id/report/generate`
- `GET /api/v1/scans/:id/compare/:other_id`
- `POST /api/v1/judge/calibrate`
- `GET /api/v1/judge/calibration-report`
- `GET /ws/scans/:id/status`

Authentication model:

- the frontend authenticates directly with Supabase
- API requests must send `Authorization: Bearer <supabase_jwt>`
- WebSocket connections pass the token with `?token=<supabase_jwt>`

Retry and DLQ behavior:

- orchestrator retries scan execution with exponential backoff (`ORCHESTRATOR_EXEC_MAX_ATTEMPTS`, `ORCHESTRATOR_EXEC_RETRY_BASE_MS`, `ORCHESTRATOR_EXEC_RETRY_MAX_MS`)
- terminal failures are persisted in `scan_dead_letters` and queryable via `GET /api/v1/scans/:id/dead-letters`

## Product Roadmap

The PRD targets a broader multi-service platform with:

- Go orchestrator and gRPC agent execution
- Python red-team and blue-team agent services
- LLM-as-Judge calibration and evaluation workflows
- OWASP LLM Top 10 report generation in JSON and PDF
- React dashboard for scan setup, monitoring, and report analysis
- Jenkins integration for security scan quality gates
- Kubernetes deployment and Prometheus/Grafana monitoring

## Notes

- `project_memory/` is a reference area and should be treated as read-only
- the current repo is an implementation slice of the full platform, not the entire product described in the PRD
- if you are using Supabase for a long-running backend service, choose the database connection mode intentionally instead of treating the example `DATABASE_URL` as production-ready verbatim text
