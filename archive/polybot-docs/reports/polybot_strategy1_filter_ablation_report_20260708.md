# Polybot Strategy 1 过滤器 Ablation 回测报告

日期：2026-07-08

## 结论摘要

- 只看 imbalance 不够。在 replay 中把所有可回放的 post-imbalance entry gate 关掉后，current live-like 场景从 `3,284.99` 变成 `-999.10`，candidate 场景从 `2,342.76` 变成 `-999.02`。这里仍保留 `empty_book`、深度不足、现金不足等物理/账户约束，也不包含无法严格回放的 taker flow、fresh requote 和账户级 halt。
- `MIN_NET_EDGE` 是本轮最有用的过滤器。取消后交易数暴增、胜率塌陷到约 49%/32%-49%、最大回撤大幅上升。这个 gate 应该保留，甚至更应该围绕 `0.015-0.020` 精调。
- `ENTRY_WINDOW_END_SEC=240` 很有风险控制价值。取消尾部限制只带来很薄的收益增量，却让最大回撤增加约 `754-968`，最长连亏增加 `5-11`。
- `ENTRY_WINDOW_START_SEC=45` 在本轮样本里疑似过严。取消开始门槛后，两个场景都明显增加收益，风险增幅相对可控。建议下一轮专门扫 `0/15/30/45` 秒。
- `OBI_VETO_THRESHOLD` 和 `MAX_SPREAD` 更像是“可调的微观结构保险”，不是强保护。当前设置可能挡掉了一些盈利交易，建议调宽而不是删除。
- `BTC_TREND_VETO_PCT` 和 `FORCE_FLATTEN_REMAIN_SEC` 在这个 snapshot replay 里边际影响为零；这不等于它们永远没用，只表示在当前其他 gate 顺序和样本下没有改变成交路径。
- `taker_flow_veto` 和 `requote_slippage` 在 live decision 日志里触发不少，但 orderbook snapshot 没有足够字段做严格反事实回放；下一步应把 taker flow 和 fresh requote 字段落库。

## 方法

- 数据窗口：`2026-06-23T00:00:00+08:00` 到 `2026-07-08T18:34:37+08:00` HKT。
- 读取 `orderbook_snapshots`：`1,764,662` 条，`3,974` 个市场。
- 参数组合：`26` 组，即 current 与 candidate 两个场景各 `13` 组。
- 安全性：SQLite 以只读 URI 打开，`query_only=1`，写入 probe 被拒绝；没有修改 live bot、Docker/systemd、钱包或 API。
- 解释方式：每个 gate 都用“all filters on”作 baseline，然后只关掉一个 gate，看净收益、交易数、胜率、最大回撤、最长连亏的变化。

## 过滤器覆盖范围

|filter|status|reading|
|---|---|---|
|ENTRY_WINDOW_START_SEC|已 ablate|`45s` 在本轮疑似过严，建议单独扫 `0/15/30/45`。|
|ENTRY_WINDOW_END_SEC|已 ablate|`240s` 有明显风险控制价值，尾部收益/回撤比差。|
|FORCE_FLATTEN_REMAIN_SEC|已近似 ablate|本轮对 entry path 无边际影响，但作为退出安全阀仍应保留。|
|REENTRY_COOLDOWN_SEC|已 ablate|取消后收益增加但回撤/连亏恶化，属于风险-收益 tradeoff。|
|BTC_TREND_VETO_PCT|已 ablate|replay 中无边际影响；live 里会触发，但可能与其他 gate 重叠。|
|empty_book|只能计数|没有可交易盘口，不是可取消的风险过滤器。|
|MAX_ENTRY_PRICE|已 ablate|几乎不影响结果，保留为极端价格保险即可。|
|MAX_SPREAD|已 ablate|当前保护证据偏弱，可测试放宽到 `0.05/0.06`。|
|OBI_VETO_THRESHOLD|已 ablate|可能偏严，建议调宽，不建议直接删除。|
|TAKER_FLOW_*|未严格 replay|live 触发很多，但 snapshot 没有历史 flow ratio / buy-sell volume。|
|entry_depth_unavailable|只能计数|没有足够深度时无法填单，不是普通参数 gate。|
|requote_* / slippage|未严格 replay|依赖 fresh book，当前 snapshot 只有单次 book。|
|MIN_NET_EDGE|已 ablate|最强保护器；取消后胜率和回撤明显恶化。|
|slug_loss_circuit / global halt|未纳入本 replay|属于账户级/市场级风控，不能直接和单笔 entry gate 混评。|

