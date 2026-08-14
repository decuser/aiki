# Milestone 15 - language-services assessment

## Question

Should Aiki extract a reusable language-services layer now, before editor or notebook adapters are implemented?

## Findings

Grammar loading and lex/parse construction are duplicated across runner, debug, fmt, lint, enginesmoke, module loading, profiling, and tests. The duplication is real, but the consumers do not all want the same result:

- runner/module loading need executable ASTs;
- debug/enginesmoke need observer-aware intermediate stages;
- formatter needs tokens, comments, and AST-preservation checks;
- linter needs structural analysis and scope information;
- LSP will need diagnostics, hover/help, definition, completion, and formatting;
- Jupyter primarily needs persistent execution, which `runner.Session` already provides.

## Decision

Defer a broad language-services refactor until there is a concrete adapter consumer.

Do not create a large facade merely to remove repeated constructor calls. The current `runner.Session` is already a suitable execution seam for a future Jupyter kernel.

When LSP work begins, extract the smallest authoritative front-end seam first:

1. authoritative grammar loading;
2. parse source in memory while preserving Aiki positions/diagnostics and optional observers;
3. then move reusable lint/format analysis out of `cmd/` as required by the adapter;
4. completion/definition must use actual Aiki visibility/environment rules rather than a second hand-written inventory.

The adapter remains non-authoritative: editor and notebook integrations translate requests to existing Aiki services rather than reimplementing language behavior.

## Restart

The ordered post-profiling queue is complete. No language-services code change is warranted until LSP/Jupyter implementation begins or another concrete consumer requires the same reusable seam.
