#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
RUN_DIR="$ROOT_DIR/.run"

is_running() {
  local pid="$1"
  if [[ -z "$pid" ]]; then
    return 1
  fi
  kill -0 "$pid" >/dev/null 2>&1
}

stop_pidfile() {
  local pidfile="$1"
  if [[ ! -f "$pidfile" ]]; then
    return 0
  fi

  local pid
  pid=$(cat "$pidfile" 2>/dev/null || true)
  if is_running "$pid"; then
    echo "Stopping process $pid (from $pidfile)"
    kill "$pid" >/dev/null 2>&1 || true
    for _ in {1..10}; do
      if ! is_running "$pid"; then
        break
      fi
      sleep 0.3
    done
    if is_running "$pid"; then
      echo "Process $pid did not stop, forcing shutdown"
      kill -9 "$pid" >/dev/null 2>&1 || true
    fi
  fi
  rm -f "$pidfile"
}

stop_pidfile "$RUN_DIR/backend.pid"
stop_pidfile "$RUN_DIR/frontend.pid"

if [[ -d "$RUN_DIR" ]]; then
  echo "Stopped. Logs are kept in $RUN_DIR."
else
  echo "No running services detected."
fi
