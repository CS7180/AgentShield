# Supabase Schema Setup (AgentShield)

This folder now contains PRD-aligned, idempotent SQL migrations for:

- `scans`
- `api_keys`
- `attack_results`
- `reports`
- `golden_set`
- `report storage integration` (Supabase Storage bucket + cleanup triggers)

## Option A: Let API Gateway Apply Migrations

1. Configure `DATABASE_URL` to your Supabase Postgres connection.
2. Start the gateway from `api-gateway/`.
3. On startup, `RunMigrations(...)` executes all `.sql` files in this folder in lexical order.

## Option B: Execute Manually in Supabase SQL Editor

Run these files in order:

1. `001_create_scans.sql`
2. `002_create_api_keys.sql`
3. `003_harden_rls_and_triggers.sql`
4. `004_create_attack_results.sql`
5. `005_create_reports.sql`
6. `006_create_golden_set.sql`
7. `007_report_storage_integration.sql`

All migrations are written with `IF NOT EXISTS`/guard blocks and are safe to re-run.

## Quick Verification Queries

```sql
-- Tables
select tablename
from pg_tables
where schemaname = 'public'
  and tablename in ('scans', 'api_keys', 'attack_results', 'reports', 'golden_set')
order by tablename;

-- RLS enabled
select tablename, rowsecurity
from pg_tables
where schemaname = 'public'
  and tablename in ('scans', 'api_keys', 'attack_results', 'reports', 'golden_set')
order by tablename;

-- Policies
select tablename, policyname, cmd
from pg_policies
where schemaname = 'public'
  and tablename in ('scans', 'api_keys', 'attack_results', 'reports', 'golden_set')
order by tablename, policyname;
```

## Notes

- Schema is aligned to `project_memory/AgentShield_PRD.md` and `CLAUDE.md` constraints.
- User-owned tables enforce RLS with `user_id = auth.uid()` style policies.
- Child tables (`attack_results`, `reports`) enforce ownership consistency with composite foreign keys to `scans(id, user_id)`.
- `007_report_storage_integration.sql` bootstraps bucket `agentshield-reports` and adds triggers that clean up `storage.objects` when report rows are deleted (including via scan-level cascade).
- Expected object path convention is `<user_id>/scans/<scan_id>/...` so storage RLS and `reports` table constraints remain consistent.
