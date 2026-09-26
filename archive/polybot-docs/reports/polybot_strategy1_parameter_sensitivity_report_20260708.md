# Polybot Strategy 1 参数敏感性回测报告

日期：2026-07-08
实验名：`strategy1_parameter_sensitivity_20260708`

## 结论摘要

- 本轮完整回测在 VPS 上直接只读读取 `/opt/polybot/data/polybot.db`，没有复制数据库到本地，也没有修改 live bot、Docker/systemd、halt/resume、钱包或 API。
- 回测窗口：`2026-06-23T00:00:00+08:00` 到 `2026-07-08T18:06:06+08:00` HKT；`1,762,425` 条 orderbook snapshot，`3,969` 个市场，`170` 个参数配置。
- 这次 replay 比上一轮更接近 live entry gate：新增重算了 `MIN_NET_EDGE`、`BTC_TREND_VETO_PCT`、`OBI_VETO_THRESHOLD`、`MAX_SPREAD`、`ENTRY_WINDOW_END_SEC`，并保留 top-10 order book VWAP 进出场。
- 当前 baseline：thr=0.25, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240，`253` 笔，净收益 `1,552.79`，胜率 `75.49%`，最大回撤 `60.16`，最长连亏 `4`。
- 严格安全候选：thr=0.2, edge=0.02, OBI=0.2, spread=0.04, trend=0.0005, end=240，`427` 笔，净收益 `2,342.76`，最大回撤 `66.42`，最长连亏 `5`。这是我最建议下一步优先验证的候选。
- 进攻型候选：thr=0.18, edge=0.015, OBI=0.4, spread=0.04, trend=0.0005, end=240，净收益 `3,460.68`，但最大回撤 `161.78`、最长连亏 `7` 都明显高于 baseline，不建议直接自动部署。

## 金融解释

- `MIN_NET_EDGE` 是扣除手续费后的最小安全边际。它越高，交易越少，胜率通常更高，但会错过一部分仍然赚钱的机会。
- `OBI_VETO_THRESHOLD` 是盘口压力过滤器。当前代码里买入侧 OBI 太负会 veto；阈值越大，veto 越宽松，允许更多卖压较重的盘口进入。
- `MAX_SPREAD` 是成交质量过滤器。过窄会错过机会，过宽会放进滑点风险。本轮看 `0.04` 到 `0.06` 差异不大，`0.02` 明显过严。
- `BTC_TREND_VETO_PCT` 是趋势反向保护。本轮在 baseline 附近几乎没有边际影响，说明更主要的保护来自 EV、OBI 和时间窗口。
- `ENTRY_WINDOW_END_SEC` 控制是否继续在尾部开仓。延长到 `255` 增加收益和样本，但回撤明显变大；缩短到 `225` 更保守。

## Safety Audit

- SQLite proof: `query_only=1`，write probe failed: `True`，error: `attempt to write a readonly database`。
- VPS `docker ps` before/after 均显示 `polybot Up ... (healthy)`；本轮没有改 live service。`127.0.0.1:8080/status` 在前后检查都未连接成功，所以报告不依赖 control endpoint 状态。

## 输出文件

本地目录：`analysis/backtests/strategy1_parameter_sensitivity_20260708/`

|file|size|
|---|---|
|summary.csv|116.5 KB|
|summary_enriched.csv|144.7 KB|
|strict_candidates.csv|4.1 KB|
|balanced_candidates.csv|42.3 KB|
|pareto_candidates.csv|18.5 KB|
|trades.jsonl|60.2 MB|
|reject_counts.json|68.5 KB|
|by_entry_elapsed_bucket.csv|400.8 KB|
|by_hkt_hour.csv|967.0 KB|
|by_side.csv|83.1 KB|
|by_param_family.csv|2.9 KB|
|run.log|7.9 KB|
|analysis_summary.json|112.1 KB|

## Top 10 Raw Results

