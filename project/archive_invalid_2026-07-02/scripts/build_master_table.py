"""
构造分析主表 analysis_master.parquet
------------------------------------
以韩国交易时段小时 (306 rows) 为索引，把 D/Y/Z/M/X 全部对齐:
  D_h  = 韩国半导体加权收益率 (r_kr, 已在 eta_hourly 表)
  Y_h  = Bitget MUUSDT 下一小时对数收益率 (处理小时的下一个 bar)
  Y2_h = Bitget SNDKUSDT 下一小时对数收益率 (对照)
  Z_h  = (eta*I_OPEN, eta*I_MID, eta*I_CLOSE) 三维 IV
  M1_h = EWY 下一个非空 bar 对数收益率 (中介 1)
  M2_h = KORU 下一个非空 bar 对数收益率 (中介 2)
  X    = r_VIX_h, r_USDJPY_h, r_SOX_lag, Y_{h-1}, kr_holiday, mu_earn_day
"""
import os
import numpy as np
import pandas as pd

DATA = "/home/user/workspace/data"
OUT = f"{DATA}/analysis"
os.makedirs(OUT, exist_ok=True)


def load_and_ret(path, name):
    df = pd.read_parquet(f"{DATA}/{path}")
    tcol = "ts_utc" if "ts_utc" in df.columns else "ts"
    df[tcol] = pd.to_datetime(df[tcol], utc=True)
    df = df[[tcol, "close"]].dropna().sort_values(tcol).rename(columns={"close": name, tcol: "ts_utc"})
    df[f"r_{name}"] = np.log(df[name] / df[name].shift(1))
    return df


# ============ 1. 加载 eta 表（已含 D_h, session, r_sox） ============
eta = pd.read_parquet(f"{OUT}/eta_hourly.parquet")
eta["ts_utc"] = pd.to_datetime(eta["ts_utc"], utc=True)
print(f"eta_hourly: {len(eta)} rows, ts {eta['ts_utc'].min()} ~ {eta['ts_utc'].max()}")

# ============ 2. Bitget MUUSDT / SNDKUSDT 下一小时收益率 ============
bit_mu = load_and_ret("bitget/muusdt_1h.parquet", "muusdt")
bit_sn = load_and_ret("bitget/sndkusdt_1h.parquet", "sndkusdt")

# Y_h = MU 在 [h+1, h+2] 的收益率 = 处理时刻 h 之后一个 bar 的价格变化
# 具体做法: merge 时用 ts_utc + 1h 对齐 bit_mu["r_muusdt"]
def next_bar_return(base, next_df, ret_col, out_col):
    tmp = next_df[["ts_utc", ret_col]].rename(columns={ret_col: out_col})
    tmp["ts_utc"] = tmp["ts_utc"] - pd.Timedelta(hours=1)  # 让 h+1 的收益率对齐到 h
    return base.merge(tmp, on="ts_utc", how="left")

df = next_bar_return(eta, bit_mu, "r_muusdt", "Y_h")
df = next_bar_return(df, bit_sn, "r_sndkusdt", "Y2_h")
print(f"合并 Bitget 后: Y_h 非空 {df['Y_h'].notna().sum()} / {len(df)}, Y2_h 非空 {df['Y2_h'].notna().sum()}")

# ============ 3. 中介 M_h: EWY / KORU ============
# 注意: EWY / KORU 在韩国盘中 (00:00-06:00 UTC) 是 US 收盘后 = ET 20:00-次日02:00
# 所以韩国 h 对应的 "EWY 盘前收益率" 应该是 ET 04:00-09:29 的 pre-market bar
# 简化处理: 用 EWY/KORU 在 h+8 ~ h+15 的下一个非空 bar 的收益率 (对应美股开盘前后)
ewy = load_and_ret("etf/ewy_1h.parquet", "ewy")
koru = load_and_ret("etf/koru_1h.parquet", "koru")


def next_avail_return(base, other_df, ret_col, out_col, min_gap_h=8, max_gap_h=15):
    """对每个 base 时间戳, 找 other_df 中 [h+min_gap, h+max_gap] 范围内第一个非空的 return."""
    other = other_df[["ts_utc", ret_col]].dropna().sort_values("ts_utc").reset_index(drop=True)
    other_ts_np = other["ts_utc"].values
    other_ret = other[ret_col].values
    results = []
    for base_ts in base["ts_utc"]:
        low = base_ts + pd.Timedelta(hours=min_gap_h)
        high = base_ts + pd.Timedelta(hours=max_gap_h)
        # 找第一个 >= low 且 <= high 的
        idx = np.searchsorted(other_ts_np, np.datetime64(low.tz_convert("UTC").tz_localize(None)))
        if idx < len(other_ts_np):
            ts_found = pd.Timestamp(other_ts_np[idx], tz="UTC")
            if ts_found <= high:
                results.append(other_ret[idx])
                continue
        results.append(np.nan)
    base[out_col] = results
    return base


df = next_avail_return(df, ewy, "r_ewy", "M1_h", min_gap_h=8, max_gap_h=16)
df = next_avail_return(df, koru, "r_koru", "M2_h", min_gap_h=8, max_gap_h=16)
print(f"中介 M1 (EWY) 非空: {df['M1_h'].notna().sum()}, M2 (KORU) 非空: {df['M2_h'].notna().sum()}")

