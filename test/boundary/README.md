# Boundary Tests

Verify architectural isolation between layers and scopes.

These tests are part of `make invariant`, not ordinary `make test`, because a
layer crossing can preserve behavior while violating the architecture.

## hal_prelude_user_gating_test.go

Verifies the HAL -> prelude -> user layering is correctly enforced:

- user scope cannot see HAL (`_`-prefixed) names;
- prelude scope can see HAL names;
- user scope sees prelude-exported names after the prelude loads.
