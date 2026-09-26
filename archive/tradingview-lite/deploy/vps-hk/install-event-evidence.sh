#!/usr/bin/env sh
set -eu

# Run on vps-hk after copying a verified Linux binary to the first argument.
# The script preserves a timestamped binary/environment backup and never prints secrets.

binary_path=${1:?usage: install-event-evidence.sh /path/to/tradingview-lite-backend}
service_dir=${SERVICE_DIR:-/home/yueyh/services/tradingview-lite-backend}
env_file="$service_dir/tradingview-lite.env"
adapter_env=${ADAPTER_ENV:-/home/yueyh/firecrawl-compatible-onyx-adapter/.env}
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
backup_dir="$service_dir/backups/$timestamp"

test -x "$binary_path"
test -f "$env_file"
test -f "$adapter_env"

adapter_key=$(sed -n 's/^ADAPTER_API_KEY=//p' "$adapter_env" | head -n 1)
test -n "$adapter_key"

mkdir -p "$backup_dir"
cp "$service_dir/tradingview-lite-backend" "$backup_dir/tradingview-lite-backend"
cp "$env_file" "$backup_dir/tradingview-lite.env"

temporary_env=$(mktemp "$service_dir/.tradingview-lite.env.XXXXXX")
trap 'rm -f "$temporary_env"' EXIT

grep -Ev '^(ENABLE_EVENT_EVIDENCE|FIRECRAWL_NEWS_URL|FIRECRAWL_API_KEY|FIRECRAWL_REQUESTS_PER_MINUTE|FIRECRAWL_NEWS_MAX_AGE_MINUTES|EVENT_SYMBOL_INTERVAL|EVENT_MACRO_INTERVAL|EVENT_CONTEXT_INTERVAL|SHADOW_SIGNAL_INTERVAL|EVENT_TRIGGERED_COOLDOWN|EVENT_SYMBOLS|SHADOW_SYMBOLS|EVENT_MACRO_QUERIES|EVENT_GROUP_QUERIES|EVENT_OFFICIAL_DOMAINS|EVENT_EVIDENCE_RETENTION|EVENT_CONTEXT_RETENTION|SHADOW_SIGNAL_RETENTION|EVENT_RUN_RETENTION|ENABLE_BEHAVIOR_POLICY_SHADOW|POLICY_CANDIDATE_INTERVAL|POLICY_OUTCOME_INTERVAL|POLICY_CANDIDATE_COOLDOWN|POLICY_DAILY_CANDIDATE_BUDGET|POLICY_SHADOW_SYMBOLS|POLICY_CANDIDATE_RETENTION|POLICY_OUTCOME_RETENTION)=' "$env_file" > "$temporary_env" || true

printf '%s\n' \
  'ENABLE_EVENT_EVIDENCE=true' \
  'FIRECRAWL_NEWS_URL=http://127.0.0.1:8793/v1/news/search' \
  "FIRECRAWL_API_KEY=$adapter_key" \
  'FIRECRAWL_REQUESTS_PER_MINUTE=10' \
  'FIRECRAWL_NEWS_MAX_AGE_MINUTES=1440' \
  'EVENT_SYMBOL_INTERVAL=15m' \
  'EVENT_MACRO_INTERVAL=10m' \
  'EVENT_CONTEXT_INTERVAL=1m' \
  'SHADOW_SIGNAL_INTERVAL=30s' \
  'EVENT_TRIGGERED_COOLDOWN=2m' \
  'EVENT_SYMBOLS=MU,NVDA,AMD,SMH,RKLB,SPCX,ARKX,TE,FCEL,IONQ,RGTI,QBTS,QUBT,SPY,QQQ,IWM,TLT,^VIX' \
  'SHADOW_SYMBOLS=MU,NVDA,RKLB,IONQ,SPY,QQQ' \
  'EVENT_MACRO_QUERIES=SPY QQQ market moving news,Federal Reserve CPI jobs report' \
  'EVENT_GROUP_QUERIES=semis_memory=semiconductor stocks market news|space=space stocks market news|energy_clean=clean energy stocks market news|quantum=quantum computing stocks market news' \
  'EVENT_OFFICIAL_DOMAINS=sec.gov,federalreserve.gov,bls.gov,bea.gov' \
  'EVENT_EVIDENCE_RETENTION=2160h' \
  'EVENT_CONTEXT_RETENTION=2160h' \
  'SHADOW_SIGNAL_RETENTION=2160h' \
  'EVENT_RUN_RETENTION=2160h' >> "$temporary_env"
printf '%s\n' \
  'ENABLE_BEHAVIOR_POLICY_SHADOW=true' \
  'POLICY_CANDIDATE_INTERVAL=1m' \
  'POLICY_OUTCOME_INTERVAL=5m' \
  'POLICY_CANDIDATE_COOLDOWN=30m' \
  'POLICY_DAILY_CANDIDATE_BUDGET=5' \
  'POLICY_SHADOW_SYMBOLS=MU,NVDA,RKLB,IONQ,SPY,QQQ' \
  'POLICY_CANDIDATE_RETENTION=2160h' \
  'POLICY_OUTCOME_RETENTION=2160h' >> "$temporary_env"

chmod 600 "$temporary_env"
mv "$temporary_env" "$env_file"
install -m 0755 "$binary_path" "$service_dir/tradingview-lite-backend"

if sudo -n true >/dev/null 2>&1; then
  sudo -n systemctl restart tradingview-lite-backend.service
  sudo -n systemctl is-active --quiet tradingview-lite-backend.service
else
  systemctl restart tradingview-lite-backend.service
  systemctl is-active --quiet tradingview-lite-backend.service
fi
printf 'backup_dir=%s\n' "$backup_dir"