## 两个 Baseline

|scenario|trades|net_pnl|win_rate|max_dd|loss_streak|
|---|---|---|---|---|---|
|Candidate: threshold=0.20, edge=0.020|427|2,342.76|75.88%|66.42|5|
|Current live-like: threshold=0.18, edge=0.015|573|3,284.99|74.87%|140.67|7|

## Imbalance Alone

|scenario|all_on_pnl|all_off_pnl|extra_trades|extra_dd|win_rate_delta|
|---|---|---|---|---|---|
|Candidate: threshold=0.20, edge=0.020|2,342.76|-999.02|1,703|1,090.56|-43.95pp|
|Current live-like: threshold=0.18, edge=0.015|3,284.99|-999.10|1,914|1,024.95|-40.49pp|

![all gates](../analysis/backtests/strategy1_filter_ablation_20260708/charts/all_gates_on_vs_off.png)

这个结果说明：raw imbalance 是信号源，但不是完整交易条件。没有这些可回放的交易质量和时间过滤，策略会被大量低质量盘口、尾部市场和负 EV 机会拖垮；而且 all-off 路径会较早耗尽本金，后续出现大量 `insufficient_cash`，所以它是一个退化压力测试，不是一个可部署策略。

## Gate Delta: Candidate: threshold=0.20, edge=0.020

|removed gate|extra_trades|delta_pnl|delta_dd|win_rate_delta|loss_streak_delta|
|---|---|---|---|---|---|
|No entry window gate|427|741.13|829.21|-10.77pp|12|
|No entry start gate|99|739.41|11.18|0.55pp|0|
|No re-entry cooldown|197|434.07|94.07|-4.56pp|2|
|No OBI veto|29|252.61|11.46|-0.88pp|3|
|No book quality pack|51|219.09|47.11|-1.82pp|3|
|No entry end gate|326|83.47|754.63|-12.13pp|11|
|No max spread|21|76.29|12.18|-0.21pp|0|
|No max entry price|5|0.44|0.01|-0.42pp|0|
|No force-flatten buffer|0|0.00|0.00|0.00pp|0|
|No BTC trend veto|0|0.00|0.00|0.00pp|0|
|No min net edge|1,163|-1,180.45|1,399.01|-27.01pp|13|
|No replayable gates|1,703|-3,341.79|1,090.56|-43.95pp|50|

![gate delta](../analysis/backtests/strategy1_filter_ablation_20260708/charts/candidate_gate_delta_pnl_dd.png)

## Gate Delta: Current live-like: threshold=0.18, edge=0.015

|removed gate|extra_trades|delta_pnl|delta_dd|win_rate_delta|loss_streak_delta|
|---|---|---|---|---|---|
|No entry start gate|129|974.25|49.93|0.20pp|0|
|No entry window gate|495|916.13|1,218.19|-8.20pp|12|
|No re-entry cooldown|322|276.14|250.38|-6.49pp|4|
|No book quality pack|69|220.34|10.94|-1.82pp|1|
|No OBI veto|36|187.40|22.44|-1.31pp|1|
|No max spread|30|106.92|-1.96|-0.41pp|0|
|No entry end gate|365|102.40|967.54|-9.73pp|5|
|No max entry price|5|0.51|-0.21|-0.30pp|0|
|No force-flatten buffer|0|0.00|0.00|0.00pp|0|
|No BTC trend veto|0|0.00|0.00|0.00pp|0|
|No min net edge|1,600|-1,785.77|1,636.82|-25.03pp|7|
|No replayable gates|1,914|-4,284.09|1,024.95|-40.49pp|22|

![gate delta](../analysis/backtests/strategy1_filter_ablation_20260708/charts/current_gate_delta_pnl_dd.png)

## 风险收益散点

![risk reward](../analysis/backtests/strategy1_filter_ablation_20260708/charts/gate_delta_risk_reward_scatter.png)

横轴是取消某个 gate 后最大回撤的变化，纵轴是净收益变化。理想区域是左上角；右上角表示收益增加但用更多风险换来。`no_entry_start_gate` 是最值得进一步研究的点，`no_min_net_edge_gate` 和 `no_entry_end_gate` 则明显暴露风险。

## Live Decision 触发频率

