# Aiki — Thompson completion and HAL redesign

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


## HAL redesign — active design phase

Cut 0 and Phase I are now **GATED** against exact baseline `v0.4.0-alpha-27`.

- Cut 0: systems-programmer affordance target, three-name model, three-phase/four-cut structure.
- Phase I Cut I.1: all 117 registered primitives inventoried.
- Phase I Cut I.2: registry responsibilities classified.
- Phase I Cut I.3: systems-programmer affordance coverage/gaps mapped.
- Phase I Cut I.4: module/spawn/profiling/file/Canvas/time-select/Thompson pressure synthesis gated.

Key finding: `time.after` already demonstrates a host-produced receive-only Aiki channel participating normally in `select`; this is evidence for Phase-II Canvas discussion, not a universal transport decision.

Current HAL artifacts:

- `proposals/hal-redesign-cut0.md`
- `proposals/hal-redesign-phase-i-inventory.md`
- `proposals/hal-redesign-phase-i.md`

**Next action:** Phase II Cut II.1 — decompose execution context (`Env` / `EvalContext`) by semantic meaning and trace call/spawn/module/HAL propagation.


## HAL redesign — Phase II gated

Phase II Cuts II.1-II.4 are **GATED** by source inspection/design consistency.

Key decisions/findings:

- context is decomposed into lexical, source, dynamic, observation, authority, and module facets;
- authority is lexical/definition-bound and explicitly separate from environment role;
- current spawn construction appears to grant `ScopePrelude` to user-spawned functions, a baseline authority leak to eliminate in migration;
- runtime/session ownership replaces package-global host relationships conceptually;
- `value.File` and `value.Canvas` currently leak concrete Go substrate types into the semantic value layer;
- trusted libraries receive explicit HAL grants rather than blanket path-derived privilege;
- `time.after` proves host-backed receive-only channels already participate normally in `select`;
- Canvas remains a pressure requirement, not the generic model.

Artifact: `proposals/hal-redesign-phase-ii.md`.

**Next action:** Phase III Cut III.1 — canonical HAL contract/metadata and registry separation.


## HAL redesign — Phase III/design gate

Phase III Cuts III.1-III.4 are **GATED**. Cut 0 plus all three design phases are complete.

Artifacts:

- `proposals/hal-redesign-cut0.md`
- `proposals/hal-redesign-phase-i-inventory.md`
- `proposals/hal-redesign-phase-i.md`
- `proposals/hal-redesign-phase-ii.md`
- `proposals/hal-redesign-phase-iii.md`

Implementation has **not** started. The serial implementation sequence is M1-M7.

**Exact next action if the design is accepted:** M1 — add canonical host-operation metadata and executable Aiki-name / HAL-name / substrate-provenance coverage beside current behavior, with no semantic change.


## HAL redesign — M1 implementation

Milestone 26 is **GATED** on branch `hal-redesign`. The first implementation
slice adds canonical metadata for the 18 settled host crossings (I/O, time,
args/environment, files) beside the existing compatibility registry and adds
executable three-name/wrapper coupling. Canvas is deliberately not canonized
by this slice.

**Gate:** the user merged the M1 drop and reported `make validate` passed on the authoritative tree.

## HAL redesign — M2 implementation

Milestone 27 is **GATED**. The old 117-entry compatibility registry is split into
five disjoint architectural roles while preserving lookup and visibility behavior.
The user reported the requested `go test ./...` checkpoint clean after the M3.a
stale-test correction.


## HAL redesign — M3 implementation

Milestone 28 is **GATED**. M3.a-d place process context, RNG, file-reader
auxiliary state, Aiki I/O, REPL presentation/user environment, module/cache/help
services, and open-Canvas lifecycle tracking under explicit `GoRuntime` ownership.
Canvas bridge/protocol representation remains reserved for M6.

**Gate:** the user reported `go test ./...` clean after the contained M3.b-d build corrections.

## HAL redesign — M4 implementation

Milestone 29 is **GATED**. The user reported the requested Go test checkpoint clean.

Scope and authority are now separate: raw runtime lookup is grant-based, trusted Aiki sources have exact declared grant sets, and isolated spawn can see prelude vocabulary without inheriting prelude raw authority.

## HAL redesign — M5 implementation

Milestone 30 is **GATED**. The agreed first-order file/system/time/path systems-programmer surface is implemented through 32 canonical host contracts and a 131-entry runtime binding registry. Runtime working directory is explicit, process execution is synchronous/captured, and path manipulation is predominantly pure Aiki.

**Gate:** the user reported `go test ./...` passed after the contained M5 documentation/test correction.


## Milestone 31 — HAL redesign M6

Milestone 31 is **GATED**. Canvas/turtle domain commands are defined in Aiki and cross one canonical `HAL.canvas.command` boundary; Canvas semantic values are opaque handles whose concrete resources and child/session lifecycle are owned by the runtime.

**Gate:** the user reported `go test ./...` passed after M6.c.


## Milestone 32 — HAL redesign M7

Milestone 32 is **GATED**. Obsolete per-command Canvas runtime aliases are removed; canonical host authority is keyed by `HAL.<domain>.<operation>` identities rather than raw substrate binding names; the remaining Aiki test-framework package globals are moved into `GoRuntime`; architecture documentation and executable couplings are strengthened.

**Gate:** after contained corrections to stale authority contract tests and pure-Aiki `path.normalize`, the user reported the final `make validate` passed on the authoritative tree. The HAL redesign implementation is COMPLETE.
