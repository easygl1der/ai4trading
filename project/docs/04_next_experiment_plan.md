# Next Experiment Plan

## Research Goal

Rebuild this project as a price-discovery and causal-design pilot, not as a completed causal report.

Core question:

\[
\text{Does Bitget RWA perpetual pricing contain useful information before conventional equity markets open?}
\]

## Experiment 1: Bitget RWA Perpetual Tracking Quality

Purpose: verify that Bitget MUUSDT/SNDKUSDT are usable price proxies before any causal work.

Design:

- Compare Bitget MUUSDT with Yahoo MU by session: pre, regular, post, overnight.
- Measure return correlation, price basis, stale-price frequency, and volume/liquidity patterns.
- Repeat for SNDKUSDT and available negative-control RWA contracts.

Pass condition:

- Bitget tracks US equity prices closely during overlap.
- Overnight moves are not dominated by stale ticks or mechanical marks.

## Experiment 2: Predictive Lead-Lag Map

Purpose: describe timing structure without causal language.

Variables:

- \(B_h\): Bitget MUUSDT return at hour \(h\)
- \(K_h\): Korean semiconductor basket return at hour \(h\)
- \(X_h\): SOX lag, VIX, FX, US market return controls

Tests:

- \(B_{h-k} \to K_h\) predictive regressions for \(k = 1,\dots,24\)
- \(K_{h-k} \to B_h\) predictive regressions for \(k = 1,\dots,24\)
- Session-specific estimates for Korean OPEN, MID, CLOSE
- Weekend and holiday gap handling, not row-shift lags

Output:

- A heatmap of predictive coefficients by lag and session.
- Clear label: predictive, not causal.

## Experiment 3: Reverse Causal Design Candidate

Purpose: test whether a real causal design for Bitget-to-Korea transmission is feasible.

Candidate treatment:

\[
D_t = \text{Bitget MUUSDT shock before Korea open}
\]

Candidate outcome:

\[
Y_t = \text{Korean semiconductor OPEN return}
\]

Candidate controls:

- Previous US SOX return
- US MU regular/post-market return
- VIX change
- USD/KRW and USD/JPY changes
- Major scheduled macro/event dummies

Identification requirement:

- Need a shock source not mechanically tied to Korean returns.
- Possible candidates: MU-specific scheduled earnings surprises, exchange-specific Bitget outages/liquidity events, or US-only semiconductor news windows.

If no credible shock exists, keep this as predictive research only.

## Experiment 4: Negative Controls

Purpose: detect spurious global-risk explanations.

Negative-control outcomes:

- Non-semiconductor Korean assets.
- Broad Korea ETF returns after removing semiconductor-heavy exposure.

Negative-control treatments:

- Bitget RWA contracts unrelated to semiconductors.
- Large-cap US names without direct Korea semiconductor supply-chain linkage.

Pass condition:

- A semiconductor-specific channel should be stronger for MU/SNDK/TSM-like names than for unrelated assets.

## Experiment 5: Correct Mediation Timing

Only attempt mediation if the time order is valid:

\[
D_h \to M_{h+a} \to Y_{h+a+b}
\]

Do not use \(M_{h+8}\) to mediate \(Y_{h+1}\).

Candidate paths:

- Korea OPEN shock \(\to\) EWY/KORU pre-market \(\to\) US regular-session MU.
- Bitget overnight MU shock \(\to\) Korea OPEN \(\to\) EWY/KORU pre-market.

## Deliverables

1. Clean data audit notebook or script.
2. Lead-lag heatmap and summary table.
3. Identification memo: causal or predictive-only.
4. If causal design passes, a new estimand document before any estimation.
