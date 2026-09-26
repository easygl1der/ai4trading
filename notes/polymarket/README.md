# Polymarket crypto paper trader

This folder contains a read-only paper-trading prototype for Polymarket BTC/ETH markets.

## Safety boundary

- No wallet integration.
- No private key handling.
- No authenticated Polymarket endpoints.
- No real order placement.

The prototype only uses:

- Gamma public search for market discovery.
- CLOB public health and price endpoints for read-only market data.
- Binance public ticker price for BTC/ETH reference prices.

## Run

From `/Users/yitwah/Documents/ai4trading`:

```bash
PYTHONPATH=notes/polymarket/src python3 notes/polymarket/src/run_paper_scan.py --max-markets 8 --bankroll 1000
```

Outputs are written to:

```text
notes/polymarket/outputs/
```

Each run creates:

- `market-scan-*.json`: full market discovery, CLOB checks, signals, orders, fills, and broker state.
- `paper-trades-*.jsonl`: one JSON decision record per scanned market.
- `run-report-*.md`: Chinese run report.

## Current validated run

The latest validated run is:

- `outputs/market-scan-20260610T033712Z.json`
- `outputs/paper-trades-20260610T033712Z.jsonl`
- `outputs/run-report-20260610T033712Z.md`

It found 68 candidate BTC/ETH markets, read CLOB prices for the top candidates, fetched Binance BTC/ETH reference prices, and generated one conservative ask-side paper fill.

