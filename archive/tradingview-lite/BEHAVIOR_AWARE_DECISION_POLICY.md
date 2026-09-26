# Behavior-Aware Market Decision Policy

Status: deployed shadow-only implementation, 2026-07-13 HKT

This is a living design record. It converts the user's stated trading, risk, and behavioral needs into constraints for TradingView-Lite. It is not investment advice, a prediction engine, or permission to automate orders.

## 1. User Objective

The primary problem is not a lack of price alerts. It is the cognitive and emotional cost of FOMO, chasing strength, selling into fear, and continuously monitoring a large watchlist from China while US markets trade overnight.

Priority order:

1. Reduce avoidable FOMO and chase/sell impulses.
2. Protect existing positions and avoid unsupported risk escalation.
3. Identify rare, high-quality opportunity states rather than frequent small-edge trades.
4. Save attention and sleep without hiding genuine overnight or session-transition risk.
5. Produce explanations and candidate actions, never automatic orders or certainty claims.

The system should prefer `BLOCKED`, `WAIT`, or `NO_CHASE` over a low-confidence directional recommendation.

## 1A. Implementation Status

The first behavior-aware policy layer is deployed on `vps-hk` in shadow-only mode.

- Candidate records, profiles, peer-map versions, feedback, and post-hoc outcomes are persisted in the isolated TradingView-Lite PostgreSQL database.
- The policy scheduler creates candidates for `MU`, `NVDA`, `RKLB`, `IONQ`, `SPY`, and `QQQ` every minute and evaluates eligible candidates after 15, 30, 60, and 180 minutes.
- All policy candidate rows are database-constrained to `delivery_mode=shadow_only`. The policy implementation has no Discord notifier dependency and the preview endpoint reports `normalDiscordWebhookInvoked=false`.
- Current policy symbols are explicitly initialized as `research_watch` with only `WAIT_CONFIRMATION` and `NO_CHASE` permissions. This is intentionally neutral; no displayed share count was treated as proof of a position.
- `semis_memory`, `space`, `energy_clean`, `quantum`, and `market_context` have proposed peer-map version 1 records. None is approved, so none currently influences policy eligibility.
- The first live opening-session records correctly moved from startup `BLOCKED` to `PRICE_DISCOVERY` and then `WAIT_CONFIRMATION` after market inputs became fresh. No market Discord message was sent.

The current range and IV calibration state is deliberately provisional. The system needs completed trading days before it can estimate coverage or bias. RWA currently exposes a descriptive contemporaneous alignment statistic only; it is not treated as evidence that RWA leads US cash equity.

## 2. Current User Profile

### Trading horizons

- Intraday speculation and medium-term investment coexist.
- The system must distinguish a position-management question from a watchlist-opportunity question before suggesting an action.
- The displayed watchlist is not yet a reliable position ledger. A visible share count in a broker/watchlist interface is not enough to infer cost basis, thesis, intended holding period, or allowed risk.

### Desired decision language

Allowed action candidates:

- `持仓观察` (hold and observe)
- `减仓复核` (review a reduction)
- `止损复核` (review a stop)
- `等待确认` (wait for confirmation)
- `禁止追高` (no-chase intervention)
- `高质量机会候选` (high-quality opportunity candidate)

The last state is an eligibility flag, not a buy instruction. The user explicitly prefers fewer, larger expected opportunities over frequent small trades. The meaning and threshold of the user's term `ERR` still needs confirmation before it can become a rule.

### Attention budget

- At most five immediate Discord notifications per day.
- Value is measured by fewer missed material events, less avoidable loss, less impulsive behavior, less screen time, and eventual evidence that the alerted state was genuinely imbalanced or unusually favorable.
- Notification volume itself is a risk metric. A correct but non-actionable alert can still be harmful if it consumes attention or disrupts sleep.

## 3. Session-Specific Products

The system must not use one global intraday threshold. Volatility and price discovery are time-dependent.

| Session product | User question | Required inputs | Initial output states |
|---|---|---|---|
| Overnight evidence brief | Did material information appear while US cash equity was closed? | Firecrawl evidence, RWA proxy, prior close, source freshness | `NO_NEW_VERIFIED_EVENT`, `EVENT_CONTEXT`, `RWA_DIVERGENCE`, `BLOCKED` |
| Premarket plan | Is the premarket gap plausible relative to expected daily movement and evidence? | prior close, premarket quote/bar, RWA, option expected move, market/sector context | `GAP_WITHIN_RANGE`, `GAP_EXTENDED`, `GAP_UNSUPPORTED`, `OPENING_RISK_UNRESOLVED` |
| Opening 0-30 minute monitor | Is the opening move still price discovery, supported continuation, or a failed impulse? | opening price, actual elapsed-time speed, finalized `1m` bars, range use, relative move, volume, evidence context | `PRICE_DISCOVERY`, `SUPPORTED_MOVE`, `UNSUPPORTED_IMPULSE`, `NO_CHASE`, `BLOCKED` |
| Midday sleep-risk brief | Before the user sleeps, is the remaining session stabilizing or still prone to re-acceleration? | realized range, post-open speed/volatility decay, option/range context, market/sector state, unresolved event risk | `STABILIZING`, `UNRESOLVED_RISK`, `RE_ACCELERATION_RISK`, `BLOCKED` |
| Post-close review | What actually happened, and which candidate alerts were useful? | final bars, event evidence, shadow observations, later excursions | structured scorecard only |

