# Negative-fixture tooling — 2026-08-14

Status: **ACTIVE — final authoritative validation pending**

Continues from the completed grammar-syntax-authority session. The project
addresses the formatter/linter issue exposed by intentionally unparsable
behavior smoke fixtures and the broader exit-status defects discovered while
investigating it.

## What is true now

- Negative fixture intent is declared textually with `# @negative parse`.
- `parse` is the only supported negative kind; unknown kinds fail.
- Smoke checks declaration against parser-failure observation in both directions
  and applies the same rule before blessing.
- Formatter and linter traversal skip declared parse-negative fixtures.
- Undeclared malformed ordinary source is retained as an error.
- Recursive formatter and lint formatting-preflight traversal accumulate
  malformed-file errors instead of stopping or swallowing them.
- CLI dispatch propagates status codes for all subcommands whose `Run` returns
  an integer status, including fmt, lint, smoke, enginesmoke, debug, profile,
  treecheck, and test.
- The convention is documented in `docs/testing.md`.
- Lint resolves public `use("name")` modules through the same registry model as runtime, while preserving relative path imports.

## Cuts

1. `01-declare-negative-fixtures.md` — explicit negative intent and smoke coupling.
2. `02-failure-semantics.md` — ordinary malformed source becomes a real tool failure.
3. `03-traversal-and-close.md` — multi-file traversal, CLI status propagation, docs/buglist reconciliation.
4. `04-linter-module-resolution.md` — align lint `use(...)` resolution with the runtime module registry after the final gate exposed latent drift.

## Validation

Local environment-limited checks passed for `cmd/internal/testfixture`,
`cmd/subcommands/dev/smoke`, and `cmd/subcommands/tools/fmt` under the available
Go 1.23 compatibility harness. The linter/full-tree gate requires the user's
Go 1.24 environment because this container cannot fetch the Ebitengine module.

## Next action

Merge the Cut 4 overlay into the authoritative tree and rerun the final
`make validate`. If the full gate passes, mark this session COMPLETE.
