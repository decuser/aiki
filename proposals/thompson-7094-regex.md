# Thompson 7094 Regex Reconstruction Proposal

## Purpose

Reconstruct Ken Thompson's 1968 regular-expression search implementation by
executing its generated IBM 7094 object code and published runtime on a minimal
IBM 7094 emulator written in Aiki, then reconstruct the compiler that produces
that code.

The project is deliberately narrower than general 7094 emulation. Machine
facilities are admitted only when Thompson's published generated code or runtime
requires them.

## Historical target and authorities

Thompson's compiler accepts a regular expression and produces an IBM 7094
program. The object program and its runtime search input text and signal each
match. The worked expression `a(b|c)*d` is the first historical reproduction
target.

The reconstruction now works from two primary authorities retained with the
project documentation:

- Ken Thompson, *Regular Expression Search Algorithm* (1968), including the
  published object program, runtime routines, and compiler listing; and
- IBM, *IBM 7094 Principles of Operation*, which is authoritative for machine
  word layout, instruction encodings, index behavior, and instruction
  semantics.

Thompson is authoritative for what program is being reconstructed; IBM is
authoritative for what the 7094 instructions mean. Where the two sources leave
a detail unstated, the reconstruction records the inference rather than silently
inventing machine behavior.

The compiler is reconstructed only after the published object-level example is
running. This separates machine correctness from compiler correctness.

### Source transcription correction

The Phase-II reconstruction initially worked from a transcript that rendered
location 4 of Thompson's example object program as `TRA CODE+13`. Rechecking the
paper image during Phase IV established that the published instruction is
`TRA CODE+16`. That value agrees with the closure case in Thompson's Stage-3
compiler and with the independently reconstructed compiler output. The earlier
`CODE+13` claim is retained only in provenance as a superseded transcription
error; it is not part of the emulator's authoritative program image.

### State

```text
memory   32,768 36-bit words
PC       15-bit instruction counter
AC       logical accumulator slice P,1-35 required by Thompson (36 bits)
XR1      15-bit index register
XR2      15-bit index register
XR4      15-bit index register
XR6      15-bit index register
XR7      15-bit index register
```

Unused:

```text
XR3
XR5
MQ
SI
IBM 7094 I/O channels
device state
trap state
timing/overlap state
```

### Word and address sizes

```text
WORD_BITS = 36
ADDR_BITS = 15
WORD_MASK = 2^36 - 1
ADDR_MASK = 2^15 - 1
```

Words and addresses are represented as non-negative Aiki exact integers whose
visible bit patterns are reduced to the appropriate width. Address and index
arithmetic wraps to 15 bits.

IBM numbers word positions from the left as `S,1,...,35`; numeric conversion in
the emulator treats position 35 as the least-significant bit. Therefore the
fields of an ordinary instruction occupy:

```text
IBM positions     numeric bits
S,1-11            35..24   operation
12-13             23..22   indirect flag (unused here)
14-17             21..18   unused for the required ordinary instructions
18-20             17..15   tag
21-35             14..0    address
```

Type-A transfer instructions (`TXI`, `TXH`, `TXL`) instead use IBM positions
`S,1,2` for the operation and positions `3-17` for the 15-bit decrement.

### Index registers

Normal seven-index-register mode is sufficient:

```text
tag 0   no index register
tag 1   XR1
tag 2   XR2
tag 4   XR4
tag 6   XR6
tag 7   XR7
```

Tags 3 and 5 are unused. Multiple-tag compatibility mode is not implemented.

For ordinary tagged instructions:

```text
EA = (Y - XR[tag]) mod 2^15
```

For tag 0:

```text
EA = Y
```

Index-register instructions use the tag to identify the operated register
rather than modify the instruction address. Indirect addressing is not
implemented.

### Thompson runtime register roles

```text
XR1   complement of current input character
XR2   control/list return index
XR4   subroutine return register
XR6   CLIST construction/copy index
XR7   NLIST construction/copy index
```

## Required instruction inventory

Exactly these instructions are initially in scope:

```text
TRA
TSX
TXL
TXH
TXI

AXC
LAC
SCA
PCA
PAC

CLA
CAL
ACL
SLW
```

Total: **14 instructions**.

The IBM operation encodings used by the emulator are the published encodings,
not emulator-private identifiers:

```text
TRA   +0020      TSX   +0074
TXI   +1000      TXH   +3000      TXL   -3000
AXC   -0774      LAC   +0535      SCA   +0636
PCA   +0756      PAC   +0737
CLA   +0500      CAL   -0500      ACL   +0361      SLW   +0602
```

For IBM negative operation codes, the sign position is set in the actual
36-bit instruction word. For example `CAL -0500` has `4500` in the 12-bit
operation field and `AXC -0774` has `4774`.

No instruction is added speculatively. If execution of Thompson's published
material demonstrates that another facility is required, the scope is expanded
from that evidence and the decision is recorded.

## Required instruction semantics

### TRA — Transfer

```text
TRA Y[,T]
PC = EA
```

### TSX — Transfer and Set Index

For a TSX stored at address `x`:

```text
TSX Y,T
XR[T] = (-x) mod 2^15
PC    = Y
```

### TXL — Transfer on Index Low or Equal

```text
TXL Y,T,D
if XR[T] <= D
    PC = Y
else
    PC = PC + 1
```

### TXH — Transfer on Index High

```text
TXH Y,T,D
if XR[T] > D
    PC = Y
else
    PC = PC + 1
```

### TXI — Transfer with Index Incremented

```text
TXI Y,T,D
XR[T] = (XR[T] + D) mod 2^15
PC    = Y
```

### AXC — Address to Index Complemented

```text
AXC Y,T
XR[T] = (-Y) mod 2^15
```

### LAC — Load Complement of Address in Index

```text
LAC Y,T
XR[T] = (-address_field(memory[Y])) mod 2^15
```

### SCA — Store Complement of Index in Address

```text
SCA Y,T
address_field(memory[Y]) = (-XR[T]) mod 2^15
```

All other word bits remain unchanged. Tag 0 stores zero in the address field.

### PCA — Place Complement of Index in Accumulator

```text
PCA ,T
clear relevant AC word positions
AC[address field] = (-XR[T]) mod 2^15
```

### PAC — Place Complement of Address in Index

```text
PAC ,T
XR[T] = (-AC[address field]) mod 2^15
```

### CLA — Clear and Add

The physical 7094 accumulator has positions outside the logical slice required
by Thompson. The emulator deliberately stores only `P,1-35`. For CLA, IBM
clears P and places memory positions `1-35` into AC positions `1-35`; the
memory sign position would go to physical AC sign S, which is outside this
minimal state. Thus the required projection is:

```text
AC[P]    = 0
AC[1-35] = memory[EA][1-35]
```

### CAL — Clear and Add Logical Word

CAL is a logical transfer: memory `S,1-35` maps directly onto accumulator
`P,1-35`:

```text
AC[P,1-35] = memory[EA][S,1-35]
```

### ACL — Add and Carry Logical Word

```text
AC[P,1-35] = end_around_add_36(AC[P,1-35], memory[EA][S,1-35])
```

A carry out of accumulator P is added back into accumulator position 35, as
specified by IBM. Positions physical S and Q are outside the required state.

### SLW — Store Logical Word

```text
memory[EA] = AC logical 36-bit word
```

AC is unchanged.

## Runtime memory structures

### CODE

Contains generated regex machine instructions. The published example for
`a(b|c)*d` occupies offsets `0..22`.

### CLIST

Contains executable `TSX address,2` instructions terminated by `TRA XCHG`.
Each TSX represents a currently viable continuation.

### NLIST

Contains executable `TSX address,2` instructions identifying CODE locations to
be considered for the next input character.

### CNODE

A call at location `x`:

```text
TSX CNODE,4
```

inserts:

```text
TSX x+1,2
```

into CLIST and resumes immediately at `x+2`.

### NNODE

A call at location `x`:

```text
TSX NNODE,4
```

inserts:

