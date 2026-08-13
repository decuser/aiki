# Aiki bug/drift list

6 open items, ordered easiest first.

## Priority

| # | Item                                                         | Effort                      |
| - | ------------------------------------------------------------ | --------------------------- |
| 1 | `math/ffi` sin/cos/sqrt can build a Number with nil `Val`    | small — one function        |
| 2 | Canvas `toInt` accepts any number, truncates through float64 | small — helper plus callers |
| 3 | Linter models block scope differently from the evaluator     | medium                      |
| 4 | `turtle/simple.new(dim)` reuses a canvas at the wrong size   | medium                      |
| 5 | Float provenance lost once an ffi result enters              | decision, possibly large    |
| 6 | Spawned fault strands a waiting receiver                     | decision, possibly large    |

---

## 1. `math/ffi` sin, cos, and sqrt can build a Number with a nil `Val`

**Bug.** An out-of-range argument produces a `*value.Number` whose `Val` is nil, handed back into the evaluator.

**Cause.** All three discard the second return of `Float64()`. A rational outside float64 range yields ±Inf → `math.Sin` yields NaN → `big.Rat.SetFloat64` returns nil. Separately, `halSqrt` tests `f < 0` on the converted float rather than the rational's own `Sign()`, which cannot underflow.

**Plan.** Check the `exact` return from `Float64()` and fault when out of range; sign-check `n.Val.Sign()` before conversion. One decision to make: fault or shaped error. `math/ffi.sqrt` already faults on negatives, so faulting is consistent. Report §6.6.2–6.6.3 describe intent correctly; no change needed.

## 2. Canvas `toInt` accepts any number and truncates through float64

**Bug.** Fractional and out-of-range values are silently admitted at every canvas entry point — coordinates, dimensions, and r/g/b components of list-form colors. Non-finite or out-of-range input hits undefined behavior in the Go conversion.

**Cause.** The helper is `f, _ := n.Val.Float64(); return int(f), true` — returns `true` for every Number, with no integrality or range check. §6.11.3 documents truncation for `pen_size` only; it doesn't license it everywhere.

**Plan.** Have `toInt` return `false` on non-integral, non-finite, or out-of-range values, and let each caller produce its own diagnostic. The helper change is one function; the work is in the callers. `halSleep`'s `time.Duration(ms)` has the same shape at lower severity — fix alongside. Watch for callers that currently *rely* on truncation, `pen_size` in particular.

## 3. Linter models block scope differently from the evaluator

**Bug.** The linter and the evaluator disagree about which constructs introduce an environment.

**Cause.** The linter creates scope for ordinary blocks. Aiki semantics say `if` and `while` bodies execute in the surrounding environment; successful `match` arms are the one case that introduces an enclosed environment.

**Plan.** Align linter scope handling with evaluator semantics. Medium rather than small because the linter walks its own tree and the change may surface previously unreported diagnostics across `lib/` — expect churn. Report §5.4 and §7.2.5 are correct as written.

## 4. `turtle/simple.new(dim)` can reuse a canvas at the wrong physical size

**Bug.** `new(400)` then `new(800)` leaves an actual 400×400 canvas with state calculated for 800×800.

**Cause.** When a live canvas exists, `new(dim)` clears and reuses it rather than recreating, but still updates its logical dimension and scale as though a new physical canvas had been made.

**Plan.** Recreate the canvas when the requested dimension differs, or more simply have `new()` always establish a fresh drawing environment as documented. The fix is small. With the recording canvas now available, verification should no longer require a human display check: the transcript can assert the dimensions of successive canvas sessions.

## 5. Float provenance is unrecoverable once an ffi result enters — DECISION

**Gap.** Mix one `math/ffi` result into an otherwise native computation and nothing in the resulting number records that a float was involved.

**Cause.** `SetFloat64` produces an exact rational encoding of a binary float, indistinguishable thereafter from any other rational. The module name quarantines the call site, not the value.

**Decision.** Is call-site visibility enough, or does the numeric model need a provenance bit? A bit touches every arithmetic operation — potentially the largest change on this list.

## 6. A fault in a spawned computation strands a waiting receiver — DECISION

**Gap.** When a spawned computation is the sole producer on a channel, its fault leaves `recv` blocked permanently. A diagnosable error becomes a hang, with the diagnostic on stderr and invisible to the program.

**Cause.** Faults don't propagate out of a spawn, and channels are unbuffered with no select, timeout, or non-blocking receive. Each rule is deliberate; the composition has the sharp edge. §6.1.7 documents the behavior.

**Decision.** Does the concurrency surface need a means of bounded waiting? Adding one changes the four-function concurrency surface the language advertises, so it's a design question rather than a fix. It also constrains anything built on Aiki concurrency — an interruptible simulated machine cannot poll a control channel between steps.

