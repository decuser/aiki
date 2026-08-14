# Aiki bug/drift list

2 open items.

## Priority

| # | Item | Effort |
|---|---|---|
| 1 | Float provenance lost once an FFI result enters Aiki | decision, possibly large |
| 2 | Spawned fault can strand a waiting receiver | decision, possibly large |

---

## 1. Float provenance lost once an FFI result enters Aiki

**Issue.** A value produced by floating-point host machinery is converted into an ordinary Aiki number. Once that happens, its provenance is lost: the value is exact as an Aiki rational representation, but it may represent an approximation that originated in floating-point computation.

This matters most for FFI-backed mathematical functions. A rational such as the exact representation of a host `float64` result can look semantically indistinguishable from a rational produced through exact Aiki arithmetic.

**Decision required.** Determine whether Aiki should:

- accept provenance loss as part of the FFI boundary;
- mark approximate values in some way;
- introduce an explicit approximate-number concept;
- or preserve provenance only through library-level conventions.

Any solution that introduces a second visible numeric category would be a substantial language-design decision and should not be made merely to annotate FFI results.

---

## 2. Spawned fault can strand a waiting receiver

**Issue.** If a spawned computation faults before sending the message another computation is waiting to receive, the receiver may wait indefinitely.

The current concurrency model deliberately avoids task handles, joins, and a separate completion mechanism. Normal completion is represented through ordinary messages.

A fault breaks that protocol because the spawned computation terminates before it can send its expected completion/result message.

**Decision required.** Determine whether fault termination needs some observable completion behavior without introducing a second general task-completion model.

Possible directions include:

- leave fault handling entirely to the spawned protocol;
- require spawned functions to catch/convert faults before communicating;
- propagate a fault value through an expected result channel;
- or add narrowly scoped runtime behavior for abnormal termination.

Any change should preserve Aiki's existing preference for explicit message-passing over futures, joins, or hidden task state.

---

## Recently closed

- `sqrt` documentation now distinguishes `math/ffi` and `math/native` examples.
- Native `iterations` and `terms` are documented as work bounds rather than direct error tolerances.
- Linter block scope now agrees with evaluator scope.
- Fractional canvas `pen_size` truncation is documented.
- `turtle/simple.new(dim)` no longer incorrectly reuses a canvas of the wrong size.
- Report and *This Is Aiki* now agree: Aiki 1 is the specified language; the reference implementation is alpha.
