# PDP-11/40 V6 Lions Laboratory Emulator Proposal

## Status

ACTIVE — project branch `v6-emulator`; implementation begins with Gate 1 core diagnostics.

## Purpose

Build, in Aiki, the smallest faithful PDP-11/40 system needed to construct and
run the V6 UNIX laboratory used to read John Lions' commentary.

This is not a general PDP-11 emulator and not a claim to support all of Sixth
Edition UNIX. Scope is admitted by the Lions laboratory execution path and the
architecturally visible behavior of the PDP-11/40, UNIBUS, and the specific
peripherals that path requires.

The first historical path is deliberately direct:

```text
PDP-11/40 core diagnostics
    -> boot the original V6 distribution tape
    -> run the standalone installation program
    -> use its own tmrk program to construct one RK05 system disk
    -> boot that disk into V6
```

No prebuilt root disk is used to bypass the tape-to-disk installation path.

## Authorities

Primary machine authorities:

- DEC, *PDP-11/40 Processor Handbook* (1972): CPU architecture, instruction
  semantics, addressing modes, PSW, traps, interrupts, and model-visible
  behavior.
- DEC, *PDP-11 UNIBUS Design Description* (1979): architecturally visible bus
  transactions, arbitration, interrupt, initialization, and DMA relationships.
- DEC, *KT11-D Memory Management Option User's Manual* (1976): memory
  management behavior required when V6 proper is entered.

Target-system authorities:

- *Setting Up UNIX—Sixth Edition: An OpenSIMH Laboratory for Reading Lions*:
  exact target configuration and installation/boot path.
- John Lions, *UNIX Operating System Source Code Level Six* and *A Commentary
  on the UNIX Operating System*: the V6 code and explanatory workload the
  machine must support.

Peripheral manuals are consulted only when a gate reaches the corresponding
peripheral. The emulator implements architecturally visible behavior required by
V6 and its standalone programs, not electromechanical detail for its own sake.

## Target machine

The completed Lions laboratory corresponds to a PDP-11/40 with the devices used
by that laboratory. Initial implementation is narrower and grows only as gates
require it.

Operator-facing vocabulary is UNIX/V6 vocabulary:

```text
CPU
memory
UNIBUS
tape
disk
console
clock
printer
paper-tape
```

DEC controller/transport names belong in source comments, documentation, and
deep debug where historically useful; they do not define the normal monitor
surface.

The emulator application and monitor are named `aiki-pdp`; the interactive
prompt is:

```text
aiki-pdp>
```

## Architectural commitments

### 1. Architectural, not microarchitectural, emulation

The emulator implements the externally visible PDP-11/40 contract. It does not
model microcode, internal pipeline behavior, electrical timing, or other
implementation detail unless later evidence shows that an externally visible
V6 behavior depends on it.

### 2. Deterministic state transition

Execution is a deterministic architectural state transition. Given the same
initial machine state, media, console input, and monitor-control events, the
machine produces the same execution sequence.

Host wall-clock time never determines guest execution order.

The CPU cycle is conceptually:

```text
recognize eligible interrupt at architectural boundary
fetch
decode
resolve operands
execute
commit architectural effects
advance deterministic machine time through UNIBUS
recognize synchronous trap/result
repeat
```

Exact PDP-11/40 ordering is taken from the handbook where architecturally
observable.

### 3. UNIBUS is the machine/device coordination boundary

All CPU-device interaction crosses the UNIBUS abstraction. UNIBUS owns the
architectural relationship among:

- memory and I/O address selection;
- device register reads/writes;
- DMA/NPR memory transfers;
- interrupt requests and priority;
- initialization/reset propagation; and
- deterministic device progress.

Devices do not mutate CPU or memory state through private back doors.

### 4. Separation of concerns by semantic directory

Folders represent machine concepts rather than generic helper buckets. Expected
initial structure:

```text
experiment/
  machine/
  cpu/
  memory/
  unibus/
  monitor/
  observe/
  diagnostics/
```

Peripheral directories are added only when their gates begin:

