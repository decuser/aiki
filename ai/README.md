# AI working method

This directory is part of the repository and part of the work product. It exists so an AI can enter an ongoing implementation effort, understand how work is conducted, and resume from an interrupted session without reconstructing state from chat history.

## Working rules

1. **Keep one authoritative working tree.** Do not create a new tree or archive for each stage unless explicitly requested.
2. **Work in small serial cuts.** Finish, validate, and record one coherent cut before extending the next.
3. **Keep the user informed while work is active.** Report meaningful checkpoints: what changed, what was discovered, what failed, and what is being validated. Never imply work continues after a turn has ended.
4. **Record progress in the repository as it happens.** The session log is not disposable scratch material. It is engineering provenance and is included in the resulting cut.
5. **Gate claims with evidence.** A cut is not `GATED` until its stated validation has actually run. Distinguish source inspection, compilation, unit tests, integration tests, race tests, smoke/gold checks, and environment-limited validation.
6. **Record discoveries, not just edits.** If work exposes a design flaw, measurement artifact, invalid assumption, or limitation, record the finding and the decision that followed.
7. **Preserve restartability.** Before a turn can reasonably end, update the current session index with the exact status and next action.
8. **Do not hide environmental limitations.** Record toolchain, dependency, network, or runtime constraints and any disposable harness used to work around them. Never treat harness code as product code.
9. **Prefer correction over accretion.** If a later cut reveals an earlier assumption was wrong, correct the implementation and record the correction rather than layering a workaround on top.
10. **Keep generated artifacts intentional.** Record which outputs are evidence worth retaining and which are temporary. Do not commit disposable stubs, caches, or environment-specific binaries by accident.

## Session layout

Sessions live under:

```text
ai/sessions/YYYY-MM-DD/
```

Each dated directory contains:

```text
README.md                  current session state and restart point
summary.md                 curated conceptual/design narrative for the session
01-<milestone>.md          first gated milestone
02-<milestone>.md          second gated milestone
03-<milestone>.md          ...
```

Milestones are sequential within the date. A milestone records:

- intent;
- relevant implementation changes;
- discoveries and decisions;
- validation actually performed;
- caveats or limitations;
- the next step established at that point.

The dated `README.md` is the session's live index. Update it whenever status changes. `summary.md` preserves the conceptual/design narrative: the problem, how the thinking developed, important distinctions, surprises, and resulting architecture. It is a curated summary, not a chat transcript. Numbered milestone files are the durable engineering record of completed gates; edit them only to correct factual errors or add clearly marked follow-up notes.

## Starting or resuming work

When beginning work:

1. Read this file.
2. Open the most recent directory under `ai/sessions/`.
3. Read that session's `README.md`.
4. Read the latest numbered milestone it references.
5. Inspect the working tree before assuming the recorded state still matches the files.
6. Continue from the recorded **Next action**.

When starting a new calendar-date session, create a new date directory. Its `README.md` should link the prior session and state what is being continued or started.

## Status vocabulary

Use these words consistently:

- `PLANNED` - defined but not started.
- `ACTIVE` - currently being changed or validated.
- `BLOCKED` - cannot proceed until a stated condition changes.
- `GATED` - implementation and stated validation for the milestone completed.
- `COMPLETE` - the session has no further planned implementation before review/delivery.
- `SUPERSEDED` - later work intentionally replaced the milestone's conclusion.

## Principle

The log should make it possible for another competent AI—or the same AI after an interrupted chat—to answer three questions without guessing:

1. What is true in the tree now?
2. Why is it that way?
3. What exactly should happen next?
