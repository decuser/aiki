# Milestone 13 — symbol/definition service and nvi tags

Status: ACTIVE — implementation and disposable package gates complete; authoritative `make validate` and live nvi tag jump pending.

## Intent

Implement Cut II.5 from the active proposal. Static symbol and definition knowledge belongs to `engine/language`; LSP and classic tags are projections over the same authority.

## Implementation

- Added `engine/language/symbols.go` with a lexically scoped source-definition index.
- Scope behavior follows Aiki structural rules: top-level bindings are precollected; ordinary blocks do not create scopes; function parameters do; match and select arms do.
- Added service `Symbols` and `Definition` operations plus observer/probe events.
- Added LSP `textDocument/definition` and `textDocument/documentSymbol` adapters with UTF-16-to-Aiki byte-position translation.
- Added `aiki tags [-o file] <paths...>` producing classic tags entries from top-level source definitions only. Local bindings remain available to the service/LSP but are intentionally excluded from nvi tags to avoid ambiguous global tag names.
- Added `tags` to the main command dispatcher.

## Evidence

A disposable Go 1.23 copy with only the graphics backend stubbed passed:

- `go test ./engine/language`
- `go test ./cmd/subcommands/tools/tags`
- `go test ./cmd/subcommands/tools/lsp`

The authoritative tree was not downgraded and no disposable stub is retained.

## Handoff gate

1. Run `make validate` on the authoritative toolchain.
2. Run `./aiki tags -o tags <chosen Aiki source path or directory>`.
3. In nvi with `:set tags=./tags`, place the cursor on a top-level name and use `^]`; confirm the jump reaches the same source definition that LSP definition reports.
