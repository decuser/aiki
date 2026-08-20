#!/bin/sh
set -eu

here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo=$(CDPATH= cd -- "$here/../../../.." && pwd)
program="$here/cut7_perf_loop.ai"
aiki_cmd=${AIKI:-aiki}
out_dir=${AIKI_PDP_PERF_OUT:-"$repo/experiments/004-v6-emulator/results/cut7-perf-structural-hotpath"}

mkdir -p "$out_dir"

printf '%s\n' "Cut 7 PDP execution scaling sweep"
printf '%s\n' "realization: fixed-domain CPU/RAM + bound execution dependencies + descriptor-free CLR/CMP/branch addressing path"
printf '%s\n' "baseline: 1x = 256 CLR/CMP/BLO iterations = 768 guest instructions"
printf '%s\n' "results:  $out_dir"
printf '\n'

for scale in 1 2 10 50 100; do
	out="$out_dir/${scale}x.txt"
	printf '%s\n' "=== ${scale}x ==="
	AIKI_PDP_PERF_SCALE="$scale" "$aiki_cmd" profile --counts "$program" | tee "$out"
	printf '\n'
done

printf '%s\n' "Sweep complete. Compare elapsed, alloc_bytes, mallocs, gc_cycles,"
printf '%s\n' "and semantic counts across 1x, 2x, 10x, 50x, and 100x."
printf '%s\n' "Use these measurements to choose the next hot-path cut."
