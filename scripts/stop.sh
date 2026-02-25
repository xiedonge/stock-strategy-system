#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
RUN_DIR="$ROOT_DIR/.run"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_PORT=${PORT:-8080}
FRONTEND_PORT=${FRONTEND_PORT:-5173}

get_listen_pids() {
  local port="$1"
  local pids=""
  if command -v lsof >/dev/null 2>&1; then
    pids=$(lsof -ti tcp:"$port" -sTCP:LISTEN 2>/dev/null || true)
  elif command -v ss >/dev/null 2>&1; then
    pids=$(ss -ltnp "sport = :$port" 2>/dev/null | awk 'NR>1 {for(i=1;i<=NF;i++){if($i ~ /pid=/){gsub(/.*pid=/,"",$i); gsub(/,.*/,"",$i); print $i}}}' | sort -u)
  elif command -v fuser >/dev/null 2>&1; then
    pids=$(fuser -n tcp "$port" 2>/dev/null | tr ' ' '\n' | sort -u)
  fi
  echo "$pids"
}

is_project_pid() {
  local pid="$1"
  local cmdline
  cmdline=$(ps -p "$pid" -o command= 2>/dev/null || true)
  if [[ -z "$cmdline" ]]; then
    return 1
  fi
  if [[ "$cmdline" == *"$ROOT_DIR"* || "$cmdline" == *"$BACKEND_DIR"* || "$cmdline" == *"$FRONTEND_DIR"* ]]; then
    return 0
  fi
  return 1
}

stop_port() {
  local port="$1"
  local label="$2"
  local pids
  pids=$(get_listen_pids "$port")
  if [[ -z "$pids" ]]; then
    return 0
  fi
  for pid in $pids; do
    if is_project_pid "$pid" || [[ "${FORCE_KILL_PORT:-}" == "1" ]]; then
      echo "Stopping $label on port $port (pid $pid)"
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
    else
      echo "Port $port is in use by pid $pid (not detected as project)."
      echo "Set FORCE_KILL_PORT=1 to kill it, or change PORT/FRONTEND_PORT."
      exit 1
    fi
  done
}

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
stop_port "$BACKEND_PORT" "backend"
stop_port "$FRONTEND_PORT" "frontend"

if [[ -d "$RUN_DIR" ]]; then
  echo "Stopped. Logs are kept in $RUN_DIR."
else
  echo "No running services detected."
fi
