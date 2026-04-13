# Agents Service

Standalone HTTP service that simulates red-team attack execution.

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
