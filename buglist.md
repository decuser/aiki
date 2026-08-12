Low-hanging first, by effort — not by importance.

Trivial, minutes each, no thinking required

#20 IDENTITY eight types → seven. One word.
#19 random help/doc. Copy the current signature in. Metadata only.
#3 grep lib/ for two-element @error. Mechanical; may find nothing.
#2 grep lib/ for unparenthesized mixed-operator arithmetic. Also mechanical, but read the hits — this one can find something real.

Small and contained, one sitting each

#9 + #11 together. Rune and boolean literals in matcher and formatter — same defect, two files, and the formatter fix is the same dispatch addition.
#7 + #8 together. Minimum arity in _apply and applyUserFunctionIsolated. You've just been in that function; the second is a three-line change. Note both will break any test that relied on the [] fallback, so run make validate expecting churn.
#6 bytes_new integrality check. One predicate.
#5 canvas toInt. Slightly larger since callers need diagnostics, but the helper change itself is small.

Real work

#1 to_decimal. Needs correct long division with rounding on the exact remainder, plus tests at magnitudes and place counts where float64 currently lies. Highest value on the list, but it isn't fruit.
#4 math/ffi nil Val. Small code, but you have to decide fault vs. shaped error to match the module's convention.
#12 linter block scope. Needs care to match evaluator semantics exactly.
#10 turtle/simple.new. Canvas lifecycle, hardest to test, and the smoke tests need a human at the screen.

Not fruit at all — leave

#13, #14, #15 are decisions. #16, #17, #18 are Report writing.

If you want one hour that pays: items 1–4 clear four entries, and #2 is the one that could surface another live bug.

# Aiki known bug/drift list

20 open items. Numbered in order of appearance.

## Summary

| # | Item | Kind |
|---|---|---|
| 1 | `to_decimal` renders exact rationals through float64 | bug — highest priority |
| 2 | Audit `lib/` for unparenthesized mixed-operator arithmetic | audit |
| 3 | Audit `lib/` for two-element `@error` values | audit |
| 4 | `math/ffi` sin/cos/sqrt can construct a Number with nil `Val` | bug |
| 5 | Canvas `toInt` accepts any number, truncates through float64 | bug |
| 6 | `bytes/native` admits non-integral values in range | bug |
| 7 | `apply` binds missing arguments to `[]` instead of faulting | bug |
| 8 | `spawn` has the same missing-argument inconsistency | bug |
| 9 | Pattern matcher omits rune and boolean literals | bug |
| 10 | `turtle/simple.new(dim)` reuses a canvas at the wrong physical size | bug |
| 11 | Formatter omits rune literal patterns | tooling |
| 12 | Linter models block scope differently from the evaluator | tooling |
| 13 | Regex byte offsets vs. rune string indexing | decision |
| 14 | Float provenance unrecoverable once an ffi result enters | decision |
| 15 | A fault in a spawned computation strands a waiting receiver | decision |
| 16 | §6.6.3's worked `sqrt` example belongs to `math/ffi` only | doc |
| 17 | §6.6.3 does not describe what `iterations` and `terms` buy | doc |
| 18 | Report and *This Is Aiki* disagree on status | doc |
| 19 | `random` help/doc describe an obsolete interface | doc |
| 20 | `IDENTITY` still says eight types | doc |

---

## Numeric path

### 1. `to_decimal` renders exact rationals through binary floating point

`halToDecimal` validates the places argument, then does `f, _ := n.Val.Float64()` and formats `f` with `%.*f`.

Every exact rational therefore reaches the user's eye through a float64 round-trip. This is the display path for the language's central claim, and it is the one used throughout the Report's own Example — including `to_decimal(abs(low_result.range - high_result.range), 6)`, which produces the published `0.000036` residual. At small magnitudes and few places the printed digits happen to be right; at large magnitudes, high place counts, or denominators outside float64's reach, they will not be.

Fix: format from the rational directly — long-divide numerator by denominator to `p` places with correct rounding on the exact remainder. No float64 anywhere in the path.

Report: §6.1.5 should state that `to_decimal` rounds the exact value.

### 2. Audit `lib/` for unparenthesized mixed-operator arithmetic

`math/native.sqrt` contained `guess = (guess + x / guess) / 2`. Under left-to-right evaluation this groups as `((guess + x) / guess) / 2`, giving the iteration x ← (1 + n/x)/2, whose fixed point is (1 + √(1+8n))/4 rather than √n. It converged linearly to the wrong value, so a larger `iterations` argument produced a more precise wrong answer.

