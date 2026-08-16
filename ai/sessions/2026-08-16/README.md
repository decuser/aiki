# Thompson 7094 regex reconstruction — completed baseline and monitor

Status: ACTIVE

Prior session: [`../2026-08-15-thompson-7094/`](../2026-08-15-thompson-7094/README.md)

## Current truth

The authoritative local run after the Phase-IV source/demo polish passed all
four corpora and the operator walkthrough:

```text
Phase I — 96 tests, 96 passed, 0 failed
Phase II — 57 tests, 57 passed, 0 failed
Phase III — 39 tests, 39 passed, 0 failed
Phase IV — 24 tests, 24 passed, 0 failed
```

Retained gate artifact:

```text
experiments/002-thompson-7094-regex/results/run-2026-08-16-004354.650.txt
```

The end-to-end reconstruction and scripted monitor walkthrough both passed.
The corrected project transcript and compiler agree on Thompson's published
`TRA CODE+16` at object word 4; the earlier `CODE+13` reading is retained only
as superseded provenance.

## Gate status

- Phase I: **GATED** (96/96).
- Phase II: **GATED** (57/57).
- Phase III: **GATED** (39/39).
- Phase IV monitor: **GATED** (24/24 plus scripted walkthrough).
- Historical reconstruction and monitor functionality: **COMPLETE**.

## Active post-gate polish

Milestone 09 is an Aiki-style refactor of the compiler plus removal of the
now-unnecessary published/generated word-4 comparison from the visible demo.
It changes Aiki source only and therefore **does not require rebuilding the
Aiki executable**.

See [`09-compiler-aiki-style-pass.md`](09-compiler-aiki-style-pass.md).

## Next action

Merge the style-pass drop and run `./run.sh` with the existing rebuilt binary.
If all four corpora and both demonstrations remain clean, mark Milestone 09
GATED and this session COMPLETE.
