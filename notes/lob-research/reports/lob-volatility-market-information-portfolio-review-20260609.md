# Limit Order Book 对波动率、市场信息与组合管理的价值：文献精读报告

日期：2026-06-09

本报告基于本地已下载的 Limit Order Book 论文与资料，重点回答你现在最关心的三个问题：

1. LOB 能不能帮助预测市场波动率或短期价格不稳定？
2. LOB 能不能反映市场信息、资金压力、买卖盘真实强弱？
3. LOB 对 portfolio，尤其是半导体和高 beta 科技组合，有什么实际好处？

报告结论先说清楚：LOB 不是一个可以直接告诉你“明天涨跌”的神奇数据源。它最有价值的地方，是在很短时间尺度上观察**流动性、订单流压力、价格冲击、短期方向风险、交易成本和组合内信息扩散**。对 portfolio 的价值更偏向风控和执行，而不是低频选股。

## 资料范围

本报告精读以下 7 篇本地 PDF：

| 文章 | 本地文件 |
| --- | --- |
| Gould et al. - Limit Order Books | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/gould-et-al-limit-order-books-survey-2013.pdf) |
| Cont, Kukanov, Stoikov - The Price Impact of Order Book Events | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/cont-kukanov-stoikov-price-impact-order-book-events-2010.pdf) |
| Xu, Gould, Howison - Multi-Level Order-Flow Imbalance in a Limit Order Book | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/xu-gould-howison-multi-level-ofi-2019.pdf) |
| Gould & Bonart - Queue Imbalance as a One-Tick-Ahead Price Predictor | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/gould-bonart-queue-imbalance-one-tick-ahead-2015.pdf) |
| Bonart & Gould - Latency and Liquidity Provision in a Limit Order Book | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/bonart-gould-latency-liquidity-provision-2016.pdf) |
| Zhang, Zohren, Roberts - DeepLOB | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/zhang-zohren-roberts-deeplob-2018.pdf) |
| Cont, Cucuringu, Zhang - Cross-Impact of Order Flow Imbalance | [PDF](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/cont-cucuringu-zhang-cross-impact-ofi-2023.pdf) |

对应全文文本在：

[fulltext 目录](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/fulltext)

## 总体判断

LOB 的价值可以分成四层。

第一层是**当前流动性状态**：spread 多宽、best bid/ask 有多厚、10 档深度是否偏向一侧。这是最直接的。

第二层是**短期价格压力**：订单流不平衡 OFI、队列不平衡 queue imbalance、多档订单流 MLOFI 能解释或预测很短窗口的 mid-price movement。

第三层是**市场信息**：大额订单、撤单、深层卖压、流动性突然消失，可能反映 adverse selection、信息性交易、等待成本、执行压力或短期情绪。但它不能告诉你交易者身份，也不能直接区分真实资金意图、spoofing、算法拆单或普通流动性调整。

第四层是**组合级信息扩散**：某些核心股票的订单流可能领先于同板块其他股票。例如半导体组合里，NVDA、AVGO、AMD、MU、TSM、ASML、AMAT、LRCX 等可能存在几分钟级 lead-lag。LOB 对 portfolio 的最大价值，是把单股盘口压力升级成组合风险监控和执行调度。

## 对你 portfolio 的直接含义

你的组合和关注方向里有较高的半导体、AI、杠杆或高 beta 暴露。根据已有本地记忆，曾经可见的组合权重中，`MVLL`、`ALAB`、`MSFU`、`MUU` 等半导体或杠杆科技相关仓位占比较高；这类仓位对短期流动性冲击和板块共振更敏感。这个事实可能已经过期，需要用当前持仓刷新，但方向上仍说明为什么 LOB 有用。

LOB 对这种组合的实际价值主要是：

- 判断加仓时机：如果价格下跌但 bid depth 仍稳定、spread 没扩、卖压没有持续补充，可能是短期冲击；如果 bid depth 消失、spread 扩大、ask depth 持续刷新，就更像流动性恶化。
- 判断减仓和执行成本：如果你要卖出高 beta 仓位，而盘口深度薄、OFI 偏负、market impact coefficient 高，就不适合用激进方式一次性卖。
- 判断板块风险扩散：如果 NVDA 或 AVGO 的 OFI 先出现异常卖压，然后 MU、MRVL、ALAB、SOXX 也跟随恶化，这比单只股票价格变动更早提示组合风险。
- 判断“市场信息”强弱：订单簿不是新闻，但它能反映新闻被交易者吸收后的微观行为。真正有价值的是压力是否持续、是否成交、是否撤单、是否扩散到相关股票。
- 做波动率预警：LOB 不能替代 VIX 或日频 realized volatility，但 depth 下降、spread 扩大、OFI 方差上升、队列极端失衡，通常意味着短期价格更容易跳动。

## 核心指标表

| 指标 | 数据来源 | 解释 | 对 portfolio 的用途 |
| --- | --- | --- | --- |
| Spread | `ask_px_00 - bid_px_00` | 交易摩擦和流动性紧张程度 | spread 扩大时减少激进交易 |
| Mid price | `(bid_px_00 + ask_px_00)/2` | 高频价格中心 | 计算短期收益和冲击 |
| Top depth | `bid_sz_00`, `ask_sz_00` | 最优档队列长度 | 判断下一跳方向和成交风险 |
| Queue imbalance | top bid/ask size | 最优队列买卖压力 | 一跳方向、短线择时 |
| 10 档 depth imbalance | `bid_sz_00...09`, `ask_sz_00...09` | 整个近端订单簿压力 | 判断压力是否只在顶层还是整本簿 |
| OFI | top-of-book 价格和数量变化 | 最优档净订单流压力 | 短期 price impact、执行成本 |
| MLOFI | 10 档价格和数量变化 | 多档净订单流压力 | 更完整的价格压力和流动性状态 |
| Liquidity refill speed | market order 后深度恢复时间 | 市场韧性 | 大单后是否继续追单 |
| Cross-asset OFI | 多股票 OFI lead-lag | 板块内信息扩散 | 半导体组合预警 |

## 文章一：Gould et al. - Limit Order Books

本地文件：[gould-et-al-limit-order-books-survey-2013.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/gould-et-al-limit-order-books-survey-2013.pdf)

### 研究定位

这是一篇 LOB 综述，不是一篇单一实验论文。它的价值是建立基本语言：什么是 LOB、为什么 LOB 是现代市场微观结构的核心、为什么它难建模、为什么它对波动率和风险有价值。

作者的基本观点是，LOB 是买卖双方通过限价单和市价单互动形成的动态系统。它不是静态表格，而是一个持续变化的状态空间。价格、深度、spread、撤单、订单到达率、成交行为、队列优先级都会互相影响。

### 历史背景

传统市场依赖 specialist 或 market maker 报价。电子交易普及后，许多市场转向 limit order book。LOB 的重要变化是：流动性提供不再只由少数指定做市商完成，而是由所有愿意提交限价单的交易者共同形成。

这带来一个关键结果：流动性变成了自组织的、可撤回的、战略性的。表面上订单簿很厚，不代表下一秒还会厚；看到大单或信息冲击后，限价单可以快速撤走。

### 对波动率的观点

综述里明确讨论了 volatility。作者强调 volatility 是风险暴露和 portfolio construction 的重要输入，但高频 LOB 环境下，波动率估计并不简单。不同采样频率、bid price、ask price、mid price 会得到不同 realized volatility。

更重要的是，普通价格序列只看到结果，LOB 能看到结果背后的状态。例如：

- spread 扩大，表示市场对风险补偿要求提高；
- depth 变薄，表示同样订单流会造成更大价格冲击；
- cancellation 增加，表示流动性供给变得不稳定；
- order flow clustering 增强，表示事件更集中，短期跳动风险更高。

所以 LOB 对波动率预测的贡献不是“直接替代 GARCH”，而是给出价格波动的微观原因。

### 对市场信息的观点

LOB 里的订单不是完全真实意图的透明窗口。限价单可能是：

- 真实交易意愿；
- 暂时提供流动性的做市行为；
- 信息交易者为了隐藏意图而拆分订单；
- 高速策略为抢队列位置而挂单；
- 试探性挂单或很快撤销的订单。

因此，不能把一个大 ask 简单解释成“大资金要卖”。更严谨的做法是看它是否持续、是否成交、是否撤单、撤单后价格是否移动、同板块是否同步出现压力。

### 对 portfolio 的价值

综述的实际启发是：LOB 更适合 portfolio 的微观执行和风险管理，而不是直接替代基本面、宏观或日频因子。你可以用它回答：

- 当前这个股票能不能交易这么多量？
- 现在买入会不会推高价格？
- 这个价格下跌是成交导致，还是流动性撤走导致？
- 同一个 portfolio 里哪些股票现在最不适合加仓或减仓？

### 局限

综述强调 LOB 研究很容易受数据质量、交易所规则、隐藏流动性、采样频率和建模假设影响。跨市场、跨年份、跨股票直接套模型很危险。

## 文章二：Cont, Kukanov, Stoikov - The Price Impact of Order Book Events

