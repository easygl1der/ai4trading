# 稳定均值回归 / 库存均值回归：订单簿与预测市场 CLOB 文献摘选

**检索**：2025-09 至 2026-09 为主（硬截止 2024-09 起），经 `academic-mcp`（arXiv / OpenAlex）检索关键词：prediction market CLOB、mean reversion、Avellaneda–Stoikov、Ornstein–Uhlenbeck、adverse selection、Polymarket/Kalshi microstructure。  
**剔除**：Telegram/Twitter 盈利截图、无方法论的「策略帖」、与 CLOB/订单簿无关的泛量化博客。  
**读者**：Polybot paper gate——只在 `|p_market − p_theory|` 足够大、且 spread/depth/fees 允许时做 fade，而不是做市库存引擎。

---

## 1. Novotny (2026) — 基本面锚定带来的价格均值回归

**URL**：https://arxiv.org/abs/2607.16970  

**机制（一段话）**：在 agent-based 订单簿里，若流动性供给锚定在一个可识别的「基本面价值」上，冲击后 mid 会向该价值回归、盘口会重新填满；把锚定强度拧弱，均值回归消失，杠杆抛售可以自我维持。作者用因果干预说明：流动性危机更像是「锚失效」而不是「做市商撤单本身传染」——这对「稳定 reversion」的定义很直接：要有可观测、短期内不太漂移的 `p_theory` 锚。

**迁到 paper gate**：把「锚强度」读成你对 theory 的信任区间（RTDS/外部参照 + 事件内慢变假设）；`|p_market − p_theory|` 触发 fade 前，先问锚是否仍在（新闻冲击、resolution 临近、参照断流 → 视为锚弱，gate 关）。  

**评分**：8/10  

**为何不是 junk**：明确因果实验 + 可复现仿真；不承诺实盘 alpha，讲的是结构条件。

**不要照搬**：模拟里的做市商行为、跨市场传染通道参数、任何「永远挂单」式库存策略。

---

## 2. Dubach (2026) — Polymarket 订单簿微观结构实测

**URL**：https://arxiv.org/abs/2604.24366  

**机制（一段话）**：30B 级 tick 订单簿事件与链上成交对齐，在预注册 600 市场面板上总结八条 stylized facts：longshot spread premium、深度剖面偏均匀、类别间有效价差差异、链上/feed 推断成交方向一致性仅 ~59% 等。核心测量结论：用公开 order-book feed 做 Lee-Ready 式方向，会系统性搞错 half-spread、Kyle λ 的符号——微观结构研究必须用链上 `OrderFilled` 定 aggressor。

**迁到 paper gate**：fade 的「可赚价差」要用**有效半价差 + 深度档位**算，不能假设 1 tick spread；按价格档位（longshot）和品类调节 `min_edge`；若你的 theory 信号依赖 OFI/方向分类，在 PM 上先验证是否与链上一致，否则 gate 会假阳性。

**评分**：9/10  

**为何不是 junk**：预注册面板、公开复制包、把「数据坑」讲清楚而非吹策略。

**不要照搬**：论文只做描述性事实，没有给出可抄的 entry 规则；不要把 feed 方向当 ground truth。

---

## 3. Zang, Andrade & Nakajima (2026) — 预测市场 CLOB：筛查租金 vs 经典噪声做市

**URL**：https://arxiv.org/abs/2609.20017  

**机制（一段话）**：交易级证据显示，大型预测市场 CLOB 上 maker 利润往往来自「持有被低估一侧直至结算」的筛查租金，而不是经典随机噪声流上的 spread capture；在共同信号带内报价会被 informed 拣选。模型里 CLOB 与 LMSR 共存：informed 流倾向 AMM，CLOB 靠 tail demand 租金支撑——与「delta-neutral AS 做市 + OU 公平价」叙事直接张力。

**迁到 paper gate**：你的 fade 是**短暂偏离**还是**方向性押注到 resolution**？若是后者，文献说这是 PM 上更常见的赚钱机制，不是「稳定 reversion」。gate 应排除：单边 informed 冲击、tail 需求主导的长价区、以及「价差够宽就均值回归」的朴素假设。

**评分**：9/10  

**为何不是 junk**：平台级交易数据 + 结构模型，对策略设计是约束而非喊单。

