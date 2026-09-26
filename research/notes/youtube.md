---
cursor:
  subagentId: "bc-60a9b796-b587-521a-b3ab-c91ec6e2b0d6"
---

# YouTube 扫描：预测市场 / 加密二元 / LLM 决策

两轮检索，工具都是 youtube-data `videos_searchVideos`，像样的再拉 `transcripts_getTranscript`。封面或标题承诺保本/暴利的直接剔除。口播数字未回论文、链上或原始盘口。

- 第一轮：`publishedAfter = 2025-09-25`，保留 5。其中 LiveTradeBench 在 2025-11，落在近 5 个月之外。
- 第二轮：`publishedAfter = 2026-04-25`（相对 2026-09-25 的近 5 个月）。只要有逐字稿、能对上 trade / skip / pause、而且正文不是收益截图。新增保留 4。同一系列的中间集并进主条，不单列。

合计保留 **9**。AssemblyAI 没有调用：这 4 条都有 YouTube 字幕。唯一关字幕、标题像纸面实验的是一场约 11 小时 55 分的直播；watch 页不是音频文件，没有整场送去转写。描述里的打分规则只记在剔除节，不当成听过的逐字稿。

门控对照：把波动、动量、盘口深度融成 **trade / skip / pause**，不算仓位。下面「相关」只谈这一层。

---

## 保留

### 1. Order-book imbalance as a 30-second alpha on Polymarket binary contracts

- 频道：FNRTrades
- 日期：2026-06-22
- URL：https://www.youtube.com/watch?v=odJ2FUudSpE
- 判断：操作向笔记，连贯。封面是图不是利润承诺。片尾有推荐链接。Sharpe / Kelly 是 sizing，不采用。同一频道后来的校准曲线短片和 PIN 短片是同一套口播模板（见「谨慎」和「剔除」），只采信这里的盘口条件，不采信任何月收益。

实质：

1. 四个最高成交二元合约、30 天、2 秒一档顶层盘口，约 1200 万观测。
2. 失衡定义为内侧 \((\text{bid size}-\text{ask size})/(\text{bid}+\text{ask})\)。
3. \(|I|>0.6\) 时，未来 30 秒中间价同向漂移约 1.1–1.3 美分；1 分钟聚合后 \(R^2\) 从 tick 上的 0.08 升到 0.31。
4. 机制被说成队列与逆向选择，不是知情流：薄的一侧被吃掉后做市商拉开，失衡约 90 秒内部分回复。
5. 扣约 2% 费用和约 40 bp 有效价差后，边缘只留在报价 15–85 美分且 \(|I|>0.75\)，约占报价更新的 4%。

相关：这是盘口深度腿的 **skip 默认**。中等失衡、价格在两尾、或价差已经吃掉漂移，都不应 trade。只有极端失衡且价格在中间带，才有资格进入 trade 候选；视频自己的净成本结论也是绝大多数时刻不交易。它不提供波动或动量，不能单独当融合门。Kelly 与回测 Sharpe 不进入门控。

### 2. 2604.24366 — The Anatomy of a Decentralized Prediction Market（Polymarket 订单簿微观结构）

- 频道：AI Paper Cast
- 日期：2026-08-10
- URL：https://www.youtube.com/watch?v=9lrBw3gQKs0
- 判断：学术口播，连贯，不是操作手册。封面是论文卡。数字是转述，未对照 arXiv。

实质：

1. 声称用约 52 天、约 300 亿条链下 websocket 事件，再与链上成交对齐。
2. `change_side` 被论证为成交后的深度更新，不是主动方；拿它当 aggressor，方向大约错 40%（口播称链下符号准确率约 59%）。
3. 用链上方向重做 Glosten–Harris 后，前 100 名合约的有效半价差、暂时成分和逆向选择 \(\phi\) 的中位数都塌到接近 0。
4. 先前论文里很大的 \(\phi\) 被说成符号弄反造成的幻影毒性流，不是做市商被知情单系统性打穿。
5. 时间衰减被说成假象；可用的状态是成交量、持续时间和价格，盘口至少看到第 10 档。链上 60/40 的主动方向里还可能混有 MEV。

