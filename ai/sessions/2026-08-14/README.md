# Session 2026-08-14 - profiling and computational visibility

Status: **COMPLETE**

## Objective

Make Aiki execution apprehensible at two levels:

1. deterministic Aiki semantic work attributable to source;
2. correlated Go substrate CPU cost across the HAL boundary.

This session established the single-tree, serial-cut working method now documented in `../../README.md`.

## Session narrative

- `summary.md` - curated conceptual summary of how the profiling design and working method developed.

## Milestones

1. `01-semantic-profiling-core.md` - semantic counters and concurrency-safe probe plumbing.
2. `02-profile-module.md` - first-class Aiki profiling surface.
3. `03-source-attribution.md` - source/function/site attribution and dynamic-call-state correction.
4. `04-go-cpu-correlation.md` - pprof labels across semantic and substrate layers.
5. `05-sweep-and-baseline.md` - representative workload sweep and developmental baseline.
6. `06-whole-tree-gate.md` - final cross-tree validation and environment caveat.
7. `07-ai-work-ledger.md` - preserve the single-tree, serial-cut, restartable AI working method.
8. `08-final-package.md` - close the session and define the final source delivery.

## Current state

All eight milestones above are gated in the same working tree. The session is complete; `summary.md` preserves the conceptual narrative and this file remains the authoritative restart/index record.

The profiling facility now distinguishes:

```text
Aiki semantic counts     exact/deterministic for the measured execution
CPU correlation          sampled/statistical, with Aiki pprof labels
interval allocations     measured quantity for the profiled interval
Go allocation profile    Go hotspots; not Aiki-label correlated
```

## Important retained findings

- Source attribution originally distorted the measurement by splitting source into lines on every semantic hit; source lines are now cached in the environment.
- Concurrency validation exposed a deeper interpreter issue: function calls were conflating lexical bindings with dynamic execution state. `NewCallEnv` now takes bindings from the defining environment while stack, stack limit, and profile probe follow the dynamic caller.
- `send` accounting occurs before the unbuffered channel handoff so a receiver cannot observe completion before the send event has been counted.
- Go pprof labels correlate CPU and goroutine profiles, not allocation profiles. The implementation and documentation do not claim source-correlated allocation profiling.
- Single-run CPU percentages are snapshots, not stable performance facts; optimization decisions require repeated measurements on the normal development machine.

## Next action

No further profiling implementation is required before review. On the normal Go 1.24 development machine:

1. rebuild Aiki from the current source;
2. run the normal validation workflow;
3. rerun the profiling sweep several times;
4. review repeated measurements before deciding on any optimization such as a small-rational fast path, list representation change, environment lookup change, or concurrency tuning.

## Prior ad-hoc ledger

The former top-level `PROFILING-PROGRESS.md` was converted into this dated milestone record. The dated session is now the authoritative progress history.
