# Polybot Hour-of-Day Trade Histogram

Generated: 2026-07-07T16:07:05+00:00

## Definition

- Scope: Strategy 1 `trades` table only, using the existing event-level extraction from the live VPS database.
- Time zone: Hong Kong time.
- X-axis: close hour of day, 0-23.
- Positive bars: profitable close events. Negative bars: losing close events.

![Hourly histogram](../analysis/polybot_trade_hour_of_day_by_1h_20260708.png)

## Summary

| Version | Closes | Wins | Losses | Win rate | Net paper PnL |
| --- | ---: | ---: | ---: | ---: | ---: |
| `previous` | 45 | 25 | 20 | 55.6% | 31.79 |
| `current` | 149 | 92 | 57 | 61.7% | 156.97 |

## Densest Hours

### previous

- Profit closes: 21:00-21:59: 13, 19:00-19:59: 6, 20:00-20:59: 4, 17:00-17:59: 1, 22:00-22:59: 1
- Loss closes: 21:00-21:59: 8, 22:00-22:59: 7, 20:00-20:59: 4, 19:00-19:59: 1
- Best net-PnL hours: 21:00-21:59 25.96, 19:00-19:59 25.61, 17:00-17:59 7.41, 20:00-20:59 2.14, 00:00-00:59 0
- Worst net-PnL hours: 22:00-22:59 -29.33, 00:00-00:59 0, 01:00-01:59 0, 02:00-02:59 0, 03:00-03:59 0

### current

- Profit closes: 21:00-21:59: 12, 22:00-22:59: 11, 08:00-08:59: 8, 23:00-23:59: 7, 04:00-04:59: 6
- Loss closes: 21:00-21:59: 8, 14:00-14:59: 6, 10:00-10:59: 6, 11:00-11:59: 5, 06:00-06:59: 4
- Best net-PnL hours: 23:00-23:59 43.6, 22:00-22:59 41.4, 19:00-19:59 38.31, 03:00-03:59 28.08, 09:00-09:59 23.05
- Worst net-PnL hours: 11:00-11:59 -24.94, 13:00-13:59 -21.09, 10:00-10:59 -19.75, 12:00-12:59 -16.46, 02:00-02:59 -11.12

## Files

- Chart PNG: `../analysis/polybot_trade_hour_of_day_by_1h_20260708.png`
- Hour-bin CSV: `../analysis/polybot_trade_hour_of_day_bins_20260708.csv`
- Source event CSV: `../analysis/polybot_trade_timing_events_20260707.csv`
