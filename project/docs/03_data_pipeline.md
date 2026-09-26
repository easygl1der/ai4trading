# Data Inventory And Caveats

This document describes reusable data assets after the 2026-07-02 reset.

It intentionally does not document the old derived causal pipeline, because that pipeline was archived as invalid.

## Reusable Raw Data

| Folder | Contents | Status |
|---|---|---|
| `data/bitget` | Bitget MUUSDT/SNDKUSDT 1h candles | reusable, requires tracking-quality audit |
| `data/yahoo` | Yahoo MU/SNDK 1h bars with pre/post labels | reusable with caveats |
| `data/korea` | Samsung and SK Hynix 1h bars | reusable with caveats |
| `data/etf` | EWY/KORU 1h bars | reusable with caveats |
| `data/macro` | SOX, VIX, USD/KRW, USD/JPY | reusable controls |
| `data/fred` | Daily macro series | reusable controls |
| `data/meta` | Holiday, earnings, event dummies | reusable, must be verified before causal use |

## Known Data Caveats

1. Yahoo pre/post bars have zero reported volume in this snapshot.

   Treat pre/post prices as indicative quotes or vendor bars until validated. Do not use them as confirmed tradeable mediator prices without further checks.

2. Bitget RWA perpetuals must be audited before use as equity-price proxies.

   Required checks:

   - overlap return correlation versus Yahoo regular-session bars;
   - basis versus Yahoo close by session;
   - stale-price frequency;
   - volume/liquidity by hour;
   - contract metadata and launch/open time.

3. Korea 06:00 UTC bars exist but were excluded in the old analysis.

   Future work must explicitly decide whether 06:00 UTC is post/closing auction/noisy data and document the rule.

4. Row-shift lags are unsafe across market gaps.

   Future lead-lag work must use real timestamp differences, not just `.shift(1)`, because the Korean series has overnight, weekend, and holiday gaps.

5. Old derived analysis files are invalid.

   Files previously under `data/analysis` were moved to `archive_invalid_2026-07-02/data_analysis`.

## Recommended New Data Layers

Create new derived data only after passing the audit above.

Suggested derived layers:

- `data/derived_tracking/`: Bitget-vs-Yahoo tracking and basis diagnostics.
- `data/derived_leadlag/`: timestamp-safe lag panels.
- `data/derived_design/`: only if a valid causal design is approved.

Every derived file should include a small metadata sidecar:

```json
{
  "status": "pilot",
  "created_at": "YYYY-MM-DD",
  "source_files": [],
  "timing_rule": "",
  "valid_for_causal_claims": false
}
```
