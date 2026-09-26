# How and When are High-Frequency Stock Returns Predictable?

日期：2026-07-09

本目录保存论文原文、MinerU Markdown 解析结果和中文精读导读。

## 论文信息

| 字段 | 内容 |
| --- | --- |
| 标题 | How and When are High-Frequency Stock Returns Predictable? |
| 作者 | Yacine Ait-Sahalia, Jianqing Fan, Lirong Xue, Xiaonan Zhu |
| 版本 | October 15, 2025 |
| 来源文件 | 原下载文件 `/Users/yitwah/Downloads/ssrn-4095405.pdf` |
| 本地重命名 | `source/ait-sahalia-fan-xue-zhu-2025-how-and-when-are-high-frequency-stock-returns-predictable_ssrn-4095405.pdf` |
| 页数 | 63 |
| 主题 | ultra-high-frequency return predictability, duration prediction, TAQ, LASSO, random forests, latency, order-flow look-ahead |

## 文件结构

| 路径 | 用途 |
| --- | --- |
| `source/ait-sahalia-fan-xue-zhu-2025-how-and-when-are-high-frequency-stock-returns-predictable_ssrn-4095405.pdf` | 重命名后的原始 PDF |
| `mineru/ait-sahalia-fan-xue-zhu-2025-how-and-when-are-high-frequency-stock-returns-predictable_ssrn-4095405.md` | MinerU 转换出的主 Markdown |
| `mineru/images/` | MinerU 提取的图表图片资源 |
| `notes/reading-guide.md` | 中文精读导读和实现映射 |

## 转换记录

- 优先尝试 VPS Smart MinerU MCP。远端服务不能直接读取本机 `/Users/...` 路径，因此先将 PDF 上传到 VPS 的 `/tmp/mineru-inputs/...`。
- Smart wrapper 分块路径提交后，上游 MinerU MCP 返回 `Server disconnected without sending a response`，未生成最终输出。
- 作为 fallback，使用本机已配置 token 的 `mineru-open-api extract --model vlm --format md` 完成整篇转换。
- 最终 Markdown、公式、HTML 表格和图片资源已经成功保存到 `mineru/`。

## 读法建议

先读 `notes/reading-guide.md`，再读 MinerU Markdown 的正文。本文不是一个完整交易系统或 PnL 回测论文，而是先量化短 horizon 的可预测性；真正转为策略回测时，需要额外处理手续费、滑点、撮合队列、延迟、撤单失败和容量约束。
