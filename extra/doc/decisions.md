# Aiki Design Decisions

Decisions that resolve open design questions. Each entry records the
question, the decision, the rationale, and the date. The buglist
references these by number.

---

## D1. Float provenance is not tracked after FFI boundary

**Date:** 2026-08-13

**Question:** Should Aiki numbers carry a provenance marker indicating
that a floating-point computation was involved in producing them?

**Decision:** No. Values returned through floating-point FFI are
converted to Aiki numbers at the boundary by representing the returned
float64 exactly as a rational. Approximation inherent in the foreign
computation is not tracked after conversion; thereafter the resulting
value participates in exact rational arithmetic normally.

**Rationale:** The meaningful event is the FFI boundary. `math/ffi`
performs computation using host floating-point arithmetic. Once that
result crosses into Aiki, Aiki receives a value and represents it
exactly as a rational. From that point forward, ordinary Aiki
arithmetic remains exact.

Tracking provenance after the boundary would introduce a second,
shadow numeric semantics — forcing answers to whether provenance
affects equality, display, lists, shapes, channels, serialization,
or whether it can ever be cleared. That turns an implementation-history
concern into a language-design concern.

An internal-only flag would avoid the visible semantic questions, but
if it has no user-facing contract it adds permanent complexity without
changing observable behavior.

The module boundary already communicates the distinction cleanly:

    math/native   Aiki rational and explicitly iterative computation
    math/ffi      host floating-point computation imported into Aiki

Provenance is expressed by which operation was called, not by a tag
carried indefinitely by the resulting number.

**Disposition:** Accepted boundary semantics. Closed. No numeric-model
change.