本地文件：[cont-kukanov-stoikov-price-impact-order-book-events-2010.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/cont-kukanov-stoikov-price-impact-order-book-events-2010.pdf)

### 研究问题

这篇文章问的是：短时间尺度上的价格变化，到底是成交量驱动，还是订单簿供需变化驱动？

作者认为只看成交量或 trade imbalance 不够，因为价格形成不仅来自 market order，也来自 limit order 和 cancellation。一个大撤单可能和一个大市价单一样改变价格压力。

### 数据

数据是 2010 年 4 月一个月内随机选取的 50 只 S&P 500 成分股，来自 NYSE TAQ consolidated quotes 和 consolidated trades。TAQ 是 Level 1 数据，只包含 best bid、best ask、对应数量和成交记录，不包含完整 10 档或逐订单 L3。

这个历史背景很重要：即使没有完整 L2/L3，作者也能用 top-of-book 构造有力的 OFI。这说明你用 Databento `mbp-10` 做 MRVL 时，数据粒度已经超过这篇早期经典论文。

### 方法

作者定义订单流不平衡 OFI。每次最优买卖盘更新后，观察 best bid/ask price 和 size 的变化。单个事件的贡献为：

\[
e_n =
\mathbf{1}_{\{P_n^B \ge P_{n-1}^B\}} q_n^B
- \mathbf{1}_{\{P_n^B \le P_{n-1}^B\}} q_{n-1}^B
- \mathbf{1}_{\{P_n^A \le P_{n-1}^A\}} q_n^A
+ \mathbf{1}_{\{P_n^A \ge P_{n-1}^A\}} q_{n-1}^A
\]

窗口内：

\[
OFI_k = \sum_{n=N(t_{k-1})+1}^{N(t_k)} e_n
\]

价格变化用 mid-price：

\[
\Delta P_k = \frac{P_k - P_{k-1}}{\delta}
\]

核心回归是：

\[
\Delta P_k = \beta OFI_k + \epsilon_k
\]

### 主要结果

在 10 秒窗口上，OFI 与 mid-price change 的线性关系很强，平均 \(R^2\) 约为 \(65\%\)。传统 trade imbalance 的平均 \(R^2\) 约为 \(32\%\)。也就是说，订单簿净供需变化比成交量本身更接近价格形成机制。

作者还发现价格冲击系数 \(\beta\) 与市场深度近似反比：

\[
\beta_i = \frac{c}{AD_i^\lambda} + \nu_i
\]

经验上 \(\lambda\) 接近 1。直觉很简单：订单簿越浅，同样的订单流不平衡越容易推动价格。

### 对波动率的价值

这篇文章最重要的 volatility 启发是：日内波动不是只由“信息多不多”解释，也可以由两个可观测变量解释：

- \(OFI\) 的波动；
- depth 的变化。

早盘订单簿浅，\(\beta\) 高，所以同样 OFI 会造成更大价格跳动。收盘附近 OFI 可能更活跃，但 depth 也更高，部分抵消冲击。这给波动率预测提供了微观结构版本：

\[
ShortTermVolatility \approx f(Var(OFI), Depth, Spread)
\]

### 对市场信息的价值

OFI 是比成交量更好的“市场压力”变量。成交量只告诉你发生了多少交易，OFI 同时吸收了：

- 买盘新增；
- 卖盘新增；
- 买盘撤单；
- 卖盘撤单；
- 市价单吃掉队列。

所以如果你想判断“资金是在压低价格，还是只是成交噪声”，OFI 比成交量更接近答案。

### 对 portfolio 的价值

这篇文章对 portfolio 的实用价值包括：

- 估计短期 market impact；
- 比较不同股票当前流动性风险；
- 在组合再平衡时决定哪些股票先交易、哪些股票慢慢交易；
- 当 \(\beta\) 高、depth 低、OFI 偏负时，避免激进卖出；
- 对杠杆半导体仓位，用 OFI 监控日内下跌是否伴随真实卖压。

### 映射到 Databento

Databento `mbp-10` 可以直接构造这篇的 Level I OFI：

| 论文变量 | Databento 字段 |
| --- | --- |
| \(P^B\) | `bid_px_00` |
| \(q^B\) | `bid_sz_00` |
| \(P^A\) | `ask_px_00` |
| \(q^A\) | `ask_sz_00` |
| event time | `ts_event` |
| mid | `(bid_px_00 + ask_px_00)/2` |

如果你只是要解释价格压力，`mbp-10` 足够。如果你要区分新增、撤单、成交和修改，应该用 `mbo` 重建订单簿。

## 文章三：Xu, Gould, Howison - Multi-Level Order-Flow Imbalance

本地文件：[xu-gould-howison-multi-level-ofi-2019.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/xu-gould-howison-multi-level-ofi-2019.pdf)

### 研究问题

这篇文章扩展 Cont 的 OFI。问题是：只看 best bid/ask 是否遗漏了深层订单簿信息？

作者提出 multi-level order-flow imbalance，简称 MLOFI。它不只看 level 1，而是在前 \(M\) 个 price levels 上分别计算 net order flow。

### 数据

数据来自 LOBSTER，覆盖 Nasdaq 上 6 只流动性股票：AMZN、TSLA、NFLX、ORCL、CSCO、MU。时间为 2016 年全年，使用每个交易日 10:00 到 15:30，排除开盘和收盘附近 30 分钟。

样本覆盖 small-tick 和 large-tick 股票。这个区分非常关键，因为不同 tick regime 下订单簿结构完全不同。

### 方法

对第 \(m\) 档 bid 和 ask，分别计算买侧和卖侧净变化，再得到：

\[
e^m(\tau_n)=\Delta W^m(\tau_n)-\Delta V^m(\tau_n)
\]

窗口内：

\[
MLOFI^m(t_{k-1},t_k)=\sum_{t_{k-1}<\tau_n\le t_k} e^m(\tau_n)
\]

回归：

\[
\Delta P(t_{i,k-1},t_{i,k})
=
\alpha
+ \sum_{m=1}^{M}\beta^m MLOFI^m(t_{i,k-1},t_{i,k})
+ \epsilon
\]

当 \(M=1\) 时，它就是 Cont 的 OFI。

### 主要结果

作者发现深层订单流确实重要，但必须小心建模。因为不同档位的 MLOFI 高度相关，OLS 会不稳定，容易误判深层不重要。使用 Ridge regression 后，深层参数显著性提高，out-of-sample RMSE 随 \(M\) 增大而下降。

与只用 OFI 相比，\(M=10\) 的 Ridge MLOFI 将 out-of-sample RMSE 降低约：

| 股票 | 改善 |
| --- | ---: |
| AMZN | 17% |
| TSLA | 15% |
| NFLX | 31% |
| ORCL | 68% |
| CSCO | 74% |
| MU | 64% |

large-tick 股票改善更大。原因是 large-tick 股票的最优价附近堆积了大量队列，深层流动性状态对未来价格形成更重要。

### 对波动率的价值

MLOFI 的 volatility 含义比 OFI 更强。只看 top-of-book 时，你可能以为卖压很小；但如果第 2 到第 10 档 ask 侧持续补量，而 bid 侧深层撤单，未来价格跳动风险会升高。

因此，对波动率预警可以从：

\[
Var(OFI)
\]

扩展到：

\[
Var(MLOFI^1,\ldots,MLOFI^{10})
\]

以及深层流动性倾斜：

\[
DepthImbalance_t =
\frac{\sum_{i=0}^{9} bid\_sz_i - \sum_{i=0}^{9} ask\_sz_i}
{\sum_{i=0}^{9} bid\_sz_i + \sum_{i=0}^{9} ask\_sz_i}
\]

### 对市场信息的价值

深层订单不一定马上成交，但它可能反映交易者对中短期价格的预期。知情交易者不一定用市价单暴露意图，也可能在深层挂单、撤单、调整价格来减少信息泄露。

所以深层 LOB 有两个信息来源：

- 机械流动性：如果顶层被吃掉，深层决定下一跳价格；
- 战略信息：交易者把预期表达在更深价格层。

这对判断“市场是不是正在变坏”很重要。真正危险的不是 level 0 一瞬间变薄，而是多档 bid 同时撤退、ask 多档持续补量。

### 对 portfolio 的价值

MLOFI 比 OFI 更适合 portfolio execution。你可以用它给组合里的股票排序：

- 哪只股票当前最不适合卖？
- 哪只股票买入冲击成本最低？
- 哪只股票盘口压力只是 top-of-book 噪声？
- 哪只股票整本订单簿都在偏向卖压？

对于半导体组合，`MU` 出现在该论文样本中，这是一个很实用的连接。你可以先用 MRVL、MU、NVDA、AVGO 做类似指标，比较各自 10 档压力。

### Databento 落地

`mbp-10` 与这篇文章非常匹配：

| MLOFI 变量 | Databento 字段 |
| --- | --- |
| \(b^m\) | `bid_px_00` 到 `bid_px_09` |
| \(r^m\) | `bid_sz_00` 到 `bid_sz_09` |
| \(a^m\) | `ask_px_00` 到 `ask_px_09` |
| \(q^m\) | `ask_sz_00` 到 `ask_sz_09` |