|reason|all_skips|high_imb_skips|avg_abs_imb|max_abs_imb|
|---|---|---|---|---|
|empty_book|97,468|97,468|0.4803|0.5000|
|after_entry_window|60,920|46,883|0.3724|0.9800|
|taker_flow_veto|22,155|18,717|0.2186|0.7041|
|obi_veto|14,203|10,645|0.1980|0.6940|
|cooldown|32,870|6,668|0.1156|0.6899|
|slug_loss_circuit|28,665|5,501|0.1333|0.6765|
|net_ev_below_min|6,807|5,196|0.2009|0.4804|
|too_late|5,426|3,706|0.3291|0.8050|
|btc_trend_veto|5,347|3,237|0.1685|0.2498|
|requote_slippage|1,713|1,596|0.2328|0.7296|
|spread_too_wide|752|651|0.2506|0.6890|
|entry_price_high|11|9|0.4039|0.4998|

![live blockers](../analysis/backtests/strategy1_filter_ablation_20260708/charts/live_high_imbalance_blockers.png)

这里的 live decision 统计只能说明触发频率，不能直接证明收益贡献。尤其是 `taker_flow_veto`、`requote_slippage`、`slug_loss_circuit`，当前 snapshot replay 没有足够字段严格复现。

## 逐 Gate 判断

|gate removed|interpretation|
|---|---|
|No min net edge|强保护：取消后交易暴增、胜率塌陷、回撤剧增；这是最有用的收益质量过滤器。|
|No entry end gate|有用：尾部入场带来很大回撤和连亏风险，收益增量很薄；建议继续限制尾部。|
|No entry start gate|疑似过严：当前样本里早期入场增加收益，风险只小幅上升；建议二次扫描 0/15/30/45 秒，而不是直接全开。|
|No re-entry cooldown|有风险控制作用：取消后收益可增加，但回撤、连亏、胜率都恶化；可测试 45 秒，不建议直接归零。|
|No OBI veto|当前可能偏严：取消后收益增加，风险也增加但不灾难；更像应调宽阈值，而不是完全删除。|
|No max spread|保护证据偏弱：取消后收益略升且风险没有明显恶化；建议测试 0.05/0.06，而不是彻底取消。|
|No max entry price|几乎无影响：触发很少，保留为保险线即可，不是第一优先调参对象。|
|No BTC trend veto|本轮 replay 边际为零：被它挡住的机会大多又被其他 gate 覆盖；需要记录/重构趋势证据后再判断是否保留 hard veto。|
|No force-flatten buffer|本轮无边际影响：在 entry_end=240 下基本没有改变 replay 结果；仍建议作为退出安全阀保留。|
|No replayable gates|强证明：只看 imbalance 会接近亏光；过滤层整体是必要的。|

## 建议调整

1. 保留并优先调 `MIN_NET_EDGE`。它是最强收益质量 gate。建议下一轮只在 `0.015, 0.020, 0.025, 0.030` 附近和 threshold 联动扫描。
2. 保留 `ENTRY_WINDOW_END_SEC=240`，不要因为尾部也有少量收益就打开最后一分钟。尾部收益/回撤比很差。
3. 专门测试 `ENTRY_WINDOW_START_SEC=0/15/30/45`。当前数据支持降低开始门槛，但需要确认早期 tick 是否有成交/建模偏差。
4. `REENTRY_COOLDOWN_SEC` 不建议直接设为 `0`；可以测试 `30/45/60`，目标是减少回撤同时别错过太多同市场二次机会。
5. `OBI_VETO_THRESHOLD` 建议调宽到 `0.30/0.40` 做二次验证；本轮显示它可能过严，但完全删除仍会提高风险。
6. `MAX_SPREAD=0.04` 保护证据不强，可以和 `0.05/0.06` 比较；不要直接无限放宽。
7. `BTC_TREND_VETO_PCT` 暂时不是第一优先调参对象。要判断它是否真有用，需要把趋势、taker flow、被拒机会后续 outcome 一起落库。
8. 对 `taker_flow_veto` 和 `requote_slippage`，先补日志字段再回测：记录 flow ratio、buy/sell volume、decision book、fresh book、requote slippage。

## 组合进场过滤回测补充

- 本轮补充回测不是新报告，而是对本报告的组合层验证；共测试 `634` 组进场参数组合。
- 数据窗口：`2026-06-23T00:00:00+08:00` 到 `2026-07-08T22:23:42+08:00` HKT，`1,784,390` 条 orderbook snapshot，`4,019` 个市场。
- 组合网格来自三条并行审查线：`threshold × MIN_NET_EDGE`、`entry_start × entry_end × cooldown`、`OBI × spread × max_entry_price`。
- 图表读法：散点图横轴是最大回撤，纵轴是净收益；虚线回撤 `90` 是 strict 风险线，横向虚线是本轮 `previous_candidate` 净收益 `2,431.84`。

