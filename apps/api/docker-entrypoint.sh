#!/bin/sh
set -e

if [ "$1" = "api" ]; then
  migrate
  exec api
fi

exec "$@"
