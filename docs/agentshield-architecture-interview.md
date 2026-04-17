# AgentShield Architecture Interview Guide

This document is based on the current implementation in this repository, not only the PRD target vision.

## 1. Executive Summary

AgentShield is a multi-service LLM security scanning platform.

Its core job is:

1. accept a user-authenticated scan request,
2. orchestrate attack execution against a target LLM endpoint,
3. evaluate the responses with a Judge service,
4. persist structured findings and generated reports,
5. stream real-time scan status back to the dashboard.

The current implementation is centered on:

- `frontend`: React dashboard
- `api-gateway`: Go REST/WebSocket gateway
- `orchestrator`: Go gRPC orchestration service
- `agents`: HTTP attack execution service
- `judge`: HTTP evaluation service
- `PostgreSQL`: system-of-record for scans, attack results, reports, calibration, DLQ
- `Redis`: rate limiting
- `Kafka`: async event bus for scan status fan-out
- `Supabase`: Auth + Postgres hosting + object storage for report artifacts

Important interview framing:

- Current state: the platform already works end-to-end as a service-oriented scan pipeline.
- Planned evolution: richer Python/CrewAI agents, stronger Judge benchmarking, and more production-grade async workers.

## 2. High-Level Architecture

```mermaid
flowchart LR
    U["User"]
    F["Frontend<br/>React + Vite"]
    S["Supabase Auth<br/>Google OAuth + JWT"]
    G["API Gateway<br/>Go + Gin"]
    O["Orchestrator<br/>Go + gRPC"]
    A["Agents Service<br/>HTTP /run-scan"]
    J["Judge Service<br/>HTTP /evaluate-batch"]
    K["Kafka<br/>agent.status topic"]
    W["WebSocket Hub"]
    P["PostgreSQL<br/>scans / attack_results / reports / judge_calibrations / scan_dead_letters"]
    R["Redis<br/>Rate Limiting"]
    B["Supabase Storage<br/>report.json / report.pdf"]
    T["Target LLM App<br/>HTTPS endpoint"]

    U --> F
    F --> S
    S --> F
    F -->|"Bearer JWT + REST"| G
    F -->|"WS ?token=JWT"| G

    G --> R
    G --> P
    G -->|"gRPC StartScan/StopScan"| O

    O -->|"POST /run-scan"| A
    A --> T
    O -->|"POST /evaluate-batch"| J
    O --> P
    O -->|"publish status"| K

    K --> G
    G --> W
    W --> F

    G -->|"upload report artifacts"| B
```

## 3. Component Responsibilities

### Frontend

- authenticates users with Supabase
- stores session and access token
- creates scans and starts/stops them
- fetches scan list, results, reports, dead letters, and Judge calibration report
- subscribes to real-time scan status over WebSocket

Primary pages:

- `Dashboard`
- `Create Scan`
- `Scan Monitoring`
- `Report Compare`
- `Judge Calibration`
- `Settings`

### API Gateway

The API gateway is the system boundary for clients.

It is responsible for:

- JWT validation
- ownership enforcement
- request validation
- REST API exposure
- report generation endpoints
- WebSocket endpoint for live scan status
- Kafka consumer wiring for event fan-out
- persistence repositories
- rate limiting with Redis
- orchestrator integration over gRPC

### Orchestrator

The orchestrator is the execution coordinator.

It is responsible for:

- accepting `StartScan` and `StopScan` over gRPC
- maintaining in-memory scan execution state and cancellation context
- calling downstream services in order
- updating progress milestones
- retrying failed executions with exponential backoff
- writing dead-letter records after retry exhaustion
- publishing scan status events to Kafka

### Agents Service

The agents service generates attack prompts and executes attacks.

Current modes:

- `simulate`: deterministic simulated responses
- `target_http`: sends HTTP requests to the target endpoint

Current attack categories:

- `prompt_injection`
- `jailbreak`
- `data_leakage`
- `constraint_drift`

### Judge Service

The Judge service evaluates attack outcomes.

Current modes:

- `rule`: deterministic heuristic evaluator
- `openai_compat`: OpenAI-compatible `/chat/completions` judging

The Judge outputs:

- `severity`
- `owasp_category`
- `confidence`
- `reasoning`
- `defense_intercepted`

