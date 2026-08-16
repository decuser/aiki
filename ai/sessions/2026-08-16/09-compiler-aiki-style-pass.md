# Milestone 09 — compiler Aiki-style pass

Status: ACTIVE

## Intent

Refactor the reconstructed Thompson compiler toward Aiki-native expression
without changing compiler behavior or the generated IBM 7094 object program.
Remove the now-unnecessary published/generated word-4 comparison from the
human-facing demonstration; retain source agreement as an executable invariant
in the Phase-III corpus.

## Changes

- Replaced regex token/operator classification chains with `match`.
- Centralized token classification in `token_kind`.
- Simplified Stage-1 validation around operand-boundary predicates rather than
  repeated nested token tests.
- Extracted Stage-2 operator-stack draining into `drain_to_open` and
  `drain_precedence`, leaving `reverse_polish` as token dispatch plus state
  updates.
- Split Stage-3 emission into `emit_alpha`, `emit_closure`, and `emit_or`; the
  main producer now dispatches with `match` instead of nested `if/else`.
- Expressed the top-level transformation with pipelines:
  source -> sieve -> reverse Polish -> producer, while retaining explicit
  shaped-error checks between stages.
- Removed the visible published/generated word-4 comparison from `demo.ai`.
  Phase III still compares every generated object word to the published object
  program, so the historical/source fidelity invariant remains covered.

## Evidence so far

The modified `compiler.ai` and `demo.ai` lex and parse successfully against the
current Aiki grammar using a disposable Go-1.23 syntax-only harness. This does
not constitute an execution gate.

No Go/substrate/CLI files changed. **No rebuild is required.**

## Gate required

Run the normal experiment driver with the already rebuilt Aiki executable:

```text
cd experiments/002-thompson-7094-regex/experiment
./run.sh
```

Required evidence:

- Phase I remains green.
- Phase II remains green.
- Phase III remains green and still reproduces all 23 published words.
- Phase IV remains green.
- End-to-end demonstration passes without the word-4 comparison block.
- Monitor walkthrough completes.

On that evidence, mark this milestone GATED and the post-Phase-IV polish
COMPLETE.
