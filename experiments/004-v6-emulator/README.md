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

Current gate: **Gate 2 — ACTIVE**

Current serial cut: **Cut 4 — raw V6 tape and six-word bootstrap**

Cuts 1–3 are GATED. Cut 4 adds the raw TUHS V6 tape media contract, the UNIBUS I/O-page projection, and only the tape-controller behavior required by the historical six-word bootstrap. Disk, console, clock, and memory management remain unimplemented.
