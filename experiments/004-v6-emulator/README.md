# Experiment 004 — PDP-11/40 V6 Lions Laboratory Emulator

This experiment builds, in Aiki, the smallest PDP-11/40 system required to
construct and run the Sixth Edition UNIX laboratory used to read John Lions'
commentary.

The governing contract is:

```text
../../proposals/pdp11-40-v6-lions.md
```

The first campaign is intentionally historical and end-to-end:

```text
Gate 1  PDP-11/40 core diagnostics
Gate 2  tape bootstrap reaches bus-visible tape operation
Gate 3  distribution tape loads standalone installer and reaches "="
Gate 4  standalone tmrk constructs one system disk
Gate 5  KT11-D memory management
Gate 6  constructed disk boots V6 to "#"
```

The implementation is not a general PDP-11 emulator. Scope is admitted by the
Lions laboratory path and by architecturally visible PDP-11/40/UNIBUS behavior.

Raw retained runs belong in `results/`; interpretation and discrepancies belong
in `analyses/`; executable reconstruction lives in `experiment/`.

## Current state

Project branch: `v6-emulator`

Current gate: **Gate 1 — ACTIVE**

Current serial cut: **Cut 3 — complete `m40.s` instruction-form execution contract and runtime audit**

Cuts 1 and 2 are GATED. Cut 3 now targets all 45 instruction forms Lions identifies in `m40.s`, plus runtime instruction/addressing-mode audit. No tape, disk, console, clock, or memory-management implementation has begun.
