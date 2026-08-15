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

---

## D2. Newline policy remains behavior-preserving while grammar analysis is made explicit

**Date:** 2026-08-14

**Question:** Once newline termination is declared in `grammar.ebnfx`, should
grammar-derived analysis immediately impose a stronger startup invariant or
change Aiki's leading-continuation behavior?

**Decision:** No. Preserve the current language behavior in this proposal and
make the structural consequences of the rule explicit and executable. The
grammar now derives expression endings, statement starts, expression
continuations, ambiguous continuations, overblocked continuations, and newline
completion entries that do not correspond to possible expression endings.
These sets are checked against the current grammar and surfaced in the engine
structural gold. No additional hard startup invariant is adopted in this cut.

For the current grammar the derived sets are:

    END(expr)
        NAME NUMBER RUNE STRING SYMBOL ) ] false true }

    CONTINUATION(expr)
        ( * + - . / < <= > >= [ and or |>

    FIRST(statement) intersect CONTINUATION(expr)
        ( - [

    unambiguous continuations blocked by the current rule
        * + . / < <= > >= and or |>

    complete expression endings not covered by the current newline rule
        }

    declared completion entries that cannot end an expression
        none

The uncovered `}` is significant. A function literal is a primary expression,
so a complete expression can end at its closing brace. Because `}` is not in
the current completion set, a newline after a function literal does not force a
statement boundary. A following continuation can therefore attach to the
function literal across that newline. This behavior is now pinned by a behavior
smoke rather than changed here.

**Rationale:** The grammar analysis shows that the existing rule embodies two
independent policy choices rather than one simple soundness condition. It
intentionally prefers a new statement for the ambiguous followers `(`, `[`, and
`-` after declared completion endings, while also forbidding several
unambiguous leading continuations. At the same time, it does not cover every
possible expression ending because `}` is omitted. Turning "all expression
endings must terminate" into a startup invariant would therefore require a
language change, and treating overblocking as an error would require a different
language change. Neither belongs in a behavior-preserving authority refactor.

The useful invariant at this stage is epistemic: the relationship is derived,
visible, and executable. Future syntax work can decide separately whether to
add `}` to the completion set or permit unambiguous leading continuations.

**Disposition:** Grammar analysis accepted. Current newline behavior preserved.
No new hard startup invariant beyond declaration validity. Syntax-policy changes
deferred to a separate proposal.
