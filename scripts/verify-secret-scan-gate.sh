#!/usr/bin/env bash
# Verifies the secret-scan gate catches synthetic secrets and the repository stays clean.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GITLEAKS="${GITLEAKS:-gitleaks}"
if ! command -v "$GITLEAKS" >/dev/null 2>&1; then
  if [[ -x /tmp/gitleaks ]]; then
    GITLEAKS=/tmp/gitleaks
  elif ! command -v "$GITLEAKS" >/dev/null 2>&1; then
    curl -sSfL https://github.com/gitleaks/gitleaks/releases/download/v8.25.0/gitleaks_8.25.0_linux_x64.tar.gz \
      | tar xz -C /tmp gitleaks
    GITLEAKS=/tmp/gitleaks
  fi
  if ! command -v "$GITLEAKS" >/dev/null 2>&1 && [[ ! -x /tmp/gitleaks ]]; then
    echo "gitleaks binary not found; install gitleaks or set GITLEAKS=/path/to/gitleaks" >&2
    exit 1
  fi
fi

CONFIG="$ROOT/.gitleaks.toml"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

run_detect() {
  "$GITLEAKS" detect --source "$1" --config "$CONFIG" --no-git --no-banner --redact "$@"
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

# 1) Custom historical-default rule (avoid embedding literal in this script).
HIST_JWT=$(printf '%s' 'ZGV2LWp3dC1zZWNyZXQ=' | base64 -d)
printf 'FILVAULT_JWT_SECRET=%s\n' "$HIST_JWT" >"$TMPDIR/historical-leak.txt"
expect_leaks "historical-default rule" "$TMPDIR/historical-leak.txt"

# 2) Default gitleaks rule in formerly path-allowlisted sensitive file shapes.
# Assemble synthetic github-pat shape at runtime so the committed script stays clean.
PAT_SUFFIX=$(printf '%s' '1234567890abcdefghijklmnopqrstuvwxyz12')
PAT_VALUE=$(printf 'ghp_%s' "$PAT_SUFFIX")
printf 'services:\n  api:\n    environment:\n      GITHUB_TOKEN: %s\n' "$PAT_VALUE" >"$TMPDIR/docker-compose.yml"
printf 'FILVAULT_JWT_SECRET=%s\n' "$PAT_VALUE" >"$TMPDIR/Makefile"
expect_leaks "default rule in docker-compose.yml" "$TMPDIR/docker-compose.yml"
expect_leaks "default rule in Makefile" "$TMPDIR/Makefile"

# 3) Documented placeholder literals in compose/Makefile shapes remain allowed.
cat >"$TMPDIR/placeholder-compose.yml" <<'EOF'
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: ${FILVAULT_POSTGRES_PASSWORD:-replace-with-local-postgres-password}
EOF
cat >"$TMPDIR/placeholder-makefile" <<'EOF'
FILVAULT_POSTGRES_PASSWORD ?= replace-with-local-postgres-password
EOF
expect_clean "documented placeholder in compose" "$TMPDIR/placeholder-compose.yml"
expect_clean "documented placeholder in Makefile" "$TMPDIR/placeholder-makefile"

# 4) Repository must stay clean with default detectors active on tracked files.
expect_clean "repository scan" "$ROOT"

echo "secret scan gate OK"