相关：这是盘口腿的 **data_trust / pause**。深度变化不能当成主动流；用错误符号算出的 OBI 或「毒性」去 trade，等于在噪声上开门。\(\phi\approx 0\) 也不是放行：它只说明不要因为幻影逆向选择而暂停，真正的 skip 仍要看价差、档位深度和报价是否可执行。波动和动量不在这篇口播里。

### 3. I Gave an LLM $406 to Beat a Prediction Market

- 频道：The Unattended Engineer
- 日期：2026-08-12
- URL：https://www.youtube.com/watch?v=xZpWnATVH1o
- 判断：操作向失败实验，连贯，反营销（明确亏了）。封面写 doomed，不是保本。因果「为什么没有边」被推到下一集，本集是架构和负结果。不是 5 分钟加密二元。

实质：

1. 约两个月、30 个版本、实盘 Polymarket、本金 $406；大约每 110 秒扫一遍，推理模型对新闻给出自己的概率。
2. 语料约 46.2 万个已结算市场，Yes 兑现率约 37%，只是基率；成交量列全空，没有流动性优势。
3. 概率分歧不等于单：接近抛硬币、仓位已经太大、已知陷阱市场、边薄于噪声，资金留下。
4. 四条策略里单边、双边、套利上过场；深度读取接好了但从未打开。
5. 实盘亏了。作者把失败定义成可复核的「没有边」，而不是回测填价。

相关：LLM 与中间价的差 **不能** 充当 trade。可用的是 skip：边不超过噪声、市场结构像陷阱、或没有盘口深度特征时直接跳过。深度策略没开、成交量缺失，说明这个代理没有把 book depth 融进决策。仓位用了 Kelly，那一层丢掉。波动和短周期动量都没有。适合当作「只有模型概率就 pause/skip」的反例，不是加密二元门的正例。

### 4. LiveTradeBench: Live Trading Benchmark for LLMs

- 频道：AI Research Roundup
- 日期：2025-11-06
- URL：https://www.youtube.com/watch?v=CYdgheUXmO4
- 判断：学术综述口播，连贯。封面是论文要点，无利润承诺。动作空间是资金权重，属于 sizing，不拿来当门控。Polymarket 是事件二元，不是 BTC 5 分钟涨跌。

实质：

1. 口播对应 2025-11-05 的 LiveTradeBench：LLM 在股票和 Polymarket 上做序贯实盘决策。
2. LM Arena 分数与累计收益的相关，股票约 0.05，Polymarket 约 \(-0.38\)。
3. 观测是持仓、实时价格和短新闻；动作是百分比配置，再映射成买、卖、持有。
4. 股票上 GPT-4.1 累计收益约 6% 且波动大；这些模型换到 Polymarket 后回撤变深、胜率下降。
5. 结论是静态榜不预测交易能力，跨市场不迁移，要的是对实时信号的序贯、风险意识决策。

相关：模型榜或「更强的 LLM」不是 trade 许可，相关为负时更接近 skip。论文把适应对象叫 live signals，但口播里的信号是价格和新闻，不是 \(\sigma\)、动量或深度的融合规则。百分比权重是 sizing，门控不应模仿这个动作空间。可保留的是否定句：没有实时微观结构适配，就 pause，不要因为模型排名放行。

### 5. Unravelling the Probabilistic Forest: Arbitrage in Prediction Markets

- 频道：Nosa Capital
- 日期：2026-05-17
- URL：https://www.youtube.com/watch?v=aonmXZ50tbM
- 判断：学术口播，机制连贯，修辞偏「漏洞/劫案」。封面是论文题。不是波动或动量策略。

实质：