## 4. End-to-End Scan Sequence

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Supabase as Supabase Auth
    participant Gateway as API Gateway
    participant Orchestrator
    participant Agents
    participant Target as Target LLM Endpoint
    participant Judge
    participant Postgres
    participant Kafka
    participant WS as WebSocket

    User->>Frontend: Sign in with Google
    Frontend->>Supabase: OAuth sign-in
    Supabase-->>Frontend: JWT session

    User->>Frontend: Create scan
    Frontend->>Gateway: POST /api/v1/scans
    Gateway->>Postgres: insert scan(status=pending)
    Gateway-->>Frontend: scan created

    User->>Frontend: Start scan
    Frontend->>Gateway: POST /api/v1/scans/:id/start
    Gateway->>Orchestrator: StartScan(scan_id, target, mode, attack_types)
    Gateway->>Postgres: mark scan started/running
    Gateway-->>Frontend: accepted

    Frontend->>Gateway: WS /ws/scans/:id/status?token=JWT
    Gateway-->>WS: subscribe client to scan room

    Orchestrator->>Kafka: publish running/progress
    Kafka->>Gateway: consume status event
    Gateway-->>Frontend: push WS update

    Orchestrator->>Agents: POST /run-scan
    Agents->>Target: POST attack payloads
    Target-->>Agents: target responses
    Agents-->>Orchestrator: attack results

    Orchestrator->>Judge: POST /evaluate-batch
    Judge-->>Orchestrator: evaluations

    Orchestrator->>Postgres: insert attack_results
    Orchestrator->>Postgres: upsert reports
    Orchestrator->>Postgres: mark scan completed

    Orchestrator->>Kafka: publish completed
    Kafka->>Gateway: consume status event
    Gateway-->>Frontend: push completed update

    Frontend->>Gateway: POST /api/v1/scans/:id/report/generate
    Gateway->>Postgres: load scan + attack_results
    Gateway->>Gateway: build report summary
    Gateway->>Supabase: upload report.json/report.pdf
    Gateway->>Postgres: upsert reports with artifact paths
    Gateway-->>Frontend: report ready
```

## 5. Runtime Control Flow

```mermaid
flowchart TD
    A["Create Scan"] --> B["Scan row inserted<br/>status=pending"]
    B --> C["Start Scan API"]
    C --> D{"Orchestrator available?"}
    D -->|No| E["status=queued"]
    D -->|Yes| F["gRPC StartScan accepted"]
    F --> G["Manager creates run context"]
    G --> H["publish running event"]
    H --> I["Executor marks running in DB"]
    I --> J["Agents service executes attacks"]
    J --> K["Judge service evaluates"]
    K --> L["Persist attack_results"]
    L --> M["Upsert report summary"]
    M --> N["Mark completed"]
    N --> O["publish completed event"]

    J --> P{"Execution error?"}
    K --> P
    L --> P
    M --> P
    N --> P

    P -->|Yes| Q["Retry with exponential backoff"]
    Q --> R{"Attempts exhausted?"}
    R -->|No| J
    R -->|Yes| S["Mark failed"]
    S --> T["Persist scan_dead_letter"]
    T --> U["publish failed event"]
```

## 6. State Machine

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> queued
    pending --> running
    queued --> running
    queued --> stopped
    running --> completed
    running --> failed
    running --> stopped
```

State meanings:

- `pending`: created but not started
- `queued`: accepted by API but orchestrator unavailable
- `running`: active execution
- `completed`: pipeline finished successfully
- `failed`: retries exhausted
- `stopped`: user-requested stop or terminal stop

## 7. Real-Time Eventing Design

```mermaid
flowchart LR
    O["Orchestrator"] -->|"Publish JSON event"| K["Kafka topic: agent.status"]
    K --> C["Gateway Consumer Group"]
    C --> D["Dispatcher"]
    D --> H["WebSocket Hub"]
    H --> R1["scan room: <scan_id>"]
    R1 --> F["Frontend Monitor"]
```

Design rationale:

- Orchestrator does not push directly to browser clients.
- Kafka decouples execution from UI delivery.
- API Gateway remains the only browser-facing surface.
- WebSocket rooms are keyed by `scan_id`, so the frontend only receives relevant events.

## 8. Authentication and Authorization

