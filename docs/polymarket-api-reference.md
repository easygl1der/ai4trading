# Polymarket API Reference

## Three APIs

| API | Base URL | Purpose | Auth Required |
|-----|----------|---------|---------------|
| **Gamma API** | `https://gamma-api.polymarket.com` | Markets, events, tags, series, comments, sports, search, public profiles | No |
| **Data API** | `https://data-api.polymarket.com` | User positions, trades, activity, holder data, open interest, leaderboards, builder analytics | No |
| **CLOB API** | `https://clob.polymarket.com` | Orderbook, pricing, order placement/cancellation, trading | Partial |

A separate **Bridge API** (`https://bridge.polymarket.com`) handles deposits/withdrawals via fun.xyz.

## Authentication

- **Gamma API** — Fully public, no authentication
- **Data API** — Fully public, no authentication
- **CLOB API** — Public endpoints for orderbook/prices; authenticated for trading

## Endpoint Categories

### Events
- `GET /events` — List events (keyset pagination)
- `GET /events` — List events (offset pagination)
- `GET /events/{id}` — Get event by ID
- `GET /events/slug/{slug}` — Get event by slug
- `GET /events/{id}/tags` — Get event tags

### Markets
- `GET /markets/keyset` — List markets (keyset pagination)
- `GET /markets` — List markets (offset pagination)
- `GET /markets/{id}` — Get market by ID
- `GET /markets/{slug}` — Get market by slug
- `GET /markets/{id}/tags` — Get market tags
- `GET /markets` — Get market by token (`?clob_token_ids=...`)
- `GET /markets/top-holders` — Get top holders (`?market_id=...`)
- `GET /markets/{id}/open-interest` — Get open interest
- `GET /events/{id}/live-volume` — Get live volume for event

### Orderbook & Pricing
- `GET /orderbook` — Get order book
- `POST /orderbooks` — Get order books (request body)
- `GET /prices` — Get market price
- `GET /prices` — Get market prices (query params)
- `POST /prices` — Get market prices (request body)
- `GET /midpoint` — Get midpoint price
- `GET /midpoints` — Get midpoint prices (query params)
- `POST /midpoints` — Get midpoint prices (request body)
- `GET /spread` — Get spread
- `POST /spreads` — Get spreads
- `GET /last-trade-price` — Get last trade price
- `GET /last-trade-prices` — Get last trade prices (query)
- `POST /last-trade-prices` — Get last trade prices (body)
- `GET /prices/history` — Get prices history
- `POST /prices/history/batch` — Get batch prices history
- `GET /fee-rate` — Get fee rate
- `GET /fee-rate/{path}` — Get fee rate by path
- `GET /tick-size` — Get tick size
- `GET /tick-size/{path}` — Get tick size by path
- `GET /info` — Get CLOB market info
- `GET /time` — Get server time

### Orders (CLOB API — Auth Required)
- `POST /orders` — Post new order
- `DELETE /orders/{id}` — Cancel single order
- `GET /orders/{id}` — Get order by ID
- `POST /orders` — Post multiple orders
- `GET /orders` — Get user orders
- `DELETE /orders` — Cancel multiple orders
- `DELETE /orders` — Cancel all orders
- `DELETE /orders` — Cancel orders for market
- `GET /orders/{id}/status` — Get order scoring status
- `POST /heartbeat` — Send heartbeat

### Trades
- `GET /trades` — Get trades
- `GET /trades/builder` — Get builder trades

### CLOB Markets
- `GET /markets` — Get simplified markets
- `GET /markets/sampling` — Get sampling markets
- `GET /markets/sampling/simplified` — Get sampling simplified markets

### Rebates
- `GET /rebates` — Get current rebated fees for maker
- `GET /rewards` — Get current active rewards configurations
- `GET /rewards/raw` — Get raw rewards for specific market
- `GET /rewards/markets` — Get multiple markets with rewards
- `GET /rewards/earnings` — Get earnings for user by date
- `GET /rewards/earnings/total` — Get total earnings for user by date
- `GET /rewards/percentages` — Get reward percentages for user
- `GET /rewards/user` — Get user earnings and markets config

### Profile
- `GET /profile/{address}` — Get public profile
- `GET /positions` — Get current positions
- `GET /positions/closed` — Get closed positions
- `GET /activity` — Get user activity
- `GET /positions/value` — Get total value of positions
- `GET /trades` — Get trades for user or markets
- `GET /trades/count` — Get total markets user has traded
- `GET /positions/{marketId}` — Get positions for market
- `GET /accounting` — Download accounting snapshot (ZIP of CSVs)

### Leaderboard
- `GET /leaderboard` — Get trader leaderboard
- `GET /leaderboard/builders` — Get aggregated builder leaderboard
- `GET /leaderboard/builders/daily` — Get daily builder volume time-series

### Search
- `GET /search` — Search markets, events, and profiles

### Tags
- `GET /tags` — List tags
- `GET /tags/{id}` — Get tag by ID
- `GET /tags/slug/{slug}` — Get tag by slug
- `GET /tags/{id}/related` — Get related tags by tag ID
- `GET /tags/slug/{slug}/related` — Get related tags by slug
- `GET /tags/{id}/related-to` — Get tags related to tag ID
- `GET /tags/slug/{slug}/related-to` — Get tags related to slug

### Series
- `GET /series` — List series
- `GET /series/{id}` — Get series by ID

### Comments
- `GET /comments` — List comments
- `GET /comments/{id}` — Get comments by ID
- `GET /comments/user/{address}` — Get comments by user

### Sports
- `GET /sports` — Get sports metadata
- `GET /sports/market-types` — Get valid sports market types
- `GET /teams` — List teams

### Bridge
- `GET /assets` — Get supported assets
- `POST /deposit-addresses` — Create deposit addresses
- `POST /quote` — Get a quote
- `GET /tx/status` — Get transaction status
- `POST /withdrawal-addresses` — Create withdrawal addresses

### Relayer
- `POST /submit` — Submit transaction
- `GET /tx/{id}` — Get transaction by ID
- `GET /txs` — Get recent transactions
- `GET /nonce` — Get current nonce
- `GET /address` — Get relayer address and nonce
- `GET /safe` — Check if safe is deployed
- `GET /keys` — Get all relayer API keys

### WebSocket
- `WSS /ws/market` — Market channel
- `WSS /ws/user` — User channel
- `WSS /ws/sports` — Sports channel
