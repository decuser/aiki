Here’s the cleaned **known bug/drift list** I’d carry forward before resuming the document check.

### Implementation bugs

1. **Pattern matcher omits rune and boolean literals**

   `matchPattern` dispatches literal matching for numbers, strings, and symbols, but not runes or the boolean terminals `true` and `false`.

   The grammar permits all five literal-pattern kinds, and `valuesEqual` already supports rune and boolean equality.

   **Fix:** extend `matchPattern` to dispatch `RUNE`, `true`, and `false` through the same literal-equality path.

   **Do not weaken Report §7.2.7.** The report and grammar describe the intended language correctly.

2. **`apply` arity differs from ordinary function application**

   Ordinary calls fault when fewer than the required ordinary arguments are supplied. The current `_apply` path instead binds missing ordinary parameters to `[]`.

   No language rule or documentation establishes that distinction as intentional.

   **Fix:** make `_apply` enforce the same minimum-arity rule as ordinary function calls.

3. **`spawn` has the same missing-argument inconsistency**

   The isolated application path used by `spawn` also fills missing ordinary parameters with `[]`.

   *This Is Aiki* says ordinary parameter/rest-parameter rules apply to spawned functions, so this is the same underlying inconsistency.

   **Fix:** enforce ordinary minimum arity for spawned user functions as well.

4. **`turtle/simple.new(dim)` can reuse a canvas with the wrong physical dimensions**

   When a live canvas already exists, `new(dim)` clears and reuses it rather than recreating it. The module nevertheless updates its logical dimension/scale as though the new physical canvas size had been created.

   Thus `new(400)` followed by `new(800)` can leave an actual 400×400 canvas with state calculated for 800×800.

   **Fix:** recreate the canvas when the requested dimension differs—or more simply, have `new()` always establish a genuinely fresh drawing environment as documented.

### Tooling bugs

5. **Formatter does not handle rune literal patterns correctly**

   The grammar permits rune patterns, but the formatter’s pattern rendering path omits `RUNE`.

   **Fix:** add rune literal handling to the formatter pattern case.

6. **Linter models block scope differently from the evaluator**

   The linter creates scope for ordinary blocks, while Aiki runtime semantics say `if` and `while` blocks execute in the surrounding environment. Successful `match` arms are the special case that introduce an enclosed environment.

   **Fix:** align linter scope handling with evaluator semantics.

### Documentation/help drift

7. **Random module help/docs describe an obsolete interface**

   `random.doc` / `random.help` still describe:

   ```text
   random()
   ```

   returning a value in `[0,1)`.

   Current behavior is:

   ```text
   random(max)
   ```

   returning an integer `n` such that:

   ```text
   0 <= n < max
   ```

   `max` must be a positive integer, and `seed(n)` requires an integer seed.

   **Fix:** update module help/doc metadata to the current interface. The Report already follows the implementation.

8. **Math documentation had drifted from the actual exported surfaces**

   This is the Report §6.6 issue we just fixed: nonexistent `round`, `tan`, `asin`, `acos`, `atan`, plus the differing `math/ffi` and `math/native` signatures.

   **Status:** document fix, not implementation bug.

9. bytes/native integer-validation bug: bytes_new(list) is intended to accept only integer elements in the range 0 through 255. The native implementation checks the range but does not verify integrality, so fractional Aiki numbers within the range may be admitted. The Report correctly states the intended byte model; no documentation change is required.



---

## Additions from Report/code cross-check

### Implementation bugs (numeric path)

10. **`to_decimal` renders exact rationals through binary floating point**

    `halToDecimal` validates the places argument, then does `f, _ := n.Val.Float64()` and formats `f` with `%.*f`.

    Every exact rational therefore reaches the user's eye through a float64 round-trip. This is the *display* path for the language's central claim, and it is the one used throughout the Report's own Example — including `to_decimal(abs(low_result.range - high_result.range), 6)`, which produces the published `0.000036` residual. At small magnitudes and few places the printed digits happen to be right; at large magnitudes, high place counts, or denominators outside float64's reach, they will not be.

    **Fix:** format from the rational directly — long-divide numerator by denominator to `p` places with correct rounding on the exact remainder. No float64 anywhere in the path.

    **Report:** §6.1.5 should state that `to_decimal` rounds the exact value, so the guarantee is on the record.

