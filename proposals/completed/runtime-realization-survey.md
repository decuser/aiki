# Proposal: Runtime Realization Survey after alpha-36

## Status

**COMPLETE**

## Baseline

Release baseline: `v0.4.0-alpha-36`, commit `875bdf0`.

The completed runtime-optimization wave established the current authoritative
three-level self-host baseline:

```text
elapsed      10.619609147s
alloc_bytes  8901681936
mallocs      50401888
gc_cycles    267
```

The semantic workload is unchanged from the earlier runtime baselines:

```text
arithmetic   1289910
comparison   1911310
call         8701738
iteration    1005458
index        1443452
```

The PDP-11 Cut 7 10x regression witness remains:

```text
elapsed      1.569385077s
alloc_bytes  1040367792
mallocs      16463466
gc_cycles    34
```

## Context

Two materially different runtime pathologies have already been established and
resolved:

1. call/runtime work reduced allocation **frequency**;
2. adaptive persistent lists reduced allocation **volume**.

The cumulative self-host improvement from the post-Number baseline is about
45.2% lower elapsed time, 68.5% lower allocated bytes, 54.4% fewer mallocs, and
70.3% fewer GC cycles for unchanged semantic work.

The demonstrated representation principle is:

> **Semantic properties are authoritative. Physical representation is negotiable.**

with the established drivers:

> **The number is exact. The representation is negotiable.**

> **The list is immutable. The representation is negotiable.**

This proposal does **not** infer a third adaptive representation by analogy.
The next family must be selected from allocation evidence.

## Question

What owns the remaining approximately 8.9 GB of allocation volume and 50.4
million allocation events in the three-level self-host workload?

The survey must distinguish at least:

- direct allocation-space ownership;
- direct allocation-object ownership;
- cumulative caller paths where useful;
- self-host-specific costs from costs also present in the PDP systems witness.

## Candidate families

Static inspection suggests several plausible families, but none is presumed to
be the answer:

- string/rune realization;
- AST/value construction;
- environments/maps that genuinely carry state;
- evaluator/parser traversal structures;
- remaining call/value wrappers;
- bytes/store/FFI or other substrate families.

String/rune work is a strong static candidate because several hot primitives
materialize `[]rune`, but this observation is not sufficient to justify an
adaptive string proposal.

## Survey evidence

Use Go allocation profiles because this phase asks **where** allocation occurs.
Aiki's measured-interval counters continue to answer **how much** allocator
activity occurs.

Primary witness:

```text
extra/profiling/selfhost-three-level.ai
```

Regression/comparison witness:

```text
experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_loop.ai
AIKI_PDP_PERF_SCALE=10
```

For each witness collect:

```text
profile --counts -allocs
pprof alloc_space, flat
pprof alloc_space, cumulative
pprof alloc_objects, flat
pprof alloc_objects, cumulative
```

Allocation profiles are ordinary Go process allocation-site profiles; they are
not claimed to carry Aiki source-label attribution.

## Attribution rule

Before proposing another representation, group the largest profile sites into
semantic realization families and account for a substantial majority of the
remaining cost.

Target:

- explain roughly 70–80% of self-host allocation bytes, where the profile makes
  that practical;
- explain a substantial fraction of allocation objects;
- distinguish costs shared with PDP from costs amplified specifically by
  self-host parsing/evaluation.

Do not optimize an isolated Go function merely because it appears high in one
view. The goal is to identify a coherent semantic realization family.

## Delivery plan — one critical stop

This is an investigation, not an implementation wave.

Proceed continuously through:

1. save the alpha-36 baseline and governing runtime results;
2. collect self-host and PDP allocation profiles;
3. classify top allocation sites by semantic family;
4. reconcile flat and cumulative views;
5. identify the dominant remaining family or conclude that cost is diffuse;
6. record rejected candidates as evidence, not as unfinished proposals.

### Critical Gate — family selection

Stop once the evidence is sufficient to answer:

> Which semantic realization family, if any, justifies the next bounded
> optimization proposal?

At that gate present:

- the dominant allocation-space families;
- the dominant allocation-object families;
- whether self-host and PDP agree or differ;
- the proposed next driver, if one is justified;
- explicit reasons not to pursue the other leading candidates.

No implementation change belongs in this survey unless required to make the
measurement itself truthful.

## Rejection condition

If no coherent family owns enough cost to justify architectural complexity,
do not manufacture an adaptive representation. Close the survey with the
finding that remaining cost is diffuse and choose smaller local work or stop
optimizing.

## Working rule

> **Measure the next semantic boundary; do not infer it from the previous one.**


## Survey result — family selected

Completed attribution on `v0.4.0-alpha-36`.

Three-level self-host allocation space is dominated by one site:

```text
evalStringIndex    6171.30 MB    73.20%
```

Inspection shows whole-string `[]rune` materialization for one-rune indexing.
The next project is therefore **Immutable String Observation Realization**.

This is intentionally narrower than Adaptive String Representation. The initial
correction keeps flat UTF-8 and removes observation-time materialization. An
alternate representation requires new post-fix evidence.

The PDP witness independently exposes parser failure bookkeeping as its dominant
allocation family. That concern is recorded separately and is not mixed into
the self-host-driven string project.
