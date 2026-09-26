# 风险感知型美股监控系统：产品研究说明

**用途**：把本文交给 Perplexity，调研市面上已经存在的产品、交互模式、数据源和功能边界，扩展产品想象力，并判断哪些能力值得在本项目中复刻。

**当前版本**：2026-07-14

**重要边界**：这是一个只提醒、解释和复盘的决策辅助系统，不自动下单，不把任何信号表述为确定的买卖建议。

## 1. 我们想做什么

我们要做的不是普通的价格提醒器，也不是预测涨跌的黑盒交易机器人。

目标是一个面向个人美股交易者的 **风险感知型市场监控与决策辅助系统**。它持续观察一个有限 watchlist，在出现真正值得注意的行情时，给出少量、可解释、与当前仓位和行为风险相关的提醒。

用户主要在中国/香港时区关注美股，因此系统尤其要解决：

- 夜盘、盘前、开盘、午盘和收盘前的市场状态切换；
- 睡觉或无法盯盘时，是否存在需要知道的风险；
- 高波动主题股快速上涨或下跌时，避免 FOMO、追涨和恐慌卖出；
- 把市场、板块、个股、期权隐含波动率、新闻和价格速度放在同一上下文中；
- 记录当时为什么提醒、之后发生了什么，逐步校准系统而不是只凭感觉。

系统第一阶段服务美股正股和 ETF。期权数据主要用于估计市场定价的风险幅度，不以自动交易期权为目标。

## 2. 目标用户体验

系统应该只在少数高价值时刻打断用户。理想状态下，每个纽约交易日不超过五条即时 Discord 消息。

一条提醒不应只说“价格涨了 5%”。它应回答：

1. **发生了什么**：例如，个股在 10 分钟内加速下跌，或盘前跳空超出预期范围。
2. **为什么异常**：相对于自身当天预期波动、同板块 ETF、同主题股票和大盘，哪里不一致。
3. **证据是否充分**：是否有公司公告、SEC 文件、宏观事件或可靠新闻；还是只有价格行为。
4. **当前应采取什么非冲动行为**：例如 `等待确认`、`禁止追高`、`持仓观察`、`减仓复核`、`止损复核` 或 `只记录`。
5. **数据是否可信**：行情时间戳、来源、延迟、期权合约更新时间和新闻证据等级。

这里的核心价值不是“每次判断都对”，而是减少低质量冲动、减少盯盘时间，并让用户能回看系统当时掌握了什么信息。

## 3. 目前已经有的系统和数据

当前后端部署在 `vps-hk`，使用 Go/Fiber 和 PostgreSQL。2026-07-14 已实际验证服务可运行、可刷新 watchlist 行情、可生成特征，并可发送 Discord 测试消息。

### 3.1 当前 watchlist 主题

| 主题 | 当前标的 |
|---|---|
| 半导体/存储 | `AMD`, `MU`, `NVDA`, `SMH` |
| 航天 | `ARKX`, `RKLB`, `SPCX` |
| 清洁能源/高 beta 能源 | `FCEL`, `TE` |
| 量子计算 | `IONQ`, `QBTS`, `QUBT`, `RGTI` |
| 市场背景 | `IWM`, `QQQ`, `SPY`, `TLT`, `^VIX` |

watchlist 未来可以拆成四种角色：真实持仓、战术观察、研究观察和市场基准。当前大部分标的仍被保守地标记为 `research_watch`，避免系统误把观察清单当成持仓。

### 3.2 已接入的数据

