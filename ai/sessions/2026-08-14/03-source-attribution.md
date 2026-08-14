# 03 - Source attribution

Status: **GATED**

## Intent

Tie semantic work to Aiki source, function identity, and operation detail while keeping the cheap counts-only path available.

## Implemented

- attributed semantic sites;
- file/line/column/function/detail identity;
- repeated-site aggregation;
- source-line caching in `Env` to avoid splitting the entire source on each semantic hit.

Exact attribution tests verify:

- call-site detail names the callee;
- function-body work carries the function name;
- source line and source text are correct;
- repeated work aggregates to the expected site count.

## Discovery: profiler overhead contaminated the first measurements

The first sweep showed source-line extraction dominating allocation profiles. Attribution was repeatedly splitting source strings into lines. Caching lines once in the environment removed that artificial allocation load.

A 250,000-iteration sanity check after the fix measured roughly 361 ms for counts-only and 440 ms with attribution, with essentially identical allocation totals. Attribution still has time cost, but it no longer manufactures dominant allocation noise.

## Discovery: lexical and dynamic execution state were conflated

Race testing exposed a deeper interpreter defect. Function calls obtained their dynamic stack through the function's lexical environment. Parent and spawned execution could therefore reconnect to shared mutable stack metadata through prelude functions.

The correction introduced `NewCallEnv` with an explicit split:

- lexical bindings come from the function's defining environment;
- stack, stack limit, and profiling probe come from the dynamic caller.

This is an interpreter correctness improvement, not merely a profiling workaround.

## Discovery: send accounting ordering

`send` originally incremented its counter after an unbuffered send completed. A receiver could resume and inspect the profile before the sender recorded the event. Successful-send accounting now occurs immediately before the channel handoff, establishing the intended happens-before relation.

## Validation

- race checks for runner, substrate, evaluator, and value packages - PASS;
- `go test ./...` - PASS;
- Aiki-native tests - 406/406 PASS;
- grammar coverage - 32/32 PASS.

## Next step established

Carry Aiki execution identity across the HAL boundary into Go CPU profiling without pretending sampled CPU data is deterministic semantic measurement.
