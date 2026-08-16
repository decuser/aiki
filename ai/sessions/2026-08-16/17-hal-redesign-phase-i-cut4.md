# Milestone 17 — HAL redesign Phase I / Cut I.4

Status: **GATED**

Pressure-tested the classification against module loading, spawn, profiling,
file resources, Canvas, time/select, and the Thompson 7094 work.

Key evidence: `time.after` already creates a receive-only Aiki `Channel` from a
Go timer callback, and the existing select evaluator/test treats it as an
ordinary selectable receive endpoint. Host-backed selectable event sources are
therefore existing behavior, not a future hypothetical. This is relevant to the
Canvas pressure case but does not justify a universal transport abstraction.

Phase I is GATED by source inspection/design consistency. No implementation
changes were made.

Next action: Phase II Cut II.1 — decompose `Env`/`EvalContext` into semantic
context facets and trace what accompanies call, spawn, module evaluation, and
host invocation.