实践上先用 MBP-10 做聚合 MLOFI；如果要精确复现 LOBSTER 逐事件逻辑，再用 MBO。

## 文章四：Gould & Bonart - Queue Imbalance

本地文件：[gould-bonart-queue-imbalance-one-tick-ahead-2015.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/gould-bonart-queue-imbalance-one-tick-ahead-2015.pdf)

### 研究问题

这篇文章问的是：最优买卖队列不平衡能不能预测下一次 mid-price 上移还是下移？

它不是预测日收益，也不是预测分钟收益，而是 one-tick-ahead 的微观方向预测。

### 数据

数据来自 LOBSTER，覆盖 Nasdaq 2014 年全年。作者排除开盘后 30 分钟和收盘前 30 分钟，只看 10:00 到 15:30。

样本包含 10 只高流动性股票：

- large-tick：MSFT、INTC、MU、CSCO、ORCL；
- small-tick：GOOG、AMZN、TSLA、PCLN、NFLX。

### 方法

队列不平衡定义为：

\[
I(t)=
\frac{n_b(b_t,t)-n_a(a_t,t)}
{n_b(b_t,t)+n_a(a_t,t)}
\]

\(I\) 接近 1 表示 best bid queue 比 best ask queue 长很多，买方压力强；接近 -1 表示卖方压力强。

作者用 logistic regression 估计：

\[
\hat y(I)=\frac{1}{1+\exp(-(x_0+x_1 I))}
\]

其中 \(\hat y(I)\) 是下一次 mid-price 上移概率。

### 结果

所有 10 只股票的 \(x_1\) 都为正，且在 99% 水平显著。也就是说，队列不平衡越偏买方，下一跳价格上移概率越高。

大 tick 股票效果更强：

- \(I\) 接近 1 时，大 tick 股票价格上移概率约 0.8 到 0.9；
- 小 tick 股票约 0.6 到 0.7；
- 大 tick 股票 ROC AUC 约 0.7 到 0.8；
- 小 tick 股票约 0.6 到 0.65。

### 对波动率的价值

这篇文章不直接预测 realized volatility，但它说明了一个重要事实：当 queue imbalance 极端时，下一跳方向更可预测，同时队列耗尽风险更集中。

从 volatility 角度看，极端 imbalance 意味着市场更接近一次微观价格跳动。连续极端 imbalance 出现时，短期价格路径更不稳定。

### 对市场信息的价值

Queue imbalance 是市场压力的低维表达。它不告诉你是谁在买卖，但能告诉你：

- 哪一侧更容易被耗尽；
- 当前价格移动的微观倾向；
- 订单簿压力是否足够极端。

作者还指出截距偏移可能来自外生新闻或方向性买卖压力。也就是说，LOB 信号会吸收新闻和私有信息交易后的行为结果。

### 对 portfolio 的价值

Queue imbalance 最适合做执行层信号：

- 买入时，如果 \(I\) 很高，价格可能更容易上跳，等待可能变贵；
- 卖出时，如果 \(I\) 很低，卖出冲击风险更高；
- 如果 \(I\) 接近 0，方向信号弱，可以避免过度反应；
- 对高 beta 仓位，可以用极端 \(I\) 作为短期风险预警。

## 文章五：Bonart & Gould - Latency and Liquidity Provision

本地文件：[bonart-gould-latency-liquidity-provision-2016.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/bonart-gould-latency-liquidity-provision-2016.pdf)

### 研究问题

这篇文章研究 market order 到达前后，限价单流动性如何变化。它关心的是反馈环：market order 不只是消耗流动性，还会改变其他交易者之后如何补单、撤单、重新排队。

### 数据

数据来自 LOBSTER，时间为 2015 年 3 月 1 日到 2015 年 9 月 1 日。样本为 5 只 Nasdaq 大 tick 高流动性股票：MSFT、INTC、YHOO、MU、CSCO。

### 方法

作者观察 market order 到达前后，best bid/ask 队列的累计净订单流。对 market order 时间 \(t_i\)，定义：

\[
W^B(t_i,\tau)=V^B(t_i+\tau)-V^B(t_i)
\]

\[
W^A(t_i,\tau)=V^A(t_i+\tau)-V^A(t_i)
\]

重点研究 price-maintaining market orders，也就是不会马上改变 best bid/ask 的 market orders。

### 主要结果

market order 到达后，same-side best quote 出现四个阶段：

1. 平台延迟：极短时间内净订单流为 0；
2. 响应延迟：净流很小；
3. 高速反应：same-side 大量撤单；
4. 较慢反应：流动性重新补充。

opposite-side 也有类似阶段，但方向不同：market order 后 opposite-side 往往快速补单，然后因为等待成本上升又逐步撤出。

### 对市场信息的价值

这篇文章最重要的观点是：成交后流动性不是平滑恢复，而是策略性变化。大 market order 会让同侧剩余限价单担心 adverse selection，于是撤单；之后才可能 stimulated refill。

这说明“盘口厚度”不是静态安全垫。真正要看的是冲击后是否补得回来，以及撤单速度有多快。

### 对波动率的价值

短期波动很多时候来自流动性突然撤走，而不只是成交本身。market order 后如果 same-side liquidity 快速撤出，下一笔订单更容易推动价格跳动。这是 LOB 对 volatility risk 的重要贡献。

### 对 portfolio 的价值

对大仓位、杠杆仓位、半导体波动仓位，这篇文章很有用：

- 大单成交后不要假设流动性马上恢复；
- 如果 market order 后同侧撤单明显，继续追单可能冲击成本很高；
- 可以估计 liquidity refill speed，作为交易执行节奏；
- 对风控来说，大单后撤单强度是短期流动性风险指标。

## 文章六：Zhang, Zohren, Roberts - DeepLOB

本地文件：[zhang-zohren-roberts-deeplob-2018.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/zhang-zohren-roberts-deeplob-2018.pdf)

### 研究问题

DeepLOB 问的是：能不能直接从原始 LOB 数据里学习未来短期价格方向，而不是手工构造 OFI、imbalance、spread 等特征？

作者提出 CNN + Inception + LSTM 架构。CNN 抓订单簿空间结构，LSTM 抓时间依赖。

### 数据

两类数据：

- FI-2010：Nasdaq Nordic 5 只股票，10 个交易日；
- LSE 数据：2017 年伦敦证券交易所 5 只高流动性股票全年数据。

LSE 数据每个时间点包含 10 档买卖盘，每档价格和数量，共 40 个特征。输入是最近 100 个订单簿状态，形状为 \(100 \times 40\)。

### 方法

标签是未来 mid-price 方向：上涨、持平、下跌。作者使用平滑后的未来/历史 mid-price 均值，减少高频噪声。

模型结构：

1. 卷积模块：提取价格层级和买卖盘局部结构；
2. Inception 模块：捕捉不同尺度模式；
3. LSTM：捕捉时间序列依赖。

### 结果

在 FI-2010 上，DeepLOB 超过多个已有方法。在 LSE 一年数据上，对训练股票和未参与训练的股票都有较稳定的 out-of-sample 表现。

对于参与训练的 5 只 LSE 股票，预测 horizon \(k=20,50,100\) 时，accuracy 大约为 70.17%、63.93%、61.52%。对未参与训练的 5 只股票，也能达到约 68.62%、63.44%、61.46%。

这说明 LOB 中存在一定跨股票可迁移的微观结构模式。

### 对波动率和市场信息的价值

DeepLOB 的意义不是某个公式，而是证明原始 LOB 状态包含可学习的短期价格信息。多档价格、数量、时间演化、局部形态共同构成信号。

它也提醒我们：手工指标不一定捕捉全部信息。例如：

- deep levels 的形态；
- 不同层之间的相对变化；
- 最近 100 个状态的动态模式；
- 价格和数量的组合形态。

这些都可能影响短期方向和 volatility clustering。

### 对 portfolio 的价值

DeepLOB 对 portfolio 的价值是建立单股微观结构预警模型：

- 每只股票输出短期上涨/持平/下跌概率；
- 按模型置信度排序组合内股票；
- 在组合再平衡前避开不利盘口；
- 对半导体组合，用统一模型比较 NVDA、AVGO、MU、MRVL、ALAB 等的盘口压力。

但它不应直接用于真实交易决策，因为论文交易模拟简化，没有完整考虑交易成本、排队、滑点、延迟和冲击。

### 局限

深度模型过拟合风险高，数据成本和计算成本高。更重要的是，可解释性有限。对你的项目来说，DeepLOB 可以作为后续路线，不应作为第一步。第一步应先做 OFI、MLOFI、imbalance、depth 这些可解释指标。

## 文章七：Cont, Cucuringu, Zhang - Cross-Impact of OFI

本地文件：[cont-cucuringu-zhang-cross-impact-ofi-2023.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/cont-cucuringu-zhang-cross-impact-ofi-2023.pdf)

### 研究问题

这篇文章问的是：一个股票的订单流不平衡，会不会影响其他股票？也就是 cross-impact。它还问：跨股票 OFI 能不能预测未来短期收益？