|family|variant|params|trades|win_rate|net_pnl|return|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|---|---|
|threshold_x_min_edge|thr=0.18,edge=0|thr=0.18, edge=0, OBI=0.2, spread=0.04, trend=0.0005, end=240|654|72.94%|3,879.53|387.95%|355.62|6|15.12%|
|threshold_x_min_edge|thr=0.18,edge=0.005|thr=0.18, edge=0.005, OBI=0.2, spread=0.04, trend=0.0005, end=240|622|73.31%|3,568.94|356.89%|373.32|7|15.58%|
|threshold_x_obi|thr=0.18,obi=0.4|thr=0.18, edge=0.015, OBI=0.4, spread=0.04, trend=0.0005, end=240|590|74.41%|3,460.68|346.07%|161.78|7|13.16%|
|threshold_x_min_edge|thr=0.18,edge=0.01|thr=0.18, edge=0.01, OBI=0.2, spread=0.04, trend=0.0005, end=240|592|74.49%|3,420.11|342.01%|181.02|7|13.58%|
|threshold_x_obi|thr=0.18,obi=0.3|thr=0.18, edge=0.015, OBI=0.3, spread=0.04, trend=0.0005, end=240|581|74.53%|3,289.43|328.94%|152.08|7|13.26%|
|threshold_x_btc_trend|thr=0.18,trend=0.0003|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0003, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0008|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0008, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0005|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.001|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.001, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|

Raw top 主要来自较低 `threshold=0.18`。它们说明低 threshold 下仍有很强信号，但风险并不都一样：`min_edge=0` 的净收益最高，同时最大回撤也最高。

![Return vs drawdown](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/scatter_return_drawdown.png)

## One-Factor Sensitivity

### MIN_NET_EDGE

|MIN_NET_EDGE|trades|net_pnl|win_rate|max_dd|loss_streak|
|---|---|---|---|---|---|
|0|281|1,635.22|73.31%|83.38|4|
|0.005|269|1,614.39|73.98%|91.38|4|
|0.01|258|1,624.64|75.97%|62.21|4|
|0.015|253|1,552.79|75.49%|60.16|4|
|0.02|243|1,412.36|75.72%|43.77|4|
|0.03|218|1,158.72|76.61%|37.53|4|

![MIN_NET_EDGE curve](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/one_factor_min_net_edge.png)

### OBI_VETO_THRESHOLD

|OBI_VETO_THRESHOLD|trades|net_pnl|win_rate|max_dd|loss_streak|
|---|---|---|---|---|---|
|0.1|246|1,533.97|76.42%|59.42|4|
|0.15|249|1,558.82|76.31%|60.33|4|
|0.2|253|1,552.79|75.49%|60.16|4|
|0.25|255|1,550.18|74.90%|60.01|5|
|0.3|256|1,550.98|74.61%|60.07|6|
|0.4|261|1,616.91|73.95%|88.87|7|

![OBI_VETO_THRESHOLD curve](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/one_factor_obi_veto_threshold.png)

### MAX_SPREAD

|MAX_SPREAD|trades|net_pnl|win_rate|max_dd|loss_streak|
|---|---|---|---|---|---|
|0.02|223|1,239.07|74.89%|46.25|3|
|0.03|244|1,432.33|74.59%|54.32|4|
|0.04|253|1,552.79|75.49%|60.16|4|
|0.05|257|1,540.53|74.71%|59.80|4|
|0.06|259|1,558.64|74.90%|60.47|4|

![MAX_SPREAD curve](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/one_factor_max_spread.png)

### BTC_TREND_VETO_PCT

|BTC_TREND_VETO_PCT|trades|net_pnl|win_rate|max_dd|loss_streak|
|---|---|---|---|---|---|
|0|253|1,552.79|75.49%|60.16|4|
|0.0003|253|1,552.79|75.49%|60.16|4|
|0.0005|253|1,552.79|75.49%|60.16|4|
|0.0008|253|1,552.79|75.49%|60.16|4|
|0.001|253|1,552.79|75.49%|60.16|4|

![BTC_TREND_VETO_PCT curve](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/one_factor_btc_trend_veto_pct.png)

### ENTRY_WINDOW_END_SEC

|ENTRY_WINDOW_END_SEC|trades|net_pnl|win_rate|max_dd|loss_streak|
|---|---|---|---|---|---|
|210|185|1,114.72|81.08%|25.91|3|
|225|207|1,485.65|80.19%|37.83|3|
|240|253|1,552.79|75.49%|60.16|4|
|255|325|1,746.61|70.77%|137.73|5|

![ENTRY_WINDOW_END_SEC curve](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/one_factor_entry_end.png)

