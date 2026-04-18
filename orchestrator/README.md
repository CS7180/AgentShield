# Orchestrator Service

Standalone gRPC orchestrator service for AgentShield.
It coordinates scan execution by calling the `agents` service and `judge` service,
then persists attack results/report aggregates into PostgreSQL.

## Endpoints

Implements `OrchestratorService` from `api-gateway/proto/orchestrator/orchestrator.proto`:

- `StartScan`
- `StopScan`
- `ScanStatus`

## Run

```bash
cd orchestrator
make run
```

Environment:

- `ORCHESTRATOR_PORT` (default: `50051`)
- `ENVIRONMENT` (`development` or `production`)
- `DATABASE_URL` (required for real pipeline execution)
- `KAFKA_BROKERS` (required for Kafka progress events, e.g. `kafka:9092`)
- `ORCHESTRATOR_SCAN_STATUS_TOPIC` (optional, default: `agent.status`)
- `ORCHESTRATOR_EXEC_MAX_ATTEMPTS` (optional, default: `3`)
- `ORCHESTRATOR_EXEC_RETRY_BASE_MS` (optional, default: `1000`)
- `ORCHESTRATOR_EXEC_RETRY_MAX_MS` (optional, default: `8000`)
- `AGENTS_SERVICE_URL` (default: `http://agents:8090`)
- `JUDGE_SERVICE_URL` (default: `http://judge:8091`)

Notes:

- The executor retries failed runs with exponential backoff.
- When retries are exhausted, a terminal failure record is persisted into `scan_dead_letters`.
- On each retry attempt, previous `attack_results` and `reports` for that scan are reset to keep data idempotent.
