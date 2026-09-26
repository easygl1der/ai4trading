# TradingView-Lite Backend

Go/Fiber backend for the risk-aware market monitoring MVP.

It uses Yahoo chart API directly for the MVP data provider, stores finalized bars in a repository, tracks quote freshness, polls a watchlist, and evaluates first-pass alert rules. The backend does not place trades.

## Run Locally

```bash
APP_PORT=8088 \
APP_HOST=127.0.0.1 \
WATCHLIST_SYMBOLS=AAPL,NVDA,QQQ \
POLL_INTERVAL=60s \
DATA_STALE_AFTER=120s \
go run ./cmd/server
```

If `DATABASE_URL` is set, the backend uses Postgres and runs the embedded versioned Goose migrations at startup. Without `DATABASE_URL`, it uses an in-memory store for local testing only.

## Endpoints

```bash
curl http://127.0.0.1:8088/health
curl http://127.0.0.1:8088/api/v1/quote/AAPL
curl 'http://127.0.0.1:8088/api/v1/bars?symbol=AAPL&range=1d&interval=1m'
curl http://127.0.0.1:8088/api/v1/watchlist
curl -X POST http://127.0.0.1:8088/api/v1/poll
curl http://127.0.0.1:8088/api/v1/quotes
curl 'http://127.0.0.1:8088/api/v1/stored-bars?symbol=AAPL&limit=5'
curl http://127.0.0.1:8088/api/v1/events
curl http://127.0.0.1:8088/api/v1/rwa/NVDA/info
curl 'http://127.0.0.1:8088/api/v1/rwa/NVDA/kline?period=1m&size=60'
curl 'http://127.0.0.1:8088/api/v1/rwa/NVDA/summary?period=1m&size=60'
curl 'http://127.0.0.1:8088/api/v1/options/NVDA/chain?near=20'
curl http://127.0.0.1:8088/api/v1/options/NVDA/expected-move
curl http://127.0.0.1:8088/api/v1/features/NVDA/intraday
curl 'http://127.0.0.1:8088/api/v1/snapshots/quotes?symbol=NVDA&limit=100'
curl 'http://127.0.0.1:8088/api/v1/snapshots/rwa?symbol=NVDA&limit=100'
curl 'http://127.0.0.1:8088/api/v1/snapshots/rwa/bars?symbol=NVDA&interval=1m&limit=60'
curl 'http://127.0.0.1:8088/api/v1/snapshots/options/contracts?symbol=NVDA&limit=100'
curl 'http://127.0.0.1:8088/api/v1/snapshots/options/iv?symbol=NVDA&limit=100'
curl 'http://127.0.0.1:8088/api/v1/snapshots/features?symbol=NVDA&limit=100'
curl 'http://127.0.0.1:8088/api/v1/daily/NVDA?date=2026-07-13'
curl http://127.0.0.1:8088/api/v1/collection/runs
curl http://127.0.0.1:8088/api/v1/collection/storage
curl -X POST http://127.0.0.1:8088/api/v1/collection/market/run
curl 'http://127.0.0.1:8088/api/v1/event-evidence?symbol=NVDA&limit=100'
curl 'http://127.0.0.1:8088/api/v1/market-events?symbol=NVDA&limit=100'
curl http://127.0.0.1:8088/api/v1/event-context/NVDA
curl 'http://127.0.0.1:8088/api/v1/shadow-signals?symbol=NVDA&limit=100'
curl http://127.0.0.1:8088/api/v1/policy/profiles
curl 'http://127.0.0.1:8088/api/v1/policy/candidates?would_notify=true&limit=100'
curl 'http://127.0.0.1:8088/api/v1/policy/calibration/range/NVDA'
curl 'http://127.0.0.1:8088/api/v1/policy/calibration/iv/NVDA'
curl 'http://127.0.0.1:8088/api/v1/policy/calibration/rwa/NVDA'
curl http://127.0.0.1:8088/api/v1/event-ingestion/runs
curl http://127.0.0.1:8088/api/v1/event-ingestion/storage
curl -X POST http://127.0.0.1:8088/api/v1/collection/event-symbol/run
curl -X POST http://127.0.0.1:8088/api/v1/collection/event-macro/run
curl -X POST http://127.0.0.1:8088/api/v1/collection/event-context/run
curl -X POST http://127.0.0.1:8088/api/v1/collection/shadow/run
curl -X POST http://127.0.0.1:8088/api/v1/collection/behavior-policy/run
curl -X POST http://127.0.0.1:8088/api/v1/collection/policy-outcomes/run
```