**不要照搬**：inventory 持有到结算的 rent 逻辑、LMSR/CLOB 路由故事、任何「双边挂单吃 spread」的 live quoting 参数表。

---

## 4. Qin & Yang (2026) — Polymarket-v1 全链成交库与方向自相关

**URL**：https://arxiv.org/abs/2606.04217  

**机制（一段话）**：12 亿笔链上成交、100% 真 aggressor 方向。Tick rule / BVC 在 PM 上接近随机，但背后是**成交方向正自相关**与集中做市——违反经典分类器嵌入的「短期均值回归」假设；错误方向会污染 VPIN、OFI 和 TCA。真 VPIN、Gibbs spread 与 Brier 的关系在 proxy 指标下会被稀释。

**迁到 paper gate**：若 `p_theory` 来自 order-flow 或「上一笔反转」类特征，在 PM 上先验就不稳；fade 触发应更依赖**慢变量 theory**（校准层、外部价格、规则化 fair value），并用 spread/深度过滤「方向持续」段。可把「方向自相关显著 + 浅深度」设为 **反 gate**（禁止 fade）。

**评分**：8/10  

**为何不是 junk**：全样本链上真值 + 对标准微观工具的系统 benchmark。

**不要照搬**：数据库里的 VPIN 交易策略、把分类器当 PM 标准管线。

---

## 5. Young (2026) OpenMarket — Polymarket×Binance 高频对齐与零结果

**URL**：https://arxiv.org/abs/2607.26245  

**机制（一段话）**：为交易 Polymarket BTC 15 分钟二元市场相对 Binance 订单流而建的毫秒级同步语料；43 个微观特征 walk-forward 逻辑回归**跑不赢**订单簿隐含概率，含费滑点模拟平均每笔约 −0.116（归一化 payoff）。数据侧贡献：1-tick spread stylized fact、跨所滞后（中位 ~347ms 响应大 Binance 波动）等。

**迁到 paper gate**：直接支持「短周期 `p_market` 相对外部参照的 reversion」在费后很难——你的 gate 必须显式扣费、扣最小 tick、扣跨所滞后；若 edge 只来自几秒内的 BTC 映射，文献暗示 OOS 可能为 null。适合当 **falsification 锚**，不是 promotable 模板。

**评分**：8/10  

**为何不是 junk**：诚实报告失败、冻结语料与可复现管线；不是事后挑窗口吹 win rate。

**不要照搬**：43 特征模型、任何「beat the book」的 ML 入场；这是数据论文，不是 alpha 配方。

---

## 6. Le (2026) — Kalshi/Polymarket 分域校准与条件「理论概率」

**URL**：https://arxiv.org/abs/2602.19520  

**机制（一段话）**：3.53 亿笔交易、42.9 万二元合约；校准随**事件域、距结算时间、成交规模**系统变化（政治市场长期 underconfidence、价格向 50% 压缩）。一阶 logistic 重校准斜率可解释大部分 in-sample 方差，且部分模式在 Polymarket 复制；Bayesian 测量误差模型提示约一半横截面斜率差可能来自估计噪声。

**迁到 paper gate**：`p_theory` 不应是全局 `p_market` 的平移，而应是 **条件校准后的 fair value**（域 × 剩余时间 × 规模桶）；`|p_market − p_theory|` 要在「该桶历史校准残差」尺度上度量，否则 longshot 区的「偏离」可能是正常压缩而非 reversion 机会。spread/fees gate 与校准桶联动。

**评分**：7/10  

**为何不是 junk**：超大样本、跨平台对照、对「概率字面意思」的严谨拆解。

**不要照搬**：政治盘 underconfidence 的套利仓位 sizing；论文不给高频 fade 时点。

---

## 7. Barzykin, Bergault & Guéant (2024) — 嵌套 OU 公平价与多时间尺度 spread 松弛

**URL**：https://arxiv.org/abs/2404.15478  

**机制（一段话）**：现货–期货 EFP 价差用**嵌套 Ornstein–Uhlenbeck**（Hull–White 利率思路）建模，多种松弛模态对应不同交易 horizon；在 HJB 框架下同时管理现货/期货库存与价差均值回归，并给出可近实时求解的数值方案。这是标准的「OU fair value + inventory-aware control」教科书级结构，虽非预测市场，但数学对象与 AS 族一致。

