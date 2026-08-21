# Proposal: Adaptive Exact Number Representation

## Status

**COMPLETE**

## Baseline

Baseline: `cbddead`, on the `runtime-optimization` project line.

This proposal supersedes `proposals/superseded/aiki-small-rational-fast-path-proposal.md`.

## Summary

Aiki has one numeric type: `number`.

That semantic boundary does not require one physical representation. The present
implementation represents ordinary Aiki numbers universally with arbitrary-
precision rational machinery. This makes the representation simple, but forces
machine-small values and machine-realizable arithmetic to pay costs their
semantics do not require.

This proposal separates **numeric semantics** from **numeric representation**.

> An Aiki number has exact rational semantics wherever the mathematical result is
> rational. The runtime may use any internal representation that preserves the
> exact Aiki value. Internal representation is never observable as a numeric type,
> promotion rule, or loss of precision.

Conceptually:

```text
                         Aiki number
                             |
              exact semantic numeric boundary
                             |
          +------------------+------------------+
          |                  |                  |
     small integer      compact rational    exact binary64
          |                  |                  |
          +------------------+------------------+
                             |
                         big.Rat
                      exact fallback
```

These are alternate hidden representations, not user-visible numeric types.

The common path should approach machine arithmetic speed. The uncommon path must
remain exact. `big.Rat` becomes the arbitrary-precision fallback, not the
definition of an Aiki number.

## Architectural principle

Aiki distinguishes programmer meaning, architectural contract, and substrate
realization. For numbers:

```text
Aiki meaning
    number

semantic contract
    exact rational value where the operation is rational

runtime realization
    selected according to the value and operation
```

The implementation may therefore vary radically without changing the language.
A number may internally be a machine integer, compact rational, exact binary64
carrier, or arbitrary rational, but `type(x)` always reports `:number`.

There is no user-visible integer/float/rational hierarchy, no numeric contagion
rule, and no operation in which representation determines semantics.

### Binary64 clarification

The binary64 carrier does **not** mean "Aiki now has floats."

It means:

> This finite IEEE-754 bit pattern is an exact dyadic rational that Aiki already
> received from the host.

A finite binary64 value denotes one exact rational number. Retaining its bits is a
compact exact representation of that returned value. It is not permission for
ordinary Aiki arithmetic to round. Ordinary arithmetic may remain on a binary64
path only when the exact Aiki result can be certified cheaply; otherwise it must
promote to an exact rational representation before computing the result.

The conservative certification gate in Cut 5 is load-bearing.

### Relationship to fixed-width machine values

Adaptive Number and the fixed-width machine FFI domain are complementary, not
competing.

Adaptive Number improves ordinary Aiki arithmetic and host-math boundaries. The
machine domain owns systems interiors such as PDP-11 and IBM 7094 emulation,
where fixed-width words, bytes, addresses, wraparound, and bit operations should
remain opaque machine values and should not re-enter ordinary Number arithmetic.

Keep these semantic boundaries distinct.

## Semantic invariants

### 1. Exact ordinary arithmetic

Existing exact rational behavior remains exact. `0.1 + 0.2` remains `3/10`; `1 / 3 * 3`
remains `1`. Representation may not introduce approximation.

### 2. Exact literals

Integer, decimal, and rational literals denote their existing exact values. In
particular, `0.1 = 1/10`, `3.14 = 157/50`, and `1/3 = 1/3`. Decimal parsing may not
silently pass through binary floating point.

### 3. One numeric type

All internal representations remain `number`. No representation query enters the
language surface.

### 4. Representation-independent equality and ordering

Two numbers denoting the same mathematical rational compare equal and order the
same regardless of representation.

### 5. Representation-independent display

Inspection, printing, conversion, equality, ordering, hashing where applicable,
and documentation behavior depend on semantic value, not internal representation.

### 6. Arbitrary precision remains available

A computation that exceeds machine representations promotes rather than
overflows at the Aiki level.

### 7. Approximation boundaries remain explicit

Operations whose mathematics already require approximation, such as host-backed
transcendental functions, may produce an approximate result according to their
existing contract. Once such a result enters Aiki as a number, the returned
numeric value is represented exactly thereafter.

### 8. No NaN or infinity as ordinary Aiki numbers

Non-finite hardware values are not rationals. Existing error/domain policy
applies rather than exposing IEEE special values as Aiki numbers.

## Candidate representations

The initial candidate representation is conceptually:

```text
Number
    SMALL_INT
    SMALL_RAT
    BINARY64_EXACT
    BIG_RAT
```

The exact Go layout is an implementation decision to be measured rather than
prescribed prematurely.

### Small integer

A signed machine integer with implicit denominator 1. This is expected to
dominate ordinary computation: counters, indexes, lengths, sizes, ordinals,
small arithmetic, and program state. Checked machine arithmetic promotes on
overflow. The ordinary path should require no heap allocation or GCD.

