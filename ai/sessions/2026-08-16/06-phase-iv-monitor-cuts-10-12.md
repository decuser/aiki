# Phase IV — monitor and observability, Cuts 10–12

Status: ACTIVE

## Intent

Restore the operator/debug environment that was part of the original Thompson
7094 project charter but was not included in the completed Phase I–III
reconstruction. The monitor is an Aiki operator environment analogous to a small
SIMH monitor; it is not emulation of the physical IBM 7094 console.

## Phase structure

Phase IV is four cuts:

1. Cut 10 — operator monitor;
2. Cut 11 — disassembly and trace;
3. Cut 12 — Thompson views and command files;
4. Cut 13 — interruptible execution, logging, and operator load.

Cuts 10–12 are implemented in the current tree. Cut 13 remains PLANNED pending
the first executable monitor gate.

## Implemented

### Machine/service boundary

- Added `machine.reset_state`, preserving memory while clearing PC/AC/XRs.
- Extended the existing spawned machine service with command-mediated AC and XR
  access, register snapshots, reset, and restoration of the prepared Thompson
  runtime/object program.
- The monitor never reaches into machine stores after handing the machine to the
  service.

### Operator monitor

Added `monitor.ai` and `console.ai` with:

- IBM-style octal addresses and 36-bit words;
- `help`, `input`, `step`, bounded `run`, `trace`;
- `show registers`, `show code`, `show clist`, `show nlist`, `show matches`;
- `examine`, `deposit`, `reset`, `do`, `quit`;
- host-aware stepping across `GETCHA`, EOF flushing, and `FOUND`;
- disassembly for the 14-instruction emulator subset.

### Replayable demonstration

Added `monitor-demo.cmd`. `run.sh` now invokes a Phase-IV test corpus and the
scripted monitor demonstration after the already-gated Phase I–III corpora.

## Validation status

Source inspection completed. A local Go/parser gate cannot run in the tool
environment because the repository requests Go 1.24 and the environment cannot
reach `proxy.golang.org` to obtain that toolchain. Therefore Cuts 10–12 are not
GATED here.

## Critical next action

On the authoritative local tree run:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

The expected new section is:

```text
--- Phase IV monitor corpus ---
PASS monitor_test.ai
```

followed by a visible `--- Monitor demonstration ---` transcript.

If that gate is clean, proceed directly to Cut 13: asynchronous operator stop,
running/halted status, logging, and operator-facing load.

## Follow-up — first authoritative monitor gate

The first local execution produced:

```text
Phase IV monitor corpus: 14 tests, 13 passed, 1 failed
monitor uses IBM-style octal presentation: parse_octal("0o1750") returned invalid octal number
```

Two warnings also reported prelude shadowing for functions named `reset`.

### Finding

Aiki string indexing yields runes, not one-character strings. The optional `0o` prefix detector compared `s[0]`/`s[1]` against string values, so the prefix was not consumed and the later digit loop rejected `o`. Plain octal such as `01750` already worked.

### Correction

- Detect `0o`/`0O` using rune code points through `ord`.
- Rename monitor `reset` to `reset_session`.
- Rename service `reset` to `reset_machine`.
- Keep the operator command named `reset`; only implementation bindings change.

Status remains **ACTIVE** pending one clean authoritative rerun.

## Follow-up — authoritative Cuts 10–12 gate

The corrected local execution passed the Phase-IV monitor corpus:

```text
14 tests, 14 passed, 0 failed
```

The monitor demonstration then entered the interactive console successfully. The only remaining warning was a helper binding named `help` shadowing the prelude function.

Cuts 10–12 are therefore **GATED**. The `help` warning is presentation/style cleanup and does not invalidate the gate.
