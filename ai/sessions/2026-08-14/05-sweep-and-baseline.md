# 05 - Sweep and developmental baseline

Status: **GATED**

## Intent

Measure real Aiki mechanisms systematically before optimizing implementation details from intuition.

## Workload corpus

Six rerunnable Aiki-owned drivers exercise distinct mechanisms:

- evaluator loop;
- proper tail recursion;
- persistent list append;
- store plus bits;
- regex/FFI;
- synchronized spawn/channel execution.

Concurrent workloads synchronize their measured work rather than relying on process timing.

## Sweep artifacts

The sweep records:

- semantic report;
- interval allocation totals;
- CPU pprof artifact;
- ordinary Go allocation profile;
- saved CPU tag report;
- environment manifest.

The manifest and tag report were added so a later comparison does not depend on chat memory or reconstructing the execution environment.

## Interpretation decisions

- semantic counts are exact for the measured execution;
- CPU percentages are statistical samples and can move materially between short runs;
- interval allocation counters answer how much allocation occurred during the measured interval;
- the ordinary Go alloc profile helps locate Go allocation hotspots but includes process-cumulative/startup activity and is not Aiki-label correlated.

The baseline report was regenerated after the attribution and concurrency fixes. Earlier contaminated measurements are not the baseline.

## Initial signals

The developmental sweep suggests:

- persistent list `append` is strongly substrate/allocation dominated;
- store+bits work is much more dominated by semantic dispatch/call-environment machinery;
- regex shows a genuine semantic/substrate split;
- evaluator loops and tail recursion are primarily evaluator costs.

These are hypotheses for repeated measurement, not optimization decisions.

## Next step established

Repeat the sweep on the normal development machine before changing representations or adding fast paths.
