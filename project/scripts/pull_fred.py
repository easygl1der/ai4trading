"""
拉 FRED 日频宏观数据 (2026-04-01 to now)
系列:
- DGS10: 10Y treasury constant maturity yield
- DCOILWTICO: WTI crude oil spot
- TEDRATE: TED spread (若已停更，用备选)
- VIXCLS: VIX close (备份对照)
- DEXKOUS: 韩元/美元汇率 (日度)
- T10Y3M: 10Y - 3M term spread
凭证走 QueryParam api_key，自动注入。
"""
import os, sys, json, time, subprocess
import pandas as pd

CA = os.environ.get("SSL_CERT_FILE", "/etc/ssl/certs/agent-proxy-ca-2.pem")

SERIES = ["DGS10", "DCOILWTICO", "TEDRATE", "VIXCLS", "DEXKOUS", "T10Y3M", "T10Y2Y"]
START = "2026-04-01"
END = "2026-07-02"
OUT = "/home/user/workspace/data/fred"
os.makedirs(OUT, exist_ok=True)

BASE = "https://api.stlouisfed.org/fred/series/observations"


def pull(series_id):
    params = {
        "series_id": series_id,
        "observation_start": START,
        "observation_end": END,
        "file_type": "json",
    }
    from urllib.parse import urlencode
    url = BASE + "?" + urlencode(params)
    cmd = ["curl", "-sS", "--cacert", CA, url]
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    if r.returncode != 0:
        print(f"[ERR] {series_id}: curl rc={r.returncode} stderr={r.stderr[:300]}")
        return None
    try:
        d = json.loads(r.stdout)
    except Exception as e:
        print(f"[ERR] {series_id}: json parse {e}: {r.stdout[:300]}")
        return None
    obs = d.get("observations", [])
    if not obs:
        print(f"[WARN] {series_id}: no obs")
        return None
    df = pd.DataFrame(obs)
    df = df[["date", "value"]].copy()
    df["date"] = pd.to_datetime(df["date"])
    # 缺失记 "." -> NaN
    df["value"] = pd.to_numeric(df["value"], errors="coerce")
    df["series"] = series_id
    out = os.path.join(OUT, f"{series_id.lower()}.parquet")
    df.to_parquet(out, index=False)
    print(f"[OK] {series_id}: {len(df)} rows, {df['value'].notna().sum()} valid -> {out}")
    return df


if __name__ == "__main__":
    all_dfs = []
    for s in SERIES:
        df = pull(s)
        if df is not None:
            all_dfs.append(df)
        time.sleep(0.3)
    # 合并宽表
    if all_dfs:
        combined = pd.concat(all_dfs, ignore_index=True)
        wide = combined.pivot(index="date", columns="series", values="value").sort_index()
        wide.to_parquet(os.path.join(OUT, "fred_daily_wide.parquet"))
        print(f"\n[OK] Combined wide table: {wide.shape} -> {OUT}/fred_daily_wide.parquet")
        print(wide.tail())
