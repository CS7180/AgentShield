-- Migration: 001_create_scans
-- Creates the scans table with RLS policies for Supabase.

CREATE TABLE IF NOT EXISTS scans (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID         NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  target_endpoint VARCHAR(500) NOT NULL,
  mode            VARCHAR(20)  NOT NULL CHECK (mode IN ('red_team', 'blue_team', 'adversarial')),
  attack_types    TEXT[]       NOT NULL,
  status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'queued', 'running', 'completed', 'failed', 'stopped')),
  created_at      TIMESTAMPTZ  DEFAULT NOW(),
  started_at      TIMESTAMPTZ,
  completed_at    TIMESTAMPTZ,
  updated_at      TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scans_user   ON scans(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_scans_status ON scans(status);

-- Row Level Security: users can only access their own scans
ALTER TABLE scans ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies WHERE tablename = 'scans' AND policyname = 'users_own_scans'
  ) THEN
    CREATE POLICY "users_own_scans" ON scans
      USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies WHERE tablename = 'scans' AND policyname = 'users_insert_own_scans'
  ) THEN
    CREATE POLICY "users_insert_own_scans" ON scans
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;
