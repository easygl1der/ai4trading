-- +goose Up

ALTER TABLE event_evidence
  ADD COLUMN IF NOT EXISTS source_event_id TEXT,
  ADD COLUMN IF NOT EXISTS body_hash TEXT,
  ADD COLUMN IF NOT EXISTS confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS data_age_seconds BIGINT,
  ADD COLUMN IF NOT EXISTS raw_payload_location TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_event_evidence_source_event
  ON event_evidence(provider, source_event_id)
  WHERE source_event_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS event_evidence_versions (
  id                  BIGSERIAL PRIMARY KEY,
  evidence_id         BIGINT NOT NULL REFERENCES event_evidence(id) ON DELETE CASCADE,
  version_number      INT NOT NULL,
  body_hash           TEXT NOT NULL,
  body_markdown       TEXT,
  raw_payload_json    JSONB NOT NULL DEFAULT '{}',
  fetched_at          TIMESTAMPTZ NOT NULL,
  fetch_status        TEXT NOT NULL DEFAULT 'ok',
  warning             TEXT,
  UNIQUE(evidence_id, version_number),
  UNIQUE(evidence_id, body_hash)
);
CREATE INDEX IF NOT EXISTS idx_event_evidence_versions_evidence
  ON event_evidence_versions(evidence_id, fetched_at DESC);

CREATE TABLE IF NOT EXISTS official_source_registry (
  id                  BIGSERIAL PRIMARY KEY,
  symbol              TEXT NOT NULL,
  source_name         TEXT NOT NULL,
  source_tier         TEXT NOT NULL DEFAULT 'T1_OFFICIAL',
  domain              TEXT NOT NULL,
  rss_url             TEXT,
  news_url            TEXT,
  allowed_paths       JSONB NOT NULL DEFAULT '[]',
  enabled             BOOLEAN NOT NULL DEFAULT TRUE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(symbol, domain)
);

CREATE TABLE IF NOT EXISTS sec_filings (
  accession_number    TEXT PRIMARY KEY,
  symbol              TEXT NOT NULL,
  cik                 TEXT NOT NULL,
  form_type           TEXT NOT NULL,
  filing_at           TIMESTAMPTZ NOT NULL,
  primary_document    TEXT,
  canonical_url       TEXT NOT NULL,
  summary             TEXT,
  raw_payload_json    JSONB NOT NULL DEFAULT '{}',
  discovered_at       TIMESTAMPTZ NOT NULL,
  received_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sec_filings_symbol_filing
  ON sec_filings(symbol, filing_at DESC);

CREATE TABLE IF NOT EXISTS discord_delivery_outbox (
  id                  BIGSERIAL PRIMARY KEY,
  idempotency_key     TEXT NOT NULL UNIQUE,
  alert_event_id      TEXT REFERENCES alert_events(id) ON DELETE CASCADE,
  level               TEXT NOT NULL,
  webhook_channel     TEXT NOT NULL,
  payload_json        JSONB NOT NULL,
  status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'delivered', 'dead_letter', 'suppressed')),
  attempts            INT NOT NULL DEFAULT 0,
  next_attempt_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  locked_at           TIMESTAMPTZ,
  delivered_at        TIMESTAMPTZ,
  last_error          TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_discord_delivery_outbox_pending
  ON discord_delivery_outbox(status, next_attempt_at, created_at)
  WHERE status IN ('pending', 'processing');

-- +goose Down

DROP TABLE IF EXISTS discord_delivery_outbox;
DROP TABLE IF EXISTS sec_filings;
DROP TABLE IF EXISTS official_source_registry;
DROP TABLE IF EXISTS event_evidence_versions;
DROP INDEX IF EXISTS idx_event_evidence_source_event;
ALTER TABLE event_evidence
  DROP COLUMN IF EXISTS raw_payload_location,
  DROP COLUMN IF EXISTS data_age_seconds,
  DROP COLUMN IF EXISTS confidence,
  DROP COLUMN IF EXISTS body_hash,
  DROP COLUMN IF EXISTS source_event_id;
