#!/usr/bin/env sh
set -eu

# Start Kafka Connect in background using the base image entrypoint
/docker-entrypoint.sh start &
CONNECT_PID=$!

# Wait for REST API
until curl -s http://localhost:8083/connectors >/dev/null 2>&1; do
  echo "[connect] Waiting for Kafka Connect REST on :8083..."
  sleep 2
done
echo "[connect] REST is ready"
echo "[connect] Listing /configs contents for debugging:"
ls -lah /configs || true

upsert_connector() {
  cfg_file="$1"
  if [ ! -f "$cfg_file" ]; then
    return 0
  fi

  name=$(jq -r '.name' "$cfg_file")
  config_json=$(jq -c '.config' "$cfg_file")

  if [ -z "$name" ] || [ "$name" = "null" ]; then
    echo "[connect] Skipping $cfg_file: missing .name"
    return 0
  fi

  echo "[connect] Ensuring connector '$name'"

  # Check if exists
  if curl -sf http://localhost:8083/connectors/"$name" >/dev/null 2>&1; then
    echo "[connect] Updating existing connector '$name'"
    # PUT expects only the flat config map
    resp=$(curl -sf -X PUT -H 'Content-Type: application/json' \
      --data "$config_json" \
      http://localhost:8083/connectors/"$name"/config)
    echo "[connect] PUT response: $resp"
  else
    echo "[connect] Creating connector '$name'"
    # POST expects {"name":"...","config":{...}}
    resp=$(curl -sf -X POST -H 'Content-Type: application/json' \
      --data @"$cfg_file" \
      http://localhost:8083/connectors)
    echo "[connect] POST response: $resp"
  fi

  # Show status after upsert
  status=$(curl -sf http://localhost:8083/connectors/"$name"/status || true)
  echo "[connect] Status for '$name': $status"
}

# Upsert known connectors from /configs
upsert_connector /configs/order-outbox.json || true

# Show final connectors list
final_list=$(curl -sf http://localhost:8083/connectors || true)
echo "[connect] Connectors: $final_list"

wait ${CONNECT_PID}
