# Milestone 08 — Language-service authority/dependency inventory

Status: GATED

## Intent

Identify the smallest acyclic extraction that can support editor-independent diagnostics without allowing LSP, an editor, or a command package to become a second language authority.

## Inventory

- **Document identity/source**: no neutral document type exists yet. Commands pass `(file, source)` directly to lexer/parser/formatter/linter.
- **Lexical/syntax authority**: `engine/syntax` owns the real lexer, grammar-driven normalization, parser, and AST. Phase-I conformance projections already validate the observable front end.
- **Parser diagnostics**: lexer/parser currently return rendered `error` strings. Parse provenance exists internally (`ParseFailure`) but is not exposed as a stable service diagnostic.
- **Structural lint analysis**: authoritative implementation lives under `cmd/subcommands/tools/lint/lint.go`; this is the principal dependency inversion to correct for diagnostics.
- **Formatter**: canonical formatting implementation lives under `cmd/subcommands/tools/fmt`; it already reuses the real parser and has an AST-preservation gate, but is command-owned and must eventually move behind a neutral capability.
- **Runtime/prelude vocabulary**: lint derives visible names from embedded prelude source and substrate `BuiltinNames(scope)`. Help data lives in `engine/runtime/help`; module resolution/export knowledge lives in the substrate module registry and lint's structural module scan.
- **Observation**: `engine.Observer` already provides neutral lexer/parser/evaluator/formatter events; `engine.DiagnosticObserver` is an optional extension.
- **Instrumentation**: `engine/observe.SemanticProbe` demonstrates the desired dependency-neutral probe pattern, but language-service work needs a sibling service-level probe rather than overloading semantic execution counts.
- **Symbols/scopes**: lint has the only reusable structural scope walk today; no definition/reference model exists yet.

## Decision

The first acyclic extraction is a new `engine/language` package that owns editor-independent document/service/result types and the structural analysis currently embedded in the lint command. `cmd/subcommands/tools/lint` becomes a CLI adapter over that capability.

The service will use `engine/syntax` directly and will not expose Go parser structs as its external contract. Phase-I conformance remains the evidence that the syntax machinery agrees with the independent Aiki front end.

Structured syntax errors will be introduced at the syntax boundary only as necessary to preserve position/category/message without parsing rendered error text.

## Next action

Implement Cut II.1: neutral `Document`, diagnostics contract, syntax error projection, and extraction of lint structural analysis into `engine/language`, with CLI/service semantic equivalence tests.
