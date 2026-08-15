# Milestone 28 — III.6 `_chars` bootstrap coupling correction

Status: ACTIVE

## Validation finding

Authoritative `make validate` after the III.6 completion drop-in exposed a
bootstrap inventory drift. `lib/string/string.ai` now implements `chars(s)`
through the private HAL primitive `_chars`, but `lib/selfhost/bootstrap.ai`
did not capture `_chars` in its encapsulated HAL capability list. As a result,
self-hosted `string_smoke.ai` and the three-level self-interpretation proof
faulted with `undefined variable: _chars`.

## Correction

- Add `_chars` to the private `hal_bindings()` inventory in
  `lib/selfhost/bootstrap.ai`.
- Extend `TestSelfhostModuleLoading` with a `string.chars("hi")` probe so the
  self-host bootstrap is required to supply the primitive exercised by the
  optimized public `string.chars` implementation.

This does not expose `_chars` to ordinary callers. The bootstrap still exports
only `run`; `_chars` remains a blessed capability installed only into
self-hosted module environments.

## Gate

ACTIVE pending authoritative `make validate`.
