# Proposal: Small-Rational Fast Path for Aiki Numbers

## Status

SUPERSEDED.

Superseded by `proposals/adaptive-exact-number-representation.md`. The original document is retained as design history.

## Summary

Aiki currently represents ordinary numbers as exact arbitrary-precision rationals. This preserves desirable language semantics:

- decimal literals remain exact;
- rational arithmetic remains exact;
- ordinary arithmetic does not silently introduce floating-point approximation;
- values such as `1/3`, `3/10`, and `157/50` retain their mathematical meaning.

The cost is that even very small values pay for arbitrary-precision rational machinery.

This proposal preserves Aiki's existing number semantics while introducing an optimized internal representation for common small rationals.

Conceptually:

```text
Aiki number
    |
    +-- small rational
    |      numerator: machine integer
    |      denominator: machine integer
    |
    +-- big rational
           arbitrary precision
```

Operations remain exact.

When a result fits safely in the small representation, execution stays on the fast path.

When an operation would overflow or otherwise exceed the small representation, the value is promoted transparently to the existing arbitrary-precision rational representation.

No Aiki syntax or user-visible numeric semantics change.

---

## Motivation

Aiki deliberately uses exact rational arithmetic.

That choice should not be weakened merely to obtain conventional floating-point performance.

However, most ordinary program values are small:

```text
0
1
42
1/2
3/10
355/113
1000
```

These values do not inherently require arbitrary-precision storage.

The optimization opportunity is therefore:

> Preserve exact rational semantics while avoiding arbitrary-precision machinery for values that fit naturally in machine-sized integers.

This is an implementation optimization, not a language-design change.

---

## Goals

1. Preserve all existing Aiki number semantics.
2. Keep decimal literals exact.
3. Keep rational arithmetic exact.
4. Make common integer and rational operations substantially cheaper.
5. Promote automatically to arbitrary precision when necessary.
6. Require no user awareness of the internal representation.
7. Keep the implementation inspectable.
8. Make performance claims measurable rather than assumed.
9. Allow the optimization to be removed or revised without affecting Aiki source programs.

---

## Non-goals

This proposal does not:

- replace rationals with floating point;
- introduce approximate ordinary numbers;
- introduce visible machine-integer types;
- add fixed-width numeric syntax;
- change equality;
- change displayed numeric representations;
- change arithmetic evaluation order;
- change overflow semantics at the language level;
- make transcendental functions exact;
- promise hardware-float-equivalent performance.

The objective is lower implementation cost for exact arithmetic, not a different number system.

---

## Proposed Representation

A conceptual representation is:

```go
type Number struct {
    kind numberKind

    smallNum int64
    smallDen int64

    big *big.Rat
}
```

The exact Go layout should be chosen after profiling and benchmarking.

The important semantic states are:

```text
small rational
big rational
```

A small rational is always canonical:

```text
denominator > 0
gcd(|numerator|, denominator) = 1
```

Examples:

```text
0       -> 0/1
5       -> 5/1
1/2     -> 1/2
-3/10   -> -3/10
```

---

## Promotion

Promotion is automatic and invisible.

For example:

```text
small + small
```

attempts a checked machine-integer computation.

If all required intermediate and final values fit:

```text
result -> small
```

If not:

```text
promote operands -> big rational
perform exact operation
result -> big
```

No overflow should be observable by Aiki code.

The language-level result remains mathematically exact.

---

## Demotion

Initial implementation should not require automatic demotion from big rationals back to small rationals.

That keeps the design simple:

```text
small -> big
```

is permitted;

```text
big -> small
```

is optional.

Later benchmarking may show that opportunistic demotion is useful, but it should not be added without evidence.

---

## Integers as a Fast Special Case

A denominator of `1` is common enough to deserve a particularly cheap path.

Conceptually:

```text
small integer
    denominator == 1
```

Operations such as:

```text
2 + 3
10 - 4
7 * 8
```

should avoid general rational reduction whenever possible.

This does not create a separate Aiki integer type.

It is only an internal optimization of the rational representation.

---

## Arithmetic Operations

### Addition and subtraction

For:

```text
a/b + c/d
```

avoid unnecessary growth where possible.

If denominators are equal:

```text
(a + c) / b
```

rather than computing the fully general cross product.

For unequal denominators, use checked operations and reduce intelligently.

