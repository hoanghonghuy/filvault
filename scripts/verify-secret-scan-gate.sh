#!/usr/bin/env bash
# Verifies the secret-scan gate catches synthetic secrets and the repository stays clean.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GITLEAKS="${GITLEAKS:-gitleaks}"
if ! command -v "$GITLEAKS" >/dev/null 2>&1; then
  if [[ -x /tmp/gitleaks ]]; then
    GITLEAKS=/tmp/gitleaks
  else
    curl -sSfL https://github.com/gitleaks/gitleaks/releases/download/v8.25.0/gitleaks_8.25.0_linux_x64.tar.gz \
      | tar xz -C /tmp gitleaks
    GITLEAKS=/tmp/gitleaks
  fi
fi
if ! command -v "$GITLEAKS" >/dev/null 2>&1 && [[ -x /tmp/gitleaks ]]; then
  GITLEAKS=/tmp/gitleaks
fi
if ! command -v "$GITLEAKS" >/dev/null 2>&1; then
  echo "gitleaks binary not found; install gitleaks or set GITLEAKS=/path/to/gitleaks" >&2
  exit 1
fi

CONFIG="$ROOT/.gitleaks.toml"
CLEANUP=()
trap 'for dir in "${CLEANUP[@]:-}"; do rm -rf "$dir"; done' EXIT

make_tempdir() {
  local dir
  dir="$(mktemp -d)"
  CLEANUP+=("$dir")
  printf '%s' "$dir"
}

run_detect() {
  "$GITLEAKS" detect --source "$1" --config "$CONFIG" --no-git --no-banner --redact "${@:2}"
}

expect_leaks() {
  local label=$1
  local source=$2
  echo "==> $label: expect findings"
  if run_detect "$source" --verbose; then
    echo "scanner self-test failed: $label was not detected" >&2
    exit 1
  fi
}

expect_clean() {
  local label=$1
  local source=$2
  echo "==> $label: expect zero findings"
  run_detect "$source" --verbose
}

# 1) Custom historical-default rule (isolated temp dir; literal assembled at runtime).
HISTDIR="$(make_tempdir)"
HIST_JWT=$(printf '%s' 'ZGV2LWp3dC1zZWNyZXQ=' | base64 -d)
printf 'FILVAULT_JWT_SECRET=%s\n' "$HIST_JWT" >"$HISTDIR/selftest-historical-leak.txt"
expect_leaks "historical-default rule" "$HISTDIR/selftest-historical-leak.txt"

# 2) Default gitleaks rules on compose/Makefile-shaped fixtures in isolated temp dirs.
# Use unique fixture names (not repo docker-compose.yml / Makefile). One fixture per
# directory — gitleaks can miss findings when multiple self-test files share a directory.
COMPOSEDIR="$(make_tempdir)"
MAKEDIR="$(make_tempdir)"
# generic-api-key example from gitleaks docs; synthetic, not a real credential.
SYNTHETIC_API_KEY=$(printf '%s' 'ODdyZnVpUnlxPXZWYzNSUnJfZWRSay1mS19fSkl0cFo=' | base64 -d)
cat >"$COMPOSEDIR/selftest-compose-shaped-fixture.yml" <<EOF
services:
  api:
    environment:
      discord_client_secret: '${SYNTHETIC_API_KEY}'
EOF
printf 'FILVAULT_JWT_SECRET=%s\n' "$SYNTHETIC_API_KEY" >"$MAKEDIR/selftest-makefile-shaped-fixture"
expect_leaks "default rule in compose-shaped fixture" "$COMPOSEDIR/selftest-compose-shaped-fixture.yml"
expect_leaks "default rule in Makefile-shaped fixture" "$MAKEDIR/selftest-makefile-shaped-fixture"

# 3) Documented placeholder literals in compose/Makefile shapes remain allowed.
PLACEHOLDER_COMPOSE_DIR="$(make_tempdir)"
PLACEHOLDER_MAKE_DIR="$(make_tempdir)"
cat >"$PLACEHOLDER_COMPOSE_DIR/selftest-placeholder-compose.yml" <<'EOF'
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: ${FILVAULT_POSTGRES_PASSWORD:-replace-with-local-postgres-password}
EOF
cat >"$PLACEHOLDER_MAKE_DIR/selftest-placeholder-makefile" <<'EOF'
FILVAULT_POSTGRES_PASSWORD ?= replace-with-local-postgres-password
EOF
expect_clean "documented placeholder in compose-shaped fixture" "$PLACEHOLDER_COMPOSE_DIR/selftest-placeholder-compose.yml"
expect_clean "documented placeholder in Makefile-shaped fixture" "$PLACEHOLDER_MAKE_DIR/selftest-placeholder-makefile"

# 4) Repository must stay clean with default detectors active on tracked files.
expect_clean "repository scan" "$ROOT"

echo "secret scan gate OK"
