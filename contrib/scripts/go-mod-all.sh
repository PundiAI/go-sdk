#!/usr/bin/env bash

set -euo pipefail

for modfile in $(find . -name go.mod); do
  dir=$(dirname $modfile)
  (
    cd $dir;
    echo "Updating $(grep "^module" go.mod) [$(date -Iseconds -u)]"
    go mod "$@"
  )
done