| 数据 | 当前来源 | 采样/粒度 | 当前用途 | 可信度边界 |
|---|---|---:|---|---|
| 美股/ETF 最新报价 | Yahoo Chart | 全表约 `60s`；重点标的 sampled quote 可约 `30s` | 当前价、开盘价、昨收、数据新鲜度、价格变化 | 适合 MVP；非正式交易所 SLA，不能当作 NBBO 或逐笔成交 |
| 美股 K 线 | Yahoo Chart | `1m` finalized bars | OHLCV、实现波动率、速度、技术结构 | Yahoo 不支持可靠的 `15s`/`30s` 历史 K 线 |
| 盘前/盘后 | Yahoo 可用时的 chart/quote | 依标的与会话而定 | gap、延长时段上下文 | 覆盖和延迟必须逐次标记 |
| RWA/tokenized equity 代理 | Bitget Wallet / Ondo | 基线约 `30s`；NVDA 快速 sampled quote 可约 `15s`；`1m` K 线 | 夜盘、盘前方向、价差和同步性研究 | 不是 Nasdaq/NYSE 官方成交、NBBO 或 Level 2；不能默认领先美股 |
| 期权链 | Yahoo Finance MCP | 液态子集约 `5m` | ATM 期权价格、bid/ask/mid、OI、volume、到期日 | 免费且非 OPRA 级实时数据；合约最后成交常明显滞后 |
| 反解 IV 与预期波动 | 后端 Black-Scholes 求解器 | 依期权采集更新 | 年化 IV、单日预期振幅、昨收/开盘/当前价范围 | 欧洲式 Black-Scholes 近似；适合风险范围，不代表方向或精确概率 |
| 实现波动率与价格速度 | 后端特征引擎 | `1m`, `3m`, `5m`, `15m` 窗口 | 识别加速、减速、区间使用率 | 当前不是逐笔或真正 `15s` bar 速度 |
| 新闻/事件证据 | VPS Firecrawl-compatible adapter | 按预算轮询 | 公司、板块、宏观事件的解释 | 新闻不是价格真相；必须记录来源等级、发布时间、接收时间和 ticker 匹配范围 |
| 持久化与回放 | PostgreSQL | append-only snapshots + finalized bars | 复盘、校准、结果归因 | 样本量尚小，不能声称统计有效性 |

### 3.3 已经能运行的功能

- 维护分组 watchlist，并持续保存报价、最终确认的 `1m` bar、期权摘要、RWA 样本和派生特征。
- 对每条行情保存 `provider_time`、`received_at`、`data_age_seconds`、session 和 provider warning；过期数据应被阻断或降级。
- 获取当前价、昨收、正规盘开盘价，并计算从不同锚点出发的合理波动区间。
- 用 near-ATM option bid/ask midpoint 反解 IV；同时保留原始 Yahoo IV、输入价格来源、合约时间戳和警告。
- 计算实现波动率、单日预期振幅、最近 `1m/3m/5m/15m` 价格速度。
- 比较 RWA proxy 与美股报价的同步性或价差，但不把它当作官方现货或领先指标。
- 获取新闻证据并按 `T1_OFFICIAL`、`T2_PRIMARY`、`T3_REPORTING`、`T4_UNVERIFIED` 分类。
- 支持传统规则：价格触及、日内幅度、相对弱势、成交量异常和数据过期。
- 支持 Discord webhook 测试、速率限制处理和健康告警。
- 支持 shadow-only 的行为策略记录：`PRICE_DISCOVERY`、`WAIT_CONFIRMATION`、`NO_CHASE`、`SUPPORTED_MOVE`、`UNSUPPORTED_IMPULSE`、`DISLOCATION_RISK`、`BLOCKED` 等状态，以及 15/30/60/180 分钟后的结果回放。

### 3.4 尚未启用或尚不可信的部分

- 行为策略目前是 **shadow-only**：会记录候选和预览，但不会自动把市场策略消息推送到 Discord。
- 真实持仓、成本、持有周期和风险预算尚未录入；因此系统不能安全地给出真正“持仓相关”的建议。
- peer/benchmark maps 已提出，但还没有经过用户批准和实证校准。
- Yahoo 期权链可以用于探索和初步 IV/range，但不能称为实时 0DTE 期权流或 OPRA 级数据。
- 没有真实美股 Level 2/order-book 数据，也不能把 Bitget RWA depth 伪装成 Nasdaq/NYSE order book。
- RWA 与美股的领先/滞后关系尚未经过足够交易日验证。
- 还没有足够完整交易日样本来证明某一异常规则具有稳定预测能力。
- 系统不会自动下单，也不应在没有长期检验前升级为自动交易。

