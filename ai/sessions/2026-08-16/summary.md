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


## HAL redesign framing and Phase-I finding

The HAL redesign is driven by the user-facing systems-programmer affordance surface, not by architectural neatness. Every irreducible host crossing is to be understood at three naming levels: Aiki name (meaning), HAL name (contract), substrate name (realization/provenance). Canvas is a pressuring requirement rather than the design driver.

Phase I source inspection of `v0.4.0-alpha-27` found 117 registrations in the current native registry. The registry conflates true host effects/resources, evaluator intrinsics, value primitives, accelerators, observation/tooling services, and runtime/session machinery. The replacement architecture must separate those classes rather than merely rename the registry.

The systems-programmer target surface is already partly present (files, args/env reads, timers, bytes/bits, Store, concurrency/select, standard I/O) but has concrete gaps in filesystem metadata/directory operations, path support, process execution, working-directory operations, and current-time access. Pure semantics should remain in Aiki libraries wherever possible.

A notable pressure-test result is that `time.after` already creates a host-produced receive-only Aiki channel and existing `select` handles it without domain-specific logic. This gives Phase II an existing architectural precedent when considering asynchronous Canvas events.


## HAL redesign Phase II

Source tracing showed that the current environment object combines lexical bindings, shape vocabulary, source provenance, dynamic stack state, observation, module state, and raw-native authority. The redesign separates these concepts. Authority follows the definition of trusted Aiki code rather than the dynamic caller; user callbacks do not acquire a trusted caller's grants. Spawn creates fresh dynamic state without inventing authority.

The baseline also exposed an authority defect: isolated spawn currently builds directly under the prelude environment and therefore inherits `ScopePrelude`, while builtin lookup is scope-based. This reinforces the need to separate prelude vocabulary/isolation from host authority.

Runtime ownership is defined conceptually around host capabilities, I/O, args/environment view, module cache/source provider, RNG, async faults, observation correlation, and resource cleanup. Sessions own evaluator interaction and environments. Concrete host resources must stop leaking substrate types into core values (`*os.File` and Canvas Go state are current examples).

`time.after` provides existing evidence that host-produced receive-only channels work with ordinary Aiki `select`; this is the architectural precedent for considering asynchronous Canvas events without creating a second concurrency model.


## HAL redesign Phase III and migration shape

The replacement design gives canonical `HAL.<domain>.<operation>` identities only to irreducible host crossings. Evaluator intrinsics, native/value primitives, native/FFI providers, and runtime/tooling services are separated conceptually rather than being renamed as HAL operations. Host contracts carry authority, context needs, effect/blocking/lifetime class, error rules, observation identity, optionality, and substrate provenance.

A Runtime owns the host world; an Evaluation Session owns evaluator interaction. Intrinsics receive evaluator context; ordinary HAL calls receive a deliberately small host-call context. Concrete host resources become runtime-owned opaque references rather than Go objects embedded in core semantic values.

Seven serial implementation migrations are planned, beginning with metadata-only three-name coverage and ending with compatibility removal and full executable hardening. The user-facing systems-programmer surface remains the criterion throughout.


## M6 Canvas pressure migration

M6 is GATED. Canvas/turtle domain commands are expressed in Aiki and cross through `_canvas_command` / `HAL.canvas.command`; six narrow Canvas contracts are canonical. `value.Canvas` is now an opaque handle and concrete Canvas process/session/drawing state is runtime-owned in the substrate. The user reported `go test ./...` passed after M6.c.


## M7 compatibility removal and hardening

M7 removes the obsolete per-command Canvas raw aliases from registration, selfhost capture, and trusted-source dependency policy. Host authority is now executable at the canonical contract layer: a trusted source that uses `_file_open` receives `HAL.file.open`, and raw `_file_open` authority alone does not authorize the crossing. Non-HAL primitives retain their implementation-name grants.

The final ownership sweep found and corrected one M3 miss: Aiki test-framework result state was still package-global. It is now owned by `GoRuntime`, and the test command executes each file on a caller-owned runtime so results can be read from the same host world. A runtime-isolation test protects this property.

`docs/hal.md` records the completed architecture outside the AI ledger. Final validation exposed only two contained cleanup defects: stale tests still encoded the superseded scope-based authority rule, and `path.normalize` had imported conventional precedence/short-circuit assumptions inconsistent with Aiki's eager left-to-right evaluation. Both were corrected without weakening the architecture. The user then reported the final `make validate` passed. M7 and the HAL redesign implementation are COMPLETE.

## Distribution formatting and formatter-stable source style

Aiki now separates canonical formatting from project presentation. `aiki fmt` remains the language formatter and does not choose width-driven expansion; `aiki distfmt` applies one repository distribution style to long declarative source. Expanded Aiki lists, calls, and parameter lists are normalized and preserved by `aiki fmt`, which makes the two passes stable rather than adversarial. The central Make `fmt` target runs Go formatting, Aiki canonical formatting, then `distfmt`.

Real-tree probing materially improved the implementation before handoff. Using the full descendant span to detect multiline Aiki calls incorrectly expanded compact calls containing multiline function bodies, so preservation was narrowed to immediate delimiter/argument layout. A line-oriented Go restyler also attempted to alter map-looking text inside a raw string; structural verification prevented the write, and Go selection was replaced with AST-position-driven rewriting. Finally, Aiki restyling was made a bounded formatter/style fixed-point computation so one invocation produces final project layout even from noncanonical input. A complete copied tree now remains unchanged under `distfmt -> canonical formatters -> distfmt`.
