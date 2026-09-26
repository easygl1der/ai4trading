# 旧对话记忆：Crypto 资料生成过程

日期：2026-06-08

这个目录保存了和 `/Users/yitwah/Documents/Codex/2026-06-07/video-youtube-video-open-document-recently` 相关的旧 Codex 会话记录。原始记录在：

`raw-sessions/`

这些 `.jsonl` 文件不是普通文章，而是 Codex 的原始会话事件记录。它们的价值是保留当时的上下文、问题推进方式、生成文件路径和最终回答。如果以后要恢复“当时为什么要生成这些材料”，可以从这里查。

## 已保存的会话记录

| 文件 | 大致内容 |
| --- | --- |
| `rollout-2026-06-07T10-17-26-019e9fde-e9d2-7830-8c63-b92f0c97affa.jsonl` | 最大的一段相关记录，包含 crypto bot frameworks、MCP safety、交易机器人安全边界等内容 |
| `rollout-2026-06-07T14-41-54-019ea0d1-0b37-7030-adce-96001c7b6d62.jsonl` | 和 crypto 学习资源、生成物路径匹配的旧记录之一 |
| `rollout-2026-06-07T16-25-05-019ea12f-7f5b-7de3-91a4-b6904ed5957a.jsonl` | 关联 crypto learning resources / trading bot 文档生成 |
| `rollout-2026-06-07T16-25-33-019ea12f-edea-7f83-9a36-89c76c350c3e.jsonl` | 关联 crypto learning resources / trading bot 文档生成 |
| `rollout-2026-06-07T16-52-12-019ea148-5516-7612-9756-a22529f42b1a.jsonl` | 关联 crypto learning resources / trading bot 文档生成 |
| `rollout-2026-06-07T16-52-47-019ea148-dd75-7a43-9281-1381cde93304.jsonl` | 关联 crypto learning resources / trading bot 文档生成 |
| `rollout-2026-06-07T16-53-21-019ea149-6142-7733-a185-2089eccd18a9.jsonl` | 关联 crypto learning resources / trading bot 文档生成 |

## 旧聊天里沉淀下来的核心想法

1. Crypto 不是单一资产类别，而是一套市场结构：链、钱包、交易所、交易对、稳定币、现货、永续合约、资金费率和 API 执行。

2. 高频/自动化交易的第一层不是 AI，而是数据、盘口、撮合、手续费、滑点、延迟和风控。AI agent 可以帮忙理解、写代码、检查策略，但不能跳过交易系统的基础安全设计。

3. 交易机器人框架要分层看：CCXT 更像交易所 API 抽象层，Freqtrade/Jesse 更接近策略和回测框架，Hummingbot 更偏做市/套利执行，NautilusTrader 更偏专业级事件驱动回测和实盘架构。

4. MCP/AI agent 接交易能力时必须分权限：只读行情和账户信息是第一阶段；paper trading 是第二阶段；真实下单必须加入限额、白名单、人工确认、撤单机制和日志审计。

5. 这条 crypto 线应该和当前项目的 Databento / LOB 线合并理解：美股订单簿研究提供严谨的 market microstructure 框架，crypto 提供 24/7、API 更开放、交易机器人生态更成熟的实验场。

## 以后怎么继续

下一步最自然的方向不是马上实盘，而是把这组资料转成一个研究路线：

1. 用当前项目的 Databento 笔记定义订单簿指标：spread、depth、imbalance、large order impact。
2. 在 crypto 交易所 API 上找相同字段：order book depth、trades、funding、open interest。
3. 用只读数据做离线研究。
4. 用 paper trading 或 sandbox 验证策略。
5. 最后才考虑小额度、硬限制、可撤销的实盘实验。