这篇对 portfolio 最直接，因为 portfolio 本来就是多资产问题。

### 数据

数据来自 LOBSTER 的 Nasdaq ITCH，样本为 S&P 500 中市值前 100 股票，时间为 2017-01-01 到 2019-12-31。作者使用前 10 档订单簿数据，按 1 分钟构造多档 OFI、integrated OFI 和收益率。排除开盘后 30 分钟和收盘前 30 分钟。

样本中包括 NVDA、QCOM、TXN、AVGO、AAPL 等科技和半导体相关股票。

### 方法

作者先计算每档 OFI，再用 PCA 第一主成分整合成 integrated OFI。原因是 10 档 OFI 高度相关，第一主成分平均解释超过 89% 方差。

然后比较：

- 单股 best-level OFI 模型；
- 单股 integrated OFI 模型；
- 加入其他股票 OFI 的 cross-impact 模型；
- 滞后 OFI 对未来收益的预测模型。

由于变量很多，作者使用 LASSO 做稀疏选择。

### 主要结果

当期收益解释：

- 只用 best-level OFI，加入 cross-impact 能提高解释力；
- 但使用 integrated OFI 后，单股自身多档信息已经解释了大部分当期收益，cross-impact 对单股的额外帮助很小。

组合层面不同。对等权组合和 eigenportfolio，cross-impact 明显更有价值。使用 integrated OFI 时，等权组合 out-of-sample \(R^2\) 从约 85.26% 提高到 87.97%，eigenportfolio 从 84.69% 提高到 87.70%。

未来收益预测：

- cross-asset lagged OFI 有短期预测价值；
- 预测力集中在 1 分钟到数分钟；
- horizon 拉长到 10、30 分钟后快速衰减；
- 交易策略上的 PnL 有改善，但论文也承认统计 \(R^2\) 很弱，且成本未完整计入。

### 对波动率和市场信息的价值

这篇文章说明订单流信息会跨股票传播。对半导体组合，这非常重要。单看 MRVL 可能太慢；如果 NVDA、AVGO 或 MU 的 OFI 已经先恶化，MRVL 可能之后才反映。

市场信息可能通过这些路径扩散：

- ETF 或行业篮子交易；
- 统计套利和 pairs trading；
- 共同基本面主题，如 AI capex、HBM、晶圆代工、设备订单；
- 投资者注意力和信息逐步反应；
- 龙头股先反映，产业链后反映。

### 对 portfolio 的价值

这是最适合 portfolio 的 LOB 论文。它给出一个组合监控框架：

1. 对每只股票计算 integrated OFI；
2. 建立 cross-impact network；
3. 找出信息输出节点和被影响节点；
4. 用短 horizon 信号监控板块风险扩散；
5. 在组合再平衡和对冲时，参考当前 cross-asset OFI。

对半导体组合，建议重点监控：

- 龙头：NVDA、AVGO、AMD；
- 存储和周期：MU、MRVL；
- 设备：AMAT、LRCX、ASML；
- 代工和 ADR：TSM；
- ETF 或杠杆产品：SOXX、SOXL、SOXS、MVLL、MSFU、MUU。

### 局限

信号很短，主要是几分钟级；交易成本可能吞掉收益；高维模型容易过拟合；跨市场半导体组合还涉及不同交易所、时区、ADR、本地股和货币。

因此，它最适合做 portfolio risk monitor，而不是直接做高频交易策略。

## 综合框架：如何把 LOB 用在你的研究里

### 第一步：单股 LOB 健康度

对每个核心持仓或 watchlist 股票，计算：

\[
spread_t = ask\_px\_00 - bid\_px\_00
\]

\[
mid_t = \frac{ask\_px\_00 + bid\_px\_00}{2}
\]

\[
QI_t =
\frac{bid\_sz\_00 - ask\_sz\_00}
{bid\_sz\_00 + ask\_sz\_00}
\]

\[
DI_t =
\frac{\sum_{i=0}^{9} bid\_sz_i - \sum_{i=0}^{9} ask\_sz_i}
{\sum_{i=0}^{9} bid\_sz_i + \sum_{i=0}^{9} ask\_sz_i}
\]

这些指标回答：当前盘口是偏买、偏卖，还是只是正常噪声？

### 第二步：订单流压力

构造 OFI：

\[
\Delta mid_t \approx \beta OFI_t
\]

构造 MLOFI：

\[
\Delta mid_t \approx \alpha + \sum_{m=1}^{10}\beta^m MLOFI^m_t
\]

这些指标回答：价格变化是否由真实订单簿压力解释？

### 第三步：短期波动风险

你可以把短期 volatility risk 拆成：

\[
VolRisk_t =
g(Spread_t, Depth_t, Var(OFI)_t, Var(MLOFI)_t, RefillSpeed_t)
\]

直觉：

- spread 扩大，交易成本和不确定性上升；
- depth 下降，同样订单流造成更大价格跳动；
- OFI 方差上升，短期压力更不稳定；
- refill speed 变慢，市场韧性下降；
- 多档 ask/bid 同向偏斜，未来跳动风险更高。

### 第四步：portfolio 层 cross-impact

对组合内股票构造一个矩阵：

| 来源股票 OFI | 目标股票未来收益 |
| --- | --- |
| NVDA OFI | MRVL future return |
| AVGO OFI | MU future return |
| MU OFI | MVLL future return |
| SOXX OFI | ALAB future return |

模型可以很简单：

\[
r_{i,t+1:t+h}
=
\alpha_i
+ \sum_j \gamma_{ij} OFI_{j,t}
+ \epsilon_{i,t}
\]

如果 \(\gamma_{ij}\) 稳定显著，说明股票 \(j\) 的订单流对股票 \(i\) 有短期信息领先作用。

## 对 Databento 的落地路线

### 先区分 dataset 和 schema

在 Databento 里，`dataset` 和 `schema` 是两层概念。

- `dataset` 表示数据从哪个交易所或哪个行情 feed 来。
- `schema` 表示你要这个 feed 里的哪种数据形态。

例如：

```python
dataset="XNAS.ITCH"
schema="mbp-10"
```

意思是：从 Nasdaq TotalView-ITCH 这个数据源里，拿前 10 档聚合订单簿。

一个简单类比：

| 概念 | 含义 | 例子 |
| --- | --- | --- |
| Dataset | 数据来源、交易所、行情 feed | `XNAS.ITCH`, `ARCX.PILLAR`, `GLBX.MDP3` |
| Schema | 数据形态、字段结构、深度 | `mbo`, `mbp-10`, `mbp-1`, `trades`, `ohlcv-1m` |

所以不是所有 dataset 都支持所有 schema。原因是原始交易所 feed 自己提供的数据深度不同。如果一个 feed 本身只提供 top-of-book，就不可能变出 10 档订单簿或逐订单数据。

### 常见 dataset 的区别

| Dataset | 是什么 | 市场 | 是否适合 LOB 研究 |
| --- | --- | --- | --- |
| `XNAS.ITCH` | Nasdaq TotalView-ITCH | 美股 Nasdaq | 非常适合，支持 `mbo` 和 `mbp-10` |
| `ARCX.PILLAR` | NYSE Arca Pillar | 美股 NYSE Arca | 适合，支持 `mbo` 和 `mbp-10` |
| `XNYS.PILLAR` | NYSE Pillar | 美股 NYSE 主交易所 | 适合，支持 `mbo` 和 `mbp-10` |
| `GLBX.MDP3` | CME Globex MDP 3.0 | 期货，例如 ES、NQ、CL | 适合期货 LOB，不是股票 |
| `IEXG.TOPS` | IEX TOPS | IEX top-of-book 美股行情 | 不适合 10 档 LOB，只适合浅层报价和成交 |

这几个 dataset 的核心差别不是“哪个好”，而是它们代表不同 venue。

`XNAS.ITCH` 是 Nasdaq 订单簿。对于 `MRVL`、`NVDA`、`MU`、`AVGO` 这种科技和半导体股票，它通常是最自然的第一站。

`ARCX.PILLAR` 是 NYSE Arca。Arca 是很重要的电子交易 venue，ETF 和美股流动性都很多。如果你想研究跨交易所承接和流动性扩散，可以加入它。

`XNYS.PILLAR` 是 NYSE 主交易所。它对 NYSE-listed 股票尤其重要；对 Nasdaq-listed 股票，是否是优先数据源要看该股票在哪些 venue 上实际流动性更强。

`GLBX.MDP3` 是 CME 期货。它适合研究 `ES`、`NQ`、利率、原油等期货 LOB。如果你想把股票 LOB 和 Nasdaq futures / E-mini futures 联动起来，它会有用；但它不是 MRVL/NVDA 这类股票订单簿。

`IEXG.TOPS` 是 IEX 的 top-of-book feed。它只提供较浅数据，例如 `mbp-1`、`tbbo`、`trades` 等。它不支持 `mbo` 或 `mbp-10`，不是因为 Databento 故意不给，而是这个 dataset 本身不是深度订单簿 feed。

### 常见 schema 的区别

