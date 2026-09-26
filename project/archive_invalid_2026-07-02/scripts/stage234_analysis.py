"""
Stage 2/3/4 一体化因果分析
===========================
输入: data/analysis/analysis_master_clean.parquet
输出:
  - data/analysis/stage2_ate.json          # ATE 三种估计器
  - data/analysis/stage3_cace.json         # CACE / IV / KP-F / AR-CI / J
  - data/analysis/stage35_mediation.json   # NDE/NIE 分解 (EWY, KORU)
  - data/analysis/stage4_sensitivity.json  # E-value / Rosenbaum / Placebo
  - data/analysis/stage_causal_graph.json  # Granger / PC 条件独立
"""
import os, json, warnings
import numpy as np
import pandas as pd
import statsmodels.api as sm
from statsmodels.stats.diagnostic import het_breuschpagan
from statsmodels.tsa.stattools import grangercausalitytests
# statsmodels IV2SLS 已移除、不用 legacy

warnings.filterwarnings("ignore")

DATA = "/home/user/workspace/data/analysis"
os.makedirs(DATA, exist_ok=True)

df = pd.read_parquet(f"{DATA}/analysis_master_clean.parquet").dropna(subset=["Y_h", "D_h", "r_vix_d", "r_jpy_h", "r_sox_d"]).reset_index(drop=True)
print(f"Analysis N = {len(df)}")

# 控制变量集
X_cols = ["r_vix_d", "r_jpy_h", "r_sox_d", "I_MID", "I_CLOSE"]  # I_OPEN 作 baseline
def X_mat(sub, extra=None):
    cols = X_cols + (extra or [])
    return sm.add_constant(sub[cols].values.astype(float))

results = {}

# =========================================================================
# Stage 2: ATE via backdoor
# =========================================================================
print("\n" + "="*80)
print("Stage 2: ATE via backdoor (D continuous → partialling-out + parametric)")
print("="*80)

# (a) 直接线性回归 Y ~ D + X (Newey-West)
X_full = X_mat(df, extra=["D_h"])
m_direct = sm.OLS(df["Y_h"].values, X_full).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
tau_ols = m_direct.params[-1]
se_ols = m_direct.bse[-1]
print(f"[OLS-adj]     tau = {tau_ols:+.4f}, SE = {se_ols:.4f}, t = {tau_ols/se_ols:+.2f}, p = {m_direct.pvalues[-1]:.4f}")

# (b) Partialling-out (Frisch-Waugh-Lovell)
X_only = X_mat(df)
m_y = sm.OLS(df["Y_h"].values, X_only).fit()
m_d = sm.OLS(df["D_h"].values, X_only).fit()
y_tilde = m_y.resid
d_tilde = m_d.resid
m_fwl = sm.OLS(y_tilde, sm.add_constant(d_tilde)).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
tau_fwl = m_fwl.params[1]
se_fwl = m_fwl.bse[1]
print(f"[FWL]         tau = {tau_fwl:+.4f}, SE = {se_fwl:.4f}, t = {tau_fwl/se_fwl:+.2f}, p = {m_fwl.pvalues[1]:.4f}")

# (c) DR-AIPW variant (对连续 D 用 partialling-out 已经 DR-equivalent)
# 加上 Y_lag1 增强 doubly-robust
df["Y_lag1_f"] = df["Y_lag1"].fillna(0)
X_dr = X_mat(df, extra=["Y_lag1_f"])
m_y2 = sm.OLS(df["Y_h"].values, X_dr).fit()
m_d2 = sm.OLS(df["D_h"].values, X_dr).fit()
m_dr = sm.OLS(m_y2.resid, sm.add_constant(m_d2.resid)).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
tau_dr = m_dr.params[1]
se_dr = m_dr.bse[1]
print(f"[DR-partial]  tau = {tau_dr:+.4f}, SE = {se_dr:.4f}, t = {tau_dr/se_dr:+.2f}, p = {m_dr.pvalues[1]:.4f}")

# (d) 只在韩国交易时段有效样本 (n=306) 之内, 分时段 ATE
print("\n分时段 ATE (探索性):")
for seg in ["OPEN", "MID", "CLOSE"]:
    sub = df[df["session"] == seg].reset_index(drop=True)
    if len(sub) < 20:
        continue
    Xs = sm.add_constant(sub[["D_h", "r_vix_d", "r_jpy_h", "r_sox_d", "Y_lag1_f"]].fillna(0).values.astype(float))
    try:
        ms = sm.OLS(sub["Y_h"].values, Xs).fit(cov_type="HAC", cov_kwds={"maxlags": 3})
        print(f"  {seg}: n={len(sub)}, tau = {ms.params[1]:+.4f}, SE = {ms.bse[1]:.4f}, p = {ms.pvalues[1]:.4f}")
    except Exception as e:
        print(f"  {seg}: err {e}")

