-- Migration: 009_create_scan_dead_letters
-- Persist final failed scan execution attempts (dead-letter queue semantics).

CREATE TABLE IF NOT EXISTS scan_dead_letters (
  id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  scan_id        UUID         NOT NULL,
  user_id        UUID         NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  attempt_count  INTEGER      NOT NULL CHECK (attempt_count > 0),
  error_stage    VARCHAR(120) NOT NULL,
  error_message  TEXT         NOT NULL,
  payload        JSONB        NOT NULL DEFAULT '{}'::jsonb,
  failed_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_scan_dead_letters_scan_user
    FOREIGN KEY (scan_id, user_id)
    REFERENCES scans(id, user_id)
    ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_scan_dead_letters_scan_unique
  ON scan_dead_letters(scan_id);
CREATE INDEX IF NOT EXISTS idx_scan_dead_letters_user_failed
  ON scan_dead_letters(user_id, failed_at DESC);

ALTER TABLE scan_dead_letters ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'scan_dead_letters' AND policyname = 'users_own_scan_dead_letters_select'
  ) THEN
    CREATE POLICY "users_own_scan_dead_letters_select" ON scan_dead_letters
      FOR SELECT USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'scan_dead_letters' AND policyname = 'users_own_scan_dead_letters_insert'
  ) THEN
    CREATE POLICY "users_own_scan_dead_letters_insert" ON scan_dead_letters
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'scan_dead_letters' AND policyname = 'users_own_scan_dead_letters_update'
  ) THEN
    CREATE POLICY "users_own_scan_dead_letters_update" ON scan_dead_letters
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
    WHERE tablename = 'scan_dead_letters' AND policyname = 'users_own_scan_dead_letters_delete'
  ) THEN
    CREATE POLICY "users_own_scan_dead_letters_delete" ON scan_dead_letters
      FOR DELETE USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_scan_dead_letters_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_scan_dead_letters_set_updated_at
      BEFORE UPDATE ON scan_dead_letters
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
