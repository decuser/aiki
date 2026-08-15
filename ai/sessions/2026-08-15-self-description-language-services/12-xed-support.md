# Milestone 12 — Xed support and grammar coupling

Status: GATED — implementation/static gates complete; live Xed gate pending on the authoritative workstation.

## Implementation

Updated `extra/editors/xed/aiki.lang` to the current lexical surface and removed
runtime/prelude vocabulary from the file. The GtkSourceView definition now owns
presentation only: comments, literals, shapes/symbols, grammar keywords, and
grammar operators.

Added `test/invariant/xed_language_test.go` so Xed's duplicated lexical facts are
executable-coupled to `grammar.ebnfx`. Keyword inventory must match exactly;
every grammar operator must be represented and known stale operators are
rejected.

Added the thin optional Xed/libpeas Python client:

- `extra/editors/xed/aiki_lsp.plugin`
- `extra/editors/xed/aiki_lsp/__init__.py`
- `extra/editors/xed/README.md`

The plugin launches `aiki lsp`, owns only JSON-RPC transport/document
synchronization/diagnostic presentation, and translates LSP UTF-16 positions
back to Xed text iterators. It contains no Aiki syntax, scope, diagnostic, or
formatting authority.

Treecheck now recognizes the Xed editor subtree structurally and requires the
plugin descriptor and Python module to exist as a pair. Directory README
ownership follows the normal active-subtree rule.

## External interface check

Current upstream Xed still documents a configurable plugin system with optional
Python support. Current in-tree Xed plugins use libpeas descriptors with
`Loader=python3` and implement `Xed.WindowActivatable`; the Aiki plugin follows
that public pattern.

No current native/generic LSP client was found in Xed, so a thin plugin remains
the appropriate adapter.

## Validation

Disposable validation only (authoritative tree unchanged):

```text
python3 -m py_compile extra/editors/xed/aiki_lsp/__init__.py

go test ./engine/language/... ./engine/syntax \
  ./cmd/subcommands/tools/lint ./cmd/subcommands/tools/lsp \
  ./cmd/subcommands/tools/treecheck -count=1

go test ./test/invariant -run TestXedLexicalInventoryMatchesGrammar -count=1
```

All passed in a Go-1.23 disposable copy with the graphics backend stubbed only
for compilation. A full disposable treecheck reports `errors=0 orphans=0`.
Generated Python bytecode was removed and is not part of the working tree.

## Handoff gate

This cut cannot be GATED here because its acceptance criterion includes the
actual editor environment. On the authoritative workstation:

1. run `make validate` on the merged tree;
2. install/enable the Xed language definition and Aiki Language Services plugin;
3. open a saved `.ai` file containing a parse error and confirm a live underline;
4. correct the source and confirm the diagnostic clears after the document change.

If the plugin cannot load because the Xed Python/libpeas loader is absent, record
that as an environment limitation; lexical `aiki.lang` support remains valid.

## Next action

Do not begin Cut II.5 until the Xed handoff gate is resolved. After II.4 is
GATED, begin symbol/definition services and the nvi tags adapter.

## Follow-up — live Xed integration findings

Authoritative workstation testing confirmed the protocol path end-to-end:
Xed launched `aiki lsp`, completed initialization, sent `didOpen`/`didChange`,
and received `textDocument/publishDiagnostics` for `let x =`.

Two Xed-adapter defects were exposed and corrected without changing the
language-service or LSP contracts:

- an untitled buffer saved later as `*.ai` was not reconciled into an open LSP
  document; the plugin now tracks URI transitions and sends the corresponding
  close/open notifications;
- an EOF parse diagnostic can legitimately have a zero-width LSP range, which
  GtkTextBuffer cannot visibly underline; the Xed presentation adapter now
  expands only the rendered tag range by one adjacent character when possible,
  leaving the authoritative LSP range unchanged.

Added `make install-xed-plugin` as an idempotent user-local installer. It removes
stale Aiki Xed plugin files before installing the current language definition,
plugin descriptor, and Python package. `make uninstall-xed-plugin` removes the
same Aiki-owned installed artifacts. The install directories are overrideable.

The remaining live gate is to install the corrected plugin, confirm the EOF
parse diagnostic is visible, then correct the source and confirm it clears.

## Live gate

Passed on the user's Xed workstation. `.ai` syntax highlighting is recognized; the plugin launches the development `aiki lsp`; `let x =` receives a visible red diagnostic underline; correcting it to `let x = 42` clears the diagnostic. Save-As URI transition and zero-width EOF diagnostic rendering were corrected during the live gate. `make install-xed-plugin` owns clean user-local installation.
