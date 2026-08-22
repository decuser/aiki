# Experiment 005 — PDP-11/40 Reconstruction

## Status

**ACTIVE — Cut 0 record complete; archival Git actions pending**

## Purpose

Reconstruct a PDP-11/40 in Aiki from documented architectural contracts, one
component at a time, with each component completed and certified before the next
component is admitted.

The evidentiary hierarchy is:

> **DEC manuals define the machine.**  
> **SIMH corroborates our interpretation and implementation.**  
> **UNIX V6 validates the completed integrated machine.**

Timing, propagation delay, electrical implementation, and mechanical behavior
are outside scope except where they affect programmer-visible semantics.

## Governing admission rule

> **Experiment 005 is a clean reconstruction. No source file, function, module,
> or device implementation from Experiment 004 may be imported, copied,
> mechanically adapted, shared, or symlinked merely because it previously
> worked.**

Experiment 004 may be read as historical evidence. It may identify failure
modes, integration hazards, useful questions, and candidate regression cases.
Every behavior admitted into Experiment 005 must instead be re-derived from the
controlling DEC documentation, implemented against the current cut's contract,
and independently certified.

SIMH may corroborate behavior after the contract is established; it does not
define the contract. UNIX V6 is an integration and acceptance workload, not an
architectural specification.

Experiment 004 tests are evidence, not automatically admissible tests. A
regression discovered there may enter Experiment 005 only after it is restated
against the appropriate DEC-derived contract.

Code similarity that arises naturally from implementing the same documented
machine is not itself a violation. Provenance is the invariant: every admitted
behavior must be explainable from the Experiment 005 contract and authorities,
not from "004 did it this way."

General Aiki infrastructure outside the PDP experiments may be used normally.
PDP-specific implementation code from `experiments/004-v6-emulator/` may not be
used as a donor implementation.

## Architectural scope

Experiment 005 models a PDP-11/40 at its programmer-visible architectural
boundaries. The UNIBUS is semantic rather than electrical: preserve data
transactions, addressability, faults, initialization, priority arbitration,
interrupt vector transfer, and NPR/DMA behavior; omit wire timing and electrical
realization that software cannot observe.

The initial configured machine uses PDP-11/40 core memory and later adds the
specific devices required for the target V6 laboratory.

## Serial cuts

### Cut 0 — Suspend and preserve Experiment 004

**ACTIVE — repository records prepared; Git archival action belongs to the user**

Purpose: terminate the current integrated-emulator development line without
repairing, normalizing, or laundering its unfinished state.

Required record:

- preserve the exact current Experiment 004 source and diagnostics;
- preserve known failures and unfinished interrupt work;
- record the last proven V6 integration boundary;
- record the architectural defect that stopped the run;
- preserve Experiment 004 as historical evidence, not a donor codebase;
- archive the development line with a branch/tag chosen by the user;
- begin Experiment 005 from a deliberate Aiki baseline, not from copied 004
  implementation.

Last proven V6 boundary:

```text
main()
  -> iinit()
  -> bread(rootdev, 1)
  -> RK transfer completes
  -> no functioning RK completion interrupt reaches/wakes V6
  -> scheduler/idle
```

Source inspection established that the Experiment 004 RK11 completion path did
not own/raise a device interrupt and the UNIBUS interrupt projection was
KW11-L-specific. That is historical evidence for 005's bus/device contracts; it
is not to be repaired as part of Cut 0.

Gate:

1. Experiment 004 exact state is committed and recoverable.
2. The old development branch is archived/preserved and no longer active.
3. The active interrupt proposal is parked as unfinished Experiment 004 work.
4. The Experiment 005 proposal is the controlling PDP reconstruction proposal.
5. A new Experiment 005 baseline is created without importing 004 PDP source.

### Cut 1 — Machine skeleton and programmer's console

**PLANNED**

Establish the permanent operator-visible machine before implementing the full
machine underneath it.

Authority: PDP-11/40 system/programmer-console documentation.

Required surface includes the documented semantics corresponding to:

```text
HALT
START
CONTINUE
STEP
EXAMINE
DEPOSIT
register display
PC display/control
PSW display
switch register
RESET / INIT distinction as documented
```

Debugger conveniences may exist, but they must be visibly distinct from
emulated programmer-console semantics. The console operates through machine
interfaces rather than arbitrary component internals.

Gate: console behavior is independently tested from DEC documentation.

### Cut 2 — Semantic UNIBUS

**PLANNED**

Establish the complete programmer-visible UNIBUS contract before memory or real
peripherals are integrated.

Contract:

```text
18-bit physical addressing
DATI
DATIP
DATO
DATOB
responder dispatch
bus timeout / nonexistent-address behavior
INIT
BR4 / BR5 / BR6 / BR7
priority arbitration
interrupt vector transfer
interrupt acknowledge
NPR / DMA abstraction
```

Synthetic bus participants prove data transactions, bus errors, priority,
vectoring, acknowledge behavior, INIT, and NPR/DMA semantics before a real
peripheral is attached.

Gate: complete synthetic UNIBUS contract passes manual-derived tests; suitable
behavior is corroborated against SIMH after the DEC contract is established.

### Cut 3 — Core memory

**PLANNED**

Attach architecturally correct PDP-11/40 core memory as a UNIBUS slave.

Initial configuration:

```text
24K words = 060000 words
physical bytes 000000..137777
first nonexistent byte 140000
```

Required semantics include word/byte read and write, DATIP behavior where
architecturally relevant, odd-word fault behavior, nonexistent-memory behavior,
and console EXAM/DEP through normal machine paths.

Model programmer-visible core-memory semantics, not ferrite timing or
electronics.

Gate: core is proven through UNIBUS transactions and console operation.

