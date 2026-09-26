# Databento LOB API 实测报告

日期：2026-06-09

这篇笔记记录本机已经跑通的 Databento Limit Order Book 测试。重点不是介绍 Databento 是什么，而是确认三件事：

1. 本机能不能用 API key 调用 Databento。
2. 哪些 dataset 可以拿到 `mbp-10` 或 `mbo` 级别的订单簿。
3. 用 MRVL 做一个真实的小样本请求，看看返回数据长什么样。

## 一句话结论

已经跑通。`XNAS.ITCH` 上的 `MRVL` 可以用 `schema="mbp-10"` 拿到 10 档 Limit Order Book 数据。

本次真实请求：

| 项目 | 值 |
| --- | --- |
| 数据源 | `XNAS.ITCH` |
| 标的 | `MRVL` |
| schema | `mbp-10` |
| 时间窗口 | `2024-06-03T13:30:00Z` 到 `2024-06-03T13:30:01Z` |
| 返回结果 | 97 行，73 列 |
| 估算成本 | 约 `$0.0028` |

注意：API key 已经保存为本机环境变量 `DATABENTO_API_KEY`，不要把完整 key 写进代码、Markdown 或 GitHub。

## 截图参考

Databento 的 Data catalog 可以按市场、dataset、schema 筛选数据。

![Databento Data catalog](/Users/yitwah/Documents/ai4trading/assets/databento/databento-catalog-2026-06-08-010917.png)

MRVL 的样本页能看到 Nasdaq TotalView-ITCH 和订单簿相关 schema。

![MRVL order book sample](/Users/yitwah/Documents/ai4trading/assets/databento/mrvl-mbo-sample-2026-06-08-010858.png)

## Dataset 支持情况

我用 Databento Python 客户端检查了几个常见 dataset：

| Dataset | `mbp-10` | `mbo` | 说明 |
| --- | --- | --- | --- |
| `XNAS.ITCH` | 支持 | 支持 | Nasdaq TotalView-ITCH，美股 Nasdaq 订单簿研究重点 |
| `ARCX.PILLAR` | 支持 | 支持 | NYSE Arca Pillar |
| `XNYS.PILLAR` | 支持 | 支持 | NYSE Pillar |
| `GLBX.MDP3` | 支持 | 支持 | CME Globex 期货 |
| `IEXG.TOPS` | 不支持 | 不支持 | 只支持 `mbp-1`、`tbbo`、`trades` 等较浅数据 |

对你现在的研究来说，半导体股票可以先从 `XNAS.ITCH` 开始，例如 `MRVL`、`NVDA`、`MU`、`AVGO`。如果之后要研究跨交易所流动性，可以再加入 `ARCX.PILLAR`、`XNYS.PILLAR` 等。

## `mbp-10` 返回什么

`mbp-10` 是 Market by Price 的 10 档聚合订单簿。它不是单个订单级别，而是每个价格档位的聚合结果。

核心字段可以分成四组：

| 字段类型 | 例子 | 含义 |
| --- | --- | --- |
| 时间 | `ts_recv`, `ts_event` | Databento 收到时间、交易所事件时间 |
| 事件 | `action`, `side`, `depth`, `price`, `size` | 本次订单簿更新是什么动作、在哪一侧、影响哪一档 |
| 最优档 | `bid_px_00`, `ask_px_00`, `bid_sz_00`, `ask_sz_00` | 最优买价、最优卖价和对应挂单量 |
| 第 10 档 | `bid_px_09`, `ask_px_09`, `bid_sz_09`, `ask_sz_09` | 第 10 档买卖盘深度 |

本次实测返回了从 `bid_px_00` 到 `bid_px_09`、从 `ask_px_00` 到 `ask_px_09` 的完整 10 档价格，也返回了每档的 size 和 count。

示例行长这样：

```text
ts_recv                              action side price  size bid_px_00 ask_px_00 bid_sz_00 ask_sz_00 bid_px_09 ask_px_09 symbol
2024-06-03 13:30:00.005967322+00:00 A      A    70.27  100  70.19     70.27     11214     100       69.68     71.28     MRVL
2024-06-03 13:30:00.006274404+00:00 A      A    70.27  100  70.19     70.27     11214     200       69.68     71.28     MRVL
2024-06-03 13:30:00.006730358+00:00 T      A    70.19  1785 70.19     70.27     11214     200       69.68     71.28     MRVL
```

读法：

