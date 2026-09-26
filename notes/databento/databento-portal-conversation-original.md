<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" style="height:64px;margin-right:32px"/>

# 我可以用这个网站做什么事情？

可以。这个网站是 Databento 的数据门户，你主要可以用它来查找、开通、下载和通过 API 使用市场数据，适合做量化研究、回测、实时行情接入和品种筛选。[^1_1]

## 能做的事

你现在这个页面是 **Data catalog**，可以按资产类别浏览数据，包括股票、股票期权、期货、期货期权和 ETF，并且能按搜索、交易所、数据集、地区等条件筛选具体标的。[^1_1]
页面里列出了很多可直接查看的市场数据源，例如美股、OPRA 美股期权、CME、CFE、Eurex、ICE 等，所以你可以先确认它有没有你要研究的市场和合约。[^1_1]

## 量化用途

如果你做量化，这个平台最直接的用途是拿历史数据做回测和研究，因为页面显示很多数据集都有较长历史覆盖，例如美股自 2018 年起、CME 自 2010 年起、OPRA 自 2013 年起。[^1_1]
它也支持 live data，且页面说明付费计划里包含实时数据，所以你还可以把它接进自己的策略、行情监控、信号生成或执行系统。[^1_1]

## 你能操作什么

从导航栏看，你不只是“看目录”，还可以进入 Plans and live data、API keys、Download center、Data usage、Billing 和 Documentation 页面。[^1_1]
这意味着你大致可以完成这几类工作：

- 开通实时数据权限，页面提示需要填写一个简短问卷来确定 subscriber status，并自动生成符合交易所许可要求的合同。[^1_1]
- 创建和管理 API key，用程序拉取数据。[^1_1]
- 去 Download center 下载数据文件，适合离线研究和本地回测。[^1_1]
- 看 Documentation 学 API 接口和字段定义。[^1_1]
- 查看 Pricing 和 Billing，评估不同数据集的成本。[^1_1]


## 适合你的场景

结合你现在打开的是 Databento，这个平台很适合做这些事：

- 研究美股 L1/L2 或成交数据，页面写到美股覆盖 15 个交易所和 30 个 ATS，并有 20,000+ symbols。[^1_1]
- 做美股期权研究，OPRA 页面显示有 1,600,000+ symbols，包含 consolidated last sale 和 national BBO。[^1_1]
- 做期货和宏观品种研究，比如 CME 的 ES、CL、SR3、ZN，或 CFE 的 VX。[^1_1]
- 做欧洲能源、利率、股指衍生品研究，例如 EEX、Eurex、ICE Europe Commodities/Financials。[^1_1]


## 你下一步该怎么用

如果你的目标是“先验证数据是否适合策略”，建议先搜索具体品种，例如 ES、ZN、SPY、QQQ、TSLA 或某个期权/期货品种，看看它对应的数据集、历史起始时间和价格。[^1_1]
如果你的目标是“尽快接入程序”，更合适的路径是先去 API keys 和 Documentation，再决定用下载方式还是 API 方式接进你的研究框架。[^1_1]

你要不要我直接按“量化交易者视角”，帮你把 Databento 在 **历史回测、实时订阅、订单簿研究、期权数据研究** 这几种用途拆开讲？

<div align="center">⁂</div>

[^1_1]: https://databento.com/portal/browse


---

# 我可以在这个网站上找到 Limited Order Book 的信息吗？

答案非常明确。Schema 下拉菜单直接列出了所有可用的数据格式，其中与 **Limit Order Book（限价订单簿）** 直接相关的有：[^2_1]

***

## 订单簿相关 Schema

| Schema | 说明 |
| :-- | :-- |
| **MBO** | Market by Order — 逐笔委托级别，最细粒度，每条记录对应一个具体订单的增删改 |
| **MBP-1** | Market by Price, 1 档 — 最优买卖各一档（即 BBO） |
| **MBP-10** | Market by Price, 10 档 — **这就是标准的 Limit Order Book，买卖各 10 档深度快照** |
| **TBBO** | Top of Book BBO — 成交时的最优买卖价快照 |
| **BBO-1s** | 每秒聚合的最优买卖价 |
| **CBBO-1s / CBBO-1m** | Consolidated BBO，跨交易所合并的最优价 |


***

## 对你最有用的是哪几个？

**MBP-10** 是最直接的 LOB 数据，包含买卖各 10 档的价格和挂单量，适合做：

- 订单流分析（Order Flow Imbalance）
- 微观结构研究（Market Microstructure）
- 高频策略回测

**MBO** 粒度更细，记录每一个订单的生命周期（新增、修改、成交、撤销），适合做：

- 重建全深度订单簿
- 隐藏单/冰山单检测
- 做市商行为研究

***

## 如何筛选

你可以直接在这个 Schema 下拉菜单中勾选 **MBP-10** 或 **MBO**，页面会自动过滤出所有提供该类订单簿数据的数据集（美股、期货、期权等），然后按 Dataset 或 Region 进一步缩小范围。

