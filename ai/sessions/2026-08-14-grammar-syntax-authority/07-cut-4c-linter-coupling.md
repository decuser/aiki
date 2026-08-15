# Milestone 07 — Cut 4c linter coupling

Status: GATED

No linter implementation rewrite was needed. Tests inspect the linter's
explicit AST node-type knowledge and require every named type to correspond to
a grammar-produced or documented synthetic node. A second test wraps an
undefined name in an unknown future node and proves generic traversal still
reaches the child.

This preserves the linter's legitimate generic traversal while detecting the
kind of unchecked structural knowledge that previously allowed scope drift.
Existing behavioral tests continue to cover the scope-sensitive cases.

Authoritative `make validate`: passed.
