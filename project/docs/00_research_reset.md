# Research Reset

Date: 2026-07-02

This project has been reset because the previous report was globally invalid as causal evidence.

## Invalidated Claims

Do not use these claims:

- Korea semiconductor returns have a validated causal effect on Bitget MUUSDT.
- The old CACE estimate identifies a complier average causal effect.
- EWY/KORU mediate Korea-to-MU effects in the old analysis.
- E-value, Rosenbaum, or Manski checks validate the old conclusion.
- Granger/placebo evidence establishes causal direction.

## Root Problems

1. The old IV was endogenous by construction.

   The old design used:

   \[
   \eta_h = D_h - \alpha_s - \beta_s r^{SOX}_{t-1}
   \]

   and then built:

   \[
   Z_h = \eta_h \cdot I_s
   \]

   Because \(Z_h\) is a transformed version of \(D_h\), first-stage strength is mechanical and does not imply valid exogenous variation.

2. The old mediation graph violated time order.

   The old outcome was Bitget MU return near \(h+1\), while EWY/KORU were measured around \(h+8\). A future ETF return cannot mediate an earlier Bitget return.

3. The old robustness checks were attached to an unidentified estimand.

   Robustness checks cannot rescue an invalid identification strategy.

## What Is Still Useful

- Raw market data snapshots.
- Bitget RWA perpetuals as a 24/7 signal source.
- Korea/US market-hours mismatch as a natural timing structure.
- Mermaid diagrams as research maps, after correction.
- The negative result as a warning: lead-lag analysis must not be converted into causal language without an external shock or a well-defined quasi-experiment.

## New Rule

Every future analysis must pass a timing and identification check before reporting estimates:

- Treatment must occur before mediator and outcome.
- Instruments must not be deterministic transformations of the treatment.
- Placebo/Granger results must be labelled predictive unless paired with a causal design.
- Derived data must carry a status label: `raw`, `pilot`, `invalid`, or `validated`.
