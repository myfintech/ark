#!/usr/bin/env bash
set -euo pipefail

# not dumb-init but it should do the trick
trap stop SIGTERM SIGINT SIGQUIT SIGHUP ERR

# copy pre-initialized vault data for dev vault
# only copies data once as long as a persistent volume is bound at $VAULT_DATA_VOLUME
init () {
  local data_dir
  data_dir="$(basename "$VAULT_BASE_STORAGE")"

  if [ ! -d "$VAULT_DATA_VOLUME/${data_dir}" ]; then
    echo "copying init data ${VAULT_DATA_VOLUME}"
    mkdir -p "$VAULT_DATA_VOLUME"
    cp -r "$VAULT_BASE_STORAGE" "$VAULT_DATA_VOLUME"
  else
    echo "${VAULT_DATA_VOLUME} already exists, skipping initialization"
  fi
  sed -i "s~/mnt/vault/data~${VAULT_DATA_VOLUME}/data~g" ./config.hcl 
}

# waits for vault to bind to its address specified in config.hcl
probe () {
  while ! curl --output /dev/null --silent --head --fail "${VAULT_ADDR}"; do
    echo "probing ${VAULT_ADDR}"
    sleep 2
  done
}

# called after probe, used to automatically unseal the dev vault
unseal () {
  echo "unsealing vault"
  vault operator unseal "$(jq -rj '.unseal_keys_b64[0]' < unseal.json)"
  echo "seal configuration"
  jq . < unseal.json
}

# dump-init trap and kill
# prevents hanging container on signal
stop () {
  kill "$(cat cmd_pid)"
}

start () {
  init

  vault server -config="${VAULT_CONFIG}" &
  echo -n "$!" > cmd_pid

  # self explanatory?
  probe
  unseal

  # wait for cmd_pid
  wait "$(cat cmd_pid)"
}

start "$@"
