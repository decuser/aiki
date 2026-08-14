# Aiki bug/drift list

7 open items, ordered easiest first.

## Priority

| # | Item | Effort |
|---|---|---|
| 1 | §6.6.3 `sqrt` example belongs to `math/ffi` only | trivial — doc |
| 2 | §6.6.3 doesn't say what `iterations`/`terms` buy | trivial — doc |
| 3 | Linter models block scope differently from the evaluator | medium |
| 4 | Canvas `toInt` truncates fractional pen_size without documenting it | small — doc |
| 5 | `turtle/simple.new(dim)` reuses a canvas at the wrong size | medium — hard to test |
| 6 | Report and *This Is Aiki* disagree on status | judgment, not code |
| 7 | Float provenance lost once an ffi result enters | decision, possibly large |
| 8 | Spawned fault strands a waiting receiver | decision, possibly large |

---

## 1. §6.6.3's worked `sqrt` example belongs to `math/ffi` only

**Drift.** The section covers both implementations but gives one example, `math.sqrt(9) ⇒ 3`. That call has one argument, so it's `math/ffi`'s, and `math/native` cannot produce it even when correct — Newton on exact rationals approaches 3 without reaching it.

**Plan.** A worked example per implementation, the native one showing an approximate result and an explicit iteration count. Generalize the rule: any section covering two implementations needs an example for each.

## 2. §6.6.3 doesn't say what `iterations` and `terms` buy

**Drift.** The Report says only that the parameters specify a count. The relationship to accuracy is neither linear nor unbounded.

**Measured** for √894, seeded at x/2:
