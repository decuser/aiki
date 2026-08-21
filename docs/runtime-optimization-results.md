# Runtime Optimization Results

## Purpose

This document records the cumulative runtime result and the architectural
lessons established by the optimization work. Individual proposals retain the
detailed implementation and gate evidence; this file preserves the cross-project
story so later profiling starts from the correct baseline and does not lose the
reasoning that produced it.

## Governing principle

Aiki optimization preserves the semantic boundary and negotiates physical
realization beneath it.

> **Semantic properties are authoritative. Physical representation is negotiable.**

This is not permission to weaken semantics for speed. An adaptive representation
is acceptable only when the programmer-observable property remains unchanged and
the representation earns its complexity under measured workloads.

The demonstrated cases are now:

### Number

Semantic property: **exactness**.

> **The number is exact. The representation is negotiable.**

Surviving realizations:

- small integer;
- compact rational;
- exact finite binary64 carrier;
- `big.Rat` escape;
- certified exact binary add/subtract/multiply where the proof succeeds.

The machine fixed-width domain remains separate from ordinary Aiki Number.

### List

Semantic property: **immutability**, with persistent sequence semantics.

> **The list is immutable. The representation is negotiable.**

Surviving realizations:

- flat immutable list;
- private synchronized frontier backing for amortized append;
- historical-prefix fork when persistence requires a branch.

The implementation may extend unused private backing storage, but an older list
value never observes a later append. Capacity and frontier state are realization
facts, not semantic state.

The detailed list proposal also uses the more implementation-specific statement
"the list is persistent" when describing branch preservation. The programmer-
facing driver is immutability; persistence is the consequence the adaptive
append path must preserve.

## Two distinct runtime pathologies

The optimization sequence exposed two materially different costs.

### 1. Allocation frequency at call/runtime boundaries

The post-Number three-level self-host initially performed enormous numbers of
small realization allocations around calls, environments, profile-label setup,
argument construction, and tail-call machinery.

The call/allocation tranche attacked **how often objects were allocated** without
changing Aiki call semantics.

Result:

```text
                         post-Number       after call/runtime      change
elapsed                  19.394718734 s    14.427563993 s          -25.6%
alloc_bytes              28,291,183,320    25,005,019,784          -11.6%
mallocs                  110,603,982       50,560,133              -54.3%
gc_cycles                899               829                     -7.8%
```

The important diagnostic was the asymmetry: malloc count fell by more than half,
while allocated bytes fell only modestly. Removing allocation frequency exposed
a second problem rather than finishing the runtime work.

### 2. Allocation volume from persistent container construction

After call/runtime cleanup, the self-host still allocated roughly 25 GB despite
only about 50.6 million mallocs. Repeated immutable list append was realizing
quadratic copied-slot volume.

Adaptive persistent lists attacked **how much memory each construction pattern
realized**, while preserving immutable/persistent list semantics.

Result:

```text
                         before lists      adaptive lists          change
elapsed                  14.427563993 s    10.619609147 s          -26.4%
alloc_bytes              25,005,019,784    8,901,681,936           -64.4%
mallocs                  50,560,133        50,401,888              -0.3%
gc_cycles                829               267                     -67.8%
```

Again the asymmetry is the result: allocation count was essentially flat, while
allocated bytes and GC pressure collapsed. This confirms that the list project
solved allocation **volume**, not another call-frequency problem.

Self-host list realization at the final gate:

```text
frontier_promoted        52928
frontier_extended        209407
frontier_grown           4039
frontier_forked          0
elements_copied          274328
backing_slots_allocated  760368
```

Historical branching—the principal risk to the frontier design—did not occur in
the motivating self-host workload.

## Overall improvement

From the original post-Number self-host baseline to the completed adaptive-list
state:

```text
                         original          current                change
elapsed                  19.394718734 s    10.619609147 s         -45.2%
alloc_bytes              28,291,183,320    8,901,681,936          -68.5%
mallocs                  110,603,982       50,401,888             -54.4%
gc_cycles                899               267                    -70.3%
```

The Aiki semantic workload remained the same across these comparisons:

```text
arithmetic   1289910
comparison   1911310
call         8701738
iteration    1005458
index        1443452
```

The cumulative result therefore reflects cheaper realization of the same
interpreter work, not a reduced workload.

## Systems regression witness

The PDP-11 Cut 7 10x witness remained materially flat across adaptive-list work:

```text
                         pre-list          adaptive list
elapsed                  1.562006621 s     1.569385077 s
alloc_bytes              1,040,319,496     1,040,367,792
mallocs                  16,463,419        16,463,466
gc_cycles                33                34
```

This witness remains useful because it pressures ordinary Aiki/system execution
without sharing the self-host parser's list-construction profile.

## Driver for future optimization

Do not infer the next adaptive representation from analogy alone.

The next runtime investigation begins from the current authoritative self-host
baseline:

```text
elapsed      10.619609147s
alloc_bytes  8901681936
mallocs      50401888
gc_cycles    267
```

First attribute the remaining allocation bytes/objects. If strings, AST values,
environments, bytes, or another family dominate, establish that with evidence
before proposing a new representation.

The reusable lesson is:

> **Optimize realization beneath the semantic boundary rather than weakening the
> semantic boundary to obtain performance.**
