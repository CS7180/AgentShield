-- Migration: 006_create_golden_set
-- Judge calibration dataset. Supports private rows and optional shared rows.

CREATE TABLE IF NOT EXISTS golden_set (
  id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                  UUID         NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  attack_type              VARCHAR(50)  NOT NULL
                           CHECK (attack_type IN ('prompt_injection', 'jailbreak', 'data_leakage', 'constraint_drift')),
  attack_prompt            TEXT         NOT NULL,
  target_response          TEXT         NOT NULL,
  expected_success         BOOLEAN      NOT NULL,
  expected_severity        VARCHAR(20)
                           CHECK (expected_severity IS NULL OR expected_severity IN ('critical', 'high', 'medium', 'low')),
  expected_owasp_category  VARCHAR(50),
  notes                    TEXT,
  is_shared                BOOLEAN      NOT NULL DEFAULT false,
  created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_golden_set_user ON golden_set(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_golden_set_attack_type ON golden_set(attack_type);
CREATE INDEX IF NOT EXISTS idx_golden_set_shared ON golden_set(is_shared) WHERE is_shared = true;

ALTER TABLE golden_set ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'golden_set' AND policyname = 'users_read_own_or_shared_golden_set'
  ) THEN
    CREATE POLICY "users_read_own_or_shared_golden_set" ON golden_set
      FOR SELECT USING (user_id = auth.uid() OR is_shared = true);
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'golden_set' AND policyname = 'users_insert_own_golden_set'
  ) THEN
    CREATE POLICY "users_insert_own_golden_set" ON golden_set
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'golden_set' AND policyname = 'users_update_own_golden_set'
  ) THEN
    CREATE POLICY "users_update_own_golden_set" ON golden_set
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
    WHERE tablename = 'golden_set' AND policyname = 'users_delete_own_golden_set'
  ) THEN
    CREATE POLICY "users_delete_own_golden_set" ON golden_set
      FOR DELETE USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_golden_set_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_golden_set_set_updated_at
      BEFORE UPDATE ON golden_set
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