Create a sample price alert:

```bash
curl -X POST http://127.0.0.1:8088/api/v1/alerts \
  -H 'Content-Type: application/json' \
  -d '{
    "symbol": "AAPL",
    "group": "semis_memory",
    "analysis_on_trigger": true,
    "rules": [
      {"type": "price_level", "operator": "above", "value": 1}
    ],
    "notify": ["log"]
  }'
```

Manually evaluate alerts:

```bash
curl -X POST http://127.0.0.1:8088/api/v1/alerts/evaluate
```

## Environment

| Variable | Default | Meaning |
|---|---:|---|
| `APP_HOST` | `127.0.0.1` | HTTP bind host. Use a reverse proxy instead of exposing the app directly. |
| `APP_PORT` | `8088` | HTTP port |
| `WATCHLIST_SYMBOLS` | built-in MVP list | Comma-separated symbols |
| `POLL_INTERVAL` | `60s` | Scheduler interval |
| `DATA_STALE_AFTER` | `120s` | Alert evaluator stale-data cutoff |
| `ENABLE_SCHEDULER` | `true` | Start background polling loop |
| `ENABLE_COLLECTORS` | `true` | Use provider-aware durable collectors instead of the legacy single poller |
| `COLLECTOR_MARKET_INTERVAL` | `60s` | Full-watchlist Yahoo quote and finalized `1m` bars cadence |
| `FOCUS_QUOTE_INTERVAL` | `30s` | Yahoo sampled-quote cadence for up to five `FOCUS_SYMBOLS` |
| `FOCUS_SYMBOLS` | empty | Optional high-attention symbols; sampled quotes are not `30s` bars |
| `RWA_COLLECT_INTERVAL` | `30s` | Baseline Bitget/Ondo RWA quote and `1m` bar cadence |
| `RWA_FOCUS_INTERVAL` | `15s` | Fast sampled RWA quote cadence for `RWA_FOCUS_SYMBOLS` |
| `RWA_SYMBOLS` | `NVDA,AMD` | Explicit verified RWA symbol subset; avoid probing every watchlist ticker |
| `RWA_FOCUS_SYMBOLS` | `NVDA` | Optional RWA fast lane, capped at five symbols |
| `OPTIONS_COLLECT_INTERVAL` | `5m` | Yahoo options-chain cadence for `OPTIONS_SYMBOLS` |
| `OPTIONS_SYMBOLS` | `MU,NVDA,AMD,SMH,SPY,QQQ` | Liquid options subset for near-ATM contract and IV snapshots |
| `FEATURE_COLLECT_INTERVAL` | `60s` | Feature snapshot cadence from the latest normalized provider inputs |
| `RETENTION_INTERVAL` | `24h` | Snapshot retention job cadence |
| `COLLECTOR_MAX_BACKOFF` | `10m` | Maximum exponential backoff after provider errors or `429` |
| `YAHOO_REQUESTS_PER_MINUTE` | `50` | Shared Yahoo chart request budget for baseline and focus quote sampling |
| `OPTIONS_REQUESTS_PER_MINUTE` | `20` | Yahoo options MCP request budget |
| `RWA_REQUESTS_PER_MINUTE` | `40` | Shared Bitget RWA request budget for baseline and focus sampling |
| `QUOTE_SNAPSHOT_RETENTION` | `336h` | Quote-snapshot retention (14 days) |
| `RWA_SNAPSHOT_RETENTION` | `336h` | RWA-snapshot retention (14 days) |
| `OPTION_SNAPSHOT_RETENTION` | `2160h` | Option-contract and IV retention (90 days) |
| `FEATURE_SNAPSHOT_RETENTION` | `2160h` | Feature-snapshot retention (90 days) |
| `COLLECTION_RUN_RETENTION` | `2160h` | Collector health-run retention (90 days) |
| `DATABASE_URL` | empty | Optional Postgres connection string |
| `DISCORD_WEBHOOK_WATCH` | empty | Discord webhook for `WATCH` |
| `DISCORD_WEBHOOK_ALERTS` | empty | Discord webhook for `ALERT` and `ANALYSIS_TRIGGER` |
| `DISCORD_WEBHOOK_URGENT` | empty | Discord webhook for `URGENT` |
| `DISCORD_WEBHOOK_DAILY` | empty | Reserved for daily briefs |
| `DISCORD_WEBHOOK_SYSTEM` | empty | Discord webhook for `BLOCKED` and system health |
| `BITGET_MCP_URL` | empty | Optional Bitget Wallet MCP HTTP endpoint, for example `http://127.0.0.1:8871/mcp` |
| `BITGET_MCP_TOKEN` | empty | Optional bearer token for Bitget Wallet MCP |
| `RWA_SYMBOL_MAP` | built-in liquid RWA map | Comma-separated mapping, for example `NVDA=NVDAon,TSLA=TSLAon` |
| `YAHOO_MCP_URL` | empty | Optional Yahoo Finance MCP HTTP endpoint, for example `http://127.0.0.1:3011/mcp` |
| `YAHOO_MCP_TOKEN` | empty | Optional bearer token for Yahoo Finance MCP |
| `OPTIONS_RISK_FREE_RATE` | `0.045` | Annualized decimal risk-free rate for Black-Scholes IV solving |

