# 08. Alternatives Not Taken（未采用的替代方案）

> 本文档记录**考虑过但未实施**的方案，帮助后续 AI 或研究者判断扩展方向。每个方案标注：*理由*、*工作量*、*预期收益*。

## 1. 数据层面的替代方案

### 1.1 扩展窗口到 200+ 天

**当前**：76 天（2026-04-11 到 2026-07-01）
**替代**：拉取 2025-01-01 起 18 个月数据（约 380 交易日）
**理由**：Bitget MUUSDT 上线时间限制了历史长度
**未采用原因**：
- MUUSDT 于 2026-03 上线，之前无数据
- 韩国半导体股票在 2025 有中国出口管制冲击，可能违反 SUTVA
**工作量**：中（需重新拉取所有源）
**预期收益**：
- 更多 IV 变异（多 20+ 个 Korean holiday × US open 组合）
- 潜在结构断点，可做 event-study

### 1.2 MU 现货（Yahoo）作为 Y 而非 Bitget MUUSDT

**当前**：Bitget MUUSDT 24/7 永续合约
**替代**：Yahoo `MU` 1h 现货
**未采用原因**：
- Yahoo 现货只在 US 交易时段有数据（14:30-21:00 UTC）
- 无法捕捉 overnight 韩国 session 的影响
- Bitget 提供全天候价格发现
**工作量**：低（数据已有）
**预期收益**：
- 排除加密溢价（basis premium）的影响
- **重要 robustness check**（推荐后续做）

### 1.3 MUUSDT basis 分析（现货 vs 期货基差）

**未做**：Bitget spot vs Bitget perpetual basis
**理由**：
- Basis 可能反映资金费率、市场情绪，与 causal treatment 混杂
- Bitget spot 数据可能与 perpetual 差异过大
**扩展方向**：分离 "spot channel" vs "perp channel"，看哪个更快响应 Korea 冲击

### 1.4 分层 mediator（M = f(EWY, KORU)）

**当前**：EWY 和 KORU 分别做 mediation
**替代**：组合 mediator，如 M = w_1·EWY + w_2·KORU（PCA 或最优权重）
**未采用原因**：
- 两个 ETF 高度相关（KORU 是 EWY 的 3× 杠杆版）
- 分别做更清晰
**扩展方向**：如果关心"总 ETF 通道"，可以用 factor model

## 2. 计量方法的替代方案

### 2.1 PC 算法完整版

**当前**：只做了部分条件独立检验作为 sanity check
**替代**：完整 PC/PC-stable/GES 算法自动学习 DAG
**未采用原因**：
- 用户已有先验 DAG（潜在结果框架先行）
- PC 在小样本 + 时间序列上性能弱
- 输出的 DAG 需要人工解释
**扩展工具**：`causal-learn`, `pgmpy`, `tigramite`（时间序列专用）
**推荐**：可用 `tigramite` 的 PCMCI 算法验证 Y→D 反向因果

### 2.2 Event Study for MU Earnings

**当前**：earnings 只作为 covariate 排除
**替代**：以 MU 财报日为 event，看 abnormal returns
**未采用原因**：
- 76 天窗口只有 1 次 MU 财报（2026-06-24）
- 样本太少无法做 event study 统计推断
**扩展方向**：扩展窗口后可以做

### 2.3 KORU 非线性 Mediation

**当前**：Mediation 用线性 a-path 和 b-path
**替代**：KORU 是 3× 杠杆 ETF，可能有非线性响应（vol drag）
**方法**：
- Semiparametric mediation (Tchetgen & Shpitser 2012)
- Machine learning mediation (double ML)
**未采用原因**：
- 小样本下 ML mediation 方差过大
- 线性近似在 daily returns 尺度已足够

### 2.4 Rolling / Time-varying CACE

**当前**：全样本单一 CACE 估计
**替代**：滚动窗口（rolling 30 天）估计时变 CACE
**未采用原因**：
- 76 天数据无法支持 rolling
- 每个 rolling 窗口 IV 会更弱
**扩展方向**：扩展窗口后可做，判断因果关系是否稳定

### 2.5 Bayesian 因果推断

**当前**：Frequentist（Newey-West SE、bootstrap CI）
**替代**：Bayesian 版本
- Bayesian IV（Rossi 2012）
- Bayesian mediation with priors on NIE/NDE
- Bayesian sensitivity for unmeasured confounding
**未采用原因**：
- 用户明确对因果推断感兴趣但没要求 Bayesian
- Frequentist 结果已经比较清晰
**扩展方向**：用户对信息几何/贝叶斯有兴趣，可作为学术扩展；用 PyMC 或 Stan
**特别推荐**：**Bayesian sensitivity analysis** ——把 unmeasured confounding 参数化为先验，比 E-value 更 informative