### Reference 对照

|case|params|trades|win_rate|net_pnl|max_dd|loss_streak|
|---|---|---|---|---|---|---|
|current_live_like|thr=0.18, edge=0.015, start=45, end=240, OBI=0.2, spread=0.04, cap=0.9, cd=60|585|74.87%|3,388.67|140.67|7|
|previous_candidate|thr=0.2, edge=0.02, start=45, end=240, OBI=0.2, spread=0.04, cap=0.9, cd=60|435|76.09%|2,431.84|66.42|5|

### 组合回测结论

- 当前最推荐的组合候选是 `thr=0.18, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60`：`641` 笔，净收益 `4,038.70`，胜率 `77.22%`，最大回撤 `68.38`，最长连亏 `6`。
- 组合结果支持 sub-agent 的共同判断：不能单独降低 `threshold`；低 threshold 必须配更强 `MIN_NET_EDGE`，并且尾部 `entry_end` 仍应保持 `240` 或更早。
- `ENTRY_WINDOW_START_SEC` 的确值得放宽，组合网格里 `0/15/30` 秒经常优于 `45` 秒；但 `entry_end=255` 仍应视作尾部风险探针，不直接上线。
- 盘口层面，`OBI=0.30/0.40` 与 `spread=0.05/0.06` 的组合能提升部分收益，但需要同时看回撤和最长连亏；`MAX_ENTRY_PRICE=0.75` 只作为保险档比较，不是主收益来源。

![combo risk reward](../analysis/backtests/strategy1_combo_entry_filters_20260708/charts/combo_risk_reward_scatter.png)

### Strict 候选

|family|params|trades|win_rate|net_pnl|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|641|77.22%|4,038.70|68.38|6|10.9%|
|signal_edge_time|thr=0.19, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|575|77.91%|3,574.08|66.62|6|11.7%|
|signal_edge_time|thr=0.19, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|604|76.66%|3,519.76|66.19|6|12.1%|
|signal_edge_time|thr=0.18, edge=0.02, start=15, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|629|75.36%|3,385.78|82.65|6|12.8%|
|time_cooldown_probe|thr=0.2, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=30|545|76.51%|3,227.24|43.01|6|12.5%|
|signal_edge_time|thr=0.18, edge=0.02, start=15, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|599|74.96%|3,167.75|78.30|6|13.0%|
|signal_edge_time|thr=0.2, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|494|78.54%|3,109.46|50.83|6|13.1%|
|time_cooldown_probe|thr=0.2, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|494|78.54%|3,109.46|50.83|6|13.1%|
|signal_edge_time|thr=0.2, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|515|76.50%|3,000.32|52.93|6|13.6%|
|time_cooldown_probe|thr=0.2, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|515|76.50%|3,000.32|52.93|6|13.6%|

### Balanced 候选

|family|params|trades|win_rate|net_pnl|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|641|77.22%|4,038.70|68.38|6|10.9%|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|675|75.85%|3,907.18|94.29|6|11.4%|
|signal_edge_time|thr=0.18, edge=0.025, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|617|77.31%|3,708.65|66.43|7|11.3%|
|signal_edge_time|thr=0.18, edge=0.03, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|601|77.54%|3,634.29|91.73|7|11.8%|
|time_cooldown_probe|thr=0.18, edge=0.03, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|601|77.54%|3,634.29|91.73|7|11.8%|
|time_cooldown_probe|thr=0.18, edge=0.03, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=30|678|74.78%|3,626.38|96.89|7|11.6%|
|signal_edge_time|thr=0.18, edge=0.025, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|651|75.88%|3,607.27|88.34|7|11.9%|
|signal_edge_time|thr=0.19, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|575|77.91%|3,574.08|66.62|6|11.7%|
|signal_edge_time|thr=0.19, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|604|76.66%|3,519.76|66.19|6|12.1%|
|signal_edge_time|thr=0.18, edge=0.03, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|635|76.06%|3,506.51|86.33|7|12.1%|
|time_cooldown_probe|thr=0.18, edge=0.03, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|635|76.06%|3,506.51|86.33|7|12.1%|
|signal_edge_time|thr=0.18, edge=0.035, start=0, end=240, OBI=0.3, spread=0.05, cap=0.9, cd=60|638|75.86%|3,421.31|130.90|7|11.9%|

