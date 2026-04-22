## Tech Stack

| Layer | Technology |
|-------|-----------|
| API Gateway | Go 1.22, Gin, pgx/v5, go-redis/v9, Sarama, gorilla/websocket |
| Orchestrator | Go 1.22, gRPC server |
| Agent frameworks | Python 3.11+, CrewAI |
| Inter-service | gRPC + Protobuf |
| Message queue | Kafka (topics: `attack.results`, `defense.results`, `judge.evaluations`, `agent.status`) |
| Auth | Supabase Auth (frontend → Supabase directly; gateway validates Supabase-issued JWT) |
| Database | Supabase PostgreSQL (pgx/v5, Transaction Pooler port 6543) |
| Cache | Redis 7 |
| LLM models | Gemini 2.5 Flash (generation), Gemini 2.5 Pro (Judge) |
| Frontend | React 18, Vite, Recharts, TailwindCSS |
| Containers | Docker, Kubernetes, Helm |
| CI/CD | GitHub Actions |
| Infrastructure | Terraform + AWS EKS |
| Monitoring | Prometheus + Grafana |

---

## System Architecture

```
React Dashboard  ──── Supabase Auth (login/register/refresh — frontend only)
       │
       │  JWT (Authorization: Bearer)          WS ?token=<jwt>
       ▼
Go API Gateway  (JWT validation, rate limiting, SSRF guard, Prometheus)
       │
       │  gRPC
       ▼
Go Orchestrator  (task scheduling, agent lifecycle, result aggregation)
       │  gRPC task distribution
  ┌────┴───────────────────────────┐
  ▼                                ▼
Red Team Agents (Python/CrewAI)   Blue Team Agents (Python/CrewAI)
  Prompt Injection                  Input Guard
  Jailbreak                         Output Filter
  Data Leakage                      Behavior Monitor
  Constraint Drift                  Constraint Persistence
  └──────────────┬─────────────────┘
                 │  Kafka
                 ▼
         LLM-as-Judge (Python)
                 │
         Result Aggregator (Go)
                 │
  Report Generator (JSON + PDF, OWASP LLM Top 10)
                 │
          PostgreSQL + Redis
```

**Auth flow:** Frontend authenticates directly with Supabase SDK. The gateway has **no auth endpoints**. Every API call carries `Authorization: Bearer <supabase_jwt>`. The gateway validates the JWT using the Supabase JWKS endpoint. WebSocket upgrades use `?token=<jwt>` because browsers cannot set headers on WS connections.

---

## Repository Structure & Ownership

```
AgentShield/
├── api-gateway/          # Go — Shanshou Li
│   ├── cmd/server/       # Binary entry point
│   ├── internal/         # All internal packages (not importable externally)
│   ├── proto/            # .proto source + generated stubs
│   └── migrations/       # SQL migrations (applied on startup)
├── orchestrator/         # Go — Shanshou Li
├── agents/               # Python/CrewAI — Shanshou Li
│   ├── red_team/
│   └── blue_team/
├── judge/                # Python — Shanshou Li
├── frontend/             # React — Yachen Wang
├── docs/                 # Sprint docs, standup logs, architecture notes
├── project_memory/       # PRD, UI mockups, reference docs (read-only)
└── CLAUDE.md
```

**Ownership rules:**
- Do not modify `project_memory/` files — they are reference artifacts.
- Do not create files outside the service directory being worked on unless explicitly asked.
- The `api-gateway/internal/` boundary is strict: no package inside it may be imported by another service.
