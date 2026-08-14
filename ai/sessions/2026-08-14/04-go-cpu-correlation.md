# 04 - Go CPU correlation

Status: **GATED**

## Intent

Correlate the semantic account with Go's realization of the same computation so the HAL boundary is inspectable rather than presenting two unrelated profilers.

## Implemented

Go `runtime/pprof` labels carry Aiki execution context across semantic and substrate regions. The correlation dimensions include Aiki layer, function, primitive, file, and source location where applicable.

The profile CLI emits CPU profiling artifacts alongside the Aiki semantic report.

## Evidence

Artifact-level checks confirmed:

- an evaluator-loop workload retained semantic/work/source labels;
- an append-heavy workload retained `function=append` across the boundary and, in one fresh sample, split approximately 82.5% substrate / 17.5% semantic CPU;
- substrate samples identified `_append` at `<prelude>:71`;
- spawned work retained `worker`, `send`, `recv`, and `spawn` identities, confirming practical goroutine label inheritance/restoration.

The percentages above are sample observations, not constants.

## Allocation limitation

Go allocation profiles do not carry these Aiki pprof labels. The surface therefore distinguishes:

- CPU profile: sampled and Aiki-label correlated;
- allocation interval totals: measured for the execution interval;
- Go alloc profile: useful Go allocation hotspots, but not source-label correlated.

Documentation was corrected to remove a stale claim for a nonexistent `frees` interval counter.

## Validation

- unit test verifies all cached pprof correlation dimensions;
- substrate and runner tests - PASS;
- fresh CPU artifact inspected with `go tool pprof -tags` - PASS.

## Next step established

Run a representative Aiki-owned workload sweep to establish a developmental baseline before making optimization decisions.