需要我帮你筛选某个特定市场（比如 CME 期货或美股）的 LOB 数据集吗？

<div align="center">⁂</div>

[^2_1]: https://databento.com/portal/browse


---

# 你看我这个页面上面有没有“Limit order book”的这个信息？

可以看到！你现在的页面是 **Nasdaq TotalView-ITCH** 数据集下 MRVL（Marvell Technology）的详情页。[^3_1]

右侧信息栏里的 **Schemas** 字段明确列出了：

> **MBO, MBP-1, MBP-10**, BBO-1s, BBO-1m, TBBO, Trades, OHLCV-1s, OHLCV-1m, OHLCV-1h...

其中：

- **MBP-10**（Market by Price, 10 档）就是标准的 **Limit Order Book**，包含买卖各 10 档的价格和挂单量
- **MBO**（Market by Order）是更细粒度的逐笔委托数据，可以用来**重建完整 LOB**

所以答案是：**有的**，这个页面上的数据集（Nasdaq TotalView-ITCH）完整支持 Limit Order Book 数据，通过 MBP-10 或 MBO schema 均可获取。

<div align="center">⁂</div>

[^3_1]: https://databento.com/portal/catalog/us-equities/XNAS.ITCH/equities/MRVL


---

# 你可以帮我滚动去读一下这一个表格里面的数据都是什么意思吗？它可能有很多行，很多列。

好，现在我已经完整读取了 MBP-10 schema 的所有字段说明。结合你发来的截图，我来逐列解释：[^4_1]

***

## 你那张截图里的列含义（MBP-10 Schema）

这张表是 **Nasdaq TotalView-ITCH / MRVL** 的 MBP-10（10档限价订单簿）样本数据，每一行代表一次**订单簿状态更新事件**。

### 时间戳列

| 列名 | 含义 |
| :-- | :-- |
| **ts_recv** | Databento **捕获服务器收到**消息的时间戳，精度到纳秒（自 UNIX epoch 起的纳秒数） |
| **ts_event** | 交易所**撮合引擎产生**事件的时间戳，精度到纳秒 |

两者的差值（`ts_recv - ts_event`）就是网络传输延迟，通常只有几百纳秒，你在截图中可以看到差值大约在 **167–230 纳秒**左右，说明数据捕获延迟极低。

### 其他列（截图中被截断的部分）

| 列名 | 含义 |
| :-- | :-- |
| **rtype** | 记录类型（Record Type），截图中全是 `10`，对应 MBP-10 schema |
| **publisher_id** | Databento 分配的发布者 ID，标识数据集来源（Nasdaq TotalView-ITCH） |
| **instrument_id** | 标的品种的数字 ID（对应 MRVL） |
| **action** | 事件类型：**A**=新增委托、**C**=撤单、**M**=修改、**R**=清空订单簿、**T**=成交 |
| **side** | 方向：**A**=卖方（Ask）、**B**=买方（Bid）、**N**=无方向 |
| **flags** | 标志位字段，标识事件是否结束、消息特征和数据质量 |
| **sequence** | 交易所分配的消息序列号，可用于排序和去重 |

### 订单簿深度列（LOB 核心数据）

这是 MBP-10 最重要的部分，`N` 从 `00`（最优档）到 `09`（第10档）：


| 列名模板 | 含义 |
| :-- | :-- |
| **bid_px_N** | 第 N+1 档的**买方价格**（N=00 为最优买价 Best Bid） |
| **ask_px_N** | 第 N+1 档的**卖方价格**（N=00 为最优卖价 Best Ask） |
| **bid_sz_N** | 第 N+1 档买方挂单的**数量**（股数） |
| **ask_sz_N** | 第 N+1 档卖方挂单的**数量**（股数） |


***

## 总结

你看到的这张表本质上是一个**10档订单簿快照流**——每当订单簿任意一档发生变动（有人挂单、撤单、成交），就产生一行新记录，记录当时完整的10档买卖价格和挂单量。这正是做**订单流分析（Order Flow Imbalance）、微观结构研究、高频策略回测**的核心原始数据。

<div align="center">⁂</div>

[^4_1]: https://databento.com/portal/catalog/us-equities/XNAS.ITCH/equities/MRVL


---

# 那它为什么叫 MBP 呢？ 那 MBO 又是什么？BBO 又是什么东西？缩写啊什么的。

这三个缩写都是金融市场微观结构（Market Microstructure）里的标准术语，解释如下：

***

## MBP — Market by Price（按价格聚合的订单簿）

全称是 **Market By Price**，意思是把订单簿里**同一价格档位的所有挂单合并**，只告诉你"这个价位上总共挂了多少量"，不区分具体是哪个人下的单。

