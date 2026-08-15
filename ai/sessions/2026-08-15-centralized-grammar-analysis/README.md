# Session: Centralized Grammar Analysis

**Date:** 2026-08-15
**Status:** ACTIVE — implementation complete; authoritative `make validate` pending.

Previous session: `../2026-08-14-negative-fixture-tooling/`

## Purpose

Implement `proposals/centralized-grammar-analysis.md`:

> Parse once. Derive shared structural facts once. Consume everywhere.

The project centralizes reusable grammar-derived knowledge without moving parser interpretation, presentation logic, evaluator semantics, or consumer-specific compiled lookup structures into the grammar package.

## Milestones

1. `01-proposal-and-inventory.md` — GATED
2. `02-central-analysis-model.md` — GATED with environment-limited package validation
3. `03-structural-consumers.md` — GATED with source audit and available package validation
4. `04-cached-newline-consumers.md` — GATED with syntax/grammar package validation
5. `05-reconciliation-and-final-gate.md` — ACTIVE

## What is true now

- `Grammar` caches one `grammar.Analysis` after EBNF parsing resolves references.
- Structural analysis owns production names, TokenRefs, AST-producible node types, and per-production terminals.
- Newline analysis is computed once as part of central analysis and cached with its error state.
- evaluator, formatter coverage, and lint grammar-knowledge tests consume central analysis instead of walking grammar expressions.
- parser and enginesmoke consume cached newline analysis rather than rederiving it.
- `AnalyzeNewlineRule()` remains as a compatibility accessor over the cached result.
- the only remaining non-grammar-package expression traversals are intentional: parser interpretation and enginesmoke presentation.

## Validation completed here

The container has Go 1.23.2 while `go.mod` requires Go 1.24.0 and has no network access. Disposable copies with the Go directive lowered to 1.23 were used only for environment-compatible validation; the authoritative tree was not altered for this purpose.

Passing local package checks:

- `go test ./engine/syntax/grammar`
- `go test ./engine/syntax`
- `go test ./cmd/subcommands/tools/fmt`

Evaluator/invariant/linter full package execution is environment-limited because the uncached Ebitengine dependency cannot be downloaded.

## Exact next action

Merge the delivered overlay into the user's current tree and run:

    make validate

If it passes, mark milestone 05 GATED and the session COMPLETE. If it fails, treat the failure as the next serial cut rather than reblessing behavior.
