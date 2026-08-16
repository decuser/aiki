# Milestone 02 — Phase I / Cut 1: representation and state

Status: GATED

## Intent

Establish the machine representation independently of instruction semantics.

## Implementation

Experiment 002 now contains `experiment/machine.ai`, which provides:

- 36-bit word normalization;
- 15-bit address/index normalization;
- 32,768 zero-initialized mutable memory words;
- zero-initialized PC, AC, XR1, XR2, XR4, XR6, and XR7;
- width-preserving memory and register accessors; and
- rejection of unsupported XR3/XR5 tags rather than aliasing them.

The machine value is a stable two-element container around a memory `store` and
private register-state `store`. The stores provide explicit mutation without
requiring a new Aiki mutable record facility.

## Validation evidence

A disposable copy of the repository was used for source-level validation with
only the `go` directive lowered from 1.24.0 to 1.23.0. No source changes from
that copy are retained.

Passed:

```text
go test ./engine/syntax -run TestThompsonSourcesParse -v
    machine.ai       lex/parse PASS
    machine_test.ai  lex/parse PASS

go test ./cmd/subcommands/tools/treecheck -v
    PASS
```

Runtime execution is environment-blocked here. The supplied repository requires
Go 1.24.0; this environment has Go 1.23.2. A disposable lowered-toolchain build
then attempted to fetch `github.com/hajimehoshi/ebiten/v2` and
`github.com/chzyer/readline`, but network access is unavailable. No prebuilt
Aiki executable exists in the supplied tree or local environment.

## Runtime validation and gate

On 2026-08-15 the user ran `experiment/run.sh` against the repository executable
(`aiki v0.4.0-alpha-26`). The experiment completed successfully and retained:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-15-231227.746.txt
```

The run emitted one non-fatal warning:

```text
warning: shadowing prelude function 'read'
```

This warning does not invalidate the Cut 1 semantics, but it is recorded for
cleanup before or during Cut 2.

GATED. Source/structural checks and runtime execution now jointly establish the
Phase I / Cut 1 representation-and-state invariant.
