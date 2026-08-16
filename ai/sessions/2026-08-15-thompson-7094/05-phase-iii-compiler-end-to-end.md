# Milestone 05 — Phase III compiler and end-to-end reconstruction

Status: ACTIVE

## Intent

Reconstruct Thompson's three-stage compiler, generate the worked 7094 object
program from `a(b|c)*d`, and execute the generated words through the already
gated emulator/runtime stack.

## Phase-II gate received

The user executed the combined Phase-I/II corpus under the repository Aiki
binary:

```text
Experiment 002 — Thompson 7094 Regex Reconstruction
Phases I–II — 7094 machine and Thompson runtime/object reproduction
Aiki version: aiki v0.4.0-alpha-26
result: experiments/002-thompson-7094-regex/results/run-2026-08-15-233403.309.txt
```

No failure or warning was reported. Phase II is therefore GATED.

## Implementation

Added `experiment/compiler.ai` with three explicit stages:

1. `sieve` — validates the core syntax exercised by the paper and inserts `.`
   for juxtaposition;
2. `reverse_polish` — conventional infix-to-RPN reconstruction;
3. `produce` — semantic reconstruction of Thompson's published ALGOL-60
   object-code producer using actual IBM 7094 words.

The first two stages are marked as reconstruction choices because Thompson
states only their contracts. Stage 3 follows the published algorithm.

Added `phase3_test.ai` covering:

- explicit juxtaposition insertion;
- Thompson's worked RPN form `abc|*.d.`;
- malformed-input rejection for the reconstructed Stage 1;
- exact 23-word equality between generated output and the compiler-derived
  Phase-II object program;
- independent generation of `TRA CODE+16` at object location 4;
- loading generated output into a fresh emulated 7094 and executing it through
  Thompson's runtime and host hooks.

`search.execute` is exported so the end-to-end test can run an independently
generated object image rather than routing through the Phase-II convenience
loader.

## Discovery

The Phase-II location-4 discrepancy is now independently resolved by compiler
execution. When closure is compiled, the operand entry is object location 4 and
`pc=16`; Thompson's published assignment
`code[stack[lc-1]] := instruction('tra', value('code')+pc, 0, 0)` therefore
produces `TRA CODE+16`. This result is not inherited from the corrected loader.

The initial compiler deliberately corresponds to Thompson's first published
compiler. His later note about closure over lambda (`a**`) is not folded into
this baseline; doing so would erase the historical progression described in
the paper.

## Validation performed here

- `compiler.ai`, `phase3_test.ai`, and modified `search.ai` lex and parse with
  the current Aiki grammar using the repository syntax package.
- `engine/syntax` parse harness passed.
- `run.sh` remains a strict shell runner and now includes the Phase-III corpus.
- Full Aiki execution is environment-limited because the container cannot
  fetch uncached `ebiten`/`readline` Go dependencies.

## Gate

Run the complete experiment with the authoritative repository executable:

```sh
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Phase III becomes GATED if all three corpora complete without failures and the
runner retains a result artifact.
