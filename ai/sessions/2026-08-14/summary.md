# Session summary - 2026-08-14

## Purpose

This session turned Aiki profiling from a narrow operation-counting proposal into a first-class computational-visibility facility. The guiding idea became that Aiki should make execution apprehensible at two boundaries: what computation occurred in Aiki, and what the Go substrate did to realize it.

## How the design developed

The work began with exact semantic counters intended to support experiments such as `profile.experiment(size, function)` and empirical complexity analysis. That quickly broadened into two complementary views:

- a deterministic semantic account of Aiki work;
- a sampled/measured substrate account of realization cost.

The important architectural claim is the correlation between them. Aiki source, evaluator activity, HAL transitions, and Go CPU samples should be relatable without collapsing unlike measurements into a single synthetic cost.

The resulting interpretation is:

```text
Aiki semantic counts     exact/deterministic for the measured execution
CPU correlation          sampled/statistical, with Aiki pprof labels
interval allocations     measured quantity for the profiled interval
Go allocation profile    Go hotspots; not Aiki-label correlated
```

This preserves an important distinction: semantic counts describe what the computation did; CPU and allocation data describe resources consumed in realizing it.

## Important implementation discoveries

### Attribution can distort the computation it observes

The first source-aware implementation repeatedly split the entire source string into lines while recording semantic sites. Profiling then reported large amounts of work caused by the profiler itself. Caching source lines in each environment removed that distortion. After the fix, attributed execution still costs more than counts-only execution, but no longer manufactures dominant allocation noise.

### Profiling exposed a deeper interpreter concurrency bug

Race testing showed that spawned execution could share mutable Aiki call-stack state with the parent through lexical/prelude environments. This was not merely a profiler issue. Function invocation was conflating lexical bindings with dynamic execution state.

The correction introduced a clean split:

- lexical bindings come from the function's defining environment;
- dynamic call stack, stack limit, and profiling probe come from the caller.

This made spawned execution genuinely stack-isolated while preserving closures.

### Channel accounting needed a happens-before rule

For unbuffered sends, incrementing the send counter after the handoff allowed a receiver to continue and inspect the profile before the send had been recorded. The send event is now counted before the handoff. Receiving a value therefore implies that the corresponding send is already visible to the profile.

### Go profiling has an asymmetric correlation boundary

CPU profiling retains Aiki pprof labels, so sampled CPU work can be grouped by Aiki function, file, line, layer, and HAL primitive. Go allocation profiles do not retain those labels. Allocation support therefore remains deliberately split between measured interval totals and ordinary Go allocation hotspots. The implementation does not claim source-correlated allocation profiling.

## Resulting Aiki surface

The profile module now supports ordinary Aiki-data results for:

- `profile.counts(function)`
- `profile.measure(function)`
- `profile.experiment(size, function)` and size sweeps
- `profile.complexity(results)`

Complexity classification is empirical. It reports observed growth consistent with candidate families rather than claiming to prove asymptotic complexity.

## Developmental baseline

A representative sweep now exercises evaluator loops, proper tail recursion, persistent list append, store/bits, regex/FFI, and synchronized concurrency. The first reliable baseline already separates qualitatively different costs:

- persistent `append` is strongly substrate/allocation dominated;
- store/bits is much more heavily dominated by semantic dispatch/call-environment work;
- regex crosses both semantic and substrate layers;
- plain loops and tail recursion mainly expose evaluator cost.

Single-run CPU percentages vary enough that they are snapshots, not stable performance constants. Optimization decisions should use repeated runs on the normal development machine.

## Working-method outcome

The session also established the repository-level `ai/` work record. Substantial AI-assisted implementation should use one authoritative tree, small serial cuts, frequent progress reports, evidence-gated milestones, a current restart index, and a curated session summary. The work record is part of the delivered cut rather than disposable chat scaffolding.

## End state

The profiling work is implemented and gated in this working tree. No additional profiling architecture is required before review. The next practical step is to rebuild with the normal Go 1.24 toolchain, rerun the standard validation workflow and profiling sweep several times, then use repeated evidence to decide whether any optimization is justified.