```text
TSX x+1,2
```

into NLIST, then returns to the next CLIST entry.

### FAIL

```text
FAIL  TRA 1,2
```

Transfers to the next entry in CLIST.

### XCHG

Responsibilities:

```text
copy NLIST to CLIST
append TRA XCHG
reset runtime list counters
obtain next character
place complement of character in XR1
start a fresh search at CODE
execute CLIST
```

Published sequence:

```text
XCHG   LAC NNODE,7
       AXC 0,6
X1     TXL X2,7,0
       TXI *+1,7,1
       CAL NLIST,7
       SLW CLIST,6
       TXI X1,6,-1
X2     CLA TRACMD
       SLW CLIST,6
       SCA CNODE,6
       SCA NNODE,0
       TSX GETCHA,4
       PAC ,1
       TSX CODE,2
       TRA CLIST
```

### INIT

```text
INIT   SCA NNODE,0
       TRA XCHG
```

Initializes the NLIST count and enters the normal XCHG cycle.

## Host services

### GETCHA

`GETCHA` is a host boundary rather than emulated 7094 device I/O. It must:

```text
obtain next input character
place it right-adjusted in AC
recognize end of input
terminate search at EOF
```

### FOUND

`FOUND` is an application-specific host boundary invoked whenever the complete
regular expression has matched. The implementation may record or report the
match. Thompson's one-character delay means an additional EOF character must be
processed to obtain complete results.

### Host channels versus IBM channels

IBM 7094 I/O channels are not emulated. Aiki channels may be used later as
host-side orchestration between the machine process, monitor, input provider,
and reporter. Those Aiki channels are not part of emulated machine state and
must not be confused with IBM 7094 I/O channels.

## Aiki substrate

The previously identified substrate requirements are now resolved:

- `store` provides explicit mutable indexed storage suitable for 32K-word core
  memory and register state;
- `bits` provides fixed-width masks, logical shifts, extraction, replacement,
  and logical bit operations over exact integral numbers;
- `select` is available for later host-side machine orchestration;
- isolated `spawn` and channels provide a natural machine/monitor boundary;
- Aiki exact rationals avoid hidden floating-point representation in the
  emulator; and
- the experiment framework provides a reproducible home for the reconstruction
  procedure, observations, and analysis.

The emulator itself remains ordinary Aiki code. No new language primitive is a
prerequisite for Phase I.

## Ownership rule

The machine is the semantic owner of its memory, PC, AC, and index registers.
Aiki `store` values can intentionally be shared across spawned computations,
but that capability must not become an accidental bypass around the machine
boundary.

When the concurrent monitor is introduced, examine, deposit, load, run, step,
and stop operations go through the machine command/reply protocol. The monitor
does not retain and mutate the machine's stores directly.

## Explicitly out of scope

```text
general IBM 7094 emulation
XR3
XR5
multiple-tag mode
indirect addressing
floating point
multiply/divide
MQ register
SI register
IBM 7094 I/O channels
peripheral devices
trap modes
interrupts
instruction timing
instruction overlap
cycle accuracy
physical IBM console behavior
loader support
full assembler support
```

## Execution rule

For each supported instruction:

```text
fetch memory[PC]
decode supported opcode and fields
advance or replace PC according to instruction semantics
apply only the required memory/register effects
reject unsupported instructions
```

Unsupported opcode behavior is a machine error. The exact Aiki-level fault
representation is settled when the execution engine is introduced in Phase I,
Cut 3.

# Work plan

The baseline reconstruction proceeds in three phases of three evidence-gated cuts each. A fourth phase adds the operator environment originally described in the project charter without expanding the emulated 7094 architecture.

## Phase I — Build the machine

### Cut 1 — Representation and state

Implement and validate:

- 36-bit word normalization;
- 15-bit address/index normalization;
- 32,768-word mutable memory;
- PC, AC, XR1, XR2, XR4, XR6, and XR7 state;
- memory and register accessors that preserve their widths.

**Gate:** machine representation is internally coherent, zero initialized, and
width preserving independently of any instruction semantics.

