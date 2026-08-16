# Milestone 14 — HAL redesign Phase I / Cut I.1

Status: **GATED**

Inventoried all 117 primitives registered by the `v0.4.0-alpha-27` Go runtime.
Each registration was traced to its Go implementation file and production
Aiki/prelude use where present. The full source-derived table is retained in
`proposals/hal-redesign-phase-i-inventory.md`.

Discovery: the registry is substantially broader than a host HAL. It contains
host effects, evaluator intrinsics, value operations, accelerators, tooling,
profiling, tests, REPL controls, and stateful resources in one native namespace.

Adjacent context/runtime state was identified separately rather than forced into
the callable inventory.
