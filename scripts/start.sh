#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
RUN_DIR="$ROOT_DIR/.run"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

mkdir -p "$RUN_DIR"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

run_as_root() {
  if [[ $(id -u) -eq 0 ]]; then
    "$@"
    return
  fi
  if command -v sudo >/dev/null 2>&1; then
    sudo "$@"
    return
  fi
  echo "Need root privileges to install dependencies. Please run with sudo." >&2
  return 1
}

detect_pkg_manager() {
  if command -v apt-get >/dev/null 2>&1; then
    echo "apt"
    return
  fi
  if command -v dnf >/dev/null 2>&1; then
    echo "dnf"
    return
  fi
  if command -v yum >/dev/null 2>&1; then
    echo "yum"
    return
  fi
  if command -v pacman >/dev/null 2>&1; then
    echo "pacman"
    return
  fi
  if command -v apk >/dev/null 2>&1; then
    echo "apk"
    return
  fi
  if command -v brew >/dev/null 2>&1; then
    echo "brew"
    return
  fi
  return 1
}

install_packages() {
  local pm
  pm=$(detect_pkg_manager) || return 1
  case "$pm" in
    apt)
      run_as_root apt-get update
      run_as_root apt-get install -y "$@"
      ;;
    dnf)
      run_as_root dnf install -y "$@"
      ;;
    yum)
      run_as_root yum install -y "$@"
      ;;
    pacman)
      run_as_root pacman -Sy --noconfirm "$@"
      ;;
    apk)
      run_as_root apk add --no-cache "$@"
      ;;
    brew)
      brew install "$@"
      ;;
    *)
      return 1
      ;;
  esac
}

ensure_command() {
  local cmd="$1"
  shift
  if command -v "$cmd" >/dev/null 2>&1; then
    return 0
  fi
  echo "Missing $cmd. Attempting to install..."
  "$@" || {
    echo "Auto-install failed for $cmd. Please install it manually." >&2
    exit 1
  }
  require_cmd "$cmd"
}

install_go() {
  local pm
  pm=$(detect_pkg_manager) || return 1
  case "$pm" in
    apt)
      install_packages golang-go
      ;;
    dnf|yum)
      install_packages golang
      ;;
    pacman|apk)
      install_packages go
      ;;
    brew)
      install_packages go
      ;;
    *)
      return 1
      ;;
  esac
}

install_node() {
  local pm
  pm=$(detect_pkg_manager) || return 1
  case "$pm" in
    apt|dnf|yum|pacman|apk)
      install_packages nodejs npm
      ;;
    brew)
      install_packages node
      ;;
    *)
      return 1
      ;;
  esac
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

ensure_command go install_go
ensure_command npm install_node

# Stop previous runs if any.
stop_pidfile "$RUN_DIR/backend.pid"
stop_pidfile "$RUN_DIR/frontend.pid"

# Backend
BACKEND_PORT=${PORT:-8080}
export PORT="$BACKEND_PORT"

if [[ ! -f "$BACKEND_DIR/go.mod" ]]; then
  echo "Backend not found at $BACKEND_DIR" >&2
  exit 1
fi

nohup bash -lc "cd '$BACKEND_DIR' && go run ./cmd/server" >"$RUN_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo "$BACKEND_PID" > "$RUN_DIR/backend.pid"

echo "Backend started on :$BACKEND_PORT (pid $BACKEND_PID)"

# Frontend
FRONTEND_PORT=${FRONTEND_PORT:-5173}

if [[ ! -f "$FRONTEND_DIR/package.json" ]]; then
  echo "Frontend not found at $FRONTEND_DIR" >&2
  exit 1
fi

if [[ ! -d "$FRONTEND_DIR/node_modules" ]]; then
  echo "Installing frontend dependencies..."
  (cd "$FRONTEND_DIR" && npm install)
fi

nohup bash -lc "cd '$FRONTEND_DIR' && npm run dev -- --host 0.0.0.0 --port $FRONTEND_PORT" >"$RUN_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "$FRONTEND_PID" > "$RUN_DIR/frontend.pid"

echo "Frontend started on :$FRONTEND_PORT (pid $FRONTEND_PID)"

echo "Logs:"
echo "  Backend:  $RUN_DIR/backend.log"
echo "  Frontend: $RUN_DIR/frontend.log"
