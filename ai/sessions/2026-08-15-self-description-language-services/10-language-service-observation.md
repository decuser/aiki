# Milestone 10 — Language-service observation and instrumentation

Status: GATED

## Implementation

Added `engine/language/observe`, a dependency-neutral sibling to the existing semantic observation package.

- Observer events currently cover diagnostics requests and produced diagnostics.
- Probe metrics cover diagnostics requests, lexer runs, parser runs, structural-analysis runs, and diagnostic count.
- `language.Service` accepts optional observer/probe attachments; neither is part of the result contract.

The vocabulary is intentionally small and reflects actual choke points implemented in Cut II.1. Later services may extend it when real queries exist.

## Validation

The disposable Go-1.23 harness runs the same document with observation/probing disabled and enabled and asserts identical diagnostics. Counts/events are asserted separately. `go test ./engine/language ./engine/language/observe ./engine/syntax` passes.

## Next action

Cut II.3: implement `aiki lsp` as a replaceable stdio JSON-RPC adapter over `language.Service`.
