# Invalid Archive

This folder preserves the old causal-analysis artifacts that were removed from the active project on 2026-07-02.

The files here are for audit only. Do not cite their numerical results as valid findings.

## Why Archived

- The IV design used a residualized version of the treatment as the instrument.
- The mediation analysis used a mediator measured after the outcome.
- Sensitivity analysis was interpreted as robustness despite invalid identification.
- Granger/placebo tests were overinterpreted as causal evidence.

## Contents

- `docs/`: old causal framework, IV design, results report, pitfalls, alternatives, and reproducibility guide.
- `scripts/`: old stage analysis scripts that generated invalid derived variables and results.
- `data_analysis/`: old derived analysis parquet/JSON outputs.

Raw source data remains in the active `data/` folder.
