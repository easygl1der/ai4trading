# Research corpus (Polybot paper gate)

Curated external material for **trade / skip / pause** gates on BTC/ETH 5m/15m Polymarket binaries. Trading flags and runtime config live elsewhere; this tree is read-only reference.

## Layout

| Path | Contents |
| --- | --- |
| [notes/](notes/) | Seven topic guides (papers, mean-reversion, regression gates, stable reversion, public discussion, YouTube) |
| [papers/manifest.json](papers/manifest.json) | Included papers: arXiv IDs, SSRN/DOI link-only entries, scores, sources |
| [papers/arxiv/](papers/arxiv/) | **21** committed PDFs (~23.7 MB); **22** arXiv IDs total — run the script for any missing file (see `pdf_in_repo: false` in manifest) |
| [scripts/download_research_papers.sh](scripts/download_research_papers.sh) | Fetch missing arXiv PDFs into `papers/arxiv/` |

## Inclusion rule

Keep if **score ≥ 6** in `notes/papers.md`, or listed in the strategy note sets (`mean-reversion`, `regression-gates`, `stable-reversion`). Dropped/junk entries from `papers.md` are excluded. Negative control [2608.21888](https://arxiv.org/abs/2608.21888) is included on purpose.

Scan window documented in notes: **2026-09-25** (corpus frozen for this branch).

## Authority

Product intent (runtime): [CONTEXT-FROM-TRADING-BOT.md](https://github.com/easygl1der/ai4trading-polybot/blob/v2/polybot-v2/docs/CONTEXT-FROM-TRADING-BOT.md) in **ai4trading-polybot**. This tree is the learning copy in **ai4trading**; it does not change trading switches.

Migrated from polybot `research/paper-library-2026-09` (PR #5, not merged into `v2` as-is). Source v2 baseline: `92402de`.
