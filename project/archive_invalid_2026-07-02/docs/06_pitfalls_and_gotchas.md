# 06. Pitfalls & Gotchas（踩坑指南）

> 本文档汇总本项目实际踩过的坑，帮助后续 AI 或研究者避免重蹈覆辙。所有条目均为**真实遇到的问题**及其解决方案。

## 1. 计量经济学库 API

### 1.1 `statsmodels.regression.linear_model.IV2SLS` 已移除

**症状**：
```python
from statsmodels.regression.linear_model import IV2SLS
# ImportError: cannot import name 'IV2SLS'
```

**原因**：statsmodels ≥ 0.13 已将 IV2SLS 移除，官方推荐迁移到 `linearmodels`。

**解决**：
```python
from linearmodels.iv import IV2SLS
# 注意：linearmodels 的 IV2SLS 参数顺序与旧版不同
# IV2SLS(dependent, exog, endog, instruments)
```

### 1.2 `linearmodels` 首阶段诊断 API 变化

**症状**：
```python
first_stage = model.first_stage
J_stat = first_stage.diagnostics['sargan']  # AttributeError / KeyError
```

**原因**：`linearmodels` v5+ 将 J-statistic 移到 `results.j_stat`，且诊断字典结构变化。

**解决**：手动计算 Hansen J：
```python
u_hat = results.resids  # 结构方程残差
# J = n × R²(u_hat 对 Z 的回归)
sm_ols = sm.OLS(u_hat, sm.add_constant(Z)).fit()
J = len(u_hat) * sm_ols.rsquared
p_J = 1 - stats.chi2.cdf(J, df=Z.shape[1] - 1)
```

### 1.3 `wald_test` API 需要 `scalar=True`

**症状**：
```python
wald = results.wald_test(R)
print(wald.statistic)  # DeprecationWarning: 返回矩阵而非标量
```

**解决**：
```python
wald = results.wald_test(R, scalar=True)
F_stat = float(wald.statistic)
```

### 1.4 `linearmodels` 外生变量与工具变量共线

**症状**：`AbsorbingEffectError` 或 `SingularMatrixError`

**触发场景**：把 `I_MID`, `I_CLOSE` 同时放到 `exog` 和 `instruments` 中（因为 η × I_MID 是工具变量，则 I_MID 是它的组成部分）。

**解决**：**从 exog 中移除 session dummies**，只保留在 instruments 侧：
```python
# ❌ 错误
IV2SLS(Y, exog=[const, I_MID, I_CLOSE], endog=[D], 
       instruments=[eta*I_OPEN, eta*I_MID, eta*I_CLOSE])

# ✅ 正确
IV2SLS(Y, exog=[const], endog=[D], 
       instruments=[eta*I_OPEN, eta*I_MID, eta*I_CLOSE])
```

## 2. Yahoo Finance 数据

### 2.1 Yahoo `scale=3` 缩放

**症状**：MU/SNDK 的 1h `close` 价格看起来是真实价格的 1/1000。

**原因**：Yahoo Chart API 对高价股（>$100）返回 `meta.priceHint=3` 或 `scale=3`，需要手动乘 10^3？实际上是 pips 表示。

**解决**：**改用对数收益率**——对 scale 免疫：
```python
ret = np.log(close).diff()  # scale 是常数因子，diff 后消掉
```

### 2.2 Yahoo `hasPrePostMarketData=false` for 韩股

**症状**：韩股 005930.KS 请求 `queryOptions.includePrePost=true` 时无盘前数据。

**原因**：韩国交易所在 Yahoo 数据源中没有盘前盘后（unlike 美股）。

**解决**：在数据 QC 阶段跳过韩股盘前请求；同时也说明 IV 设计需以美国 session 为基准（韩股的"盘前 KORU"其实是 ETF 在 NY 的盘前）。

### 2.3 部分小时缺失

**症状**：MU 1h 数据在某些日子的第一/最后一小时缺失（如 09:30 或 15:30 的 partial bar）。

**解决**：QC 脚本 `qc_manifest.py` 检测并 forward-fill 或 drop；建模时用 `pd.merge_asof` 对齐时间戳。

## 3. FRED / 宏观数据

### 3.1 SSL 证书问题

**症状**：
```python
requests.get("https://api.stlouisfed.org/...", ...)
# SSLError: CERTIFICATE_VERIFY_FAILED
```

**原因**：Perplexity Sandbox 环境的内部 HTTPS 代理需要自定义 CA。

**解决**：使用 `curl` 而非 `requests`，并指定内部 CA：
```bash
curl --cacert /etc/ssl/certs/agent-proxy-ca-2.pem \
     "https://api.stlouisfed.org/fred/series/observations?series_id=DGS10&..."
```

或在 Python 中：
```python
import os
os.environ['SSL_CERT_FILE'] = '/etc/ssl/certs/agent-proxy-ca-2.pem'
```

### 3.2 TEDRATE 已被废弃

**症状**：FRED API 返回 `TEDRATE` 序列有效，但数据在 2022 年后停止更新。

**原因**：FRED 已停止发布 TED spread（TED = 3M LIBOR - 3M T-bill），因为 LIBOR 于 2023 年被 SOFR 取代。

**解决**：改用 `SOFR - DGS3MO` 或直接跳过 TED spread，因为本项目主要用 VIX / DGS10 / DFF 作为宏观 covariate。

## 4. 工具变量设计的教训

### 4.1 日频 IV 太弱

**首次尝试**：用 "韩国休市 × 美国开市" 作为日频 IV（KP-F ≈ 4.57）。

