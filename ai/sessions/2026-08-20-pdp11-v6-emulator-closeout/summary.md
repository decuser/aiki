# Summary — PDP-11/40 V6 emulator effort

The PDP-11/40 experiment began as an attempt to run Research UNIX V6 on a machine
implemented in Aiki. It grew into a useful architectural stress test for the
language: instruction dispatch, addressing modes, PSW behavior, UNIBUS traffic,
DMA, tape, disk, console I/O, observer instrumentation, and the interactive
monitor all exercised parts of Aiki that smaller programs do not pressure in the
same way.

Functionally, the reconstruction reached a credible machine core. The CPU and
addressing-mode work was gated in increasingly broad cuts; TM11/TU10 could DMA a
512-byte record using the historical six-word bootstrap; RK11/RK05 and KL11 paths
were implemented and tested; the monitor and observer surfaces could operate in
octal-first PDP-11 terms. The effort did not, however, complete the entire V6
installation-and-disk-boot objective at a satisfactory interactive speed.

Profiling changed the character of the work. Early results showed that the cost
was not primarily PDP arithmetic. A small number of guest instructions expanded
into a very large number of Aiki calls and allocations. Successive cuts removed
avoidable work: PSW delta scans where flags could not change, reads of irrelevant
condition bits, repeated audit lookup, dead hot-path parameters, interpreted
wrappers that merely forwarded to substrate builtins, redundant state facades,
and unnecessary UNIBUS dispatch on ordinary RAM.

A more important discovery was that the real emulator was not using the same
bounded execution path being optimized in the synthetic profile. The interactive
monitor returned through its host loop after every single guest instruction.
Changing the live path to execute bounded 128-instruction slices finally made the
performance work relevant to real `tmrk` execution. That required correcting KL11
output ownership as well: a one-character host latch was only safe while the host
drained output every instruction, so KL11 now queues transmitted characters after
they leave TPB.

Those changes roughly halved the delivered 10x semantic call count, from about
1.041 million calls to about 542 thousand. The machine nevertheless remains slow
on the real standalone V6 `tmrk` path.

That result is the reason to stop rather than a reason to keep shaving the PDP
implementation. The experiment has already identified a system-level problem:
Aiki's execution/runtime mechanics can impose enough per-operation cost to
dominate a systems workload even after local call structure is cleaned up.
Further PDP-specific tuning would increasingly risk collapsing useful
architectural boundaries for the sake of one benchmark.

The most important deferred issue is `store`. The intended abstraction is
isolated mutable mapped memory, but the current runtime permits a store to cross
`spawn` and therefore synchronizes accesses. If the intended isolation contract
is restored or made explicit enough to guarantee single authority, per-access
mutex locking becomes an implementation error rather than a required semantic
cost. That question belongs in the next system project, not in the PDP emulator.

The experiment therefore closes in a deliberately honest state: functionally
substantial, architecturally useful, measurably improved, still too slow, and now
serving as evidence and a regression workload for runtime-level work.