单参数结论：`MIN_NET_EDGE`、`MAX_SPREAD`、`ENTRY_WINDOW_END_SEC` 有清楚影响；`OBI_VETO_THRESHOLD` 在 baseline 附近是平台型影响；`BTC_TREND_VETO_PCT` 在这一轮几乎不影响结果。

## Interaction Results

|family|configs|positive|avg_return|best_return|best_net_pnl|best_trades|best_max_dd|
|---|---|---|---|---|---|---|---|
|baseline|1|1/1|155.28%|155.28%|1,552.79|253|60.16|
|interaction:entry_end_x_force_flatten|16|16/16|147.45%|174.66%|1,746.61|325|137.73|
|interaction:max_spread_x_max_entry_price|25|25/25|146.67%|156.65%|1,566.50|258|60.77|
|interaction:threshold_x_btc_trend|30|30/30|198.87%|328.50%|3,284.99|573|140.67|
|interaction:threshold_x_min_edge|36|36/36|196.03%|387.95%|3,879.53|654|355.62|
|interaction:threshold_x_obi|36|36/36|200.18%|346.07%|3,460.68|590|161.78|
|one_factor:btc_trend_veto_pct|5|5/5|155.28%|155.28%|1,552.79|253|60.16|
|one_factor:entry_end|4|4/4|147.49%|174.66%|1,746.61|325|137.73|
|one_factor:max_spread|5|5/5|146.47%|155.86%|1,558.64|259|60.47|
|one_factor:min_net_edge|6|6/6|149.97%|163.52%|1,635.22|281|83.38|
|one_factor:obi_veto_threshold|6|6/6|156.06%|161.69%|1,616.91|261|88.87|

### Threshold x MIN_NET_EDGE

低 threshold 需要配合更强的 EV 过滤。`threshold=0.20, min_edge=0.02` 是严格候选里最好的折中；`threshold=0.18, min_edge=0` 收益最高但回撤明显放大。

![threshold x min edge](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/heatmap_threshold_min_edge_net_pnl.png)

### Threshold x OBI

`threshold=0.18` 配合较宽松的 `OBI_VETO_THRESHOLD=0.30-0.40` 产生高收益，但最长连亏和回撤上升。金融含义是：低 threshold 捕捉更多 mispricing，同时放松 OBI 让更多有短期盘口压力的机会进入，这在样本内赚钱，但更依赖后续出场流动性。

![threshold x OBI](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/heatmap_threshold_obi_net_pnl.png)

### MAX_SPREAD x MAX_ENTRY_PRICE

这一组不是主要收益来源。`MAX_SPREAD=0.04-0.06` 附近都可以，`MAX_ENTRY_PRICE=0.75` 在交互里略好，但提升很小，不建议优先动。

![spread x entry cap](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/heatmap_spread_entry_net_pnl.png)

### ENTRY_END x FORCE_FLATTEN

`entry_end=255, force_flatten=5` 收益最高，但回撤明显高于 baseline。`entry_end=225` 更防守，`entry_end=240` 仍是比较稳的中间点。

![entry end x force flatten](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/heatmap_entry_end_force_net_pnl.png)

### Threshold x BTC Trend Veto

trend veto 在这轮不是主要杠杆。不同 `BTC_TREND_VETO_PCT` 的结果基本重合，说明真正起作用的是 threshold、EV 和盘口过滤。

![threshold x trend](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/heatmap_threshold_trend_net_pnl.png)

## Candidate Sets

### Strict Candidates

严格候选要求：`trades >= 300`、正收益、最大回撤不超过 baseline 的 1.2 倍、最长连亏不超过 6、top10 winner share 不超过 60%。

|family|variant|params|trades|win_rate|net_pnl|return|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|---|---|
|threshold_x_min_edge|thr=0.2,edge=0.02|thr=0.2, edge=0.02, OBI=0.2, spread=0.04, trend=0.0005, end=240|427|75.88%|2,342.76|234.28%|66.42|5|16.95%|
|threshold_x_min_edge|thr=0.2,edge=0.03|thr=0.2, edge=0.03, OBI=0.2, spread=0.04, trend=0.0005, end=240|389|76.86%|2,029.88|202.99%|56.42|4|18.03%|
|threshold_x_min_edge|thr=0.22,edge=0.02|thr=0.22, edge=0.02, OBI=0.2, spread=0.04, trend=0.0005, end=240|345|76.81%|1,945.09|194.51%|58.93|4|20.02%|
|threshold_x_min_edge|thr=0.22,edge=0.03|thr=0.22, edge=0.03, OBI=0.2, spread=0.04, trend=0.0005, end=240|314|78.66%|1,756.33|175.63%|41.89|4|20.67%|

