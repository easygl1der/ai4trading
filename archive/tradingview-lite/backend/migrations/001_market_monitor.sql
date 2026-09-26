-- +goose Up

CREATE TABLE IF NOT EXISTS watchlist_items (
  symbol      TEXT PRIMARY KEY,
  group_name  TEXT NOT NULL DEFAULT 'custom',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ohlcv_bars (
  symbol      TEXT NOT NULL,
  timeframe   TEXT NOT NULL,
  provider    TEXT NOT NULL,
  ts          TIMESTAMPTZ NOT NULL,
  open        DOUBLE PRECISION NOT NULL,
  high        DOUBLE PRECISION NOT NULL,
  low         DOUBLE PRECISION NOT NULL,
  close       DOUBLE PRECISION NOT NULL,
  volume      DOUBLE PRECISION NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (symbol, timeframe, ts)
);
CREATE INDEX IF NOT EXISTS idx_ohlcv_bars_symbol_tf_ts
  ON ohlcv_bars(symbol, timeframe, ts DESC);

CREATE TABLE IF NOT EXISTS latest_quotes (
  symbol                TEXT PRIMARY KEY,
  provider              TEXT NOT NULL,
  price                 DOUBLE PRECISION NOT NULL,
  provider_time         TIMESTAMPTZ NOT NULL,
  received_at           TIMESTAMPTZ NOT NULL,
  data_age_seconds      BIGINT NOT NULL,
  exchange_name         TEXT,
  exchange_timezone     TEXT,
  data_granularity      TEXT,
  provider_warning      TEXT,
  payload_json          JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_configs (
  id                   TEXT PRIMARY KEY,
  symbol               TEXT NOT NULL,
  group_name           TEXT,
  rules_json           JSONB NOT NULL,
  notify_json          JSONB NOT NULL DEFAULT '[]',
  analysis_on_trigger  BOOLEAN NOT NULL DEFAULT FALSE,
  enabled              BOOLEAN NOT NULL DEFAULT TRUE,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_configs_symbol_enabled
  ON alert_configs(symbol, enabled);

CREATE TABLE IF NOT EXISTS alert_events (
  id                   TEXT PRIMARY KEY,
  alert_id             TEXT,
  symbol               TEXT NOT NULL,
  level                TEXT NOT NULL,
  rule_type            TEXT,
  message              TEXT NOT NULL,
  observed_value       DOUBLE PRECISION,
  threshold_value      DOUBLE PRECISION,
  benchmark            TEXT,
  data_age_seconds     BIGINT NOT NULL,
  provider_time        TIMESTAMPTZ NOT NULL,
  received_at          TIMESTAMPTZ NOT NULL,
  triggered_at         TIMESTAMPTZ NOT NULL,
  analysis_on_trigger  BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_alert_events_triggered_at
  ON alert_events(triggered_at DESC);

-- +goose Down

DROP TABLE IF EXISTS alert_events;
DROP TABLE IF EXISTS alert_configs;
DROP TABLE IF EXISTS latest_quotes;
DROP TABLE IF EXISTS ohlcv_bars;
DROP TABLE IF EXISTS watchlist_items;
