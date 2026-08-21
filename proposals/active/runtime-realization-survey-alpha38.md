# Proposal: Runtime Realization Survey after alpha-38

## Status

**ACTIVE**

## Baseline

Release baseline: `v0.4.0-alpha-38`, commit `36d5a0e`.

Branch: `runtime-survey-alpha38`.

Authoritative three-level self-host baseline after environment realization:

```text
elapsed      8.792440682s
alloc_bytes  1928120376
mallocs      49812106
gc_cycles    64
```

PDP-11 Cut 7 10x baseline:

```text
elapsed      1.545114031s
alloc_bytes  1016470360
mallocs      16440547
gc_cycles    33
```

The cumulative runtime work has reduced self-host allocation from roughly
28.291 GB to 1.928 GB. The residual ranking must therefore be measured again;
pre-environment profile ordering is not authoritative.

## Question

Which semantic realization family, if any, owns enough of the remaining
self-host allocation space and/or object count to justify another bounded
optimization project?

The survey must distinguish:

- direct allocation-space ownership;
- direct allocation-object ownership;
- cumulative caller paths where they clarify ownership;
- self-host-specific pressure from pressure shared with the PDP systems witness.

## Method

Reuse the established allocation survey without changing its profiling method:

```text
profile --counts -allocs
pprof alloc_space, flat
pprof alloc_space, cumulative
pprof alloc_objects, flat
pprof alloc_objects, cumulative
```

The wrapper:

```sh
extra/profiling/alpha38-runtime-survey.sh
```

records the exact repository/executable baseline, captures independent semantic
counts, runs the existing pprof survey, and packages evidence under `/tmp`.

## Selection rule

Do not carry forward the old candidate order by assumption.

Classify the largest sites into semantic families and prefer a new project only
when one family explains a material share of the remaining cost and has a clean
semantic boundary.

A large isolated Go function is not sufficient evidence for architectural work.
Likewise, a site that dominates PDP but not self-host may belong to a separate
systems/parser project rather than the main runtime sequence.

## Candidate names are not commitments

Earlier profiles contained sites such as:

```text
evalCallArgs
Parser.recordFailure
NewNumberFromString
```

These names are retained only as historical context. Alpha-38 evidence may
reorder them, shrink them, or expose a different family entirely.

## Critical gate

This survey has one stop: family selection.

At the gate record:

1. dominant self-host allocation-space families;
2. dominant self-host allocation-object families;
3. the corresponding PDP view;
4. the next bounded semantic family, if one is justified;
5. explicit rejection/defer reasons for other leading sites.

If the remaining cost is diffuse, close the survey without manufacturing a new
adaptive representation.

## Working rule

> **Measure the next semantic boundary; do not inherit the previous ranking.**


## Evidence-suite expansion: Four-Way Life

The alpha-38 survey is widened before the next post-parser family is selected.

The permanent evidence suite now contains three qualitatively different
workloads:

```text
selfhost       language implementation / evaluator pressure
PDP-11         systems-emulator / interpreter-loop pressure
Four-Way Life  multiprocess / IPC / heterogeneous library pressure
```

Four-Way Life is a genuine five-process application. Profiling only
`coordinator.ai` would therefore be incomplete. The survey preserves five
independent profiles:

```text
coordinator    5 headless generations, 60x40, seed 42
worker A       one deterministic generation frame
worker B       one deterministic generation frame
worker C       one deterministic generation frame
worker D       one deterministic generation frame
```

A separate complete coordinator + four-worker run records whole-application
wall time.

The worker profiles are deliberately not merged. Their load-bearing domains are
different:

- A: file/path/string/bytes;
- B: environment/random/hash;
- C: subprocess/regex/base conversion;
- D: lists/sort/filter/reduce/shapes/math.

### Cross-workload selection rule

Prefer a runtime project when the same semantic realization problem is material
in at least two independent evidence legs.

This is a preference, not a mechanical threshold. A single-workload pathology
may still warrant work when it is extreme (for example parser failure
materialization in PDP), but a representation change must not be justified only
by one convenient benchmark.

Do not sum the five Life pprof files before classification. First preserve
process identity, identify each process's dominant semantic families, and only
then reason about cross-workload recurrence.

The expanded survey command remains:

```sh
extra/profiling/alpha38-runtime-survey.sh
```

All generated evidence remains under `/tmp`.


## Method refinement — startup versus execution

The expanded evidence suite exposed a necessary classification boundary.

Lexer tokenization, grammar traversal, module scanning, help parsing, and AST
construction are predominantly one-shot process/module startup costs. They
remain measured and reported, but they are no longer ranked mechanically
against repeatedly executed evaluator/runtime realization.

Future selection reports two families:

```text
startup
  lexer / parser / module scan / help metadata / AST construction

execution
  evaluator / value realization / call / environment / collections / HAL use
```

A startup site can still justify bounded work when it is pathological (as the
speculative parser mismatch allocation was), but ordinary AST/token
construction is not optimized merely because short-lived benchmark processes
make startup a large percentage of total allocation.

Four-Way Life is especially sensitive to this distinction because the canonical
run starts five independent Aiki processes.

Working rule:

> **Remove pathological startup materialization; optimize steady execution by
> repeated semantic lifetime.**