The opening 30 minutes must not be assumed to be automatically tradable. It is treated as a high-information, high-noise price-discovery window. A normal default is to collect evidence and suppress chase recommendations until a session-specific confirmation condition is met.

## 4. Range and Volatility Policy

The preferred range anchor is a combination of previous close, premarket/overnight path, opening price, and option-implied expected movement. Historical realized volatility is a calibration baseline and a tail-context feature, not the sole driver.

For each symbol/session, preserve separate range components:

\[
M_{\mathrm{IV}} = S\sigma_{\mathrm{IV}}\sqrt{1/252}
\]

where \(S\) is the contemporaneous underlying spot and \(\sigma_{\mathrm{IV}}\) is the annualized IV re-solved from the selected option price.

Also preserve the directly observed near-ATM straddle move. These two estimates are cross-checks, not interchangeable values.

Core diagnostics:

\[
\mathrm{range\ use}_t = \frac{|P_t-P_{\mathrm{anchor}}|}{M_{\mathrm{selected}}}
\]

- `anchor` is explicitly labeled as prior close, premarket reference, or regular-session open.
- A move beyond an expected range is evidence of a dislocation, not proof of reversal.
- The continuation-versus-overreaction judgment requires follow-through, retracement behavior, relative strength, evidence quality, and later calibration.

### Same-day and 0DTE options

The user is especially interested in same-day/0DTE options. They can be informative about the market's priced short-horizon amplitude, but current free Yahoo options data is not an OPRA-quality, low-latency feed.

Initial policy:

- Re-solve IV from bid/ask midpoint; do not trust Yahoo's raw IV field.
- Preserve expiry, exact option price source, spread, underlying timestamp, contract timestamp, and freshness warning.
- Use 0DTE data as a provisional range/context input, not an execution or microsecond-level signal.
- Do not infer order-flow direction or a precise probability of profit from a free delayed chain.
- A paid/broker-grade options source is a later prerequisite for reliable same-day skew, flow, and very short-horizon IV monitoring.

## 5. RWA / Tokenized Overnight Policy

Bitget/Ondo RWA data may help describe overnight direction, gap risk, liquidity divergence, and reactions to information released while US cash equity is closed. It is not Nasdaq/NYSE trade data, NBBO, Level 2, or automatically authoritative price discovery.

The system will measure rather than assume usefulness:

1. Overlapping-session correlation with the US-equity source.
2. Lead/lag relationship at 1-minute and sampled-quote horizons.
3. RWA-to-equity premium/discount before open and its subsequent convergence or divergence.
4. Freshness, quoted volume, provider-market status, and data gaps.
5. Conditional behavior on high-impact evidence days versus quiet days.

Until this calibration is complete, RWA may change an explanation confidence label but cannot independently create an entry/exit candidate.

## 6. Continuation, Overreaction, and Panic-Like Dislocation

The system should not label a move `panic` merely because it is large. The initial implementation records a feature vector and evidence state, then evaluates which combinations predict poor follow-through or reversal for each symbol and session.

Candidate observable features:

- price speed and acceleration using actual elapsed sample time;
- repeated impulse, horizontal pause, and renewed breakdown/rebound pattern;
- retracement from the local extreme and failure to reclaim an impulse midpoint or VWAP;
- realized range use versus IV/straddle range;
- relative move versus market, sector, and peer leaders;
- volume persistence or rapid fade after an impulse;
- fresh, source-tiered event evidence;
- RWA divergence as a low-confidence overnight context feature.

Initial states are descriptive:

| State | Meaning | Allowed action candidate |
|---|---|---|
| `PRICE_DISCOVERY` | open/gap still being resolved | `等待确认` |
| `SUPPORTED_MOVE` | price action and context have not invalidated follow-through | `持仓观察` only |
| `UNSUPPORTED_IMPULSE` | speed/range extension lacks adequate confirmation | `禁止追高` |
| `DISLOCATION_RISK` | repeated/extreme behavior exceeds calibrated local baseline | `减仓复核` or `止损复核` |
| `BLOCKED` | stale/missing/ambiguous data | no action candidate |

No weighted composite score is fixed before shadow data establishes empirical distributions. The initial purpose is to detect conditions worth reviewing, not to claim that a particular feature combination predicts a specific return.

## 7. FOMO and Attention Protocol

The system is a behavioral guardrail. Every candidate alert must answer four questions before delivery:

1. Is the data fresh and provider-qualified?
2. Is the move unusual for this symbol and this session, not merely large in raw percentage terms?
3. Is there verified evidence or is the explanation still absent/weak?
4. Is there a non-impulsive action candidate, including `no action`?

