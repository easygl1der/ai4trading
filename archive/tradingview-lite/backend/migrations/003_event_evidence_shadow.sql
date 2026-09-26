-- +goose Up

CREATE TABLE IF NOT EXISTS event_ingestion_runs (
  id                BIGSERIAL PRIMARY KEY,
  provider          TEXT NOT NULL,
  query             TEXT NOT NULL,
  scope             TEXT NOT NULL,
  group_name        TEXT,
  requested_symbols JSONB NOT NULL DEFAULT '[]',
  strict_ticker     BOOLEAN NOT NULL DEFAULT FALSE,
  provider_route    TEXT,
  started_at        TIMESTAMPTZ NOT NULL,
  finished_at       TIMESTAMPTZ NOT NULL,
  result_count      INT NOT NULL DEFAULT 0,
  records_written   INT NOT NULL DEFAULT 0,
  rate_limited      BOOLEAN NOT NULL DEFAULT FALSE,
  backoff_seconds   BIGINT NOT NULL DEFAULT 0,
  error_summary     TEXT,
  metadata_json     JSONB NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_event_ingestion_runs_finished
  ON event_ingestion_runs(finished_at DESC);

CREATE TABLE IF NOT EXISTS event_evidence (
  id                BIGSERIAL PRIMARY KEY,
  ingestion_run_id  BIGINT REFERENCES event_ingestion_runs(id) ON DELETE SET NULL,
  canonical_url     TEXT NOT NULL UNIQUE,
  headline_hash     TEXT NOT NULL,
  title             TEXT NOT NULL,
  snippet           TEXT,
  source_domain     TEXT,
  provider          TEXT NOT NULL,
  publisher         TEXT,
  source_tier       TEXT NOT NULL,
  published_at      TIMESTAMPTZ,
  discovered_at     TIMESTAMPTZ,
  received_at       TIMESTAMPTZ NOT NULL,
  freshness         TEXT,
  matched_symbols   JSONB NOT NULL DEFAULT '[]',
  scope             TEXT NOT NULL,
  group_name        TEXT,
  event_type        TEXT NOT NULL,
  metadata_json     JSONB NOT NULL DEFAULT '{}',
  warning           TEXT
);
CREATE INDEX IF NOT EXISTS idx_event_evidence_symbols_received
  ON event_evidence(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_event_evidence_scope_received
  ON event_evidence(scope, received_at DESC);

CREATE TABLE IF NOT EXISTS market_events (
  id                BIGSERIAL PRIMARY KEY,
  event_key         TEXT NOT NULL UNIQUE,
  scope             TEXT NOT NULL,
  primary_symbol    TEXT,
  group_name        TEXT,
  event_type        TEXT NOT NULL,
  evidence_status   TEXT NOT NULL,
  severity          TEXT NOT NULL DEFAULT 'unknown',
  state             TEXT NOT NULL DEFAULT 'active',
  first_seen_at     TIMESTAMPTZ NOT NULL,
  last_seen_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_market_events_symbol_seen
  ON market_events(primary_symbol, last_seen_at DESC);

CREATE TABLE IF NOT EXISTS market_event_evidence (
  market_event_id BIGINT NOT NULL REFERENCES market_events(id) ON DELETE CASCADE,
  evidence_id     BIGINT NOT NULL REFERENCES event_evidence(id) ON DELETE CASCADE,
  relation_type   TEXT NOT NULL DEFAULT 'supports',
  confidence      DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY(market_event_id, evidence_id)
);

CREATE TABLE IF NOT EXISTS event_context_snapshots (
  id                    BIGSERIAL PRIMARY KEY,
  symbol                TEXT NOT NULL,
  as_of                 TIMESTAMPTZ NOT NULL,
  market_data_fresh     BOOLEAN NOT NULL,
  quote_age_seconds     BIGINT NOT NULL,
  session               TEXT NOT NULL,
  selected_move_pct     DOUBLE PRECISION NOT NULL DEFAULT 0,
  evidence_coverage     TEXT NOT NULL,
  last_ingestion_at     TIMESTAMPTZ,
  known_event_ids_json  JSONB NOT NULL DEFAULT '[]',
  risk_state            TEXT NOT NULL,
  reason_codes_json     JSONB NOT NULL DEFAULT '[]',
  warnings_json         JSONB NOT NULL DEFAULT '[]',
  payload_json          JSONB NOT NULL DEFAULT '{}',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_event_context_symbol_asof
  ON event_context_snapshots(symbol, as_of DESC);

CREATE TABLE IF NOT EXISTS shadow_signal_observations (
  id                  BIGSERIAL PRIMARY KEY,
  signal_type         TEXT NOT NULL,
  symbol              TEXT NOT NULL,
  as_of               TIMESTAMPTZ NOT NULL,
  state               TEXT NOT NULL,
  eligible            BOOLEAN NOT NULL,
  blocked_reason      TEXT,
  threshold_version   TEXT NOT NULL,
  event_context_id    BIGINT REFERENCES event_context_snapshots(id) ON DELETE SET NULL,
  feature_snapshot_id BIGINT REFERENCES intraday_feature_snapshots(id) ON DELETE SET NULL,
  reason_codes_json   JSONB NOT NULL DEFAULT '[]',
  metrics_json        JSONB NOT NULL DEFAULT '{}',
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_shadow_signal_symbol_asof
  ON shadow_signal_observations(symbol, as_of DESC);

-- +goose Down

DROP TABLE IF EXISTS shadow_signal_observations;
DROP TABLE IF EXISTS event_context_snapshots;
DROP TABLE IF EXISTS market_event_evidence;
DROP TABLE IF EXISTS market_events;
DROP TABLE IF EXISTS event_evidence;
DROP TABLE IF EXISTS event_ingestion_runs;
