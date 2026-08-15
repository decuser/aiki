# Milestone 17 — Phase II handoff

Status: GATED

## Intent

Close the language-services/adapters phase on authoritative repository and live
editor evidence before resuming the independent Aiki interpreter.

## Authoritative and live evidence

The user reported authoritative `make validate` passing through Cut II.7 and
then exercised the final VS Code client live after the II.8 corrections.

The final VS Code gate established all intended surfaces in a real client:

- `.ai` recognition and lexical highlighting;
- diagnostics appear for malformed source and clear after correction;
- completion works;
- hover works for authored/source-backed information;
- Go to Definition works with the caret anywhere inside an identifier;
- Format Document uses canonical Aiki formatting.

Live testing exposed and corrected three adapter/install defects rather than
language-service authority defects:

1. successful nullable JSON-RPC replies must contain explicit `result: null`;
2. definition requests must first resolve the containing NAME token rather than
   assuming the editor caret is at the token's first byte;
3. VS Code development installation must use a packaged VSIX, with npm/vsce
   work performed out of tree, rather than copying a directory under
   `~/.vscode/extensions`.

The tested VS Code installation also demonstrated that reinstalling the same
extension version can require explicit `code --uninstall-extension aiki.aiki-language-services` first; the exact command is recorded in the
editor installation notes. (Spacing in this sentence is prose only; the actual
extension id is `aiki.aiki-language-services`.)

## Conclusion

Phase II is GATED and complete. Language knowledge remains below replaceable
adapters; Xed, nvi, and VS Code now consume the same underlying authorities.

## Next action

Begin Phase III / Cut III.0: runtime data model and environments for the
independent Aiki evaluator.