**迁到 paper gate**：把 `p_theory` 看成 OU 均值、把 `p_market − p_theory` 看成平稳（或分段平稳）的 spread 状态；**仅在估计的 half-life 内、且 reversion 速度 × 预期幅度 > fees + 有效半价差** 时允许 fade。多时间尺度 → gate 可分「快 reversion（流动性补偿）」与「慢漂移（信息）」两档，慢档默认关。

**评分**：7/10  

**为何不是 junk**：Guéant 一脉的严格随机控制 + 贵金属实盘结构；可检验的 OU 参数而非叙事。

**不要照搬**：EFP 双资产库存最优报价、实时做市 HJB 策略；paper 只做单边 fade，不做双边 quote 优化。

---

## 8. Feys (2026) — Avellaneda–Stoikov 与 Cartea–Jaimungal 的单一风险参数

**URL**：https://arxiv.org/abs/2606.01477  

**机制（一段话）**：在 entropic 确定性等价 + 动态一致性公理下，AS 是唯一规范形式；CJ 的 running penalty `φ` 与 terminal `α` 由单一 `γ` 强制：`φ = γσ²/2`（并可与清算成本二阶项联动）。独立校准 `γ` 与 `φ` 在理论上是不自洽的——两框架是同一对象的不同展开，而非可随意混搭的「库存惩罚旋钮」。

**迁到 paper gate**：若你从 AS 文献借「库存均值回归强度」来缩放 fade 仓位，文献警告：**不要把 inventory skew 当成与 vol 无关的第四参数**；paper gate 里 inventory 应硬 cap 或零（你不做市），`γ` 只应进入「风险调整后的最小 edge」，而不是 live skew 报价。与「只 fade、不挂单」一致。

**评分**：6/10（对 PM 间接，但对 AS 误用极有用）  

**为何不是 junk**：公理化唯一性定理，用来防 desk 式参数拼凑。

**不要照搬**：_entire_ AS/CJ 做市控制律、reservation price 双边报价、任何随库存连续挪价逻辑。

---

## 汇总结论（给 coordinator）

| 主题 | 文献共识 | 对 paper gate 的含义 |
|------|----------|----------------------|
| 稳定 reversion | 需要锚定 fair value（Novotny）；PM 上方向常持续而非反转（Qin–Yang） | theory 要慢、要可辩护；flow 自相关高 → 关 gate |
| 预测市场 CLOB | 利润结构偏筛查/持有，非经典噪声 MM（Zang） | fade 不是「挂双边吃 spread」 |
| spread/depth/fees | longshot premium、1-tick、有效半价差（Dubach；Young 费后 null） | `min_edge` 必须品类与价位分桶 |
| `p_theory` | 条件校准（Le）；OU 均值可选（Barzykin） | 偏离度要桶内标准化 |
| AS/库存 | `φ` 与 `γ` 不独立（Feys） | **禁止** inventory skew 做市；仅借风险定价不等式 |

### 明确不要从这类文献抄进 Polybot paper 的东西

1. **Inventory sizing / 库存均值回归做市**：AS/CJ/Guéant 最优报价、reservation price 随库存平移、bid–ask 双侧 live quoting。  
2. **持有到结算的筛查租金**（Zang）当作「高频 fade」。  
3. **用 order-book feed 方向**做 OFI/VPIN 门控而不链上校验（Dubach、Qin–Yang）。  
4. **短周期 cross-venue ML edge**（Young 已 OOS 失败）当默认 alpha。  
5. **社交证明式回测截图**（本次检索已排除，也不应作为证据来源）。

### 缺口（未凑满 8 篇里的「理想型」）

近 24 个月内**专门**写「预测市场 CLOB + OU mid + 费后 fade gate」的合一论文很少；OU/AS 多仍在传统 LOB（Barzykin、Meteykin 2025 HJBQVI+alpha 等）。若以后要补，可在 academic-mcp 跟踪 Meteykin (https://arxiv.org/abs/2512.20850) 的 alpha-aware MM 数值解，但迁到 paper 时仍只取「有信号才缩 spread / 无信号不交易」的**门控思想**，不取做市控制律。

---

*文档路径：`/cursor/stores/self/docs/research/stable-reversion.md`（与 Project store 同步）。未修改 `papers.md` / `README.md`。*
