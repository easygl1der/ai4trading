# Polybot Trade Timing Histogram

Generated: 2026-07-07T16:03:43+00:00 UTC

## Definition

- Scope: Strategy 1 `trades` table only. `late_trades` is excluded from the chart because the recent two strategy versions have `LATE_STRATEGY_ENABLED=false`.
- Event time: close timestamp minus the 5-minute market start timestamp.
- Bin width: 10 seconds over the 0-300 second market window.
- Positive bars: profitable close events. Negative bars: losing close events.
- PnL reconstruction for `trades`: `close_notional - close_fee - open_notional - open_fee`, paired by `(slug, side)` in chronological order.

![Trade timing histogram](../analysis/polybot_trade_timing_by_10s_20260707.png)

## Version Windows

| Version | UTC window | HKT window | Key settings |
| --- | --- | --- | --- |
| `previous` | 2026-06-25T09:18:15+00:00 to 2026-06-26T05:13:31+00:00 | 2026-06-25T17:18:15+08:00 to 2026-06-26T13:13:31+08:00 | 上一版：IMB 18pp, entry 10s, SL 12%, cooldown 0s |
| `current` | 2026-06-26T05:13:31+00:00 to now | 2026-06-26T13:13:31+08:00 to now | 当前版：IMB 18pp, entry 45s, SL 16%, cooldown 60s |

## Summary

| Version | Closes | Wins | Losses | Win rate | Net paper PnL | Win close timing | Loss close timing |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| `previous` | 45 | 25 | 20 | 55.6% | 31.79 | p25=41.3s, median=69.7s, p75=163.6s | p25=30.1s, median=42.0s, p75=74.5s |
| `current` | 149 | 92 | 57 | 61.7% | 156.97 | p25=101.5s, median=195.8s, p75=243.1s | p25=194.4s, median=247.9s, p75=274.1s |

## Densest 10-second Bins

### previous

- Profit closes: 30-40s: 3, 20-30s: 3, 60-70s: 3, 50-60s: 2, 40-50s: 2
- Loss closes: 30-40s: 4, 10-20s: 3, 70-80s: 3, 40-50s: 2, 20-30s: 2
- Exit reason groups: `{'take_profit': 25, 'stop_loss': 20}`
- Observed close window: 2026-06-25T09:30:58+00:00 to 2026-06-25T14:15:18+00:00 UTC

### current

- Profit closes: 70-80s: 9, 240-250s: 9, 250-260s: 6, 220-230s: 5, 280-290s: 5
- Loss closes: 270-280s: 10, 280-290s: 7, 260-270s: 5, 240-250s: 5, 190-200s: 4
- Exit reason groups: `{'take_profit': 92, 'stop_loss': 54, 'force_flatten': 3}`
- Observed close window: 2026-06-26T06:13:29+00:00 to 2026-07-04T09:54:42+00:00 UTC

## Files

- Chart PNG: `../analysis/polybot_trade_timing_by_10s_20260707.png`
- Event CSV: `../analysis/polybot_trade_timing_events_20260707.csv`
- Bin CSV: `../analysis/polybot_trade_timing_bins_20260707.csv`

## Data Quality Notes

- Main trade pairs reconstructed: 194; unmatched closes: 0; unmatched opens left after pairing: 0.
- Recent-window `late_trades` close rows found: 0; chart excludes them by design.
- The bot is live-writing the SQLite database, so extraction used read-only connection with retry.
