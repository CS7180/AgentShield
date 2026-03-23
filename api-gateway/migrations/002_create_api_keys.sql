-- Migration: 002_create_api_keys
-- API keys for service-to-service auth (e.g. Jenkins CI).

CREATE TABLE IF NOT EXISTS api_keys (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID         NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  key_hash    VARCHAR(255) UNIQUE NOT NULL,
  name        VARCHAR(100),
  created_at  TIMESTAMPTZ  DEFAULT NOW(),
  last_used_at TIMESTAMPTZ,
  expires_at  TIMESTAMPTZ,
  active      BOOLEAN      DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user   ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(key_hash) WHERE active = true;
