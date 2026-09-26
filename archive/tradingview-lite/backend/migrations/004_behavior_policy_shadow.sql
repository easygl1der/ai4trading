-- +goose Up

CREATE TABLE IF NOT EXISTS symbol_profiles (
  symbol                 TEXT PRIMARY KEY,
  role                   TEXT NOT NULL CHECK (role IN ('position', 'tactical_watch', 'research_watch', 'benchmark')),
  horizon                TEXT NOT NULL CHECK (horizon IN ('intraday', 'swing', 'medium_term', 'unspecified')),
  action_permissions_json JSONB NOT NULL DEFAULT '[]',
  cost_basis             DOUBLE PRECISION,
  notes                  TEXT,
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS peer_map_versions (
  id               BIGSERIAL PRIMARY KEY,
  map_key          TEXT NOT NULL,
  version          INT NOT NULL,
  group_name       TEXT,
  benchmark_symbol TEXT,
  methodology      TEXT NOT NULL,
  review_status    TEXT NOT NULL CHECK (review_status IN ('proposed', 'approved', 'retired')),
  effective_from   TIMESTAMPTZ NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(map_key, version)
);
CREATE INDEX IF NOT EXISTS idx_peer_map_versions_status_effective
  ON peer_map_versions(review_status, effective_from DESC);

CREATE TABLE IF NOT EXISTS peer_map_members (
  peer_map_version_id BIGINT NOT NULL REFERENCES peer_map_versions(id) ON DELETE CASCADE,
  symbol              TEXT NOT NULL,
  relation            TEXT NOT NULL,
  weight              DOUBLE PRECISION NOT NULL DEFAULT 1,
  PRIMARY KEY(peer_map_version_id, symbol)
);
CREATE INDEX IF NOT EXISTS idx_peer_map_members_symbol
  ON peer_map_members(symbol);

CREATE TABLE IF NOT EXISTS policy_candidate_alerts (
  id                   BIGSERIAL PRIMARY KEY,
  symbol               TEXT NOT NULL,
  trading_date         DATE NOT NULL,
  as_of                TIMESTAMPTZ NOT NULL,
  profile_role         TEXT NOT NULL,
  horizon              TEXT NOT NULL,
  session_product      TEXT NOT NULL,
  state                TEXT NOT NULL,
  action_candidate     TEXT NOT NULL,
  direction            TEXT,
  reference_price      DOUBLE PRECISION NOT NULL DEFAULT 0,
  eligible             BOOLEAN NOT NULL DEFAULT FALSE,
  would_notify         BOOLEAN NOT NULL DEFAULT FALSE,
  budget_slot          INT,
  delivery_mode        TEXT NOT NULL DEFAULT 'shadow_only' CHECK (delivery_mode = 'shadow_only'),
  suppressed_reason    TEXT,
  threshold_version    TEXT NOT NULL,
  event_context_id     BIGINT REFERENCES event_context_snapshots(id) ON DELETE SET NULL,
  peer_map_version_id  BIGINT REFERENCES peer_map_versions(id) ON DELETE SET NULL,
  reason_codes_json    JSONB NOT NULL DEFAULT '[]',
  metrics_json         JSONB NOT NULL DEFAULT '{}',
  warnings_json        JSONB NOT NULL DEFAULT '[]',
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_policy_candidate_alerts_symbol_asof
  ON policy_candidate_alerts(symbol, as_of DESC);
CREATE INDEX IF NOT EXISTS idx_policy_candidate_alerts_date_notify
  ON policy_candidate_alerts(trading_date, would_notify, as_of);

CREATE TABLE IF NOT EXISTS policy_candidate_outcomes (
  id                   BIGSERIAL PRIMARY KEY,
  candidate_alert_id   BIGINT NOT NULL REFERENCES policy_candidate_alerts(id) ON DELETE CASCADE,
  horizon_minutes      INT NOT NULL,
  target_at            TIMESTAMPTZ NOT NULL,
  observed_at          TIMESTAMPTZ NOT NULL,
  observed_price       DOUBLE PRECISION NOT NULL,
  return_percent       DOUBLE PRECISION NOT NULL,
  max_up_percent       DOUBLE PRECISION NOT NULL,
  max_down_percent     DOUBLE PRECISION NOT NULL,
  sample_count         INT NOT NULL,
  warning              TEXT,
  evaluated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(candidate_alert_id, horizon_minutes)
);
CREATE INDEX IF NOT EXISTS idx_policy_candidate_outcomes_evaluated
  ON policy_candidate_outcomes(evaluated_at DESC);

CREATE TABLE IF NOT EXISTS policy_candidate_feedback (
  candidate_alert_id BIGINT PRIMARY KEY REFERENCES policy_candidate_alerts(id) ON DELETE CASCADE,
  helpful            BOOLEAN,
  acted              BOOLEAN,
  emotion_intensity  INT CHECK (emotion_intensity BETWEEN 0 AND 5),
  notes              TEXT,
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS policy_candidate_feedback;
DROP TABLE IF EXISTS policy_candidate_outcomes;
DROP TABLE IF EXISTS policy_candidate_alerts;
DROP TABLE IF EXISTS peer_map_members;
DROP TABLE IF EXISTS peer_map_versions;
DROP TABLE IF EXISTS symbol_profiles;
