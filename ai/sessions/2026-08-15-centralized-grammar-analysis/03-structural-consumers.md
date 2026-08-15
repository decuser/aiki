# 03 — Structural consumers

**Status:** GATED

Evaluator now consumes `Analysis.ASTNodeTypes()` and `Analysis.TerminalAlternatives("BINOP")`; the local grammar walkers were deleted. Evaluator policy for synthetic `TERMINAL` remains local.

Lint grammar-knowledge tests now consume `Analysis.ASTNodeTypes()`; the local AST-node grammar walker was deleted.

Formatter coverage consumes the central production inventory while retaining formatter-specific disposition tables.

Tests that deliberately mutate grammar structure call `Reanalyze()` explicitly before exercising coverage.

Available formatter package validation passes. Evaluator/lint full package validation is environment-limited by unavailable Ebitengine dependency.
