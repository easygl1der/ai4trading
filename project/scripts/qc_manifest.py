"""
数据质量核查 + 生成 manifest.
- 检查 NaN / 时间戳单调 / 频率断点 / 时间对齐
- 生成 manifest.json 记录每份数据的元信息
- 生成 manifest.md 便于人工阅读
"""
import os, json, hashlib
from datetime import datetime, timezone
import pandas as pd

DATA_ROOT = "/home/user/workspace/data"

FILES = [
    # (path, symbol, source, tz_col, price_col, notes)
    ("bitget/muusdt_1h.parquet",  "MUUSDT",   "Bitget CEX (perp)",  "ts",     "close", "USDT-FUTURES perp, includes 24/7 overnight"),
    ("bitget/sndkusdt_1h.parquet","SNDKUSDT", "Bitget CEX (perp)",  "ts",     "close", "USDT-FUTURES perp"),
    ("yahoo/mu_1h.parquet",       "MU",       "Yahoo Finance",      "ts_utc", "close", "US equity, includePrePost=true, scale=3"),
    ("yahoo/sndk_1h.parquet",     "SNDK",     "Yahoo Finance",      "ts_utc", "close", "US equity, pre/post"),
    ("korea/005930_ks_1h.parquet","005930.KS","Yahoo Finance",      "ts_utc", "close", "Samsung Electronics, KRW"),
    ("korea/000660_ks_1h.parquet","000660.KS","Yahoo Finance",      "ts_utc", "close", "SK Hynix, KRW"),
    ("etf/koru_1h.parquet",       "KORU",     "Yahoo Finance",      "ts_utc", "close", "Direxion 3x Korea Bull"),
    ("etf/ewy_1h.parquet",        "EWY",      "Yahoo Finance",      "ts_utc", "close", "iShares MSCI Korea ETF"),
    ("macro/sox_1h.parquet",      "^SOX",     "Yahoo Finance",      "ts_utc", "close", "PHLX Semiconductor Index"),
    ("macro/vix_1h.parquet",      "^VIX",     "Yahoo Finance",      "ts_utc", "close", "CBOE Volatility Index"),
    ("macro/krwx_1h.parquet",     "KRW=X",    "Yahoo Finance",      "ts_utc", "close", "USD/KRW spot"),
    ("macro/usdjpyx_1h.parquet",  "USDJPY=X", "Yahoo Finance",      "ts_utc", "close", "USD/JPY spot"),
    ("fred/fred_daily_wide.parquet", "FRED_WIDE", "FRED",           None,     None,    "Daily macro: DGS10, WTI, VIXCLS, DEXKOUS, T10Y3M, T10Y2Y"),
    ("meta/korea_holidays.parquet","KR_HOLIDAYS","meta",            None,     None,    "KRX full closures in window"),
    ("meta/earnings.parquet",     "EARNINGS", "meta",               None,     None,    "MU / SNDK earnings dates"),
    ("meta/events_daily.parquet", "EVENTS",   "meta",               None,     None,    "Daily event dummies"),
]

