"""
Stage 1 分段回归诊断
=====================
目标: 验证 eta_h = r_KR_h - alpha_s - beta_s * r_SOX_{t-1} 中,
      三个时段 (OPEN, MID, CLOSE) 的 beta_s 是否显著异质,
      以及全球共同因子被残差化剔除多少。

设计:
- 韩国半导体板块 = 三星(005930.KS) + 海力士(000660.KS) 市值加权
- SOX 用日度前一交易日收盘对开盘的日度收益率 (韩国时段内美股闭市, 只能用 lag)
- 三段: KR-OPEN 00:00-01:00 UTC, KR-MID 01:00-05:00 UTC, KR-CLOSE 05:00-06:00 UTC
- Newey-West 稳健标准误 (lag=6h)
"""
import os
import numpy as np
import pandas as pd
import statsmodels.api as sm
from statsmodels.stats.diagnostic import het_breuschpagan

DATA = "/home/user/workspace/data"
OUT = "/home/user/workspace/data/analysis"
os.makedirs(OUT, exist_ok=True)


# ============ 1. 加载韩股 1h 数据 ============
def load_kr(sym_file):
    df = pd.read_parquet(f"{DATA}/korea/{sym_file}")
    df = df[["ts_utc", "close", "volume"]].copy()
    df["ts_utc"] = pd.to_datetime(df["ts_utc"], utc=True)
    df = df.dropna(subset=["close"]).sort_values("ts_utc").reset_index(drop=True)
    return df


ss = load_kr("005930_ks_1h.parquet")  # Samsung
sk = load_kr("000660_ks_1h.parquet")  # SK Hynix

# 合并对齐
kr = ss.rename(columns={"close": "ss_close", "volume": "ss_vol"}).merge(
    sk.rename(columns={"close": "sk_close", "volume": "sk_vol"}),
    on="ts_utc", how="inner"
)

# ============ 2. 市值加权 (用简化的固定权重) ============
# 三星总股本 ~ 59.7 亿股, 海力士 ~ 7.28 亿股
# 用窗口首日收盘价计算市值
shares_ss = 5.97e9
shares_sk = 0.728e9
mktcap_ss = shares_ss * kr["ss_close"].iloc[0]
mktcap_sk = shares_sk * kr["sk_close"].iloc[0]
w_ss = mktcap_ss / (mktcap_ss + mktcap_sk)
w_sk = 1 - w_ss
print(f"市值权重: Samsung={w_ss:.3f}, SK Hynix={w_sk:.3f}")

# 板块加权指数 (取对数便于线性组合)
kr["ss_logp"] = np.log(kr["ss_close"])
kr["sk_logp"] = np.log(kr["sk_close"])
kr["kr_semi_logp"] = w_ss * kr["ss_logp"] + w_sk * kr["sk_logp"]

# 小时收益率 = 一阶差分
kr["r_kr"] = kr["kr_semi_logp"].diff()

# ============ 3. 构造时段标签 ============
kr["hour_utc"] = kr["ts_utc"].dt.hour
def label_session(h):
    if h == 0:
        return "OPEN"     # KR 09:00-10:00 KST
    elif 1 <= h <= 4:
        return "MID"      # KR 10:00-14:00 KST
    elif h == 5:
        return "CLOSE"    # KR 14:00-15:00 KST
    else:
        return None       # 06:00 UTC 是 15:00 KST 收盘拍卖前, 剔除避免噪声
kr["session"] = kr["hour_utc"].apply(label_session)

# ============ 4. 构造 SOX 日度前置收益率 r_SOX_{t-1} ============
sox = pd.read_parquet(f"{DATA}/macro/sox_1h.parquet")
sox["ts_utc"] = pd.to_datetime(sox["ts_utc"], utc=True)
sox = sox.dropna(subset=["close"]).sort_values("ts_utc").reset_index(drop=True)

# SOX 每交易日的开盘价 & 收盘价
sox["us_date"] = sox["ts_utc"].dt.tz_convert("America/New_York").dt.date
sox_daily = sox.groupby("us_date").agg(
    sox_open=("open", "first"),
    sox_close=("close", "last"),
).reset_index()
# 日度收益率 (close_t / close_{t-1})
sox_daily["r_sox_d"] = np.log(sox_daily["sox_close"] / sox_daily["sox_close"].shift(1))
# 日度 open-to-close (盘中)
sox_daily["r_sox_intra"] = np.log(sox_daily["sox_close"] / sox_daily["sox_open"])
# 日度 close-to-open (隔夜跳空)
sox_daily["r_sox_overnight"] = np.log(sox_daily["sox_open"] / sox_daily["sox_close"].shift(1))
print(f"\nSOX 日度记录数: {len(sox_daily)}")

