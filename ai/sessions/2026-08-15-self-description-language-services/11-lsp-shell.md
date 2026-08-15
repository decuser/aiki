# Milestone 11 — `aiki lsp` protocol shell

Status: GATED

## Implementation

Added `cmd/subcommands/tools/lsp` and registered `aiki lsp`.

Initial protocol surface:

- Content-Length framed JSON-RPC 2.0 over stdio;
- `initialize`, `initialized`, `shutdown`, `exit`;
- full text synchronization for `didOpen`, `didChange`, `didClose`;
- `textDocument/publishDiagnostics` from `engine/language.Service`;
- no LSP structs leak into the language-service contract.

The adapter declares UTF-16 LSP positions and explicitly translates Aiki's existing 1-based byte columns into 0-based UTF-16 positions. This resolves the Phase-I byte/rune discovery at the protocol boundary without changing Aiki's internal source-position policy.

## External specification check

Implementation was checked against the current Microsoft LSP specification for base framing, JSON-RPC message shape, text synchronization, diagnostics, and negotiated position encoding. No third-party LSP library or new module dependency was added.

## Validation

In a disposable Go-1.23 copy with the graphics backend replaced only for compilation (no authoritative source changed):

```text
go test ./engine/language/... ./engine/syntax ./cmd/subcommands/tools/lint ./cmd/subcommands/tools/lsp
```

passes. LSP tests cover initialize, diagnostic publication, shutdown/exit framing, and non-ASCII byte-column to UTF-16 conversion.

Full repository validation remains a Phase-II handoff requirement on the authoritative machine.

## Next action

Cut II.4: bring Xed lexical support back to current grammar authority, add executable coupling, and provide the thinnest practical Xed client for `aiki lsp`.
