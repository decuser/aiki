# Proposal: Environment Realization and Binding Storage

## Status

**COMPLETE**

## Baseline

Branch: `environment-optimization`

Release baseline: `v0.4.0-alpha-37`.

Authoritative three-level self-host baseline after immutable string observation:

```text
elapsed      9.224568502s
alloc_bytes  2395612824
mallocs      50296696
gc_cycles    79
```

PDP-11 Cut 7 10x regression baseline:

```text
elapsed      1.556346663s
alloc_bytes  1040471456
mallocs      16463422
gc_cycles    33
```

Post-string self-host allocation-space leaders:

```text
NewCallEnv             ~979.71 MB   ~42.86%
evalCallArgs           ~229.51 MB   ~10.04%
Parser.recordFailure   ~180.06 MB    ~7.88%
NewNumberFromString    ~157.01 MB    ~6.87%
Env.Set                ~155.04 MB    ~6.78%
```

This proposal owns `NewCallEnv` and `Env.Set` only. Parser failure bookkeeping,
Number parsing, and call-argument realization are separate concerns.

## Problem

The environment subsystem now owns roughly half of the remaining self-host
allocation volume.

Call-realization counters provide an important first decomposition:

```text
user_entry       4,664,042
tail_env_reuse     143,766
```

Approximately 4.52 million user invocations therefore require a fresh physical
call environment. `NewCallEnv` owns about 980 MB, or roughly 227 bytes per fresh
call environment. This points first to the physical Env footprint rather than
to maps alone.

At the same time, `Env.Set` owns about 155 MB. Whether that cost justifies a
compact binding representation depends on the actual maximum local-binding
cardinality of call and enclosed environments.

These are two distinct hypotheses and must not be conflated.

## Semantic boundary

Environment representation is not an Aiki-visible value type. The authoritative
properties are lexical lookup, dynamic execution metadata, scope, authority,
shadowing, update/delete semantics, closure capture, stack behavior, module
metadata, shape visibility, and spawn isolation.

A useful working statement is:

> **Scope and binding semantics are authoritative. Environment storage is
> negotiable.**

This is a proposal driver, not a new language slogan. It survives only if the
implementation remains observationally identical.

## Hypothesis A — hot Env carries cold metadata

The current `Env` object directly stores:

- shape map;
- file string;
- source string / cached source lines;
- module exports;
- package name.

Ordinary call environments almost never own these fields but pay their object
footprint on every physical call allocation.

The first tranche moves this metadata behind one lazy cold sidecar. The hot Env
retains only execution, binding, lexical, authority, stack, and profiling state.

This is evidence-backed before new instrumentation because `NewCallEnv` itself
dominates allocation space.

### Constraint

Do not replace the removed fields with an equally large inline optimization.
The rejected call-argument Cut 3 already demonstrated that enlarging every Env
to save allocations elsewhere is a losing trade.

## Hypothesis B — local maps may be overpowered

`Env.Set` allocates a Go map on the first ordinary local binding.

Do **not** add inline binding slots yet.

First measure maximum local-binding cardinality per logical environment
lifetime:

```text
0
1
2
3-4
5+
```

Call parameters borrowed through `callParamNames/callParamValues` are excluded;
the histogram measures only actual `store` bindings.

A logical call lifetime is distinct from a physical Env allocation. Tail-env
reuse begins a new logical call lifetime while reusing one physical Env object.

If the histogram shows that the overwhelming majority of stateful environments
peak at a small cardinality, a later tranche may introduce compact **external**
binding storage that promotes to a map. It must not fatten every Env.

If the histogram does not support that representation, stop after the cold
sidecar.

## Instrumentation

`aiki profile --counts` adds:

```text
Environment realization
  physical_call
  logical_call
  physical_enclosed
  physical_isolated

  call_local_max_0
  call_local_max_1
  call_local_max_2
  call_local_max_3_4
  call_local_max_5_plus

  enclosed_local_max_0
  enclosed_local_max_1
  enclosed_local_max_2
  enclosed_local_max_3_4
  enclosed_local_max_5_plus

  call_map_promotions
  enclosed_map_promotions
  call_local_new
  call_local_update
  enclosed_local_new
  enclosed_local_update
```

The cardinality buckets are derived from threshold crossings and therefore
describe the **maximum ordinary-local binding count reached during one logical
environment lifetime**, not the final count after deletes.

Profiling is observation only and must not alter semantic behavior.

## Tranche A — footprint + measurement

Proceed continuously:

1. move cold Env metadata behind a lazy sidecar;
2. preserve all inheritance/reset/module/shape semantics;
3. add physical/logical environment realization counters;
4. add local-binding maximum-cardinality counters;
5. add focused value/evaluator tests;
6. reconcile alpha-37 string completion evidence in durable results/history;
7. run whole-tree validation and the self-host/PDP witnesses.

### Critical Gate 1

Stop only when Tranche A is complete.

Gate 1 decides two things:

1. whether the cold sidecar survives on correctness/performance evidence;
2. whether compact binding storage is justified by the measured histogram.

If compact binding storage is not overwhelmingly supported, this proposal may
close after Tranche A.

## Possible Tranche B — compact binding storage