# 对齐到韩国时段: 韩国 t 日的 00:00-06:00 UTC 对应美国 t-1 日的完整交易日之后
# 韩国 UTC 时间 h 对应的美股"最近已知信息"日期 = h.date() 的前一天的美股交易日
def latest_us_trading_day(kr_ts_utc, sox_dates):
    kr_date = kr_ts_utc.date()
    # 韩国 UTC 早上, 对应美股是 kr_date-1 的完整交易日 (14:30-21:00 UTC 那天)
    candidate = kr_date - pd.Timedelta(days=1).to_pytimedelta()
    # 若 candidate 不是美股交易日 (周末/假日), 往前找
    dates_set = set(sox_dates)
    while candidate not in dates_set:
        candidate = candidate - pd.Timedelta(days=1).to_pytimedelta()
        if (kr_date.year - candidate.year) > 1:
            return None
    return candidate

sox_date_set = set(sox_daily["us_date"])
kr["prev_us_date"] = kr["ts_utc"].apply(lambda t: latest_us_trading_day(t, sox_date_set))
kr = kr.merge(
    sox_daily[["us_date", "r_sox_d", "r_sox_intra", "r_sox_overnight"]],
    left_on="prev_us_date", right_on="us_date", how="left"
).drop(columns=["us_date"])

# ============ 5. 清洗: 保留有 session 标签且非空 ============
data = kr.dropna(subset=["r_kr", "r_sox_d", "session"]).reset_index(drop=True)
print(f"\n有效样本数: {len(data)}")
print(f"时段分布:\n{data['session'].value_counts()}")

# ============ 6. 分段回归 ============
def run_segment(subset, segment_name):
    """对某个时段跑 r_kr ~ 1 + r_sox_d, Newey-West lag=6."""
    y = subset["r_kr"].values
    X = sm.add_constant(subset["r_sox_d"].values)
    # OLS
    m_ols = sm.OLS(y, X).fit()
    # Newey-West HAC
    m_nw = sm.OLS(y, X).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
    # Breusch-Pagan 异方差检验
    try:
        bp = het_breuschpagan(m_ols.resid, X)
        bp_p = bp[1]
    except Exception:
        bp_p = np.nan
    return {
        "segment": segment_name,
        "n": len(subset),
        "alpha": m_nw.params[0],
        "alpha_se_nw": m_nw.bse[0],
        "alpha_t_nw": m_nw.tvalues[0],
        "beta": m_nw.params[1],
        "beta_se_nw": m_nw.bse[1],
        "beta_t_nw": m_nw.tvalues[1],
        "beta_pval_nw": m_nw.pvalues[1],
        "r2": m_ols.rsquared,
        "r2_adj": m_ols.rsquared_adj,
        "resid_std": m_ols.resid.std(),
        "bp_pval": bp_p,
        "model": m_nw,
    }

results = []
for seg in ["OPEN", "MID", "CLOSE"]:
    sub = data[data["session"] == seg]
    if len(sub) < 10:
        print(f"[WARN] {seg} 样本太少: {len(sub)}")
        continue
    r = run_segment(sub, seg)
    results.append(r)

# 池化 (对照组: 不分段, 只加 session 固定效应)
pool = data.copy()
pool_D = pd.get_dummies(pool["session"], prefix="s", drop_first=True).astype(float)
X_pool = pd.concat([pd.Series(1.0, index=pool.index, name="const"),
                    pool_D,
                    pool[["r_sox_d"]].rename(columns={"r_sox_d": "beta_pool"})], axis=1)
m_pool_ols = sm.OLS(pool["r_kr"].values, X_pool.values).fit()
m_pool_nw = sm.OLS(pool["r_kr"].values, X_pool.values).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
pool_result = {
    "segment": "POOLED (session FE)",
    "n": len(pool),
    "alpha": m_pool_nw.params[0],
    "alpha_se_nw": m_pool_nw.bse[0],
    "alpha_t_nw": m_pool_nw.tvalues[0],
    "beta": m_pool_nw.params[-1],
    "beta_se_nw": m_pool_nw.bse[-1],
    "beta_t_nw": m_pool_nw.tvalues[-1],
    "beta_pval_nw": m_pool_nw.pvalues[-1],
    "r2": m_pool_ols.rsquared,
    "r2_adj": m_pool_ols.rsquared_adj,
    "resid_std": m_pool_ols.resid.std(),
    "bp_pval": np.nan,
    "model": m_pool_nw,
}
results.append(pool_result)

