# 01 — Profiling implementation review

Status: **GATED**

## Intent

Independent full-tree validation and architectural assessment of the
profiling cut delivered in `2026-08-14/`.

## Validation performed

Environment: Go 1.24.0 linux/amd64, container with xvfb for ebiten tests.

```text
go build -buildvcs=false ./...      PASS
go test ./...                       all packages PASS
go test -race ./...                 all packages PASS (full tree)
```

No tests were skipped. The ebiten/canvas packages required xvfb but
otherwise ran without issue.

## Architectural assessment

The profiling architecture is sound:

- Two-view design (deterministic semantic counts vs. sampled substrate cost)
  is the right separation; refusing to collapse them into a synthetic number
  will age well.
- Probe interface is clean: `SemanticProbe` for cheap counts,
  `AttributionProbe` for opt-in site detail.
- pprof label correlation across the HAL boundary is the part that would be
  hardest to add later and it is already in place.
- `complexity` classifier written in Aiki rather than Go eats its own cooking.

The concurrency fixes are the real payload of the profiling session:

- `SetContext`/`currentCtx` elimination removes a data race under concurrent
  spawn.
- `ContextCallable` interface carries context with the call rather than
  storing it in shared runtime state.
- `NewCallEnv` splits lexical bindings from dynamic execution state (stack,
  stack limit, probe). This is an interpreter correctness fix.
- `NewIsolatedEnclosedEnv` gives spawned computations their own call stack.
- Send-before-handoff ordering establishes the intended happens-before.

All `Builtin` values implement both `Callable` and `ContextCallable`. The
evaluator's dispatch asserts `ContextCallable` first, so no live nil-context
panic path exists. The test framework's `callable.Call(nil)` fallback only
fires for non-Builtin callables passed to `test.run`/`test.faults`, which
in practice take the `ctx.Eval` path.

## Issues identified

1. **`value` imports `engine`** — the value model has acquired a dependency on
   profiling types. Compiles without cycle but the dependency arrow points the
   wrong way. An interface declared locally in `value` and satisfied by
   `engine` would keep the boundary clean. Will is addressing this.

2. **Store isolation exception** — `Store` has a `sync.RWMutex` per instance
   and can be captured by a closure and shared across spawns. This is
   intentional (store is explicitly "systems work that requires mutation")
   but should be documented as an explicit exception to spawn isolation.

## Next action

No further review required. The two issues above are tracked; neither blocks
other work.
