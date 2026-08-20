# 01 - Close PDP-11/40 V6 emulator performance effort

Status: **GATED**

## Intent

Reconcile the PDP-11/40 V6 emulator effort against what was actually achieved and
close the project without turning remaining systemic runtime cost into another
round of PDP-specific optimization.

## Evidence considered

The final live-fast-path implementation had already passed the focused monitor,
console, bounded-slice, structural, fixed-domain, architecture, CPU-extension,
and RK gates recorded in the project work.

The delivered 10x semantic profile for 7,680 guest instructions reported about
541,586 Aiki calls, down from about 1,040,917 on the immediately preceding
profiled tree.

The real standalone V6 `tmrk` path remained slow after those improvements.

## Findings

1. PDP instruction semantics are not the dominant remaining optimization target.
2. A large part of the measured cost came from general Aiki execution structure:
   interpreted forwarding calls, repeated facade/look-up work, allocation, and
   host-loop boundaries.
3. The live monitor's one-instruction host turn was a material architectural
   performance problem; bounded 128-instruction slices corrected that without
   abandoning bounded control responsiveness.
4. KL11 required an ordered output queue once host draining was no longer
   instruction-by-instruction.
5. Continued PDP-local optimization now has diminishing architectural value.
6. `store` synchronization cannot be removed safely until the language/runtime
   contract for isolation versus cross-`spawn` sharing is reconciled.

## Decision

Close the PDP emulator performance project now.

Do not claim completion of the full V6 tape-to-disk installation or RK05 boot
milestones. Preserve the experiment as a systems stress test and future runtime
regression workload.

Move the next engineering effort to the Aiki system/runtime level. The first
design question is the `store` authority model; subsequent runtime work should be
driven by measured general costs rather than by special cases in the PDP code.

## Validation

This closeout is a record-only cut. It reconciles previously gated implementation
work with the observed real-machine result and records an explicit disposition
for the remaining performance limitation and the `store` issue.

No source, binary, gold, proposal, or experiment file is modified by this
closeout overlay.

## Next action

Create a new named system-level project branch and proposal as appropriate for the
runtime contract change. Treat the PDP emulator as a measurement/regression
consumer of that work, not as its implementation site.
