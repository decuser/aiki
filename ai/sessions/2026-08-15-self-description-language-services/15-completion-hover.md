# Milestone 15 — completion and hover/inspect

Status: ACTIVE — implementation and disposable local gates complete; authoritative `make validate` pending.

## Intent

Implement Cut II.7 from the active proposal. Completion must derive from Aiki's existing lexical/runtime visibility rules, and hover must project authored Aiki help or known source-definition facts. Neither LSP nor an editor adapter may maintain a second semantic inventory.

## Implementation

- Added neutral `Service.Completion` and `Service.Hover` operations.
- Completion reuses the lexical scope index established by II.5. The symbol walker records the names visible at each source NAME token, preserving Aiki's actual top-level, function, match-arm, and select-arm scope rules.
- Extended the neutral workspace catalog from runtime builtin enumeration to `VisibleNames(scope)` plus authored `Help(name)`. The Go workspace adapter combines runtime-visible primitives with the embedded Aiki prelude surface and parses `prelude.help` / `prelude.doc` through the existing help parser.
- Source-defined hover reports the known binding kind and definition position. Prelude hover returns the authored template, one-line help, and documentation when present.
- Added behavior-neutral completion/hover observer events and probe metrics.
- LSP now advertises `completionProvider` and `hoverProvider` and projects the neutral results. It contains no Aiki completion list or documentation database.

## Evidence

A disposable Go 1.23 validation copy, used only because this environment cannot fetch the repository-required Go toolchain/dependencies, passed:

```text
go test ./engine/language/...
go test ./cmd/subcommands/tools/lsp
```

Tests prove:

- lexical completion inside a function includes its parameter/local and mutually visible top-level definitions;
- catalog-provided prelude/runtime names are merged without adapter inventory;
- lexical hover resolves to the same definition authority used by II.5;
- authored prelude help is projected through the catalog;
- LSP completion/hover capabilities and responses come from service results.

No disposable toolchain or graphics stub is retained in the authoritative tree.

## Gate

Run authoritative `make validate`. If it passes, mark II.7 GATED and begin II.8, the thin VS Code client / Phase-II completion handoff.
