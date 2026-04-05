-- Migration: 005_create_reports
-- Aggregated scan report output and scorecard storage.

CREATE TABLE IF NOT EXISTS reports (
  id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  scan_id          UUID          NOT NULL,
  user_id          UUID          NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  overall_score    NUMERIC(5,2)
                   CHECK (overall_score IS NULL OR (overall_score >= 0 AND overall_score <= 100)),
  critical_count   INTEGER       NOT NULL DEFAULT 0 CHECK (critical_count >= 0),
  high_count       INTEGER       NOT NULL DEFAULT 0 CHECK (high_count >= 0),
  medium_count     INTEGER       NOT NULL DEFAULT 0 CHECK (medium_count >= 0),
  low_count        INTEGER       NOT NULL DEFAULT 0 CHECK (low_count >= 0),
  owasp_scorecard  JSONB         NOT NULL DEFAULT '{}'::jsonb,
  report_json      JSONB         NOT NULL DEFAULT '{}'::jsonb,
  report_pdf_path  VARCHAR(500),
  created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_reports_scan_user
    FOREIGN KEY (scan_id, user_id)
    REFERENCES scans(id, user_id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reports_scan ON reports(scan_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_reports_scan_unique ON reports(scan_id);
CREATE INDEX IF NOT EXISTS idx_reports_user_created ON reports(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reports_report_json_gin ON reports USING GIN (report_json);

ALTER TABLE reports ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'reports' AND policyname = 'users_own_reports_select'
  ) THEN
    CREATE POLICY "users_own_reports_select" ON reports
      FOR SELECT USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'reports' AND policyname = 'users_own_reports_insert'
  ) THEN
    CREATE POLICY "users_own_reports_insert" ON reports
      FOR INSERT WITH CHECK (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_policies
    WHERE tablename = 'reports' AND policyname = 'users_own_reports_update'
  ) THEN
    CREATE POLICY "users_own_reports_update" ON reports
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
    WHERE tablename = 'reports' AND policyname = 'users_own_reports_delete'
  ) THEN
    CREATE POLICY "users_own_reports_delete" ON reports
      FOR DELETE USING (user_id = auth.uid());
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_reports_set_updated_at'
  ) THEN
    CREATE TRIGGER trg_reports_set_updated_at
      BEFORE UPDATE ON reports
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;
