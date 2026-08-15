# Milestone 20 — Dynamic value construction surface

Status: GATED locally; authoritative repository gate pending.

## Decision

Adopt two ordinary prelude-level operations:

- `to_symbol("foo")` -> `:foo`. The colon is source notation, not symbol content.
- `shaped(:point, [1, 2])` -> `[@point, 1, 2]`.

`to_symbol` is a new value-construction capability. `shaped` exposes the already-existing HAL shaped-list constructor through the ordinary prelude surface.

## Implementation

Added `_to_symbol` in the HAL substrate and wrappers/documentation/tests for `to_symbol` and `shaped` in the prelude. The self-host front end now uses ordinary shaped errors rather than relying on undeclared local `@error` metadata.

## Evidence

Disposable execution produced `:foo` and `[@point, 1, 2]` from runtime strings/data. HAL package tests pass locally.