The expression is correct in any language with a conventional operator-precedence table and wrong in Aiki, and it was written by someone who knew the rule. That makes it a habit error rather than a typo, and habits repeat.

One instance found — by accident, through a discrepancy between two independent computations in the revised Report Example. The rest of `lib/` has not been checked. Candidates are any expression mixing `+` or `-` with `*` or `/` without explicit grouping; `hash/native`'s modulo reduction is worth particular attention.

### 3. Audit `lib/` for two-element `@error` values

`math/native.sqrt` returned `[@error, "sqrt of negative"]`, omitting the kind symbol. The documented convention is `[@error, :kind, "message"]`, and a three-element pattern does not match the two-element form — so the documented way of handling a recoverable error failed on the one this module produced.

Corrected there. No other module has been checked.

### 4. `math/ffi` sin, cos, and sqrt can construct a Number with a nil `Val`

All three discard the second return of `Float64()`. A rational outside float64 range yields ±Inf; `math.Sin`/`math.Cos` then yield NaN; `big.Rat.SetFloat64` returns nil for non-finite input; the result is `&value.Number{Val: nil}` handed back into the evaluator.

`halSqrt` additionally tests `f < 0` — the sign of the converted float rather than of the rational. The rational's own `Sign()` is the correct test and cannot underflow.