11. **`math/ffi` sin, cos, and sqrt can construct a Number with a nil `Val`**

    All three discard the second return of `Float64()`. A rational outside float64 range yields ±Inf; `math.Sin`/`math.Cos` then yield NaN; `big.Rat.SetFloat64` returns nil for non-finite input; the result is `&value.Number{Val: nil}` handed back into the evaluator.

    `halSqrt` additionally tests `f < 0` — the sign of the *converted float* rather than of the rational. The rational's own `Sign()` is the correct test and cannot underflow.

    **Fix:** check the `exact` return from `Float64()` and fault (or return a shaped `@error`, matching the module's existing convention) when the argument is out of range; sign-check `n.Val.Sign()` before conversion.

    **Report:** §6.6.2–6.6.3 describe intent correctly; no change needed.

12. **Canvas `toInt` accepts any number and truncates through float64**

    The helper is `f, _ := n.Val.Float64(); return int(f), true` — it returns `true` for every Number, with no integrality check, no range check, and undefined behavior in the Go conversion for non-finite or out-of-range values.

    It is used for coordinates, dimensions, and the r/g/b components of list-form colors, so fractional and out-of-range values are admitted silently at every canvas entry point. (§6.11.3 documents truncation for `pen_size` specifically; it does not license it everywhere.)

    **Fix:** have `toInt` return `false` on non-integral, non-finite, or out-of-range values, and let each caller produce its own diagnostic. Note the same unchecked-conversion pattern in `halSleep` (`time.Duration(ms)`), which is lower-severity but the same shape.

### Design gaps (not surface defects; both need a decision before beta)

13. **Regex positions and string indexing use different coordinate systems**

    §6.8.2 specifies `[@match, start, end, text]` positions as byte offsets into the UTF-8 representation. §4.9 specifies that `s[i]` indexes by rune, and the string builtins convert through `[]rune` throughout.

    A match position therefore cannot be used to index the string it came from. Both documents state their halves correctly, so this is an inconsistency in the language as specified rather than a drift between doc and code — which is why it needs a decision rather than a patch. It is invisible on ASCII and silently wrong on anything else.

    **Options:** return rune offsets from `regex/ffi` (converting at the boundary); or keep byte offsets and supply explicit conversion procedures in the `string` module; or make `@match` carry both. The first is most consistent with the rest of the language.

14. **Float provenance is unrecoverable once an ffi result enters a computation**

    `SetFloat64` produces an exact rational encoding of a binary float, indistinguishable thereafter from any other rational. Quarantining approximation behind a module *name* quarantines the call site, not the value: mix one `math/ffi` result into an otherwise native computation and nothing in the resulting number records that a float was involved.

    The naming convention tells a reader where approximation entered when they are reading source. It cannot tell them when they are holding the number, and `inspect` will not either.

    **Question for beta:** is call-site visibility sufficient, or does the numeric model need a provenance bit? A bit is a real cost — it touches every arithmetic operation — and may not be worth it. Worth deciding deliberately rather than by default.

### Documentation drift

15. **The Report's Example does not state why `math/native` was chosen over `math/ffi`**

    The narrative says the velocity components come from the native module and that its sine and cosine use Taylor series over Aiki numbers. It never says that native is used *so that no floating-point value enters the computation at all*, nor that `PI` and `TERMS` are passed at the spawn site because a call site should declare its own epsilons.

    Stated as mechanism without intent, the parameter count reads as an apology for the isolation rule rather than as the design working. One clause in each place fixes it.

16. **Example binds pattern names it does not use**

    `[@result, :low, range, peak, duration]` and its `:high` counterpart bind three names, then the arm body uses `message`. `[@result, :low, _, _, _]` is the disciplined form, and the Report's showcase example is where discipline should be visible.

17. **`IDENTITY` still says eight types**

    The repo document folds `bytes` into the basic type set. Both books have since moved it out: seven basic language types, with bytes among the runtime-supplied system types.

18. **The two documents disagree on status**

    `ra1` presents Aiki 1 as a defining specification whose features "are intended to remain stable within this report." *This Is Aiki* describes the same surface as alpha and asks which parts should be frozen for beta.

    Both framings are defensible; holding both simultaneously is not. Reconciling them is a prerequisite for the Report functioning as a specification others can implement against.