### 2.6 Synthetic Control / DiD

**未采用**：Korean semiconductor shock 事件太多，无法定义清晰"treatment date"
**扩展方向**：如果只关心某一次大事件（如 2026-05 出口管制新闻）

### 2.7 反向因果的形式化处理

**发现**：Y→D 显著（F=93.8），D→Y 不显著
**当前**：placebo test 确认，但只是**报告结果**
**替代**：把假设倒过来，做 "MU → Korea" 的 CACE
- 找 US 侧的工具变量（如 MU 财报日、FOMC 会议）
- 估计 US → Korea 的 causal effect
**推荐**：**这可能是本项目最有价值的扩展**——重新做 identification，方向反过来

## 3. 因果框架的替代方案

### 3.1 SCM (Structural Causal Model) 完整版

**当前**：Rubin potential outcomes
**替代**：Pearl SCM，写出所有结构方程
```
D = f_D(η, U_D)
M = f_M(D, η, U_M)
Y = f_Y(D, M, η, U_Y)
```
**未采用原因**：
- Rubin 框架足够识别 ATE/CACE/NDE/NIE
- SCM 需要更多假设（functional form）
**扩展方向**：写完整 SCM 有助于 counterfactual query，但对本项目 marginal effect 不大

### 3.2 Front-door Criterion

**未采用**：Front-door 要求 mediator 完全阻断 D→Y 的直接路径
**理由**：
- EWY/KORU 不完全 block（存在 X 通过其他 ETF 或宏观通道影响 Y）
- 我们用 IV + back-door 组合更合适

### 3.3 Do-Calculus 派生 identifiability

**未采用**：用户已有明确 DAG，不需要 do-calculus 自动推导
**扩展方向**：用 `dagitty` 或 `DoWhy` 验证 identification 假设

### 3.4 Interventional vs Counterfactual estimand

**当前**：Interventional ATE / CACE
**替代**：
- E[Y_i(1) - Y_i(0) | D=1] (ATT)
- E[Y_i(1) - Y_i(0) | Y=y_obs] (counterfactual)
**未采用原因**：ATT 需要不同 IV，本项目 IV design 更适合 ITT/CACE

## 4. 敏感性分析的替代方案

### 4.1 Rosenbaum bounds

**当前**：E-value = 11.12
**替代**：Rosenbaum Γ-sensitivity（更 rigorous for matching-based inference）
**未采用原因**：
- 本项目非 matching design（是 IV + regression）
- E-value 已足够 informative
**部分实施**：Manski bounds [-0.047, +0.050] 已提供

### 4.2 Negative Control Outcome

**未采用**：找一个"不应受 η 影响"的 outcome（如 TSLA、AAPL 等无半导体关系股）作为 negative control
**扩展方向**：若 negative control 也显示 τ>0，说明混杂严重

### 4.3 Instrumental Variable Falsification

**未做**：假设 η 通过 exclusion restriction 只影响 D 不影响 Y。可以检验：
- η vs 前期 Y 的相关性（应该为 0）
- η vs 无关 outcome（如 gold price）的相关性
**扩展方向**：加强 IV 假设的经验支持

## 5. 呈现层的替代方案

### 5.1 Interactive Dashboard

**当前**：静态 Markdown 报告
**替代**：Streamlit / Dash / Observable notebook
**扩展方向**：
- 时间序列交互图（Y_h, D_h, η_h overlaid）
- Bootstrap 分布 histogram
- Mermaid DAG hover 显示效应大小

### 5.2 LaTeX 论文格式

**未做**：把项目结论写成正式论文（AER-style）
**扩展方向**：如果用户想投稿或作为 thesis chapter

### 5.3 Reproducible Notebook

**当前**：Python scripts
**替代**：Jupyter/Quarto notebook（inline output）
**扩展方向**：`quarto render` 一键生成 HTML/PDF

## 6. 后续推荐优先级

**High priority**（下一步应做）：
1. ⭐ **反向 identification**：重新做 MU → Korea 的 CACE（承认 Y→D 主导）
2. ⭐ **扩展窗口到 200+ 天**（用 Yahoo MU 代替 Bitget MUUSDT）
3. ⭐ **Negative control outcome**（验证 IV exclusion restriction）

**Medium priority**：
4. Bayesian sensitivity analysis
5. Rolling CACE（判断因果稳定性）
6. Tigramite PCMCI 时间序列 DAG 学习

**Low priority**（学术兴趣）：
7. 完整 SCM + counterfactual query
8. Semiparametric mediation
9. Interactive dashboard

