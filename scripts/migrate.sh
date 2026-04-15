#!/bin/sh

set -eu

if [ "$#" -eq 0 ]; then
  set -- up
fi

exec go run ./cmd/migrate "$@"
