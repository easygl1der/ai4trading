# Databento 研究笔记：订单簿、大单冲击和免费 credit 使用

日期：2026-06-08

## 已保存材料

- 原始对话 Markdown：`notes/databento/databento-portal-conversation-original.md`
- Databento Data catalog 截图：`assets/databento/databento-catalog-2026-06-08-010917.png`
- MRVL / Nasdaq TotalView-ITCH 样本数据截图：`assets/databento/mrvl-mbo-sample-2026-06-08-010858.png`

## 当前确认的信息

Databento 是市场数据平台，可以通过 portal 浏览数据集，也可以通过 Download center 或 API 拉历史/实时数据。官方文档说明，新账号有 125 美元 free data credits，可用于历史数据，或抵扣首月订阅费用；credits 按 team 共享，6 个月后过期。

你当前打开的是 Databento portal 的 data catalog 和 MRVL 的 Nasdaq TotalView-ITCH 数据详情页。截图和 Markdown 里反复出现的关键 schema 是：

| Schema | 含义 | 对你的研究价值 |
| --- | --- | --- |
| MBO | Market by Order，逐笔订单级数据 | 能看到单个订单的新增、修改、撤销、成交，适合追踪大额挂单生命周期 |
| MBP-10 | Market by Price，买卖各 10 档价格聚合 | 直接看 10 档限价订单簿，适合订单簿压力、深度、价差和失衡指标 |
| MBP-1 / BBO | 最优买价和最优卖价 | 适合轻量观察 spread、top-of-book liquidity |
| TBBO | 成交时刻的 BBO | 适合把成交和当时最优买卖价关联起来，判断成交发生在 bid/ask 附近 |
| Trades | 成交记录 | 适合和订单簿变化配合，研究卖压是否真的造成成交和价格移动 |

官方 historical API 支持 `mbo`、`mbp-10`、`mbp-1`、`tbbo`、`trades`、`ohlcv-*` 等 schema，也可以用 `metadata.list_fields` 查看字段定义。历史数据 API 的核心请求是按 dataset、symbols、schema、start/end 时间段拉取数据。

## 我判断你真正感兴趣的点

你不是只想知道“Databento 是什么网站”。你真正关心的是：

1. **能不能看到真实订单簿**

   你反复问 “Limit Order Book 有没有”“能不能看到买价卖价和量”。这说明你的第一层需求是确认数据粒度够不够细，是否真的能看到 bid/ask price、bid/ask size、订单深度，而不是只看 K 线或成交价。

2. **能不能看到单个大单**

   你特别问 “我能不能看到这个订单”“大额售卖对股市的影响”。这更偏向 MBO，不只是 MBP-10。MBP-10 可以看某个价格档位的总量，MBO 才能追踪单个 `order_id` 的生命周期。

3. **大额卖单是否会压低价格**

   你的核心研究问题可以写成：当 ask side 出现异常大的挂单或卖出压力时，之后短窗口内价格、spread、买盘深度、成交方向会怎样变化？

4. **补货/加仓时机**

   你说“我现在很想补货”，说明你不是纯学术观察，而是想把订单簿数据转成交易决策辅助：如果某只股票短期被大额卖单压住，是否代表更好的买入区间，还是说明流动性正在恶化。

5. **MRVL 和半导体相关标的**

   你当前页面是 MRVL，结合你已有持仓/观察列表背景，半导体和相关杠杆 ETF 是更高优先级。MRVL、ALAB、MU、NVDA、AVGO、SOXX/SOXS、以及你持仓里的 MVLL/MSFU/MUU 都比泛泛的 SPY/QQQ 更贴近你的需求。

6. **用 125 美元 credit 低成本验证**

   你明确提到 125 美元免费 credit。这里最合理的思路不是一次性下载大范围全市场 L2/L3，而是先用少量 ticker、少量日期、少量 schema 验证研究问题是否成立。