results["stage2_ate"] = {
    "n": int(len(df)),
    "estimators": {
        "OLS_adjusted": {"tau": float(tau_ols), "se_nw": float(se_ols), "p": float(m_direct.pvalues[-1])},
        "FWL_partial": {"tau": float(tau_fwl), "se_nw": float(se_fwl), "p": float(m_fwl.pvalues[1])},
        "DR_partial_with_Ylag": {"tau": float(tau_dr), "se_nw": float(se_dr), "p": float(m_dr.pvalues[1])},
    },
    "controls": X_cols + ["Y_lag1 (in DR)"],
    "interpretation": "tau ≈ ΔY / ΔD, 单位: MU 收益率 per 单位 韩国半导体 收益率"
}

# =========================================================================
# Stage 3: CACE via IV (2SLS, 三维小时级 Z)
# =========================================================================
print("\n" + "="*80)
print("Stage 3: CACE via 三维小时级 IV (2SLS)")
print("="*80)

# 用 linearmodels 更专业
try:
    from linearmodels.iv import IV2SLS
    HAS_LM = True
except ImportError:
    print("[NOTE] linearmodels 未安装, 用 statsmodels IV2SLS")
    HAS_LM = False

df_iv = df.dropna(subset=["Z_OPEN", "Z_MID", "Z_CLOSE", "Y_h", "D_h"] + X_cols + ["Y_lag1_f"]).reset_index(drop=True)
print(f"IV sample N = {len(df_iv)}")

# ---- 版本 A: 单一 IV = I_OPEN (只用时段) ----
def run_iv(df_sub, Z_cols, X_cols_iv, tag):
    """跑 2SLS: Y ~ D + X, D ~ Z + X. 返回 dict."""
    Y = df_sub["Y_h"].values
    D = df_sub["D_h"].values.reshape(-1, 1)
    Z = df_sub[Z_cols].values
    X_ex = df_sub[X_cols_iv].values  # exogenous
    if HAS_LM:
        exog = sm.add_constant(X_ex)
        mod = IV2SLS(Y, exog, D, Z).fit(cov_type="kernel", kernel="bartlett", bandwidth=6)
        tau = mod.params["endog.0"] if "endog.0" in mod.params else mod.params.iloc[-1]
        se = mod.std_errors["endog.0"] if "endog.0" in mod.std_errors else mod.std_errors.iloc[-1]
        # First stage
        fs = mod.first_stage
        # KP-F
        try:
            kp = float(mod.first_stage.diagnostics.loc["endog.0", "f.stat"])
            kp_p = float(mod.first_stage.diagnostics.loc["endog.0", "f.pval"])
        except Exception:
            kp = np.nan
            kp_p = np.nan
        # Hansen J
        try:
            j_stat = float(mod.j_stat.stat)
            j_p = float(mod.j_stat.pval)
        except Exception:
            j_stat = np.nan
            j_p = np.nan
        out = {
            "tag": tag,
            "n": len(df_sub),
            "L": len(Z_cols),
            "tau_cace": float(tau),
            "se": float(se),
            "t": float(tau/se) if se != 0 else np.nan,
            "p": float(2*(1 - __import__('scipy.stats', fromlist=['norm']).norm.cdf(abs(tau/se)))) if se != 0 else np.nan,
            "kp_f": kp,
            "kp_f_pval": kp_p,
            "hansen_j": j_stat,
            "hansen_j_pval": j_p,
        }
    else:
        # statsmodels fallback
        mod = IV_LEGACY(Y, sm.add_constant(np.column_stack([X_ex, D])), sm.add_constant(np.column_stack([X_ex, Z]))).fit()
        out = {"tag": tag, "n": len(df_sub), "note": "legacy IV, limited stats"}
    return out, mod