## 4. 有了这些数据，可以实现什么功能

下面是从基础到高级的能力地图。Perplexity 应帮助寻找已成熟的产品体验、功能细节和数据要求，而不是只列出泛泛的“AI 选股”。

### 4.1 基础监控与通知

- 固定价格位、日内涨跌幅、gap、前高/前低、盘前/盘后异动提醒。
- 开盘、午盘、收盘前、盘后和夜盘的定时摘要。
- 数据源失效、延迟、限流、期权更新时间过旧等系统健康提醒。
- watchlist 分组和按主题的风险热力图。

这类能力容易实现，但不是最终差异化。

### 4.2 动态“正常范围”与异常雷达

对每个标的，动态估计：今天从昨收、开盘价和当前价出发，什么范围内的波动仍属常态，什么程度开始值得人工关注。

可做的功能：

- `expected move`、一倍/两倍预期振幅区间和当前 `range use`。
- 盘前 gap 是否已经消耗掉当天较大部分预期波动。
- 以当日到期期权、短期限期权和实现波动率交叉校验风险范围。
- 标记“涨跌很大但仍在该标的正常高波动范围内”和“幅度不大但对 SPY/QQQ 而言异常”的区别。
- 当前剩余交易时间内的风险范围，而不是只有全天范围。

这类能力最适合做成 `WATCH`、`ALERT` 和 `NO_CHASE` 的基础，而不是直接给方向。

### 4.3 速度、加速度与失败延续

用户特别关心短时间暴涨暴跌后，价格是否继续、横盘、回撤或重新加速。

可做的功能：

- 多窗口 price speed：`1m`、`3m`、`5m`、`15m` 的收益率和单位时间速度。
- speed acceleration/deceleration：快速移动是否正在减速。
- 一次急拉/急跌后是否能继续突破，还是只在局部极值附近横盘并失去动量。
- impulse rejection：快速扩张后没有延续，随后出现明显回撤的行为状态。
- 重新加速风险：午盘或尾盘的波动衰减后再次加速。
- 未来可升级到 VWAP、volume persistence、trade count、真实 order-flow 和 Level 2 imbalance。

这类功能应优先输出“等待确认”或“禁止追高”，而不是把一次速度异常解释成必然反转。

### 4.4 相对强弱、主题篮子和关系网络

单看某只股票经常会误判。系统可把个股放回市场、板块和主题中比较。

可做的功能：

- `MU`/`NVDA`/`AMD` 相对 `SMH` 和 `QQQ` 的超额强弱。
- `IONQ`/`RGTI`/`QBTS`/`QUBT` 的主题同步、分化和领涨/领跌。
- `RKLB` 相对 `SPCX`/`ARKX` 的特异性 move。
- “跟跌但不跟涨”的滞后者识别。
- 龙头先走、跟随者确认，或龙头反弹而弱者无法恢复的结构判断。
- 主题 risk-on/risk-off heatmap，以及当个股偏离篮子时的异常解释。

这类功能是把普通 TradingView alert 升级为风险解释和机会筛选的关键。

### 4.5 事件、新闻和证据链

可以把 Firecrawl、SEC、公司 IR、宏观日历和财经报道变成“证据层”，而不是让 LLM 自由编造原因。

可做的功能：

- 夜间发生的公司公告、SEC filing、财报、指引、评级变化和宏观事件摘要。
- 盘前 brief：gap 与已知事件是否相称，是否存在未解释的 gap。
- 新闻传播时间线：发布时间、系统发现时间、价格开始反应时间。
- 公司级证据、板块级证据、宏观级证据分开显示。
- 在证据不足时，明确输出“原因未确认”，而不是编造叙事。
- earnings、FOMC、CPI、jobs report、Fed speaker 等事件日风险日历。

