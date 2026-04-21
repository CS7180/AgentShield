## REST API Reference

```
# No auth endpoints — handled entirely by Supabase frontend SDK

GET  /health
GET  /metrics                              (Prometheus)

POST /api/v1/scans                         (JWTAuth + ScanCreateRateLimit)
GET  /api/v1/scans                         (JWTAuth)
GET  /api/v1/scans/:id                     (JWTAuth + Ownership)
POST /api/v1/scans/:id/start               (JWTAuth + Ownership)
POST /api/v1/scans/:id/stop                (JWTAuth + Ownership)
POST /api/v1/scans/:id/attack-results      (JWTAuth + Ownership)
GET  /api/v1/scans/:id/attack-results      (JWTAuth + Ownership)
GET  /api/v1/scans/:id/dead-letters        (JWTAuth + Ownership)
PUT  /api/v1/scans/:id/report              (JWTAuth + Ownership)
GET  /api/v1/scans/:id/report              (JWTAuth + Ownership)
GET  /api/v1/scans/:id/report/pdf          (JWTAuth + Ownership) → 501 until Sprint 2
POST /api/v1/scans/:id/report/generate     (JWTAuth + Ownership)
GET  /api/v1/scans/:id/compare/:other_id   (JWTAuth + Ownership)
POST /api/v1/judge/calibrate               (JWTAuth)
GET  /api/v1/judge/calibration-report      (JWTAuth)

WS   /ws/scans/:id/status                  (?token=<supabase_jwt>)
```

**Error envelope** (all errors):
```json
{ "error": "...", "code": "SNAKE_CASE", "status_code": 400, "timestamp": "...", "request_id": "..." }
```

---

## Database Schema

Tables: `scans`, `attack_results`, `reports`, `scan_dead_letters`, `judge_calibrations` (in Supabase PostgreSQL; `auth.users` provided by Supabase).

Key `attack_results` fields: `attack_type`, `attack_prompt`, `target_response`, `attack_success`, `severity` (critical/high/medium/low), `owasp_category`, `defense_intercepted`, `judge_confidence`, `latency_ms`, `tokens_used`.

Migrations live in `api-gateway/migrations/` and are applied automatically on gateway startup. Use `IF NOT EXISTS` and idempotent guards. Never use `DROP` or destructive DDL in migrations.

**RLS:** Every table that stores user data must have `ENABLE ROW LEVEL SECURITY` and a `user_id = auth.uid()` policy. The gateway also enforces ownership in the `Ownership` middleware as a defence-in-depth layer.

---

## gRPC / Protobuf Contract

```protobuf
// proto/orchestrator/orchestrator.proto
service OrchestratorService {
  rpc StartScan  (StartScanRequest)  returns (StartScanResponse);
  rpc StopScan   (StopScanRequest)   returns (StopScanResponse);
  rpc ScanStatus (ScanStatusRequest) returns (ScanStatusResponse);
}
```

**Proto rules:**
- Never change field numbers on existing messages — add new fields only.
- Always regenerate Go stubs with `make proto` after editing `.proto` files.
- Python agents use `grpcio-tools` to generate stubs: `python -m grpc_tools.protoc ...`
- The `.proto` source is the single source of truth. Generated files are derived artifacts.
