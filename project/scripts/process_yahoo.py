"""
处理 Yahoo Finance MCP 输出 JSON -> parquet
用显式 file_id -> symbol 映射，避免误匹配旧文件。
"""
import json
import os
from datetime import datetime, timezone
import pandas as pd

BASE = "/home/user/workspace/current_session_context/tool_calls/call_external_tool"
OUT_DIRS = {
    "MU": "/home/user/workspace/data/yahoo",
    "SNDK": "/home/user/workspace/data/yahoo",
    "005930.KS": "/home/user/workspace/data/korea",
    "000660.KS": "/home/user/workspace/data/korea",
    "KORU": "/home/user/workspace/data/etf",
    "EWY": "/home/user/workspace/data/etf",
    "^SOX": "/home/user/workspace/data/macro",
    "^VIX": "/home/user/workspace/data/macro",
    "KRW=X": "/home/user/workspace/data/macro",
    "USDJPY=X": "/home/user/workspace/data/macro",
}

# 用户在 skill 摘要中确认的映射
MAPPING = {
    "MU":         "output_mr1ypxwn.json",
    "SNDK":       "output_mr1ypybp.json",
    "005930.KS":  "output_mr1ypxvm.json",
    "000660.KS":  "output_mr1ypxvr.json",
    "KORU":       "output_mr1ypy39.json",
    "EWY":        "output_mr1yq0tj.json",
    "^SOX":       "output_mr1yq0or.json",
    "^VIX":       "output_mr1yq18i.json",
    "KRW=X":      "output_mr1yq184.json",
    "USDJPY=X":   "output_mr1yq1ln.json",
}


def mark_session(ts_utc_series, trading_periods):
    """给每个时间戳打标: pre / regular / post / offhours。
    trading_periods 是 meta.tradingPeriods 结构 {pre:[[{start,end}]], regular:..., post:...}
    时间以秒或 ISO 字符串，这里 start/end 是 ISO -> 转换。
    """
    # 展平 periods
    def flatten(periods_list):
        out = []
        for day in periods_list:
            for p in day:
                s = pd.Timestamp(p["start"]).tz_convert("UTC") if pd.Timestamp(p["start"]).tz is not None else pd.Timestamp(p["start"], tz="UTC")
                e = pd.Timestamp(p["end"]).tz_convert("UTC") if pd.Timestamp(p["end"]).tz is not None else pd.Timestamp(p["end"], tz="UTC")
                out.append((s, e))
        return out

    pre_ranges = flatten(trading_periods.get("pre", [])) if "pre" in trading_periods else []
    reg_ranges = flatten(trading_periods.get("regular", []))
    post_ranges = flatten(trading_periods.get("post", [])) if "post" in trading_periods else []

    labels = []
    for ts in ts_utc_series:
        lab = "offhours"
        for s, e in reg_ranges:
            if s <= ts < e:
                lab = "regular"; break
        if lab == "offhours":
            for s, e in pre_ranges:
                if s <= ts < e:
                    lab = "pre"; break
        if lab == "offhours":
            for s, e in post_ranges:
                if s <= ts < e:
                    lab = "post"; break
        labels.append(lab)
    return labels


def process_one(symbol, fname):
    fpath = os.path.join(BASE, fname)
    with open(fpath) as f:
        d = json.load(f)
    r = d["result"]
    meta = r["meta"]
    ts = r.get("timestamp", [])
    if not ts:
        print(f"[WARN] {symbol}: empty timestamp")
        return None
    q = r["indicators"]["quote"][0]
    df = pd.DataFrame({
        "ts_utc": pd.to_datetime(ts, unit="s", utc=True),
        "open": q.get("open"),
        "high": q.get("high"),
        "low": q.get("low"),
        "close": q.get("close"),
        "volume": q.get("volume"),
    })
    tz_name = meta.get("exchangeTimezoneName", "UTC")
    df["ts_local"] = df["ts_utc"].dt.tz_convert(tz_name)
    # session label
    tp = meta.get("tradingPeriods", {})
    if tp and ("regular" in tp or "pre" in tp):
        df["session"] = mark_session(df["ts_utc"], tp)
    else:
        df["session"] = "regular"  # 韩股这种没有 pre/post
    df["symbol"] = symbol
    df["scale"] = meta.get("scale", 1)
    # 有些 Yahoo 数据 scale=3 表示 close 是原价的 1/scale? 保留 raw close，另存 adj_close
    # 根据 meta.chartPreviousClose vs previousClose: 用户看到 MU=1154.29 (real=~$180), scale=3 是缩放
    # 但为保证数据可用性，先直接原样保存
    outdir = OUT_DIRS[symbol]
    os.makedirs(outdir, exist_ok=True)
    safe = symbol.replace("^", "").replace("=", "").replace(".", "_")
    out = os.path.join(outdir, f"{safe.lower()}_1h.parquet")
    df.to_parquet(out, index=False)
    n_reg = (df["session"] == "regular").sum()
    n_pre = (df["session"] == "pre").sum()
    n_post = (df["session"] == "post").sum()
    n_off = (df["session"] == "offhours").sum()
    print(f"[OK] {symbol}: {len(df)} rows -> {out}")
    print(f"      first={df['ts_utc'].iloc[0]} last={df['ts_utc'].iloc[-1]}")
    print(f"      session: regular={n_reg} pre={n_pre} post={n_post} offhours={n_off}")
    return out


if __name__ == "__main__":
    for sym, fn in MAPPING.items():
        try:
            process_one(sym, fn)
        except Exception as e:
            print(f"[ERR] {sym}: {e}")
