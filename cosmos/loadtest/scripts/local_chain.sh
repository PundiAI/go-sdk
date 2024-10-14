#!/usr/bin/env bash

set -eo pipefail

CUR_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

modify_toml() {
  file_path=$1 key=$2 value=$3
  python3 "$CUR_DIR/tq.py" "$file_path" | jq -c ".$key = \"$value\"" | xargs -0 python3 "$CUR_DIR/tq.py" "$file_path"
}

update_config() {
  modify_toml "$APP_HOME/config/client.toml" "\"keyring-backend\"" "test"

  modify_toml "$APP_HOME/config/app.toml" "\"minimum-gas-prices\"" "0$APP_DENOM"
  modify_toml "$APP_HOME/config/app.toml" "pruning" "everything"
  modify_toml "$APP_HOME/config/app.toml" "grpc.enable" "false"
  modify_toml "$APP_HOME/config/app.toml" "api.enable" "false"
  modify_toml "$APP_HOME/config/app.toml" "\"grpc-web\".enable" "false"
  modify_toml "$APP_HOME/config/app.toml" "mempool.\"max-txs\"" "10000"
  modify_toml "$APP_HOME/config/app.toml" "\"json-rpc\".enable" "false"
  modify_toml "$APP_HOME/config/app.toml" "\"json-rpc\".\"enable-indexer\"" "false"

  modify_toml "$APP_HOME/config/config.toml" "db_backend" "goleveldb"
  modify_toml "$APP_HOME/config/config.toml" "rpc.laddr" "tcp://0.0.0.0:26657"
  modify_toml "$APP_HOME/config/config.toml" "rpc.pprof_laddr" "localhost:6060"
  modify_toml "$APP_HOME/config/config.toml" "p2p.pex" "false"
  modify_toml "$APP_HOME/config/config.toml" "mempool.recheck" "false"
  modify_toml "$APP_HOME/config/config.toml" "mempool.broadcast" "false"
  modify_toml "$APP_HOME/config/config.toml" "mempool.size" "10000"
  modify_toml "$APP_HOME/config/config.toml" "mempool.max_txs_bytes" "1073741824"
  modify_toml "$APP_HOME/config/config.toml" "mempool.cache_size" "10000"
  TIMEOUT_PROPOSE=${TIMEOUT_PROPOSE:-"200ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_propose" "$TIMEOUT_PROPOSE"
  TIMEOUT_PROPOSE_DELTA=${TIMEOUT_PROPOSE_DELTA:-"50ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_propose_delta" "$TIMEOUT_PROPOSE_DELTA"
  TIMEOUT_PREVOTE=${TIMEOUT_PREVOTE:-"200ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_prevote" "$TIMEOUT_PREVOTE"
  TIMEOUT_PREVOTE_DELTA=${TIMEOUT_PREVOTE_DELTA:-"50ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_prevote_delta" "$TIMEOUT_PREVOTE_DELTA"
  TIMEOUT_PRECOMMIT=${TIMEOUT_PRECOMMIT:-"200ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_precommit" "$TIMEOUT_PRECOMMIT"
  TIMEOUT_PRECOMMIT_DELTA=${TIMEOUT_PRECOMMIT_DELTA:-"50ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_precommit_delta" "$TIMEOUT_PRECOMMIT_DELTA"
  TIMEOUT_COMMIT=${TIMEOUT_COMMIT:-"500ms"}
  modify_toml "$APP_HOME/config/config.toml" "consensus.timeout_commit" "$TIMEOUT_COMMIT"
  modify_toml "$APP_HOME/config/config.toml" "storage.discard_abci_responses" "true"
  modify_toml "$APP_HOME/config/config.toml" "tx_index.indexer" "null"
  modify_toml "$APP_HOME/config/config.toml" "instrumentation.prometheus" "true"
  modify_toml "$APP_HOME/config/config.toml" "instrumentation.namespace" "cometbft"
}

init_genesis() {
  rm -rf "$APP_HOME" "$APP_TEST_ACCOUNTS_DIR"

  $APP_BIN init "$APP_CHAIN_ID" --chain-id "$APP_CHAIN_ID" --home "$APP_HOME"

  $APP_BIN keys add "$APP_ADMIN_KEY_NAME" --keyring-backend test --output json --home "$APP_HOME" >"$APP_HOME/config/admin_key.json"

  admin_address=$($APP_BIN keys show "$APP_ADMIN_KEY_NAME" -a --keyring-backend test --home "$APP_HOME")
  if $APP_BIN --help | grep "add-genesis-account" >/dev/null; then
    $APP_BIN add-genesis-account "$admin_address" "4000000000000000000000$APP_DENOM" --home "$APP_HOME"
    python3 "$CUR_DIR/populate_genesis_accounts.py" "$APP_TEST_ACCOUNTS" "$APP_HOME/config/genesis.json"
    $APP_BIN gentx "$APP_ADMIN_KEY_NAME" "100000000000000000000$APP_DENOM" --chain-id "$APP_CHAIN_ID" --keyring-backend test --home "$APP_HOME"
    $APP_BIN collect-gentxs --home "$APP_HOME"
  else
    $APP_BIN genesis add-genesis-account "$admin_address" "1000000000000000000000$APP_DENOM" --home "$APP_HOME"
    python3 "$CUR_DIR/populate_genesis_accounts.py" "$APP_TEST_ACCOUNTS" "$APP_HOME/config/genesis.json"
    $APP_BIN genesis gentx "$APP_ADMIN_KEY_NAME" "100000000000000000000$APP_DENOM" --chain-id "$APP_CHAIN_ID" --keyring-backend test --home "$APP_HOME"
    $APP_BIN genesis collect-gentxs --home "$APP_HOME"
  fi
  #  update_config
}

start_node() {
  if $APP_BIN --help | grep "comet" >/dev/null; then
    $APP_BIN comet reset-state --home "$APP_HOME" && $APP_BIN comet unsafe-reset-all --home "$APP_HOME"
  else
    rm -rf "$APP_HOME/data" && mkdir -p "$APP_HOME/data"
    echo '{"height":"0","round":0,"step":0}' | jq >"$APP_HOME/data/priv_validator_state.json"
  fi
  if $APP_BIN start --help | grep "mempool.max-txs" >/dev/null; then
    if [ -z "$1" ]; then
      $APP_BIN start --log_level=warn --home "$APP_HOME" --mempool.max-txs -1 >"$APP_HOME/app.log" 2>&1 &
    else
      $APP_BIN start --log_level=warn --home "$APP_HOME" --mempool.max-txs -1
    fi
  else
    if [ -z "$1" ]; then
      $APP_BIN start --log_level=warn --home "$APP_HOME" >"$APP_HOME/app.log" 2>&1 &
    else
      $APP_BIN start --log_level=warn --home "$APP_HOME"
    fi
  fi
  pgrep -f "$APP_BIN" >"$APP_HOME/app.pid"
}

export APP_BIN=${APP_BIN:-$(which simd 2>/dev/null)}
export APP_HOME=${APP_HOME:-$HOME/.simapp}
export APP_DENOM=${APP_DENOM:-stake}
export APP_CHAIN_ID=${APP_CHAIN_ID:-cosmos}
export APP_ADMIN_KEY_NAME=${APP_ADMIN_KEY_NAME:-admin}
export APP_TEST_ACCOUNTS=${APP_TEST_ACCOUNTS:-3000}
export APP_TEST_ACCOUNTS_DIR=${APP_TEST_ACCOUNTS_DIR:-"$HOME/test_accounts"}

[[ "$#" -gt 0 && "$(type -t "$1")" != "function" ]] && echo "invalid args: $1" && exit 1
"$@" || (echo "failed: $0" "$@" && exit 1)