The event-evidence and shadow-signal settings are documented in [`.env.example`](./.env.example). `FIRECRAWL_API_KEY` is required only by the event collector and must remain outside version control.

Behavior-policy settings are also in [`.env.example`](./.env.example). `ENABLE_BEHAVIOR_POLICY_SHADOW=true` writes session-aware candidate states, a hard five-candidate daily budget, and post-hoc outcomes. It does not invoke the Discord notifier for market candidates. Symbol roles default to `research_watch` until the user explicitly classifies positions and tactical watch symbols.

## Bitget RWA Provider

The Bitget RWA provider is an optional proxy source for tokenized US equity exposure. It is useful for overnight and off-hours context, but it is not official Nasdaq/NYSE market data and it is not a real US equity Level 2 book.

Enable it by pointing the backend at an authenticated Bitget Wallet MCP endpoint:

```bash
BITGET_MCP_URL="http://127.0.0.1:8871/mcp" \
BITGET_MCP_TOKEN="$BITGET_MCP_TOKEN" \
go run ./cmd/server
```

Useful endpoints:

- `GET /api/v1/rwa/:symbol/info`: RWA stock info, sessions, latest price, and provider warning.
- `GET /api/v1/rwa/:symbol/kline?period=1m&size=60`: RWA K-line candles.
- `GET /api/v1/rwa/:symbol/summary`: stock info and recent RWA K-line in one response.

Default symbol mapping currently covers `AAPL`, `AMD`, `GOOGL`, `META`, `MSFT`, `NVDA`, and `TSLA`. Extend it with `RWA_SYMBOL_MAP` when more `*on` tickers are verified.

## Free Yahoo Options Provider

The Yahoo options provider is optional and uses the existing Yahoo Finance MCP. It is a free/unofficial exploration source for the options risk layer.

Enable it with:

```bash
YAHOO_MCP_URL="http://127.0.0.1:3011/mcp" \
YAHOO_MCP_TOKEN="$YAHOO_MCP_TOKEN" \
go run ./cmd/server
```

Useful endpoints:

- `GET /api/v1/options/:symbol/chain?near=20`: nearest-expiration option chain around the ATM strike.
- `GET /api/v1/options/:symbol/expected-move`: ATM straddle-implied move summary.

The response includes Yahoo raw `impliedVolatility`, but the backend marks it as advisory. For production-grade IV, store bid/ask/mid and re-solve IV from option prices through a controlled model.

The backend now also re-solves IV from Yahoo bid/ask midpoint using a Black-Scholes bisection solver:

