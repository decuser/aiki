---
name: aiki
description: >
  Use this skill for any substantial conversation involving the Aiki programming
  language project: implementation, review, design, planning, proposals, papers,
  repository work, or any session involving an Aiki tarball, the ai/ working
  record, proposals, docs, tests, or language-specific architecture. The working
  method applies to coding and non-coding work alike.
---

# Aiki working method

## Purpose

Aiki work follows a serial-cut, evidence-gated, restartable method. The goal is
not merely to produce code or prose, but to leave behind a repository state from
which another AI instance — or the same instance after interruption — can answer
without guessing:

1. What is true now?
2. Why is it that way?
3. What exactly should happen next?

The repository, not the chat, is the durable record.

The user is the language designer and works from Aiki's observable behavior and
documented surface. Do not make reading the Go implementation a prerequisite for
understanding or deciding language behavior.

## Core principles

1. **One authoritative working tree.** Work in one tree unless a separate tree is
   explicitly requested. Do not manufacture a new source archive for every cut.
2. **Small serial cuts.** Constrain the next coherent change, complete it, validate
   it, and record it before extending the work.
3. **Evidence before claims.** A cut is not `GATED` until its stated validation has
   actually run. Distinguish inspection, compilation, unit tests, integration
   tests, race tests, smoke/gold checks, and environment-limited validation.
4. **Record discoveries as well as edits.** Invalid assumptions, design flaws,
   limitations, measurement artifacts, and unexpected behavior are engineering
   results and must survive the chat.
5. **Prefer correction over accretion.** When later evidence disproves an earlier
   assumption, correct the design or implementation and record the correction;
   do not layer a workaround over a false premise.
6. **Preserve restartability.** At any real interruption or handoff, leave an exact
   status and next action in the current session record.
7. **Make limitations explicit.** Record unavailable toolchains, dependencies,
   network constraints, disposable harnesses, and anything that limits the
   strength of a validation claim.
8. **Keep generated artifacts intentional.** Distinguish retained evidence from
   disposable output. Do not allow accidental files to become part of the tree.
9. **Single authority for facts.** When multiple subsystems need the same fact,
   locate the declaration that owns it, derive reusable facts near that authority,
   and make consumers use the derived view rather than reconstructing it.
   Consumer-specific meaning, representation, and policy remain local.
10. **Keep the user informed during substantial work.** Report meaningful
    checkpoints, discoveries, failures, and validation results. Never imply that
    work continues after the turn has ended.

## Project lifecycle

Substantial work normally follows this flow:

```text
finding / idea
    ↓
proposal (when the work needs an explicit design contract)
    ↓
named project branch
    ↓
small serial cuts
    ↓
evidence gates
    ↓
project reconciliation
    ↓
non-fast-forward merge to master
    ↓
validation of integrated master
    ↓
synchronize meaningful branches and master to remotes
```

Not every small repair needs a proposal. Use one when the work introduces or
changes architecture, language behavior, a cross-cutting invariant, or enough
scope that intent could drift during execution.

### Findings, proposals, bugs, and sessions

These artifacts have different jobs:

- `docs/audit-findings.md` is the repository-wide ledger of credible engineering
  findings from audits and reviews. Findings receive stable IDs and a disposition
  such as `OPEN`, `PLANNED`, `RESOLVED`, `ACCEPTED`, or `DEFERRED`.
- `proposals/` defines bounded intended work: what will change, why, constraints,
  acceptance criteria, planned cuts, and which audit findings it addresses.
- `buglist.md` contains current unresolved defects. It is not a history of every
  observation, accepted limitation, or resolved audit finding.
- `ai/sessions/` records execution: decisions, discoveries, departures from the
  proposal, validation evidence, caveats, and restart state.

A project may close audit findings without adding them to the buglist. A finding
that remains a genuine unresolved defect should remain represented in the
repository-wide defect record.

## Serial cuts and status

Use these status words consistently:

- `PLANNED` — defined but not started.
- `ACTIVE` — currently being changed or validated.
- `BLOCKED` — cannot proceed until a stated condition changes.
- `GATED` — the cut's implementation or decision and its stated validation are
  complete.
- `COMPLETE` — the project/session has no remaining planned work before normal
  review, merge, or delivery.
- `SUPERSEDED` — later work intentionally replaced the earlier conclusion.

A cut should have one coherent purpose. If a gate exposes a different defect,
record it and either add a new bounded cut or defer it; do not silently fold an
unrelated repair into the current claim.

Golds and other evidence artifacts may be corrected while establishing a new
baseline if the expected artifact itself is wrong. After a baseline is gated,
change evidence only for an explicit, planned behavior or representation change.

## Working record

The durable AI record lives in the repository:

