# AgentShield

AgentShield is a multi-agent red-blue teaming platform for discovering security vulnerabilities in LLM-powered applications before deployment.

This repository currently contains the Go API gateway and product reference artifacts. The broader platform described in the PRD includes a Go orchestrator, Python-based red and blue team agents, an LLM-as-Judge service, a report generator, and a React dashboard.

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
- `project_memory/` contains the PRD and UI references
- `orchestrator/`, `agents/`, `judge/`, and `frontend/` are planned but not yet present in this repo

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

The following features are scaffolded but not implemented yet:

- judge calibration endpoints
- report JSON generation
- PDF report generation
- scan comparison
- production orchestrator integration beyond the current stub mode

When `ORCHESTRATOR_ENABLED=false`, the gateway uses a stub orchestrator and newly started scans are queued instead of being executed end-to-end.

## Repository Layout

```text
AgentShield/
├── api-gateway/      Go API gateway
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
- `ALLOWED_ORIGINS=http://localhost:3000` controls CORS

### Run Locally

Start Redis and Kafka:

```bash
cd api-gateway
make docker-up
```

Run the gateway:

```bash
cd api-gateway
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
- `GET /api/v1/scans/:id/report`
- `GET /api/v1/scans/:id/report/pdf`
- `GET /api/v1/scans/:id/compare/:other_id`
- `POST /api/v1/judge/calibrate`
- `GET /ws/scans/:id/status`

Authentication model:

- the frontend authenticates directly with Supabase
- API requests must send `Authorization: Bearer <supabase_jwt>`
- WebSocket connections pass the token with `?token=<supabase_jwt>`

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