### Compact rational

A canonical machine-sized numerator and denominator, conceptually signed `int64`
numerator plus positive machine denominator. This representation covers values
such as `1/2`, `1/3`, `3/10`, `157/50`, and `355/113` without arbitrary-precision
objects. Arithmetic uses cancellation and checked operations where useful.

### Exact binary64 carrier

A finite IEEE-754 binary64 value is an exact representation of a dyadic rational.
The carrier is especially valuable at host math boundaries, where the host has
already returned a finite binary64 approximation. Expanding those same 64 bits
immediately into a large numerator and power-of-two denominator adds no semantic
precision.

### Big rational

`big.Rat` remains the universal exact fallback for arbitrarily large integers,
rationals exceeding compact representation, and operations that cannot be
certified exact on a cheaper path.

## Promotion and demotion

Promotion is invisible and value-preserving. Values do not have to traverse every
tier. A host-math result may enter directly as binary64; `1/3` may enter directly
as compact rational; a huge literal may enter directly as big rational.

Demotion is permitted but not required on every operation. Eager demotion is a
performance policy, never a semantic requirement.

## Binary64 arithmetic policy

A binary64 operation may remain on the binary64 path only when exactness can be
established without approximation. If exactness cannot be certified, operands
must transition to an exact rational representation and the operation must be
performed there.

Implementation begins conservatively. Candidate techniques include
exponent/significand analysis, checked integer operations over decomposed binary
values, and hardware-assisted/error-free transforms. No technique is accepted
without both semantic and performance evidence.

## Semantic authority

The project must stop treating `*big.Rat` as synonymous with `Aiki number`.
`Number` becomes the authority for numeric semantics. Consumers ask that
abstraction for arithmetic, comparison, sign, integer conversion,
numerator/denominator when semantically required, host conversion, and
inspection rather than reaching through to representation fields.

## Performance objective

> Execute the largest possible proportion of ordinary Aiki numeric operations on
> allocation-free or near-allocation-free machine paths while preserving the
> existing number boundary exactly.

Expected hot-path priority:

```text
1. small integer
2. compact rational
3. exact binary64 where naturally applicable
4. arbitrary rational fallback
```

## Instrumentation

Diagnostic instrumentation should distinguish number construction, small-integer
operations, compact-rational operations, binary64-carrier operations, big-rational
operations, promotion paths, fallback from attempted fast arithmetic,
GCD/reduction operations, and number allocations where measurable.

Instrumentation is optional and not part of numeric semantics.

## Benchmark workloads

Measure both microbenchmarks and real Aiki programs. Microbenchmarks include
integer construction/arithmetic/comparison, compact rational operations,
binary64 carrier construction/exact arithmetic/fallback, promotion-triggering
arithmetic, already-big arithmetic, and display/conversion.

Record at minimum `ns/op`, allocations/op, and bytes/op where applicable.

Whole-program witnesses include loops/counters, indexed/list processing,
recursive numeric programs, exact decimal calculations, rational-heavy
workloads, math/native and math/ffi workloads, self-host acceptance, and the
PDP-11 emulator performance workload. The emulator is evidence and regression
workload, not the architectural driver.

# Planned cuts

## Cut 0 — Numeric baseline and representation audit

Status: **GATED**

No behavior changes.

Inventory every direct dependency on `Number.Val`/`big.Rat`, constructors and
parsers, evaluator arithmetic, comparisons/equality, conversions, HAL/math
consumers, self-host assumptions, representation-level tests, and documentation
that says `big.Rat` when it means exact Aiki number. Capture baseline performance
for representative numeric workloads and establish semantic differential
fixtures.

### Gate

- complete inventory recorded;
- current semantic corpus green;
- baseline benchmark/profile evidence retained;
- no implementation change.

## Cut 1 — Establish the Number semantic authority

Status: **GATED**

Refactor `Number` so callers no longer depend directly on `*big.Rat`. Initially,
`big.Rat` may remain the only physical representation. Introduce
representation-independent internal operations required by current consumers and
replace representation assertions with semantic assertions.

This cut deliberately gains no performance. Its purpose is to separate semantic
authority from storage realization before optimization begins.

### Gate

- whole existing numeric behavior unchanged;
- no uncontrolled direct `big.Rat` access remains outside numeric representation authority;
- focused number tests green;
- self-host numeric acceptance green;
- `make validate` green.

## Cut 2 — Small-integer fast path

Status: **GATED**

Add an internal signed machine-integer representation and allocation-free checked
fast paths for construction, `+`, `-`, `*`, comparison, equality, sign, zero,
integer predicates, and small integral conversions. Division remains exact;
integral results may remain integer and fractional results transition exactly.
Overflow promotes rather than wraps.

