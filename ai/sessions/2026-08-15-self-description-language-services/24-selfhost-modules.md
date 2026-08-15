# Milestone 24 — Self-hosted module loading

Status: GATED locally; authoritative repository gate pending.

## Decision

Cut III.4 self-hosts Aiki source modules end to end. The loader resolves source paths, reads source through ordinary file facilities, then uses the independent lexer, newline normalizer, parser, and evaluator. It creates an isolated interpreted module environment, records package/export metadata, and caches the resulting interpreted module value.

The boundary is explicit:

- Aiki source modules are loaded and evaluated by the self-hosted interpreter.
- HAL primitives remain host capabilities.
- A privileged `lib/selfhost/bootstrap` module privately captures the HAL function values needed by blessed standard-library source and configures the self-host evaluator. It exports only `run(source, file_name)`; raw `_` bindings are never exposed to callers.

This supersedes Milestone 23's unresolved module-value boundary.

## Implementation

- `selfhost/runtime.ai` now records package name, exports, source file identity, and outer environment access through runtime methods.
- `selfhost/evaluator.ai` implements module resolution, cache, import/use/export intrinsics, package statements, module field access, qualified pipe access, and relative `./` imports.
- `lib/selfhost/bootstrap.ai` is the scoped bootstrap entrypoint.
- `test/invariant/selfhost_module_test.go` couples representative ordinary, transitive, HAL-backed, and selective imports to the reference implementation.

## Evidence

Disposable runtime probes successfully self-host-loaded all 20 current standard-library module families. Representative execution matched the reference implementation, including:

- `list.sum([1,2,3,4])` → `10`;
- `string.trim("  hello  ")` → `"hello"` through transitive `list` + `math/ffi` imports;
- relative `./import_target` loading;
- pure and FFI-backed modules.

## Follow-up: executable documentation correction

Authoritative `make validate` exposed that `lib/selfhost/bootstrap.doc` had been
written in help-style markup (`# selfhost/bootstrap`, `@func run`) rather than
the repository's `.doc` entry grammar. The module implementation was unaffected.
The document was corrected to a single `run` entry with an executable `# 9`
expectation so `TestDocEntryDisposition`, `TestDocExamplesExecutable`, and
`TestLibDocCoversExports` all observe the same exported surface.