def md5_of(fpath):
    h = hashlib.md5()
    with open(fpath, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


def profile(fpath, sym, source, tz_col, price_col, notes):
    df = pd.read_parquet(fpath)
    n = len(df)
    info = {
        "path": os.path.relpath(fpath, DATA_ROOT),
        "symbol": sym,
        "source": source,
        "rows": n,
        "cols": list(df.columns),
        "size_bytes": os.path.getsize(fpath),
        "md5": md5_of(fpath),
        "notes": notes,
    }
    if tz_col and tz_col in df.columns:
        ts = df[tz_col]
        info["ts_first"] = str(ts.iloc[0])
        info["ts_last"] = str(ts.iloc[-1])
        # 单调
        info["ts_monotonic"] = bool(pd.Series(ts).is_monotonic_increasing)
        # 唯一
        info["ts_unique_ratio"] = round(ts.nunique() / n, 4)
        # 频率断点：相邻间隔的中位数与超过 2x 的比例
        if pd.api.types.is_datetime64_any_dtype(ts):
            diff = ts.diff().dropna()
            if len(diff):
                med = diff.median()
                info["ts_median_gap"] = str(med)
                info["ts_gap_gt_2x_pct"] = round((diff > 2 * med).mean() * 100, 2)
    if price_col and price_col in df.columns:
        p = df[price_col]
        info["price_nan_pct"] = round(p.isna().mean() * 100, 2)
        info["price_min"] = float(p.min()) if p.notna().any() else None
        info["price_max"] = float(p.max()) if p.notna().any() else None
        info["price_mean"] = float(p.mean()) if p.notna().any() else None
    if "session" in df.columns:
        info["session_counts"] = df["session"].value_counts().to_dict()
    return info


results = []
for rel, sym, src, tz_col, price_col, notes in FILES:
    fpath = os.path.join(DATA_ROOT, rel)
    if not os.path.exists(fpath):
        print(f"[MISS] {fpath}")
        continue
    try:
        info = profile(fpath, sym, src, tz_col, price_col, notes)
        results.append(info)
        print(f"[OK] {sym}: {info['rows']} rows, md5={info['md5'][:8]}")
    except Exception as e:
        print(f"[ERR] {sym}: {e}")

manifest = {
    "window": {"start": "2026-04-11", "end": "2026-07-01"},
    "generated_at": datetime.now(timezone.utc).isoformat(),
    "purpose": "Raw market-data snapshot for Bitget RWA perpetual tracking, Korea/US semiconductor lead-lag mapping, and future causal-design feasibility checks. No causal claim is validated by this manifest.",
    "files": results,
    "known_issues": [
        "TEDRATE series discontinued at FRED (no obs in window).",
        "Yahoo MU/SNDK prices are scaled (scale=3): raw close ~1150 USD, real MU price ~$180 - need /scale for actual analysis.",
        "005930.KS / 000660.KS have hasPrePostMarketData=false; only regular Korean session hours.",
        "Yahoo pre/post bars should be audited before use because reported pre/post volume is often zero.",
        "Korean holiday mismatch alone is too sparse for a reliable IV design in this 76-day window.",
        "USDJPYX=X and KRW=X return continuous FX data (~1380 hourly rows) - 24h FX market.",
        "^SOX and Korean stocks: no pre/post pre-market at exchange level.",
        "Bitget MUUSDT starts 2026-04-11, SNDKUSDT starts 2026-04-15.",
    ],
    "next_steps": [
        "Audit Bitget-vs-Yahoo tracking quality by session before using Bitget as an equity proxy.",
        "Build timestamp-safe predictive lead-lag panels; do not use row-shift lags across market gaps.",
        "Search for a real external shock before making any causal claim.",
        "Keep mediation analysis disabled until treatment, mediator, and outcome timing are valid.",
    ],
}

out_json = f"{DATA_ROOT}/manifest.json"
with open(out_json, "w") as f:
    json.dump(manifest, f, indent=2, ensure_ascii=False, default=str)
print(f"\n[OK] Manifest written: {out_json}")

# Markdown 版本便于阅读
out_md = f"{DATA_ROOT}/manifest.md"
lines = []
lines.append(f"# 数据 Manifest\n")
lines.append(f"- **窗口**: {manifest['window']['start']} → {manifest['window']['end']}")
lines.append(f"- **生成时间 (UTC)**: {manifest['generated_at']}")
lines.append(f"- **用途**: {manifest['purpose']}\n")
lines.append("## 数据文件清单\n")
lines.append("| 符号 | 来源 | 行数 | 时间范围 | 价格 min/max/mean | 备注 |")
lines.append("|---|---|---|---|---|---|")
for r in results:
    ts_range = ""
    if "ts_first" in r:
        ts_range = f"{r['ts_first'][:19]} ~ {r['ts_last'][:19]}"
    pmm = ""
    if "price_min" in r and r["price_min"] is not None:
        pmm = f"{r['price_min']:.2f} / {r['price_max']:.2f} / {r['price_mean']:.2f}"
    lines.append(f"| {r['symbol']} | {r['source']} | {r['rows']} | {ts_range} | {pmm} | {r['notes']} |")
lines.append("\n## 已知问题\n")
for k in manifest["known_issues"]:
    lines.append(f"- {k}")
lines.append("\n## 下一步（分析阶段）\n")
for k in manifest["next_steps"]:
    lines.append(f"- {k}")
with open(out_md, "w") as f:
    f.write("\n".join(lines))
print(f"[OK] Manifest MD: {out_md}")