Only if Gate 1 earns it.

A possible representation is:

```text
no ordinary locals
      |
      v
small external binding block
      |
      | exceeds measured threshold
      v
Go map
```

The small block is external to Env so zero-local environments do not pay its
size.

Exact capacity is determined by the histogram, not by taste.

Required semantics include:

- lexical shadowing;
- `Set`, `Update`, `Delete`, `HasOwn`;
- delete revealing outer bindings;
- closure capture;
- tail reset/reuse;
- match/select enclosed scopes;
- module environments;
- spawn isolation;
- shape/module cold metadata independence.

### Critical Gate 2

If Tranche B exists, final survival requires:

- `make validate` PASS;
- unchanged Aiki semantic/call counts;
- materially lower self-host allocation bytes and/or malloc count;
- favorable or materially flat elapsed time;
- PDP materially flat;
- no representation leakage.

## Rejection discipline

Do not rescue weak environment storage by adding:

- environment pooling across escaping closures;
- global symbol IDs;
- compiler escape analysis;
- hash-consed scopes;
- frame arenas;
- parser-specific environment shortcuts.

Those are separate projects.

The purpose of this proposal is to remove cold physical footprint and admit
compact binding storage only if real binding cardinality earns it.


## Critical Gate 1 evidence — PASSED

Authoritative Tranche A self-host:

```text
elapsed      8.864493056s
alloc_bytes  2022264336
mallocs      50296557
gc_cycles    68
```

Compared with alpha-37:

```text
elapsed      9.224568502s -> 8.864493056s
alloc_bytes  2395612824   -> 2022264336
```

The cold metadata sidecar therefore survives.

Authoritative logical call-binding maxima:

```text
logical_call                 4809241
call_local_max_0             4276125
call_local_max_1              338033
call_local_max_2              119289
call_local_max_3_4             54570
call_local_max_5_plus          21224
```

Interpretation:

- about 88.9% of logical call environments acquire no ordinary locals;
- among environments that do acquire locals, about 96% peak at four or fewer;
- only about 0.44% of all logical call lifetimes reach five or more locals.

This decisively earns Tranche B with a four-binding **external** compact block.
Zero-local environments pay no compact-storage allocation and Env itself is not
enlarged. The fifth ordinary local promotes the external block to a Go map.

The PDP Tranche A witness remained materially flat:

```text
elapsed      1.570065039s
alloc_bytes  1021104192
mallocs      16463682
gc_cycles    33
```

One instrumentation discrepancy (`enclosed_local_max_0 = -1`) was caused by the
top-level profiled user environment being created before the probe was attached.
Tranche B explicitly registers that already-live environment after probe
attachment; the semantic implementation was unaffected.

## Tranche B implementation

Status: **COMPLETE**

The surviving binding realization is:

```text
no ordinary locals
      |
      | first Set
      v
external compact block (capacity 4)
      |
      | fifth distinct local
      v
Go map
```

Tail-reset physical frames retain the external block allocation but reset it
back to compact-empty representation, even if a previous logical invocation
promoted to map storage.

Profiling distinguishes compact-block allocation from map promotion.


## Critical Gate 2 evidence — PASSED

Authoritative final self-host:

```text
elapsed      8.792440682s
alloc_bytes  1928120376
mallocs      49812106
gc_cycles    64
```

Compared with the alpha-37 baseline:

```text
                         alpha-37         final
elapsed                  9.224568502s     8.792440682s
alloc_bytes              2395612824      1928120376
mallocs                  50296696        49812106
gc_cycles                79              64
```

Approximate change:

- elapsed: 4.7% lower;
- allocated bytes: 19.5% lower;
- mallocs: 1.0% lower;
- GC cycles: 19.0% lower.

The final call-binding realization is:

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

The map-promotion count exactly equals the number of logical call lifetimes that
reached five or more ordinary locals. This confirms the four-binding promotion
boundary is being realized exactly as designed.

There are 533,116 logical call lifetimes that acquire at least one ordinary
local:

```text
4809241 - 4276125 = 533116
```

Only 526,837 external compact blocks were allocated. The lower allocation count
is expected because tail-reset physical environments may retain and reuse an
already allocated compact block across logical call lifetimes.

The final PDP-11 Cut 7 10x witness is:

```text
elapsed      1.545114031s
alloc_bytes  1016470360
mallocs      16440547
gc_cycles    33
```

Compared with the alpha-37 PDP baseline, elapsed and allocated bytes are slightly
better and GC cycles are unchanged. The systems workload therefore remains
materially flat.

## Completion decision

Completed 2026-08-21.

Both environment hypotheses survived:

1. cold module/source/shape metadata belongs behind a lazy sidecar rather than
   in every hot Env object;
2. ordinary local bindings benefit from lazy external compact storage with a
   measured four-binding capacity and promotion on the fifth distinct local.

No additional environment pooling, arenas, symbol IDs, compiler escape
analysis, or frame specialization is admitted.

The architectural driver survives the evidence:

> **Scope and binding semantics are authoritative. Environment storage is
> negotiable.**

The representation remains subordinate to lexical lookup, shadowing, update,
delete/reveal, closure capture, tail reuse, module metadata, authority, and
isolation semantics.