```text
  devices/tape/
  devices/disk/
  devices/console/
  devices/clock/
  devices/printer/
  devices/paper-tape/
```

### 5. State ownership

The machine is the authoritative owner of mutable architectural state. CPU,
memory, UNIBUS, and devices expose explicit operations. Monitor, logger, and
debugger inspect or request operations through those boundaries; they do not
retain private authoritative copies or mutate arbitrary internals.

### 6. Built-in PDP-11 reference help

The emulator provides an operator-facing PDP-11/40 reference derived from the
processor handbook. The same reference authority is used by the standalone
help entry point and, once present, the `aiki-pdp>` monitor. Initial topics
include addressing modes, registers, PSW, instruction groups and octal base
codes, branches, traps, and selected mnemonic detail. Help is reference material,
not a second source of machine semantics; executable decode/behavior remains
validated independently against the same handbook authority.

### 7. Observation is not execution

The runtime has three planes:

```text
control      aiki-pdp monitor
machine      CPU / memory / UNIBUS / devices
observation  structured event log / debug-state projection
```

Logging and debugging may observe machine events and snapshots but must not pace
or alter guest execution.

A later UI may open independent read-only observer windows. UNIBUS has its own
transaction/arbitration/DMA view, and each configured device family has its own
semantic view (tape, disk, paper tape, printer, console, clock). CPU/debug state
remains separate. These windows consume structured events and snapshots; no
observer window owns or mutates machine state. Their exact presentation is
deferred and the observation interfaces must not depend on a particular UI
technology.

### 8. Host terminal control contract

While a monitor command or guest execution is active:

```text
CTRL-T   status snapshot; no machine-state change
CTRL-E   suspend emulation and return to aiki-pdp>
CTRL-C   normal foreground interrupt behavior
CTRL-D   normal EOF behavior
```

`CTRL-T` and `CTRL-E` must remain effective even if the guest is looping or a
long transfer is in progress. The guest may hang; the emulator monitor must not.

### 9. Guest faults vs emulator faults

Architectural conditions such as odd-address traps, reserved instructions,
memory-management violations, and bus errors are guest-machine behavior.
Malformed host configuration, missing required media, impossible internal
state, or violated emulator invariants are emulator failures. These categories
must remain distinct.

### 10. Media policy

Original distribution tape is read-only by default. Disk images are explicitly
writable when constructing the V6 system. Host attachment policy must not allow
accidental mutation of master historical media.

### 11. Replayability

The architecture must preserve the possibility of exact replay from:

```text
machine configuration
initial machine state
attached media
console input
monitor control events
```

Trace recording/replay is not required for Gate 1, but hidden nondeterminism is
not permitted.

## Project gates

Each gate is a serial evidence boundary. Work stops after each gate for user
validation and input before the next gate begins.

### Gate 1 — PDP-11/40 core is demonstrably functional

No peripheral is required.

Required capabilities:

- known reset state;
- word/byte memory representation and access;
- registers R0-R7, with R6/SP and R7/PC architectural behavior;
- PSW condition state needed by exercised instructions;
- deposit/examine through the monitor boundary;
- run and single-step;
- `CTRL-T` status and `CTRL-E` suspension contract represented in the control
  architecture, with host-terminal wiring allowed as a later cut if the current
  Aiki terminal surface requires substrate support;
- deterministic fetch/decode/resolve/execute/commit loop;
- representative instruction behavior sufficient to exercise every PDP-11
  addressing mode;
- before/after diagnostic fixtures for addressing side effects, PC, registers,
  memory, and PSW where relevant;
- bounded control-flow/stack diagnostics sufficient to establish that the core
  is usable before tape is introduced.

Addressing diagnostics cover, in source and destination positions where
meaningful:

```text
register
register deferred
autoincrement
autoincrement deferred
autodecrement
autodecrement deferred
indexed
indexed deferred
PC-derived immediate / absolute / relative / relative-deferred cases
SP byte/word increment/decrement special behavior
```

