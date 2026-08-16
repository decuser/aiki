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

## Post-gate polish

Milestone 09 is an Aiki-style refactor of the compiler plus removal of the
now-unnecessary published/generated word-4 comparison from the visible demo.
It changes Aiki source only and therefore **does not require rebuilding the
Aiki executable**.

See [`09-compiler-aiki-style-pass.md`](09-compiler-aiki-style-pass.md).

## Active validation cleanup

Post-merge validation exposed a runner regression-test context defect and Aiki
module-field naming warnings in the Thompson experiment. Milestone 10 corrects
both without changing machine or compiler semantics.

See [`10-runner-args-and-lint-cleanup.md`](10-runner-args-and-lint-cleanup.md).

## Next action

Run `go test ./...` and `./aiki lint ./...`. Because this cut includes Go test
code, rebuild with `make` before rerunning the experiment `./run.sh`. If all
checks are clean, mark Milestone 10 GATED and the session COMPLETE.

## Follow-up — spawned relative imports

Milestone 11 is ACTIVE. Repository-wide `./aiki test ./...` showed that relative imports inside spawned functions lost their defining-file anchor and fell back to process CWD. Engine correction and regression test are prepared. Next action: run the Milestone 11 local gate, rebuilding Aiki before Aiki-level tests.

## Follow-up — restored machine reset primitive

Milestone 12 is **ACTIVE**. After Milestone 11 allowed repository-root tests to reach the spawned monitor service, three Phase-IV tests exposed that `service.ai` still called `machine.reset_state` while the machine module no longer defined/exported it. The narrow reset primitive has been restored. This cut is Aiki-source only and does **not** require `make`. Next action: run `./aiki test ./...` and the experiment `./run.sh`; gate Milestone 12 if both are clean.
