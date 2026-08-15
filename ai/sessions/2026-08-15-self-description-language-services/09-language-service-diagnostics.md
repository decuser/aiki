# Milestone 09 — Service contract, documents, and diagnostics

Status: GATED

## Implementation

Created `engine/language` as the editor-independent service layer.

- `Document` carries stable identity, path, source, and client generation/version.
- `Diagnostic` carries Aiki-native position, severity, source (`lex`, `parse`, `lint`), and message.
- `Service.Diagnostics` runs the real lexer/parser and then the extracted structural analyzer.
- Lexer/parser failures now expose `syntax.SourceError` so services can obtain position/category/message without scraping rendered caret text; `Error()` preserves the existing CLI rendering.
- Structural lint analysis moved from the lint command into `engine/language`.
- The service core does **not** import the Go substrate. A narrow `language.Catalog` contract supplies runtime/workspace facts; `engine/language/workspace` is the concrete Go-runtime/module adapter.
- `cmd/subcommands/tools/lint` is now an adapter over the extracted structural analysis.

## Discovery

The prior lint implementation directly depended on `substrate.NewGoRuntime` and `ModuleRegistry`. Carrying that dependency into the service core would have violated the intended neutral boundary. The catalog inversion was therefore required before LSP work.

## Validation

A disposable Go-1.23 copy (module version lowered only in the disposable copy) ran:

```text
go test ./engine/language ./engine/syntax
```

Both packages passed. This validates the neutral core and structured syntax-error changes without requiring the unavailable Go-1.24 toolchain/Ebiten download in this environment. Full repository validation remains a Phase-II handoff requirement on the authoritative machine.

## Next action

Cut II.2: add optional, behavior-neutral language-service observation and instrumentation at actual service choke points.
