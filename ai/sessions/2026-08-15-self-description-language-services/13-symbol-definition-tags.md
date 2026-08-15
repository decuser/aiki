# Milestone 13 — symbol/definition service and nvi tags

Status: GATED — authoritative validation and live nvi tag jump passed.

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

## Authoritative and live evidence

- `make validate` passed on the authoritative tree/toolchain.
- Generated tags for the actual file under test with:

  ```sh
  ./aiki tags -o tags extra/samples/string_utils.ai
  ```

- In nvi, `:set tags=./tags` followed by `^]` on the `my_slice` call jumped to the top-level `my_slice` definition; `^T` returns.
- An earlier `tag not found` result was correctly traced to using a tags file generated for `selfhost/` rather than the source corpus being edited. Tags are corpus-specific.

## Operational note

For nvi, regenerate `tags` for the source tree/file being edited. The minimal workflow is:

```sh
./aiki tags -o tags <path>
nvi <file.ai>
```

Then `:set tags=./tags`, `^]` to jump, and `^T` to return.