### 4.6 期权风险与波动率工作台

即使不交易期权，期权市场也可用于风险定价参考。

可做的功能：

- ATM expected move、近端到期 IV、IV 与实现波动率的差异。
- 当天实际 move 已使用多少期权预期振幅。
- IV percentile/IV rank：需要积累历史 IV，且需要质量更高的 option source 才能做严肃结论。
- put/call skew、期限结构、volume/OI 异常：需要更稳定、低延迟的期权数据源。
- earnings 前后的 IV build-up、earnings 后 IV crush 和实际 gap 对照。
- “价格上涨但 IV 同时抬升”与“价格上涨而 IV 回落”的风险解释。

### 4.7 行为守栏与个性化风险控制

这是本系统最不同于普通行情软件的部分：系统要防止用户在最容易冲动的时刻增加风险。

可做的功能：

- `NO_CHASE`：快速上涨已经扩张、缺少板块/成交量/证据确认时，只提醒不要追高。
- `NO_CATCHING_KNIVES`：快速下跌、仍在加速、没有企稳证据时，不把“大跌”误认为买点。
- 开盘前 30 分钟默认只观察，只有非常明确的例外条件才允许提示高质量候选。
- 用户睡前风险 brief：剩余时段是稳定、未决、重新加速，还是数据不足。
- 针对真实持仓：把提示限制在用户预先允许的行动，例如仅 `持仓观察` 或 `减仓复核`。
- 每次提醒后收集 `helpful/not helpful`、`acted/did not act` 和情绪强度，衡量提醒是否减少 FOMO。
- 每日提醒预算、同一 symbol/state cooldown、digest 合并，控制噪音。

### 4.8 复盘、学习和“市场时间机器”

由于数据会落库，系统未来可以做比实时提醒更有价值的事：复盘。

可做的功能：

- 在任意历史时点回放当时价格、IV、新闻、速度、RWA 和系统状态，而不是事后使用未来信息。
- 对每种提示计算 15/30/60/180 分钟后的最大有利/不利波动。
- 判断 `NO_CHASE` 是否真的减少了追高，以及 `WAIT_CONFIRMATION` 是否过于保守。
- 比较不同 session、标的、市场 regime 下同一信号的表现。
- 生成每周“系统判断质量报告”：提醒数量、误报、漏报、节省的盯盘时间、用户情绪反馈。
- 逐步做个体化 calibration，而不是拿通用指标参数套所有高 beta 股票。

## 5. 市面上值得研究的产品方向

请不要只找“股票提醒 app”。应按以下产品类别分别调研，寻找其最佳交互、数据能力和付费门槛。

| 产品类别 | 需要研究的问题 |
|---|---|
| 图表与条件提醒 | 多条件 alert、跨标的/跨周期 alert、webhook、脚本化策略、手机/Discord 通知如何设计？ |
| 实时市场扫描器 | 如何发现 unusual volume、gap、relative strength、sector rotation、top movers？ |
| 期权分析与交易风险工具 | expected move、IV rank、skew、0DTE 风险、options flow、earnings volatility 如何展示？ |
| Order-flow / market microstructure 工具 | 哪些产品提供 Level 2、DOM、heatmap、liquidity、imbalance、tape；其数据成本和适用市场是什么？ |
| 新闻与事件终端 | 如何将新闻、SEC filing、财报电话会、宏观日历与价格时间线关联？ |
| 投资组合与风险管理工具 | 如何做仓位感知提醒、集中度、行业暴露、beta、止损复核、情景分析？ |
| AI research / financial copilot | 哪些产品提供有来源的解释、主题关系图、可追溯研究，而不是无证据的“买卖建议”？ |
| 交易日记与行为管理 | 如何记录交易理由、情绪、FOMO、冲动、复盘和规则违例？ |
| Overnight / global / tokenized-equity context | 哪些产品研究盘前、夜盘、ADR、期货、RWA/tokenized stocks 与美股开盘的关系？ |
| Alert routing / workflow | Discord、Telegram、Email、移动推送、quiet hours、digest、升级路径如何减少噪音？ |

