#!/bin/sh
set -eu

AIKI=${AIKI:-./aiki}
OUT=${1:-/tmp/aiki-string-observation-gate}
SURVEY="$OUT/runtime-realization-survey"
ARCHIVE="${OUT}.tar.gz"

rm -rf "$OUT" "$ARCHIVE"
mkdir -p "$OUT"

echo "== validate =="
make validate 2>&1 | tee "$OUT/validate.txt"

echo "== focused string witness =="
"$AIKI" profile --counts extra/profiling/string-observation.ai \
    2>&1 | tee "$OUT/string-observation.counts.txt"

echo "== post-fix runtime realization survey =="
extra/profiling/runtime-realization-survey.sh "$SURVEY" \
    2>&1 | tee "$OUT/survey-run.txt"

{
    echo "generated_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "git=$(git describe --tags --always --dirty 2>/dev/null || echo unavailable)"
    echo "head=$(git rev-parse HEAD 2>/dev/null || echo unavailable)"
    echo "aiki=$($AIKI version 2>/dev/null || echo unknown)"
    echo "go=$(go version 2>/dev/null || echo unavailable)"
} > "$OUT/gate-manifest.txt"

tar -czf "$ARCHIVE" -C "$(dirname "$OUT")" "$(basename "$OUT")"

echo
echo "gate evidence: $OUT"
echo "archive: $ARCHIVE"
