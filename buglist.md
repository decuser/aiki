# Aiki bug/drift list

11 open items, ordered easiest first.

## Priority

| # | Item | Effort |
|---|---|---|
| 1 | §6.6.3 `sqrt` example belongs to `math/ffi` only | trivial — doc |
| 2 | §6.6.3 doesn't say what `iterations`/`terms` buy | trivial — doc |
| 3 | `math/ffi` sin/cos/sqrt can build a Number with nil `Val` | small — one function |
| 4 | Canvas `toInt` accepts any number, truncates through float64 | small — helper plus callers |
| 5 | Linter models block scope differently from the evaluator | medium |
| 6 | `to_decimal` renders exact rationals through float64 | medium — needs real tests |
| 7 | `turtle/simple.new(dim)` reuses a canvas at the wrong size | medium — hard to test |
| 8 | Report and *This Is Aiki* disagree on status | judgment, not code |
| 9 | Regex byte offsets vs. rune string indexing | decision, then medium |
| 10 | Float provenance lost once an ffi result enters | decision, possibly large |
| 11 | Spawned fault strands a waiting receiver | decision, possibly large |

---

## 1. §6.6.3's worked `sqrt` example belongs to `math/ffi` only

**Drift.** The section covers both implementations but gives one example, `math.sqrt(9) ⇒ 3`. That call has one argument, so it's `math/ffi`'s, and `math/native` cannot produce it even when correct — Newton on exact rationals approaches 3 without reaching it.

**Consequence.** The Report's only worked example of `sqrt` exercised the implementation that wasn't broken, which is how the recurrence bug survived a document review.

**Plan.** A worked example per implementation, the native one showing an approximate result and an explicit iteration count. Generalize the rule: any section covering two implementations needs an example for each.

## 2. §6.6.3 doesn't say what `iterations` and `terms` buy

**Drift.** The Report says only that the parameters specify a count. The relationship to accuracy is neither linear nor unbounded, and the useful count depends on the argument's magnitude, since the seed is x/2 — √9 converges by 5 iterations, √894 by 8.

**Measured** for √894, seeded at x/2:

```
iteration 5    30.7330            42 denominator digits
iteration 6    29.9111            85
iteration 7    29.8998349        172
iteration 8    29.8998327754     345
iteration 9    unchanged         692
```

**Plan.** State that the parameter bounds work rather than error, that accuracy stops improving once the iteration converges, and that the representation keeps growing regardless. Pair with item 1 — same section, one editing pass.

## 3. `math/ffi` sin, cos, and sqrt can build a Number with a nil `Val`

**Bug.** An out-of-range argument produces a `*value.Number` whose `Val` is nil, handed back into the evaluator.

**Cause.** All three discard the second return of `Float64()`. A rational outside float64 range yields ±Inf → `math.Sin` yields NaN → `big.Rat.SetFloat64` returns nil. Separately, `halSqrt` tests `f < 0` on the converted float rather than the rational's own `Sign()`, which cannot underflow.

**Plan.** Check the `exact` return from `Float64()` and fault when out of range; sign-check `n.Val.Sign()` before conversion. One decision to make: fault or shaped error. `math/ffi.sqrt` already faults on negatives, so faulting is consistent. Report §6.6.2–6.6.3 describe intent correctly; no change needed.

## 4. Canvas `toInt` accepts any number and truncates through float64

**Bug.** Fractional and out-of-range values are silently admitted at every canvas entry point — coordinates, dimensions, and r/g/b components of list-form colors. Non-finite or out-of-range input hits undefined behavior in the Go conversion.

**Cause.** The helper is `f, _ := n.Val.Float64(); return int(f), true` — returns `true` for every Number, with no integrality or range check. §6.11.3 documents truncation for `pen_size` only; it doesn't license it everywhere.

**Plan.** Have `toInt` return `false` on non-integral, non-finite, or out-of-range values, and let each caller produce its own diagnostic. The helper change is one function; the work is in the callers. `halSleep`'s `time.Duration(ms)` has the same shape at lower severity — fix alongside. Watch for callers that currently *rely* on truncation, `pen_size` in particular.

## 5. Linter models block scope differently from the evaluator

**Bug.** The linter and the evaluator disagree about which constructs introduce an environment.

**Cause.** The linter creates scope for ordinary blocks. Aiki semantics say `if` and `while` bodies execute in the surrounding environment; successful `match` arms are the one case that introduces an enclosed environment.

**Plan.** Align linter scope handling with evaluator semantics. Medium rather than small because the linter walks its own tree and the change may surface previously-unreported diagnostics across `lib/` — expect churn. Report §5.4 and §7.2.5 are correct as written.

## 6. `to_decimal` renders exact rationals through float64

**Bug.** Every number printed with `to_decimal` passes through a binary float before display. At small magnitudes the digits are right; at large magnitudes, high place counts, or denominators outside float64's reach, they are not.

