# 研究语料入口

给 Polybot 决策门用的外部材料。只收近一年、读过、讲得通、能对照到「代码算特征 → 概率门控 trade/skip/pause → 纸面/回放」的东西。营销号和「交易所领先结算」当 alpha 的，不进保留表。

本目录会续写。新材料追加到对应文件，再在下面改一行状态。意图以仓库 [CONTEXT-FROM-TRADING-BOT.md](https://github.com/easygl1der/ai4trading-polybot/blob/v2/polybot-v2/docs/CONTEXT-FROM-TRADING-BOT.md) 为准；本目录不改交易开关。

内部对照：

- [决策生命周期](/cursor/stores/self/docs/decision-tree-lifecycle.md)
- [Jev 清单](/cursor/stores/self/docs/jev-decision-tree-checklist.md)
- [采集路径](/cursor/stores/self/docs/collection-pipeline-v2.md)

扫描日：2026-09-25。窗口：优先 2025-09 之后。

## 分册

| 册 | 状态 | 用途 |
| --- | --- | --- |
| [论文](/cursor/stores/self/docs/research/papers.md) | 正文打分 18 篇 | 高分另加：5 分钟做市利润来自结算、到期附近波动发散 |
| [公开讨论](/cursor/stores/self/docs/research/public-discussion.md) | 保留 21 | X 与可打开的原文；近 3 个月有官方结算窗口与 taker delay |
| [YouTube](/cursor/stores/self/docs/research/youtube.md) | 保留 9 | 近 5 个月补了负结果口播；保本封面已剔除 |
| [稳健 fade](/cursor/stores/self/docs/research/stable-reversion.md) | 8 篇 | 只借回归机制；不借库存和双边报价 |
| [回归门](/cursor/stores/self/docs/research/regression-gates.md) | 8 篇 | 校准和分位数当否决，不当价格 |
| [均值回归](/cursor/stores/self/docs/research/mean-reversion.md) | 8 篇 + 负对照 | fade 错价；15 分钟现货反转扣费后不可交易 |

## 对模型构建的约束（从这轮筛出来的）

1. 五分钟结算前数秒的现货推动是暂停条件，不是信号。回放不要把「Binance 先动、Chainlink 后打印」记成 edge。
2. 公开盘口的改档方向不等于主动成交方向。用它推断 aggressor / OFI 会接近抛硬币。
3. Resolution、oracle proposal、adapter 终态、赎回不是同一个时间戳。标签不要压成一个 `resolved_at`。
4. LLM 适合做可被否决的 skip/observe 门，不适合报价格或仓位。已有实盘口播是亏钱样本，不是参数。
5. 窗口内没有一篇正式发表、且直接测量 Chainlink Data Streams 对 BTC/ETH 5m/15m 的论文。金标准仍以仓库代码和 CONTEXT 为准。