Fix: check the `exact` return from `Float64()` and fault (or return a shaped `@error`, matching the module's convention) when the argument is out of range; sign-check `n.Val.Sign()` before conversion.

Report: §6.6.2–6.6.3 describe intent correctly; no change needed.

### 5. Canvas `toInt` accepts any number and truncates through float64

The helper is `f, _ := n.Val.Float64(); return int(f), true` — it returns `true` for every Number, with no integrality check, no range check, and undefined behavior in the Go conversion for non-finite or out-of-range values.

It is used for coordinates, dimensions, and the r/g/b components of list-form colors, so fractional and out-of-range values are admitted silently at every canvas entry point. §6.11.3 documents truncation for `pen_size` specifically; it does not license it everywhere.

Fix: have `toInt` return `false` on non-integral, non-finite, or out-of-range values, and let each caller produce its own diagnostic. The same unchecked-conversion pattern appears in `halSleep` (`time.Duration(ms)`) — lower severity, same shape.

### 6. `bytes/native` integer validation

`bytes_new(list)` is intended to accept only integer elements in the range 0 through 255. The native implementation checks the range but does not verify integrality, so fractional Aiki numbers within the range may be admitted.

The Report correctly states the intended byte model; no documentation change is required.

## Evaluator and application

### 7. `apply` arity differs from ordinary function application

Ordinary calls fault when fewer than the required ordinary arguments are supplied. The `_apply` path instead binds missing ordinary parameters to `[]`. No language rule or documentation establishes that distinction as intentional.

Fix: make `_apply` enforce the same minimum-arity rule as ordinary function calls.

### 8. `spawn` has the same missing-argument inconsistency

The isolated application path also fills missing ordinary parameters with `[]` — confirmed in `applyUserFunctionIsolated`, which contains the same `callEnv.Set(param, value.EMPTY)` fallback.

Report §7.2.4 states that a call supplying fewer than n arguments faults, and its isolated-application paragraph applies the parameter rules unchanged, so the implementation contradicts the specification in both paths. Fix alongside item 7.

### 9. Pattern matcher omits rune and boolean literals

`matchPattern` dispatches literal matching for numbers, strings, and symbols, but not runes or the boolean terminals `true` and `false`. The grammar permits all five literal-pattern kinds, and `valuesEqual` already supports rune and boolean equality.

Fix: extend `matchPattern` to dispatch `RUNE`, `true`, and `false` through the same literal-equality path. Do not weaken Report §7.2.7 — the report and grammar describe the intended language correctly.

## Library and tooling

### 10. `turtle/simple.new(dim)` can reuse a canvas with the wrong physical dimensions

When a live canvas already exists, `new(dim)` clears and reuses it rather than recreating it. The module nevertheless updates its logical dimension and scale as though the new physical canvas size had been created. Thus `new(400)` followed by `new(800)` can leave an actual 400×400 canvas with state calculated for 800×800.

Fix: recreate the canvas when the requested dimension differs — or more simply, have `new()` always establish a genuinely fresh drawing environment as documented.

### 11. Formatter does not handle rune literal patterns

The grammar permits rune patterns, but the formatter's pattern rendering path omits `RUNE`. Same defect as item 9, in a different tool; fix together.

### 12. Linter models block scope differently from the evaluator

The linter creates scope for ordinary blocks, while Aiki semantics say `if` and `while` blocks execute in the surrounding environment. Successful `match` arms are the special case that introduce an enclosed environment.

Fix: align linter scope handling with evaluator semantics.

## Design gaps — each needs a decision before beta

### 13. Regex positions and string indexing use different coordinate systems

§6.8.2 specifies `[@match, start, end, text]` positions as byte offsets into the UTF-8 representation. §4.9 specifies that `s[i]` indexes by rune, and the string builtins convert through `[]rune` throughout.

A match position therefore cannot be used to index the string it came from. Both documents state their halves correctly, so the inconsistency is in the language as specified rather than a drift between doc and code — which is why it needs a decision rather than a patch. It is invisible on ASCII and silently wrong on anything else.

Options: return rune offsets from `regex/ffi`, converting at the boundary; or keep byte offsets and supply explicit conversion procedures in the `string` module; or make `@match` carry both. The first is most consistent with the rest of the language.

### 14. Float provenance is unrecoverable once an ffi result enters a computation

`SetFloat64` produces an exact rational encoding of a binary float, indistinguishable thereafter from any other rational. Quarantining approximation behind a module name quarantines the call site, not the value: mix one `math/ffi` result into an otherwise native computation and nothing in the resulting number records that a float was involved.

The naming convention tells a reader where approximation entered when they are reading source. It cannot tell them when they are holding the number, and `inspect` will not either.

Question: is call-site visibility sufficient, or does the numeric model need a provenance bit? A bit touches every arithmetic operation and may not be worth it. Cheaper to decide alongside item 13, since both concern a value carrying less information than the reader assumes.

### 15. A fault in a spawned computation strands a waiting receiver

Faults do not propagate from a spawned computation, and channels are unbuffered with no select, timeout, or non-blocking receive. When the spawned computation is the sole producer, its fault leaves `recv` blocked permanently: a diagnosable error becomes a hang, with the diagnostic on stderr and invisible to the program.

Not a defect in itself — each rule is deliberate — but the composition has a sharp edge. §6.1.7 now states the behavior.

Question: whether the concurrency surface needs any means of bounded waiting. It also constrains designs built on Aiki concurrency — an interruptible simulated machine cannot poll a control channel between steps.

## Documentation and help drift

### 16. §6.6.3's worked `sqrt` example belongs to `math/ffi` only

The section covering both implementations gives `math.sqrt(9) ⇒ 3`. That call has one argument, so it is `math/ffi`'s, and `math/native` cannot produce it even when correct — Newton on exact rationals approaches 3 without reaching it.

Consequence: the Report's only worked example of `sqrt` exercised the implementation that was broken, and the defect survived a document review. The lesson generalizes — any section covering two implementations needs a worked example for each.

Fix: give a worked example per implementation, with the native one showing an approximate result and an explicit iteration count.

### 17. §6.6.3 does not describe what `iterations` and `terms` buy

The Report says only that the parameters specify the number of iterations or terms. The relationship to accuracy is neither linear nor unbounded: convergence is quadratic until the value stops changing, after which further iterations alter nothing while roughly doubling the size of the rational produced. The useful count also depends on the magnitude of the argument, since the seed is x/2 — √9 converges by 5 iterations, √894 by 8.

Measured for √894, seeded at x/2:

```
iteration 5    30.7330            42 denominator digits
iteration 6    29.9111            85
iteration 7    29.8998349        172
iteration 8    29.8998327754     345
iteration 9    unchanged         692
```

Fix: state that the parameter bounds work rather than error, that accuracy stops improving once the iteration converges, and that the representation continues to grow.

### 18. The two documents disagree on status

`ra1` presents Aiki 1 as a defining specification whose features are intended to remain stable within the report. *This Is Aiki* describes the same surface as alpha and asks which parts should be frozen for beta.

Both framings are defensible; holding both simultaneously is not. Reconciling them is a prerequisite for the Report functioning as a specification others can implement against.

### 19. Random module help and doc describe an obsolete interface

`random.doc` and `random.help` still describe `random()` returning a value in `[0,1)`. Current behavior is `random(max)` returning an integer `n` such that `0 <= n < max`. `max` must be a positive integer, and `seed(n)` requires an integer seed.

Fix: update module help and doc metadata to the current interface. The Report already follows the implementation.

### 20. `IDENTITY` still says eight types

The repo document folds `bytes` into the basic type set. Both books have since moved it out: seven basic language types, with bytes among the runtime-supplied system types.