**Cause.** `halToDecimal` does `f, _ := n.Val.Float64()` then `%.*f`. This is the display path for the language's central claim, and it's used throughout the Report's own Example — including the published `0.000036` residual.

**Plan.** Format from the rational directly: long-divide numerator by denominator to `p` places, rounding on the exact remainder. The code is not hard; the tests are the work — you need cases at magnitudes and place counts where float64 currently lies, and negative values and exact halves for rounding. Then state in Report §6.1.5 that `to_decimal` rounds the exact value. Highest value on the list despite the position.

## 7. `turtle/simple.new(dim)` can reuse a canvas at the wrong physical size

**Bug.** `new(400)` then `new(800)` leaves an actual 400×400 canvas with state calculated for 800×800.

**Cause.** When a live canvas exists, `new(dim)` clears and reuses it rather than recreating, but still updates its logical dimension and scale as though a new physical canvas had been made.

**Plan.** Recreate the canvas when the requested dimension differs, or more simply have `new()` always establish a fresh drawing environment as documented. The fix is small; verification is the hard part, since canvas smoke tests need a human at the screen and the ebiten one-window-per-process constraint shapes what's possible.

## 8. The two documents disagree on status

**Drift.** `ra1` presents Aiki 1 as a defining specification whose features are intended to remain stable within the report. *This Is Aiki* describes the same surface as alpha and asks which parts should be frozen for beta.

**Plan.** No code, no research — but it's a real decision about what you're claiming, and the remaining three items partly depend on it. Both framings are defensible; holding both is not. Reconciling them is a prerequisite for the Report functioning as a specification others can implement against.

## 9. Regex byte offsets vs. rune string indexing — DECISION

**Gap.** A match position cannot be used to index the string it came from. Invisible on ASCII, silently wrong on anything else.

**Cause.** §6.8.2 specifies `[@match, start, end, text]` positions as byte offsets into UTF-8; §4.9 specifies `s[i]` indexes by rune, and the string builtins convert through `[]rune`. Both documents state their halves correctly, so the language as specified is inconsistent — a decision, not a patch.

**Options.** Return rune offsets from `regex/ffi`, converting at the boundary; or keep byte offsets and add explicit conversion procedures to the `string` module; or make `@match` carry both. The first is most consistent with the rest of the language and the least work once decided.

## 10. Float provenance is unrecoverable once an ffi result enters — DECISION

**Gap.** Mix one `math/ffi` result into an otherwise native computation and nothing in the resulting number records that a float was involved.

**Cause.** `SetFloat64` produces an exact rational encoding of a binary float, indistinguishable thereafter from any other rational. The module name quarantines the call site, not the value.

**Decision.** Is call-site visibility enough, or does the numeric model need a provenance bit? A bit touches every arithmetic operation — potentially the largest change on this list. Cheaper to decide alongside item 9; both concern a value carrying less information than the reader assumes.

## 11. A fault in a spawned computation strands a waiting receiver — DECISION

**Gap.** When a spawned computation is the sole producer on a channel, its fault leaves `recv` blocked permanently. A diagnosable error becomes a hang, with the diagnostic on stderr and invisible to the program.

**Cause.** Faults don't propagate out of a spawn, and channels are unbuffered with no select, timeout, or non-blocking receive. Each rule is deliberate; the composition has the sharp edge. §6.1.7 now documents the behavior.

**Decision.** Does the concurrency surface need a means of bounded waiting? Adding one changes the four-function concurrency surface the language advertises, so it's a design question rather than a fix. It also constrains anything built on Aiki concurrency — an interruptible simulated machine cannot poll a control channel between steps.

---

## Closed

| Item | Resolution |
|---|---|
| `math/native.sqrt` wrong Newton recurrence | fixed — parenthesization |
| `math/native.sqrt` two-element error value | fixed — `[@error, :math, …]` |
| Shape definitions didn't cross the spawn boundary | fixed — `Env.CollectShapes()` |
| `apply` bound missing arguments to `[]` | fixed — minimum-arity rule |
| `spawn` had the same inconsistency | fixed — same rule, isolated path |
| Pattern matcher omitted rune and boolean literals | fixed |
| Formatter dropped rune patterns, producing invalid source | fixed |
| `bytes/native` admitted non-integral values | fixed — integrality check |
| Two-element `@error` values across `lib/` | fixed — six sites, all in `bytes/native` |
| Unparenthesized mixed-operator arithmetic in `lib/` | audited — `sqrt` was the only instance |
| `random` help/doc described an obsolete interface | fixed |
| `math/ffi.doc` claimed `sqrt` returns an error | fixed — it faults |
| Both hash docs named `[@error, :not_found]` | fixed — actual is `[@error, :key, …]` |
| Bare `[@error]` in list/file/regex/hash help | fixed — full three-element form |
| `constraints.md` said eight types | fixed — seven basic, five system |
| Report Example: native-vs-ffi rationale, unused bindings | fixed |