Cross-cancellation or GCD-based reduction may prevent avoidable overflow.

### Multiplication

Before multiplying:

```text
a/b * c/d
```

cross-reduce before multiplying.

This both reduces intermediate magnitude and lowers the probability of promotion.

### Division

Division similarly benefits from cross-reduction before multiplication by the reciprocal.

Division by zero retains the existing Aiki fault behavior.

### Negation

Negation remains on the small path unless negating the minimum signed machine integer would overflow.

That case promotes before negation.

### Comparison

Comparison should avoid promotion where safe.

Cross multiplication must use checked arithmetic.

If machine-sized comparison cannot be performed safely, promote or use a wider checked intermediate representation.

---

## Literal Construction

### Integer literals

Small integer literals should normally enter directly as small rationals with denominator `1`.

### Decimal literals

Exact decimal construction should remain unchanged semantically.

For example:

```text
0.1  -> 1/10
3.14 -> 157/50
```

If the reduced numerator and denominator fit the small representation, the value remains small.

Otherwise it is created as a big rational.

### Rational expressions

Expressions such as:

```text
1 / 3
```

remain ordinary exact arithmetic.

If the result fits, it remains small.

---

## Equality

Numeric equality must remain mathematical equality.

Examples:

```text
1 == 1/1
2/4 == 1/2
0.1 == 1/10
```

must continue to behave exactly as they do now.

Internal representation must not affect equality.

Therefore:

```text
small 1/2
big   1/2
```

must compare equal.

---

## Display

Displayed representation must remain independent of storage representation.

A user must never be able to tell whether a number is stored as small or big merely by printing it.

Examples remain:

```text
1/3
3/10
157/50
42
```

No implementation tags or machine-width details should leak into ordinary output.

---

## Error Semantics

Machine overflow must never become an Aiki arithmetic fault merely because the small representation is exhausted.

Instead:

```text
small overflow
    ->
promotion
    ->
exact big-rational operation
```

The user-visible arithmetic model remains unbounded exact rational arithmetic.

---

## Interaction with `math/native`

The small-rational optimization applies to ordinary exact arithmetic.

It does not make irrational results rational.

Functions such as:

```text
sqrt(2, n)
sin(x, n)
cos(x, n)
```

remain approximation algorithms expressed in rational arithmetic.

A faster rational substrate may improve these implementations substantially because their intermediate terms can remain on the small path for some portion of execution.

However, this proposal should not be justified solely by transcendental-function performance.

Its primary target is ordinary arithmetic.

---

## Relationship to `math/nativeo`

A future optimized native math implementation could build on this work.

Possible module taxonomy:

```text
math/native
    clear, inspectable algorithms

math/nativeo
    optimized Aiki-native algorithms

math/ffi
    host-backed implementation
```

The small-rational representation is lower-level than this distinction.

This proposal does not require `math/nativeo`.

---

## Inspectability

The optimization should remain understandable.

The core rule should be explainable as:

> Aiki keeps small exact fractions in machine-sized numerator and denominator fields. If they become too large, it transparently switches to arbitrary-precision rationals.

Avoid representation schemes whose performance advantage depends on obscure encodings or difficult invariants unless measurement shows that the simpler design is inadequate.

---

## Instrumentation

Before and after implementation, gather empirical data.

Useful counters may include:

```text
small-number constructions
small arithmetic operations
promotions to big rational
big-rational operations
GCD reductions
```

These counters should be diagnostic and optional.

Representative workloads can then answer:

```text
What percentage of arithmetic remained small?
Which operations cause promotion?
How much time is spent reducing fractions?
```

---

## Benchmark Plan

Benchmark against the current arbitrary-precision rational implementation.

### Microbenchmarks

Measure:

```text
integer addition
integer subtraction
integer multiplication
rational addition with equal denominators
rational addition with unequal denominators
rational multiplication
division
comparison
negation
literal creation
```

Use both:

```text
small values
promotion-triggering values
already-big values
```

### Representative Aiki workloads

Use actual Aiki programs:

```text
list processing with arithmetic
loops and counters
recursive numeric programs
exact decimal calculations
ballistics example
native sqrt
native sin/cos
profiling experiments
```

The optimization should materially improve ordinary Aiki execution, not merely microbenchmarks.

---