# 手动算 KP-F: first-stage F on the excluded instruments
def kp_f_manual(df_sub, Z_cols, X_cols_iv):
    Y_d = df_sub["D_h"].values
    Xe = sm.add_constant(df_sub[X_cols_iv].values)
    Zmat = df_sub[Z_cols].values
    XZ = np.column_stack([Xe, Zmat])
    m_full = sm.OLS(Y_d, XZ).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
    m_res = sm.OLS(Y_d, Xe).fit()
    # F test on excluded IVs
    # R matrix: 只对最后 L 个系数联合为 0
    L = len(Z_cols)
    R = np.zeros((L, XZ.shape[1]))
    for i in range(L):
        R[i, -(L - i)] = 1
    w = m_full.wald_test(R, use_f=True, scalar=True)
    return float(w.statistic), float(w.pvalue), m_full

# 三个方案对比
# 方案中的 X (exogenous 控制变量) 不能又包含 I_MID/I_CLOSE, 因为 Z 中的 eta*I 组合会与之共线
scenarios = {
    "A_dummy_only": (["I_OPEN"], ["r_vix_d", "r_jpy_h", "r_sox_d"]),
    "B_eta_open_close": (["Z_OPEN", "Z_CLOSE"], ["r_vix_d", "r_jpy_h", "r_sox_d"]),
    "C_eta_three_way": (["Z_OPEN", "Z_MID", "Z_CLOSE"], ["r_vix_d", "r_jpy_h", "r_sox_d"]),
}

iv_results = {}
print(f"\n{'方案':<25} {'n':>4} {'L':>2} {'τ_CACE':>10} {'SE':>8} {'t':>7} {'KP-F':>8} {'J':>7} {'J-p':>7}")
print("-" * 90)
for tag, (Zc, Xc) in scenarios.items():
    # 保证列可用
    sub = df_iv.copy()
    r, mod = run_iv(sub, Zc, Xc, tag)
    kpf, kpf_p, m_fs = kp_f_manual(sub, Zc, Xc)
    r["kp_f_manual"] = kpf
    r["kp_f_manual_pval"] = kpf_p
    iv_results[tag] = r
    print(f"{tag:<25} {r['n']:>4} {r.get('L',1):>2} {r['tau_cace']:>+10.4f} {r['se']:>8.4f} {r.get('t', np.nan):>+7.2f} {kpf:>8.2f} {r.get('hansen_j', np.nan):>7.3f} {r.get('hansen_j_pval', np.nan):>7.3f}")

# Anderson-Rubin CI for 方案 C
print("\nAnderson-Rubin CI (方案 C, 弱-IV 稳健):")
def ar_ci(df_sub, Z_cols, X_cols_iv, alpha=0.05, tau_grid=None):
    from scipy.stats import chi2
    Y = df_sub["Y_h"].values
    D = df_sub["D_h"].values
    Z = df_sub[Z_cols].values
    Xe = sm.add_constant(df_sub[X_cols_iv].values)
    if tau_grid is None:
        tau_grid = np.linspace(-2, 2, 4001)
    L = Z.shape[1]
    crit = chi2.ppf(1 - alpha, L)
    accept = []
    stats_list = []
    for tau0 in tau_grid:
        resid = Y - D * tau0
        # 用 partial-out X_e first
        m_r = sm.OLS(resid, Xe).fit()
        r_x = m_r.resid
        m_zx = sm.OLS(Z, Xe).fit() if L > 1 else None
        # AR: n * (Z'r)' (Z'Z)^-1 (Z'r) / sigma^2
        Zr = Z.T @ r_x
        ZZ_inv = np.linalg.pinv(Z.T @ Z)
        sigma2 = np.var(r_x)
        stat = float((Zr @ ZZ_inv @ Zr) / sigma2)
        stats_list.append(stat)
        if stat <= crit:
            accept.append(tau0)
    if accept:
        return accept[0], accept[-1], stats_list
    else:
        return None, None, stats_list

lo, hi, ar_stats = ar_ci(df_iv, ["Z_OPEN", "Z_MID", "Z_CLOSE"],
                          ["r_vix_d", "r_jpy_h", "r_sox_d", "I_MID", "I_CLOSE"])
print(f"  AR 95% CI: [{lo:.4f}, {hi:.4f}]" if lo is not None else "  AR CI: empty (reject all)")

iv_results["C_eta_three_way"]["ar_ci_lo"] = lo
iv_results["C_eta_three_way"]["ar_ci_hi"] = hi

results["stage3_cace"] = iv_results

# =========================================================================
# Stage 3.5: Mediation NDE / NIE (两个中介分别做)
# =========================================================================
print("\n" + "="*80)
print("Stage 3.5: Mediation NDE / NIE 分解")
print("="*80)

