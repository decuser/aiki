# Phase II first real gate

Status: ACTIVE

## Trigger

The first corrected `aiki test` run produced trustworthy results:

```text
Phase I: 96 tests, 96 passed, 0 failed
Phase II: 45 tests, 39 passed, 6 failed
```

Retained result:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-16-000933.738.txt
```

This is the first run in which the experiment driver actually surfaced assertion and caught-fault failures.

## Failure analysis

The six Phase-II reports collapse to two Aiki-level mistakes rather than six independent 7094 defects.

1. Shaped lists were treated as though the shape marker occupied element zero. In Aiki the shape is metadata: `shape([@stopped, ...])` yields `:stopped`, while indexing begins with the first payload field. Thus the XCHG test read the step count (`10`) where it expected `:stopped`, and the search tests similarly used shifted payload indexes.

2. `search.execute` used `while not done and steps < limit`. Aiki evaluates binary expressions left-to-right and both operands of `and`; the ungrouped expression therefore combined a boolean with the numeric comparison path and produced `cannot compare boolean and number`. The condition now groups the boolean and numeric predicates explicitly.

The same shaped-list indexing mistake was found proactively in Phase III and in the visible demonstration. `compiler.compile` also extracted the shaped `@object` result one position too far to the right. These are corrected before the next gate run rather than waiting for Phase III to expose them serially.

## Source check

No IBM 7094 semantic correction is made in this cut. The IBM manual specifies TXL as transferring when the selected XR is less than or equal to D, and Thompson's runtime uses that instruction exactly in XCHG and the generated character tests. The current failure signatures arise before evidence requires changing those machine semantics.

## Changes

- `search.ai`: group the loop predicates explicitly.
- `phase2_test.ai`: test shaped result identities with `shape(...)` and use payload indexes correctly.
- `compiler.ai`: extract `@object` payload fields correctly.
- `phase3_test.ai`: correct `@compiled` and `@search` payload indexing proactively.
- `demo.ai`: correct compiled/search shaped payload indexing.

## Validation required

Run `./run.sh` again. Phase I must remain 96/96. Phase II must now either pass or expose the next actual emulator/runtime defect. Phase III and the visible demonstration are not considered gated until reached by the corrected runner.
