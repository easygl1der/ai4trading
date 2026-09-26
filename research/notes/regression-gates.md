# 回归 / 校准当「门」：文献速写（2025-09～2026-09）

**检索**：academic-mcp（OpenAlex + arXiv；Semantic Scholar 当次 429 限流）。  
**窗口**：首发/上线约 2025-09 至 2026-09。  
**Polybot 对照**：已有 `p_theory`（GBM）、`p_market`；`imbalance = p_theory − p_market`；后续可用经典 logistic / GBM **否决**坏单，**不**用回归直接报价。Jev 树 **不输出价格**。

**评分**：1–10 = 与「概率校准 / 公平价值偏差 + 交易门控（veto）」的贴合度，不是 alpha 宣传分。

---

## 保留（≤8）

### 1. Portnaya — Polymarket Yes 价 vs 期权隐含二元公平价（**10**）

- **URL**：https://arxiv.org/abs/2606.19517  
- **回归在干什么**：逐小时匹配 Polymarket 阈值合约与 Binance 同标的期权，构造 **定价缺口（gap）**；用 **横截面回归** 看缺口与期权隐含概率、期限等的关系；时间序列上还有 AR(1) 半衰期。  
- **目标是否偷看未来**：缺口用 **当时** 的市价与期权曲面；结算结果只用于经济解释，不是回归里的因变量。  
- **和 veto 的关系**：本质是 **双源概率楔子**——和 Polybot 的 `p_market` vs `p_theory` 同构；回归描述楔子形态，**不**给下单价；可支撑「楔子过大/结构异常 → 不交易」。  
- **为何不是 junk**：可复核的跨市场基准、HAC/bootstrap；明确讲分段与投机需求，不是「AI 战胜市场」。

### 2. Le — 预测市场分域 logistic 再校准斜率（**9**，邻域文献仅一笔）

- **URL**：https://arxiv.org/abs/2602.19520（详见 `papers.md` 长摘要，此处不重复）  
- **回归在干什么**：按领域 / 距结算 / 成交规模分格，对 **成交价 → 实现频率** 做 **logistic 再校准斜率**；再用分解解释格间差异。  
- **目标是否偷看未来**：实现结果在合约结算后才知道；分格估计若严格按 **过去已结算** 样本滚动则可用作门；全文主分析是 **描述性大样本**，部署时要自己切成 walk-forward。  
- **和 veto 的关系**：告诉你 **raw `p_market` 在该格是否系统性偏高/偏低**——适合做成 **校准层或否决**（例如政治盘压缩向 50%），**不是**价格预测器。  
- **为何不是 junk**：3.53 亿笔、Kalshi+Polymarket；crypto 短周期格上斜率近 1 与你们观测一致。

### 3. Hao, Chen & Qi — 短周期 crypto 事件合约 + 自适应分位数门（**8**）

- **URL**：https://doi.org/10.3390/a19080704  
- **回归在干什么**：GBM 集成先打出 **方向胜率分数**；再用 **过去 14–28 日分数的日度自适应分位数阈值**（coverage 约束）决定 **是否出手**；并和固定校准窗、ACI 等对比。  
- **目标是否偷看未来**：阈值只用 **过去分数**；标签是合约结算方向，在 **冻结配置 + 伪 OOS + 2026 上半年 holdout** 协议下评估。  
- **和 veto 的关系**：典型 **选择性预测 / 分位数门**——模型分数过线才交易，**不**改报价；和「logistic veto 坏单」同一层。  
- **为何不是 junk**：明确 coverage 运营带、多重检验、承认只是 PoC；但底层仍是 ML 特征，**不能**当「crypto 可预测」证据。

### 4. Dalen — 预测市场 logit 跳跃扩散 + 校准管线（**8**）

- **URL**：https://arxiv.org/abs/2510.15205  
- **回归在干什么**：把成交价 \(p_t\) 当 **Q-鞅**；EM 分离扩散与跳跃，估 **belief volatility** 曲面（校准管线），不是逐笔 alpha 回归。  
- **目标是否偷看未来**：参数用历史路径估计；实验含合成与真实事件，强调 **短视界 belief 方差** 预测误差。  
- **和 veto 的关系**：给 **`p_market` 的「合理波动带」**——belief vol 异常时可 veto；**不**输出目标价。  
- **为何不是 junk**：和 BS 类比清楚、机制可解释；仍偏做市/对冲语言，落地要减维。

### 5. Semenas — 短天期 BTC 二元市场公平价与 stat arb 边（**7**）