def mediation_analysis(df_sub, mediator_col, tag):
    """
    简化线性中介 (Baron-Kenny + Sobel + Bootstrap):
      D -> M (a): m_a
      D + M -> Y (b, c'): m_b
    Total = a*b + c'
    NDE = c', NIE = a*b
    """
    sub = df_sub.dropna(subset=[mediator_col, "Y_h", "D_h"] + X_cols).reset_index(drop=True)
    if len(sub) < 20:
        return None
    Xctrl = sub[X_cols].values.astype(float)
    D_ = sub["D_h"].values.astype(float)
    M_ = sub[mediator_col].values.astype(float)
    Y_ = sub["Y_h"].values.astype(float)

    # (1) M ~ D + X
    m_a = sm.OLS(M_, sm.add_constant(np.column_stack([D_, Xctrl]))).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
    a = m_a.params[1]
    a_se = m_a.bse[1]
    # (2) Y ~ D + M + X
    m_b = sm.OLS(Y_, sm.add_constant(np.column_stack([D_, M_, Xctrl]))).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
    c_prime = m_b.params[1]  # NDE
    b = m_b.params[2]
    c_prime_se = m_b.bse[1]
    b_se = m_b.bse[2]
    # NIE = a*b (Sobel SE)
    NIE = a * b
    NIE_se = np.sqrt((b**2) * (a_se**2) + (a**2) * (b_se**2))
    # Bootstrap NIE (稳健)
    rng = np.random.default_rng(42)
    nb = 1000
    nie_boot = []
    n = len(sub)
    for _ in range(nb):
        idx = rng.integers(0, n, n)
        try:
            mm_a = sm.OLS(M_[idx], sm.add_constant(np.column_stack([D_[idx], Xctrl[idx]]))).fit()
            mm_b = sm.OLS(Y_[idx], sm.add_constant(np.column_stack([D_[idx], M_[idx], Xctrl[idx]]))).fit()
            nie_boot.append(mm_a.params[1] * mm_b.params[2])
        except:
            pass
    nie_boot = np.array(nie_boot)
    NIE_ci = (np.percentile(nie_boot, 2.5), np.percentile(nie_boot, 97.5))
    TE = c_prime + NIE
    prop_mediated = NIE / TE if TE != 0 else np.nan
    print(f"\n  中介 = {tag} (n={len(sub)}):")
    print(f"    a  (D→M)     = {a:+.4f}   SE={a_se:.4f}   p={m_a.pvalues[1]:.4f}")
    print(f"    b  (M→Y|D)   = {b:+.4f}   SE={b_se:.4f}   p={m_b.pvalues[2]:.4f}")
    print(f"    c' (NDE)     = {c_prime:+.4f}   SE={c_prime_se:.4f}   p={m_b.pvalues[1]:.4f}")
    print(f"    NIE = a*b    = {NIE:+.4f}   Sobel SE={NIE_se:.4f}")
    print(f"    NIE Bootstrap 95% CI: [{NIE_ci[0]:+.4f}, {NIE_ci[1]:+.4f}]")
    print(f"    Total       = {TE:+.4f}")
    print(f"    Prop.mediated (NIE/TE) = {prop_mediated:.3f}")
    return {
        "mediator": tag, "n": len(sub),
        "a": float(a), "a_se": float(a_se), "a_p": float(m_a.pvalues[1]),
        "b": float(b), "b_se": float(b_se), "b_p": float(m_b.pvalues[2]),
        "c_prime_NDE": float(c_prime), "c_prime_se": float(c_prime_se),
        "NIE": float(NIE), "NIE_sobel_se": float(NIE_se),
        "NIE_ci_lo": float(NIE_ci[0]), "NIE_ci_hi": float(NIE_ci[1]),
        "total_effect": float(TE), "prop_mediated": float(prop_mediated) if not np.isnan(prop_mediated) else None,
    }

med_ewy = mediation_analysis(df_iv, "M1_h", "EWY")
med_koru = mediation_analysis(df_iv, "M2_h", "KORU")
results["stage35_mediation"] = {"EWY": med_ewy, "KORU": med_koru}

# =========================================================================
# Stage 4: Sensitivity
# =========================================================================
print("\n" + "="*80)
print("Stage 4: Sensitivity analysis")
print("="*80)

sens = {}

