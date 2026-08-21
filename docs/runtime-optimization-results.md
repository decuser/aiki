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

The controlling next-step contract is
`proposals/runtime-realization-survey.md`. It intentionally has one critical
gate: select the next semantic realization family only after self-host/PDP
allocation-space and allocation-object attribution is reconciled.

The reusable lesson is:

> **Optimize realization beneath the semantic boundary rather than weakening the
> semantic boundary to obtain performance.**


## Alpha-36 runtime-realization survey

The first post-list allocation survey changed the diagnosis again.

Three-level self-host allocation-space ownership:

```text
evalStringIndex                  6171.30 MB   73.20%
NewCallEnv                        931.20 MB   11.04%
evalCallArgs                      210.01 MB    2.49%
Parser.recordFailure              192.56 MB    2.28%
```

The remaining self-host byte volume is therefore dominated by **string
observation materialization**: rune indexing converts the complete immutable
string to `[]rune` to retrieve one rune.

This is distinct from both earlier pathologies:

```text
call/runtime       allocation frequency
persistent lists   allocation volume from copy-on-append
string indexing    observation-time whole-value materialization
```

The initial string response is deliberately not adaptive representation.
Flat UTF-8 remains authoritative while rune-aware observation becomes
allocation-free.

The PDP-11 survey independently identifies parser failure bookkeeping as its
dominant allocation family (`Parser.recordFailure`, about 60% of allocation
space). That remains a separate future parser-realization concern.


## String observation survey selection

The alpha-36 allocation survey identified `evalStringIndex` as the dominant
remaining self-host allocation-space site:

```text
evalStringIndex   ~6.17 GB   ~73.2% of allocated bytes
```

This selected immutable string observation as the next runtime family. The
initial response is deliberately narrower than an adaptive representation:

> **The string is immutable. Observation need not materialize it.**

Whole-string `[]rune` materialization is removed from observation operations
before any alternate string representation is considered.


## Immutable string observation — COMPLETE

The alpha-37 string-observation wave removed whole-string rune materialization
from immutable observation without introducing a second string representation.

Three-level self-host:

```text
                         pre-fix          post-fix
elapsed                  10.6196 s        9.2246 s
allocated                 8.902 GB        2.396 GB
mallocs                  50.402 M        50.297 M
GC cycles                267             79
```

The former `evalStringIndex` site (~6.17 GB / 73.2% of allocation space)
disappeared from the leaders. PDP-11 Cut 7 10x remained materially flat.

Driver:

> **The string is immutable. Observation need not materialize it.**

Cumulative self-host change from the original post-Number baseline:

```text
elapsed      19.3947 s -> 9.2246 s
allocated    28.291 GB -> 2.396 GB
mallocs      110.604 M -> 50.297 M
GC cycles    899 -> 79
```


## Environment realization and binding storage — COMPLETE

The alpha-37 post-string allocation survey selected the environment subsystem:

```text
NewCallEnv   ~979.7 MB
Env.Set      ~155.0 MB
```

The project separated two costs:

1. hot physical Env footprint;
2. ordinary local-binding storage.

Cold module/source/shape metadata moved behind a lazy sidecar. Measurement then
showed that 88.9% of logical call environments acquire no ordinary locals and,
among stateful call environments, about 96% peak at four or fewer locals.

That evidence admitted a lazy external four-binding block, promoting to a Go map
on the fifth distinct ordinary local.

Final three-level self-host:

```text
                         alpha-37         final
elapsed                  9.2246 s         8.7924 s
allocated                2.396 GB         1.928 GB
mallocs                  50.297 M         49.812 M
GC cycles                79               64
```

Approximate environment-wave change:

- elapsed: -4.7%;
- allocated bytes: -19.5%;
- mallocs: -1.0%;
- GC cycles: -19.0%.

Final realization evidence:

```text
logical_call                 4809241
call_local_max_0             4276125
call_local_max_1              338033
call_local_max_2              119289
call_local_max_3_4             54570
call_local_max_5_plus          21224
call_compact_allocations      526837
call_map_promotions            21224
```

The promotion count exactly matches the measured five-plus population.

PDP-11 Cut 7 10x remained materially flat/slightly improved:

```text
elapsed      1.545114031s
allocated    1.016470360 GB
mallocs      16.440547 M
GC cycles    33
```

Driver:

> **Scope and binding semantics are authoritative. Environment storage is
> negotiable.**

### Updated cumulative self-host improvement

Across the runtime optimization sequence:

```text
                         original          current
elapsed                  19.3947 s         8.7924 s
allocated                28.291 GB         1.928 GB
mallocs                  110.604 M         49.812 M
GC cycles                899               64
```

Approximate cumulative improvement:

- elapsed: 54.7% lower;
- allocated bytes: 93.2% lower;
- mallocs: 55.0% lower;
- GC cycles: 92.9% lower.

The sequence now exposes four distinct realization pathologies:

1. call/runtime work reduced allocation **frequency**;
2. adaptive persistent lists reduced container allocation **volume**;
3. immutable string observation removed whole-value **observation
   materialization**;
4. environment realization reduced hot-frame **object footprint** and
   local-binding **storage volume**.

The common discipline remains:

> Optimize realization beneath the semantic boundary rather than weakening the
> semantic boundary to obtain performance.
