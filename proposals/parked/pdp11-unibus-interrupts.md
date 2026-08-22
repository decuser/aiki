# PDP-11/40 UNIBUS Interrupt Fabric

## Status

**PARKED — unfinished Experiment 004 work; superseded as an active development
line by the clean Experiment 005 reconstruction proposal**

## Baseline

Release baseline: `v0.4.0-alpha-39`, commit `bffff5c`.

Development branch at suspension: `pdp11-unibus-interrupts`.

The working tree was intentionally dirty with the active PDP-11/V6 Cut 10
development set. This proposal records the unfinished interrupt project exactly
as historical Experiment 004 evidence. It is not an implementation source for
Experiment 005.

## Driver at suspension

Live V6 reached `iinit()` at `005572`. Its `bread(rootdev, 1)` call entered at
`005634` and did not return to `005640`. While the kernel slept in the block-I/O
path, RK11 state showed a completed block-1 transfer with no controller error,
exhausted word count, and RKDA advanced to block 2.

Source inspection found the architectural defect: the Experiment 004 RK11 model
only marked the controller READY on completion and owned no interrupt request;
UNIBUS interrupt projection and acknowledge were hard-wired to the KW11-L clock.

## Device-document contract discovered during 004

- KW11-L: vector `000100`, BR6.
- RK11-D: vector `000220`, BR5; IDE-enabled completion requests service.
- TM11/TU10: vector `000224`, BR5 with documented ready/error/rewind interrupt
  conditions.
- KL11: receiver vector `000060`, transmitter vector `000064`, both BR4, with
  independent requests and receiver priority on a simultaneous request.

These facts are evidence to be re-derived from the DEC manuals in Experiment
005. This proposal does not grant its implementation admission to 005.

## Architectural invariant discovered

> Interrupt-causing conditions belong to devices; selection among pending
> requests belongs to the UNIBUS; acceptance and vector entry belong to the CPU.

The unfinished project intended to add device-owned request state and a generic
UNIBUS arbiter. That work is intentionally not completed here.

## Unfinished cuts at suspension

### Cut 1 — RK11 completion + generic arbiter

**SUSPENDED / NOT GATED**

The intended work was RK11-owned BR5/vector-220 request state, generic UNIBUS
arbitration, source-specific acknowledge, and clock-vs-RK priority witnesses.
Focused diagnostics had been developed during the branch, but the project was
stopped before repository reconciliation and live V6 acceptance.

### Cut 2 — KL11 receiver/transmitter interrupts

**NOT STARTED**

### Cut 3 — TM11 completion interrupts

**NOT STARTED**

## Disposition

Experiment 004 is being archived as a record of integration discoveries. Do not
resume this proposal on the Experiment 005 development line and do not copy its
PDP-specific implementation into Experiment 005.

The controlling successor proposal is:

`proposals/active/pdp11-40-reconstruction.md`
