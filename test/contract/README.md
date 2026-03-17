# Contract Tests

Verify interfaces are satisfied correctly.

## interface_test.go

- `RuntimeContract` — GoRuntime methods return sensible values
- `Value` interface — all concrete types implement Type() and Inspect()
- `Callable` interface — Builtin and Function can be called
