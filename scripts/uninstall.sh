#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
RUN_DIR="$ROOT_DIR/.run"

"$ROOT_DIR/scripts/stop.sh"

# Remove generated artifacts.
rm -rf "$ROOT_DIR/backend/data" 2>/dev/null || true
rm -rf "$ROOT_DIR/frontend/node_modules" 2>/dev/null || true
rm -rf "$ROOT_DIR/frontend/dist" 2>/dev/null || true
rm -rf "$RUN_DIR" 2>/dev/null || true

echo "Uninstall complete. Generated data and dependencies removed."