1. 二元长套利：Yes 与 No 买价之和低于 1，两边都买锁定差价；短套利是概率和超过 1，转去买 No。
2. 跨市场：子集事件的价格不能高于超集（赢两球的价格不能高于赢球）。违反时组合 Yes/No 锁定价差。
3. 依赖关系靠 LLM 读市场文案来找；「flipping a state」这类词没有历史上下文时模型会错。
4. 口播把 2024 年美国大选 Polymarket 成交量说成超过 37 亿美元，用来说明错误定价的规模，不是策略成绩。
5. 结尾问题是：纠错如果主要由机器人完成，报价还算不算人群信念。

相关：这是盘口一致性 **硬 skip**，发生在波动/动量融合之前。Yes+No 明显偏离 1，或嵌套合约价格颠倒，书上的中间价就不是概率，动量和 \(\sigma\) 都不该打开 trade。LLM 在这里是依赖解析器，不是决策器；解析失败（上下文不够）本身就是 pause，而不是一笔交易。没有深度阈值，也没有仓位规则值得搬。

---

## 近 5 个月补扫

下面 4 条都在 2026-04-25 之后，逐字稿读过。数字是口播，不是我们复算的。

### 6. 48 Zeros: My AI Trading Edge Was Dead

- 频道：The Unattended Engineer
- 日期：2026-08-15
- URL：https://www.youtube.com/watch?v=zTjtF6vOaMQ
- 判断：第 3 条的续集。预先登记的负结果，反营销。不是 5 分钟加密二元。Kelly 不是要点。

实质：

1. 预先哈希的检验：540 条实盘预测对已实现结果。口播给出 \(p = 2.3\times 10^{-49}\)，高置信、大声称优势、分品类切片都没有正的一块。
2. 跨所价差在存在的地方大约 1.5 个点。2026 年 Polymarket 的动态 taker 费吃掉大约 94% 的这段差；费是百分比，把本金做大并不能把费吃掉的比例缩小。
3. Maker 可以避开 taker 费，但单腿成交（legging）在 $406 这种不能扛挂单的本金上是致命的。
4. 方向性被判死，盲 taker 接近抛硬币，maker 套利也被单腿打死。作者把交易关掉，用的词是 buried，不是暂停后再开。
5. 本集没有波动、动量或深度阈值。关掉的理由是费用相对价差，以及本金扛不住未配对的一腿。

相关：模型概率仍然不是 trade。费用吃掉价差是硬 skip。本金不能仓储挂着的一腿，maker 套利整条 pause。这和「把 \(\sigma\)、动量、深度融成开门」不是同一件事：它给出的是开门之前的费用与库存约束。

### 7. The Honest Verdict: Two Markets Can't Disagree