| Schema | 粒度 | 能看到什么 | 适合做什么 |
| --- | --- | --- | --- |
| `mbo` | L3，Market by Order | 单个订单新增、修改、撤单、成交 | 重建订单簿、追踪大单生命周期、研究撤单 |
| `mbp-10` | L2，Market by Price 10 档 | 买卖各 10 档价格、数量、订单数 | 订单簿压力、深度失衡、MLOFI、短期波动风险 |
| `mbp-1` | L1，Market by Price 1 档 | 最优买价、最优卖价和数量 | spread、top-of-book imbalance |
| `tbbo` | Trade with BBO | 成交发生时的 best bid/ask | 判断成交发生在 bid/ask 附近 |
| `trades` | 成交记录 | 成交价、成交量、时间 | 成交量、主动买卖压力粗略研究 |
| `ohlcv-*` | K 线聚合 | open、high、low、close、volume | 低频回测、趋势、波动率基础特征 |

更深的数据可以生成更浅的数据，但反过来不行。

例如，`mbo` 理论上可以重建 `mbp-10`，也可以生成 `mbp-1`。但是只有 `mbp-1` 时，你无法还原第 2 到第 10 档，更无法知道单个订单是不是被撤掉。

可以理解成：

\[
mbo \rightarrow mbp\text{-}10 \rightarrow mbp\text{-}1 \rightarrow trades/OHLCV
\]

信息会越来越少。

### 为什么有的支持，有的不支持？

是否支持 `mbo` 或 `mbp-10`，取决于原始 feed 是否包含深度订单簿事件。

如果一个 dataset 原始 feed 有逐订单或深度档位信息，就可以支持：

- `mbo`
- `mbp-10`
- `mbp-1`
- `trades`
- `ohlcv`

如果一个 dataset 原始 feed 只发布最优买卖价，就只能支持：

- `mbp-1`
- `tbbo`
- `trades`
- `ohlcv`

它不能支持：

- `mbo`
- `mbp-10`

因为这些信息从源头就没有。

### 对当前研究的选择

你现在研究的是半导体股票、科技股回调、无明显坏消息时的资金撤退和 LOB stress。因此优先级应该是：

1. 先用 `XNAS.ITCH` + `mbp-10`。
2. 对 `MRVL`、`NVDA`、`MU`、`AVGO` 计算 spread、depth、OFI、MLOFI。
3. 如果要追踪“大单挂了又撤”，再改用 `XNAS.ITCH` + `mbo`。
4. 如果要研究跨交易所流动性，再加入 `ARCX.PILLAR` 和 `XNYS.PILLAR`。
5. 如果研究指数期货或科技股与期货联动，再加入 `GLBX.MDP3`。
6. 不要用 `IEXG.TOPS` 做 10 档 LOB，因为它没有这个深度。

最务实的第一阶段组合是：

```text
dataset = XNAS.ITCH
schema  = mbp-10
symbols = MRVL, NVDA, MU, AVGO
```

等这个跑通后，再扩展到多 dataset 和 `mbo`。

### 阶段 1：用 `mbp-10` 做可解释指标

先不要直接上深度学习。用 Databento `XNAS.ITCH` 的 `mbp-10` 做：

- spread；
- mid；
- top queue imbalance；
- 10 档 depth imbalance；
- OFI；
- MLOFI；
- short-window realized volatility。

优先标的：

- MRVL；
- NVDA；
- MU；
- AVGO；
- ALAB；
- SOXX 或相关 ETF。

### 阶段 2：事件研究

定义事件：

- ask depth 突然增加；
- bid depth 突然消失；
- OFI 极端负值；
- MLOFI 多档同向卖压；
- spread 突然扩大。

观察后续：

- 1 秒；
- 5 秒；
- 30 秒；
- 1 分钟；
- 5 分钟。

结果变量：

- mid return；
- realized volatility；
- spread 变化；
- depth 恢复速度；
- 是否扩散到相关股票。

### 阶段 3：用 `mbo` 研究大单生命周期

当你想问“大单是不是真的”“是不是挂了又撤”“有没有 spoofing-like 行为”，`mbp-10` 不够。要用 `mbo`：

- 订单新增；
- 修改；
- 撤单；
- 成交；
- order id 生命周期；
- 队列位置；
- 大单持续时间。

`mbo` 更贵、更复杂，但对“真实意图”更接近。

### 阶段 4：组合监控

最后再做 cross-impact：

- 每只股票每分钟 integrated OFI；
- 股票之间 lead-lag；
- 板块风险扩散图；
- portfolio 加权风险；
- 根据仓位权重和流动性风险排序。

对你的 portfolio，不应该只看哪个股票新闻最热，而应该看：

\[
PortfolioLOBStress_t =
\sum_i w_i \cdot LOBStress_{i,t}
\]

其中 \(w_i\) 是组合权重，\(LOBStress_{i,t}\) 可以由 spread、depth、OFI、MLOFI、cross-impact risk 综合得到。

## 最重要的实践结论

1. LOB 对预测波动率有帮助，但主要是短期微观波动，不是直接预测月度或季度波动。
2. OFI 和 depth 是最核心的两个变量：一个代表压力，一个代表承受压力的能力。
3. 多档订单簿比 top-of-book 更有价值，尤其是判断卖压是否深入整本簿。
4. Queue imbalance 可以预测下一跳方向，但对小 tick 股票效果弱一些。
5. Market order 后流动性可能先撤出再恢复，不能把当前深度当作稳定安全垫。
6. DeepLOB 说明 LOB 原始矩阵有预测信息，但第一步不建议直接上黑箱模型。
7. Cross-asset OFI 对 portfolio 最有价值，尤其是半导体这种主题和产业链高度联动的组合。
8. LOB 最适合用于执行、风控、异常预警和组合短线压力监控，而不是单独决定中长期投资。

## 具体到你的研究问题

### 我能不能用 LOB 判断市场波动率？

可以，但要限定在短期。你可以用：

- spread；
- depth；
- OFI variance；
- MLOFI variance；
- queue imbalance extremeness；
- liquidity refill speed。

这些可以预测未来几秒到几分钟的微观价格不稳定，帮助你判断“现在是不是容易跳价”。

### 我能不能用 LOB 判断市场信息？

可以部分判断，但不能过度解释。

LOB 可以告诉你：

- 市场压力是否偏买或偏卖；
- 卖压是否持续；
- 大单是否撤掉；
- 成交后流动性是否恢复；
- 相关股票是否同步恶化。

LOB 不能直接告诉你：

- 谁在交易；
- 是机构还是散户；
- 是真实卖出还是策略性挂单；
- 是否一定有内幕信息；
- 价格明天一定怎么走。

### LOB 对 portfolio 的好处是什么？

最有价值的三件事：

1. **执行成本控制**：决定何时加仓、减仓、拆单、挂单还是吃单。
2. **短期风险预警**：发现高权重股票的流动性恶化和卖压扩散。
3. **组合联动监控**：识别 NVDA/AVGO/MU/MRVL/ALAB/SOXX 等之间的订单流 lead-lag。

## 推荐下一步

第一步：用 MRVL 做 1 天或 30 分钟的 `mbp-10` 样本，计算 spread、mid、QI、DI、OFI、MLOFI。

第二步：画 4 张图：

- mid price；
- spread；
- top queue imbalance；
- 10 档 depth imbalance；
- OFI/MLOFI。

第三步：做事件研究：

- 当 ask depth 突然增大时，后续 mid 是否下跌？
- 当 bid depth 消失时，后续 volatility 是否上升？
- 当 OFI 极端负时，后续 1 分钟是否更容易下跌？

第四步：扩展到 NVDA、MU、AVGO、ALAB，做 cross-impact。

第五步：把结果和 portfolio 权重结合，输出一个组合级 LOB stress dashboard。

## 追加思考：新闻解释不了的科技股大跌，LOB 能不能更早看出来？

这一节是根据你新的想法补充的。你提出的问题非常重要：有些下跌并不是因为出现了显而易见的坏消息，而是因为资金层面的风险偏好变化、估值重定价、利率预期变化、拥挤交易出清，或者大资金从热门板块撤出。新闻往往只能事后给解释，但你真正想要的是更早看到“市场内部已经开始变坏”的证据。

### 最近这个例子的市场背景

以 2026-06-05 这次美股科技和半导体大跌为例，公开新闻报道并不是传统意义上的“公司爆雷”或“宏观崩坏”。AP 和 Reuters 系列报道的共同解释是：强于预期的就业数据推高了市场对美联储更鹰派、甚至再度加息的担忧，长端利率上行，进而压缩高估值科技股和 AI/半导体股票的估值。Axios 也把它概括成一种 “good news is bad news” 的市场反应：就业强本来是经济好消息，但对高估值股票而言意味着利率压力更大。

同一天，科技和芯片板块跌幅特别集中。报道提到 Nasdaq 大跌，Nvidia、Broadcom、Micron、AMD、Marvell 等芯片或 AI 相关股票承压，芯片板块市值蒸发超过 1 万亿美元量级。这类下跌很符合你的观察：新闻表面上不是“坏消息”，但市场价格已经发生了非常剧烈的重新定价。

这说明只看 news、OHLCV、历史成交量，确实可能太慢。原因是：