- Input option price is `bid_ask_mid` when both bid and ask exist.
- If bid/ask are incomplete, it falls back to last price and marks `ivInputSource`.
- IV is annualized decimal, for example `0.25` means 25%.
- Time to expiry is calendar time to 16:00 America/New_York on the Yahoo expiration date.
- The model currently assumes zero dividend yield and uses `OPTIONS_RISK_FREE_RATE`.

This is a European approximation for a US equity option. It is good enough for a free MVP risk-range estimate, but should remain tagged as a model-derived estimate.

## Intraday Features

`GET /api/v1/features/:symbol/intraday` combines:

- Yahoo quote and `1m` bars.
- Yahoo options chain and resolved IV.
- Bitget/Ondo RWA proxy data when available.
- Realized volatility from `1m` log returns.
- Price speed over `1m`, `3m`, `5m`, and `15m`.
- RWA proxy speed over the same horizons when RWA K-line exists.
- Reasonable range from previous close, regular open, and current price.

Reasonable range selection is conservative: it uses the larger of option-derived one-day move and projected daily realized move.

## Discord Webhook Notifier

The Discord notifier is webhook-based. A Discord bot is not required.

Recommended first setup:

- `#market-alerts` -> `DISCORD_WEBHOOK_ALERTS`
- `#system-health` -> `DISCORD_WEBHOOK_SYSTEM`

Create the webhook in Discord:

1. Open your private Discord server.
2. Open the target channel, for example `#market-alerts`.
3. Channel Settings -> Integrations -> Webhooks.
4. Create a webhook named `Risk Agent Alerts`.
5. Copy the webhook URL.
6. Export it as an environment variable. Do not commit webhook URLs.

```bash
export DISCORD_WEBHOOK_ALERTS="https://discord.com/api/webhooks/..."
export DISCORD_WEBHOOK_SYSTEM="https://discord.com/api/webhooks/..."
```

Run with Discord enabled:

```bash
APP_PORT=8088 \
APP_HOST=127.0.0.1 \
WATCHLIST_SYMBOLS=AAPL,NVDA,QQQ \
DISCORD_WEBHOOK_ALERTS="$DISCORD_WEBHOOK_ALERTS" \
DISCORD_WEBHOOK_SYSTEM="$DISCORD_WEBHOOK_SYSTEM" \
go run ./cmd/server
```

Send a test message:

```bash
curl -X POST http://127.0.0.1:8088/api/v1/notifiers/discord/test
```

Discord messages are sent as embeds with mentions disabled. If Discord returns a `429`, the notifier honors `Retry-After` once and retries once. Notification failures are logged but do not stop polling or local event storage.

## Current Scope

- Uses Yahoo chart API directly, not Codex MCP, so the backend can run by itself.
- Returns quote-like intrabar points separately from finalized bars.
- Filters Yahoo's null/current incomplete rows out of the `bars` response.
- Includes `dataAgeSeconds` so alert logic can reject stale market data.
- Polls watchlist symbols and upserts finalized `1m` bars.
- Stores latest quotes and alert events.
- Supports price level, intraday move, relative underperformance, volume spike, and stale-data blocking rules.
- Supports memory mode for local tests and Postgres mode through `DATABASE_URL`.
- Optionally exposes Bitget/Ondo RWA stock info and K-line endpoints as an overnight proxy.
- Optionally exposes Yahoo Finance MCP options-chain and ATM straddle expected-move endpoints for a free prototype.
- Computes resolved IV, realized volatility, price speed, and MVP reasonable-range features.

## Durable Collection and Postgres

The database schema is versioned under [migrations](./migrations) and is applied through embedded Goose migrations when the process starts with `DATABASE_URL` configured. The VPS deployment uses the isolated `Postgres 16` compose service in [deploy/vps-hk/docker-compose.yml](../deploy/vps-hk/docker-compose.yml); it does not reuse Onyx's database.

The collector persists append-only Yahoo quote samples, RWA samples, selected near-ATM option contracts, resolved-IV summaries, derived feature states, and collector-run health. It separately upserts finalized Yahoo/RWA `1m` bars and daily market summaries.

History endpoints accept `from` and `to` in RFC3339 plus `limit`. Every quote sample preserves provider time, receipt time, data age, session, provider identity, and warnings. RWA data remains an Ondo/Bitget proxy and is never represented as official US-equity or Level 2 data.

