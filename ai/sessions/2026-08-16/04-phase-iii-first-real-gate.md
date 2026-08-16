# Phase III first real gate

Status: ACTIVE

## Trigger

The corrected runner produced:

```text
Phase I: 96 tests, 96 passed, 0 failed
Phase II: 59 tests, 59 passed, 0 failed
Phase III: 10 tests, 4 passed, 6 failed
```

Retained result:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-16-001301.759.txt
```

This establishes Phase I and Phase II as genuinely gated under `aiki test` and narrows the remaining defects to the compiler reconstruction.

## Failure analysis

The six Phase-III failures collapse to two Stage-1/Stage-2 reconstruction bugs.

1. `sieve` treated the empty sentinel `previous = ""` as an atom because `is_atom` was defined as simply `not is_operator`. That caused an explicit juxtaposition dot to be emitted before the first source character. `can_end_operand` now rejects the empty sentinel structurally before asking whether the character is an atom.

2. `without_last` used `while i < length(xs) - 1`. Under Aiki's left-to-right expression semantics this is evaluated as `(i < length(xs)) - 1`, producing the reported `cannot subtract boolean and number`. The arithmetic is now explicitly grouped as `i < (length(xs) - 1)`.

The Stage-2 fault cascaded into Stage 3 and the end-to-end compiler test, so no 7094 machine or Thompson-runtime change is justified by this failure set.

## Changes

- `compiler.ai`: prevent the empty Stage-1 sentinel from ending an operand.
- `compiler.ai`: group the arithmetic bound in `without_last` so comparison occurs after subtraction.

## Validation required

Run `./run.sh` again. Phase I must remain 96/96 and Phase II 59/59. Phase III should either pass or expose the next genuine compiler defect. Do not alter the already-gated machine/runtime layers unless new evidence requires it.