- news 解释是离散的，订单流是连续的；
- news 告诉你“可能的原因”，LOB 告诉你“现在谁在承受压力”；
- OHLCV 看到的是结果，LOB 看到的是结果形成前的流动性和订单流变化；
- 大资金撤退不一定会提前写在新闻里，但可能先表现为撤单、买盘变薄、卖盘持续补量、跨股票同步 OFI 恶化。

### 你的核心假设

你的想法可以整理成一个研究假设：

**当科技股或半导体股出现无明显坏消息的大跌时，新闻和历史 OHLCV 可能无法提前解释，但实时 LOB 可能提前显示资金撤退、流动性变薄和卖压扩散。**

更具体地说：

1. 大资金不是通过新闻表达意图，而是通过订单行为表达意图。
2. 大资金撤退不一定表现为单笔巨大卖单，因为它可能拆单、撤掉买盘、减少提供流动性，或通过 ETF/篮子交易扩散。
3. 科技股大跌前，可能先看到 LOB 中的“承接能力下降”，而不是立刻看到成交量暴增。
4. 对 portfolio 来说，最危险的不是一只股票出现卖压，而是高权重股票和相关股票同时出现 LOB stress。

### LOB 能比新闻多看到什么？

LOB 可以补充新闻看不到的四类信息。

第一，**买盘是否撤退**。

如果大资金不愿意接盘，不一定表现为主动卖出，也可能表现为 bid side depth 下降：

\[
BidDepth_t = \sum_{i=0}^{9} bid\_sz_i
\]

如果价格还没明显跌，但 \(BidDepth_t\) 已经持续下降，说明承接盘在撤。

第二，**卖盘是否持续补充**。

单个大 ask 不一定重要，因为可能撤掉。但如果 ask side 多档持续补量，说明上方供应压力持续存在：

\[
AskDepth_t = \sum_{i=0}^{9} ask\_sz_i
\]

真正危险的是：

\[
\Delta AskDepth_t > 0
\]

同时：

\[
\Delta BidDepth_t < 0
\]

第三，**订单流是否从局部变成系统性**。

用 OFI 或 MLOFI 看：

\[
OFI_t < 0
\]

说明 top-of-book 净订单流偏卖压。

如果 10 档 MLOFI 多数为负：

\[
MLOFI^1_t, MLOFI^2_t, \ldots, MLOFI^{10}_t < 0
\]

说明不是 top level 噪声，而是整本近端订单簿都在偏卖。

第四，**板块内是否扩散**。

如果 NVDA、AVGO、MU、MRVL、AMD、SOXX 同时出现负 OFI，或者龙头先出现负 OFI，其他股票随后恶化，这就是 portfolio 级别信号。

可以定义：

\[
SectorOFI_t = \sum_i w_i OFI_{i,t}
\]

其中 \(w_i\) 可以是你的 portfolio 权重，也可以是市值权重、成交额权重或风险权重。

### “大资金撤退”在 LOB 里可能长什么样？

不能直接从 LOB 看到“华尔街某机构正在撤资”。LOB 没有交易者身份，也看不到暗池、隐藏订单、内部化成交和 OTC 流动性。但是，大资金撤退可能留下以下可观测痕迹。

#### 1. Bid side depth 变薄

大资金如果原本愿意在下方接盘，现在不愿意接，前 10 档 bid depth 会下降。

可观察信号：

- `bid_sz_00` 到 `bid_sz_09` 同时下降；
- 10 档 bid depth 低于过去 30 分钟分位数；
- spread 扩大；
- mid-price 下跌时 bid depth 没有恢复。

#### 2. Ask side depth 持续补量

如果卖方持续压上来，ask depth 会在价格反弹时继续补量。

可观察信号：

- ask depth 上升但价格无法突破；
- ask side 大单出现后没有立刻撤走；
- ask side 被成交后又快速 refill；
- MLOFI 在多个 ask levels 上持续偏负。

#### 3. 成交后流动性不恢复

Bonart & Gould 的 liquidity provision 文章很适合解释这一点。market order 到达后，流动性不一定马上恢复；在信息风险高的时候，same-side limit orders 可能先撤走。

如果大跌当天出现：

- 卖出 market orders 后 bid side 不补；
- 或每次小反弹后 ask side 马上补量；
- 或 liquidity refill speed 变慢；

这说明市场韧性下降，后续更容易继续跳跌。

#### 4. Queue imbalance 极端偏负

Gould & Bonart 的 queue imbalance 文章说明，最优买卖队列不平衡对下一跳方向有预测力，尤其对 large-tick 股票更明显。

对高流动性科技股，可以观察：

\[
QI_t =
\frac{bid\_sz\_00 - ask\_sz\_00}
{bid\_sz\_00 + ask\_sz\_00}
\]

如果 \(QI_t\) 持续接近 -1，说明最优卖盘相对强、买盘相对弱，短期下跳风险更高。

#### 5. Cross-asset OFI 同步恶化

Cont, Cucuringu, Zhang 的 cross-impact OFI 文章对你的 portfolio 最关键。真正的板块撤退，不一定从单只 MRVL 开始，而可能先从 NVDA、AVGO、SOXX 或 QQQ/NDX 相关产品开始，再传到 MRVL、MU、ALAB、杠杆 ETF。

可观察信号：

\[
OFI_{NVDA,t} < 0
\]

随后：

\[
OFI_{MRVL,t+h} < 0,\quad OFI_{MU,t+h} < 0,\quad OFI_{SOXX,t+h} < 0
\]

其中 \(h\) 可能是几十秒到几分钟。

### 新闻、OHLCV 和 LOB 的层级差异

可以把信息层级理解成这样：

| 信息层级 | 能看到什么 | 看不到什么 | 对这类大跌的缺陷 |
| --- | --- | --- | --- |
| News | 事件解释、宏观叙事、公司公告 | 实时订单压力、撤单、承接盘 | 常常滞后，且会事后找原因 |
| OHLCV | 价格和成交量结果 | 下跌前的流动性结构 | 看到的是已经发生后的聚合结果 |
| Level 1 | best bid/ask 和 top size | 深层卖压和多档撤单 | 比 OHLCV 好，但仍太浅 |
| MBP-10 | 前 10 档价格和深度 | 单个订单身份和生命周期 | 能看多档压力，是第一阶段最佳选择 |
| MBO | 单订单新增、撤单、修改、成交 | 隐藏订单、暗池、真实身份 | 最接近研究大单生命周期，但成本高 |

因此你的研究路径应该是：

先用 `MBP-10` 判断是否存在实时压力，再用 `MBO` 验证大单生命周期。

### 一个可落地的“无新闻大跌”检测框架

你可以把这个问题变成一个事件检测系统。

#### Step 1：定义 LOB stress

对每只股票定义：

\[
LOBStress_{i,t}
=
z(Spread_{i,t})
- z(BidDepth_{i,t})
+ z(AskDepth_{i,t})
- z(OFI_{i,t})
- z(MLOFI_{i,t})
\]

这里 \(z(\cdot)\) 表示用过去一段时间标准化。符号含义是：

- spread 越大，stress 越高；
- bid depth 越低，stress 越高；
- ask depth 越高，stress 越高；
- OFI 越负，stress 越高；
- MLOFI 越负，stress 越高。

可以先不用复杂模型，直接做 z-score 加权。

#### Step 2：定义 portfolio LOB stress

按你的仓位权重加权：

\[
PortfolioLOBStress_t
=
\sum_i w_i LOBStress_{i,t}
\]

如果 \(w_i\) 是真实 portfolio 权重，这个指标就直接回答：你的组合现在承受的订单簿压力有多大。

#### Step 3：定义 sector diffusion

对半导体 watchlist：

\[
SemiLOBStress_t
=
\sum_i v_i LOBStress_{i,t}
\]

其中 \(v_i\) 可以先用等权，后面改成成交额权重或市值权重。

如果 `SemiLOBStress` 先升高，然后你的 portfolio 开始跌，说明板块 LOB 压力有预警价值。

#### Step 4：定义“news mismatch”

你的核心问题是：新闻没坏，但市场大跌。可以做一个 mismatch 指标：

\[
Mismatch_t =
LOBStress_t - NewsStress_t
\]

其中 `NewsStress` 可以先人工标注，或者用新闻情绪模型。若新闻偏中性或利好，但 \(LOBStress_t\) 很高，就说明市场内部资金行为和新闻叙事不一致。

这类时刻最值得研究。

### 如何验证这个想法？

你不能只凭直觉说 LOB 可以预测大跌，需要做一个小型实证。

#### 实验 1：2026-06-05 半导体大跌回放

标的：

- `NVDA`
- `AVGO`
- `MU`
- `MRVL`
- `AMD`
- `SOXX` 或相关 ETF

数据：

- Databento `XNAS.ITCH`
- schema 先用 `mbp-10`
- 时间窗口：2026-06-05 开盘前后、就业数据发布时间后、正式交易时段前 60 分钟和开盘后 60 分钟

要观察：

