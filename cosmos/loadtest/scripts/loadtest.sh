#!/usr/bin/env bash

set -eo pipefail

function json_processor() {
  local json_file=$1 && shift
  local jq_opt=("$@")

  jq "${jq_opt[@]}" "$json_file" >"$json_file.tmp" &&
    mv "$json_file.tmp" "$json_file"
}

function datetime_since() {
  python3 -c "import time; print((time.mktime(time.strptime('${1%.*}', '%Y-%m-%dT%H:%M:%S'))-time.mktime(time.strptime('${2%.*}', '%Y-%m-%dT%H:%M:%S'))))" 2>/dev/null
}

function current_time() {
  python3 -c "from datetime import datetime, timezone; print(datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%S.%fZ'))"
}

function get_latest_block_and_time() {
  # get latest block header
  latest_block_header=$(curl -s "http://127.0.0.1:26657/commit" | jq -r '.result.signed_header.header')
  # get latest block time
  latest_block_time=$(echo "$latest_block_header" | jq -r '.time')
  # get latest block height
  latest_block_height=$(echo "$latest_block_header" | jq -r '.height')
  echo "$latest_block_height $latest_block_time"
}

function avg_block_time_interval() {
  local block_interval=10
  read -r latest_block_height latest_block_time < <(get_latest_block_and_time)
  # get block time of latest_block - block_interval
  block_time=$(curl -s "http://127.0.0.1:26657/commit?height=$((latest_block_height - block_interval))" | jq -r '.result.signed_header.header.time')
  # calculate avg block time interval
  block_time_interval=$(datetime_since "$latest_block_time" "$block_time")
  python3 -c "print(${block_time_interval:-"0"}/$block_interval)"
}

function pprof() {
    go tool pprof -http 127.0.0.1:8088 "http://127.0.0.1:6060/debug/pprof/profile?seconds=60"
}

function node_catching_up() {
  local node_url=${1:-"http://127.0.0.1:26657"}
  local timeout=${2:-"30"}
  for i in $(seq "$timeout"); do
    sync_state=$(curl -s "$node_url/status" | jq -r '.result.sync_info.catching_up')
    latest_block_height=$(curl -s "$node_url/status" | jq -r '.result.sync_info.latest_block_height')
    if [[ "$sync_state" != "false" || $latest_block_height -le 5 ]]; then
      sleep 1 && echo "Node is syncing... $i, latest block height: $latest_block_height" && continue
    fi
    return 0
  done
  echo "Timeout: Node is not catching up"
  return 1
}

function block_timout_with_3s() {
  export TIMEOUT_PROPOSE="1s" TIMEOUT_PROPOSE_DELTA="50ms"
  export TIMEOUT_PREVOTE="100ms" TIMEOUT_PREVOTE_DELTA="50ms"
  export TIMEOUT_PRECOMMIT="100ms" TIMEOUT_PRECOMMIT_DELTA="50ms"
  export TIMEOUT_COMMIT="3s"
}

function block_timout_with_1s() {
  export TIMEOUT_PROPOSE="500ms" TIMEOUT_PROPOSE_DELTA="50ms"
  export TIMEOUT_PREVOTE="100ms" TIMEOUT_PREVOTE_DELTA="50ms"
  export TIMEOUT_PRECOMMIT="100ms" TIMEOUT_PRECOMMIT_DELTA="50ms"
  export TIMEOUT_COMMIT="1s"
}

function block_timout_with_500ms() {
  export TIMEOUT_PROPOSE="200ms" TIMEOUT_PROPOSE_DELTA="50ms"
  export TIMEOUT_PREVOTE="100ms" TIMEOUT_PREVOTE_DELTA="50ms"
  export TIMEOUT_PRECOMMIT="100ms" TIMEOUT_PRECOMMIT_DELTA="50ms"
  export TIMEOUT_COMMIT="500ms"
}

function block_timout_with_200ms() {
  export TIMEOUT_PROPOSE="100ms" TIMEOUT_PROPOSE_DELTA="50ms"
  export TIMEOUT_PREVOTE="100ms" TIMEOUT_PREVOTE_DELTA="50ms"
  export TIMEOUT_PRECOMMIT="100ms" TIMEOUT_PRECOMMIT_DELTA="50ms"
  export TIMEOUT_COMMIT="200ms"
}

function block_timout_with_100ms() {
  export TIMEOUT_PROPOSE="50ms" TIMEOUT_PROPOSE_DELTA="50ms"
  export TIMEOUT_PREVOTE="50ms" TIMEOUT_PREVOTE_DELTA="50ms"
  export TIMEOUT_PRECOMMIT="50ms" TIMEOUT_PRECOMMIT_DELTA="50ms"
  export TIMEOUT_COMMIT="100ms"
}

function run() {
  block_time=$1
  echo "Running loadtest with block time: $block_time"
  $block_time
  update_config
  json_processor "$APP_HOME/config/genesis.json" '.genesis_time = "'"$(current_time)"'"'
  start_node
  node_catching_up "" "" || exit 1
  ./loadtest "http://127.0.0.1:26657" "./config.json" >"$APP_HOME/loadtest.log" 2>&1
  avg_block_time_interval
  kill -9 "$(cat "$APP_HOME/app.pid")"
  sleep 10
}

function run_fxcore() {
  export APP_BIN=$(which fxcored 2>/dev/null)
  export APP_HOME="$HOME/.fxcore"
  export APP_DENOM="FX"
  export APP_CHAIN_ID="fxcore"
  export APP_ADMIN_KEY_NAME="admin"
  export APP_TEST_ACCOUNTS=3000
  export APP_TEST_ACCOUNTS_DIR="$HOME/test_accounts"
  if [ "$1" == "init" ]; then
    init_genesis
    json_processor "$APP_HOME/config/genesis.json" '.consensus_params.block.max_bytes = "22020096"'
    json_processor "$APP_HOME/config/genesis.json" '.consensus_params.block.max_gas = "-1"'
    json_processor "$APP_HOME/config/genesis.json" '.consensus_params.block.time_iota_ms = "100"'
    exit 0
  fi
  for block_time in "block_timout_with_1s"; do
    run $block_time
  done
}

# shellcheck source=/dev/null
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/local_chain.sh"