```mermaid
flowchart TD
    A["User signs in with Google"] --> B["Supabase issues JWT"]
    B --> C["Frontend stores session"]
    C --> D["REST calls include Authorization: Bearer <jwt>"]
    C --> E["WS connection includes ?token=<jwt>"]

    D --> F["Gateway validates JWT with SUPABASE_JWT_SECRET"]
    E --> F
    F --> G["Extract user_id from sub claim"]
    G --> H["Ownership middleware checks scan ownership"]
    H --> I["User can only access own scans/results/reports"]
```

Security boundaries:

- browser authenticates with Supabase, not with the gateway directly
- gateway validates Supabase-issued JWTs
- database tables enforce Row Level Security
- middleware also enforces per-user ownership at API layer

## 9. Persistence Model

```mermaid
erDiagram
    AUTH_USERS ||--o{ SCANS : owns
    SCANS ||--o{ ATTACK_RESULTS : has
    SCANS ||--|| REPORTS : summarizes
    SCANS ||--o| SCAN_DEAD_LETTERS : may_fail_into
    AUTH_USERS ||--|| JUDGE_CALIBRATIONS : has_latest

    SCANS {
        uuid id PK
        uuid user_id FK
        string target_endpoint
        string mode
        string[] attack_types
        string status
        timestamptz created_at
        timestamptz started_at
        timestamptz completed_at
    }

    ATTACK_RESULTS {
        uuid id PK
        uuid scan_id FK
        uuid user_id FK
        string attack_type
        text attack_prompt
        text target_response
        boolean attack_success
        string severity
        string owasp_category
        boolean defense_intercepted
        float judge_confidence
        text judge_reasoning
    }

    REPORTS {
        uuid id PK
        uuid scan_id FK
        uuid user_id FK
        numeric overall_score
        int critical_count
        int high_count
        int medium_count
        int low_count
        jsonb owasp_scorecard
        jsonb report_json
        string report_pdf_path
    }

    SCAN_DEAD_LETTERS {
        uuid id PK
        uuid scan_id FK
        uuid user_id FK
        int attempt_count
        string error_stage
        text error_message
        jsonb payload
        timestamptz failed_at
    }

    JUDGE_CALIBRATIONS {
        uuid user_id PK
        int sample_count
        jsonb report_json
        timestamptz generated_at
    }
```

Persistence strategy:

- `scans` is the lifecycle root aggregate
- `attack_results` stores per-attack evidence
- `reports` stores aggregated security posture and artifact metadata
- `scan_dead_letters` stores terminal failures for retry visibility
- `judge_calibrations` stores the latest per-user Judge quality report

## 10. Report Generation Design

```mermaid
flowchart TD
    A["Load scan"] --> B["Load all attack_results"]
    B --> C["Aggregate counts by severity"]
    C --> D["Build OWASP scorecard"]
    D --> E["Compute overall_score"]
    E --> F["Build report_json payload"]
    F --> G{"Uploader configured?"}
    G -->|Yes| H["Upload report.json"]
    H --> I["Optionally upload report.pdf"]
    I --> J["Upsert reports row with artifact paths"]
    G -->|No| K["Persist partial report without storage artifacts"]
```

Current report semantics:

- score is penalty-based from successful findings
- JSON report is the canonical structured artifact
- PDF is a simplified generated artifact
- object storage paths are stored in the `reports` table

## 11. Deployment Topology

```mermaid
flowchart LR
    subgraph Browser
        FE["React Frontend"]
    end

    subgraph Platform
        GW["API Gateway"]
        OR["Orchestrator"]
        AG["Agents"]
        JU["Judge"]
        RD["Redis"]
        KF["Kafka"]
    end

    subgraph Supabase
        AU["Auth"]
        PG["Postgres"]
        ST["Storage"]
    end

    FE --> GW
    FE --> AU
    GW --> OR
    GW --> RD
    GW --> KF
    GW --> PG
    GW --> ST
    OR --> AG
    OR --> JU
    OR --> KF
    OR --> PG
```

Current local-development deployment:

- `agents`, `judge`, `orchestrator`, `redis`, `kafka`, and `api-gateway` run via Docker Compose
- database and auth are expected from Supabase
- frontend runs separately with Vite

## 12. Design Tradeoffs

### Why separate API Gateway and Orchestrator

