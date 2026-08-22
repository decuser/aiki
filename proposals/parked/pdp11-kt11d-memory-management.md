# PDP-11/40 KT11-D Memory Management — Cut 10

## Status

**PARKED — unfinished Experiment 004 Cut 10; historical evidence only**

## Driver

The V6 RK05 boot now reaches `rkunix` and enters the kernel, but `main()` never
finishes physical-memory discovery. Live observation showed `UISD0 = 000006`
and a steadily advancing `UISA0`, exactly the moving user-space window used by
V6 `clearseg()`/`fuibyte()`.

The emulator had processor-visible APR/SSR addresses but no address translation;
those registers were ordinary backing storage. KT11-D is therefore the actual
next machine boundary.

## Architectural contract

The PDP-11/40 KT11-D provides:

- one Kernel and one User bank of eight PAR/PDR pairs;
- 16-bit virtual to 18-bit physical relocation in 32-word blocks;
- PDR access control, expansion direction, page length and written state;
- SSR0 abort/enable state and SSR2 virtual instruction address;
- current-mode translation for ordinary references;
- previous-mode transfer for MFPI/MTPI;
- memory-management abort through vector `0250`.

Unassigned physical I/O-page references must produce a bus timeout through
vector `0004`; they are not RAM.

## Realization rule

> **CPU addresses are virtual when KT11-D is enabled. Processor-internal
> registers and UNIBUS devices exist only after physical address formation.**

There is one authoritative translated machine-access path. The bounded profiling
executor may operate while management is disabled, but falls back to that path
when SSR0 enables KT11-D rather than maintaining a second MMU implementation.

## Gate 1

Focused tests must establish:

1. disabled-management addressing remains compatible with the PDP-11/40 high
   page projection;
2. Kernel PAR/PDR relocation reaches the intended physical word;
3. previous-mode transfer selects the User APR bank;
4. read-only writes abort and record SSR0;
5. processor PSW address `177776` is backed by the real CPU PSW;
6. an unassigned physical I/O address produces bus timeout;
7. SR2 records the virtual PC of a successful instruction fetch and is not
   updated by a failed instruction fetch;
8. installed core is distinct from 18-bit physical address capacity: the Lions
   lab defaults to 24K words (`000000..137777` bytes), with `140000` absent,
   while alternate machine configurations may supply a different core extent.

Then the live acceptance target is:

```text
boot disk 0
@rkunix
mem = ...
```

The next gate begins only after this is observed. Clock probing and user-mode
process execution are downstream.

## Explicit Cut 10 boundary

Trap and interrupt entry are certified only for Kernel-current execution in this
cut. Separate Kernel/User stack-pointer realization and User-to-Kernel
trap/interrupt entry are deliberately deferred until live V6 first reaches that
boundary. Cut 10 does not claim complete user-mode trap semantics.

## Live gate observation

The post-SR2 live `@rkunix` run reaches V6 `clearseg` and then remains in the
octal `003100..003104` loop:

```text
003100  005046  clr -(sp)
003102  006620  mtpi (r0)+
003104  077103  sob r1,003100
```

At suspension the machine reports `PC 003102`, `PSW 030344`, `IR 006620`,
`UISD0 000106`, and `UISA0 005203`. The written bit in `UISD0` is evidence that
previous-mode stores have succeeded; the failure is therefore not yet classified
as a broken MTPI implementation. Cut 10 remains ACTIVE until the exact stalled
transition is isolated and the live `mem = ...` gate is reached.

## Installed-memory correction

The live post-SR2 run disproved the earlier assumption that `003102` was a
terminal MTPI loop. V6 progressed through `clearseg()` and continued the
`main()` memory-sizing loop. The remaining false premise was treating every
ordinary physical address below the I/O page as installed RAM. The 18-bit
physical address space is capacity, not presence.

