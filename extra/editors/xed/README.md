# Aiki support for Xed

`aiki.lang` provides GtkSourceView lexical presentation. It contains only
lexical Aiki facts and is executable-coupled to `grammar.ebnfx`; runtime
builtins are intentionally not duplicated there.

`aiki_lsp.plugin` and `aiki_lsp/` are a thin optional Xed client for
`aiki lsp`. They contain protocol transport, document synchronization, and
visual diagnostic presentation only. All Aiki diagnostics come from the
language-service process.

Install or refresh the user-local integration from the repository with:

```sh
make install-xed-plugin
```

The target removes any previously installed Aiki plugin package, descriptor,
language definition, and Python cache carried inside the old package before
copying the current repository files. By default it installs under
`$XDG_DATA_HOME` when set, otherwise `~/.local/share`: Xed plugins go to
`xed/plugins/` and the language definition goes to
`gtksourceview-4/language-specs/`. Override `XED_PLUGIN_DIR` or `XED_LANG_DIR`
for a nonstandard installation. `make uninstall-xed-plugin` removes only the
Aiki integration.

After installation, restart Xed, enable **Aiki Language Services** in Xed's
plugin preferences, and ensure `aiki` is on Xed's `PATH`.

The plugin uses Xed's current libpeas `WindowActivatable`/Python-3 plugin
surface. Python support is optional in Xed builds; if the Python loader is not
installed, the lexical `aiki.lang` support remains usable but the LSP plugin
cannot load.