- `bid_px_00=70.19`，`ask_px_00=70.27`，所以当时最优买卖价差是 `$0.08`。
- `bid_sz_00=11214` 表示最优买价这一档聚合挂单量是 11214。
- `ask_sz_00=100` 到下一行变成 200，说明最优卖价这一档的挂单量增加了。
- `action=A` 表示新增，`action=C` 表示撤单或取消，`action=T` 表示成交事件。

## 和 `mbo` 的区别

| Schema | 粒度 | 适合做什么 |
| --- | --- | --- |
| `mbp-10` | 价格档位聚合，买卖各 10 档 | depth、spread、imbalance、短窗口价格压力 |
| `mbo` | 单个订单级别 | 追踪某个订单的新增、修改、撤销、成交生命周期 |

如果你的问题是“当前买卖盘前 10 档压力怎么样”，先用 `mbp-10`。

如果你的问题是“某个大单是不是挂出来后又撤掉了”，要用 `mbo`。

## 本机怎么调用

先确认环境变量：

```bash
echo $DATABENTO_API_KEY
```

不要把输出粘贴到公开位置。只要看到以 `db-` 开头，就说明 shell 已经读到了 key。

Python 示例：

```python
import os
import databento as db

client = db.Historical(os.environ["DATABENTO_API_KEY"])

data = client.timeseries.get_range(
    dataset="XNAS.ITCH",
    symbols="MRVL",
    schema="mbp-10",
    start="2024-06-03T13:30:00Z",
    end="2024-06-03T13:30:01Z",
)

df = data.to_df()
print(df.head())
print(df.columns)
```

成本估算建议先跑：

```python
estimate = client.metadata.get_cost(
    dataset="XNAS.ITCH",
    symbols="MRVL",
    schema="mbp-10",
    start="2024-06-03T13:30:00Z",
    end="2024-06-03T13:30:01Z",
)
print(estimate)
```

技巧：先用 `get_cost` 和 `get_record_count` 估算，不要一上来拉全天、多 ticker、多 schema。

## 本次成本估算

| 请求 | 记录数估算 | 成本估算 |
| --- | ---: | ---: |
| `XNAS.ITCH` / `NVDA` / `mbp-10` / 1 秒 | 60,791 | `$0.0083` |
| `XNAS.ITCH` / `MRVL` / `mbp-10` / 1 秒 | 20,393 | `$0.0028` |
| `GLBX.MDP3` / `ESM4` / `mbp-10` / 1 秒 | 289,524 | `$0.0496` |

这些数字说明：高流动性标的和期货合约的数据量会非常大。研究时应该先做小窗口，确认字段、指标、逻辑都对，再扩大范围。

## 可以马上加工的指标

用 `mbp-10` 可以先做这些基础特征：

| 指标 | 含义 |
| --- | --- |
| spread | `ask_px_00 - bid_px_00` |
| mid price | `(ask_px_00 + bid_px_00) / 2` |
| top depth imbalance | `(bid_sz_00 - ask_sz_00) / (bid_sz_00 + ask_sz_00)` |
| 10 档 bid depth | `bid_sz_00 + ... + bid_sz_09` |
| 10 档 ask depth | `ask_sz_00 + ... + ask_sz_09` |
| 10 档 depth imbalance | `(sum_bid_size - sum_ask_size) / (sum_bid_size + sum_ask_size)` |

如果之后写论文式研究，可以把核心定义写成：

\[
mid_t = \frac{bestBid_t + bestAsk_t}{2}
\]

\[
Imbalance_t = \frac{BidDepth_t - AskDepth_t}{BidDepth_t + AskDepth_t}
\]

## 建议下一步

1. 先写一个小脚本，把 `MRVL` 的 1 分钟 `mbp-10` 保存成 Parquet 或 CSV。
2. 画出 spread、mid price、10 档 bid/ask depth 的时间序列。
3. 找出 `ask_sz_00` 或 10 档 ask depth 突然变大的时刻。
4. 看这些时刻之后 1 秒、5 秒、30 秒的 mid price 是否下跌。
5. 如果需要追踪“大单是否撤掉”，再切换到 `schema="mbo"`。

这条路线比直接下载全天 L3 数据更稳：先用 `mbp-10` 做低成本验证，再决定是否进入更贵、更细的 `mbo`。

## 参考链接

- Databento MBP-10 schema: https://databento.com/docs/schemas-and-data-formats/mbp-10
- Databento quickstart: https://databento.com/docs/quickstart
- Databento historical time series API: https://databento.com/docs/api-reference-historical/timeseries/timeseries-get-range
- Databento data catalog: https://databento.com/portal/browse