Gate evidence must include a terse aggregate diagnostic result and exact
expected-vs-actual before/after state on failure.

### Gate 2 — tape bootstrap executes against the bus boundary

Add only the tape behavior required by the six-word V6 bootstrap. Establish
that the already-gated CPU executes the bootstrap correctly up to and through
its bus-visible tape operations.

### Gate 3 — distribution tape boots standalone V6 installer

The six-word bootstrap reads the first distribution-tape record into memory at
address 0. Execution of address 0 reaches the standalone `=` prompt.

This is the first historical whole-artifact gate.

### Gate 4 — standalone V6 constructs one system disk

Add one disk device. Run V6's own standalone `tmrk` program to copy:

```text
tape offset 100 -> disk block 0, count 1
tape offset 101 -> disk block 1, count 3999
```

The resulting disk image must be produced by the emulated machine, not imported
as a completed root disk.

### Gate 5 — KT11-D memory management required by V6 proper

Add the PDP-11/40 memory-management behavior required by `rkunix`: Kernel/User
mode, the active PAR/PDR sets, SR0/SR2 behavior used by V6, relocation,
protection, and architectural abort/trap behavior. This gate is deliberately
after standalone tape-to-disk construction so CPU, tape, and disk faults are not
confounded with address translation.

### Gate 6 — boot the constructed disk into V6

Boot the disk, reach the `@` bootstrap prompt, load `rkunix`, and reach the V6
super-user shell `#`.

This closes the first major project campaign:

```text
original V6 distribution tape
    -> standalone installer
    -> emulated disk transfer
    -> KT11-D memory management
    -> disk bootstrap
    -> running V6
```

### Later gates — complete the Lions laboratory

After Gate 6 and explicit user approval, extend only as required to rebuild and
run the complete Lions configuration, including clock, additional RK05
filesystems, and the configured peripherals. The complete laboratory, not all
possible V6 configurations, remains the acceptance workload.

## Gate 1 diagnostic model

Diagnostics are declarative before/program/after fixtures. Each establishes a
known machine state, executes a bounded instruction sequence, and compares the
complete relevant architectural after-state.

Conceptually:

```text
before
  registers
  psw
  memory

program
  encoded words
  bounded steps

after
  registers
  psw
  memory
  pc
```

A passing aggregate should be mechanically readable, for example:

```text
AIKI PDP-11/40 CORE DIAGNOSTICS

register addressing ........ PASS
register deferred .......... PASS
autoincrement .............. PASS
autoincrement deferred ..... PASS
autodecrement .............. PASS
autodecrement deferred ..... PASS
indexed .................... PASS
indexed deferred ........... PASS
PC-derived modes ........... PASS
SP byte/word behavior ...... PASS
control flow / stack ....... PASS
PSW/NZVC ................... PASS

CORE PASS
```

The exact cases and expected state are derived from the PDP-11/40 handbook, not
from remembered generic PDP-11 behavior.

## Initial serial cuts

### Cut 0 — project state and contract

Status: GATED when branch, proposal, experiment skeleton, and durable restart
state are committed from the clean baseline.

### Cut 1 — machine state, memory, and Gate-1 fixture harness

Status: GATED — user validation: 42 tests passed, 0 failed.

Build the smallest state model and diagnostic runner required to express exact
before/after tests. No instruction implementation beyond what is necessary to
validate the harness itself.

### Cut 2 — fetch/decode and addressing resolution

Status: GATED — addressing repair validated by the user after the lexical shape-boundary correction; PDP help separately passed 16/16.

Implement the instruction word decoding and addressing-mode resolver from the
PDP-11/40 handbook, including PC/SP special cases. Gate with focused addressing
fixtures.

### Cut 3 — m40.s execution contract, PSW, control flow, stack, and audit

Status: GATED — user validation completed the execution and `m40.s` suites; instruction/addressing/PSW audit is active.

Implement every instruction form exercised by Lions' `m40.s` workload, with semantic diagnostics rather than decode-only acceptance. Add runtime audit counters for instruction use and source/destination addressing modes so later V6 runs can be compared against Lions' static description and the source listing. Machine-facing numeric presentation is octal-first.

