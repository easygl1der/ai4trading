# Limit Order Book 阅读资料与本地来源索引

日期：2026-06-09

目标：筛选真正有价值的 Limit Order Book 资料，并保存到本地，方便后续研究 MRVL、NVDA 等股票的订单簿、盘口压力、撤单和价格冲击。

我按两条线筛选：

1. **怎么获得 LOB 数据**：Databento、Nasdaq TotalView-ITCH、NYSE OpenBook/Integrated Feed、LOBSTER 等。
2. **拿到 LOB 后怎么研究**：OFI、queue imbalance、多档订单流失衡、深度学习预测、流动性补给。

## 本地目录

| 目录 | 内容 |
| --- | --- |
| `notes/lob-research/sources/papers/` | 已下载 PDF 论文 |
| `notes/lob-research/sources/text/` | 从 PDF 前 1-2 页抽取的文本，用于快速确认内容 |
| `notes/lob-research/sources/webpages/` | 可保存的官方网页快照，部分现代前端页面只保存到壳，不作为主要证据 |

## 最推荐先读的 5 篇

### 1. Gould et al. - Limit Order Books

本地文件：[gould-et-al-limit-order-books-survey-2013.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/gould-et-al-limit-order-books-survey-2013.pdf)

价值：这是 LOB 综述，适合从零理解订单簿是什么、为什么 LOB 数据是市场微观结构研究的核心。它讨论 LOB 机制、实证事实、模型和研究局限。

适合你用来回答：

- LOB 到底记录了什么？
- 为什么它比 K 线更接近市场真实压力？
- 研究 LOB 时有哪些常见陷阱？

### 2. Cont, Kukanov, Stoikov - The Price Impact of Order Book Events

本地文件：[cont-kukanov-stoikov-price-impact-order-book-events-2010.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/cont-kukanov-stoikov-price-impact-order-book-events-2010.pdf)

价值：这是 OFI 方向的经典短文。它把价格变化和 order flow imbalance 连接起来，适合你研究“大单、撤单、买卖盘变化是否真的推动价格”。

这篇最适合转成 MRVL/Databento 的第一版实验：

\[
OFI_t = \Delta BidSize_t - \Delta AskSize_t
\]

然后看短窗口内：

\[
\Delta mid_t
\]

是否跟 OFI 有稳定关系。

### 3. Gould & Bonart - Queue Imbalance as a One-Tick-Ahead Price Predictor

本地文件：[gould-bonart-queue-imbalance-one-tick-ahead-2015.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/gould-bonart-queue-imbalance-one-tick-ahead-2015.pdf)

价值：这篇专门研究 bid/ask queue imbalance 是否能预测下一次 mid-price 方向。它非常贴近你的“盘口能不能帮我判断短线压力”的问题。

可以直接映射到 `mbp-10` 的 top-of-book：

\[
QI_t = \frac{bid\_sz\_00 - ask\_sz\_00}{bid\_sz\_00 + ask\_sz\_00}
\]

如果 \(QI_t\) 很高，说明最优买盘队列相对强；如果很低，说明最优卖盘队列相对强。

### 4. Xu, Gould, Howison - Multi-Level Order-Flow Imbalance in a Limit Order Book

本地文件：[xu-gould-howison-multi-level-ofi-2019.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/xu-gould-howison-multi-level-ofi-2019.pdf)

价值：这篇把 OFI 从最优买卖价扩展到多档订单簿。它非常适合 Databento `mbp-10`，因为 `mbp-10` 正好有买卖各 10 档。

你可以先做 10 档 depth imbalance，再进一步做多档 OFI：

\[
DepthImbalance_t = \frac{\sum_{i=0}^{9} bid\_sz_i - \sum_{i=0}^{9} ask\_sz_i}{\sum_{i=0}^{9} bid\_sz_i + \sum_{i=0}^{9} ask\_sz_i}
\]

### 5. Zhang, Zohren, Roberts - DeepLOB

本地文件：[zhang-zohren-roberts-deeplob-2018.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/zhang-zohren-roberts-deeplob-2018.pdf)

价值：这是用深度学习处理 LOB 数据的经典方向。它不适合作为第一步，因为你现在更需要理解数据和特征；但当你有较长时间序列后，它可以作为建模方向参考。

适合以后做：

- 用多档 bid/ask price 与 size 作为输入矩阵。
- 预测短窗口 mid-price 上涨、下跌、持平。
- 比较传统特征模型和深度模型。

## 其他值得保存的论文

### Cont, Cucuringu, Zhang - Cross-Impact of Order Flow Imbalance

本地文件：[cont-cucuringu-zhang-cross-impact-ofi-2023.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/cont-cucuringu-zhang-cross-impact-ofi-2023.pdf)