### Gate

- differential equivalence against exact rational reference behavior;
- overflow/promotion boundaries tested;
- no observable type/display/equality change;
- common integer operations materially reduce allocations and runtime;
- representative workloads prove the path is exercised;
- `make validate` green.

## Cut 3 — Compact-rational fast path

Status: **GATED**

Add canonical machine-sized rational representation. Implement checked/cancelled
construction, arithmetic, comparison, equality, and negation. Preserve exact
decimal parsing. Promote where machine-sized exact computation is unsafe.

### Gate

- randomized differential testing against `big.Rat`;
- machine-boundary edge tests;
- canonicalization invariants hold;
- exact decimals remain exact;
- rational workloads materially reduce arbitrary-precision work;
- `make validate` green.

## Cut 4 — Exact binary64 carrier

Status: **GATED**

Introduce finite binary64 as a hidden exact dyadic-rational representation,
initially where a binary64 value naturally already exists at host numeric
boundaries. Implement representation-independent equality, ordering, inspection,
conversion, and promotion. Normalize negative zero according to Aiki equality;
non-finite values remain outside ordinary number semantics.

### Gate

- every accepted finite binary64 carrier denotes the same exact rational as
  reference conversion;
- cross-representation equality/order is differential-tested;
- host math results no longer require immediate `big.Rat` expansion;
- public output is unchanged;
- `make validate` green.

## Cut 5 — Certified machine arithmetic for exact binary values

Status: **GATED**

Investigate and benchmark exactness certification for binary64 `+`, `-`, `*`,
and `/`. Enable an operation only where exactness can be established cheaply. If
not certified, promote and compute exactly.

### Gate

A candidate operation enters the runtime only if strong differential/adversarial
testing validates exactness, no rounded result can silently become an Aiki
ordinary arithmetic result, and measured performance beats immediate rational
fallback on relevant workloads.

Cut 5 gate evidence on the authoritative Go 1.24 host: binary carrier construction ~21 ns / 32 B / 1 allocation; certified exact add ~22 ns / 32 B / 1 allocation; certified exact multiply ~27 ns / 32 B / 1 allocation; common failed certification falls exactly through compact rational at ~52 ns / 32 B / 1 allocation; only genuinely large dyadic fallback reaches `big.Rat` (~1.85 us / 2209 B / 26 allocations). Certified division is not admitted. Cut 7 workload evidence did not justify another proof path: common exact binary fallback is already about 52 ns / 32 B / 1 allocation on the authoritative host, and the measured witnesses obtain their useful binary coverage from certified add/subtract/multiply plus compact-rational fallback.

## Cut 6 — Host math and numeric boundary reconciliation

Status: **GATED**

Route math/FFI numeric conversion through Number authority. Retain finite host
binary64 approximations compactly as exact returned values rather than expanding
them immediately to arbitrary rational form. Audit transcendental functions,
random/time numeric paths, canvas/numeric consumers, and numeric FFI conversions.

### Gate

- host-boundary behavior matches documented semantics;
- no user-visible float type exists;
- returned approximations remain stable exact Aiki numbers thereafter;
- math workloads reduce conversion/allocation cost;
- `make validate` green.

## Cut 7 — Representation distribution and workload proof

Status: **GATED**

Run the adaptive representation against whole-program workloads and record
percentage of operations by representation, promotion/fallback rates, allocation
change, and elapsed change.

### Gate

- majority of ordinary numeric operations avoid `big.Rat` on representative workloads;
- integer-heavy workloads approach machine-arithmetic cost;
- rational-heavy programs retain exact semantics;
- host-numerical workloads benefit from the binary carrier;
- no catastrophic representation-churn regression.

If a representation does not justify its complexity with measured coverage,
**remove it**. The proposal does not require all candidate representations to
survive.

### Gate evidence

Authoritative Go 1.24 workload evidence:

- Three-level self-host interpretation (`host -> selfhost -> selfhost -> 1 + 2 * 3`) recorded 1,122,252 small-integer arithmetic results and no compact-rational, binary-carrier, or big-rational arithmetic results. The workload performed 1,289,910 arithmetic events, 8,701,738 calls, and 1,443,452 indexes. This proves ordinary interpreter machinery is overwhelmingly served by the small-integer realization.
- Experiment 004 PDP-11 Cut 7 at 10x / 7,680 guest instructions recorded 20,834 small-integer arithmetic results and no rational/binary/big results while the PDP interior remained in the distinct fixed-width machine domain. Against the immediately preceding same-work semantic profile, elapsed fell from about 2.534 s to 1.924 s, allocation from about 1.667 GB to 1.302 GB, and mallocs from about 26.47 M to 21.12 M without reducing guest instructions or Aiki call count. This demonstrates that adaptive Number and the machine FFI domain are complementary.
- The rational witness recorded 300,001 small-integer and 200,004 compact-rational arithmetic results, zero binary and zero big-rational results, while preserving outputs `0`, `3/10`, and `1`. Compact rational therefore earns its place independently of the integer path.
- The host-math witness recorded 399,998 binary-carrier Number returns from FFI calls. Arithmetic realization was 200,002 small integer, 54,486 compact rational, 145,436 binary carrier, and only 76 big rational. Of 299,997 measured binary arithmetic attempts, 245,435 were certified exact and 54,562 fell back exactly; only 76 promoted to arbitrary precision. The carrier and certified add/subtract/multiply therefore earn their place.
- Certified division is not admitted because none of the workload evidence establishes a need large enough to justify another exactness-proof path.

