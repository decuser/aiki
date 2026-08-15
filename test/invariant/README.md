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
10. grammar token authority <-> self-host token inventory (`selfhost_token_authority_test.go`)
11. Go lexer <-> independent Aiki lexer <-> reviewed lexical fixtures (`selfhost_lexer_conformance_test.go`)
12. Go newline normalization <-> independent Aiki normalization <-> reviewed newline fixtures (`selfhost_newline_conformance_test.go`)
13. Go parser <-> independent Aiki parser <-> engine parse-gold corpus (`selfhost_parser_conformance_test.go`)
14. grammar lexical inventory <-> Xed GtkSourceView declaration (`xed_language_test.go`)
15. grammar lexical inventory <-> VS Code TextMate declaration (`vscode_language_test.go`)
16. VS Code language registration/client launch <-> thin `aiki lsp` adapter contract (`vscode_client_test.go`)
17. self-host runtime environment semantics <-> focused lexical-scope probe (`selfhost_runtime_test.go`)
18. reference evaluator <-> independent Aiki expression evaluator (`selfhost_evaluator_expr_test.go`)
19. reference evaluator <-> independent Aiki statement/function evaluator (`selfhost_evaluator_program_test.go`)
20. prelude exports <-> explicit self-host host-value bridge (`selfhost_prelude_bridge_test.go`)
21. reference module behavior <-> self-hosted source-module loader (`selfhost_module_test.go`)
22. behavior gold corpus <-> independent self-host evaluator (`selfhost_behavior_acceptance_test.go`)
23. Go bootstrap -> Aiki interpreter -> Aiki interpreter -> third-level Aiki result (`selfhost_self_interpretation_test.go`)

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


## selfhost_*_test.go

Verify Phase-I self-description/conformance couplings. The independent Aiki
front end does not reuse the Go lexer or parser. Both implementations are
checked against reviewed language-owned artifacts: extracted grammar facts,
lexical/newline conformance fixtures, and the existing engine parse-gold
corpus. Phase III adds runtime-environment, expression, statement/function, prelude-bridge, module-loading, and behavior-acceptance couplings. These compare the independent evaluator against reference behavior while keeping the bridge vocabulary coupled to the real prelude authority and preserving self-hosted Aiki-source module loading. The final self-interpretation invariant runs the Aiki-written interpreter through itself and requires the third-level program to produce the same Aiki result.
