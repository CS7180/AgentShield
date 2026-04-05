-- Migration: 003_harden_rls_and_triggers
-- Adds shared trigger utilities, hardens api_keys with RLS, and prepares
-- a composite key on scans for child tables that enforce user ownership.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Utility trigger function to keep updated_at fresh on write.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

-- Ensure api_keys has updated_at for future-safe writes.
ALTER TABLE api_keys
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- A composite uniqueness key allows child tables to reference (scan_id, user_id)
-- and guarantee ownership consistency.
CREATE UNIQUE INDEX IF NOT EXISTS idx_scans_id_user_unique ON scans(id, user_id);

ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'api_keys' AND policyname = 'users_own_api_keys_select'
  ) THEN
    CREATE POLICY "users_own_api_keys_select" ON api_keys
      FOR SELECT USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'api_keys' AND policyname = 'users_own_api_keys_insert'
  ) THEN
    CREATE POLICY "users_own_api_keys_insert" ON api_keys
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'api_keys' AND policyname = 'users_own_api_keys_update'
  ) THEN
    CREATE POLICY "users_own_api_keys_update" ON api_keys
      FOR UPDATE
      USING (user_id = auth.uid())
      WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'api_keys' AND policyname = 'users_own_api_keys_delete'
  ) THEN
    CREATE POLICY "users_own_api_keys_delete" ON api_keys
      FOR DELETE USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_scans_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_scans_set_updated_at
      BEFORE UPDATE ON scans
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_api_keys_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_api_keys_set_updated_at
      BEFORE UPDATE ON api_keys
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
