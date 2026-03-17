# Invariant Tests

Verify the machine is shaped correctly — structural properties that must hold.

## handler_validation_test.go

Verifies grammar-evaluator coupling:
- Every grammar production has a handler
- Every token has a handler
- Missing handlers panic at startup