`MFPI`/`MTPI` retain their instruction/stack-transfer contract here, but previous-space translation and full mode semantics remain explicitly owned by Gate 5 KT11-D memory management. `WAIT`, `RESET`, and `RTT` establish their current architectural seams; interrupts, attached-device INIT propagation, and complete mode/trap behavior are strengthened by their later owning gates.

### Cut 4 — raw V6 tape and Gate 2 bootstrap

Status: GATED — synthetic/controller suite 31/31 and real TUHS raw-media bootstrap reached PC `100012` with the expected first-record signature and `TAPE BOOTSTRAP PASS`.

Use the TUHS raw `v6.tape` representation directly: exactly 12,100 fixed
512-byte records with the known terminal tape mark synthesized by the media
layer. Do not require OpenSIMH `enblock` framing. Add the PDP-11/40 16-bit to
18-bit UNIBUS I/O-page projection, the narrow TM11 register/read/DMA behavior
required by the six-word historical bootstrap, and a bit-exact real-media
bootstrap witness.

Gate: the six-word bootstrap executes from `100000`, transfers raw tape record
0 into memory through UNIBUS DMA, advances the tape by one record, and remains
in the loop at `100012`. Stop for user validation before adding console or
executing the standalone installer.


### Cut 5 — `aiki-pdp` monitor and read-only observer windows

Status: ACTIVE.

Build the normal operator control surface before Gate 3 so the historical path is no longer dependent on a purpose-built bootstrap runner. The monitor owns command interpretation and remains the only writer of machine state. A raw-terminal input reader sends keystrokes to the monitor event loop; CTRL-T reports status without mutation, CTRL-E suspends execution and returns to `aiki-pdp>`, CTRL-C interrupts the foreground run, and CTRL-D at an empty prompt is EOF/quit.

Add an observation broker inspired by Experiment 003 but stronger about execution isolation: one read-only CPU/debug view, one UNIBUS view, and one tape view for currently implemented components. Each observer has its own sender process and a latest-snapshot mailbox so blocked terminal/TCP output cannot pace guest execution. Later disk, paper-tape, printer, console, and clock windows use the same contract. Window creation remains host presentation in `showcase.sh`; closing an observer must not affect the machine.

Gate this cut with focused monitor/view diagnostics and a manual run that starts `showcase.sh`, attaches the real raw tape, deposits the six bootstrap words through `aiki-pdp>`, runs from `100000`, verifies CTRL-T/CTRL-E, examines memory 0, and confirms the three observer windows track CPU, UNIBUS, and tape state. Gate 3 standalone execution does not begin until this control surface is validated.

## Validation discipline

Focused Aiki diagnostics are used during serial cuts. Gate 1 additionally runs
the strongest relevant repository validation available, normally `make
validate` or stronger if the project touches surfaces covered by `make
rigorous`.

A cut is not `GATED` until its stated evidence has actually run. Any environment
limitation is recorded exactly rather than silently weakening the claim.

No expected/gold artifact is changed after a gated baseline except as an
explicit planned change.

## Explicit non-goals

Until the Lions workload proves otherwise, this project does not attempt:

- general PDP-11 family compatibility;
- PDP-11/45 or PDP-11/70 behavior;
- cycle-accurate microarchitecture;
- electrical UNIBUS simulation;
- arbitrary DEC peripheral coverage;
- networking;
- all historical V6 configurations;
- SIMH command compatibility; or
- reproduction of DEC operator-console vocabulary at the normal user surface.


### Current execution gate — RK11/RK05

Cut 7 adds the Lions-lab RK path as one controller with eight attachable RK05 drives. The gate is not controller unit tests alone: the real V6 standalone `tmrk` must use TM11 space-forward/read plus RK11 NPR write operations to create disk 0 from the original distribution tape. The accepted transfer sequence is tape offset 100 -> disk block 0 count 1, then tape offset 101 -> disk block 1 count 3999. Booting that resulting pack is the following gate.