For the Lions laboratory, the configured machine has 24K words of core. The
UNIBUS therefore responds as RAM for physical bytes `000000..137777`; physical
`140000` is the first absent byte and must produce bus timeout. The backing
store remains 18-bit so bootstrap ROM and I/O-page realization are not conflated
with installed core.

## Observer and console isolation correction

Live Cut 10 execution established two host-boundary requirements that are now
part of the acceptance contract:

1. Observation must not perform translated guest memory references. The CPU
   observer displays the processor's latched IR rather than re-reading the word
   at PC, so observing a faulting mapping cannot create or modify MMU abort
   state.
2. Host terminal speed must not overwrite the KL11 receive buffer. Host
   keystrokes are queued as they arrive and delivered through the KL11 only
   while its receiver DONE bit is clear. Guest echo remains the typing
   feedback, and Return remains the guest line terminator; no host-side line
   editor is interposed.

These are emulator/monitor boundary corrections, not changes to PDP-11 or V6
semantics.

## Mode-stack architecture correction

Live V6 trap recovery reached the common `mfpi sp` path with Kernel current mode
and User previous mode. That boundary invalidated the earlier Cut 10 deferral of
separate stack pointers.

For the PDP-11/40 target, CPU state now follows these rules:

- R0-R5 and PC are shared across Kernel/User mode;
- R6 is banked as KSP and USP, selected by PSW current mode for ordinary
  addressing;
- direct `MFPI SP` reads the previous-mode SP bank and pushes it on the current
  stack;
- direct `MTPI SP` pops the current stack and writes the previous-mode SP bank;
- trap/interrupt vectors are fetched through Kernel virtual space;
- the vector PSW selects the new current-mode stack before old PS and PC are
  pushed, and the interrupted current mode is copied into the new previous-mode
  field;
- RESET outside Kernel mode is suppressed.

Focused acceptance now includes User-to-Kernel fault and interrupt entry in
addition to the original Kernel-current Cut 10 cases.

NPR/DMA memory-presence behavior is explicitly separate. Controllers must report
non-existent memory through their own error/status semantics (RK11 NXM, etc.);
CPU bus-timeout trapping is not a substitute. This remains a deferred controller
cut and does not block the current `mem = ...` gate unless live V6 reaches it.

## Console switch register correction

Live low-water tracing identified a missing processor-console register outside the KT11-D itself.
V6 `putchar()` begins with `tst *$177570`; physical `777570` is the PDP-11 console switch
register. Treating that address as unassigned I/O caused a bus timeout, whose diagnostic path
recursively called `putchar()` and consumed KSP until kernel data were overwritten.

The emulator now realizes the switch register at the CPU boundary with the conventional Lions/V6
operator value `173030`. Guest reads are supported as word and bytes, guest stores do not alter the
physical front-panel setting, and the monitor/operator may change it explicitly. This correction is
part of the live Cut 10 gate because it removes the recursive false bus-error path that prevented
V6 initialization from progressing.

## M40 instruction-contract audit correction

After the live gate reached the V6 banner it stopped on `006700` (`SXT R0`).
The existing M40 contract had incorrectly equated "direct mnemonic occurrence in
`m40.s`" with the complete instruction vocabulary assumed by its PDP-11/40
backup logic. Lions explicitly classifies SXT in the `u6` single-op/MFPI/MTPI/SXT
case even though SXT has zero direct mnemonic occurrences in `m40.s`.

The contract now distinguishes direct source counts from backup-classified
architectural forms. SXT is required by the contract and must be decoded,
executed, disassembled, and regression-tested before the live gate resumes.



## Suspension disposition

Experiment 004 is intentionally suspended before this proposal reached its live gate.
This document remains as historical evidence of discoveries made during 004 and is not
a donor implementation or controlling contract for Experiment 005. The successor
contract is `proposals/active/pdp11-40-reconstruction.md`; KT11-D is re-derived in
that proposal's Cut 7 from DEC documentation.
