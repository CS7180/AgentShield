-- Migration: 007_report_storage_integration
-- Integrates reports table with Supabase Storage bucket lifecycle.
-- Includes:
--   1) report object path columns
--   2) path ownership constraints
--   3) storage bucket bootstrap (when storage schema exists)
--   4) storage object cleanup on report delete/update

ALTER TABLE reports
  ADD COLUMN IF NOT EXISTS report_bucket VARCHAR(100) NOT NULL DEFAULT 'agentshield-reports',
  ADD COLUMN IF NOT EXISTS report_json_path VARCHAR(500);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_reports_pdf_path_owned_by_user'
  ) THEN
    ALTER TABLE reports
      ADD CONSTRAINT chk_reports_pdf_path_owned_by_user
      CHECK (
        report_pdf_path IS NULL
        OR report_pdf_path = ''
        OR report_pdf_path LIKE user_id::text || '/%'
      );
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_reports_json_path_owned_by_user'
  ) THEN
    ALTER TABLE reports
      ADD CONSTRAINT chk_reports_json_path_owned_by_user
      CHECK (
        report_json_path IS NULL
        OR report_json_path = ''
        OR report_json_path LIKE user_id::text || '/%'
      );
  END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_reports_bucket_pdf_path
  ON reports(report_bucket, report_pdf_path);
CREATE INDEX IF NOT EXISTS idx_reports_bucket_json_path
  ON reports(report_bucket, report_json_path);

-- Bootstrap bucket + policies only when Supabase storage schema exists.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
    EXECUTE $sql$
      INSERT INTO storage.buckets (id, name, public)
      VALUES ('agentshield-reports', 'agentshield-reports', false)
      ON CONFLICT (id) DO NOTHING
    $sql$;
  END IF;
END$$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
    IF NOT EXISTS (
      SELECT 1
      FROM pg_policies
      WHERE schemaname = 'storage'
        AND tablename = 'objects'
        AND policyname = 'agentshield_reports_read'
    ) THEN
      EXECUTE $sql$
        CREATE POLICY "agentshield_reports_read" ON storage.objects
          FOR SELECT TO authenticated
          USING (
            bucket_id = 'agentshield-reports'
            AND (storage.foldername(name))[1] = auth.uid()::text
          )
      $sql$;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM pg_policies
      WHERE schemaname = 'storage'
        AND tablename = 'objects'
        AND policyname = 'agentshield_reports_insert'
    ) THEN
      EXECUTE $sql$
        CREATE POLICY "agentshield_reports_insert" ON storage.objects
          FOR INSERT TO authenticated
          WITH CHECK (
            bucket_id = 'agentshield-reports'
            AND (storage.foldername(name))[1] = auth.uid()::text
          )
      $sql$;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM pg_policies
      WHERE schemaname = 'storage'
        AND tablename = 'objects'
        AND policyname = 'agentshield_reports_update'
    ) THEN
      EXECUTE $sql$
        CREATE POLICY "agentshield_reports_update" ON storage.objects
          FOR UPDATE TO authenticated
          USING (
            bucket_id = 'agentshield-reports'
            AND (storage.foldername(name))[1] = auth.uid()::text
          )
          WITH CHECK (
            bucket_id = 'agentshield-reports'
            AND (storage.foldername(name))[1] = auth.uid()::text
          )
      $sql$;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM pg_policies
      WHERE schemaname = 'storage'
        AND tablename = 'objects'
        AND policyname = 'agentshield_reports_delete'
    ) THEN
      EXECUTE $sql$
        CREATE POLICY "agentshield_reports_delete" ON storage.objects
          FOR DELETE TO authenticated
          USING (
            bucket_id = 'agentshield-reports'
            AND (storage.foldername(name))[1] = auth.uid()::text
          )
      $sql$;
    END IF;
  END IF;
END$$;

CREATE OR REPLACE FUNCTION public.delete_storage_object_if_exists(p_bucket TEXT, p_name TEXT)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  IF p_name IS NULL OR p_name = '' THEN
    RETURN;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
    EXECUTE 'DELETE FROM storage.objects WHERE bucket_id = $1 AND name = $2'
      USING p_bucket, p_name;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION public.cleanup_report_storage_objects()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    PERFORM public.delete_storage_object_if_exists(OLD.report_bucket, OLD.report_pdf_path);
    PERFORM public.delete_storage_object_if_exists(OLD.report_bucket, OLD.report_json_path);
    RETURN OLD;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    IF OLD.report_bucket IS DISTINCT FROM NEW.report_bucket
       OR OLD.report_pdf_path IS DISTINCT FROM NEW.report_pdf_path THEN
      PERFORM public.delete_storage_object_if_exists(OLD.report_bucket, OLD.report_pdf_path);
    END IF;

    IF OLD.report_bucket IS DISTINCT FROM NEW.report_bucket
       OR OLD.report_json_path IS DISTINCT FROM NEW.report_json_path THEN
      PERFORM public.delete_storage_object_if_exists(OLD.report_bucket, OLD.report_json_path);
    END IF;
    RETURN NEW;
  END IF;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_reports_cleanup_storage_on_delete'
  ) THEN
    CREATE TRIGGER trg_reports_cleanup_storage_on_delete
      AFTER DELETE ON reports
      FOR EACH ROW
      EXECUTE FUNCTION public.cleanup_report_storage_objects();
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_reports_cleanup_storage_on_update'
  ) THEN
    CREATE TRIGGER trg_reports_cleanup_storage_on_update
      AFTER UPDATE OF report_bucket, report_pdf_path, report_json_path ON reports
      FOR EACH ROW
      EXECUTE FUNCTION public.cleanup_report_storage_objects();
  END IF;
END$$;