## 最适合先做的研究问题

### 方向 A：大额卖单冲击

目标：找出 MRVL 或半导体股票中 ask side 大额挂单出现后的短期影响。

可以定义：

- 大额卖单：`side = A` 且 `size` 位于当天同标的 ask order size 的 95% 或 99% 分位以上。
- 观察窗口：事件后 1 秒、5 秒、30 秒、1 分钟。
- 结果变量：mid-price return、spread 变化、bid depth 变化、ask depth 变化、成交是否向下穿透。

核心指标：

\[
mid_t = \frac{bestBid_t + bestAsk_t}{2}
\]

\[
OFI_t = \Delta BidSize_t - \Delta AskSize_t
\]

如果大额 ask 出现后 \(mid\) 下跌、spread 扩大、bid depth 下降，就说明卖压可能有短期冲击。

### 方向 B：订单簿失衡和短线反转

目标：判断订单簿 imbalance 是否能预测短线价格变化。

\[
Imbalance_t = \frac{BidDepth_t - AskDepth_t}{BidDepth_t + AskDepth_t}
\]

可以先用 MBP-10 做，不必一开始就用 MBO。MBP-10 的数据量更可控，也更适合用免费 credit 快速试验。

### 方向 C：加仓辅助信号

目标：把订单簿指标变成“是否适合补货”的辅助判断。

例子：

- 如果价格下跌但 bid depth 没有消失，且 spread 没有明显扩大，可能是短期流动性冲击。
- 如果价格下跌同时 bid depth 快速消失、spread 扩大、卖单持续刷新，可能不是好买点。
- 如果大额 ask 频繁出现但很快撤单，要警惕 spoofing-like 行为，不能简单解读成真实卖压。

## 推荐的免费 credit 使用顺序

第一步：不要先买全市场、全日期、全 schema。

建议先做一个非常小的样本：

- 标的：`MRVL`
- 数据集：Nasdaq TotalView-ITCH / XNAS.ITCH
- schema：先 `mbp-10`，再 `mbo`
- 日期：选 1 个交易日，最好是 MRVL 有明显波动或你关心新闻的日期
- 输出格式：优先 DBN 或 Parquet；如果只是手工检查，可导出 CSV/JSON 小样本

第二步：验证字段和数据量。

- 用 `metadata.list_fields` 看 `mbo` 和 `mbp-10` 字段。
- 先拉 5 到 30 分钟窗口，估算每分钟数据大小和成本。
- 再决定是否扩大到全天、多日、多 ticker。

第三步：扩展到一组半导体/AI 相关标的。

优先级可以是：

1. MRVL
2. ALAB
3. MU
4. NVDA
5. AVGO
6. SOXX / SOXS 或其他 ETF

如果要研究你自己的持仓风险，应该优先匹配你真实持仓和 watchlist，而不是随机选标的。

## 后续项目可以怎么落地

建议在本项目里逐步形成这几个模块：

| 模块 | 作用 |
| --- | --- |
| `data/raw/databento/` | 保存原始 Databento 下载文件 |
| `notebooks/` | 快速探索字段、成本、样本数据 |
| `src/ai4trading/databento_loader.py` | 统一读取 DBN/CSV/Parquet |
| `src/ai4trading/orderbook_features.py` | 计算 spread、depth、imbalance、OFI |
| `src/ai4trading/large_order_impact.py` | 识别大额卖单并做事件研究 |
| `reports/` | 输出图表和结论 |

## 参考链接

- Databento portal browse: https://databento.com/portal/browse
- Databento usage pricing and data credits: https://databento.com/docs/faqs/usage-pricing-and-data-credits
- Databento quickstart: https://databento.com/docs/quickstart/new-user-guides
- Databento schemas and data formats: https://databento.com/docs/schemas-and-data-formats
- Historical `timeseries.get_range`: https://databento.com/docs/api-reference-historical/timeseries/timeseries-get-range

