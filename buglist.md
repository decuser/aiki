# Aiki bug/drift list

No open implementation bugs are recorded here at this time.

## Resolved

### B1. Recursive fmt/lint and intentionally invalid behavior fixtures

**Status:** Resolved.

The grammar-as-sole-syntax-authority work introduced behavior smoke specimens
whose intended outcome is a parser failure. Recursive `aiki fmt ./...` and
`aiki lint ./...` originally treated those specimens as ordinary source.
Investigation also exposed two broader tool defects:

- recursive fmt printed parse failures but discarded them, so `fmt` could report
  malformed ordinary source and still return success internally;
- lint's formatting preflight stopped at the first malformed file, and the
  top-level CLI discarded the return codes from both fmt and lint, making both
  commands exit zero at the shell.

The resolution establishes an explicit negative-fixture declaration:

```text
# @negative parse
```

`parse` is currently the only supported negative kind. The smoke runner checks
the declaration against observed parser-failure behavior in both directions;
blessing performs the same check before writing a gold. Formatter and linter
source traversal skip declared parse-negative specimens, while undeclared
malformed source remains an error. Recursive fmt and lint formatting preflight
continue across multiple malformed ordinary files and report the accumulated
failures. The CLI now propagates fmt/lint status codes to the process.

The convention is documented in `docs/testing.md`.

The first full validation after this repair exposed a related pre-existing lint
resolution drift: valid fixtures using `use("list")` were now reached, but lint
looked for a flat filesystem form such as `lib/list.ai` while runtime resolves
public package names through the module registry (`list` ->
`lib/list/list.ai`). Lint public-module resolution now uses the same registry
model as runtime; relative path imports remain path-resolved.

## Design follow-up (not classified as a bug)

The current newline policy's overblocked continuations and uncovered `}`
expression ending are recorded in `docs/decisions.md` D2. They are intentional
open language-design questions, not implementation drift.
