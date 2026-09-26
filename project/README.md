# Bitget x Korea Semis Research Workspace

This workspace has been reset on 2026-07-02.

The previous causal report is invalid and must not be cited as a result. The raw data snapshot and the research idea are still useful: Bitget RWA perpetuals may provide a 24/7 price-discovery signal around US semiconductor names, and Korean semiconductor equities provide a natural market-hours contrast.

## Current Status

Valid assets:

- Raw data snapshots under `data/bitget`, `data/yahoo`, `data/korea`, `data/etf`, `data/macro`, `data/fred`, and `data/meta`.
- The data manifest under `data/manifest.md` and `data/manifest.json`.
- The corrected Mermaid research maps in `docs/02_dag_mermaid.md`.
- The next-round experiment plan in `docs/04_next_experiment_plan.md`.
- User and communication preferences in `docs/07_user_preferences.md`.

Invalid and archived:

- Old ATE/CACE/IV/mediation/sensitivity reports.
- Old derived analysis parquet files.
- Old stage analysis scripts that generated invalid causal claims.

They are preserved for audit only under:

`archive_invalid_2026-07-02/`

## Why The Old Report Was Invalid

The old report made causal claims that were not identified.

1. The proposed instruments were constructed from the treatment itself. `eta` is a residualized version of `D_h`, so `Z = eta * session` is not an external instrument.
2. The mediation timing was impossible. The mediator EWY/KORU was measured around `h+8`, while the outcome `Y_h` was measured around `h+1`.
3. Granger and placebo tests were interpreted as causal direction, but they only establish predictive lead-lag structure.
4. E-value, Rosenbaum, and Manski sections were used as robustness language without a valid underlying identification design.

## What Remains Valuable

The valuable idea is not "Korea causes MU" or "MU causes Korea" as a completed conclusion. The valuable idea is a research design question:

Can 24/7 Bitget RWA perpetual markets reveal price discovery before conventional equity markets open, and can that signal be separated from ordinary global semiconductor shocks?

The useful next research direction is to treat the current dataset as a pilot and rebuild from a clean estimand:

- Price discovery: Does Bitget MUUSDT move before Korean semiconductor equities open?
- Reverse transmission: Does a Bitget MU shock predict Korean semiconductor returns after controlling for US semiconductor/global risk factors?
- Robustness: Does the same pattern appear for negative controls and non-semiconductor RWA contracts?
- Basis: How closely do Bitget RWA perpetuals track Yahoo/US equity prices by session?

## Suggested Reading Order

1. `docs/00_research_reset.md`
2. `docs/02_dag_mermaid.md`
3. `docs/03_data_pipeline.md`
4. `docs/04_next_experiment_plan.md`
5. `archive_invalid_2026-07-02/README.md` only if auditing old mistakes

## Do Not Use

Do not use archived ATE/CACE/NDE/NIE/E-value numbers as research findings. They are preserved only to document what failed and why.
