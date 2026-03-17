# Boundary Tests

Verify isolation between layers and scopes.

## hal_prelude_user_gating_test.go

Verifies the HAL → prelude → user layering is correctly enforced:
- User scope cannot see HAL (_prefixed) names
- Prelude scope can see HAL names
- User scope sees prelude-exported names after prelude loads
