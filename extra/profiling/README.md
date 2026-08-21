# Profiling workloads

Small, durable workloads intended for `aiki profile --counts`.

These are not ordinary language examples. They exist to pressure particular
runtime paths and make semantic work, Number realization, allocation, and
substrate costs visible.

- `selfhost-three-level.ai` — host interpreter -> selfhost -> selfhost -> `1 + 2 * 3`
- `adaptive-number-rational.ai` — compact exact rational arithmetic
- `adaptive-number-host-math.ai` — host `math/ffi` binary64-return boundary

## Adaptive Number acceptance witnesses

The three adaptive-number workloads are retained as durable profiling witnesses:

- `selfhost-three-level.ai` demonstrates small-integer dominance through three interpreter levels;
- `adaptive-number-rational.ai` demonstrates compact-rational coverage with no arbitrary-precision promotion;
- `adaptive-number-host-math.ai` separates host call-return carriers from arithmetic realization and exercises certified versus fallback binary arithmetic.

Acceptance measurements are recorded in `docs/adaptive-number-results.md`.

## Alpha-36 runtime realization survey

`runtime-realization-survey.sh` captures the allocation-space and
allocation-object views used by `proposals/runtime-realization-survey.md`. It
runs the three-level self-host workload and the PDP-11 Cut 7 10x comparison
witness, writing ordinary Go allocation profiles plus flat/cumulative pprof
tables.

This survey is attribution-only: it selects the next semantic realization
family from evidence and does not presume that strings, AST values, or another
family require an adaptive representation.


### Immutable string observation final gate

Run the complete correctness/performance gate without leaving generated output
inside the repository:

```sh
extra/profiling/string-observation-gate.sh
```

The default evidence directory and archive are:

```text
/tmp/aiki-string-observation-gate/
/tmp/aiki-string-observation-gate.tar.gz
```

The gate runs `make validate`, the focused mixed-width Unicode indexing witness,
and the post-fix self-host/PDP allocation survey.
