#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

echo "Aiki PDP-11/40 V6 emulator experiment"
echo
echo "This experiment is suspended and retained as a systems stress/regression workload."
echo "Focused diagnostics live under:"
echo "  $ROOT/diagnostics"
echo
echo "Suggested smoke gate:"
echo "  aiki test $ROOT/diagnostics/cut7_live_slice_test.ai"
