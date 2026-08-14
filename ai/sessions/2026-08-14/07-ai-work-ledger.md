# 07 - Persistent AI work ledger

Status: **GATED**

## Intent

Preserve the working approach that made the profiling implementation reliable across interrupted app-chat turns: one authoritative tree, smaller serial cuts, frequent visible checkpoints, and an in-repository restart record.

## Implemented

Created the top-level `ai/` work-history structure:

```text
ai/
  README.md
  sessions/
    README.md
    2026-08-14/
      README.md
      01-semantic-profiling-core.md
      02-profile-module.md
      03-source-attribution.md
      04-go-cpu-correlation.md
      05-sweep-and-baseline.md
      06-whole-tree-gate.md
      07-ai-work-ledger.md
```

`ai/README.md` defines the durable working rules, status vocabulary, session layout, and resume procedure.

The dated session index is the mutable restart point. Numbered milestones preserve the sequential engineering record: intent, implementation, discoveries, decisions, validation, caveats, and next action.

## Decision

The AI work log is part of the repository work product and should be included with the cut. It is not temporary chat scaffolding.

For future substantial work:

- continue in the same tree;
- create or resume the current dated session;
- work in small serial cuts;
- update the session index as status changes;
- add a numbered milestone when a coherent cut is gated;
- begin a new date directory on a later calendar day and link the prior session;
- never claim background progress after a turn has ended.

## Validation

- directory and milestone sequence inspected in the authoritative tree;
- today's prior profiling ledger was decomposed into the sequential dated milestone files;
- no second top-level progress ledger remains to compete with `ai/` as the source of truth.

## Next action

Use this structure for the next implementation/review activity. If work resumes on another date, create `ai/sessions/YYYY-MM-DD/README.md`, point it to this session, and state the exact continuation point before making changes.