![combo top candidates](../analysis/backtests/strategy1_combo_entry_filters_20260708/charts/combo_top_candidates.png)

### Raw Top 仅供研究

|family|params|trades|win_rate|net_pnl|max_dd|loss_streak|top10_share|
|---|---|---|---|---|---|---|---|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=240, OBI=0.3, spread=0.05, cap=0.9, cd=60|718|75.21%|4,204.53|165.35|7|11.0%|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=60|641|77.22%|4,038.70|68.38|6|10.9%|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=240, OBI=0.3, spread=0.05, cap=0.9, cd=45|752|73.94%|4,019.12|182.55|7|11.7%|
|time_cooldown_probe|thr=0.18, edge=0.03, start=0, end=255, OBI=0.3, spread=0.05, cap=0.9, cd=60|749|73.56%|3,929.56|232.49|10|11.7%|
|signal_edge_time|thr=0.18, edge=0.02, start=0, end=225, OBI=0.3, spread=0.05, cap=0.9, cd=45|675|75.85%|3,907.18|94.29|6|11.4%|
|time_cooldown_probe|thr=0.18, edge=0.03, start=0, end=255, OBI=0.3, spread=0.05, cap=0.9, cd=30|838|71.12%|3,846.20|272.01|11|11.7%|
|signal_edge_time|thr=0.18, edge=0.025, start=0, end=240, OBI=0.3, spread=0.05, cap=0.9, cd=60|684|75.44%|3,799.45|153.84|7|11.3%|
|time_cooldown_probe|thr=0.18, edge=0.03, start=0, end=255, OBI=0.3, spread=0.05, cap=0.9, cd=45|782|72.38%|3,760.50|237.82|10|12.0%|
|microstructure_quality|thr=0.18, edge=0.02, start=15, end=240, OBI=0.4, spread=0.06, cap=0.9, cd=45|723|73.31%|3,720.91|205.59|7|12.8%|
|signal_edge_time|thr=0.19, edge=0.02, start=0, end=240, OBI=0.3, spread=0.05, cap=0.9, cd=60|644|75.62%|3,718.26|146.73|7|11.8%|

Raw top 不能直接部署；如果它们主要靠 `entry_end=255`、过低 `MIN_NET_EDGE` 或较高回撤取得收益，应归入 research bucket。

### 关键交互图

![threshold edge](../analysis/backtests/strategy1_combo_entry_filters_20260708/charts/combo_threshold_edge_heatmap.png)

![threshold start](../analysis/backtests/strategy1_combo_entry_filters_20260708/charts/combo_threshold_start_heatmap.png)

![obi spread](../analysis/backtests/strategy1_combo_entry_filters_20260708/charts/combo_obi_spread_heatmap.png)

### 更新后的进场判断建议

1. 主线继续围绕 `threshold=0.20-0.22` 与 `MIN_NET_EDGE=0.020-0.030`，不要测试无 edge 或极低 edge 的部署候选。
2. `ENTRY_WINDOW_START_SEC` 下一步可以从 `45` 下调到 `15/30` 做 paper validation；`0` 秒如果样本内很好，也要先检查 early tick 可成交性。
3. `ENTRY_WINDOW_END_SEC` 维持 `240`；`255` 只保留为研究探针。
4. `REENTRY_COOLDOWN_SEC` 不建议归零；组合里优先比较 `45` 与 `60`，如回撤放大则回到 `60`。
5. 盘口质量建议用 `OBI=0.30`、`MAX_SPREAD=0.05` 作为下一轮中心点，再用 `OBI=0.20/0.40` 和 `spread=0.04/0.06` 做邻域确认。
6. 真正上线前还需要 walk-forward 或 holdout；本节仍是样本内 replay。


## 局限性

- 本报告 replay 的是 `orderbook_snapshots`，不是逐笔成交流；没有真实撮合延迟和 partial fill。
- `taker_flow_veto` 没有足够历史字段做严格反事实；live 统计只代表触发频率。
- `requote_slippage` 依赖重新拉取 fresh book，snapshot 里只有单次 book 状态，所以只能近似。
- 取消某个 gate 后，后续仓位、资金曲线、cooldown 状态都会变，delta 不是完全独立的单笔归因。
- 当前结果仍是同一历史窗口内的样本内结论，真正改参数前应做 holdout 或下一段 live paper 验证。
