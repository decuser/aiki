# Summary — proof and demonstration corrected

Adding a human-readable demonstration exposed a Stage-2 compiler fault that the nominal Phase III corpus had not surfaced. The immediate bug was an operator-stack scan that relied on compound boolean `and` as if it guarded later operands; an empty stack was consequently indexed at -1. The parser now performs explicit structural emptiness checks before top-of-stack access.

The deeper finding was in the experiment driver. It ran `*_test.ai` files as ordinary Aiki programs. Aiki's test assertions accumulate results, and `test.run` catches faults, but only the `aiki test` command reports those accumulated failures and returns a failing exit status. Thus the previous runs were execution evidence, not valid assertion gates.

The runner now separates three concerns cleanly: each phase corpus is executed by `aiki test` and must explicitly pass; only after those gates succeed does `demo.ai` present the historical reconstruction to a human; `tee` retains the complete transcript. Earlier GATED/COMPLETE claims are superseded until one corrected local run passes all three corpora and the demonstration.

## First trustworthy Phase-II failure set

The corrected gate immediately justified itself. Phase I passed 96/96, while Phase II exposed six failures. Examination showed that the failures were dominated by Aiki semantics rather than 7094 semantics: shaped lists carry their shape as metadata rather than element zero, and a mixed ungrouped `not ... and ... < ...` expression was invalid under Aiki's eager, left-to-right binary evaluation. Correcting those mistakes also exposed equivalent off-by-one shaped-result handling in the compiler and demo, which was fixed proactively. The reconstruction remains deliberately evidence-driven: no IBM instruction semantic was changed without a failure that requires it.


## First trustworthy Phase-III failure set

The next corrected run left Phase I at 96/96 and gated Phase II at 59/59, then exposed six Phase-III failures. These again localize to Aiki expression semantics in the reconstructed compiler rather than to the emulated machine: the Stage-1 sentinel empty string was classified as an atom, yielding a leading juxtaposition operator, and the Stage-2 list-pop helper wrote `i < length(xs) - 1`, which Aiki evaluates left-to-right as a boolean followed by subtraction. The fixes are deliberately confined to `compiler.ai`; the already-gated 7094 and Thompson runtime remain untouched.

## Final trustworthy gate

The subsequent local run completed all corrected assertion gates: Phase I 96/96, Phase II 59/59, and Phase III 39/39. The end-to-end demonstration then compiled `a(b|c)*d` through the reconstructed sieve, reverse-Polish stage, and source-derived Stage 3 into 23 IBM 7094 words, executed those words on the Aiki 7094 emulator with Thompson's runtime, and produced the expected FOUND offsets for positive and negative examples. The retained evidence is `run-2026-08-16-001438.448.txt`.

This run supersedes all earlier nominal completion claims. The baseline reconstruction is now COMPLETE.


## Phase IV — operator environment

Review of the original charter after the Phase I–III reconstruction showed that the planned SIMH-like monitor had not been delivered. Phase IV therefore treats the completed machine/compiler reconstruction as a stable substrate and adds a human-facing operator layer. The monitor uses octal presentation, disassembly, machine-service-mediated state access, host-aware stepping, Thompson-specific `CODE`/`CLIST`/`NLIST` views, and replayable command files. Physical IBM console behavior remains out of scope. The final cut will add interruptible execution and logging after the first local monitor gate.