# ============ 4. 控制变量 X ============
vix = load_and_ret("macro/vix_1h.parquet", "vix")   # VIX 只在 US 时段
jpy = load_and_ret("macro/usdjpyx_1h.parquet", "jpy")   # FX 24h

# VIX: 用同一 US 交易日的 VIX 日度变化 (韩国时段 VIX 不动, 只能用 daily)
vix["us_date"] = vix["ts_utc"].dt.tz_convert("America/New_York").dt.date
vix_daily = vix.groupby("us_date").agg(
    vix_open=("vix", "first"),
    vix_close=("vix", "last"),
).reset_index()
vix_daily["r_vix_d"] = np.log(vix_daily["vix_close"] / vix_daily["vix_close"].shift(1))

# 韩国 h 对齐的 US 交易日 = 前一个 US 交易日 (已在 eta 里做过, 可复用)
# 这里直接根据 ts_utc 找 prev US trading date
sox_dates = set(vix_daily["us_date"])
def prev_us_date(ts):
    d = ts.date() - pd.Timedelta(days=1).to_pytimedelta()
    while d not in sox_dates:
        d = d - pd.Timedelta(days=1).to_pytimedelta()
        if abs((ts.date().year - d.year)) > 1:
            return None
    return d

df["_prev_us_date"] = df["ts_utc"].apply(prev_us_date)
df = df.merge(vix_daily[["us_date", "r_vix_d"]], left_on="_prev_us_date", right_on="us_date", how="left").drop(columns=["us_date"])

# JPY 24h, 直接同小时对齐
jpy_h = jpy[["ts_utc", "r_jpy"]].copy()
df = df.merge(jpy_h, on="ts_utc", how="left")
df = df.rename(columns={"r_jpy": "r_jpy_h"})

# Y_{h-1} 自回归项 (处理内生性 + 减少 U_us 遗留)
df["Y_lag1"] = df["Y_h"].shift(1)

# 事件哑变量
events = pd.read_parquet(f"{DATA}/meta/events_daily.parquet")
events["date"] = pd.to_datetime(events["date"]).dt.date
df["_date"] = df["ts_utc"].dt.date
df = df.merge(events[["date", "kr_holiday", "mu_earnings", "sndk_earnings"]],
              left_on="_date", right_on="date", how="left").drop(columns=["date"])
df[["kr_holiday", "mu_earnings", "sndk_earnings"]] = df[["kr_holiday", "mu_earnings", "sndk_earnings"]].fillna(0)

# ============ 5. 构造 Z_h 三维 IV ============
df["I_OPEN"] = (df["session"] == "OPEN").astype(float)
df["I_MID"] = (df["session"] == "MID").astype(float)
df["I_CLOSE"] = (df["session"] == "CLOSE").astype(float)
df["Z_OPEN"] = df["eta"] * df["I_OPEN"]
df["Z_MID"] = df["eta"] * df["I_MID"]
df["Z_CLOSE"] = df["eta"] * df["I_CLOSE"]

# 处理变量 D_h = r_kr
df["D_h"] = df["r_kr"]

# ============ 6. 清洗并保存 ============
keep = ["ts_utc", "session", "D_h", "Y_h", "Y2_h", "M1_h", "M2_h",
        "Z_OPEN", "Z_MID", "Z_CLOSE", "eta", "I_OPEN", "I_MID", "I_CLOSE",
        "r_vix_d", "r_jpy_h", "r_sox_d", "Y_lag1",
        "kr_holiday", "mu_earnings", "sndk_earnings",
        "alpha_s", "beta_s"]
master = df[keep].copy()
# 只保留 D 和 Z 都非空的行
master_clean = master.dropna(subset=["D_h", "Y_h", "Z_OPEN", "Z_MID", "Z_CLOSE"]).reset_index(drop=True)
print(f"\n主表 (含缺失): {len(master)} rows")
print(f"主表 (清洗后): {len(master_clean)} rows")
print(f"  Y_h 非空: {master_clean['Y_h'].notna().sum()}")
print(f"  M1_h 非空: {master_clean['M1_h'].notna().sum()}")
print(f"  M2_h 非空: {master_clean['M2_h'].notna().sum()}")
print(f"  r_vix_d 非空: {master_clean['r_vix_d'].notna().sum()}")
print(f"  r_jpy_h 非空: {master_clean['r_jpy_h'].notna().sum()}")
print(f"  时段分布: {master_clean['session'].value_counts().to_dict()}")

master.to_parquet(f"{OUT}/analysis_master.parquet", index=False)
master_clean.to_parquet(f"{OUT}/analysis_master_clean.parquet", index=False)
print(f"\n[OK] 保存: {OUT}/analysis_master.parquet ({len(master)} rows)")
print(f"[OK] 保存: {OUT}/analysis_master_clean.parquet ({len(master_clean)} rows)")

# 描述性统计
print("\n=== 主变量描述性统计 ===")
print(master_clean[["D_h", "Y_h", "Y2_h", "M1_h", "M2_h", "r_vix_d", "r_jpy_h"]].describe())
