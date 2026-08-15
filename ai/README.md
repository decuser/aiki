---
name: aiki
description: >
  Use this skill for any conversation involving the Aiki programming language
  project — implementation work, code review, design discussion, paper review,
  planning, or any session where the user provides an aiki tarball or refers to
  the Aiki codebase, its ai/ working record, proposals, docs, or lib modules.
  Also trigger when the user mentions concepts specific to Aiki: HAL boundary,
  exact rationals, isolated spawn, semantic profiling, store/bits/select, the
  7094 emulator, executable documentation, or the SPLASH-E submission. Trigger
  even for non-coding sessions (discussion, review, planning) because the
  working method applies to all substantial Aiki work and the session record
  must be maintained.
---

# Aiki project skill

## What Aiki is

Aiki is a programming language hosted in Go. Its defining commitments are
exact rational arithmetic (no hidden floating-point state) and isolated
spawn-based concurrency (no shared mutable memory between spawned
computations). The implementation comprises a lexer, parser, and tree-walking
evaluator, a HAL (hardware abstraction layer) contract separating the
evaluator from the Go substrate, a scope-gated builtin registry, a module
system, a prelude, and a standard library.

The user (Will) is the language designer. He works from observable behavior
and documented surface, not from reading the Go implementation directly.
This is deliberate and constitutive of the project's method — do not suggest
he read the implementation code as a remedy for any problem.

## Working method

All substantial Aiki work — including non-coding sessions — follows a
serial-cut, evidence-gated, restartable method. The authoritative record
lives in `ai/` inside the repository.

### Rules

1. Keep one authoritative working tree. Do not create a new tree or archive
   for each stage unless explicitly requested.
2. Work in small serial cuts. Finish, validate, and record one coherent cut
   before extending the next.
3. Keep the user informed while work is active. Report meaningful
   checkpoints: what changed, what was discovered, what failed, and what is
   being validated. Never imply work continues after a turn has ended.
4. Record progress in the repository as it happens. The session log is not
   disposable scratch material. It is engineering provenance.
5. Gate claims with evidence. A cut is not GATED until its stated validation
   has actually run. Distinguish source inspection, compilation, unit tests,
   integration tests, race tests, smoke/gold checks, and environment-limited
   validation.
6. Record discoveries, not just edits. If work exposes a design flaw,
   measurement artifact, invalid assumption, or limitation, record the
   finding and the decision that followed.
7. Preserve restartability. Before a turn can reasonably end, update the
   current session index with the exact status and next action.
8. Do not hide environmental limitations. Record toolchain, dependency,
   network, or runtime constraints and any disposable harness used to work
   around them.
9. Prefer correction over accretion. If a later cut reveals an earlier
   assumption was wrong, correct the implementation and record the correction
   rather than layering a workaround on top.
10. Keep generated artifacts intentional. Record which outputs are evidence
    worth retaining and which are temporary.

### Session layout

```text
ai/
  README.md                    durable working rules (do not overwrite)
  sessions/
    README.md                  pointer to current session
    YYYY-MM-DD/
      README.md                live session index / restart point
      summary.md               curated conceptual narrative
      01-<milestone>.md        first gated milestone
      02-<milestone>.md        ...
```

- The dated `README.md` is the mutable restart point. Update it whenever
  status changes.
- `summary.md` preserves the conceptual/design narrative: the problem, how
  thinking developed, important distinctions, surprises, and resulting
  architecture. It is a curated summary, not a chat transcript.
- Numbered milestone files are the durable engineering record. Edit them only
  to correct factual errors or add clearly marked follow-up notes.

### Status vocabulary

Use these words consistently in milestone and session records:

- `PLANNED` — defined but not started.
- `ACTIVE` — currently being changed or validated.
- `BLOCKED` — cannot proceed until a stated condition changes.
- `GATED` — implementation and stated validation completed.
- `COMPLETE` — the session has no further planned work before review/delivery.
- `SUPERSEDED` — later work intentionally replaced the milestone's conclusion.

### Git workflow

Substantial proposal/feature work should preserve its development line in Git history.

1. Create a named project branch for the work.
2. Commit coherent project milestones on that branch and push meaningful branches to all configured remotes unless there is a specific reason not to.
3. Integrate completed work into `master` with a **non-fast-forward merge** so the branch-and-merge structure remains visible in the commit graph. Do not fast-forward completed project branches.
4. Validate the integrated `master` tree before pushing the merge to remotes.
5. After the branch is merged, validated, and present on the remotes, the local branch pointer may be deleted to keep the local branch list uncluttered. Deleting the local pointer does not remove the branch ancestry or merge structure from history.
6. Keep release tags separate from project branches: tags mark exact release commits; non-fast-forward merges preserve project development lines.

Typical lifecycle:

