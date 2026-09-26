# Crypto 高频交易与自动化交易资料库

日期：2026-06-08

这个目录从旧对话工作区复制而来：

`/Users/yitwah/Documents/Codex/2026-06-07/video-youtube-video-open-document-recently`

它保存的是你关于 crypto、自动化交易、交易机器人、链上/交易所基础知识，以及把这些内容接到 AI4Trading/finance 理解体系里的材料。原始输出、渲染预览、生成脚本和相关旧聊天记录都保留在这里。

## 目录结构

| 路径 | 用途 |
| --- | --- |
| `outputs/` | 最终生成物，主要是 `.docx`、`.pdf`、`.tex`、Excalidraw 图 |
| `work/` | 生成这些文档的脚本、渲染版 PDF、页面截图和中间产物 |
| `chat-memory/` | 和这组 crypto 材料相关的旧 Codex 会话记录与可读摘要 |
| `material-index.md` | 按学习/研究主题整理的文件索引 |

## 建议先看

1. `outputs/crypto_learning_resources_2026_with_chain_diagram.docx`

   这是入口型资料，适合先建立 crypto 学习地图和链之间的关系。

2. `outputs/crypto_bot_frameworks_and_mcp_safety_guide_zh.docx`

   这是最接近“能不能用 AI agent/MCP 做 crypto 交易机器人”的资料。它把 Freqtrade、Jesse、Hummingbot、NautilusTrader、CCXT 的定位和风险边界整理出来。

3. `outputs/crypto_automation_tools_guide_2026.docx`

   这是工具层索引，适合从“我想自动化”转到“应该用哪些工具、每个工具负责什么”。

4. `outputs/youtube_v8_crypto_trading_bot_masterclass_notes_zh.docx`

   这是偏交易机器人实作路线的 YouTube 笔记。

5. `outputs/crypto_algorithmic_trading_mastery_summary_zh.docx`

   这是偏算法交易框架的总结，适合和订单簿、回测、执行系统连接起来读。

## 和当前 ai4trading 项目的关系

当前项目已经有 Databento / Limit Order Book 方向的笔记，重点在美股订单簿、MBO、MBP-10、大单冲击、短窗口价格变化。这个 crypto 模块可以作为另一条并行线：

- Databento/LOB 线：研究真实订单簿数据、订单流失衡、大单冲击和补货时机。
- Crypto 线：研究 24/7 市场、交易所 API、交易机器人框架、链上数据、自动执行风险。
- 交叉点：订单簿微观结构、回测、风险控制、执行系统、AI agent 是否应该触达下单权限。

一个合理的研究路线是：先用 Databento 的 LOB 数据把微观结构问题做扎实，再把同样的问题迁移到 crypto 交易所的 order book 和 trade stream 上。不要一开始就让 agent 直接下实盘单；应该先做到只读数据、离线回测、纸面交易、限额实盘。

