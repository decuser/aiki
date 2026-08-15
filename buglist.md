# Aiki bug/drift list

## B1. Recursive fmt/lint traverse intentionally invalid behavior fixtures

**Status:** Open, deferred from the grammar-as-sole-syntax-authority work.

`make validate` runs `./aiki fmt ./...` and `./aiki lint ./...`. The negative
newline behavior fixtures are intentionally unparsable, so those recursive
tool invocations print their expected parser diagnostics even though validation
continues successfully.

Observed fixtures include:

- `test/behavior/break_access_smoke.ai`
- `test/behavior/break_after_smoke.ai`
- `test/behavior/break_binop_smoke.ai`

This is not a parser correctness failure and did not invalidate the grammar
authority work. It is a tooling/fixture-disposition issue: recursive formatter
and linter traversal currently has no convention for files whose purpose is to
be syntactically invalid.

**Disposition:** Do not fix inside the newline authority refactor. Address in a
separate tooling cut, choosing deliberately between excluding known-negative
fixtures from recursive fmt/lint traversal or introducing an explicit fixture
convention understood by those tools.

## Design follow-up (not classified as a bug)

The current newline policy's overblocked continuations and uncovered `}`
expression ending are recorded in `docs/decisions.md` D2. They are intentional
open language-design questions, not implementation drift in this completed
refactor.
