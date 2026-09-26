# 金融 Agent 与 AI4Trading 历史探讨汇总

日期：2026-06-08

这个目录用于集中保存你以前和 Codex 讨论过的金融 agent、交易 agent、MCP、crypto 自动化、订单簿和 AI4Trading 研究内容。以后如果继续在 `/Users/yitwah/Documents/ai4trading` 里讨论金融方向，优先把新结论追加或整理到这里，而不是散落在旧线程里。

## 历史线程来源

| 线程 | 主题 | 已保留内容 |
| --- | --- | --- |
| `codex://threads/019e9c5a-2232-7d33-aaee-db1e5ab3bc59` | 金融/交易/crypto agent 与 MCP 选型调研；Google Docs 交付；TradingAgents 源码分析 | 已复制到 `imported-2026-06-06/` |
| `codex://threads/019e9cb6-06e9-7a13-9c34-ad6fa9fccbfa` | FinRobot clone、源码结构分析、项目文档 | 已复制到 `imported-2026-06-06/` |

## 已导入文档

| 文档 | 说明 |
| --- | --- |
| `imported-2026-06-06/finance-trading-agent-research-2026.md` | 金融/交易 agent、crypto/onchain agent、MCP 数据/执行层、企业级金融 AI 产品的选型调研 |
| `imported-2026-06-06/tradingagents-project-analysis.md` | TradingAgents 源码分析：LangGraph 多 agent 投研与交易建议流程 |
| `imported-2026-06-06/finrobot-project-analysis.md` | FinRobot 源码分析：AutoGen 金融 agent 框架与 equity research report pipeline |

旧 Google Docs 版本：

- 金融与交易 Agent / MCP 选型调研：`https://docs.google.com/document/d/1jxHGZ_9mrboc2ZjLQ6f3wcM1GiYyP58xSm6NYWZPDyE/edit`

## 已下载到本机的相关项目

| 项目 | 本机路径 | 当前理解 |
| --- | --- | --- |
| TradingAgents | `/Users/yitwah/Projects/TradingAgents` | 基于 LangGraph 的多 agent 投研/交易建议框架；不是 broker 执行系统 |
| FinRobot | `/Users/yitwah/Projects/FinRobot` | 包含旧 AutoGen 金融 agent 框架和较新的 equity research report 系统 |

这两个项目的代码状态是 2026-06-06 当时检查的快照，后续如果要基于它们开发，应重新拉取并确认当前 commit、依赖和 README 变化。

## 当前核心结论

1. 比较稳的金融 agent 架构不是让 LLM 直接下单，而是拆成数据层、研究层、执行层、风控层。
2. 股票/ETF/期权方向可以把 `OpenBB + TradingAgents/FinRobot + Alpaca MCP` 当作研究和 paper trading 的候选组合。
3. Crypto/onchain 方向可以关注 `Coinbase AgentKit`、交易所 API、链上数据和自研 policy engine，但钱包/转账/swap/approve 必须有硬限制和人工确认。
4. TradingAgents 和 FinRobot 更适合做研究、报告、交易建议和决策日志，不应直接当成安全的实盘自动交易系统。
5. 企业级成熟金融 AI 更偏 Bloomberg、FactSet、AlphaSense、S&P/Kensho 这类研究/数据/文档平台，而不是开源 auto-trader。

## 和当前 ai4trading 方向的连接

当前项目里还有几条已经形成的研究线：

| 方向 | 已有入口 | 作用 |
| --- | --- | --- |
| Databento / LOB | `../databento/research-interests-and-next-steps.md` | 用 Nasdaq TotalView-ITCH、MBO、MBP-10 研究大额订单、盘口失衡和短期冲击 |
| 用户交易想法日志 | `../lob-research/user-idea-log-20260608.md` | 保留你关于新闻、盘口、市场意图、大资金行为的原始问题和研究原则 |
| Crypto 高频交易 | `../../crypto-high-frequency-trading/material-index.md` | 保存 crypto 市场结构、自动化工具、交易 bot、MCP 安全边界和高频交易资料 |

组合起来看，比较自然的路线是：

1. 先用 Databento 小窗口样本验证订单簿指标，例如 spread、depth、imbalance、OFI、大额 ask/bid 后的短窗口价格冲击。
2. 用 TradingAgents/FinRobot 类型框架生成研究报告和交易计划，但不要让它直接控制资金。
3. 用 paper trading 或回测层验证策略建议，而不是凭新闻或单次 agent 输出下单。
4. 最后才考虑 Alpaca、IBKR、交易所 API、Coinbase AgentKit 等执行层，并且必须放在硬风控后面。

## 组合架构草案

```text
market data / news / filings / order book
        |
        v
feature layer: LOB features, news events, fundamentals, portfolio exposure
        |
        v
research agents: TradingAgents / FinRobot / custom notebooks
        |
        v
decision log: thesis, risk, evidence, invalidation, confidence
        |
        v
paper trading / backtest / replay
        |
        v
risk gate: position limit, loss limit, ticker allowlist, human approval
        |
        v
broker / exchange execution
```

## 以后维护规则

- 涉及金融公式较多时，优先写 `.tex` 并编译 PDF；普通索引和项目说明可以用 Markdown。
- 所有交易相关结论都要标明是“研究建议”“paper trading 方案”还是“实盘执行方案”。
- 涉及持仓、watchlist、权重、行情、项目 star 数、当前 API/MCP 能力时都要刷新，因为这些会过期。
- 不要把 LOB 中单个大单直接解释成市场真实意图；应配合成交、撤单、深度恢复、spread、短窗口收益和多日样本验证。
- 自动交易能力必须默认关闭 live trading，先做只读、回放、paper trading 和人工确认。