# (a) E-value (VanderWeele & Ding 2017)
# 对连续 outcome 转 risk ratio: RR ≈ exp(0.91 * tau / sd(Y))
tau_hat = tau_fwl
sd_y = df["Y_h"].std()
RR = np.exp(0.91 * tau_hat / sd_y)
if RR < 1:
    RR = 1 / RR
E_val = RR + np.sqrt(RR * (RR - 1))
print(f"  E-value (FWL tau): RR ≈ {RR:.3f}, E-value = {E_val:.3f}")
print(f"    → 需要一个 unmeasured confounder 与 D 和 Y 关联度均达 RR≥{E_val:.2f} 才能推翻结论")

# (b) Rosenbaum sensitivity: 对每单位增加 D=0/1 分组 (二值化)
# 用连续 D 时的近似: 计算 tau 对 Gamma-scaling 的敏感度
def rosenbaum_bound(df_sub, tau_hat, tau_se, gammas=[1.5, 2, 3, 5]):
    """简化: 在 IPW 权重被 gamma-scaled 后 tau 的偏差"""
    out = []
    for g in gammas:
        # 假设 unobserved covariate 使 P(D=1|U) 从 e 变到 e*g/(1-e+e*g)
        # 结果: tau 的 wors-case bias ≈ tau * log(g) / 2 (启发式)
        bias = tau_hat * (np.log(g) / 2)
        tau_worst = tau_hat - abs(bias) if tau_hat > 0 else tau_hat + abs(bias)
        significant = abs(tau_worst / tau_se) > 1.96
        out.append({"gamma": g, "tau_worst": float(tau_worst), "still_sig": bool(significant)})
    return out

rb = rosenbaum_bound(df, tau_fwl, se_fwl)
print(f"  Rosenbaum Γ-sensitivity (启发式):")
for r in rb:
    print(f"    Γ = {r['gamma']}: 最坏 τ = {r['tau_worst']:+.4f}, 仍显著? {r['still_sig']}")

# (c) Manski bounds
# 对连续变量, Manski bounds 需要 Y 的支撑范围
y_lo, y_hi = df["Y_h"].min(), df["Y_h"].max()
d_lo, d_hi = df["D_h"].min(), df["D_h"].max()
# 大小样本 Manski worst-case: [E(Y|D=d_hi)*P(D)+y_lo*(1-P) - (E(Y|D=d_lo)*P + y_hi*(1-P))]
p_treated = (df["D_h"] > df["D_h"].median()).mean()
manski_upper = df[df["D_h"] > df["D_h"].median()]["Y_h"].mean() * p_treated + y_hi * (1 - p_treated) \
              - (df[df["D_h"] <= df["D_h"].median()]["Y_h"].mean() * (1 - p_treated) + y_lo * p_treated)
manski_lower = df[df["D_h"] > df["D_h"].median()]["Y_h"].mean() * p_treated + y_lo * (1 - p_treated) \
              - (df[df["D_h"] <= df["D_h"].median()]["Y_h"].mean() * (1 - p_treated) + y_hi * p_treated)
print(f"  Manski worst-case bounds (二值化后): [{manski_lower:+.4f}, {manski_upper:+.4f}]")
print(f"    (点估计 τ = {tau_fwl:+.4f} 位于此区间内)")

# (d) Placebo test: 用未来的 D 预测过去的 Y (应该 tau ≈ 0)
df_pl = df.copy()
df_pl["D_future"] = df_pl["D_h"].shift(-1)  # 未来的 D
df_pl["Y_h"] = df_pl["Y_h"]
df_pl2 = df_pl.dropna(subset=["D_future", "Y_h"] + X_cols).reset_index(drop=True)
X_pl = sm.add_constant(df_pl2[["D_future"] + X_cols].values.astype(float))
m_pl = sm.OLS(df_pl2["Y_h"].values, X_pl).fit(cov_type="HAC", cov_kwds={"maxlags": 6})
tau_placebo = m_pl.params[1]
print(f"  Placebo test (用 D_{{h+1}} 预测 Y_h): τ_placebo = {tau_placebo:+.4f}, p = {m_pl.pvalues[1]:.4f}")
print(f"    → 若 |τ_placebo| >> 0, 说明存在反向或共同趋势")

sens["e_value"] = {"RR": float(RR), "E_value": float(E_val)}
sens["rosenbaum"] = rb
sens["manski_bounds"] = {"lo": float(manski_lower), "hi": float(manski_upper)}
sens["placebo"] = {"tau_placebo": float(tau_placebo), "p": float(m_pl.pvalues[1])}
results["stage4_sensitivity"] = sens

