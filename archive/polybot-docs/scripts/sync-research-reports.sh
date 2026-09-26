#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source_docs="$root/docs"
target="$root/polybot-docs/reports"
assets="$root/polybot-docs/analysis"

mkdir -p "$target"
for report in polybot_strategy1_parameter_sensitivity_report_20260708.md polybot_strategy1_filter_ablation_report_20260708.md polybot_trade_timing_report_20260707.md polybot_trade_hour_of_day_report_20260708.md polybot_orderbook_snapshot_insights_20260706.md polybot_strategy_lab_incident_20260716.md polybot_strategy_lab_comparative_review_20260717.md; do
  cp "$source_docs/$report" "$target/$report"
done
perl -0pi -e 's/<!-- COMBO_ENTRY_FILTERS_(?:START|END) -->\n//g' "$target/polybot_strategy1_filter_ablation_report_20260708.md"
perl -0pi -e 's/(Exit reason groups: )(\{[^}]+\})/$1`$2`/g' "$target/polybot_trade_timing_report_20260707.md"
perl -0pi -e 's/\\\[(.*?)\\\]/\$\$$1\$\$/gs; s/\\\((.*?)\\\)/\$$1\$/gs' "$target"/*.md
rg --no-filename -o '\.\./analysis/[A-Za-z0-9_./-]+' "$target" | sort -u | while IFS= read -r reference; do
  relative="${reference#../analysis/}"
  source="$root/analysis/$relative"
  destination="$assets/$relative"
  if [[ -f "$source" ]]; then
    mkdir -p "$(dirname "$destination")"
    cp "$source" "$destination"
  fi
done
echo "Synchronized $(find "$target" -maxdepth 1 -name '*.md' | wc -l | tr -d ' ') Mintlify report pages."
