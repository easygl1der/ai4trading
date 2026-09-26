-- +goose Up

ALTER TABLE ohlcv_bars
  ADD COLUMN IF NOT EXISTS session TEXT NOT NULL DEFAULT 'offhours';

CREATE TABLE IF NOT EXISTS market_quote_snapshots (
  id               BIGSERIAL PRIMARY KEY,
  symbol           TEXT NOT NULL,
  provider         TEXT NOT NULL,
  price            DOUBLE PRECISION NOT NULL,
  provider_time    TIMESTAMPTZ NOT NULL,
  received_at      TIMESTAMPTZ NOT NULL,
  data_age_seconds BIGINT NOT NULL,
  session          TEXT NOT NULL,
  market_state     TEXT,
  warning          TEXT
);
CREATE INDEX IF NOT EXISTS idx_market_quote_snapshots_symbol_provider_received
  ON market_quote_snapshots(symbol, provider, received_at DESC);

CREATE TABLE IF NOT EXISTS rwa_quote_snapshots (
  id                     BIGSERIAL PRIMARY KEY,
  symbol                 TEXT NOT NULL,
  ticker                 TEXT NOT NULL,
  provider               TEXT NOT NULL,
  data_source            TEXT,
  price                  DOUBLE PRECISION NOT NULL,
  received_at            TIMESTAMPTZ NOT NULL,
  session                TEXT NOT NULL,
  provider_market_status TEXT,
  warning                TEXT
);
CREATE INDEX IF NOT EXISTS idx_rwa_quote_snapshots_symbol_provider_received
  ON rwa_quote_snapshots(symbol, provider, received_at DESC);

CREATE TABLE IF NOT EXISTS rwa_ohlcv_bars (
  symbol      TEXT NOT NULL,
  ticker      TEXT NOT NULL,
  provider    TEXT NOT NULL,
  timeframe   TEXT NOT NULL,
  ts          TIMESTAMPTZ NOT NULL,
  open        DOUBLE PRECISION NOT NULL,
  high        DOUBLE PRECISION NOT NULL,
  low         DOUBLE PRECISION NOT NULL,
  close       DOUBLE PRECISION NOT NULL,
  volume      DOUBLE PRECISION NOT NULL,
  amount      DOUBLE PRECISION NOT NULL,
  received_at TIMESTAMPTZ NOT NULL,
  session     TEXT NOT NULL,
  PRIMARY KEY(symbol, provider, timeframe, ts)
);
CREATE INDEX IF NOT EXISTS idx_rwa_ohlcv_bars_symbol_tf_ts
  ON rwa_ohlcv_bars(symbol, timeframe, ts DESC);

CREATE TABLE IF NOT EXISTS option_contract_snapshots (
  id                          BIGSERIAL PRIMARY KEY,
  symbol                      TEXT NOT NULL,
  provider                    TEXT NOT NULL,
  contract_symbol             TEXT NOT NULL,
  expiration_date             TIMESTAMPTZ NOT NULL,
  strike                      DOUBLE PRECISION NOT NULL,
  right_type                  TEXT NOT NULL,
  bid                         DOUBLE PRECISION NOT NULL,
  ask                         DOUBLE PRECISION NOT NULL,
  mid                         DOUBLE PRECISION NOT NULL,
  last_price                  DOUBLE PRECISION NOT NULL,
  volume                      BIGINT NOT NULL,
  open_interest               BIGINT NOT NULL,
  raw_implied_volatility      DOUBLE PRECISION NOT NULL,
  resolved_implied_volatility DOUBLE PRECISION NOT NULL,
  resolved_iv_error           TEXT,
  iv_input_price              DOUBLE PRECISION NOT NULL,
  iv_input_source             TEXT,
  contract_trade_at           TIMESTAMPTZ,
  underlying_spot             DOUBLE PRECISION NOT NULL,
  underlying_provider_time    TIMESTAMPTZ,
  received_at                 TIMESTAMPTZ NOT NULL,
  warning                     TEXT
);
CREATE INDEX IF NOT EXISTS idx_option_contract_snapshots_symbol_exp_received
  ON option_contract_snapshots(symbol, expiration_date, received_at DESC);

