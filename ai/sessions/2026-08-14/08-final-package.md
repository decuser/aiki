# Milestone 08 - final package

Status: **GATED**

## Intent

Close the 2026-08-14 profiling/computational-visibility session as one coherent source delivery while preserving the AI work record that made the session restartable.

## Included

The final source archive contains the authoritative `/mnt/data/aiki-prof` working tree, including:

- all profiling implementation and tests;
- `lib/profile` and profile CLI support;
- source attribution and dynamic-call-state corrections;
- CPU correlation support;
- profiling sweep, baseline, and documentation;
- the complete `ai/` working-method and session record;
- `summary.md`, which preserves the conceptual/design narrative separately from engineering milestones.

## Excluded from the source archive

- `.git/` repository metadata;
- the top-level `aiki` executable, because it predates the final source edits and was intentionally not replaced by the disposable Go-1.23/stub harness build.

No disposable dependency stubs or offline compile-harness files are part of the authoritative working tree or final archive.

## Validation inherited by this package

The final implementation gate completed before packaging:

```text
go test ./...                 PASS
Aiki tests                    406 / 406
behavior smokes               34 / 34
grammar coverage              32 / 32
engine gold inputs            10 / 10
race checks on concurrent
profiled packages             PASS
```

The final packaging edits affect only the `ai/` work record and do not modify executable Aiki/Go code.

## Delivery state

This milestone closes the session. The source should next be rebuilt and normally validated on the Go 1.24 development machine, followed by repeated profiling sweeps before making optimization decisions.
