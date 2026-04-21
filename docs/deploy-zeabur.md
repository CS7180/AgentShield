# Deploy AgentShield on Zeabur

This repository is a monorepo with multiple deployable services:

- `frontend`: React + Vite dashboard
- `api-gateway`: Go HTTP API
- `orchestrator`: Go gRPC service
- `agents`: Go HTTP execution service
- `judge`: Go HTTP evaluation service

Zeabur does not use the local `docker-compose.yml` directly for production deployment.
Deploy each service separately inside the same Zeabur project and connect them over Zeabur private networking.

## Recommended deployment paths

### Path A: minimal usable deployment

Deploy these first:

- `frontend`
- `api-gateway`
- `Redis`

Use Supabase for:

- Auth
- PostgreSQL
- Storage

Set `ORCHESTRATOR_ENABLED=false` on `api-gateway`.

This mode is the easiest to get online, but scans will stay in queued/stub mode and will not execute end-to-end.

### Path B: full platform deployment

Deploy these inside one Zeabur project:

- `frontend`
- `api-gateway`
- `orchestrator`
- `agents`
- `judge`
- `Redis`
- `Kafka`

Keep Supabase for:

- Auth
- PostgreSQL
- Storage

This mode enables real scan execution and WebSocket status propagation.

## Monorepo Dockerfiles for Zeabur

The repository root includes Zeabur-friendly Dockerfiles:

- `Dockerfile.frontend`
- `Dockerfile.api-gateway`
- `Dockerfile.orchestrator`
- `Dockerfile.agents`
- `Dockerfile.judge`

For each Zeabur service created from this repo:

1. Set the repository root as the source directory.
2. Set `ZBPACK_DOCKERFILE_NAME` to the matching service name.

Examples:

- `frontend` service: `ZBPACK_DOCKERFILE_NAME=frontend`
- `api-gateway` service: `ZBPACK_DOCKERFILE_NAME=api-gateway`
- `orchestrator` service: `ZBPACK_DOCKERFILE_NAME=orchestrator`
- `agents` service: `ZBPACK_DOCKERFILE_NAME=agents`
- `judge` service: `ZBPACK_DOCKERFILE_NAME=judge`

## Create the Zeabur project

Create one project in Zeabur, then add services:

1. Add Git services from this repository:
   - `frontend`
   - `api-gateway`
   - optional: `orchestrator`, `agents`, `judge`
2. Add managed services:
   - `Redis`
   - `Kafka` (recommended via Zeabur template)

If you only want the minimal deployment, skip `orchestrator`, `agents`, `judge`, and `Kafka`.

## Environment variables

### `frontend`

Build-time variables:

- `VITE_SUPABASE_URL=https://<your-project-ref>.supabase.co`
- `VITE_SUPABASE_ANON_KEY=<your-supabase-anon-key>`
- `VITE_API_URL=https://<your-api-gateway-domain>`

Notes:

- `VITE_API_URL` must be the public HTTPS URL of the deployed `api-gateway`.
- The frontend Dockerfile reads these at build time.

### `api-gateway`

Required:

- `ENVIRONMENT=production`
- `SUPABASE_JWT_SECRET=<supabase-jwt-secret>`
- `SUPABASE_URL=https://<your-project-ref>.supabase.co`
- `SUPABASE_SERVICE_ROLE_KEY=<supabase-service-role-key>`
- `DATABASE_URL=<supabase-postgres-connection-string>`
- `REDIS_URL=<zeabur-redis-connection-string>`
- `ALLOWED_ORIGINS=https://<your-frontend-domain>`

Recommended:

- `SERVER_PORT=8080`
- `SUPABASE_REPORTS_BUCKET=agentshield-reports`
- `KAFKA_GROUP_ID=api-gateway-ws`

Minimal deployment:

- `ORCHESTRATOR_ENABLED=false`

Full deployment:

- `ORCHESTRATOR_ENABLED=true`
- `ORCHESTRATOR_ADDR=<orchestrator-private-host>:50051`
- `KAFKA_BROKERS=<kafka-private-host>:<kafka-private-port>`

### `orchestrator`

Required:

- `ENVIRONMENT=production`
- `ORCHESTRATOR_PORT=50051`
- `DATABASE_URL=<supabase-postgres-connection-string>`
- `KAFKA_BROKERS=<kafka-private-host>:<kafka-private-port>`
- `AGENTS_SERVICE_URL=http://<agents-private-host>:8090`
- `JUDGE_SERVICE_URL=http://<judge-private-host>:8091`

Optional:

- `ORCHESTRATOR_SCAN_STATUS_TOPIC=agent.status`
- `ORCHESTRATOR_EXEC_MAX_ATTEMPTS=3`
- `ORCHESTRATOR_EXEC_RETRY_BASE_MS=1000`
- `ORCHESTRATOR_EXEC_RETRY_MAX_MS=8000`

### `agents`

Required:

- `AGENTS_PORT=8090`

Optional:

- `AGENTS_EXECUTION_MODE=simulate`

If you want it to hit a real target endpoint instead of local simulation:

- `AGENTS_EXECUTION_MODE=target_http`

### `judge`

Required:

- `JUDGE_PORT=8091`

Local rule mode:

- `JUDGE_EVAL_MODE=rule`

Model-based mode:

- `JUDGE_EVAL_MODE=openai_compat`
- `JUDGE_LLM_BASE_URL=<compatible-chat-completions-base-url>`
- `JUDGE_LLM_API_KEY=<api-key>`
- `JUDGE_LLM_MODEL=<model-name>`

## Private networking

Use Zeabur private networking values from the service networking panel for:

- `REDIS_URL`
- `KAFKA_BROKERS`
- `ORCHESTRATOR_ADDR`
- `AGENTS_SERVICE_URL`
- `JUDGE_SERVICE_URL`

Do not use public domains for internal service-to-service traffic unless you have a specific reason to.

## Deployment order

### Minimal deployment order

1. `Redis`
2. `api-gateway`
3. `frontend`

### Full deployment order

1. `Redis`
2. `Kafka`
3. `agents`
4. `judge`
5. `orchestrator`
6. `api-gateway`
7. `frontend`

## Post-deploy checks

Check these endpoints after deployment:

- `GET https://<api-gateway-domain>/health`
- Open the frontend and confirm login works.
- Create a scan and confirm the API accepts it.

For full deployment, also verify:

- `orchestrator` is reachable from `api-gateway`
- `agents` and `judge` are reachable from `orchestrator`
- WebSocket status updates arrive in the dashboard

## Important caveats

- `api-gateway` requires Redis at startup; it cannot run without a reachable `REDIS_URL`.
- `api-gateway` requires `DATABASE_URL` and `SUPABASE_JWT_SECRET`.
- With `ORCHESTRATOR_ENABLED=false`, the system is online but not executing real scans end-to-end.
- The frontend is a SPA; `Dockerfile.frontend` uses Nginx with an `index.html` fallback for React Router refreshes.