### Cut 7 architectural performance refinement

Real V6 standalone execution and staged semantic profiling established that the
remaining emulator cost was representation and dispatch overhead rather than
PDP-11 work. The execution realization therefore follows the physical machine
more directly while preserving the architectural contracts above:

- CPU architectural state is compact mutable storage; rich Aiki structures are
  produced at observation boundaries rather than on every instruction.
- Physical core is word-backed while remaining byte-addressable. Word fetches
  and stores are one indexed storage operation; byte operations select/update a
  byte lane.
- Backplane configuration establishes I/O address ownership once. Ordinary RAM
  takes the direct UNIBUS memory path; an I/O-page address selects the configured
  controller rather than polling every controller with `handles(address)`.
- UNIBUS routes transactions. Controllers own bus-visible registers, commands,
  interrupts/NPR sequencing. Attached devices own their own media and local
  device state. In the current machine this is explicit as TM11 -> TU10 and
  RK11 -> RK05 drive.
- Idle controllers receive no per-instruction advancement. Controller work is
  serviced only while a command is pending; the authoritative machine loop
  remains the sole committer of machine state.
- Instruction decode is classified once per fetched word and dispatched from
  that result; the executor does not repeat decoder-family searches.
- Observer processes remain concurrent, disposable subscribers. `show TYPE`
  prints a snapshot in the monitor; `observe TYPE` (alias `open TYPE`) launches
  another observer window when one has been closed.

These are execution-realization changes, not shortcuts around the PDP-11/40,
UNIBUS, or controller semantics. The same historical `tmrk` gate and the same
`CLR/CMP/BLO` scaling workload remain the acceptance evidence.

### Cut 7 fixed-width execution-domain refinement

Aiki `number` remains exact rational. PDP machine-width representation is not a
new Aiki numeric type and is not implemented by hidden integer/float promotion.
The emulator explicitly imports a provider-backed machine capability instead:

```text
ordinary Aiki exact number
        |
        | explicit boundary conversion
        v
opaque machine/ffi value
        |
        | fixed-width machine operations + typed store/ffi
        v
opaque machine/ffi value
        |
        | explicit observation/control conversion
        v
ordinary Aiki exact number
```

The following invariants govern this realization:

- No `int`, `float`, `:word`, `:byte`, or `:addr18` semantic is added to Aiki.
  Provider width/kind is private; ordinary source sees only an opaque capability
  value and cannot apply ordinary Aiki arithmetic to it.
- Conversion is deliberate. Exact rationals enter the machine domain at monitor,
  diagnostic, configuration, or media/controller boundaries and return only at
  observation/control boundaries.
- PDP data words, registers, IR, and PSW are 16-bit machine values; bytes are
  8-bit machine values; physical/UNIBUS addresses are 18-bit byte addresses.
  An 18-bit address is never treated as an 18-bit data word.
- `store/ffi` is typed provider-backed storage, not a faster arbitrary-value
  Aiki store. Physical machine storage uses fixed-width `word`, `byte`, and
  address backing. Execution-only non-negative counters use an explicit counter
  block with provider-side `add`, avoiding `store.get -> Aiki arithmetic ->
  store.set` bookkeeping in the CPU hot path. Counter blocks are implementation
  capabilities, not an Aiki integer type. Physical RAM remains a word store
  indexed from the 18-bit byte address; byte operations select/update a word lane.
- Functions model machine relationships; storage models state owned by the
  hardware being represented. Rich Aiki lists/shapes are observation/control
  representations, not the normal CPU/RAM execution representation.
- The current serial cut applies this invariant to CPU/RAM and PSW audit first.
  Legacy numeric controller-register APIs remain an explicit I/O-boundary
  compatibility island until the following representation cut. EIS wider signed
  operations are similarly isolated until given a fixed-width implementation.

The same historical diagnostics and `CLR/CMP/BLO` scaling workload remain the
acceptance evidence. A performance improvement does not excuse a semantic
regression, and a semantic pass does not by itself gate the real `tmrk` transfer.
