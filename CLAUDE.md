# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**AgentShield** is a multi-agent red-blue teaming platform that automatically discovers security vulnerabilities in LLM-powered applications. It simulates adversarial attacks and defenses running simultaneously against a target LLM system to identify vulnerabilities before deployment.

The full product requirements are in `project_memory/AgentShield_PRD.md`. React UI mockups (dark-themed, purple accent) are in `project_memory/AgentShield_*.jsx`.

## Planned Tech Stack

| Layer | Technology |
|-------|-----------|
| API Gateway + Orchestrator | Go |
| Agent frameworks | Python + CrewAI |
| Inter-service communication | gRPC + Protobuf |
| Message queue | Kafka |
| Database | PostgreSQL |
| Cache / sessions | Redis |
| LLM models | Gemini 2.5 Flash (generation), Gemini 2.5 Pro (Judge) |
| Frontend | React |
| Container orchestration | Docker + Kubernetes + Helm |
| CI/CD | Jenkins + GitLab CI |
| Infrastructure | Terraform + AWS EKS |
| Monitoring | Prometheus + Grafana |

## System Architecture

```
React Dashboard (WebSocket for real-time)
        ↓
Go API Gateway  (auth, rate limiting)
        ↓
Go Orchestrator (task scheduling, agent lifecycle, result aggregation)
        ↓  gRPC task distribution
  ┌─────┴──────────────────────┐
  ↓                            ↓
Red Team Agents (Python/CrewAI)   Blue Team Agents (Python/CrewAI)
  - Prompt Injection Agent         - Input Guard Agent
  - Jailbreak Agent                - Output Filter Agent
  - Data Leakage Agent             - Behavior Monitor Agent
  - Constraint Drift Agent         - Constraint Persistence Agent
  └──────────────┬──────────────┘
                 ↓ Kafka (attack.results, defense.results, judge.evaluations, agent.status)
         LLM-as-Judge (Python)
                 ↓
     Result Aggregator (Go)
                 ↓
  Report Generator (JSON + PDF, OWASP LLM Top 10 aligned)
                 ↓
        PostgreSQL + Redis
```

**Three scan modes:**
- **Red Team Mode** — Attacks only against baseline (no defense)
- **Blue Team Mode** — Measures false positive rate on normal traffic
- **Adversarial Mode** — Red + blue simultaneous with comparative analysis

## gRPC Service

```protobuf
service AgentService {
  rpc ExecuteAttack (AttackRequest) returns (AttackResponse);
  rpc ExecuteDefense (DefenseRequest) returns (DefenseResponse);
  rpc EvaluateResult (JudgeRequest) returns (JudgeResponse);
  rpc HealthCheck (Empty) returns (HealthStatus);
}
```

**Kafka topics:** `attack.results`, `defense.results`, `judge.evaluations`, `agent.status`

## REST API (Go API Gateway)

```
POST /api/v1/auth/register|login|refresh
POST /api/v1/scans                     # create scan
GET  /api/v1/scans/:id
POST /api/v1/scans/:id/start|stop
GET  /api/v1/scans/:id/report          # JSON
GET  /api/v1/scans/:id/report/pdf
GET  /api/v1/scans/:id/compare/:other_id
WS   /ws/scans/:id/status              # real-time WebSocket
POST /api/v1/judge/calibrate
```

## Database (PostgreSQL)

Tables: `users`, `scans`, `attack_results`, `reports`, `golden_set`

Key fields in `attack_results`: `attack_type`, `attack_prompt`, `target_response`, `attack_success`, `severity` (critical/high/medium/low), `owasp_category`, `defense_intercepted`, `judge_confidence`, `latency_ms`, `tokens_used`

## Testing Strategy

- Unit tests: ~70% coverage on attack generation, response parsing, Judge evaluation, orchestration, defense agents
- Integration tests: gRPC chain, Kafka flow, end-to-end scans
- Evaluation tests: 50+ golden labeled attack/response pairs for LLM-as-Judge calibration
- Target: 80%+ code coverage

## Sprint Plan

- **Sprint 1 (Weeks 1-2):** Go API Gateway + Orchestrator skeleton, gRPC/Protobuf, Kafka, PostgreSQL/Redis, 3 Red Team Agents, LLM-as-Judge (30 golden pairs), React skeleton, Jenkins/Docker CI → **Red team mode end-to-end**
- **Sprint 2 (Weeks 3-4):** Constraint Drift Agent, 4 Blue Team Agents, Adversarial mode, Result Aggregator, OWASP scorecard + PDF reports, WebSocket dashboard, Judge calibration (50+ pairs), K8s/Helm/Terraform → **Full platform, all 3 modes**

## Team

- **Shanshou Li:** Go Orchestrator, gRPC, Kafka, architecture, red/blue team agents, LLM-as-Judge, Jenkins CI/CD, K8s deployment
- **Yachen Wang:** React dashboard, frontend visualization, WebSocket integration, PDF report generation
