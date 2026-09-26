# 预测市场错价消退 / 均值回归文献速览（Polybot 向）

检索窗口：主搜 2025-09～2026-09；经典可至 2024-09。工具：`academic-mcp` 的 `academic_search` + `paper_detail`（Semantic Scholar 一度限流，部分用 arXiv/OpenAlex/Crossref 补摘要）。主题：**二元合约 / 短周期 CLOB 上，相对“公平概率”的错价是否会回落**——不是现货上「Binance 领先 Chainlink 就追 lag」那类跨源延迟套利。

Polybot 语境：用模型给出 BTC/ETH 5m/15m 的**理论价**，相对 Polymarket 盘口有 dislocation 时考虑 **fade**；决策层只做 skip / observe / pause，且必须 **fee-aware**。

---

## 负对照（别和 fade-dislocation 混为一谈）

- **Kitron & Wengrowicz (2026)** — [https://arxiv.org/abs/2608.21888](https://arxiv.org/abs/2608.21888)  
  15 分钟 **现货**方向性均值回归在 Binance 上统计上很强，但毛利约 **1.3 bp/笔**，低于约 **5 bp** 往返成本；和「Polymarket 合约价回到模型公平概率」不是同一条腿。用作提醒：**短 horizon 信号可以“存在但不可交易”**，fade gate 必须把手续费、价差、深度和 round-trip 算进去。

---

## 精选（最多 8 篇；均读过摘要或 `paper_detail`，不止标题）

### 1. Fair Value Pricing in Short Dated Bitcoin Binary Markets（Semenas, 2026）

- **URL**：https://doi.org/10.2139/ssrn.7134801  
- **说了什么**：把短期 BTC 二元合约当成 **drift 可忽略的 cash-or-nothing digital**，公平价只依赖现价 + 短窗波动；当市价与 fair value 差超过 buffer 才入场。22–23 Jun 2026 共 182 个已结算合约：胜率 67.6% vs 按均价 0.63 隐含的 63% 盈亏平衡，点估计 edge +4.6pp，但样本小、bootstrap 含零、fair value 并不优于市价做概率预测，且收益集中在少数交易。  
- **和 fade gate**：和 Polybot「理论 vs 市场」同构——**dislocation 阈值 + 波动定价**；作者自己强调单窗、未证显著，适合当 **正向模板 + 过拟合警示**，不是已验证 alpha。  
- **为什么不是 junk**：SSRN 工作论文、机制清晰、诚实报局限；直接点名 favorite–longshot 等开放问题。  
- **评分**：**8/10**（场景贴合；证据仍弱）

### 2. A Fee-Aware Break-Even Framework for Binary Prediction Markets（Charre, 2026）

- **URL**：https://doi.org/10.2139/ssrn.7186118  
- **说了什么**：纯解析、无回测：Polymarket 二次 taker 费 \(\phi = C r p(1-p)\) 下，持有到期的 **概率 edge 下限** 为 \(r p(1-p)\)，在 \(p=0.5\) 最大为 \(r/4\)；相对 deployed capital 为 \(r(1-p)\)，**低价 longshot 区 hurdle 更高**；短周期 round-trip 近似 **双倍** 单腿 hurdle。不含 spread、滑点、队列。  
- **和 fade gate**：fade 前应先算 **|model_prob − market| 是否超过确定性费用地板**；在 0.5 附近和低价区策略要不同阈值。和 Polybot「只 observe/pause」一致：raw edge ≠ tradable edge。  
- **为什么不是 junk**：不吹收益；公式可嵌入 gate，和平台公开费率对齐。  
- **评分**：**9/10**（工程必备，非经验 MR 论文）

### 3. The Anatomy of a Decentralized Prediction Market（Dubach, 2026）

- **URL**：https://arxiv.org/abs/2604.24366  
- **说了什么**：30B 档 CLOB 事件 + 链上成交对齐；longshot **spread premium**、深度形态、feed 推断成交方向与链上仅 ~59% 一致等。强调微观结构研究必须用 **OrderFilled** 定方向。  
- **和 fade gate**：说明 **dislocation 可能是 spread/深度假象**；fade 信号应结合有效半价差、可成交量；公共 feed 单独做 Lee-Ready 类推断不可靠。  
- **为什么不是 junk**：预注册面板、可复制包；直接针对 Polymarket CLOB。  
- **评分**：**8/10**（基础设施；不直接给 MR 策略）

### 4. Decomposing Crowd Wisdom: Domain-Specific Calibration（Le, 2026）

- **URL**：https://arxiv.org/abs/2602.19520  
- **说了什么**：Kalshi + Polymarket 上 3.53 亿笔交易；**校准随领域、距到期、规模变化**；政治市场系统性 **underconfidence（价格向 50% 压缩）** 在 Polymarket 可复制。  
- **和 fade gate**：「模型公平概率」若把 **市价当无偏概率** 会错；fade 应 **按品类/剩余时间做 recalibration**，否则会把结构性压缩当成可 fade 的 dislocation。  
- **为什么不是 junk**：大规模、双平台；解释的是条件概率而非炒作。  
- **评分**：**7/10**（偏校准而非短窗 MR，但对 gate 极重要）

### 5. The Favorite–Longshot Bias in Prediction Markets（Cardozo & Rivero-Wildemauwe, 2026）

- **URL**：https://arxiv.org/abs/2609.12878  
- **说了什么**：Polymarket 5.88 亿笔交易；低价买入长期亏损、高价端略赚；**Crypto / Politics 有两端偏差，Sports 不明显**；按 event 聚合会改变结论幅度。  
- **和 fade gate**：系统性 **价位依赖的误定价**；若模型在极端概率区和市场对抗，可能是在 **吃 FLB 错误一侧** 或与之对齐——crypto 类 up/down 要和 FLB 文献对照。  
- **为什么不是 junk**：全平台账户级、分品类；和 Semenas 文互证。  
- **评分**：**7/10**（慢变量偏差，非秒级 fade；但定方向）

### 6. Volatility in Prediction Markets: A Structural Approach（2026）

- **URL**：https://arxiv.org/abs/2607.08199  
- **说了什么**：为 **有到期二元合约** 建波动模型：Wright–Fisher 截止分解 + Glosten–Milgrom 订单流；Kalshi 大样本上结构项压过纯 ARCH/GARCH；波动在 **50¢ 附近、临近 resolution** 更高。  
- **和 fade gate**：给 **理论价动态和“还剩多少可消的 dislocation”**；体育更像 jump，经济类更像平滑 deadline——BTC 5m/15m 更接近哪一类要实证。  
- **为什么不是 junk**：结构可解释、OOS 比较；不是黑箱预测。  
- **评分**：**8/10**（理论核；Kalshi 为主，需外推到 Polymarket crypto）

### 7. Arbitrage Analysis in Polymarket NBA Markets（Cheng et al., 2026）

- **URL**：https://arxiv.org/abs/2605.00864  
- **说了什么**：7500 万 LOB 快照；单市场可执行套利极少（173 场仅 7 次，中位持续 **3.6s**）；组合套利多但 **深度瓶颈**（多数机会 ~14.8 shares）。  
- **和 fade gate**：**错价存在但极短、极小规模**；fade 若指望反复吃固定 dislocation，要和 **持续时间与可成交量** 对齐，否则 observe/pause 更合理。  
- **为什么不是 junk**：高频 LOB 重建；结论偏「效率高」而非营销。  
- **评分**：**6/10**（NBA 非 crypto up/down，但 CLOB 纪律可迁移）

### 8. Executable Arbitrage and Market Efficiency in Prediction Markets（Gebele et al., 2026）

- **URL**：https://arxiv.org/abs/2608.00666  
- **说了什么**：区分 **payoff 无套利** vs **协议可执行无套利**；NegRisk 仅部分方向可转换，CLOB 上违反多集中在 **不可执行一侧**；深度感知组合价值 + 链上轨迹估套利利润 ~$1.12M。  
- **和 fade gate**：Polybot 的 theory-vs-market fade **不是** combinatorial / converter 套利；此文说明 **很多“看起来错价”在结算前锁死**。gate 应排除需特殊协议路径的 dislocation。  
- **为什么不是 junk**：机制级、Polymarket 专用、可执行性定义严谨。  
- **评分**：**7/10**（套利非 MR，但划清边界）

---

## 未纳入主表（dropped：营销味 / 现货 ML alpha / 错题）

| 题名 / 线索 | 丢弃原因 |
|-------------|----------|
| PolySwarm 等多智能体 LLM「预测市场 + latency arbitrage」类 (如 arXiv:2604.03888) | Agent 叙事、latency 套利，贴近用户明确不要的 cross-venue lag |
| OpenMarket：Polymarket–Binance 同步数据集 (arXiv:2607.26245) | 数据工程有用，但不论证 MR/fade |
| MoFE 等 crypto 价格预测 + 模拟盘高 Sharpe (arXiv:2608.17342) | 现货 ML alpha 话术 |
| Binary Tree + Random Forest 分钟级 SPY 方向 (arXiv:2507.16701) | 传统标的微结构 ML，非预测市场二元 |
| Bitcoin power law 估值 + 分位交易 (2026 Pressacademia) | 长周期现货叙事，非 CLOB 概率 fade |
| Do Prediction Markets Match Option Prices? Binance–Polymarket (arXiv:2606.19517) | 阈值期权 **跨市场定价一致性**，不是 intra-Polymarket dislocation 消退 |
| Kalshi NBA underreaction / hot-hand (JBEF 2026) | `paper_detail` 无摘要可读；体育慢信息，与 5m crypto 合约距离大 |
| Toward Black Scholes for Prediction Markets (arXiv:2510.15205) | 偏做市/belief vol 工具书，未读全文前不抢 Semenas/结构 vol 的坑位——可作后续扩展 |

---

## 对 Polybot gate 的一句话收束

文献共识大致是：**短窗上“错价”要么被费用吃掉（现货 15m 负对照 + Polymarket 费率解析），要么存在但受校准偏差、FLB、深度与可执行性约束**；可交易的 fade 更像 **带 buffer 的 fair-value 偏离 + 明确 round-trip 成本地板**，且 NBA/CLOB 经验提示 **窗口常以秒计、规模很小**。模型层继续 **skip / observe / pause** 与这些证据一致，而不是默认每笔 dislocation 都可 fade。

*检索日期：2026-09-25。*
