# News And Alerts Audit

Audited: 2026-07-20 UTC. This document records the deployed state observed on `vps-hk` and separates it from the intended next-stage design. No secret values are recorded.

## Executive Summary

The production backend is healthy and collecting Yahoo quotes/bars, Yahoo options, Bitget RWA proxy data, and Firecrawl news evidence into PostgreSQL. Discord is configured and its existing test endpoint is available. The system is not yet an evidence-first alert system: the durable collector scheduler does not invoke the legacy price-alert evaluator, evidence contains search results but not scraped article bodies, and Discord delivery is synchronous rather than durable.

## Observed Production State

| Component | Status | Actual configuration or observation | Assessment |
| --- | --- | --- | --- |
| Backend service | Enabled | `tradingview-lite-backend.service` active since 2026-07-13; `/health` returned `ok: true` | Available |
| Store | Enabled | Dedicated `tradingview-lite-postgres` container healthy; 18 latest quotes, 4,621 evidence rows, 16,836 ingestion runs at audit time | Available |
| Scheduler | Enabled | Collector scheduler enabled, market collection every 60s | Available |
| Yahoo quote/bar | Enabled | Full watchlist quote + 1m bar collection every 60s; focus quotes every 30s | MVP only, not exchange-grade realtime |
| Yahoo options | Enabled | Selected symbols every 5m; 20 requests/minute budget | Exploration/risk-range input only |
| Bitget RWA | Enabled | General polling every 30s, focus polling every 15s; 40 requests/minute budget | Auxiliary RWA proxy, not US equity reference/NBBO/L2 |
| Firecrawl evidence | Enabled | Symbol news every 15m, macro/group news every 10m, 10 requests/minute, search window 1,440m | Search evidence only; not prompt enough for major-news alerts |
| Event context/shadow policy | Enabled | Context every 1m, shadow signals every 30s, policy candidate budget 5/day | Shadow only; no market candidate is delivered to Discord |
| Legacy alert evaluator | Present but not scheduled | `alert_events=0`; collector scheduler never calls the evaluator | Missing production alert path |
| Discord | Enabled | Webhooks configured; test endpoint exists | Direct send only; not durable |

## Firecrawl Adapter: Actual Capability Audit

The VPS quality gateway was healthy and its upstream adapter was reachable with 11 tools. It exposes `news_search`, `scrape`, `batch_scrape`, `map`, `crawl`, `extract`, and `answer`.

Safe read-only probes performed during this audit:

| Probe | Result | Interpretation |
| --- | --- | --- |
| `news_search` for `NVDA latest news` | Returned current candidates, including items discovered in the same minute; `cacheHit=false`; one upstream provider was circuit-open | Useful for near-realtime candidate discovery, but provider availability is partial and `discoveredAt` is not necessarily article publication time |
| `scrape` of IANA public page | HTTP 200 with Markdown and discovered links | Known static pages can be fetched and archived |
| `scrape` of Federal Reserve page | HTTP 403 with empty content | Official sites can block the adapter; this must be a recorded degradation, not silently treated as no news |
| `map` of two public sites | Returned an empty link list | Tool is exposed but is not currently reliable enough to use as a primary source-discovery mechanism |
| `scrape` with `renderJs=true` | Explicitly rejected: JS rendering unavailable | Dynamic pages require RSS/API/direct provider alternatives |

This is the deployed adapter's behavior. It is narrower than Firecrawl Cloud documentation: no verified cache-control input, no JS rendering, and no successful site-map discovery were demonstrated in this audit. The backend currently calls only the local `news/search` HTTP endpoint and does not call `scrape`, `map`, or `crawl`.

## Current Evidence Model

`event_evidence` persists canonical URL, title/hash, snippet, source domain, provider, source tier, published/discovered/received timestamps, matched symbols, event type, metadata, and warnings. It deduplicates by canonical URL. This is a useful base, but it lacks body hash, full-text/version history, source-specific identifier, CIK/accession fields, data-age field, explicit confidence, and raw-payload location.

The current source-tier classifier recognizes official domains but uses generic tiers for most reporting. It does not yet have a source registry, SEC/EDGAR ingestion, RSS-first official sources, or versioned body storage.

## Current Options And Risk Range

The option collector records raw Yahoo `impliedVolatility`, a Black-Scholes bisection `resolvedImpliedVolatility`, bid/ask/mid/last, volume, open interest, underlying timestamp, and contract last-trade time. The current chain selection uses the first chain returned by Yahoo rather than explicit `0DTE_or_nearest_valid`, `7D`, and `30D` targets. It has no liquidity, spread, crossed-quote, contract-age, or target-expiry quality gate.

The derived range currently takes the larger of resolved-IV one-trading-day move and a realized-volatility projection. It rejects stale underlying quotes but does not reject stale option contract data. Yahoo options are free/unofficial and must remain labelled as delayed risk context rather than OPRA/0DTE flow.

## Current Discord Delivery

The notifier disables mentions and maps levels to separate webhooks. It persists legacy alert events before sending and performs one synchronous retry for a `429` only when `Retry-After` is at most five seconds. There is no persistent outbox, idempotency key, delivery state, worker, multi-attempt retry, low-priority digest, or durable daily delivery budget. Existing policy candidates are database-constrained to `shadow_only`.

## Classification

### Available

- Persistent quote, 1m bar, RWA proxy, option-contract, IV, feature, search-evidence, and shadow-policy storage.
- Freshness timestamps for core market data and partial freshness metadata for news/options.
- Existing basic rules: price level, intraday move, relative underperformance, volume spike, stale-data block.
- Discord webhook transport with disabled mentions and a safe test endpoint.

### Present But Not Trusted For Trading Decisions

- Yahoo quote/bar/options data: free/unofficial, with provider timestamp and contract staleness limitations.
- Bitget RWA: auxiliary overnight/market-state signal, not official US equity market data or order book.
- Firecrawl search result age: candidate discovery only; publication must be sourced from article metadata or official feed.
- Resolved IV and expected move: useful risk-range approximation after quality gates, not an American-option valuation or live OPRA flow.

### Missing

- SEC ticker-to-CIK mapping, SEC submissions ingestion, accession de-duplication, and compliant SEC User-Agent configuration.
- Configurable official-source registry and RSS/API-first ingestion.
- Full-text evidence capture, body hashing, evidence versioning, and explicit content failure records.
- Scheduled alert evaluation under the collector scheduler.
- Session-aware hysteresis and durable daily alert budget.
- Durable Discord outbox with idempotency, multi-attempt 429 retry, and digest delivery.
- Explicit option expiry buckets and contract quality gates.

## Recommended Changes

1. Add a versioned evidence schema, official source registry, and SEC filing tables; retain raw evidence and failure metadata.
2. Add a dedicated official-source/SEC collector and a Firecrawl full-text fetch step for newly discovered URLs. Keep `map` low-frequency and optional; do not use broad crawl in the hot path.
3. Move price-rule evaluation into the active collector scheduler and route all deliveries through a durable outbox.
4. Add expiry buckets and quality status to option snapshots before option data can upgrade an alert.
5. Keep every newly implemented market/news alert in shadow or test mode until the production env diff and alert thresholds are explicitly approved.

## Production Change Boundary

No production polling intervals, secrets, systemd units, or live notification rules were changed by this audit. The implementation should be validated locally and through read-only production endpoints before an explicit deployment approval.
