# Milestone 04 — Phase II / Cuts 4–6: runtime and object reproduction

Status: ACTIVE — source implementation and source reconciliation complete; authoritative Aiki runtime gate required.

## Intent

Proceed through all three Phase-II cuts without artificial stopping points:
reconstruct Thompson's runtime, load the worked object program, and reproduce
machine-level matching behavior.

## Runtime implementation

`runtime.ai` installs Thompson's published routines directly as 7094 words:

- `CNODE`
- `NNODE`
- `FAIL`
- `XCHG`
- `INIT`
- `TSXCMD` and `TRACMD` constants

The runtime uses the Phase-I emulator unchanged. Self-modifying list counts live
in the address fields of the leading `AXC` instructions exactly as described by
Thompson. `CLIST` and `NLIST` are executable word arrays in emulated core.

`GETCHA` and `FOUND` are explicit host hooks. `GETCHA` follows Thompson's stated
contract: supply the next character right-adjusted in AC and terminate at end of
input. The experiment passes one nonmatching zero sentinel before termination to
satisfy Thompson's stated one-character FOUND delay. Thompson leaves FOUND
application-dependent; the experiment records the signal and resumes with the
next CLIST entry, matching FAIL's continuation convention so remaining search
paths can run.

## Primary-source discrepancy discovered

The printed final object listing gives location 4 as:

```text
TRA CODE+13
```

Thompson's own object-code producer contradicts that word. At the closure step it
stores the previous operand-entry word at `pc+1` and overwrites the operand entry
with `TRA CODE+pc`. In the worked example the operand entry is location 4 and
`pc = 16`, requiring:

```text
TRA CODE+16
```

Figure 5 likewise routes the path after `a` through the closure CNODE before the
`b|c` alternation. With the printed `CODE+13`, the null branch of `*` is
unreachable at that point and `ad` cannot match.

The implementation therefore preserves both forms:

- `install_verbatim_code` — exact printed `CODE+13` word;
- `install_corrected_code` — compiler-derived `CODE+16` word, all other 22
  locations identical.

This is evidence, not a silent erratum. Phase III should independently regenerate
`CODE+16` from the reconstructed compiler.

## Behavioral corpus

`phase2_test.ai` checks:

- runtime constants and self-modifying count fields;
- all 23 published object operations;
- CLIST initialization by XCHG;
- the verbatim/corrected location-4 distinction;
- zero-repetition match `ad` under the compiler-derived form;
- `b` and `c` closure paths;
- embedded substring matching;
- multiple and adjacent matches; and
- representative nonmatches.

## Validation available in this environment

The new Aiki sources lex and parse successfully against the baseline grammar via
a disposable parser harness using `engine/syntax`.

A full Aiki execution cannot run in this environment because the available Go
1.23.2 toolchain cannot build the repository executable without uncached
`ebiten` and `readline` dependencies, and network access is unavailable. This is
an environmental limitation, not a source failure.

## Critical gate

Run the authoritative repository experiment:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Phase II is GATED only if the Phase-I regression corpus and Phase-II Thompson
corpus both pass. A failure in Phase II must be resolved before reconstructing
the compiler because Phase III depends on this machine/runtime behavior.

## Follow-up — authoritative runtime gate

GATED under the repository executable on 2026-08-15:

```text
Aiki version: aiki v0.4.0-alpha-26
result: experiments/002-thompson-7094-regex/results/run-2026-08-15-233403.309.txt
```

The combined Phase-I and Phase-II corpora completed without reported failure.
Phase III proceeded from this gate.
