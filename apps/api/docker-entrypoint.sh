#!/bin/sh
set -e

if [ "$1" = "api" ]; then
  migrate
  if [ "$FILVAULT_SEED_DEV_USER" = "true" ]; then
    seed
  fi
  exec api
fi

exec "$@"