```text
ai/
  README.md                    durable working method
  sessions/
    README.md                  pointer to current/latest session
    YYYY-MM-DD[-project]/
      README.md                live index and restart point
      summary.md               curated conceptual narrative
      01-<milestone>.md        durable milestone record
      02-<milestone>.md
      ...
```

The exact dated-directory naming convention may include a project suffix when
multiple substantial projects occur on the same date.

### Session files

- The dated `README.md` is the mutable project/session index. It states current
  status, gated milestones, unresolved discoveries, and the exact next action.
- `summary.md` preserves the conceptual story: the problem, how the reasoning
  developed, important distinctions, surprises, and resulting architecture. It
  is curated provenance, not a chat transcript.
- Numbered milestone files are the durable engineering record. Once gated, edit
  them only to correct factual errors or add clearly marked follow-up notes.

Maintain the record as work proceeds, but normally **deliver the consolidated
session material at project completion or at a genuine interruption/handoff**.
Do not generate a separate `ai/` tarball after every routine cut merely because a
cut completed.

For non-coding Aiki work — design discussion, paper review, proposal development,
planning, or code review without edits — use the same record when the discussion
produces durable decisions, findings, or next actions that are not already
captured elsewhere.

## Starting or resuming work

When substantial Aiki work begins:

1. Read this working method.
2. If a baseline/archive was supplied, inspect it rather than assuming the chat
   describes its current state.
3. Read `ai/sessions/README.md`, then the current/latest session `README.md` and
   latest relevant milestone.
4. Read the controlling proposal and any audit-finding IDs it cites.
5. Inspect Git status, branch, and relevant tree state before making changes.
6. Continue from the recorded next action or create a new project session when
   starting genuinely new work.

Do not ask the user to repeat information that the repository record already
contains.

## Git workflow

Substantial proposal/feature work uses named project branches and preserves the
development line in history.

1. Create a named project branch.
2. Commit coherent work on that branch.
3. Push meaningful project branches to all configured remotes unless there is a
   specific reason not to.
4. Integrate completed project branches into `master` with a **non-fast-forward
   merge**. Do not fast-forward completed project branches.
5. Validate the integrated `master` tree.
6. Push the merged `master` and meaningful project branches to the remotes.
7. After merge, validation, and synchronization, the local project-branch pointer
   may be deleted to keep the local branch list small. The merge ancestry and
   remote branch refs remain.

Typical lifecycle:

```text
git switch -c <project-branch>
# work, validate, commit, push branch

git switch master
git merge --no-ff <project-branch>
make validate
# push master and project branch to configured remotes

git branch -d <project-branch>   # optional local cleanup
```

Prefer non-fast-forward merges globally:

```text
git config --global merge.ff false
```

Release tags are separate from project branches. Tags identify exact release
commits; non-fast-forward merges preserve bounded development lines.

## Validation

Use the strongest relevant evidence for the cut. The standard repository gate is
`make validate`; when a cut is small, focused tests may be used during execution,
but they do not substitute for the final repository gate when the project claims
integration-level correctness.

The standard validation surface includes, as applicable:

```text
go test ./...                 Go tests
./aiki test ./...             Aiki-native tests
./aiki smoke test/behavior/   behavior/gold tests
./aiki enginesmoke ...        grammar coverage and engine structural golds
./aiki treecheck              distribution-tree integrity
race checks                   concurrent packages when relevant
```

When environment constraints prevent the authoritative gate, record exactly what
was run and what remains unverified. A locally compatible harness is useful
evidence but must not be described as the full gate.

For non-coding work, validation means that decisions and findings are internally
consistent, rationale is recorded, open questions have dispositions, and the next
action is explicit.

## Handoff and delivery

Before a project/session ends or work is handed off:

- reconcile the proposal against what was actually implemented;
- update any referenced audit findings and unresolved bugs;
- record gated and blocked cuts accurately;
- distinguish intentional behavior changes from baseline-preserving work;
- update the session `README.md` and `summary.md`;
- ensure the working tree contains no unexplained generated files;
- state the exact next Git or engineering action if the project is not complete.

When delivering only the AI working record, package it rooted at `ai/` so it can
merge directly into the repository:

```bash
tar czf ai-session.tgz ai/
```

Do not overwrite durable working-method files unless they were intentionally
changed as part of the project.

## Minimal project orientation

Aiki's detailed architecture belongs in its source, proposals, and documentation,
not in this working-method file. For orientation only:

```text
cmd/         command-line tools
engine/      syntax, evaluator, runtime, runners
lib/         Aiki standard-library modules
docs/        user-facing and architectural documentation
proposals/   bounded design contracts
test/        behavioral, structural, invariant, fuzz, and property tests
ai/          AI working provenance
```

Current implementation state and priorities belong in the latest session/project
record, not here. This file should change only when the **method** changes.