- 开盘前或开盘初期 bid depth 是否先下降；
- ask depth 是否持续补量；
- OFI 是否在价格大跌前先转负；
- NVDA/AVGO 的 OFI 是否领先 MRVL/MU；
- spread 是否扩大；
- liquidity refill 是否变慢。

#### 实验 2：正常交易日对照

选一个没有明显科技股大跌的正常交易日，重复同样指标。

比较：

- LOBStress 是否明显低；
- OFI 是否没有持续单边；
- sector diffusion 是否弱；
- bid depth 是否更稳定。

#### 实验 3：新闻好但股票跌的日期集合

长期可以收集这类日期：

- 就业好但科技跌；
- CPI 不差但利率上行导致科技跌；
- 公司财报不差但估值压缩；
- 龙头股轻微利空触发全板块撤退。

然后比较：

\[
P(\text{large drawdown} \mid LOBStress\ high, NewsStress\ low)
\]

是否高于：

\[
P(\text{large drawdown} \mid NewsStress\ low)
\]

如果显著更高，说明 LOB 确实补充了新闻无法解释的信息。

### 对“good news is bad news”的解释

“Good news is bad news” 本质上不是新闻本身坏，而是市场状态和定价条件变了。

强就业数据本身是经济好消息，但如果市场已经高度拥挤在 AI/科技/半导体，并且估值依赖低利率，那么强就业会带来：

- 利率上行；
- 折现率上升；
- 高估值股票估值压缩；
- 做多科技的拥挤仓位减风险；
- 被动和量化组合同步降暴露；
- ETF/期货/篮子交易传导到单股。

新闻只告诉你第一层因果：就业强、利率上行。LOB 可能看到第二层行为：买盘撤退、卖压补充、板块同步恶化。

所以 LOB 的价值不是替代宏观解释，而是把宏观解释落实到可交易的微观证据。

### 我对这个研究方向的判断

我认为你的想法有价值，而且比“直接用新闻预测股价”更现实。原因是：

1. 新闻通常是低频、离散、模糊的；
2. LOB 是高频、连续、直接反映交易行为的；
3. 科技股和半导体的下跌常常是估值、利率、拥挤交易和风险偏好共同作用，不一定有单一坏消息；
4. 这种情况下，资金行为可能比新闻文本更早反映风险；
5. 对 portfolio，LOB stress 可以按仓位加权，直接变成组合风险预警。

但也要保持边界：

- LOB 不能识别具体华尔街机构；
- LOB 不能看到暗池和隐藏流动性；
- LOB 信号有效期很短；
- 高频数据成本高；
- 交易成本会吞掉很多看似可预测的收益；
- 大跌前可能也没有稳定预警，尤其是由隔夜新闻或盘前宏观数据直接触发时。

因此，最合理的定位是：

**LOB 不是预测所有大跌的水晶球，而是帮助你识别新闻解释之外的实时资金压力、流动性撤退和组合级风险扩散。**

### 建议加入报告的下一步任务

下一步可以直接做一个小实验，不需要马上做复杂模型：

1. 下载 2026-06-05 `NVDA`、`MRVL`、`MU`、`AVGO` 的 `mbp-10`。
2. 每 1 秒聚合一次 spread、bid depth、ask depth、OFI、MLOFI。
3. 画出这些指标和 mid-price。
4. 标记就业数据发布时间、开盘时间、价格第一次明显下跌时间。
5. 看 LOBStress 是否早于价格或同步价格恶化。
6. 再拿一个正常交易日做对照。

如果这个实验显示，在价格大跌前或大跌初期，LOBStress 已经明显异常，那么你的研究方向就成立，可以继续扩展到 portfolio-level dashboard。

## 本节参考的公开市场背景资料

- AP: https://apnews.com/article/b5e10863b81cb1d6399f688ad8885c46
- AP index recap: https://apnews.com/article/b9d2661cbba6cc326c618c06769d8291
- Axios jobs/rates framing: https://www.axios.com/2026/06/08/stocks-jobs-interest-rates
- Axios chip selloff recap: https://www.axios.com/2026/06/05/stocks-nasdaq-tech-stocks
- Reuters via Investing.com, tech selloff and rates: https://www.investing.com/news/stock-market-news/hot-jobs-report-rising-rates-send-wall-streets-tech-favorites-sprawling-4729227
- Reuters via Yahoo Finance, chip selloff market value: https://finance.yahoo.com/markets/stocks/articles/chip-selloff-erases-over-1-200538289.html

## 补充：股票涨跌的原理、LOB 的位置，以及它和 Crypto 的区别

这一节专门解释一个更基础但很关键的问题：股票到底为什么涨跌？LOB 在这个过程里扮演什么角色？它和 crypto 的涨跌机制有什么不同？

### 股票价格涨跌的三层结构

股票涨跌可以分成三层：

1. 估值层：公司未来现金流、利润、增长率、利率、风险溢价决定“合理价格大概在哪里”。
2. 资金层：投资者、基金、ETF、量化策略、做市商、期权交易商、杠杆资金决定“谁在买、谁在卖、谁被迫调仓”。
3. 微观交易层：订单如何进入市场，谁吃掉流动性，谁撤单，谁补单，最后在 LOB 里形成每一笔成交和每一次报价变化。

很多人只看第一层：公司好不好、新闻好不好、财报好不好。但短期大跌经常发生在第二层和第三层。也就是说，公司没有坏消息，但资金层在降风险，微观交易层的买盘承接突然变弱，价格就会跌。

### 股票的“合理价值”和“成交价格”不是一回事

一个股票的理论价值可以粗略写成：

\[
Price \approx \sum_{t=1}^{\infty} \frac{ExpectedCashFlow_t}{(1+r+\rho)^t}
\]

其中 \(r\) 是无风险利率，\(\rho\) 是风险溢价。

如果就业数据强，市场担心利率更高，那么 \(r\) 上升。对 AI、半导体、高成长科技股来说，很多价值来自很远的未来现金流，所以折现率上升会明显压低估值。这就是 “good news is bad news” 的估值层解释。

但是，这只是解释“为什么市场可能愿意给更低估值”。真正从一个价格跌到另一个价格，需要交易发生。这个过程就在 LOB 里完成。

### LOB 是价格变化的微观发动机

股票在交易所里的价格不是连续滑动的，而是通过订单簿一步一步更新。

简化地说，订单簿里有两边：

- bid side：愿意买的人，按价格从高到低排队；
- ask side：愿意卖的人，按价格从低到高排队。

最优买价是 `bid_px_00`，最优卖价是 `ask_px_00`。

中间价是：

\[
mid_t = \frac{bid\_px\_00 + ask\_px\_00}{2}
\]

如果有人用 market sell order 去打 bid，最优买盘会被消耗。如果 bid queue 被吃掉或撤掉，新的最优买价会下移，mid-price 也会下移。

所以短期价格下跌通常不只是“有人卖”，而是：

1. 卖方主动吃掉 bid；
2. 原本的买方撤掉 bid；
3. 做市商降低报价；
4. ask side 持续补量压住反弹；
5. 相关股票和 ETF 同时出现类似压力。

这些都能在 LOB 中看到一部分痕迹。

### 股票为什么会突然无明显坏消息地下跌？

没有坏消息也能跌，常见原因包括：

#### 1. 利率和折现率变化

宏观数据本身可能是好消息，但如果它让市场相信利率会更高，高估值股票会被重新定价。科技股、AI 股、半导体股通常 duration 更长，对利率更敏感。

LOB 里的表现可能是：

- 高估值股票 bid depth 变薄；
- ask side 持续补量；
- OFI 转负；
- ETF 和龙头股先出现卖压。

#### 2. 仓位拥挤

如果很多基金都持有同一批 AI/半导体股票，当风险偏好下降时，它们可能同时减仓。新闻看起来不坏，但资金层发生了拥挤交易出清。

LOB 里的表现可能是：

- 多只相关股票同步负 OFI；
- 龙头先跌，二线股票滞后跟跌；
- 反弹时 ask depth 快速恢复；
- market sell 后 bid depth 不恢复。

#### 3. ETF、指数和篮子交易

很多资金不是一只一只股票卖，而是通过 QQQ、SOXX、SMH、行业篮子、期货、期权对冲来卖。这会导致单股之间出现联动。

LOB 里的表现可能是：

- NVDA、AVGO、AMD、MU、MRVL 同时出现卖压；
- ETF 的压力传到成分股；
- cross-asset OFI 先于某些股票价格反应。

#### 4. 期权和 dealer hedging

如果市场下跌触发期权 dealer 调整 delta hedge，dealer 可能需要卖出股票或指数期货来对冲。这会放大短期下跌。

LOB 里的表现可能是：

- 下跌过程中 sell market orders 加速；
- bid side 补不回来；
- spread 扩大；
- 波动率上升与深度下降同时出现。

#### 5. 流动性撤退

做市商和高频交易者在不确定性上升时会减少挂单或扩大 spread。这样即使卖单不大，也能造成明显价格下跌。

LOB 里的表现非常直接：

\[
Depth_t \downarrow,\quad Spread_t \uparrow,\quad \beta_t \uparrow
\]

这里 \(\beta_t\) 可以理解为价格冲击系数。同样大小的 OFI，在浅订单簿中会造成更大价格变化。