- **URL**：https://papers.ssrn.com/sol3/papers.cfm?abstract_id=7134801  
- **回归在干什么**（据题名与 SSRN 元数据）：在 **单一时间窗** 里，用公平价框架检验 **候选统计套利边** 是否存在；核心是 **市场相对公平价的偏离**，不是长期价格回归。  
- **目标是否偷看未来**：若边用同期价差定义则有 **同期相关性** 风险；标题已写明 **single window**，当 **假说生成** 而非生产门控。  
- **和 veto 的关系**：支持「只有偏离超过成本才玩」的 **门控思维**；需外推成滚动校准才像 Polybot。  
- **为何不是 junk**：诚实标注单窗；和 5m/15m 二元场景近。OpenAlex 无摘要，细节需读 PDF。

### 6. Mitra & Tyagi — 库存风险下的 quasi fair value（**7**）

- **URL**：https://papers.ssrn.com/sol3/papers.cfm?abstract_id=7237578  
- **回归在干什么**（题名 + 同类 MM 文献）：在预测市场里定义 **含库存风险的准公平价值**，报价围绕该值偏移；通常伴随 **信念/库存状态** 的实证或结构估计。  
- **目标是否偷看未来**：公平价应只依赖 **当时** 状态与风险厌恶；若用全样本估库存惩罚需 walk-forward。  
- **和 veto 的关系**：公平价是 **参照系**，不是 alpha 预测；偏离过大可 **拒绝跟单**（尤其你们不做 MM）。  
- **为何不是 junk**：问题设定对口 PM；2026 SSRN，摘要缺失，读前降级为「待精读」。

### 7. Feil & Nendel — 预测市场最优做市与隐含概率（**6**）

- **URL**：https://arxiv.org/abs/2607.17991  
- **回归在干什么**：**随机控制**（非经典回归）： latent belief 扩散 → 条件概率价；解 HJB 得最优买卖价。  
- **目标是否偷看未来**：模型内前瞻；数值实验是 **模拟** 路径。  
- **和 veto 的关系**：告诉你 **库存 + 结算风险** 下报价应偏离「裸概率」多少——可启发 **何时 imbalance 不可信**，但不是现成的 logistic veto。  
- **为何不是 junk**：二元结算结构建模干净；和 Polybot **方向相关、工具不同**。

### 8. Curtin — Kalshi 上零成本否定散户叙事（**6**）

- **URL**：https://papers.ssrn.com/sol3/papers.cfm?abstract_id=7115858  
- **回归在干什么**（题名）：在 **受监管预测市场** 上用 **不占用本金** 的检验 **否定** 三条零售交易叙事；方法应是 **可观测隐含概率 vs 可实现 PnL 路径** 的 falsification，而非因子 zoo。  
- **目标是否偷看未来**：若用历史成交与结算，标签在事后；设计目标是 **证伪** 而非拟合未来价。  
- **和 veto 的关系**：和 Strategy Lab 的 **「先证伪再上线门」** 同精神——回归/检验用于 **关掉一类规则**，不是报价。  
- **为何不是 junk**：强调 falsify、零风险资本；摘要未入库，未确认是否含 logistic 校准。

---

## 淘汰（代表项）

| 文献 | 原因 |
|------|------|
| [MoFE / Fourier 专家 crypto 预测](https://arxiv.org/abs/2608.17342) | 「SOTA + 高 Sharpe」叙事；直接报价格，不是校准门。 |
| [PolySwarm LLM 预测市场交易](https://arxiv.org/abs/2604.03888) | Agent/延迟套利；无稳健概率校准门。 |
| [Unravelling the Probabilistic Forest 组合套利](https://arxiv.org/abs/2508.03474) | 组合无套利扫描；**不是**回归楔子/veto。 |
| [Executable Arbitrage / NegRisk](https://arxiv.org/abs/2608.00666) | 协议可执行性与套利利润；无校准回归。 |
| [Dynamic Grid Trading crypto](https://arxiv.org/abs/2506.11921) | 网格 backtest 跑赢叙事。 |
| 股指 LSTM+GRU+VaR 分位数门（academic-mcp 命中） | 股票因子/技术指标 zoo +「LSTM 经济价值」。 |
| 医疗/信贷 isotonic、Platt 校准综述 | 方法相关，但 **域外**；不进入 PM 门控清单。 |
| [Optimal MM 博士论文（Jain, Harvard DASH)](https://dash.harvard.edu/handle/1/42743947) | 与 Feil 文重叠、偏 RL/MM；未精读不占位。 |

---

## 对 Polybot 的一句收束

- **最贴架构**：Portnaya（双源概率楔子）+ Le（分格 logistic 斜率，crypto 短周期可近似恒等）+ Hao（分位数 / coverage **veto**）。  
- **公平价 / 波动带**：Dalen、Semenas、Mitra 线 → 约束「imbalance 多大才算数」，仍 **不** 替代 Jev 出价。  
- **落地注意**：凡用 **结算结果** 估校准的，训练窗必须 **严格早于** 交易窗；描述性大样本（Le）只能当 **先验**，不能直接当 live 门参。

*生成：2026-09-25，academic-mcp 检索。*