- **MBP-1**：只看最优的 1 档（最优买价 + 最优卖价），也叫 Best Bid \& Offer
- **MBP-10**：展示买卖各 **10 个价格档位**，这就是你刚才看到的 LOB 数据，`bid_px_00 ~ bid_px_09`、`ask_px_00 ~ ask_px_09` 以此类推

页面上你可以看到 MBP-10 下拉里有 MBP-1 和 MBP-10 两个选项，正是这个意思。[^5_1]

***

## MBO — Market by Order（按订单粒度的数据）

全称是 **Market By Order**，比 MBP 粒度**更细**。它不做任何聚合，每一行记录对应一个**具体订单**的生命周期事件：

- 新增一笔委托（Add）
- 撤销一笔委托（Cancel）
- 修改一笔委托（Modify）
- 成交（Trade）

也就是说，同一个价格档位上可能有 50 个人在排队挂单，MBP 只告诉你总量，而 **MBO 会把这 50 笔订单每一笔都单独记录**，你能看到每个订单的 order ID、价格、数量、时间，以及它后来是被撤还是成交了。

MBO 数据可以用来**完全重建订单簿**，是高频交易研究中最细粒度的数据。

***

## BBO — Best Bid and Offer（最优买卖价）

全称是 **Best Bid and Offer**，也叫 **NBBO（National Best Bid and Offer）**，就是整个订单簿里：

- **Best Bid**：所有买方挂单中**价格最高**的那一档（愿意出最高价买入的人）
- **Best Offer / Best Ask**：所有卖方挂单中**价格最低**的那一档（愿意出最低价卖出的人）

这两者之间的差距叫做 **Bid-Ask Spread（买卖价差）**，是衡量市场流动性的核心指标。

Databento 上还有几个 BBO 变体：

- **BBO-1s**：每秒聚合一次的最优买卖价快照
- **BBO-1m**：每分钟聚合一次
- **TBBO**：成交发生时刻的最优买卖价（Trade BBO）

***

## 三者关系总结

```
MBO（最细）  →  MBP（价格聚合）  →  BBO（只看最优档）
每一笔订单        按档位汇总           只看 Top 1
```

粒度从左到右越来越粗，数据量也越来越小。做高频微观结构研究用 MBO，做 LOB 深度分析用 MBP-10，只需要价差和流动性指标用 BBO 就够了。

<div align="center">⁂</div>

[^5_1]: https://databento.com/portal/catalog/us-equities/XNAS.ITCH/equities/MRVL


---

# 我的意思是说，我很好奇我能不能看到这个订单。你现在能看到买价和卖价吗？我看看看不看得到这个买卖的这个量。我现在很想补货，就是想知道这个大额的售卖对这个股市的影响，在 Limit Order Book 里面。

现在看到的是 **MBO schema 的 JSON 样本数据**，这里已经可以看到每笔订单的详细信息了！我来帮你解读。[^6_1]

***

## 你现在看到的数据：MBO 逐笔订单数据

以截图中第一条记录（第156-172行）为例：

```json
"action": "A"        ← 新增挂单（Add）
"side": "B"          ← 买方挂单（Bid）
"price": "23.940000000"  ← 买价 $23.94
"size": 100          ← 挂单量 100 股
"order_id": "547059" ← 该订单的唯一 ID
"symbol": "MRVL"     ← 标的是 MRVL
```

第二条记录（第173行起）：

```json
"action": "A"        ← 新增挂单（Add）
"side": "A"          ← 卖方挂单（Ask）
"price": "85.510000000"  ← 卖价 $85.51
"size": 100          ← 挂单量 100 股
```


***

## 关于你的问题：大额卖单对股价的影响

**可以的，MBO 数据完全支持这个分析。** 具体来说：

1. **能看到买价和卖价** — 每条记录都有 `price` 和 `side`（B=买/A=卖）
2. **能看到挂单量** — `size` 字段就是每笔订单的股数
3. **能追踪大单的生命周期** — 通过 `order_id` 可以追踪同一笔大单从挂出（A）、到修改（M）、到撤单（C）或成交（T）的全过程
4. **分析大额卖单压力** — 你可以筛选出 `side=A`（卖方）且 `size` 很大的订单，观察它们出现后订单簿的变化，比如：
    - 买方是否开始撤单（买盘变薄）
    - 价差是否扩大
    - 最终是否触发价格下跌

***

## 但有一点要注意

这个页面展示的是**样本数据**，时间只到 2023 年 9 月。要做完整的大单影响分析，你需要通过 **API 或 Download center** 下载完整历史数据，然后在本地用 Python 重建订单簿，再做统计分析。页面底部也提示了可以用 API 下载 DBN/CSV/JSON/Parquet 格式。

你想不想了解怎么用 Databento 的 API 拉取 MBO 数据来分析这个问题？

<div align="center">⁂</div>

[^6_1]: https://databento.com/portal/catalog/us-equities/XNAS.ITCH/equities/MRVL