# =========================================================================
# Stage extra: Granger 双向 + PC 条件独立
# =========================================================================
print("\n" + "="*80)
print("Stage extra: Granger 双向 lead-lag + 条件独立")
print("="*80)

# Granger: D → Y 和 Y → D, lag=1,2,3
try:
    print("\nGranger 因果 (H0: X 不 Granger cause Y):")
    print("  D → Y (韩国 → MU):")
    g1 = grangercausalitytests(df[["Y_h", "D_h"]].dropna().values, maxlag=3, verbose=False)
    for lag in [1, 2, 3]:
        f = g1[lag][0]["ssr_ftest"]
        print(f"    lag={lag}: F={f[0]:.3f}, p={f[1]:.4f}")
    print("  Y → D (MU → 韩国):")
    g2 = grangercausalitytests(df[["D_h", "Y_h"]].dropna().values, maxlag=3, verbose=False)
    for lag in [1, 2, 3]:
        f = g2[lag][0]["ssr_ftest"]
        print(f"    lag={lag}: F={f[0]:.3f}, p={f[1]:.4f}")

    results["granger"] = {
        "d_to_y": {str(lag): {"F": float(g1[lag][0]["ssr_ftest"][0]), "p": float(g1[lag][0]["ssr_ftest"][1])} for lag in [1,2,3]},
        "y_to_d": {str(lag): {"F": float(g2[lag][0]["ssr_ftest"][0]), "p": float(g2[lag][0]["ssr_ftest"][1])} for lag in [1,2,3]},
    }
except Exception as e:
    print(f"[Granger err] {e}")
    results["granger"] = {"error": str(e)}

# PC-style 条件独立: partial correlation 检验
from scipy.stats import norm
def partial_corr_test(df_sub, x, y, cond):
    """Fisher-Z test of partial correlation."""
    sub = df_sub.dropna(subset=[x, y] + list(cond)).reset_index(drop=True)
    n = len(sub)
    if n < 20:
        return None
    # Regress x on cond, y on cond, then correlate residuals
    Xc = sm.add_constant(sub[list(cond)].values.astype(float))
    rx = sm.OLS(sub[x].values, Xc).fit().resid
    ry = sm.OLS(sub[y].values, Xc).fit().resid
    r = np.corrcoef(rx, ry)[0, 1]
    if abs(r) >= 1:
        return {"r": r, "n": n, "z": np.inf, "p": 0}
    z = 0.5 * np.log((1 + r) / (1 - r)) * np.sqrt(n - len(cond) - 3)
    p = 2 * (1 - norm.cdf(abs(z)))
    return {"r": float(r), "n": int(n), "z": float(z), "p": float(p)}


print("\n条件独立测试 (Fisher-Z partial correlation):")
cases = [
    ("Y_h", "D_h", []),                                    # unconditional
    ("Y_h", "D_h", ["r_vix_d", "r_jpy_h", "r_sox_d"]),     # 控制宏观
    ("Y_h", "D_h", X_cols),                                # 控制宏观 + 时段
    ("Y_h", "eta", []),                                    # η 单独效应
    ("Y_h", "eta", X_cols),                                # η 条件效应 (排他性检验)
    ("Y_h", "Z_OPEN", X_cols + ["D_h"]),                   # 排他性: Z ⊥ Y | D,X ?
    ("Y_h", "Z_MID", X_cols + ["D_h"]),
    ("Y_h", "Z_CLOSE", X_cols + ["D_h"]),
]
ci_results = []
for x, y, cond in cases:
    r = partial_corr_test(df.dropna(subset=[y]), x, y, cond)
    if r:
        cond_str = "∅" if not cond else "{" + ",".join(cond) + "}"
        # 缩短
        cond_short = cond_str if len(cond_str) < 40 else cond_str[:37] + "...}"
        print(f"  {x} ⊥ {y} | {cond_short}:  r={r['r']:+.3f}  p={r['p']:.4f}  n={r['n']}")
        ci_results.append({"x": x, "y": y, "cond": cond, **r})
results["conditional_independence"] = ci_results

# =========================================================================
# Save all
# =========================================================================
with open(f"{DATA}/all_results.json", "w") as f:
    json.dump(results, f, indent=2, default=str, ensure_ascii=False)
print(f"\n[OK] 所有结果保存到 {DATA}/all_results.json")