![strict/balanced candidate pnl](../analysis/backtests/strategy1_parameter_sensitivity_20260708/charts/balanced_candidates_net_pnl.png)

### Balanced / Aggressive Candidates

Balanced 候选放宽到最大回撤不超过 baseline 的 3 倍、最长连亏不超过 7。这里可以看到更强的收益潜力，但不是我建议直接上线的第一选择。

|family|variant|params|trades|win_rate|net_pnl|return|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|---|---|
|threshold_x_obi|thr=0.18,obi=0.4|thr=0.18, edge=0.015, OBI=0.4, spread=0.04, trend=0.0005, end=240|590|74.41%|3,460.68|346.07%|161.78|7|13.16%|
|threshold_x_obi|thr=0.18,obi=0.3|thr=0.18, edge=0.015, OBI=0.3, spread=0.04, trend=0.0005, end=240|581|74.53%|3,289.43|328.94%|152.08|7|13.26%|
|threshold_x_min_edge|thr=0.18,edge=0.015|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_obi|thr=0.18,obi=0.2|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0003|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0003, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0005|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0008|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0008, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.001|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.001, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_obi|thr=0.18,obi=0.25|thr=0.18, edge=0.015, OBI=0.25, spread=0.04, trend=0.0005, end=240|576|74.65%|3,283.27|328.33%|138.29|7|13.25%|

### Pareto Candidates

|family|variant|params|trades|win_rate|net_pnl|return|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|---|---|
|threshold_x_min_edge|thr=0.18,edge=0|thr=0.18, edge=0, OBI=0.2, spread=0.04, trend=0.0005, end=240|654|72.94%|3,879.53|387.95%|355.62|6|15.12%|
|threshold_x_obi|thr=0.18,obi=0.4|thr=0.18, edge=0.015, OBI=0.4, spread=0.04, trend=0.0005, end=240|590|74.41%|3,460.68|346.07%|161.78|7|13.16%|
|threshold_x_min_edge|thr=0.18,edge=0.01|thr=0.18, edge=0.01, OBI=0.2, spread=0.04, trend=0.0005, end=240|592|74.49%|3,420.11|342.01%|181.02|7|13.58%|
|threshold_x_obi|thr=0.18,obi=0.3|thr=0.18, edge=0.015, OBI=0.3, spread=0.04, trend=0.0005, end=240|581|74.53%|3,289.43|328.94%|152.08|7|13.26%|
|threshold_x_min_edge|thr=0.18,edge=0.015|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_obi|thr=0.18,obi=0.2|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0003|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0003, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0005|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.0008|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.0008, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_btc_trend|thr=0.18,trend=0.001|thr=0.18, edge=0.015, OBI=0.2, spread=0.04, trend=0.001, end=240|573|74.87%|3,284.99|328.50%|140.67|7|13.23%|
|threshold_x_obi|thr=0.18,obi=0.25|thr=0.18, edge=0.015, OBI=0.25, spread=0.04, trend=0.0005, end=240|576|74.65%|3,283.27|328.33%|138.29|7|13.25%|

## Baseline vs Recommended Candidate

### Baseline

|metric|value|
|---|---|
|params|thr=0.25, edge=0.015, OBI=0.2, spread=0.04, trend=0.0005, end=240|
|trades|253|
|net_pnl|1,552.79|
|return|155.28%|
|win_rate|75.49%|
|max_drawdown|60.16|
|longest_loss_streak|4|
|top10_winner_share|24.83%|
|top_rejects|`after_entry_window` 149290, `empty_book` 34117, `net_ev_below_min` 5054, `obi_veto` 1722, `cooldown` 792, `before_entry_window` 527|

|side|trades|win_rate|net_pnl|avg_pnl|
|---|---|---|---|---|
|YES|130|77.69%|823.23|6.33|
|NO|123|73.17%|729.56|5.93|

