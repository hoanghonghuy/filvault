#!/usr/bin/env bash
# Generate self-signed TLS material for the production-like Postgres service.
# Required before first `make prod-like-up` — production validation rejects sslmode=disable.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="$ROOT/certs"
mkdir -p "$CERT_DIR"

# postgres:16-alpine uses Alpine's standard postgres uid/gid 70 (debian-based images use 999).
POSTGRES_UID=70
POSTGRES_GID=70

fix_cert_ownership() {
  if [[ ! -f "$CERT_DIR/server.crt" || ! -f "$CERT_DIR/server.key" ]]; then
    return 0
  fi
  if [[ $(id -u) -eq 0 ]]; then
    chown "${POSTGRES_UID}:${POSTGRES_GID}" "$CERT_DIR/server.crt" "$CERT_DIR/server.key"
  elif command -v sudo >/dev/null 2>&1; then
    sudo chown "${POSTGRES_UID}:${POSTGRES_GID}" "$CERT_DIR/server.crt" "$CERT_DIR/server.key"
  else
    echo "Cannot chown certs to ${POSTGRES_UID}:${POSTGRES_GID}; re-run as root or with sudo" >&2
    exit 1
  fi
  chmod 644 "$CERT_DIR/server.crt"
  chmod 600 "$CERT_DIR/server.key"
}

if [[ -f "$CERT_DIR/server.crt" && -f "$CERT_DIR/server.key" ]]; then
  fix_cert_ownership
  echo "Postgres TLS certs already exist in $CERT_DIR (ownership refreshed for uid ${POSTGRES_UID})"
  exit 0
fi

if command -v docker >/dev/null 2>&1; then
  docker run --rm \
    -v "$CERT_DIR:/certs" \
    -u "${POSTGRES_UID}:${POSTGRES_GID}" \
    postgres:16-alpine \
    sh -c 'openssl req -new -x509 -days 3650 -nodes \
      -out /certs/server.crt -keyout /certs/server.key \
      -subj "/CN=filvault-postgres" && chmod 600 /certs/server.key'
else
  openssl req -new -x509 -days 3650 -nodes -text \
    -out "$CERT_DIR/server.crt" \
    -keyout "$CERT_DIR/server.key" \
    -subj "/CN=filvault-postgres"
  fix_cert_ownership
fi

echo "Generated Postgres TLS certs in $CERT_DIR (owner uid ${POSTGRES_UID})"
