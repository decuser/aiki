# Invariant Tests

Verify the machine is shaped correctly — structural properties that must hold.

## Executable couplings

Each test enforces that two parts of the distribution agree:

1. grammar productions <-> evaluator handlers (`handler_validation_test.go`)
2. module exports <-> help entries (`lib_help_test.go`)
3. module exports <-> doc entries (`lib_help_test.go`)
4. runtime module roots <-> invariant module discovery (`lib_help_test.go`)
5. help templates <-> function signatures (`lib_help_test.go`)
6. doc examples <-> stated expected values (`doc_examples_test.go`)
7. doc entry disposition: every entry is checked or @unchecked (`doc_examples_test.go`)
8. every shipped module has both help and doc (`lib_help_test.go`, `doc_examples_test.go`)
9. ebiten import confined to one file (`graphics_boundary_test.go`)

## handler_validation_test.go

Verifies grammar-evaluator coupling:
- Every grammar production has a handler
- Every token has a handler
- Missing handlers panic at startup

## lib_help_test.go

Verifies that shipped library modules have complete, well-formed
help and doc files covering all exports. Uses the same module
registry (`DistributionModuleRoots`) that the runtime uses, so the
tests discover exactly the modules users see.

## doc_examples_test.go

Runs every doc entry's example code as an Aiki program. Entries
with `# expected` comments are checked against `inspect()` output.
Entries marked `@unchecked` run the preamble only (verifying the
module loads). Every entry must have one disposition or the other.

## graphics_boundary_test.go

Confirms ebiten is imported by exactly one file and that the
language core packages are free of it.
