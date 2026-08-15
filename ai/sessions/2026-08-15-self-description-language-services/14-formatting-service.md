# Milestone 14 — formatting service

Status: ACTIVE — implementation and disposable local gates complete; authoritative `make validate` pending.

## Intent

Implement Cut II.6 from the active proposal. Formatting remains one canonical implementation with its existing parse-preservation safety gate. The language service and LSP must adapt that authority rather than reimplement it.

## Planned extraction

- Move the reusable formatter implementation below command packages.
- Keep `aiki fmt` as a command adapter over the same implementation.
- Add `Service.Format(Document)` with language-service observation/instrumentation.
- Add LSP `textDocument/formatting` as a full-document edit projection.
- Prove command-level formatter output and service output are identical for shared fixtures.
- Preserve the formatter's existing rejection of invalid source or AST-changing transformations.

## Gate

Affected package tests pass locally, formatter equivalence is explicit, and the authoritative `make validate` passes before II.6 is marked GATED.

## Implementation

- Moved the canonical formatter implementation from `cmd/subcommands/tools/fmt` to neutral `engine/formatter`.
- Kept `aiki fmt` as a thin compatibility/command adapter; no printer implementation remains in `cmd/`.
- Updated lint format checking and debug formatting to consume `engine/formatter` directly.
- Added `Service.Format(Document)` plus format request/produced/rejected observer and probe events.
- Added LSP `documentFormattingProvider` and `textDocument/formatting`, projected as a full-document edit.
- LSP formatting errors preserve formatter rejection rather than manufacturing edits.

## Evidence

A disposable Go 1.23 copy (with only the graphics backend stubbed because network/toolchain dependencies are unavailable here) passed:

```text
go test ./engine/formatter
go test ./engine/language
go test ./cmd/subcommands/tools/fmt
go test ./cmd/subcommands/tools/lsp
go test ./cmd/subcommands/tools/treecheck
go test ./cmd/subcommands/tools/lint
go test ./cmd/subcommands/dev/debug
```

The formatter/service equivalence test calls both the language service and canonical `engine/formatter` on the same source and requires byte-identical output. Invalid source is rejected by both the service and LSP path. `git diff --check` passes.

A full disposable command build could not run because the environment lacks network access to fetch the existing readline dependency; this is environmental and does not gate the authoritative tree.

## Handoff

Run `make validate` on the authoritative tree. If it passes, mark II.6 GATED and begin II.7 completion and hover/inspect.