可重点搜索但必须核实当前状态的产品/类别包括：TradingView、TrendSpider、Koyfin、Benzinga Pro、Bookmap、Quantower、Sierra Chart、Market Chameleon、OptionStrat、SpotGamma、Unusual Whales、Quartr、Fiscal.ai、FinChat、Interactive Brokers、Alpaca、Polygon/Massive、ThetaData、Tradier、ORATS、Cboe LiveVol。

产品名称只是调研起点，不代表推荐、可用性、实时性或价格已验证。

## 6. 希望 Perplexity 帮我们回答的问题

### 6.1 产品能力

1. 目前有哪些产品能把实时价格、技术条件、相对强弱、新闻、期权 IV 和仓位风险整合在同一个工作流中？
2. 哪些产品最擅长“少而重要”的提醒，而不是高频 spam？
3. 哪些产品在防止 FOMO、交易日志、规则执行和心理纪律方面有值得复刻的设计？
4. 有哪些成熟的开盘、盘前、午盘、收盘前和隔夜 briefing 产品形态？
5. 有哪些工具能把“异常发生”与“为什么异常、证据是否足够、后续如何复盘”连起来？

### 6.2 数据与工程可行性

1. 对美股个人用户，什么 provider 能提供实时 quote、`1m`/sub-minute bars、WebSocket、Level 1、Level 2、trades 和 order book？免费层、延迟层和付费层分别怎样？
2. 哪些 options data provider 能可靠支持 near-real-time 0DTE bid/ask、IV、skew、OI 和 volume？它们是否有个人可负担的计划？
3. 个人开发者使用 Discord webhook、Telegram、Email、手机推送时，怎样设计 rate limit、cooldown、quiet hours 和 escalation？
4. 什么数据可以合法、稳定地用于个人项目，哪些数据不能从免费 Yahoo/网页抓取层面声称是实时交易所数据？
5. 对 RWA/tokenized stocks、盘前价格、ADR、指数期货和美股现货，什么情况下能作为相关上下文，什么情况下不该做 lead-lag 主张？

### 6.3 功能优先级

1. 哪些能力只依赖现有 Yahoo + Firecrawl + Bitget RWA + PostgreSQL，即可作为 v1 实现？
2. 哪些能力必须增加付费/正式数据源才可信？
3. 哪些功能看起来酷但容易导致假精确、过度交易或错误因果解释？
4. 哪些产品特征最适合先复刻，以提高用户注意力效率而非增加提醒数量？

## 7. 给 Perplexity 的可直接粘贴任务

