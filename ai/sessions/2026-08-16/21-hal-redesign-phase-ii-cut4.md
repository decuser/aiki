# Milestone 21 — HAL redesign Phase II / Cut II.4

Status: **GATED**

Pressure-tested stateful/asynchronous host resources. Baseline `time.after`
already creates a receive-only host-produced Aiki channel and ordinary `select`
consumes it; host-backed selectable events therefore have existing precedent.

Canvas remains a pressure case: resource/session acquisition plus Aiki-defined
protocol and ordinary communication machinery where possible. No Canvas-aware
select and no universal transport abstraction are justified.

Next action: Phase III Cut III.1 — define canonical HAL contract metadata and
separate host operations from evaluator/native/tooling registries.