CREATE TABLE IF NOT EXISTS iv_snapshots (
  id                               BIGSERIAL PRIMARY KEY,
  symbol                           TEXT NOT NULL,
  provider                         TEXT NOT NULL,
  expiration_date                  TIMESTAMPTZ NOT NULL,
  atm_strike                       DOUBLE PRECISION NOT NULL,
  spot                             DOUBLE PRECISION NOT NULL,
  call_resolved_iv                 DOUBLE PRECISION NOT NULL,
  put_resolved_iv                  DOUBLE PRECISION NOT NULL,
  average_resolved_iv              DOUBLE PRECISION NOT NULL,
  straddle_move                    DOUBLE PRECISION NOT NULL,
  straddle_move_percent            DOUBLE PRECISION NOT NULL,
  one_day_move                     DOUBLE PRECISION NOT NULL,
  one_day_move_percent             DOUBLE PRECISION NOT NULL,
  risk_free_rate                   DOUBLE PRECISION NOT NULL,
  dividend_yield_assumption        DOUBLE PRECISION NOT NULL,
  underlying_provider_time         TIMESTAMPTZ,
  underlying_data_age_seconds      BIGINT NOT NULL,
  contract_trade_data_age_seconds  BIGINT NOT NULL,
  received_at                      TIMESTAMPTZ NOT NULL,
  warning                          TEXT
);
CREATE INDEX IF NOT EXISTS idx_iv_snapshots_symbol_received
  ON iv_snapshots(symbol, received_at DESC);

CREATE TABLE IF NOT EXISTS intraday_feature_snapshots (
  id                   BIGSERIAL PRIMARY KEY,
  symbol               TEXT NOT NULL,
  computed_at          TIMESTAMPTZ NOT NULL,
  provider_times_json  JSONB NOT NULL,
  input_freshness_json JSONB NOT NULL,
  selected_source      TEXT,
  selected_move_pct    DOUBLE PRECISION NOT NULL,
  warnings_json        JSONB NOT NULL,
  payload_json         JSONB NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intraday_feature_snapshots_symbol_computed
  ON intraday_feature_snapshots(symbol, computed_at DESC);

CREATE TABLE IF NOT EXISTS collection_runs (
  id                BIGSERIAL PRIMARY KEY,
  collector         TEXT NOT NULL,
  provider          TEXT NOT NULL,
  started_at        TIMESTAMPTZ NOT NULL,
  finished_at       TIMESTAMPTZ NOT NULL,
  symbols_attempted INT NOT NULL,
  symbols_succeeded INT NOT NULL,
  records_written   INT NOT NULL,
  rate_limited      BOOLEAN NOT NULL DEFAULT FALSE,
  backoff_seconds   BIGINT NOT NULL DEFAULT 0,
  error_summary     TEXT
);
CREATE INDEX IF NOT EXISTS idx_collection_runs_collector_finished
  ON collection_runs(collector, finished_at DESC);

CREATE TABLE IF NOT EXISTS daily_market_summaries (
  symbol               TEXT NOT NULL,
  trading_date         DATE NOT NULL,
  provider             TEXT NOT NULL,
  previous_close       DOUBLE PRECISION NOT NULL DEFAULT 0,
  premarket_high       DOUBLE PRECISION NOT NULL DEFAULT 0,
  premarket_low        DOUBLE PRECISION NOT NULL DEFAULT 0,
  preopen_five_minute  DOUBLE PRECISION NOT NULL DEFAULT 0,
  regular_open         DOUBLE PRECISION NOT NULL DEFAULT 0,
  daily_high           DOUBLE PRECISION NOT NULL DEFAULT 0,
  daily_low            DOUBLE PRECISION NOT NULL DEFAULT 0,
  selected_move        DOUBLE PRECISION NOT NULL DEFAULT 0,
  selected_move_pct    DOUBLE PRECISION NOT NULL DEFAULT 0,
  selected_source      TEXT,
  updated_at           TIMESTAMPTZ NOT NULL,
  PRIMARY KEY(symbol, trading_date, provider)
);

-- +goose Down

DROP TABLE IF EXISTS daily_market_summaries;
DROP TABLE IF EXISTS collection_runs;
DROP TABLE IF EXISTS intraday_feature_snapshots;
DROP TABLE IF EXISTS iv_snapshots;
DROP TABLE IF EXISTS option_contract_snapshots;
DROP TABLE IF EXISTS rwa_ohlcv_bars;
DROP TABLE IF EXISTS rwa_quote_snapshots;
DROP TABLE IF EXISTS market_quote_snapshots;
ALTER TABLE ohlcv_bars DROP COLUMN IF EXISTS session;
