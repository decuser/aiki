#!/bin/sh
set -eu

HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$HERE"

TMP=$(mktemp)
cleanup() {
    rm -f "$TMP"
}
trap cleanup EXIT HUP TERM

echo "Four-Way Life Gate 4 hardening acceptance"
echo

echo "--- deterministic five-process acceptance ---"
./run.sh
echo

echo "--- Engine C subprocess failure remains local ---"
if ! FWL_HELPER_FAIL=1 aiki coordinator.ai headless 8 24 18 42 >"$TMP" 2>&1; then
    cat "$TMP"
    echo "FAIL: coordinator did not survive Engine C helper failure" >&2
    exit 1
fi
if ! grep -qx 'PASS' "$TMP"; then
    cat "$TMP"
    echo "FAIL: failure-injection run did not reach PASS" >&2
    exit 1
fi
if ! grep -q '^GEN 8 ' "$TMP"; then
    cat "$TMP"
    echo "FAIL: failure-injection run did not complete requested generations" >&2
    exit 1
fi
echo "Engine C local failure recovery: PASS"
echo

echo "--- systems spine acceptance ---"
./gate3.sh
echo

echo "GATE 4 PASS"