```text
git switch -c <project-branch>
# work, validate, commit, and push the branch to remotes
git switch master
git merge --no-ff <project-branch>
# validate integrated master
git push <remote> master
# ensure <project-branch> is pushed to the remotes
git branch -d <project-branch>   # optional local cleanup after merge/sync
```

Set Git to prefer this merge policy by default:

```text
git config --global merge.ff false
```

### Starting or resuming work

When a tarball arrives or Aiki work begins:

1. Read this skill.
2. If a tarball was provided, extract it and find the most recent directory
   under `ai/sessions/`.
3. Read that session's `README.md` and the latest numbered milestone.
4. Inspect the working tree before assuming the recorded state matches files.
5. Continue from the recorded next action.

When starting a new calendar-date session, create a new date directory. Its
`README.md` should link the prior session and state what is being continued
or started.

### Non-coding sessions

The working method applies to sessions that involve no code changes —
design discussion, paper review, planning, proposal writing, code review
without modification, or any other substantial exchange about Aiki.

For such sessions:

- Create or continue a dated session directory.
- Record milestones for substantive decisions, reviews, or findings.
  A milestone for a non-coding session records intent, the discussion or
  review conducted, decisions reached, and any action items established.
- Use status `GATED` when the discussion reached a conclusion or decision;
  `COMPLETE` when the session has no further planned discussion.
- Write a `summary.md` when the session produced conceptual or design
  insight worth preserving beyond the milestone record.

The test is whether another instance of Claude — or the same instance after
an interrupted chat — could answer three questions without guessing:

1. What is true now?
2. Why is it that way?
3. What exactly should happen next?

If the session produced answers to any of those that aren't already recorded,
it warrants a session entry.

## Delivering session records

At the end of any session that produced `ai/` content, package the updated
`ai/` directory as a drop-in tarball the user can extract and merge into the
repository.

```bash
# Build a tarball rooted at ai/ that merges cleanly
tar czf ai-session.tgz ai/
```

The tarball must:
- Be rooted at `ai/` so it extracts into the right place.
- Never overwrite `ai/README.md` or `ai/sessions/README.md` unless those
  files were intentionally changed during the session.
- Only add or update files under the current session's dated directory,
  plus updating `ai/sessions/README.md` to point to the current session
  if a new date directory was created.
- Never include files outside `ai/`.

Present the tarball to the user at the end of the session.

## Project structure

Key directories and what they contain:

```text
cmd/                    CLI entry points (aiki, subcommands)
engine/
  profiling.go          semantic probe types (SemanticKind, SemanticProbe, etc.)
  runner/               session, profile, smoke runners
  runtime/
    hal/                HAL contract (RuntimeContract, ContextCallable, etc.)
      substrate/        Go implementation of HAL (builtins, registry, store)
    help/               help system
    prelude/            prelude loader
  semantics/
    evaluator/          tree-walking evaluator, counters
    value/              value model (Number, List, Store, Channel, Env, etc.)
  syntax/               lexer, parser, grammar
lib/                    standard library modules (Aiki source + help + doc)
docs/                   user-facing documentation
proposals/              design proposals
extra/                  editors, profiling workloads, samples
test/                   boundary, canary, contract, fuzz, invariant, property tests
ai/                     AI working record (this method)
```

## Key architectural concepts

- **HAL boundary**: the evaluator delegates all host effects through
  `RuntimeContract`. Builtins are scope-gated (prelude vs user).
- **Value model**: exact rationals via `big.Rat`, persistent immutable lists,
  explicit mutable `Store`, opaque `Channel`. No floats.
- **Isolated spawn**: spawned computations get `NewIsolatedEnclosedEnv` —
  separate call stack, shared prelude vocabulary, no mutable state leakage.
  `Store` is an explicit, documented exception.
- **NewCallEnv**: lexical bindings from the defining environment, dynamic
  execution state (stack, probe) from the caller. This split is load-bearing
  for concurrency correctness.
- **Executable documentation**: doc examples are parsed and run as
  subprocess tests. The doc corpus is the spec; the Go is a fungible
  implementation.
- **Semantic profiling**: deterministic Aiki-level counts (arithmetic,
  comparison, call, iteration, index, send, receive, store_read, store_write)
  plus opt-in source attribution, plus correlated Go CPU profiling via pprof
  labels across the HAL boundary.

## Validation workflow

The standard validation sequence (when doing code work):

```text
go test ./...                 all Go tests
Aiki-native tests             lib/*/test_*.ai via test framework
behavior smokes               test/behavior/*.ai vs .gold
grammar coverage              32 productions across 10 inputs
engine gold check             test/structure/engine/ inputs
race checks                   go test -race on concurrent packages
```

For non-coding sessions, validation means: session record is internally
consistent, decisions are recorded with rationale, and next actions are
specified.

## Current state and priorities

Read the latest `ai/sessions/` entry for current state. The identified
feature priority order is: `bits` → `store` → `select` → profiler.
Store and profiler are now implemented. The executable documentation effort
(Stages 4–5) is in progress.
