# 06 — Phase I handoff

Status: ACTIVE — implementation complete; authoritative gate pending.

## Phase boundary reached

Phase I implementation is complete through Cut I.4:

1. Aiki-native grammar token/newline authority extraction;
2. authority/representation inventory;
3. independent Aiki lexer;
4. independent Aiki skip/newline normalizer;
5. independent Aiki recursive-descent parser;
6. reusable lexical/newline conformance fixtures and reuse of the existing
   engine parse-gold corpus as the neutral syntax surface.

The Go and Aiki implementations remain independent. Enumerated duplicated facts
are executable-coupled to grammar authority.

## Validation evidence available here

- `gofmt` completed on all added/changed Go files.
- Disposable Go-1.23 syntax package: `go test ./engine/syntax/...` passed after
  extracting `syntax.NormalizeTokens`.
- Real Go syntax engine parsed every `selfhost/*.ai` file in the disposable
  syntax-only copy.
- Disposable non-graphics runtime using the real Aiki engine/HAL/prelude/module
  path matched 4/4 lexer fixtures, 4/4 newline fixtures, and 10/10 engine parse
  golds.
- No disposable runtime, graphics stub, generated harness, or temporary binary
  remains in the working tree.

## Validation not claimed

The repository's required Go 1.24 toolchain cannot be downloaded in this
execution environment. Therefore `make validate` and the new permanent Go
invariant tests have not run under the authoritative toolchain. This phase is
not marked GATED or COMPLETE.

## Exact next action

On the authoritative development machine run:

```text
make validate
```

If it passes, update Phase I/session records to GATED and begin Phase II Cut
II.0 from this same tree. If it fails, preserve the failure output and correct
Phase I before starting language-service implementation.

## Deferred design issue

Non-ASCII column units remain intentionally unresolved: Go currently counts
UTF-8 bytes while Aiki string indexing is rune-based. Resolve this as a
language-service/source-position policy before editor/LSP position translation.

## Follow-up — authoritative validation correction

The first authoritative `make validate` attempt reached `./aiki fmt ./...` and
failed because the lexical and newline conformance inputs had been stored with
`.ai` extensions. Several of those files are intentionally lexical fragments or
syntactically invalid programs; the formatter correctly rejected them before
the invariant tests could run.

This was a fixture-packaging error, not a lexer/parser conformance failure. The
raw front-end inputs are data, not executable Aiki programs. They have therefore
been renamed from `*.ai` to `*.input`, and the invariant tests now discover the
`*.input` files while retaining the adjacent reviewed `*.tokens` projections.
No fixture contents or expected projections changed.

Status remains ACTIVE. Re-run the full authoritative gate from the corrected
same tree before advancing to Phase II.

## Follow-up — treecheck integration correction

The second authoritative `make validate` attempt passed build, formatting, and
Aiki lint apart from one naming warning, then failed at `aiki treecheck`.
Treecheck correctly identified the newly introduced self-host implementation and
syntax-conformance corpus as distribution artifacts with no recognized
relationship. The local validation patch file was also reported because it had
been copied into the repository root; it is not a repository artifact and must
be removed before validation.

The correction preserves treecheck as an invariant rather than adding broad
allow-list exceptions:

- `selfhost/*.ai` is a recognized first-class self-host implementation source;
- `selfhost/token_authority.gold` is structurally owned by
  `selfhost/token_authority.ai`;
- each `test/conformance/syntax/**/*.input` must have an adjacent `.tokens`
  projection, and each `.tokens` projection must have an adjacent `.input`;
- treecheck tests cover both valid relationships and missing-companion errors;
- the lint-only `parse_BINOP` helper is renamed `parse_binop`; its emitted
  `BINOP` syntax node remains unchanged.

The treecheck package passes its tests in a disposable Go 1.23 module, and the
corrected checker reports zero errors and zero orphans against the reconstructed
Phase-I tree. The authoritative gate remains pending until `make validate` is
re-run on the development machine.
