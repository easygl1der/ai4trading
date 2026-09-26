#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="tradingview-lite-backend.service"
APP_DIR="/home/yueyh/services/tradingview-lite-backend"
ENV_FILE="${APP_DIR}/tradingview-lite.env"

validate_webhook() {
  local name="$1"
  local value="$2"
  if [[ -z "$value" ]]; then
    echo "${name} is required." >&2
    return 1
  fi
  case "$value" in
    https://discord.com/api/webhooks/*|https://discordapp.com/api/webhooks/*)
      return 0
      ;;
    *)
      echo "${name} does not look like a Discord webhook URL." >&2
      return 1
      ;;
  esac
}

mkdir -p "$APP_DIR"
umask 077

printf "Paste #market-alerts webhook URL: "
IFS= read -r -s alerts_webhook
printf "\nPaste #system-health webhook URL: "
IFS= read -r -s system_webhook
printf "\n"

validate_webhook "DISCORD_WEBHOOK_ALERTS" "$alerts_webhook"
validate_webhook "DISCORD_WEBHOOK_SYSTEM" "$system_webhook"

cat > "$ENV_FILE" <<EOF
APP_HOST=127.0.0.1
APP_PORT=8088
WATCHLIST_SYMBOLS=MU,NVDA,AMD,SMH,RKLB,SPCX,ARKX,TE,FCEL,IONQ,RGTI,QBTS,QUBT,SPY,QQQ,IWM,TLT,^VIX
POLL_INTERVAL=60s
DATA_STALE_AFTER=120s
ENABLE_SCHEDULER=true
DATABASE_URL=
DISCORD_WEBHOOK_ALERTS=${alerts_webhook}
DISCORD_WEBHOOK_SYSTEM=${system_webhook}
EOF

chmod 600 "$ENV_FILE"

if command -v systemctl >/dev/null 2>&1; then
  sudo systemctl restart "$SERVICE_NAME"
  sudo systemctl --no-pager --full status "$SERVICE_NAME" | sed -n '1,18p'
fi

echo "Discord webhooks saved to ${ENV_FILE} and ${SERVICE_NAME} restarted."