**问题**：76 天窗口中只有 **3 天** 韩休美开的组合 → 变异太少，IV 严重弱。

**解决**：升级到**小时级 3-D IV**：η × [I_OPEN, I_MID, I_CLOSE]（session dummies）。这样每天有 3 个小时级 session 提供变异，KP-F 从 4.57 → **951.97**。

### 4.2 单一 session IV 也弱

**方案 A**：只用 `η × I_OPEN` 作为唯一 IV。**KP-F = 4.57**（弱 IV 阈值 10）。

**原因**：只有 22.5 hr 数据点被 I_OPEN 激活；且 β_OPEN=0.485 虽有 R²=22%，但 leverage 集中在少数极端 η 值。

**解决**：三段联合识别（方案 C）。同时 Hansen J 变得有意义，可检验过度识别限制。

### 4.3 内生性验证：η 是否真的外生？

**风险**：η_h（Korean 半导体加权收益）可能与 U_h（未观测混杂：全球风险偏好）相关。

**部分缓解**：
- 控制 VIX_1h、SOX_1h、USDJPY_1h、DGS10 → partialling-out
- Placebo test（下节讨论）

**根本无法解决**：如果 U_h 通过 SOX（费城半导体指数，包含 MU + 韩国 ADR）影响 MU，则 η 与 U 相关，IV 不外生。这是本项目**主要 identification threat**。

## 5. 数据对齐与时区

### 5.1 时区混乱

**坑**：
- Bitget: UTC
- Yahoo (MU/SNDK): US/Eastern
- 韩国交易所: Asia/Seoul
- FRED: US/Central

**解决**：**全部统一到 UTC**，用 `pd.Timestamp.tz_convert("UTC")`。所有 `session_h` 判断以 UTC 为基准（US Open = 13:30 UTC 冬令时 / 14:30 夏令时，需注意 DST）。

### 5.2 韩国节假日 vs 美国节假日

**关键**：IV design 依赖"韩国休市 vs 美国开市"的时间错配。需分别构造两个 holiday calendar：
- 韩国：`korea_holidays.parquet`（KOSPI 官方）
- 美国：内置 `pandas.tseries.holiday.USFederalHolidayCalendar`

## 6. 因果推断方法学的坑

### 6.1 反向识别（最大发现）

**惊人结果**：Granger F 检验显示 **Y → D 显著（F=93.8）**，而 **D → Y 不显著（F=1.09）**。

**含义**：Bitget MUUSDT（24/7 交易的 MU 永续合约）**领先于**韩国半导体现货，而不是相反。

**为什么**：
- Bitget 是 24/7 CEX，MU 价格信号先于韩国现货市场
- 韩国现货交易时段有限（09:00-15:30 KST），无法即时反映全球定价
- 全球定价通道：美国期货 → Bitget 永续 → 韩国现货

**处理**：
1. 明确报告"MU→Korea"而非原假设"Korea→MU"
2. Placebo test（用 D_{h+1} 预测 Y_h）→ τ=+0.219, p<0.0001，**证实反向因果**
3. 不再声称"Korean overnight causes MU"，而是"数据表明信息流向相反"

### 6.2 CACE ≈ 0 是"零效应"还是"弱识别"

**τ_CACE = +0.018, SE = 0.035, KP-F = 951.97**

**判断**：
- KP-F 远超弱 IV 阈值 → **强 IV**
- AR 95% CI [-0.049, +0.083] → **对 IV 弱性稳健**
- Hansen J p-value = NaN（linearmodels API 问题，需手动）

**结论**：τ ≈ 0 是**真实的零效应**，不是识别失败。但**方向要修正**（是 MU→Korea，不是 Korea→MU 的零效应）。

### 6.3 Mediation 结果解读

- **NIE via EWY = +0.021 (bootstrap CI [+0.001, +0.049])** → 通过 EWY 通道，Korea 对 MU 有微弱正向间接效应
- **NDE = -0.006** → 直接效应为零/微负
- **总效应 ≈ NIE + NDE ≈ +0.015**

**注意**：即使反向因果为主，Mediation 分析仍有部分意义：可以解释为"Korea 通过 ETF 通道对 MU 有次要影响，主导方向仍是 MU→Korea"。

## 7. 实现细节

### 7.1 Bitget API 无需认证但有速率限制

- 公共 REST：`https://api.bitget.com/api/v2/mix/market/history-candles`
- 每分钟 20 次调用
- 单次最多返回 1000 根 K 线
- 分页：用 `endTime` 参数向前回溯

### 7.2 Newey-West 滞后阶数选择

- 小时频数据 + 每日周期性 → 建议 `maxlags = 24`（一天）
- 更保守：`maxlags = int(4 * (T/100)^(2/9))` (Newey-West 1994 推荐)
- 本项目 T=302 → maxlags ≈ 5，但用 24 更稳健

### 7.3 Bootstrap CI for NIE/NDE

- **必须用 block bootstrap**（避免破坏时间序列自相关）
- Block size = 24（1 天）
- B = 1000 次重抽样
- 存 seed 便于复现：`rng = np.random.default_rng(42)`

## 8. 可复现性坑

### 8.1 环境依赖版本

必须固定的版本：
- `linearmodels >= 5.0`（API 大变化）
- `statsmodels >= 0.14`
- `pandas >= 2.0`（时区处理）
- `numpy < 2.0`（scipy 兼容）

已导出到 `project/requirements.txt`（待生成）。

### 8.2 数据快照

**警告**：Bitget/Yahoo API 数据会**回填修正**——同一时段过几天再拉可能不完全一样。

**解决**：保存原始 parquet 快照，`data/manifest.json` 记录拉取时间。