### Cut 2 — Instruction semantics

Implement instruction encoding/decoding, effective addressing, and all 14
instruction semantics with focused state-transition tests.

**Gate:** every required instruction is independently demonstrated against its
specified state transition.

### Cut 3 — Execution and control

Implement fetch/decode/execute, PC progression, unsupported-instruction/fault
behavior, observational tracing, and the channel-based machine monitor
boundary.

**Gate:** a controllable minimal 7094 executes small hand-authored programs and
retains machine ownership of state.

## Phase II — Reproduce Thompson's machine-level result

### Cut 4 — Runtime reconstruction

Implement `CNODE`, `NNODE`, `FAIL`, `XCHG`, `INIT`, and the `GETCHA`/`FOUND`
host boundary.

### Cut 5 — Published object program

Encode and load Thompson's published `a(b|c)*d` object program.

### Cut 6 — Historical reproduction

Run representative inputs, retain execution evidence, compare observations with
Thompson's description, and document reconstruction ambiguities.

**Phase II gate:** Thompson's published runtime executes on the Aiki 7094; the
verbatim printed object listing is characterized as printed, and the
compiler-derived one-field correction reproduces the `a(b|c)*d` behavior shown
and described in the paper.

## Phase III — Reconstruct the compiler

### Cut 7 — Compiler structure

Reconstruct the syntax sieve, reverse-Polish conversion, and object-code
producer structure described by Thompson.

### Cut 8 — Code generation

Generate 7094 code from regex source, beginning with the published example.

### Cut 9 — End-to-end experiment

Execute:

```text
regular expression
    -> Thompson compiler
    -> generated 7094 code
    -> Aiki 7094 emulator
    -> input matching result
```

Retain the procedure, raw observations, and interpretation separately.

**Phase III gate:** the complete Thompson compile-search path is reproduced.


## Phase IV — Monitor and observability

Phase IV completes the human-facing machine environment envisioned by the
project charter. It does **not** emulate the physical IBM 7094 console. It is an
Aiki operator monitor, analogous in role to a small SIMH monitor, layered over
the already-gated machine boundary. Architectural state remains owned by the
machine service; the monitor acts only through commands and replies.

### Cut 10 — Operator monitor

Add an interactive console with IBM-style octal presentation and the basic
operator vocabulary:

```text
help
reset
step
run
examine
deposit
show registers
input
quit
```

The monitor must service Thompson's `GETCHA` and `FOUND` host boundaries while
stepping so operator inspection does not fall out of the reconstructed runtime.

**Gate:** a user can load the prepared Thompson machine, inspect state, step the
machine across both ordinary instructions and host boundaries, modify memory,
and run a bounded search without direct access to machine stores.

### Cut 11 — Disassembly and trace

Add 12-digit octal word display, 5-digit octal addresses, disassembly of all 14
supported instructions, bounded instruction/host tracing, and register displays.
Trace remains observational and does not alter machine semantics.

**Gate:** a short Thompson execution can be followed instruction-by-instruction
with raw word, decoded instruction, control transfer, and host events visible.

### Cut 12 — Thompson views and command files

Add named views for `CODE`, `CLIST`, and `NLIST`, retained `FOUND` offsets, and
`do` command files so an operator walkthrough can be replayed as experiment
evidence.

**Gate:** a scripted monitor session can expose the generated program and the
evolution of Thompson's runtime lists while reproducing a search.

### Cut 13 — Interruptible execution and logging

Complete the original monitor architecture with asynchronous operator stop at an
instruction boundary, explicit running/halted status, monitor-controlled logging,
an operator-facing load command, and command-mediated register modification.
The monitor implementation should use Aiki-native forms deliberately: `match`
for dispatch/classification and pipelines for compositional transformations,
while retaining explicit control flow where control itself is the point. This is
simulator control, not an IBM trap or interrupt facility.

**Phase IV gate:** the targeted 7094 can be operated interactively and through
command files as a small historical laboratory, including stopping a running
machine, inspecting/modifying state, tracing Thompson's runtime, and replaying a
retained session.