### Cut 4 — Complete KD11-A processor

**PLANNED**

This is a complete architectural CPU cut, not an opcode-only executor.

First produce a DEC-derived instruction inventory. Each architecturally distinct
instruction family records:

```text
mnemonic
encoding / mask
operand form
legal addressing modes
byte / word behavior
condition-code effects
trap / fault behavior
option status
canonical witnesses
```

Implement the complete configured KD11-A architectural contract, including:

```text
R0-R7, PC, SP, PSW
all addressing modes
complete basic instruction repertoire
11/40-specific instruction forms
configured EIS behavior
condition codes
stack behavior
instruction fetch
odd-address trap
bus-error / timeout trap
reserved-instruction trap
BPT / IOT / EMT / TRAP
WAIT
RTI / RTT
RESET
interrupt priority comparison
interrupt acceptance
vector entry
PC/PS stacking
nested traps and interrupts
```

Synthetic UNIBUS interrupt requests prove CPU interrupt entry before any real
interrupting device exists.

SIMH corroboration uses identical initial architectural state and compares
registers, PSW, memory effects, and trap/vector results. DEC remains authoritative
when a discrepancy appears.

Gate: every instruction-inventory row and every trap/interrupt contract is
certified before peripheral work begins.

### Cut 5 — KW11-L line clock

**PLANNED**

First real device. Implement documented register/state/interrupt semantics,
including interrupt enable, INIT/reset behavior, BR6, and vector `000100`.

Acceptance chain:

```text
clock event
  -> KW11-L request
  -> UNIBUS arbitration
  -> KD11-A acceptance
  -> vector 000100
  -> handler entry
  -> acknowledge / documented device transition
```

No clock-specific bypasses.

Gate: DEC-derived tests plus SIMH corroboration.

### Cut 6 — KL11 console interface

**PLANNED**

Implement receiver and transmitter as independent documented functions,
including registers, ready/done transitions, interrupt enable behavior,
receiver/transmitter requests, BR4, vectors `000060` and `000064`, simultaneous
request behavior, and INIT/reset semantics.

Host terminal I/O is an adapter around the architectural device.

Gate: DEC-derived tests plus SIMH corroboration.

### Cut 7 — KT11-D memory management

**PLANNED**

Introduce virtual memory only after the physical machine is stable.

Layering:

```text
KD11-A virtual reference
        -> KT11-D
        -> 18-bit physical address
        -> semantic UNIBUS
```

Implement kernel/user and current/previous mode, PAR/PDR translation, KISA/KISD,
UISA/UISD, access control, page length and expansion direction, status registers,
fault state, previous-space operations, trap interaction, I/O-page mapping, and
MMU-disabled projection as defined by the KT11-D contract.

Gate: KT11-D manual-derived tests plus SIMH corroboration; CPU tests remain green
with memory management disabled.

### Cut 8 — RK11-D / RK05

**PLANNED**

First DMA block device. Implement the complete programmer-visible controller
contract: documented registers, drive selection, functions, address/count/disk
progression, meaningful errors, NPR/DMA, IDE/READY/completion behavior, BR5,
vector `000220`, and INIT/reset semantics.

The disk image is media, not the controller contract.

Experiment 004's failed completion interrupt becomes a regression requirement
only after restatement from the RK11 and UNIBUS manuals: a successful
IDE-enabled operation must produce the documented BR5/vector-220 interrupt path.

Gate: DEC-derived RK contract plus SIMH corroboration.

### Cut 9 — TM11 / TU10

**PLANNED**

Implement the programmer-visible tape controller/transport contract, including
registers, record operations, tape marks/EOF, BOT/EOT, spacing, rewind,
read/write semantics in scope, NPR/DMA, counter/address progression, documented
interrupt conditions, BR5, vector `000224`, and INIT/reset behavior.

Authority is the TM11 system/controller documentation plus TU10 transport
documentation where transport behavior matters.

Gate: DEC-derived tests plus SIMH corroboration.

### Cut 10 — Remaining V6-relevant peripherals

**PLANNED**

Add only devices required by the target machine configuration. Candidates include
DL11, LP11, and PC11/PR11. Each device follows the same sequence:

```text
DEC manual -> contract -> implementation -> tests -> SIMH corroboration
```

No minimal stubs merely to advance the guest unless the proposal explicitly
excludes the device from the emulated configuration.

### Cut 11 — UNIX V6 acceptance

**PLANNED**

Only after the component contracts are gated do we boot UNIX V6.

Acceptance proceeds serially:

```text
bootstrap
  -> kernel load
  -> kernel initialization
  -> root filesystem
  -> console
  -> init
  -> shell
```

Final acceptance goals are boot from tape, RK05 recognition, construction or
transfer of the disk system, boot from RK05, and an interactive V6 system.

A V6 failure first triggers reconciliation against already-certified component
contracts. It does not justify an integration-only architectural shortcut.

## Project rules

> **Manual first. Component complete. Integration last.**

> **DEC defines. SIMH corroborates. UNIX validates.**

> **Experiment 004 is evidence, never a donor implementation.**

> **No component receives a special bypass merely to advance the guest.**

> **The semantic UNIBUS preserves software-visible behavior and omits invisible
> wire/timing realization.**

> **Architectural completeness comes from the hardware contract, not from what
> UNIX happens to execute.**

## Cut 0 handoff

The assistant prepares only durable repository records for Cut 0. The user owns
Git archival operations and creation of the new baseline.

After the user archives Experiment 004 and establishes the new baseline, the
next engineering action is to create the empty Experiment 005 implementation
surface and begin Cut 1 from the PDP-11/40 programmer-console contract without
copying Experiment 004 PDP source.