## Cut 8 — Documentation, invariants, and self-description

Status: **GATED**

Correct documentation and tests that conflate semantic exactness with universal
arbitrary-precision representation. Replace representation invariants such as
"numbers are big.Rat" and "no floats exist internally" with semantic invariants:
ordinary rational arithmetic is exact, representation is unobservable, no fast
path may introduce rounding, and all representations agree on observable value.

### Gate

- language report, user docs, executable docs, invariants, and implementation agree;
- documentation does not expose the representation ladder as a programming concept;
- self-host remains conformant;
- `make validate` green.

## Cut 9 — Reconciliation and whole-tree gate

Status: **GATED**

Reconcile this proposal, the superseded fast-rational proposal, audit findings,
session history, profiling, documentation, tests, benchmarks, and remaining
limitations. Remove representation machinery that did not survive its gate.

### Gate

- semantic equivalence established;
- representation boundary encapsulated;
- performance benefit measured;
- strongest repository validation green;
- documentation reconciled;
- no unexplained generated artifacts.

# Acceptance criteria

1. Aiki still exposes exactly one numeric type: `number`.
2. Integer, decimal, and rational literals retain current values.
3. Ordinary rational arithmetic remains exact.
4. Arbitrary numeric magnitude remains supported.
5. No machine overflow becomes an Aiki arithmetic result.
6. All surviving internal representations are observationally equivalent.
7. No rounded binary64 operation is silently accepted as exact ordinary arithmetic.
8. Finite host binary64 results may be retained as exact compact carriers of the returned approximation.
9. `big.Rat` is fallback representation, not the public definition of `number`.
10. Common integer arithmetic avoids arbitrary-precision rational machinery.
11. Common compact rational arithmetic avoids arbitrary precision when safe.
12. Representation distribution and promotion behavior are measurable.
13. The optimization materially improves representative programs.
14. Full project validation remains green.
15. No user program must know or care which representation is in use.
16. Fixed-width machine-value domains remain distinct and do not route systems interiors back through Number.

# Explicit non-goals

This proposal does not introduce user-visible integer or floating types, an
exact/inexact numeric hierarchy, IEEE NaN/infinity as ordinary numbers, coercion
rules based on representation, user-selectable numeric storage, representation
syntax, approximate decimal literals, representation-specific equality, or
implementation-specific arithmetic behavior.

It does not promise every mathematical function is mathematically exact. It
promises that **representation never adds approximation beyond the operation's
defined semantic boundary**.

# Failure conditions

Stop or simplify if representation dispatch erases expected gains, compact
rationals promote too frequently, binary64 certification costs more than
fallback, representation churn creates major regressions, semantic equivalence
requires fragile special cases, or complexity is disproportionate to measured
benefit.

If measured coverage later shows one representation does not pay for itself,
drop it.

# Resulting architectural claim

> Aiki has one numeric type with an exact semantic boundary. The runtime chooses
> the cheapest representation that can preserve the value: machine integer,
> compact rational, exact binary carrier, or arbitrary rational. These
> representations are implementation details. When a cheap representation can no
> longer preserve the exact result, Aiki changes representation rather than
> changing the number.

Or, more compactly:

> **The number is exact. The representation is negotiable.**


# Completion

Completed 2026-08-20.

The strongest whole-tree `make validate` gate passed after documentation,
invariant, profiling, and treecheck reconciliation. No additional numeric
representation mechanism was required.

The surviving implementation is:

- hidden small integer for the dominant ordinary path;
- hidden compact rational for exact non-integral arithmetic and common binary
  fallback;
- hidden finite binary64 carrier only as an exact dyadic representation of
  host-returned values;
- certified exact binary add/subtract/multiply where cheap proof succeeds;
- `big.Rat` as the arbitrary-precision escape representation.

Certified binary division was deliberately not admitted because measured
workloads did not justify another proof path.

The controlling architectural statement is:

> **The number is exact. The representation is negotiable.**

The fixed-width machine FFI domain remains a separate boundary for systems
interiors such as the PDP-11/7094 experiments.
