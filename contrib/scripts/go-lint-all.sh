#!/usr/bin/env bash

set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
export REPO_ROOT

for modfile in $(find . -name go.mod); do
  dir=$(dirname $modfile)
  (
    cd $dir
    echo "Linting $(grep "^module" go.mod) [$(date -Iseconds -u)]"
    golangci-lint run -v --out-format=tab ./... -c "${REPO_ROOT}/.golangci.yml" "$@"
  )
done
