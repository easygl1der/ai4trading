# Data Source Research for Risk-Aware Market Agent

Last updated: 2026-07-11 HKT

This note records the practical data-source choices for the watchlist monitoring system. The key boundary is simple:

- Free sources can support an MVP for quotes, bars, SEC filings, macro data, news crawling, delayed options exploration, and Bitget/Ondo RWA proxy signals.
- Real NBBO, full Level 2 order book, real-time OPRA options, reliable production Greeks/IV, low-latency news, and commercial redistribution require paid vendors and licensing.

## Recommended MVP Stack

| Layer | Source | Cost | Use | Caveat |
|---|---:|---:|---|---|
| US equity/ETF quote and bars | Yahoo chart path | Free, unofficial | MVP quote snapshot and `1m` bars | No SLA, not Level 2, API may change |
| Overnight / off-hours proxy | Bitget Wallet RWA / Ondo `*on` tickers | Free through existing MCP | RWA stock info and K-line, especially for `NVDAon`, `TSLAon`, `AAPLon` | Proxy only, not official Nasdaq/SIP truth |
| News/event evidence | Firecrawl + official pages/RSS | Free to low-cost | Headlines, IR pages, SEC release pages, macro release pages | Crawling is not low-latency professional news |
| Filings | SEC EDGAR APIs | Free | 8-K, 10-Q, 10-K, ownership filings | Respect SEC rate limits |
| Macro | FRED, BLS, BEA, Treasury | Free | CPI/jobs/Fed/macro context | Event calendars still need extra handling |
| Options exploration | Yahoo / delayed chains | Free | Prototype option-chain shape and IV solver | Do not rely on Yahoo IV as production truth |

Current implementation uses Yahoo Finance MCP for the free options prototype:

- `GET /api/v1/options/:symbol/chain?near=20`
- `GET /api/v1/options/:symbol/expected-move`
- `GET /api/v1/features/:symbol/intraday`

This is enough to estimate a rough nearest-expiration ATM straddle move for risk-range calibration. The backend also re-solves IV from bid/ask midpoint using a Black-Scholes bisection solver, while keeping Yahoo raw IV as advisory metadata.

## Low-Cost Upgrade Path

| Priority | Provider | Approximate cost | Adds | Notes |
|---:|---|---:|---|---|
| 1 | MarketData.app Trader | around $30/month annual or $75/month monthly | Real-time OPRA options for non-professional internal use, chain, IV, Greeks | Strong options-first MVP candidate |
| 2 | ThetaData Standard | around $80/month | OPRA NBBO options, Greeks, IV, tick/1m/history | Better options research depth |
| 3 | Alpaca Algo Trader Plus | around $99/month | SIP US equity data and OPRA options feed access | Good personal production quote upgrade |
| 4 | Massive Advanced | around $199/month | More complete stocks/options real-time bundle | Useful if one vendor should cover more layers |
| 5 | Databento, Nasdaq TotalView, Cboe One, SIP direct | hundreds to thousands/month | Real microstructure and depth | Not MVP-scope |

## What Is Not Free in Practice

| Need | Reality |
|---|---|
| Official NBBO / SIP | Usually requires paid market-data access. Free IEX-only quotes are not full-market NBBO. |
| Real Level 2 / order book | Requires licensed depth feeds such as Nasdaq TotalView, NYSE/OpenBook, Cboe depth, Databento, or broker/vendor feeds. |
| Real-time OPRA options | Requires OPRA-backed provider or broker entitlement. |
| Production-grade IV/Greeks | Vendor values are model outputs; for auditability, store inputs and re-solve IV from bid/ask/mid. |
| Low-latency market news | Free crawling can explain, but cannot replace professional event feeds. |
| Commercial redistribution | Usually requires explicit licensing even when a provider has a cheap individual plan. |

## Provider Notes

### Yahoo

Use Yahoo for the current MVP because it is easy and already verified in this repo. The backend records provider timestamps and `dataAgeSeconds` so the signal engine can reject stale data.

Do not treat Yahoo as:

- official real-time NBBO,
- Level 2/order-book data,
- a durable production SLA,
- a reliable source of production IV.

### Bitget / Ondo RWA

Use Bitget RWA as an off-hours proxy. It can answer questions like:

- Did `NVDAon` move meaningfully overnight?
- Is there a tokenized-stock premium/discount relative to the official previous close?
- Did RWA momentum confirm or diverge from the US premarket move?

Do not treat it as:

- the official US equity price,
- official NBBO,
- Nasdaq/NYSE depth,
- a direct substitute for SIP data.

The current backend exposes:

- `GET /api/v1/rwa/:symbol/info`
- `GET /api/v1/rwa/:symbol/kline?period=1m&size=60`
- `GET /api/v1/rwa/:symbol/summary`

### Options and IV

For the agent's "reasonable range" model, options matter because they measure priced-in amplitude, not direction. The production approach should be:

1. Pull near-term option quotes from an OPRA-backed source.
2. Store bid, ask, mid, volume, open interest, quote timestamp, and provider delay.
3. Re-solve ATM IV from option prices.
4. Convert annualized IV into a one-day expected move.
5. Compare realized move against implied move and current news context.

Yahoo option-chain IV can be used for exploration, but the workflow should not trust the raw `impliedVolatility` field as final.

The free-version rule is:

- Use Yahoo options for chain shape, ATM strike selection, bid/ask/mid, volume, open interest, and a rough straddle expected move.
- Record provider, received timestamp, quote timestamp, last trade timestamp, and warnings.
- Do not call it real-time OPRA options.
- Do not use raw Yahoo IV as the final IV signal; use the backend's resolved IV from option prices before using it in serious alerts.

## Source Links

- Alpaca market data docs: https://docs.alpaca.markets/us/docs/about-market-data-api
- Alpaca real-time stock data docs: https://docs.alpaca.markets/us/docs/real-time-stock-pricing-data
- Massive stock snapshot docs: https://massive.com/docs/rest/stocks/snapshots/single-ticker-snapshot
- Tradier market data docs: https://docs.tradier.com/docs/market-data
- MarketData.app pricing: https://www.marketdata.app/pricing/
- ThetaData options data: https://www.thetadata.net/options-data
- ThetaData pricing: https://www.thetadata.net/pricing
- Tradier options chain docs: https://docs.tradier.com/reference/brokerage-api-markets-get-options-chains
- ORATS near-EOD data: https://orats.com/near-eod-data
- SEC EDGAR APIs: https://www.sec.gov/search-filings/edgar-application-programming-interfaces
- SEC rate limits: https://www.sec.gov/filergroup/announcements-old/new-rate-control-limits
- BLS API features: https://www.bls.gov/bls/api_features.htm
- FRED API terms: https://fred.stlouisfed.org/docs/api/terms_of_use.html
- BEA API signup: https://apps.bea.gov/api/signup/
- Bitget Wallet RWA docs: https://web3.bitget.com/en/docs/market/rwa
- Bitget Wallet WebSocket docs: https://web3.bitget.com/en/docs/market/websocket
- Ondo Global Markets overview: https://docs.ondo.finance/ondo-global-markets/overview
- Ondo important notes: https://docs.ondo.finance/ondo-global-markets/important-notes
- Chainlink Ondo tokenized equity feeds: https://docs.chain.link/data-feeds/tokenized-equity-feeds/ondo
