# Milestone 10 — runner argument regression and lint cleanup

Status: ACTIVE

## Trigger

Post-merge validation exposed two independent issues:

1. `go test ./...` failed in `TestRunExposesProgramArguments` with
   `unknown shape: error` at `system.args()`.
2. `./aiki lint ./...` reported module-field naming warnings for the Thompson
   reconstruction (`runtime.CODE`, `runtime.FAIL`, etc.) and
   `instruction.ADDRESS_SHIFT`.

## Finding — runner test context

The CLI plumbing added in Phase IV was not the source of the test failure.
The regression test executes from the `engine/runner` package directory. The
standard module-root policy therefore could not discover the repository
`lib/`, so `import("system")` returned a shaped error. Field access on that
error then produced `unknown shape: error`.

Existing runner module tests already establish repository-root context with
`withRepoRoot(t)`. The program-argument regression now does the same. This
makes the test exercise the intended ordinary runner path with the standard
library available.

## Finding — Aiki module naming

The uppercase names in the reconstruction are useful internal historical
labels (`CODE`, `CNODE`, `NNODE`, `CLIST`, etc.), but exporting those spellings
as Aiki module fields conflicts with the language's snake_case field style.
The runtime now retains the uppercase locals internally and exposes
snake_case aliases at the module boundary. The instruction module does the
same for `ADDRESS_SHIFT` / `address_shift`. External experiment references
were updated to the module-facing snake_case names.

This preserves Thompson/IBM terminology inside the reconstruction while
making the Aiki surface idiomatic.

## Validation

Source inspection confirms no remaining `runtime.<UPPERCASE>` or
`instruction.<UPPERCASE>` field accesses in the experiment.

The targeted Go regression could not be executed in the assistant container:
Go attempted to download toolchain 1.24 and network access is unavailable.
Therefore this milestone remains ACTIVE pending authoritative local
validation.

## Required local gate

Because `engine/runner/program_args_test.go` is Go code, rebuild before the
Aiki experiment validation:

```text
go test ./...
./aiki lint ./...
make
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Expected: Go suite passes; the reported field-naming warnings disappear; all
four experiment corpora and demonstrations remain clean.
