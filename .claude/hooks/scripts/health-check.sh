#!/bin/bash
# Health check for Langfuse Hook
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAIL=0

check() { # $1=condition, $2=ok_msg, $3=error_msg
  if eval "$1"; then echo "[OK] $2"; else echo "[ERROR] $3"; ((FAIL++)) || true; fi
}

echo "=== Langfuse Hook Health Check ==="
check '[ -n "$LANGFUSE_PUBLIC_KEY" ]' "LANGFUSE_PUBLIC_KEY set (${LANGFUSE_PUBLIC_KEY:0:8}...)" "LANGFUSE_PUBLIC_KEY not set"
check '[ -n "$LANGFUSE_SECRET_KEY" ]' "LANGFUSE_SECRET_KEY set (${LANGFUSE_SECRET_KEY:0:8}...)" "LANGFUSE_SECRET_KEY not set"
check '[ -f "$DIR/dist/langfuse-hook.js" ]' "Hook compiled" "Hook not compiled - run: npm run build"
check '[ -d "$DIR/node_modules/langfuse" ]' "langfuse installed" "langfuse not installed - run: npm install"

echo
[ $FAIL -eq 0 ] && echo "[SUCCESS] All checks passed" || { echo "[FAILED] $FAIL check(s) failed"; exit 1; }
