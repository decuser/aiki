# VS Code support

This directory is a thin VS Code client for `aiki lsp`. VS Code owns editor
registration, lexical presentation, and launch configuration; Aiki semantics
remain in the language-service core and LSP adapter.

## Install for development

From the Aiki repository:

```sh
make install-vscode-plugin
```

Restart VS Code after installation. The installer stages npm dependencies,
packages a VSIX, and invokes `code --install-extension` entirely from a
temporary directory. It does not create `node_modules`, a package lock, or a
VSIX in the Aiki repository.

If desktop-launched VS Code cannot find the development `aiki` executable, set
**Aiki: Server Path** (`aiki.server.path`) to an absolute path, for example:

```text
/home/wsenn/forge/dev/aiki/aiki
```

This is preferable to inferring the editor's PATH from an interactive shell.

## Live gate

Open an existing `.ai` file and verify:

1. lexical highlighting is active;
2. `let x =` produces an Aiki diagnostic and fixing it clears the diagnostic;
3. completion offers visible lexical/prelude names;
4. hover on a source-defined or documented prelude name shows Aiki information;
5. Go to Definition jumps to a source-defined binding;
6. Format Document produces the same canonical result as `aiki fmt`.

All semantic results above must come from `aiki lsp`; this extension contains no
parser, scope model, formatter, builtin list, or semantic documentation.

## Uninstall

```sh
make uninstall-vscode-plugin
```