`NO_CHASE` is an explicit valuable output. It can be delivered once per symbol/session only when a rapid move is extended relative to the selected range and lacks the required continuation evidence. It must never be phrased as a prediction that price will reverse.

Daily notification policy to design and shadow-test:

- hard budget: at most five immediate messages per New York trading date;
- per symbol/state cooldown;
- `URGENT` reserved for position-aware, fresh, source-qualified material risk;
- `INFO` and repeated `BLOCKED` states remain dashboard/summary-only;
- a bedtime brief is a single summary, not a stream of alerts.

## 8. Evaluation Framework

The first scorecard is not PnL. It evaluates whether the system reduces bad decisions and attention cost without suppressing material information.

For every shadow observation, record later outcomes without pretending that they were known at decision time:

- maximum favorable and adverse excursion over 15, 30, 60, and 180 minutes;
- range extension/reversion relative to the anchor;
- whether the state changed after new evidence arrived;
- alert eligibility, suppression reason, and notification-budget use;
- user feedback: `acted`, `did_not_act`, `helpful`, `not_helpful`, and optional emotion intensity from 0 to 5;
- time saved and whether the alert avoided unnecessary screen checking.

Candidate policy promotion requires both quantitative and behavioral evidence. A high raw hit rate is insufficient if messages increase FOMO or disrupt sleep.

The next review gate should require all of the following before a user considers enabling any normal policy delivery:

1. At least 20 completed, matched IV/range days for a symbol before interpreting calibration coverage or bias.
2. A sufficient regular-session sample for the impulse baseline; the existing shadow model requires at least five sessions and 30 observed speed samples.
3. Review of every would-notify candidate, including its 15/30/60/180-minute excursions and whether its action language reduced or increased the urge to chase.
4. Explicit approval of both the relevant symbol profile and its peer-map version.
5. An explicit user decision to change the delivery policy. There is no automatic promotion path.

## 9. Research Notes and Boundaries

The following sources motivate the architecture, not a mechanical trading rule:

- Andersen and Bollerslev document intraday periodicity and volatility persistence, supporting separate session-aware baselines rather than one all-day threshold: [DOI: 10.1016/S0927-5398(97)00004-2](https://doi.org/10.1016/S0927-5398(97)00004-2).
- Boudt, Croux, and Laurent study time-varying intraday periodicity, reinforcing that a fixed opening-window heuristic can become stale: [DOI: 10.1080/01621459.2018.1512864](https://doi.org/10.1080/01621459.2018.1512864).
- Blair, Poon, and Taylor study the incremental information in implied volatility and high-frequency returns for future volatility, supporting IV plus realized data as an amplitude framework rather than direction forecasting: [DOI: 10.1016/S0304-4076(01)00068-9](https://doi.org/10.1016/S0304-4076(01)00068-9).
- A recent tokenized-stocks working paper is a useful research locator, but it is not enough to justify treating a tokenized proxy as US-equity price truth: [DOI: 10.2139/ssrn.5937314](https://doi.org/10.2139/ssrn.5937314).

The current literature scan is intentionally incomplete. Some public search endpoints were rate-limited or blocked during this pass. Before a production claim about opening-auction or RWA predictive power, add a source-level literature review and evaluate the effect on this watchlist's own stored samples.

## 10. Next Goal

The implemented layer now provides:

1. Add a user position/watchlist profile with explicit role, horizon, cost basis optionality, and action permissions.
2. Add session-specific premarket, opening, midday, and overnight context snapshots.
3. Add benchmark/peer maps that are versioned, reviewable, and empirically audited rather than hard-coded assertions.
4. Add range-use, gap, unresolved-risk, and `NO_CHASE` candidate states to the shadow store.
5. Add a post-hoc evaluation ledger, including user feedback and attention/notification budget metrics.
6. Produce up to five daily Discord candidate previews in dry-run form only; do not deliver normal market messages until policy review is explicitly approved.

The remaining goal is observation and policy review, not more automatic action: collect sufficient completed-session samples, classify actual positions and tactical watch symbols, review the proposed peer maps, and decide whether any candidate class deserves normal delivery.

## 11. Remaining Decisions for the User

1. Which symbols are actual current positions, and for each, what is the intended horizon: intraday, swing, or medium term?
2. What does `ERR` mean in the user's language, and what minimum ratio/threshold makes an opportunity worth attention?
3. What is the maximum acceptable loss or drawdown review point for an intraday position versus a medium-term investment?
4. Should the default opening policy be a hard no-new-entry window for the first 30 minutes, unless a separate high-confidence exception is proved?
5. At what China/Hong Kong time should the one-per-day midday sleep-risk brief arrive?
6. Is a one-tap user feedback loop acceptable after an alert: helpful/not helpful, acted/did not act, emotion 0-5?
7. May the system maintain a separate `position`, `tactical_watch`, and `research_watch` list rather than treating all displayed symbols equally?