# ============ 7. 输出汇总表 ============
tbl = pd.DataFrame([{k: v for k, v in r.items() if k != "model"} for r in results])
tbl = tbl[["segment", "n", "alpha", "alpha_t_nw", "beta", "beta_se_nw",
           "beta_t_nw", "beta_pval_nw", "r2", "resid_std", "bp_pval"]]
print("\n" + "="*100)
print("分段回归汇总 (Newey-West HAC, lag=6h)")
print("="*100)
print(tbl.to_string(index=False, float_format=lambda x: f"{x:.4f}" if abs(x) < 1e4 else f"{x:.2e}"))

# ============ 8. 检验 beta 是否显著异质: 池化交互模型 ============
# r_kr = alpha_s + beta_s * r_sox_d + eps
# 用 session dummy * r_sox_d 交互项检验
inter = data.copy()
inter["is_OPEN"] = (inter["session"] == "OPEN").astype(float)
inter["is_MID"] = (inter["session"] == "MID").astype(float)
inter["is_CLOSE"] = (inter["session"] == "CLOSE").astype(float)
inter["sox_x_MID"] = inter["is_MID"] * inter["r_sox_d"]
inter["sox_x_CLOSE"] = inter["is_CLOSE"] * inter["r_sox_d"]
X_inter = sm.add_constant(inter[["is_MID", "is_CLOSE", "r_sox_d", "sox_x_MID", "sox_x_CLOSE"]].values)
m_inter = sm.OLS(inter["r_kr"].values, X_inter).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
print("\n" + "="*100)
print("交互模型: beta_OPEN 是基准, 检验 (beta_MID - beta_OPEN) 和 (beta_CLOSE - beta_OPEN)")
print("="*100)
labels = ["const=alpha_OPEN", "delta_alpha_MID", "delta_alpha_CLOSE",
          "beta_OPEN", "delta_beta_MID", "delta_beta_CLOSE"]
for lab, coef, se, t, p in zip(labels, m_inter.params, m_inter.bse, m_inter.tvalues, m_inter.pvalues):
    print(f"  {lab:22s}: coef={coef:+.5f}  se={se:.5f}  t={t:+.2f}  p={p:.3f}")

# Wald test: H0: delta_beta_MID = delta_beta_CLOSE = 0
R = np.zeros((2, len(m_inter.params)))
R[0, 4] = 1  # sox_x_MID
R[1, 5] = 1  # sox_x_CLOSE
w = m_inter.wald_test(R, use_f=False, scalar=True)
print(f"\n联合 Wald 检验 H0: beta 三段相等")
print(f"  chi2 = {float(w.statistic):.3f}, df=2, p-value = {float(w.pvalue):.4f}")

# ============ 9. 保存 eta_h (用分段回归的残差) ============
def compute_eta(subset, seg):
    y = subset["r_kr"].values
    X = sm.add_constant(subset["r_sox_d"].values)
    m = sm.OLS(y, X).fit()
    return m.resid, m.params

eta_all = []
for seg in ["OPEN", "MID", "CLOSE"]:
    sub = data[data["session"] == seg].copy()
    resid, params = compute_eta(sub, seg)
    sub["eta"] = resid
    sub["alpha_s"] = params[0]
    sub["beta_s"] = params[1]
    eta_all.append(sub)
eta_df = pd.concat(eta_all).sort_values("ts_utc").reset_index(drop=True)

eta_out = eta_df[["ts_utc", "session", "r_kr", "r_sox_d", "alpha_s", "beta_s", "eta"]]
eta_out.to_parquet(f"{OUT}/eta_hourly.parquet", index=False)
print(f"\n[OK] eta_h 保存到 {OUT}/eta_hourly.parquet ({len(eta_out)} rows)")

# 残差诊断
print("\n=== eta 描述性统计 (按时段) ===")
print(eta_out.groupby("session")["eta"].describe()[["count", "mean", "std", "min", "max"]])

# 保存汇总表
tbl.to_csv(f"{OUT}/stage1_regression_summary.csv", index=False)
print(f"[OK] 表格保存到 {OUT}/stage1_regression_summary.csv")