价值：研究 OFI 的跨资产影响。等你不只研究 MRVL，而是同时看 NVDA、MU、AVGO、SOXX 等半导体相关标的时，这篇会有用。

### Bonart & Gould - Latency and Liquidity Provision in a Limit Order Book

本地文件：[bonart-gould-latency-liquidity-provision-2016.pdf](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/papers/bonart-gould-latency-liquidity-provision-2016.pdf)

价值：研究市场订单到达后，limit order liquidity 如何补充或撤退。它适合你后面研究“大额成交之后盘口是否恢复”的问题。

## 怎么获得 LOB 数据

### 1. Databento

适合当前项目，因为本机已经跑通 API key，并且 `XNAS.ITCH` 可以拿 `MRVL` 的 `mbp-10` 和 `mbo`。

官方入口：

- Databento MBP-10 schema: https://databento.com/docs/schemas-and-data-formats/mbp-10
- Databento MBO schema: https://databento.com/docs/schemas-and-data-formats/mbo
- Databento reconstruct order book example: https://databento.com/docs/examples/order-book/reconstructing-the-order-book
- Databento historical get_range API: https://databento.com/docs/api-reference-historical/timeseries/timeseries-get-range

本地保存的 HTML 快照在：

- [databento-mbp-10-schema.html](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/webpages/databento-mbp-10-schema.html)
- [databento-mbo-schema.html](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/webpages/databento-mbo-schema.html)
- [databento-reconstruct-order-book.html](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/webpages/databento-reconstruct-order-book.html)
- [databento-get-range-api.html](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/webpages/databento-get-range-api.html)

说明：这些本地 HTML 是现代文档站的前端快照，可能不是完整正文；正式使用时以在线官方文档为准。

### 2. Nasdaq TotalView-ITCH

用途：美股 Nasdaq 订单簿的重要源头，Databento 的 `XNAS.ITCH` 就属于这个方向。适合 MRVL、NVDA、MU 等 Nasdaq 交易股票。

官方入口：

- Nasdaq TotalView: https://www.nasdaqtrader.com/Trader.aspx?id=TotalView2

本地保存：

- [nasdaq-totalview.html](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/webpages/nasdaq-totalview.html)

说明：这个本地文件抓到的正文有限，更适合作为入口链接，不作为主要可读资料。

### 3. NYSE Historical Data Products

用途：NYSE 官方历史数据产品里有 TAQ、OpenBook、Integrated Feed 等。它说明了订单级和历史 market data 产品的官方获取路径。

本地文件：[nyse-historical-data-products.html](/Users/yitwah/Documents/ai4trading/notes/lob-research/sources/webpages/nyse-historical-data-products.html)

我检查到页面里包含 `TAQ NYSE OpenBook`、`TAQ NYSE Integrated Feed` 等内容。其中 Integrated Feed 说明是 order-by-order view，适合理解 NYSE 侧的官方数据路径。

### 4. LOBSTER

用途：学术界常用的 Nasdaq LOB 数据重建和下载服务。适合做研究和复现实验，但通常不是免费全量数据。

官方入口：

- LOBSTER: https://lobsterdata.com/

说明：LOBSTER 页面是现代前端，本地直接下载只得到网页壳，所以我没有把它当作已保存的可读正文。但它仍然是重要数据源入口。

## 当前项目的实用路线

### 路线 A：先做 Databento `mbp-10`

理由：已经验证能拿到 MRVL 数据，成本低，字段直接覆盖前 10 档 bid/ask。

第一批指标：

- `spread = ask_px_00 - bid_px_00`
- `mid = (ask_px_00 + bid_px_00) / 2`
- top queue imbalance
- 10 档 depth imbalance
- 大额 ask/bid depth 突增事件

### 路线 B：再做 Databento `mbo`

理由：当你要问“这个大单是不是撤掉了”“是不是 spoofing-like vanish pattern”，`mbp-10` 不够，要看单订单生命周期。

第一批问题：

- 新增大单是否很快撤单？
- 大单挂出后是否真实成交？
- 撤单是否发生在价格移动前？

### 路线 C：多标的扩展

当 MRVL 流程跑通后，再扩展到：

- `NVDA`
- `MU`
- `AVGO`
- `ALAB`
- `SOXX` 或相关 ETF

这时可以参考 cross-impact OFI 那篇，研究半导体板块内部的盘口压力传导。

## 已剔除或降级的材料

- 普通博客文章：很多只解释“order book 是买卖挂单表”，价值不如综述论文和官方文档。
- 只讲 crypto 交易所 API 的文章：对股票 LOB 获取路径帮助有限，但以后做 crypto 高频可以单独整理。
- 无法保存正文的前端网页：保留官方入口链接，但不当作已读正文证据。

