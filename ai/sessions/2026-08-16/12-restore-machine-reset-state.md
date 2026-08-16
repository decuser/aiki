# Milestone 12 — restore machine reset primitive

Status: ACTIVE

## Trigger

After the spawned-relative-import engine correction, repository-wide Aiki tests
reached the Phase-IV monitor service correctly and exposed a separate source
regression:

```text
'reset_state' not exported by module 'thompson_7094_machine'
```

`service.ai` invokes `emu.reset_state(machine)` for both operator reset and
Thompson preparation. The Phase-IV Cut 10–12 record states that the machine
reset primitive had been added, but later cleanup/drop merging left the caller
without the corresponding machine definition/export.

## Correction

Restore `machine.reset_state(machine)` as the narrow machine-level operation it
was intended to be:

- clear PC;
- clear logical AC;
- clear modeled XR1, XR2, XR4, XR6, XR7;
- preserve all core memory;
- return `true`;
- export as `reset_state`.

No monitor, runtime, compiler, or Go/engine semantics are changed.

## Validation gate

This cut is Aiki source only; no executable rebuild is required.

Run:

```text
./aiki test ./...
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Expected: repository-wide Aiki corpus clean and Phase IV returns to 24/24 with
the scripted monitor walkthrough completing.
