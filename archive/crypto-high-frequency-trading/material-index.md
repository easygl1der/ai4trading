# Crypto 高频交易资料索引

日期：2026-06-08

## 总览与学习地图

| 文件 | 说明 |
| --- | --- |
| `outputs/crypto_learning_resources_2026.docx` | Crypto 学习资源总览 |
| `outputs/crypto_learning_resources_2026_with_chain_diagram.docx` | 带链关系图的学习资源总览，建议作为主入口 |
| `outputs/chain_relationships_explained.excalidraw` | 链、资产、交易所、钱包等关系图源文件 |
| `outputs/chain_relationships_explained.share.json` | Excalidraw 分享数据 |

## Crypto 基础课程与市场结构

| 文件 | 说明 |
| --- | --- |
| `outputs/crypto_course_simplilearn_notes.pdf` | Crypto 基础课程 PDF |
| `outputs/crypto_course_simplilearn_notes.tex` | 对应 LaTeX 源文件 |
| `outputs/crypto_course_simplilearn_notes_v2.pdf` | 第二版课程 PDF |
| `outputs/crypto_course_simplilearn_notes_v2.tex` | 第二版 LaTeX 源文件 |
| `outputs/kraken_introduction_to_crypto_trading_summary_zh.docx` | Kraken 的 crypto trading 入门总结 |
| `outputs/cmc_ranking_market_pair_cryptoasset_summary_zh.docx` | CoinMarketCap 排名、市场对、cryptoasset 概念总结 |
| `outputs/cmc_how_can_i_buy_coins_tokens_summary_zh.docx` | 如何买 coins/tokens 的基础流程总结 |

## 交易机器人与自动化

| 文件 | 说明 |
| --- | --- |
| `outputs/crypto_automation_tools_guide_2026.docx` | Crypto 自动化工具指南 |
| `outputs/crypto_bot_frameworks_and_mcp_safety_guide_zh.docx` | 交易机器人框架、MCP/AI agent 与安全边界 |
| `outputs/youtube_v6_openclaw_self_healing_trading_bot_notes_zh.docx` | OpenClaw 自修复 trading bot 视频笔记 |
| `outputs/youtube_v7_claude_code_trading_bot_notes_zh.docx` | Claude Code trading bot 视频笔记 |
| `outputs/youtube_v8_crypto_trading_bot_masterclass_notes_zh.docx` | Crypto trading bot masterclass 视频笔记 |
| `outputs/crypto_algorithmic_trading_mastery_summary_zh.docx` | Crypto algorithmic trading 总结 |

## 高频交易论文与精读材料

| 文件 | 说明 |
| --- | --- |
| `papers/how-and-when-are-high-frequency-stock-returns-predictable/README.md` | Ait-Sahalia, Fan, Xue, Zhu (2025) 高频收益可预测性论文包入口；含重命名 PDF、MinerU Markdown 和中文精读导读 |
| `papers/how-and-when-are-high-frequency-stock-returns-predictable/notes/reading-guide.md` | 这篇论文和高频回测、短线量化、订单簿特征、latency ablation 的实现映射 |

## 渲染预览与可再生成材料

`work/` 目录保留了生成脚本和渲染结果。里面的 `rendered_*` 子目录包含对应文档的 PDF 和页面 PNG，可以快速预览文档排版。

重要脚本：

| 脚本 | 用途 |
| --- | --- |
| `work/build_crypto_doc.py` | 生成 crypto 学习资源文档 |
| `work/build_crypto_automation_tools_doc.py` | 生成自动化工具指南 |
| `work/build_crypto_bot_frameworks_safety_guide.py` | 生成交易机器人框架与 MCP 安全指南 |
| `work/build_crypto_algorithmic_trading_mastery_summary.py` | 生成算法交易总结 |
| `work/create_chain_excalidraw.py` | 生成链关系 Excalidraw 图 |

## 对 finance / AI4Trading 的研究价值

这组资料最有价值的不是“知道有哪些币”，而是把 crypto 市场拆成几层：

- 市场层：交易所、交易对、流动性、盘口、成交。
- 工具层：CCXT、Hummingbot、Freqtrade、Jesse、NautilusTrader。
- 数据层：订单簿、成交流、K 线、链上数据、资金费率。
- 策略层：做市、套利、趋势、均值回归、事件驱动。
- 安全层：API key 权限、只读模式、撤单权限、限额、沙盒、纸面交易。
- Agent 层：AI 可以辅助研究、生成代码、做风控检查，但实盘下单必须有硬限制。

和 Databento / LOB 研究连接时，优先考虑这些问题：

1. Crypto order book 是否能复现美股 LOB 中的 spread、depth、imbalance、large order impact。
2. 24/7 市场是否让短窗口事件研究更容易收集样本，但也更容易遇到 regime shift。
3. 交易机器人框架是否适合先做 paper trading，而不是直接接真实 API key。
4. 高频策略的核心不只是预测方向，还包括数据延迟、撮合规则、手续费、滑点、撤单失败和风控。
