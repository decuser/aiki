# Milestone 16 — VS Code client

Status: GATED

## Intent

Complete Phase II with a conventional VS Code client that consumes `aiki lsp`
without acquiring language semantics of its own.

## Implementation

Added `extra/editors/vscode/` containing:

- `package.json` — `.ai`/Aiki registration, the `aiki.server.path` launch setting,
  and the `vscode-languageclient` dependency;
- `extension.js` — a thin client that launches the configured executable as
  `aiki lsp` and selects file-backed Aiki documents;
- `language-configuration.json` — comments, brackets, and editor pairing;
- `syntaxes/aiki.tmLanguage.json` — lexical presentation only;
- `README.md` — install, PATH-independent development configuration, live gate,
  and uninstall procedure.

The TextMate keyword/operator inventory is deliberate duplicated presentation
metadata and is executable-coupled to `grammar.ebnfx`, matching the existing
Xed policy. The extension contains no parser, scope model, formatter, builtin
catalog, or semantic documentation.

Added `make install-vscode-plugin` / `make uninstall-vscode-plugin`. The first
installer copied a staged extension directly into the user extension directory;
live testing showed VS Code did not discover that installation reliably. The
installer now uses the supported packaging path entirely out of tree: copy the
extension to a temporary directory, resolve runtime npm dependencies there,
package a VSIX with `@vscode/vsce`, install it with
`code --install-extension --force`, and remove the temporary tree. No
`node_modules`, package lock, or VSIX is created in the Aiki repository.

## Authority and dependency notes

Current Microsoft VS Code guidance continues to model a language-server
extension as a normal JS/TS Language Client that launches a separate Language
Server and communicates through LSP. Current `vscode-languageclient` 10.x is
therefore used rather than implementing protocol handling in the extension.

The earlier Xed PATH discovery is handled directly: `aiki.server.path` may be
set to an absolute development binary path, so desktop VS Code need not inherit
an interactive shell PATH.

## Validation

Source/static checks completed locally:

- `node --check extra/editors/vscode/extension.js`;
- all extension JSON parses;
- treecheck relationships and negative companion tests added;
- executable couplings added for TextMate lexical inventory and thin-client
  registration/launch shape.

Direct npm resolution could not be completed in the tool environment because
external npm network access is unavailable. The installer is designed so this
network dependency occurs only at explicit install time and never during
`make validate`.

## Gate

Pending authoritative `make validate` and live VS Code exercise:

1. install with `make install-vscode-plugin`;
2. set `aiki.server.path` if required;
3. verify `.ai` highlighting;
4. verify diagnostic appears and clears;
5. verify completion;
6. verify hover;
7. verify Go to Definition;
8. verify Format Document matches canonical Aiki formatting.

When those pass, mark II.8 GATED and Phase II COMPLETE/GATED at its handoff.

## Live-gate finding — explicit JSON-RPC null results

The first real VS Code hover request exposed a protocol-envelope defect that the
Xed diagnostics-only client could not exercise. `response.Result` used
`json:\"result,omitempty\"`, so a legitimate LSP null response was serialized
without either `result` or `error`. VS Code rejected that as an invalid JSON-RPC
response.

Corrected the response path so successful replies first marshal their result to
`json.RawMessage`. A nil result therefore serializes explicitly as
`\"result\":null`, while error replies continue to omit `result`. Added a
regression test covering hover miss, definition miss, and shutdown; all three
must contain explicit null results and no error member.

This is a protocol adapter correction only. No language-service semantics or
hover/definition authority changed.

## Live-gate finding — definition caret position

After the JSON-RPC null-result correction, live VS Code hover and formatting
worked, but Go to Definition on `my_slice` reported no definition. The
language-service definition index was keyed by the authoritative start position
of each `NAME` token, while `Definition` attempted lookup using the editor's
raw caret position. VS Code legitimately reports the caret anywhere inside the
identifier, so definition succeeded only when the caret happened to be on the
first character.

Corrected `Service.Definition` to resolve the containing `NAME` token first and
then query the existing lexical definition index by that token's authoritative
position. This matches the already-correct hover behavior and does not change
scope authority. Added a regression case with the caret inside an identifier.

## Live-gate finding — VSIX installation

Manual live packaging established the required build/install flow. `vsce`
requires the runtime `vscode-languageclient` dependency to be present while
packaging, so packaging must occur after `npm install` in the disposable build
tree. Directly copying the staged extension into `~/.vscode/extensions` was not
discovered by the tested VS Code installation; `code --install-extension` of a
VSIX was. The Make target now encodes that proven flow.

The latest `@vscode/vsce` dependency chain emitted Node >=22 engine warnings on
Node 20.19.2 during the live experiment, but packaging completed successfully;
those warnings were not the cause of the earlier failure.


## Authoritative live gate — COMPLETE

After the null-result, definition-position, and VSIX-install corrections, the user exercised the installed extension in VS Code and confirmed the complete client surface: syntax highlighting, diagnostic appearance and clearing, completion, hover, Go to Definition, and canonical Format Document all work. The extension was installed through the out-of-tree VSIX path.

Cut II.8 is GATED. Phase II is complete.
