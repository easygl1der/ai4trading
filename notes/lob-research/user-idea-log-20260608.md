# 用户研究想法日志：Limit Order Book 与市场意图判断

日期：2026-06-08

## 原始动机

用户正在学习金融知识，希望把对市场的观察、疑惑和交易直觉持续记录下来，而不是让这些想法散落在对话里。当前核心方向是研究 Limit Order Book 是否能提供新闻之外的、更接近市场真实资金行为的信息。

## 用户当前问题

用户原话中的 “naming order book” 应理解为 Limit Order Book。用户关心的是：

- Limit Order Book 里蕴含什么额外信息？
- 这些信息是否可以帮助对市场行情做出更准确判断？
- 是否能通过 LOB 判断市场到底更想做多还是做空某只股票？
- 当新闻很好但市场或半导体/科技板块大跌时，LOB 是否能解释新闻无法解释的短期抛压？

## 用户当前假设

1. 大额资金卖出可能会提前或同时体现在 LOB 中。

   用户怀疑：如果出现华尔街级别的大额资金撤离，可能意味着有人在刻意压低某只股票，或者市场内部真实资金流向和新闻表面含义相反。

2. 大额资金买入可能意味着背后有资本在做多。

   如果 LOB 中突然出现持续性大额买入、买盘补充、买方队列增强，可能说明有大资金准备推高或支撑股价。

3. 新闻信号不够。

   用户之前想用新闻做交易：新闻好，按直觉应该利好市场。但最近遇到“新闻很好，半导体和科技整体大跌”的情况，因此意识到新闻可能无法解释短期价格行为。

4. “Good news is bad news” 需要微观结构解释。

   用户观察到：宏观或新闻层面看起来是好消息，但市场反而暴跌。用户希望知道这种情况是否可能来自机构资金、流动性、预期差、交易拥挤、去杠杆、做空压力或订单簿层面的抛压。

5. 目标不是只看单个大单，而是判断市场意图。

   用户真正想判断的是：市场当前更偏向做多还是做空；某只股票的短期下跌是暂时流动性冲击，还是大资金系统性撤离。

## 需要保留的研究原则

- 不要把单个大挂单直接解释为“华尔街要做多/做空”。大单可能是真实交易意图，也可能是撤单、分拆执行、做市库存调整、流动性提供、spoofing/layering 或普通订单簿噪声。
- 研究应从可检验指标出发：order flow imbalance、queue imbalance、depth imbalance、cancel/add ratio、trade sign、price impact、resiliency、spread、realized volatility。
- 对用户交易决策最有价值的不是“发现一个大单”，而是判断这个大单之后是否真的带来成交、价格冲击、流动性消失或订单簿恢复。

## 已纠正的数据来源认知

来源线程：`codex://threads/019ea297-372f-75e0-87fe-32b0e7a0fcbb`

用户曾询问 MRVL（Marvell Technology，Nasdaq 上市）上一个交易日的 Limit Order Book 数据在哪里能公开获取。需要保留的结论是：

- 美股单只股票的完整历史 full-depth Limit Order Book 通常不是免费、开源、公开数据；它属于交易所 proprietary market data。
- “开源”通常只能找到处理订单簿数据的代码或研究方法，不等于能免费拿到 Nasdaq/Cboe/NYSE/IEX 等全市场历史订单簿原始数据。
- MRVL 的主要官方深度来源应优先理解为 Nasdaq TotalView / Historical TotalView-ITCH；Databento、LOBSTER、BMLL、Cboe DataShop、Bloomberg、LSEG Refinitiv Tick History 等是更现实的数据获取路线，但多为付费。
- IEX HIST/DEEP 是相对接近公开可下载的历史深度数据来源，但只代表 IEX 这个 venue 的订单簿，不能代表 MRVL 的全市场盘口。
- Nasdaq BookViewer、Bookmap + dxFeed 这类工具更适合看实时或可视化盘口，不应默认等同于可以下载完整历史 LOB。
- 研究“资金是什么样子”时，应明确区分 lit venue 公开挂单、隐藏单、冰山单、暗池成交和真实资金身份。LOB 能看到公开挂单和成交/撤单行为，不能直接看到资金身份或完整意图。

以后项目里如果 AI 说“可以免费公开拿到 MRVL 完整历史订单簿”，应视为错误假设并先纠正。更合理的默认方案是：用 Databento 的 `XNAS.ITCH` / Nasdaq TotalView-ITCH 小窗口样本做验证；预算为零时，只能用 IEX venue 数据做不完整样本。
