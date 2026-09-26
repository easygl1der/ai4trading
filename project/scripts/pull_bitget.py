"""
Pull Bitget CEX MUUSDT / SNDKUSDT perpetual klines - 76 days, 1h granularity.
Public REST API, no auth needed. Uses history-candles endpoint with pagination.

Time window: 2026-04-11 to 2026-07-01 (~82 days)
Output: parquet files in /home/user/workspace/data/bitget/
"""
import requests
import pandas as pd
from datetime import datetime, timezone
import time
import json
from pathlib import Path

BASE = "https://api.bitget.com/api/v2/mix/market/history-candles"
OUT_DIR = Path("/home/user/workspace/data/bitget")
OUT_DIR.mkdir(parents=True, exist_ok=True)

# Window: from 2026-04-11 00:00 UTC to now
end_ts = int(datetime.now(timezone.utc).timestamp() * 1000)
start_ts = int(datetime(2026, 4, 11, tzinfo=timezone.utc).timestamp() * 1000)

def pull_symbol(symbol: str, granularity: str = "1H"):
    """Walk backwards from now using endTime pagination."""
    all_rows = []
    cur_end = end_ts
    seen = set()
    max_iters = 200
    for it in range(max_iters):
        params = {
            "symbol": symbol,
            "productType": "USDT-FUTURES",
            "granularity": granularity,
            "endTime": cur_end,
            "limit": 200,
        }
        r = requests.get(BASE, params=params, timeout=30)
        r.raise_for_status()
        data = r.json()
        if data.get("code") != "00000":
            print(f"  [!] API error: {data}")
            break
        rows = data.get("data", [])
        if not rows:
            print(f"  [{it}] empty response, stopping")
            break
        # Deduplicate + track earliest
        new_rows = 0
        earliest = None
        for row in rows:
            ts = int(row[0])
            if ts in seen:
                continue
            seen.add(ts)
            all_rows.append(row)
            new_rows += 1
            if earliest is None or ts < earliest:
                earliest = ts
        if new_rows == 0:
            print(f"  [{it}] no new rows, done")
            break
        if earliest <= start_ts:
            print(f"  [{it}] reached start window ({datetime.fromtimestamp(earliest/1000, tz=timezone.utc)})")
            break
        # Next page: use earliest as new endTime
        cur_end = earliest
        time.sleep(0.15)  # Be nice to the API
    return all_rows

for sym in ["MUUSDT", "SNDKUSDT"]:
    print(f"\n=== {sym} ===")
    rows = pull_symbol(sym, "1H")
    if not rows:
        print(f"  [!] No data for {sym}")
        continue
    df = pd.DataFrame(rows, columns=["ts", "open", "high", "low", "close", "volume_base", "volume_quote"])
    df["ts"] = pd.to_datetime(df["ts"].astype(int), unit="ms", utc=True)
    for c in ["open", "high", "low", "close", "volume_base", "volume_quote"]:
        df[c] = pd.to_numeric(df[c], errors="coerce")
    df = df.sort_values("ts").reset_index(drop=True)
    # Filter to window
    df = df[df["ts"] >= pd.Timestamp("2026-04-11", tz="UTC")].reset_index(drop=True)
    out = OUT_DIR / f"{sym.lower()}_1h.parquet"
    df.to_parquet(out, index=False)
    print(f"  {sym}: {len(df)} rows, {df['ts'].min()} → {df['ts'].max()}")
    print(f"  saved → {out}")
    print(f"  sample close: {df['close'].iloc[-3:].tolist()}")

print("\n=== DONE ===")
