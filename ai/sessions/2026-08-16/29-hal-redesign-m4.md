# Milestone 29 — HAL redesign implementation M4

Status: **ACTIVE — compile/test checkpoint required**

## Purpose

Separate lexical/tooling scope from executable authority and eliminate the known
spawn authority leak without changing the Aiki programmer-facing library surface.

## Implemented

- added immutable `value.Authority`, an explicit set of raw runtime primitive grants;
- `Env` now carries authority independently of `Scope`;
- ordinary function calls inherit definition-bound authority from the lexical environment;
- isolated spawn construction can use prelude vocabulary as its lexical outer while carrying the spawned function definition's independent authority;
- evaluator builtin lookup now passes `Authority`, not `Scope`, to `RuntimeContract`;
- `GoRuntime.HasBuiltin/GetBuiltin` resolve raw primitives only when explicitly granted;
- constitutive `import`/`use`/`export` remain available as language intrinsics without raw grants;
- trusted-source authority policy is explicit and exact per Aiki source; merely living beneath `lib/` grants nothing;
- prelude authority is the 32 raw primitives actually referenced by `prelude.ai`, not universal substrate privilege;
- standard library implementations receive only their declared raw dependencies;
- `selfhost/bootstrap.ai` is intentionally broad because it captures the substrate binding table for the independent interpreter;
- module evaluation assigns authority from runtime policy rather than deriving privilege from `ScopePrelude`;
- top-level user/REPL environments explicitly carry no raw authority.

## Executable/static couplings

- authority policy tests require narrow grants and prove unlisted `lib/` paths receive none;
- a source-coupling invariant reads every declared trusted Aiki source, intersects identifiers with registered raw primitives, and requires the declared grant set to match exactly (no missing or surplus privilege);
- environment tests prove `ScopePrelude` alone confers no authority and isolated environments can retain prelude vocabulary without inheriting prelude grants;
- existing boundary/contract tests were rewritten to test explicit authority instead of scope-derived privilege.

## Validation status

`gofmt` and `git diff --check` are clean. A Python structural check also confirmed the declared trusted-source grant sets match current registered raw dependencies, including all 85 bindings intentionally captured by `lib/selfhost/bootstrap.ai`.

The delivery environment cannot execute Go 1.24 tests. Because M4 changes the evaluator/runtime contract signature and authority propagation, `go test ./...` on the authoritative tree is required before M5 is stacked on top. Full `make validate` is not required at this checkpoint.
