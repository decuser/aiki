# Milestone 03 — Phase I / Cuts 2–3: instruction semantics, execution, and control

Status: GATED.

## Intent

Complete the remainder of Phase I without stopping between Cuts 2 and 3: establish
IBM-grounded instruction encoding/semantics, then fetch/decode/execute and the
machine-control boundary.

## Primary-source correction

During Cut 2 the project gained the IBM 7094 *Principles of Operation* alongside
Thompson's 1968 paper/transcription. This invalidated an interim implementation
choice that used emulator-private canonical opcode identifiers because raw IBM
operation encodings had not yet been verified.

The interim encoding was removed rather than layered around.

The IBM manual is now authoritative for:

- the 36-bit instruction geometry;
- 15-bit address, tag, and decrement placement;
- actual operation encodings;
- effective-address rules; and
- execution semantics of the fourteen required instructions.

Thompson remains authoritative for the generated object program and runtime being
reconstructed.

## Important discovery

IBM numbers a word from the left as `S,1,...,35`. When represented as a
non-negative 36-bit integer, IBM position 35 is the least-significant bit.
Therefore the address field is the low 15 bits, tag is the next three bits, and
ordinary operation coding occupies the high 12 bits. The earlier in-progress Cut
2 source had this numeric orientation reversed; it was corrected before Cut 2
was gated or delivered.

Type-A transfers (`TXI`, `TXH`, `TXL`) use positions `S,1,2` for operation and
positions `3-17` for decrement, exactly as documented by IBM.

## IBM operation encodings now used

```text
TRA   +0020      TSX   +0074
TXI   +1000      TXH   +3000      TXL   -3000
AXC   -0774      LAC   +0535      SCA   +0636
PCA   +0756      PAC   +0737
CLA   +0500      CAL   -0500      ACL   +0361      SLW   +0602
```

Negative IBM operation codes are represented by setting the word sign position;
for example CAL has `4500` and AXC `4774` in the actual 12-bit operation field.

## Accumulator projection

The complete physical 7094 accumulator includes positions not required by
Thompson. The emulator explicitly represents only logical `P,1-35` as a 36-bit
value.

This is sufficient for Thompson's use of PCA/PAC/CAL/ACL/CLA/SLW, but CLA and
CAL are deliberately distinct:

- CLA clears P and loads memory positions 1-35 into AC 1-35;
- CAL maps memory S,1-35 directly to AC P,1-35;
- ACL performs the documented 36-bit logical addition with carry from P wrapped
  into position 35.

The proposal now states this projection rather than calling AC generically a
36-bit physical accumulator.

## Cut 2 implementation

`instruction.ai` now provides actual IBM encoding/decoding and field operations.
`machine.ai` implements all fourteen instruction transitions and effective
addressing. `machine_test.ai` contains exact known-word assertions expressed as
12-digit IBM octal words in comments and decimal Aiki values in executable
checks.

## Cut 3 implementation

The same machine module now provides:

- fetch/decode/execute stepping;
- sequential IC progression with transfers replacing it;
- host instruction limits and stop addresses;
- observational execution traces; and
- deterministic rejection of unsupported words.

`service.ai` provides the machine command/reply boundary using Aiki spawn and
channels. The spawned worker imports the emulator inside its isolated
environment; state is passed explicitly. Monitor operations include load,
examine, deposit, step, run, trace, PC access, and stop.

## Validation evidence available here

The four Phase-I Aiki sources lex and parse successfully against the current Aiki
grammar using the existing disposable syntax-check tree:

```text
go test ./engine/syntax -run TestThompsonFilesParseTmp -v
PASS
```

Independent conversion of the IBM octal layouts confirms the decimal word
constants used in the executable encoding tests.

## Runtime gate — PASSED

The authoritative repository executed the complete Phase-I corpus successfully:

```text
Experiment 002 — Thompson 7094 Regex Reconstruction
Phase I — Machine: representation, instructions, execution, control
Aiki executable: /home/wsenn/forge/dev/aiki/aiki
Aiki version:    aiki v0.4.0-alpha-26
result: /home/wsenn/forge/dev/aiki/experiments/002-thompson-7094-regex/results/run-2026-08-15-232508.454.txt
```

No warnings or failures were reported. This supplies the runtime evidence that
was intentionally withheld after the primary-source correction. Cuts 2 and 3,
and therefore Phase I as a whole, are GATED.
