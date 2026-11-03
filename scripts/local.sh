#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
env_file=.env.local
action=${1:-up}
if [[ "$action" == init ]]; then
  if [[ -e "$env_file" ]]; then
    echo "$env_file already exists; keeping its secrets."
    exit 0
  fi
  umask 077
  {
    printf 'POSTGRES_PASSWORD=%s\n' "$(openssl rand -hex 24)"
    printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)"
    printf 'NEXTAUTH_SECRET=%s\n' "$(openssl rand -hex 32)"
    printf 'DKIM_MASTER_KEY=%s\n' "$(openssl rand -hex 32)"
  } > "$env_file"
  echo "Created $env_file with private permissions."
  exit 0
fi
if [[ ! -f "$env_file" ]]; then
  echo 'Run scripts/local.sh init first.' >&2
  exit 1
fi
compose=(docker compose --env-file "$env_file" -f compose.local.yml)
case "$action" in
  up) "${compose[@]}" up --build -d --wait ;;
  down) "${compose[@]}" down ;;
  logs) "${compose[@]}" logs --tail 100 ;;
  setup) "${compose[@]}" exec api mailit setup --config /etc/mailit/mailit.yaml ;;
  backup)
    umask 077
    mkdir -p data/backups
    file="data/backups/mailit-$(date -u +%Y%m%dT%H%M%SZ).sql"
    "${compose[@]}" exec -T postgres pg_dump -U mailit -d mailit --no-owner --no-acl > "$file"
    echo "Saved $file. Keep .env.local with it to decrypt existing DKIM keys."
    ;;
  *) echo 'Usage: scripts/local.sh {init|up|down|logs|setup|backup}' >&2; exit 2 ;;
esac
