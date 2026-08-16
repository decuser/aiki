# Milestone 16 — HAL redesign Phase I / Cut I.3

Status: **GATED**

Mapped the systems-programmer affordance target against the baseline.

Current strengths include file handles/read/write/random access, arguments,
environment reads, timers, bytes/bits, Store, channels/select, and standard
I/O. Concrete first-order gaps are filesystem metadata/directory operations,
path utilities, process execution, working-directory operations, and a current
time source.

Decision: programmer affordances do not map one-to-one to HAL operations. Pure
operations such as most path manipulation belong in Aiki; rich Aiki modules may
be built over narrower irreducible host contracts.