|bucket|trades|win_rate|net_pnl|avg_pnl|
|---|---|---|---|---|
|030-060s|30|73.33%|110.27|3.68|
|060-090s|30|80.00%|216.42|7.21|
|090-120s|34|76.47%|220.76|6.49|
|120-150s|28|85.71%|232.58|8.31|
|150-180s|26|92.31%|224.21|8.62|
|180-210s|36|83.33%|272.55|7.57|
|210-240s|66|59.09%|251.23|3.81|
|240-270s|3|66.67%|24.77|8.26|

### Recommended Strict Candidate

|metric|value|
|---|---|
|params|thr=0.2, edge=0.02, OBI=0.2, spread=0.04, trend=0.0005, end=240|
|trades|427|
|net_pnl|2,342.76|
|return|234.28%|
|win_rate|75.88%|
|max_drawdown|66.42|
|longest_loss_streak|5|
|top10_winner_share|16.95%|
|top_rejects|`after_entry_window` 161647, `empty_book` 34044, `net_ev_below_min` 17212, `obi_veto` 7397, `btc_trend_veto` 2634, `cooldown` 1825|

|side|trades|win_rate|net_pnl|avg_pnl|
|---|---|---|---|---|
|YES|222|75.68%|1,232.76|5.55|
|NO|205|76.10%|1,110.00|5.41|

|bucket|trades|win_rate|net_pnl|avg_pnl|
|---|---|---|---|---|
|030-060s|65|64.62%|131.25|2.02|
|060-090s|50|88.00%|407.89|8.16|
|090-120s|49|77.55%|231.98|4.73|
|120-150s|47|85.11%|368.48|7.84|
|150-180s|51|80.39%|294.67|5.78|
|180-210s|69|82.61%|467.43|6.77|
|210-240s|91|64.84%|405.49|4.46|
|240-270s|5|60.00%|35.56|7.11|

## Recommendations

1. **优先验证 `threshold=0.20, MIN_NET_EDGE=0.020`。** 这是严格安全候选里收益最高的配置，交易数从 baseline 的 253 提升到 427，净收益从 1,552.79 提升到 2,342.76，最大回撤只从 60.16 升到 66.42。
2. **如果更保守，选 `threshold=0.22, MIN_NET_EDGE=0.030`。** 它净收益 1,756.33，交易数 314，最大回撤 41.89，最长连亏 4，比 baseline 风险更低但样本量仍过 300。
3. **不要优先改 `BTC_TREND_VETO_PCT`。** 本轮在 baseline 附近没有可见边际贡献。保留当前值即可。
4. **`MAX_SPREAD` 可保持 0.04。** 0.06 略高但提升太小；0.02 过严，牺牲太多交易。
5. **`entry_end=255` 暂不直接上线。** 它增加收益，但回撤和尾部风险明显升高；如果要试，应配合更严格的 holdout validation。
6. **`threshold=0.18` 的组合值得做第二轮验证，但不要直接部署。** 它显示信号潜力很强，特别是配合 OBI 或 edge，但最大回撤和最长连亏都上升。

## Limitations

- 这是 snapshot replay，不是完整交易所模拟；没有 queue priority、真实 latency、cancel failure、部分成交排队。
- 本轮仍使用已记录的 `p_theory/imbalance`，没有重新计算 volatility lookback。
- 当前数据窗口仍偏短，所有候选都需要 walk-forward 或后续 holdout 验证。
- `TAKER_FLOW_*` 没有历史落库，无法在本轮可靠重放。
- 本轮回测 endpoint 包含 2026-07-08 白天新增数据，所以不能和上一份只到 2026-07-08 00:40 HKT 的报告逐项硬比。

## Next Step

下一轮建议只围绕 strict candidates 做 neighborhood search：

```text
threshold: 0.19, 0.20, 0.21, 0.22
MIN_NET_EDGE: 0.015, 0.020, 0.025, 0.030
OBI_VETO_THRESHOLD: 0.15, 0.20, 0.25
ENTRY_WINDOW_END_SEC: 225, 240
```

这一步会比继续扩大全局 grid 更有价值，因为我们已经知道主要交互集中在 `threshold × MIN_NET_EDGE`。
