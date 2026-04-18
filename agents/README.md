# Agents Service

Standalone HTTP service for red-team attack execution.

It supports two execution modes:

- `simulate` (default): deterministic local simulation
- `target_http`: sends attack payloads to the scan target endpoint and evaluates response heuristically

## Endpoints

- `GET /health`
- `POST /run-scan`

## Run

```bash
cd agents
make run
```

Environment:

- `AGENTS_PORT` (default `8090`)
- `AGENTS_EXECUTION_MODE` (`simulate` or `target_http`, default `simulate`)