## Acceptance Criteria

### Semantic equivalence

The complete existing test suite must pass unchanged.

Additional tests must show that small and big representations are semantically indistinguishable.

### Exactness

No arithmetic operation may silently approximate.

### Promotion safety

Operations near machine limits must promote correctly rather than overflow.

### Performance

There should be a substantial measured improvement on common exact-arithmetic workloads.

A precise target should be chosen only after establishing the baseline.

Decision rule:

> If complexity increases materially but representative workloads do not improve enough to matter, retain the current implementation.

### Maintainability

The implementation must remain understandable and testable.

---

## Required Tests

### Representation-independent equality

Verify equality between values constructed through small and promoted paths.

### Promotion boundaries

Exercise:

```text
max safe numerator
max safe denominator
addition overflow
subtraction overflow
multiplication overflow
division intermediate overflow
negation boundary
```

All must produce correct exact values.

### Canonical form

Verify:

```text
2/4 -> 1/2
-2/-4 -> 1/2
2/-4 -> -1/2
0/n -> 0
```

according to existing Aiki display and value semantics.

### Cross-reduction

Use values where naïve intermediate multiplication would overflow but cross-reduction permits the exact result to remain small.

### Decimal exactness

Retain tests such as:

```text
0.1 + 0.2 == 3/10
0.1 + 0.1 + 0.1 == 3/10
3.14 == 157/50
```

### Big-rational fallback

Construct values large enough to force promotion and verify subsequent arithmetic.

### Mixed representation

Exercise operations where one operand is small and one is big.

---

## Risks

### Added implementation complexity

The current representation is conceptually simple because every number uses one arbitrary-precision rational model.

A dual representation adds branches and invariants.

This is the primary cost.

### GCD cost

Exact rational normalization is not free.

A small representation does not automatically make rational arithmetic float-like.

Poor reduction strategy could consume much of the expected gain.

### Overflow bugs

Checked arithmetic must be comprehensive.

A missed overflow path would violate one of Aiki's strongest semantic guarantees.

### Premature optimization

The optimization should not be implemented simply because it seems clever.

Measure first.

### Architecture leakage

No library or user code should begin depending on whether a number is small or big.

The representation must stay behind the value abstraction.

---

## Alternatives Considered

### Keep the current representation

This remains the default until measurement justifies change.

Advantages:

- simplest implementation;
- strongest uniformity;
- already correct;
- arbitrary precision everywhere.

### Floating point

Rejected for ordinary Aiki numbers.

It changes language semantics and loses exact rational behavior.

### Dyadic rationals

A machine-oriented representation of:

```text
integer * 2^exponent
```

can be efficient, but it cannot represent values such as `1/3` or `1/10` exactly.

Rejected as a replacement for Aiki rationals.

### Arbitrary-precision floating point

Useful for approximation work, but still not exact general rational arithmetic.

It addresses a different problem.

### External optimized rational library

Potentially useful later, but adds dependency and boundary complexity.

The internal small-rational fast path should be evaluated before introducing a major external arithmetic dependency.

---

## Implementation Sketch

### Phase 1 — Benchmark baseline

Record current performance of `big.Rat`-backed Aiki arithmetic.

No semantic changes.

### Phase 2 — Prototype small representation

Implement a private experimental number type with:

```text
int64 numerator
int64 denominator
promotion to big.Rat
```

Exercise it independently of the evaluator.

### Phase 3 — Differential testing

Generate arithmetic cases and compare:

```text
small/big hybrid result
```

against:

```text
current big.Rat result
```

for equality.

Include randomized and boundary-heavy cases.

### Phase 4 — Integrate behind `value.Number`

Replace internal storage without changing public APIs.

Run all tests, golds, smokes, invariants, and Aiki-native tests.

### Phase 5 — Measure

Compare representative workloads against baseline.

### Phase 6 — Decide

Keep the optimization only if the measured benefit justifies the added machinery.

---

## Deferred Decision

No implementation is proposed now.

The current arbitrary-precision rational representation remains authoritative.

This proposal exists so that future optimization work begins from a clear constraint:

> Performance may change. Aiki's exact rational semantics may not.

The optimization is worth revisiting when profiling shows that rational arithmetic is a meaningful execution cost.

Until then, correctness, inspectability, and architectural simplicity take precedence.