- the API Gateway handles client-facing concerns: auth, REST, WebSocket, ownership, rate limiting
- the Orchestrator handles long-running execution concerns: retries, cancellation, sequencing, progress
- this separation keeps the external API layer simpler and makes the execution pipeline independently evolvable

### Why Kafka for status events

- decouples pipeline execution from UI connection management
- supports future additional consumers like analytics, notifications, or audit streams
- prevents the Orchestrator from becoming browser-aware

### Why Postgres as source of truth

- scans, results, reports, and DLQ entries need durable persistence
- relational ownership constraints fit the per-user security model
- JSONB is used where report payloads need flexible structure

### Why Redis only for rate limiting

- Redis is used as a fast operational control plane, not as system-of-record
- durable domain state remains in Postgres

### Why Supabase

- gives Auth, Postgres, RLS, and object storage in one managed stack
- reduces glue code for early-stage product development

## 13. Current vs Target Architecture

```mermaid
flowchart LR
    A["Current Implementation"] --> B["Go gateway + Go orchestrator + HTTP agents + HTTP judge + Kafka + WS + Postgres"]
    C["Target Evolution"] --> D["Richer Python/CrewAI red-blue agents + stronger Judge benchmarking + more production-grade async workers"]
```

How to say this in an interview:

- "What is fully implemented is the service skeleton and the end-to-end control plane."
- "What is still evolving is the sophistication of the attack agents and the Judge intelligence layer."

That answer is accurate and technically honest.

## 14. Suggested Interview Talk Track

Use this order:

1. "AgentShield is a multi-service LLM security scanning platform."
2. "The user interacts only with the React dashboard and the Go API gateway."
3. "The gateway authenticates with Supabase JWT, persists scan metadata, and delegates execution over gRPC."
4. "The orchestrator runs the scan pipeline: call Agents, call Judge, persist findings, and publish progress."
5. "Kafka decouples execution events from UI delivery, and the gateway fans them out over WebSocket."
6. "Postgres is the source of truth, Redis is for rate limiting, and Supabase Storage holds the report artifacts."
7. "If execution fails, the orchestrator retries with backoff and persists a dead-letter record after exhaustion."
8. "The current platform is already end-to-end, while richer multi-agent intelligence is the next evolution."

## 15. Likely Interview Questions

### Why not let the frontend call the orchestrator directly?

Because the orchestrator is an internal execution service, not a trust boundary. Authentication, ownership, throttling, and client protocol stability belong in the API Gateway.

### Why use gRPC between gateway and orchestrator?

The interface is internal, latency-sensitive, and typed. gRPC is a good fit for service-to-service control calls like `StartScan`, `StopScan`, and status queries.

### Why not make everything synchronous?

Because scan execution is long-running and failure-prone. Decoupling execution, event streaming, and report generation makes the system more resilient and easier to observe.

### What happens if the orchestrator is down?

The gateway falls back to stub behavior and marks scans as `queued`, which preserves the user request and avoids hard failure at the API boundary.

### What makes this secure from a multi-tenant perspective?

Three layers:

- Supabase JWT authentication
- API ownership middleware
- Postgres Row Level Security

## 16. Source Pointers

- API gateway bootstrap: `api-gateway/cmd/server/main.go`
- scan lifecycle handler: `api-gateway/internal/handler/scan_handler.go`
- report generation: `api-gateway/internal/handler/report_generation_handler.go`
- WebSocket auth/upgrade: `api-gateway/internal/handler/ws_handler.go`
- Kafka dispatch: `api-gateway/internal/kafka/dispatcher.go`
- orchestrator server: `orchestrator/internal/orchestrator/server.go`
- orchestrator manager: `orchestrator/internal/orchestrator/manager.go`
- orchestrator executor: `orchestrator/internal/orchestrator/executor.go`
- agents service: `agents/internal/service/service.go`
- judge service: `judge/internal/service/service.go`
- scan domain/state machine: `api-gateway/internal/domain/scan.go`, `api-gateway/internal/domain/state_machine.go`
- DB schema: `api-gateway/migrations/001_create_scans.sql`, `004_create_attack_results.sql`, `005_create_reports.sql`, `008_create_judge_calibrations.sql`, `009_create_scan_dead_letters.sql`
