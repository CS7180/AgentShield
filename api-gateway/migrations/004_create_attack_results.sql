-- Migration: 004_create_attack_results
-- PRD-aligned attack results persisted per scan with ownership-safe constraints.

CREATE TABLE IF NOT EXISTS attack_results (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  scan_id              UUID         NOT NULL,
  user_id              UUID         NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  attack_type          VARCHAR(50)  NOT NULL
                       CHECK (attack_type IN ('prompt_injection', 'jailbreak', 'data_leakage', 'constraint_drift')),
  attack_prompt        TEXT         NOT NULL,
  target_response      TEXT         NOT NULL,
  attack_success       BOOLEAN      NOT NULL,
  severity             VARCHAR(20)
                       CHECK (severity IN ('critical', 'high', 'medium', 'low')),
  owasp_category       VARCHAR(50),
  defense_intercepted  BOOLEAN,
  judge_confidence     DOUBLE PRECISION
                       CHECK (judge_confidence IS NULL OR (judge_confidence >= 0 AND judge_confidence <= 1)),
  judge_reasoning      TEXT,
  latency_ms           INTEGER
                       CHECK (latency_ms IS NULL OR latency_ms >= 0),
  tokens_used          INTEGER
                       CHECK (tokens_used IS NULL OR tokens_used >= 0),
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_attack_results_scan_user
    FOREIGN KEY (scan_id, user_id)
    REFERENCES scans(id, user_id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attack_results_scan_created
  ON attack_results(scan_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attack_results_user_created
  ON attack_results(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attack_results_attack_type
  ON attack_results(attack_type);
CREATE INDEX IF NOT EXISTS idx_attack_results_severity
  ON attack_results(severity);

ALTER TABLE attack_results ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'attack_results' AND policyname = 'users_own_attack_results_select'
  ) THEN
    CREATE POLICY "users_own_attack_results_select" ON attack_results
      FOR SELECT USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'attack_results' AND policyname = 'users_own_attack_results_insert'
  ) THEN
    CREATE POLICY "users_own_attack_results_insert" ON attack_results
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'attack_results' AND policyname = 'users_own_attack_results_update'
  ) THEN
    CREATE POLICY "users_own_attack_results_update" ON attack_results
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
    WHERE tablename = 'attack_results' AND policyname = 'users_own_attack_results_delete'
  ) THEN
    CREATE POLICY "users_own_attack_results_delete" ON attack_results
      FOR DELETE USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_attack_results_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_attack_results_set_updated_at
      BEFORE UPDATE ON attack_results
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
