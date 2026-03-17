# Canary Tests

Tests that exist solely to fail if a feature disappears.

## tco_test.go

Verifies tail call optimization is present and working:
- Explicit return in tail position
- Implicit tail in if branches
- Implicit tail in match arms
- Stack limit enforced for non-tail recursion
