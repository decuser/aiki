# 01 - Semantic profiling core

Status: **GATED**

## Intent

Establish exact semantic counting at Aiki execution choke points before adding a user-facing profiling surface.

## Implemented

- engine semantic kinds and probe contract;
- atomic semantic counters;
- proper tail-call call accounting;
- `select` chosen-receive accounting;
- HAL `send`/`recv` accounting;
- store read/write accounting;
- dynamic probe propagation through function, `apply`, and `spawn` execution.

## Cleanup and decisions

Removed the old mutable `RuntimeContract.SetContext/currentCtx` compatibility state. The context-aware callable path made it dead shared state, and retaining it would create unnecessary concurrency risk.

Added negative semantic tests establishing:

- a `select` default arm counts zero receives;
- failed store access counts zero successful reads.

## Validation

- `go test ./...` - PASS in the disposable offline compile harness.
- Aiki-native tests - 400/400 PASS at this milestone.
- structural engine coverage - 32/32 grammar productions across 10 inputs PASS.

## Environment note

The repository requires Go 1.24 while the container supplied Go 1.23.2 and could not download a toolchain. Compilation used a disposable harness outside the authoritative tree with local dependency stubs. No stub code belongs to the product tree.

## Next step established

Expose the semantic measurement through an ordinary Aiki module without introducing an opaque profiler runtime value.
