#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
RUN_DIR="$ROOT_DIR/.run"
VENV_DIR="$RUN_DIR/akshare-venv"
PYTHON_BIN="$VENV_DIR/bin/python"

mkdir -p "$RUN_DIR"

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

ensure_python() {
  if command -v python3 >/dev/null 2>&1; then
    return 0
  fi
  echo "Missing python3. Attempting to install..."
  local pm
  pm=$(detect_pkg_manager) || {
    echo "No supported package manager found. Install python3 manually." >&2
    exit 1
  }
  case "$pm" in
    apt)
      install_packages python3 python3-venv
      ;;
    dnf|yum)
      install_packages python3
      ;;
    pacman)
      install_packages python
      ;;
    apk)
      install_packages python3
      ;;
    brew)
      install_packages python
      ;;
    *)
      echo "Unsupported package manager. Install python3 manually." >&2
      exit 1
      ;;
  esac
}

ensure_pip() {
  if python3 -m pip --version >/dev/null 2>&1; then
    return 0
  fi
  echo "Missing pip. Attempting to install..."
  local pm
  pm=$(detect_pkg_manager) || {
    echo "No supported package manager found. Install pip manually." >&2
    exit 1
  }
  case "$pm" in
    apt|dnf|yum)
      install_packages python3-pip
      ;;
    pacman)
      install_packages python-pip
      ;;
    apk)
      install_packages py3-pip
      ;;
    brew)
      return 0
      ;;
    *)
      echo "Unsupported package manager. Install pip manually." >&2
      exit 1
      ;;
  esac
}

ensure_python
ensure_pip

if [[ ! -x "$PYTHON_BIN" ]]; then
  if ! python3 -m venv "$VENV_DIR"; then
    echo "Failed to create venv. Please ensure python3-venv is installed." >&2
    exit 1
  fi
fi

"$PYTHON_BIN" -m pip install --upgrade pip >/dev/null 2>&1 || true
if ! "$PYTHON_BIN" -c "import akshare" >/dev/null 2>&1; then
  echo "Installing akshare..."
  "$PYTHON_BIN" -m pip install --upgrade akshare
fi

exec "$PYTHON_BIN" "$ROOT_DIR/scripts/akshare_sync.py" "$@"
