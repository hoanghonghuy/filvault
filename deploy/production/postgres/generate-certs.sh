#!/usr/bin/env bash
# Generate self-signed TLS material for the production-like Postgres service.
# Required before first `make prod-like-up` — production validation rejects sslmode=disable.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="$ROOT/certs"
mkdir -p "$CERT_DIR"

if [[ -f "$CERT_DIR/server.crt" && -f "$CERT_DIR/server.key" ]]; then
  echo "Postgres TLS certs already exist in $CERT_DIR"
  exit 0
fi

openssl req -new -x509 -days 3650 -nodes -text \
  -out "$CERT_DIR/server.crt" \
  -keyout "$CERT_DIR/server.key" \
  -subj "/CN=filvault-postgres"
chmod 600 "$CERT_DIR/server.key"
echo "Generated Postgres TLS certs in $CERT_DIR"
