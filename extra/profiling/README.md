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