- 频道：The Unattended Engineer
- 日期：2026-08-21
- URL：https://www.youtube.com/watch?v=-61xFNh9QnM
- 判断：同一系列的收束。没有机器人。中间集 [The Phantom 14° Gap](https://www.youtube.com/watch?v=RIFBwp9B2Wc)（2026-08-18）的数字在本集重述，不单列。

实质：

1. LLM 可以提议一对「同一个市场」；只有机械核对（同一指标、同一主体、同一参照期）才算确认。政治：82 次匹配尝试，美国受监管的 Polymarket 对 Kalshi 确认数为 0。名词相同，可结算事件不同。此前 25 对「双胞胎」是离岸 Polymarket，不是这个受监管对照。
2. 体育：15 对进去，0 对出来。真正能对齐的是标量梯子（天气、增长率）。拉了 4,251 个 Kalshi 标量市场。隐含中位数差到大约 \(0.01\)–\(0.11^\circ\mathrm{F}\)，增长率差到大约 0.01 个百分点这一档（口播：hundredths of a percent）。这些差都落在 Kalshi 买卖价差大约 1.5–6 个点里面，费还没扣，差已经没了。
3. 五类标签错误会造出假价差：\(\ge\) 对 \(>\) 大约 12 个假点；区间对尾部；序列污染（日高、日低、午后拼在一起，大约 \(14^\circ\)、假的约 98 点缺口）；主体坍缩（美国违约把国家和同比/环比并在一起）；正则丢掉负号，梯子加总到大约 126。
4. 幻影那一集里，一侧报价陈旧大约 30 小时。身份没对齐之前，价格差不是分歧。
5. 放出约 46.2 万个已结算市场。没有下单。

相关：跨所「分歧」在身份成立之前是 skip，不是 trade。加总超过 1 的梯子是概率写错，不是动量。陈旧大约 30 小时的一侧是 data_trust pause。波动和深度融合排在身份核对之后；本集没有通过身份的加密二元样本。

### 8. Arbitrage and Insider Trading in Polymarket（Data Con LA，Yurii Nosov 代 Guang Cheng）

- 频道：Data Con LA
- 日期：2026-09-06
- URL：https://www.youtube.com/watch?v=sJF6CMoETHQ
- 判断：会议报告，研究助理代讲，连贯。不是 BTC 5 分钟策略，也不是产品演示。内幕「确认」被讲者自己标成括号，来源含 NPR，软。

实质：

1. 口播引用：头部 1% 拿走大约 70% 的赢利；这 1% 里大约 70% 是机器人，赚的是错误定价的小价差，不是预测。因此不要跟每一个赢的钱包。例子：有钱包每注约 $10,000，胜率 23%。
2. Isolation forest 用在下注规模和时点上，列出 34 个加引号的「确认」内幕。讲者明确说无法 100% 确认。
3. 单市场套利：Yes 20 美分加 No 75 美分，快的话锁 5 美分；和超过 1 就 mint 1 美元拆成 Yes 和 No 再卖掉。他们的窗口 2026-02-04 到 2026-03-04 只找到大约 7 次，\(7\times 5\) 美分 \(= 35\) 美分。讲者直接说靠这个活不了。
4. 子集价格不能高于超集。这种嵌套错价比单市场上的 Yes+No 之和更难被场地自己清掉。
5. CLOB 的挂单和撤单不在公开链上历史里。$50,000 的 80 美分买单然后撤掉，不自己连续下载就看不见。问答里托管机房是玩笑，不是配方；外部推文没有自动化接进信号。

相关：Yes+No 偏离 1 是稀有的一致性旗，频率在他们这一个月里低到不能当活。它不是波动或动量的融合。盘口状态必须自己录，否则撤单不可见，深度腿没有数据就不该 trade。跟赢家钱包是 sizing 和过程的复制，门控不做。

### 9. Building a Polymarket Bot: We Were Wrong About the $90K/Month Wallet

- 频道：Adi
- 日期：2026-07-28
- URL：https://www.youtube.com/watch?v=FFwsW0TFdDU
- 判断：标题是 PnL 诱饵，正文是失败复盘，保留时把这个区分写死。最接近 BTC 5 分钟的实践尝试。速度优势的回测作者说留到下一集，本集没做。

实质：

1. 钱包 “Doggy Sty” 被说成大约每天 $5,000、每月约 $100,000。作者拉了订单簿历史（resolvemarkets.com 企业数据）和该钱包的成交。
2. 假设是两侧限价挂在 49/49。模拟里配对的两边能赚，未配对的残差到期亏掉，整本书不赚钱。
3. predicts.guru 上该钱包付出的费用大约 $265,000，净盈亏小于费用，等于在给费用打工。Maker 返佣的故事也对不上。
4. 手数是 50/100 的整数，不成交部分、到期前不卖。形状像 fill-or-kill，也可能是市价单（口播称成本大约 3%）。
5. 作者未检验的猜测：用比 Chainlink 更快的价格（Binance / Hyperliquid）去打陈旧的盘口，再对冲。需要 Rust 和托管。本集没有 100/200/300 毫秒的回测结果。

相关：没有残差规则的双侧做市应 skip。钱包盈亏不是策略。比结算预言机更快是延迟，不是 \(\sigma\)、动量、深度的融合；在那个速度回测出现之前，不能因为「别人每月九万」打开 trade。

---

## 谨慎：有一个可用数字，不单列、不采信收益

[Polymarket calibration curve](https://www.youtube.com/watch?v=_aEyhRnVifs)（FNRTrades，2026-06-19）。口播：2024-01 到 2026-04 约 41,000 个已结算二元，用结算前 1 小时的最终中间价。85–95 美分的 Yes 实际兑现 Yes 只有约 79%（大约 9 个点的缺口）；5–15 美分的 Yes 兑现约 14%（高了大约 4 个点）。后面接到 Kelly 淡出、月毛收益 4.1%、2026 年 3 月费率后 2.7%，以及推荐链接。模板和第 1 条相同。

可用的只有价格带否决：两尾报价相对兑现率偏了，高价热门和低价冷门默认 skip。月收益、Kelly、推荐链接不进入门控。

---

## 剔除

封面或标题已是保本/暴利，不采用（多数未再拉全字幕）：

| 视频 | 日期 | 原因 |
| --- | --- | --- |
| [$71,000 With a Polymarket Trading Bot…](https://www.youtube.com/watch?v=SufBKOtgUqQ) | 2026-08-25 | 封面大字 $71,000 profit。 |
| [This Polymarket Trading Bot has never lost](https://www.youtube.com/watch?v=tZpCEaEcVgM) | 2026-06-12 | 封面「never lost」加 $13,000。 |
| [Polymarket Bot That Steals 5% Every 5 Minutes](https://www.youtube.com/watch?v=iXBXUGAngbU) | 2026-06-07 | 封面承诺每 5 分钟稳定 5%。 |
| […Steal Free Money (Real PnL)](https://www.youtube.com/watch?v=0v8jAwOT5_Q) | 2026-06-19 | 封面 FREE MONEY。 |
| [I Made 45% ROI Per Week…](https://www.youtube.com/watch?v=lCpYtO9Lhtc) | 2026-06-10 | 封面 45% ROI per week。 |
| [$24 to $104,000 in 60 Days](https://www.youtube.com/watch?v=-01njdqdZBg) | 2026-01-17 | 标题即收益奇迹，带跟单链接。 |
| [A Polymarket Bot Made $438,000 In 30 Days](https://www.youtube.com/watch?v=BiqG3it0gY0) | 2026-04-07 | 标题以 30 天利润做钩子。 |
| [Make $200 in 5 Minutes / Arbitrage Bot Tutorial](https://www.youtube.com/watch?v=UuUo3f5BLYY) | 2026-07-29 | 描述直接写 5 分钟赚 $200。 |

看过字幕后剔除：

| 视频 | 日期 | 原因 |
| --- | --- | --- |
| [POLYMARKET BTC 5-MINUTE TRADING STRATEGY RESEARCH](https://www.youtube.com/watch?v=8u6jy8v56ww)（Amon） | 2026-02-15 | 封面像研究，机制却绑着不可信成绩：前 4 分钟同向则动量延续，口播给出约 78% 混合胜率、前两分钟波动超过 $100/$200 时 88%/93%，最大回撤 0.5%，12,272 段里最长连亏 6。后段又说盘口会被机器人吃穿、边缘会消失。动量加「书被吃空就别做」的形状和门控相近，但数字是营销，不保留。 |
| [Arbitrage vs Market Making vs Momentum](https://www.youtube.com/watch?v=NfEq_e5ImS0)（Altcoin Buzz） | 2026-02-18 | 钱包截图串：$63→$131k、$290 万、66% 胜率赚 $76 万、动量钱包 $160 万。15 分钟 BTC 上「波动把 No 打到几美分再买」没有 skip 条件。营销讲解。 |
| [Prediction market alpha: the 3.14% who Win…](https://www.youtube.com/watch?v=HoviJS2BGd4) | 2026-06-05 | 前半是延迟套利和买卖价差（费用约 2% 才打平），后半转到内幕起诉和刑期。不是 vol/动量/深度的 trade/skip 规则。 |
| [Can AI Beat the Prediction Market?](https://www.youtube.com/watch?v=a5eF_t7nOTU) | 2026-01-15 | 作者声明不是下注策略。多代理只做事件 Yes/No 和置信度，对照 ForecastBench。没有盘口、波动或暂停。 |
| [Why CLOB Beats AMM…](https://www.youtube.com/watch?v=jp8cA4q1vFw) | 2026-09-20 | 市场设计口播：多结果 CLOB 的加价阶梯，AMM 才能给出连贯概率。与加密二元的三特征门控无关。 |
| [PIN / informed flow shorts](https://www.youtube.com/watch?v=-ruDJe9JQzE)（FNRTrades；同题材还有 [2026-06-18](https://www.youtube.com/watch?v=WuMe1mXrJIQ)） | 2026-06-23 | 口播：前 50 名、6 个月，PIN 0.22，夹在纽交所 0.10 与 OTC 0.30 之间；结算前 60 分钟知情到达约 4.1 倍；60 美分、1 美分价差上不知情 maker 每笔知情成交亏约 0.38 美分；有效价差约 12 bp；最后 30 分钟整本撤掉。与第 2 条「符号修正后 \(\phi\approx 0\)」直接冲突，又是推荐链接模板。22% 知情和 30 分钟撤单不采用。 |
| [Shorter OBI cut](https://www.youtube.com/watch?v=BPUrnDH75H0)（FNRTrades） | 2026-06-17 | 第 1 条的短版，不重复计数。 |
| [Copy-trade leaderboard ratings](https://www.youtube.com/watch?v=T0JkMGeSW8o) | 2026-05-31 | 榜单口播：$158M / $149M / $1.14M。跟「最好的」在 $30 上亏了 $18（60%）。没有对方的风险过程。过程本身是 sizing，门控不做。 |
| [CLOB V2 migration](https://www.youtube.com/watch?v=Q2DDBqsbSoc)（Alphastack） | 2026-04-29 | 基础设施迁移，不是 trade / skip / pause。 |

标题或描述已是产品/暴利，本轮只做了标题与描述筛选，没有逐字稿，不保留：

| 视频 | 日期 | 原因 |
| --- | --- | --- |
| [HFT on Polymarket and Kalshi](https://www.youtube.com/watch?v=T6kehDDMpfs)（Apex Executions） | 2026-08-27 | 标题写 exploiting。 |
| [Crubble AI lives](https://www.youtube.com/watch?v=0mUzoW_avH0)（另有 [omJZtHKkFk4](https://www.youtube.com/watch?v=omJZtHKkFk4)、[kLLWctcKm7s](https://www.youtube.com/watch?v=kLLWctcKm7s)） | 2026 年夏 | 「第一只合法套利机器人」加组合证明。 |
| [EdgePilot](https://www.youtube.com/watch?v=mQczmB485ik) | — | 付款链接。 |
| [PolyBot free arb](https://www.youtube.com/watch?v=Lkp3U2sjybY) | — | 两侧都低于 50 美分就买。 |
| [Oxus Tech walkthrough](https://www.youtube.com/watch?v=5xWPM1Cxa6w) | — | 产品演示。 |
| [How to earn money](https://www.youtube.com/watch?v=J5C5MbBEgdY) | — | 赚钱标题。 |
| [PolyBot product demo](https://www.youtube.com/watch?v=D-t2ZS6q3WM) | — | 产品演示。 |

未拉到字幕、因此不保留：

- [vumaq Polymarket 5m BTC paper-trading live](https://www.youtube.com/watch?v=05TfuaCDROk)（2026-05-28）：时长 PT11H55M，字幕关闭。描述（不是逐字稿）写 5 分钟 BTC 的合流分数 0–25，来自 RSI、CVD、VWAP 偏离、OBI、taker flow、资金费率，八档（Draco 0–2 到 Cartwheel 20+），每笔纸面 $10，并有亏损加码。加码是 sizing / 回本，不在门控里。没有整场送去 AssemblyAI：watch 页不是媒体文件，12 小时直播的转写成本与「描述已经自曝加码」不对等。描述不能升成保留条。
- [Logic to Prototype](https://www.youtube.com/watch?v=6Nztuu021o4)（Ethan Sadleir，2026-08-12）：未拉逐字稿，不保留。
