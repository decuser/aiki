# Milestone 19 - documentation example coupling fix

Status: **GATED**

## Intent

Resolve executable-documentation failures introduced by the new `file` and
`system` module entries.

## Findings

`lib/file/file.doc` uses `@preamble use("file")`, which imports the module's
exports directly. The new `list` example incorrectly called `file.list(...)`,
so the executable documentation harness failed with `undefined: file`.

The new `read_at`, `write_at`, `system.args`, and `system.env` entries also had
no executable expectation and were not marked `@unchecked`, violating the doc
entry disposition invariant.

## Changes

- Changed the `file.list` example to call `list("lib")` under the existing
  `use("file")` preamble.
- Marked `file.read_at` and `file.write_at` `@unchecked`; their examples require
  prepared file handles/fixtures and are not self-contained doc executions.
- Made `system.args` executable with `type(system.args()) # :list`.
- Made `system.env` executable by checking a deliberately absent variable with
  `is_error(...) # true`.

## Validation

- `TestDocEntryDisposition`: PASS.
- Corrected `file.list` example executed and produced `:list`.
- `system.args` example executed and produced `:list`.
- `system.env` absent-variable example executed and produced `true`.

The full executable-doc invariant was not rerun in the offline harness because
canvas examples require the real Ebiten child lifecycle. The exact failures
reported from the normal development machine are individually resolved by the
checks above.

## Next action

Run `make validate` on the normal Go 1.24 development machine. If it is green,
no further documentation-coupling work is required for these entries.
