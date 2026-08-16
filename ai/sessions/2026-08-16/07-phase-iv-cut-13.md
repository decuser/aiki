# Phase IV — Cut 13: interruptible execution and operator completion

Status: ACTIVE

## Intent

Complete the SIMH-like operator layer promised by the project charter without
expanding the emulated IBM 7094 architecture. This cut adds asynchronous
operator run/stop control, explicit status, logging, operator loading, and
register modification. It also applies a focused Aiki-native style pass to the
new monitor code before extending it.

## Aiki-style refinement

The monitor implementation had begun to accumulate host-language-shaped
classification logic. The Phase-IV surface is now deliberately expressed using
Aiki idioms where they clarify the computation:

- `match` for command, opcode, prefix, register, and state classification;
- pipelines for value transformations such as trim/split/octal parsing;
- explicit `if`/`while` where the operation is genuinely control flow rather
  than classification or composition.

This is a focused monitor refactor, not a mechanical rewrite of working Phase
I–III code.

## Implemented

### Asynchronous operator control

Added `operator.ai`, a monitor-side controller distinct from the machine
service. `run` starts execution and returns the interactive prompt. The
controller receives an operator stop over a dedicated Aiki channel and observes
it between machine/host steps, so stop occurs at an instruction boundary.

Architectural state remains behind the spawned machine service. The operator
controller shares only explicit monitor bookkeeping state (`active` and the
last stop/result) and interacts with the machine through `monitor.step` and
service commands.

New operator commands:

```text
run [LIMIT]
stop
status
wait
```

While the machine is running, state-mutating/inspection commands are rejected;
`stop`, `status`, `wait`, and `help` remain available.

### Register and memory loading

Added command-mediated register setters and bulk memory loading:

```text
set ic OCTAL
set ac OCTAL
set xr1|xr2|xr4|xr6|xr7 OCTAL
load OCTAL FILE
```

`load` accepts one octal 36-bit word per nonblank, non-comment line and writes
the words sequentially from the requested origin through the existing machine
service.

### Logging

Added:

```text
log FILE
log off
```

Monitor output emitted after logging is enabled is retained in the requested
file. This is operator/transcript logging; trace semantics remain observational
and unchanged.

### Cleanup

- renamed the internal `help` helper to `show_help`, eliminating the remaining
  prelude-shadowing warning;
- retained IBM-style octal presentation and the already-gated monitor views;
- kept the Phase I–III machine/runtime/compiler code untouched.

## Validation completed here

The changed Aiki sources were lexed/parsed with the repository grammar using a
disposable Go-1.23 syntax-package harness:

```text
monitor.ai       PASS
operator.ai      PASS
console.ai       PASS
monitor_test.ai  PASS
```

The full Aiki executable cannot run in this environment because the repository
requires Go 1.24 and the toolchain cannot be downloaded. Therefore Cut 13 is
not yet GATED.

## Authoritative local gate

Run:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Required evidence:

1. Phases I–III remain green.
2. Phase IV monitor corpus passes, including the new register/load and
   asynchronous-stop tests.
3. No prelude-shadowing warnings appear.
4. The scripted monitor walkthrough still completes and enters/exits cleanly.

If those pass, Phase IV can be marked GATED/COMPLETE.

## Discovery — `system.args()` was not wired through ordinary file execution

The authoritative Cuts 10–12 run entered an interactive prompt instead of
executing `monitor-demo.cmd`, even though `run.sh` invoked:

```text
aiki console.ai monitor-demo.cmd
```

Inspection showed that `system.args()` existed in the HAL and was documented as
returning arguments supplied to an Aiki program, but the ordinary CLI file path
called `runner.Run(filename)` without passing the remaining command-line
arguments. No production caller invoked `SetProgramArgs`; therefore ordinary
standalone programs always saw an empty argument list.

This is an Aiki executable-coupling defect exposed by the concrete monitor use
case, not a reason to work around arguments in the experiment.

### Correction

- `runner.Run` now accepts optional program arguments, installs them through the
  existing HAL argument facility for the duration of the run, and clears them
  afterward.
- `cmd/aiki` passes `flag.Args()[1:]` when executing a source file.
- Added `engine/runner/program_args_test.go`, which executes an Aiki file and
  verifies that three supplied arguments are visible through `system.args()`.

The runner-level Go test cannot execute in this environment because the Ebiten
dependency is not cached and network access is blocked. It must be included in
the authoritative local validation/build.

## Authoritative local gate — passed

The user rebuilt the Aiki executable and reran the full experiment. The rebuilt
binary identified itself as:

```text
aiki v0.4.0-alpha-26-1-ga79cbc6-dirty
```

The gate passed:

```text
Phase I   96/96
Phase II  59/59
Phase III 39/39
Phase IV  24/24
```

The command-file argument path also passed: `monitor-demo.cmd` executed
non-interactively and exited. Retained result:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-16-003826.232.txt
```

Cut 13 is therefore **GATED**. A prior run made before rebuilding dropped into
the interactive prompt; this was an old-binary validation error, not a remaining
monitor defect. Future drops that change Go/substrate/CLI code must explicitly
state that `make`/rebuild is required before validation.
