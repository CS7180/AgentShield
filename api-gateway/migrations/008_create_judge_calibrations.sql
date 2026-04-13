-- Migration: 008_create_judge_calibrations
-- Persists latest calibration report per user.

CREATE TABLE IF NOT EXISTS judge_calibrations (
  user_id       UUID        PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  sample_count  INTEGER     NOT NULL DEFAULT 0 CHECK (sample_count >= 0),
  report_json   JSONB       NOT NULL DEFAULT '{}'::jsonb,
  generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE judge_calibrations ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'judge_calibrations' AND policyname = 'users_own_judge_calibrations_select'
  ) THEN
    CREATE POLICY "users_own_judge_calibrations_select" ON judge_calibrations
      FOR SELECT USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'judge_calibrations' AND policyname = 'users_own_judge_calibrations_insert'
  ) THEN
    CREATE POLICY "users_own_judge_calibrations_insert" ON judge_calibrations
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'judge_calibrations' AND policyname = 'users_own_judge_calibrations_update'
  ) THEN
    CREATE POLICY "users_own_judge_calibrations_update" ON judge_calibrations
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
    WHERE tablename = 'judge_calibrations' AND policyname = 'users_own_judge_calibrations_delete'
  ) THEN
    CREATE POLICY "users_own_judge_calibrations_delete" ON judge_calibrations
      FOR DELETE USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_judge_calibrations_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_judge_calibrations_set_updated_at
      BEFORE UPDATE ON judge_calibrations
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
