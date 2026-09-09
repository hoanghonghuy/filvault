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

run_detect_range() {
  local log_opts=$1
  local source=$2
  "$GITLEAKS" detect --source "$source" --config "$CONFIG" --log-opts="$log_opts" --no-banner --redact --verbose
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

expect_range_leaks() {
  local label=$1
  local source=$2
  local log_opts=$3
  echo "==> $label: expect findings ($log_opts)"
  if run_detect_range "$log_opts" "$source"; then
    echo "scanner self-test failed: $label was not detected" >&2
    exit 1
  fi
}

expect_range_clean() {
  local label=$1
  local log_opts=$2
  local source=$3
  echo "==> $label: expect zero findings ($log_opts)"
  run_detect_range "$log_opts" "$source"
}

# 1) Custom historical-default rule (isolated temp dir; literal assembled at runtime).
HISTDIR="$(make_tempdir)"
HIST_JWT=$(printf '%s' 'ZGV2LWp3dC1zZWNyZXQ=' | base64 -d)
printf 'FILVAULT_JWT_SECRET=%s\n' "$HIST_JWT" >"$HISTDIR/selftest-historical-leak.txt"
expect_leaks "historical-default rule" "$HISTDIR/selftest-historical-leak.txt"

# 2) Default gitleaks rules on compose/Makefile-shaped fixtures in isolated temp dirs.
# Fixtures live only under private temp dirs (never tracked docker-compose.yml / Makefile).
# Note: synthetic github-pat (ghp_…) values do not trip default detectors in gitleaks v8.25.0
# with this repo config (extend.useDefault = true) — CI reproduced ~97-byte fixture, zero findings.
# Use stable built-in rules instead: discord-client-secret (compose) + generic-api-key (Makefile).
COMPOSEDIR="$(make_tempdir)"
MAKEDIR="$(make_tempdir)"
# discord-client-secret example from gitleaks docs; base64 keeps the literal out of this script.
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
expect_clean "repository scan (current tree)" "$ROOT"

# 5) Commit-range self-test: synthetic leak introduced then removed must still fail the range scan.
RANGE_TEST_REPO="$(make_tempdir)"
git -C "$RANGE_TEST_REPO" init -q
git -C "$RANGE_TEST_REPO" config user.email "selftest@filvault.local"
git -C "$RANGE_TEST_REPO" config user.name "selftest"
printf 'ok\n' >"$RANGE_TEST_REPO/README"
git -C "$RANGE_TEST_REPO" add README
git -C "$RANGE_TEST_REPO" commit -q -m "baseline"
RANGE_BASE="$(git -C "$RANGE_TEST_REPO" rev-parse HEAD)"
printf 'FILVAULT_JWT_SECRET=%s\n' "$HIST_JWT" >"$RANGE_TEST_REPO/leak.txt"
git -C "$RANGE_TEST_REPO" add leak.txt
git -C "$RANGE_TEST_REPO" commit -q -m "introduce synthetic leak"
git -C "$RANGE_TEST_REPO" rm -q leak.txt
git -C "$RANGE_TEST_REPO" commit -q -m "remove leak"
expect_range_leaks "commit-range scan catches removed leak" "$RANGE_TEST_REPO" "${RANGE_BASE}..HEAD"

# 6) PR commit-range scan vs develop (only commits on this branch; allowlists cover documented placeholders).
BASE_REF="${GITLEAKS_BASE_REF:-origin/develop}"
if git rev-parse --verify "$BASE_REF" >/dev/null 2>&1; then
  COMMIT_COUNT="$(git rev-list --count "${BASE_REF}..HEAD" 2>/dev/null || echo 0)"
  if [[ "$COMMIT_COUNT" -gt 0 ]]; then
    expect_range_clean "PR commit-range scan" "${BASE_REF}..HEAD" "$ROOT"
  else
    echo "==> PR commit-range scan: skip (no commits ahead of $BASE_REF)"
  fi
else
  echo "==> PR commit-range scan: skip ($BASE_REF not found; fetch base ref for full gate)" >&2
fi

echo "secret scan gate OK"