Retention deletes only short-horizon snapshots; finalized `1m` bars and daily summaries are retained. `GET /api/v1/collection/storage` reports PostgreSQL table sizes and estimated row counts.

## Event Evidence and Shadow Signals

The Firecrawl-compatible news adapter is an event-evidence source, not price truth and not a trade-recommendation engine. The backend calls its VPS-local `/v1/news/search` endpoint with an environment-provided adapter key, then persists only short attributable metadata: canonical URL, headline hash, source domain, provider, publication/discovery/receipt times, ticker set, event type, and provider route.

Single-symbol searches use `strictTicker=true`. Market and group queries use `strictTicker=false` and are stored with `macro` or `group` scope, so they cannot silently appear as issuer-specific evidence. Source tiers are explicit:

- `T1_OFFICIAL`: configured official domains, `sec.gov`, or conventional investor-relations domains.
- `T2_PRIMARY`: government and primary macro sources.
- `T3_REPORTING`: reporting/context only.
- `T4_UNVERIFIED`: weak ticker match or incomplete provenance; retained for audit but not promoted.

`GET /api/v1/event-context/:symbol` returns the latest persisted point-in-time context. It filters evidence by `received_at <= as_of`, so historical evaluation cannot use a future-discovered article. A fresh market quote plus supported evidence can raise risk context; stale or missing market data produces `blocked`.

`IMPULSE_REJECTION` is shadow-only. It records `IDLE`, `NO_BASELINE`, `BLOCKED`, `IMPULSE_FORMING`, `CONTINUING`, or `REJECTED` together with the actual elapsed sample time, baseline sample count, range feature reference, and event-context ID. It does not call the notifier or produce a trading command. In particular, Yahoo weekend/stale samples correctly produce `BLOCKED` rather than an invented intraday signal.

The scheduler uses a distinct Firecrawl request budget and records every request, empty result, provider route, `429`, and error in `event_ingestion_runs`. Retention and storage diagnostics cover evidence, contexts, and shadow observations alongside existing market snapshots.

## Scope Boundary

The durable collector does not create normal Discord event/signal alerts, news/LLM trade interpretation, order-book claims, or automatic trading. Yahoo options remain a free/unofficial input and Black-Scholes results remain a model-derived risk-range estimate, not a direction signal.

## Behavior Policy Shadow Layer

The policy layer is deliberately separate from the existing alert evaluator. It records:

- explicit symbol profiles: `position`, `tactical_watch`, `research_watch`, or `benchmark`;
- versioned, reviewable peer/benchmark maps;
- overnight, premarket, opening, midday, and regular-session candidate states;
- range-use and RWA calibration diagnostics;
- `NO_CHASE`, `WAIT_CONFIRMATION`, `HOLD_OBSERVE`, `REDUCTION_REVIEW`, and `STOP_REVIEW` action candidates;
- later 15/30/60/180-minute excursions and optional user feedback.

Every policy candidate is database-constrained to `delivery_mode=shadow_only`. The Discord preview endpoint returns an embed-shaped preview and explicitly reports that no webhook was invoked. Normal market Discord delivery requires a later policy-review change; it is not enabled by this layer.

Useful endpoints:

```bash
curl -X PUT http://127.0.0.1:8088/api/v1/policy/profiles/NVDA \
  -H 'Content-Type: application/json' \
  -d '{"role":"research_watch","horizon":"intraday","actionPermissions":["WAIT_CONFIRMATION","NO_CHASE"]}'

curl -X POST http://127.0.0.1:8088/api/v1/policy/NVDA/evaluate
curl 'http://127.0.0.1:8088/api/v1/policy/candidates?symbol=NVDA&limit=20'
curl 'http://127.0.0.1:8088/api/v1/policy/candidates/1/discord-preview'
curl -X POST http://127.0.0.1:8088/api/v1/policy/NVDA/outcomes/evaluate
curl -X PUT http://127.0.0.1:8088/api/v1/policy/candidates/1/feedback \
  -H 'Content-Type: application/json' \
  -d '{"helpful":true,"acted":false,"emotionIntensity":1,"notes":"减少了追涨冲动"}'
```
