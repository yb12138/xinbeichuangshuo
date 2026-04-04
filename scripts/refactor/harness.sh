#!/usr/bin/env bash
set -euo pipefail

profile="${1:-smoke}"
run_dir="${2:-}"

if [ -n "$run_dir" ]; then
  mkdir -p "${run_dir}/logs"
fi

timestamp() {
  date '+%Y-%m-%d %H:%M:%S'
}

run_step() {
  local name="$1"
  local cmd="$2"
  local logfile=""
  if [ -n "$run_dir" ]; then
    logfile="${run_dir}/logs/${name}.log"
  fi

  echo "[$(timestamp)] >>> ${name}"
  echo "[$(timestamp)] CMD: ${cmd}"
  if [ -n "$logfile" ]; then
    if bash -lc "$cmd" >"${logfile}" 2>&1; then
      echo "[$(timestamp)] PASS ${name}"
    else
      echo "[$(timestamp)] FAIL ${name} (see ${logfile})"
      return 1
    fi
  else
    if bash -lc "$cmd"; then
      echo "[$(timestamp)] PASS ${name}"
    else
      echo "[$(timestamp)] FAIL ${name}"
      return 1
    fi
  fi
}

case "$profile" in
  compile)
    run_step "compile" "go test ./... -run '^$'"
    ;;
  smoke)
    run_step "compile" "go test ./... -run '^$'"
    run_step "tests" "go test ./tests/... -count=1"
    ;;
  engine)
    run_step "engine" "go test ./internal/engine/... -count=1"
    ;;
  server)
    run_step "server" "go test ./internal/server/... -count=1"
    ;;
  full)
    run_step "compile" "go test ./... -run '^$'"
    run_step "tests" "go test ./tests/... -count=1"
    run_step "engine" "go test ./internal/engine/... -count=1"
    run_step "server" "go test ./internal/server/... -count=1"
    run_step "rules" "go test ./internal/rules/... -count=1"
    ;;
  *)
    echo "Unknown profile: ${profile}"
    echo "Available profiles: compile | smoke | engine | server | full"
    exit 1
    ;;
esac

echo "Harness profile '${profile}' completed."