```text
我正在开发一个个人使用、只提醒不自动下单的“风险感知型美股市场监控与决策辅助系统”。

现有技术与数据：
- Go/Fiber + PostgreSQL 后端部署在 VPS；Discord webhook 通知。
- Yahoo：近实时 quote、1m bars、盘前/盘后可用时的数据、免费期权链。
- 后端从期权 bid/ask midpoint 反解 Black-Scholes IV，计算 ATM expected move、从昨收/开盘/当前价出发的合理波动范围、实现波动率、1m/3m/5m/15m price speed。
- Bitget/Ondo tokenized-equity/RWA：作为盘前、夜盘和离线时段的辅助 proxy，不当作官方美股价格或 Level 2。
- Firecrawl-compatible 新闻层：公司、板块、宏观事件证据，带来源等级和时间戳。
- 当前 watchlist 聚焦 semis/memory、space、quantum、energy high beta，以及 SPY/QQQ/IWM/TLT/VIX 市场背景。

目标用户：位于中国/香港时区，主要做美股日内投机和中期投资；非常重视降低 FOMO、追涨杀跌和盯盘压力。每天最多接受约 5 条即时提醒。系统必须给出解释和“等待确认/不追高/持仓观察/减仓复核”等候选动作，而不是自动交易或确定性买卖建议。

请用截至今天的官方资料、产品文档、定价页、可信产品评测和实际产品截图/演示，完成以下调研：

1. 按产品类别梳理领先产品：图表与 alert、实时 scanner、options analytics、order-flow/Level 2、新闻事件终端、portfolio risk、AI financial copilot、trade journal/behavioral guardrails、overnight/global context 和 notification workflow。
2. 对每个最相关产品说明：核心功能、实时性与数据来源、是否支持价格/相对强弱/IV/新闻/仓位感知/Discord 或 webhook、免费或个人计划价格、主要局限、最值得复刻的交互设计。
3. 区分三层实现难度：
   - 可以用现有 Yahoo + Firecrawl + Bitget RWA + PostgreSQL 做到；
   - 需要较低成本的正式数据源；
   - 必须有付费交易所/OPRA/Level 2 数据才应该做。
4. 特别寻找“少而重要的异常提醒”“防 FOMO”“开盘/午盘/隔夜 brief”“expected move vs realized move”“relative strength/theme basket”“news evidence timeline”“alert review and journaling”的优秀产品案例。
5. 提出 15 个我们当前没有想到、但对该系统有价值的功能。每个功能说明：用户价值、所需数据、可否在当前技术栈实现、风险或容易误导的地方、建议优先级。
6. 不要把 Yahoo 免费数据或 tokenized/RWA proxy 描述为官方实时美股、OPRA options feed 或 Nasdaq/NYSE Level 2；不要推荐自动下单作为第一阶段。

请给出带链接的 Markdown 报告，优先引用官方文档和定价页。最后给一个“最值得先复刻的 10 个功能”清单，并按价值、实现难度、数据成本、误导风险、减少 FOMO 的效果排序。
```

## 8. 希望收到的调研输出格式

让 Perplexity 对每个候选功能使用同一结构，以便后续能直接转成 engineering backlog：

| 字段 | 要求 |
|---|---|
| 功能名称 | 用户能看懂的短名称 |
| 参考产品 | 具体产品和官方链接 |
| 用户问题 | 它解决哪一种风险、注意力或决策问题 |
| 输入数据 | quote、bar、options、news、position、Level 2 等 |
| 延迟要求 | `1m` 足够、10-30 秒可用、还是必须 WebSocket/Level 2 |
| 现有数据能否支持 | 可直接做、可降级做、还是必须新增 provider |
| 最小可行交互 | Discord、dashboard、daily brief、replay 等 |
| 主要误导风险 | 数据延迟、幸存者偏差、假因果、过度提醒等 |
| 复刻优先级 | `P0`、`P1` 或 `P2`，并说明理由 |

报告应明确区分：**已验证事实、产品宣称、研究推断和需要我们自己回测验证的假设**。

## 9. 产品原则

- 少而重要优先于全量提醒。
- 数据新鲜度和来源透明优先于漂亮但不可靠的结论。
- 异常不等于反转，不应把“超范围”自动解释为买点或卖点。
- 价格、板块、期权、新闻和仓位必须分层表达，避免单一故事解释一切。
- 先 shadow test 和复盘，再允许真正的 Discord 市场提醒。
- 用户行为负担也是一个指标：提醒即使方向正确，但诱发 FOMO 或打断睡眠，也可能是失败。
- 任何“领先指标”“RWA 价格发现”“订单簿 imbalance 有预测力”的主张，都需要在本 watchlist 的存储样本上独立验证。

## 10. 本轮调研的成功标准

调研完成后，我们应该能回答：

1. 这个系统与 TradingView 式 price alert 相比，真正的产品差异在哪里？
2. 哪些成熟产品能力值得直接借鉴或复刻？
3. 哪些能力已经能在现有免费/低成本数据条件下落地？
4. 哪些看似重要但必须购买高质量行情、期权或 order-book 数据后再做？
5. 怎样把系统做成一个减少 FOMO 和盯盘负担的工具，而不是另一个制造焦虑的行情屏幕？