### LOB 和股票涨跌的关系：它不是原因的全部，但它是价格形成的现场

LOB 不是所有涨跌的根本原因。根本原因可能是利率、盈利、新闻、风险偏好、仓位、期权、ETF 流。但这些原因最后必须通过订单进入市场，才会形成价格。

可以把关系写成：

\[
Macro/Fundamental/Positioning\ Shock
\rightarrow
Order\ Flow
\rightarrow
LOB\ State
\rightarrow
Price\ Change
\]

新闻和基本面主要看第一段。OHLCV 看最后一段。LOB 看中间两段。

这就是它的价值：它不能告诉你所有原因，但它能更早显示“市场内部行为正在改变”。

### 股票 LOB 中最应该关注的几种信号

#### 1. 承接盘是否消失

\[
BidDepth_t = \sum_{i=0}^{9} bid\_sz_i
\]

如果价格还没明显跌，但 \(BidDepth_t\) 已经下降，说明买方撤退。

#### 2. 上方卖压是否持续

\[
AskDepth_t = \sum_{i=0}^{9} ask\_sz_i
\]

如果每次价格反弹，ask depth 都快速增加，说明上方有持续供应。

#### 3. 订单流是否转负

\[
OFI_t < 0
\]

表示最优档净订单流偏卖压。

#### 4. 深层订单簿是否同向变差

\[
MLOFI^1_t,\ldots,MLOFI^{10}_t < 0
\]

如果多档同时偏负，比单一 top-of-book 信号更危险。

#### 5. 板块是否同步恶化

\[
SectorOFI_t = \sum_i w_i OFI_{i,t}
\]

如果半导体股票一起负 OFI，说明不是单股噪声，而是组合级压力。

### 股票和 Crypto 涨跌机制的共同点

股票和 crypto 都有订单簿，都通过买卖订单形成价格。共同点包括：

- 价格短期由供需不平衡推动；
- market orders 消耗流动性；
- limit orders 提供流动性；
- 撤单会使深度变薄；
- spread 扩大表示流动性紧张；
- OFI、depth imbalance、queue imbalance 都有意义；
- 杠杆清算和流动性撤退都会放大波动。

所以在微观结构层面，很多 LOB 指标可以迁移：

\[
Spread,\quad Depth,\quad OFI,\quad MLOFI,\quad Imbalance
\]

这些对股票和 crypto 都有用。

### 股票和 Crypto 的根本区别

但股票和 crypto 的价格机制有几个很大的不同。

#### 1. 股票有公司现金流，crypto 很多没有传统现金流

股票背后是公司，有收入、利润、资产、回购、分红、财报和监管披露。长期股票价格最终要和未来现金流、利率和风险溢价相关。

很多 crypto 没有传统现金流，价格更多来自：

- 网络采用；
- token 供需；
- staking 或 fee capture；
- 叙事；
- 流动性；
- 杠杆；
- 交易所资金费率；
- 风险偏好。

因此，股票 LOB 更多是“基本面和资金层冲击的微观表达”；crypto LOB 更多时候直接就是“市场叙事、杠杆和流动性的主要现场”。

#### 2. 股票市场有交易时间，crypto 24/7

股票有开盘、收盘、盘前盘后、财报窗口、宏观数据发布时间。大量信息会在非交易时段积累，然后开盘集中反映。

Crypto 24/7 交易，信息释放更连续，但流动性也会在周末、夜间、地区时段明显变化。

这意味着：

- 股票 LOB 开盘前后特别重要；
- crypto LOB 的时段效应和跨交易所流动性更重要；
- 股票的大 gap 可能来自隔夜信息，LOB 只能在盘前或开盘后看到承接；
- crypto 更容易实时看到冲击传播。

#### 3. 股票市场更碎片化但更监管化，crypto 更跨交易所且质量不一

美股一个股票可能在 Nasdaq、NYSE、Arca、IEX、暗池、内部化系统中同时交易。你看到的 `XNAS.ITCH` 是 Nasdaq TotalView，不是全市场全部流动性。

Crypto 也碎片化，BTC/ETH 同时在 Binance、Coinbase、OKX、Bybit、Kraken 等交易。但 crypto 的跨交易所价格通过套利更快联动，同时不同交易所存在数据质量、刷量、API 延迟、风控和撮合规则差异。

所以：

- 股票 LOB 要注意单一 venue 不等于全市场；
- crypto LOB 要注意交易所之间的深度和价格差异；
- 股票有 NBBO 和监管框架；
- crypto 的 consolidated book 需要自己构建。

#### 4. 股票有 ETF、指数、期权 dealer 和机构再平衡

股票的大跌常常通过 ETF、指数、期权和组合再平衡传导。例如 QQQ、SOXX、SMH、NDX futures、single-stock options 都会影响成分股。

Crypto 也有衍生品，但机制不同，主要包括：

- perpetual futures；
- funding rate；
- liquidation cascade；
- basis trade；
- stablecoin 流动性；
- on-chain flows；
- exchange reserves。

股票里，你要重点看：

\[
ETF/Index/Options \rightarrow Single\ Stocks
\]

Crypto 里，你要重点看：

\[
Perps/Funding/Liquidations/Stablecoins \rightarrow Spot
\]

#### 5. 股票跌可能来自估值重定价，crypto 跌更多来自流动性和杠杆

股票的科技股大跌，常见机制是：

- 利率上升；
- 估值倍数压缩；
- 基金降科技 beta；
- ETF/期权链条放大；
- LOB 中买盘撤退。

Crypto 大跌，常见机制是：

- 杠杆多头爆仓；
- funding 过热后反转；
- stablecoin 流动性收缩；
- 大户转币到交易所；
- perp 价格带动 spot；
- 链上风险或交易所风险；
- LOB 深度突然消失。

两者都可能无明显坏消息地下跌，但背后的“资金结构”不同。

### LOB 在股票和 Crypto 中的信息含义不同

同样看到 bid depth 下降，在股票和 crypto 中解释不同。

在股票中，bid depth 下降可能表示：

- 做市商因为宏观风险撤单；
- 机构减少承接；
- ETF 或行业篮子卖压；
- 期权 dealer 对冲；
- 单一交易所流动性转移到其他 venue；
- 盘前信息等待开盘消化。

在 crypto 中，bid depth 下降可能表示：

- 交易所流动性撤出；
- 做市商降低库存风险；
- perp 多头清算压力；
- funding crowded long；
- stablecoin 购买力不足；
- 大户转移资产准备卖出；
- 跨交易所套利资金撤退。

所以不能把股票 LOB 和 crypto LOB 的信号一模一样解释。指标可以相似，经济含义要分别建模。

### 股票 LOB 更适合回答什么？

股票 LOB 更适合回答：

- 这只股票现在流动性好不好？
- 大资金是否正在减少承接？
- 板块内部压力是否同步？
- ETF/指数冲击是否传到单股？
- 我现在加仓或减仓会不会冲击价格？
- 高权重股票是否出现组合级压力？

它尤其适合 portfolio execution 和 intraday risk。

### Crypto LOB 更适合回答什么？

Crypto LOB 更适合回答：

- 哪个交易所先出现卖压？
- perp 和 spot 谁领先？
- funding 过热是否开始反转？
- liquidation cascade 是否正在发生？
- stablecoin 买盘是否不足？
- 跨交易所深度是否同步消失？
- whale 行为是否转化为交易所卖压？

它更适合实时监控杠杆和流动性链条。

### 对你的研究的结论

你的想法可以进一步升级成一个统一框架：

**股票价格大跌并不一定先出现在 news 里，而可能先出现在 LOB 的资金行为里。Crypto 也是类似，但 crypto 的 LOB 更多反映杠杆、交易所流动性和清算压力；股票 LOB 更多反映估值冲击、机构调仓、ETF/期权传导和交易所内流动性撤退。**

对你的 AI4Trading 项目，最实用的路径是：

1. 股票先做 Databento `mbp-10`：MRVL、NVDA、MU、AVGO、ALAB、SOXX。
2. 指标先做 spread、depth、OFI、MLOFI、LOBStress。
3. portfolio 层按真实仓位加权，做 `PortfolioLOBStress`。
4. crypto 另开一套类似指标，但额外加入 funding、open interest、liquidation、exchange flows。
5. 不要把 LOB 当作“预测新闻”的工具，而要把它当作“观察资金行为是否已经变坏”的工具。

最终，你可以形成两个并行系统：

| 市场 | 核心驱动 | LOB 主要作用 | 额外必须看的数据 |
| --- | --- | --- | --- |
| 股票 | 现金流、利率、风险溢价、机构调仓、ETF/期权 | 观察承接、卖压、流动性撤退、板块扩散 | 新闻、财报、利率、ETF、期权、portfolio 权重 |
| Crypto | 流动性、杠杆、叙事、资金费率、链上流 | 观察交易所深度、卖压、清算前兆、跨所传导 | Funding、open interest、liquidations、stablecoin、on-chain flows |

这也是为什么 LOB 对你的研究有价值：它不是替代新闻和历史数据，而是补上“资金正在怎样交易”这一层。
