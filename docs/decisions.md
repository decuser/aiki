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

## D3. Derived grammar knowledge is centralized

**Date:** 2026-08-15

**Question:** Should evaluator, formatter, linter, parser diagnostics, and
structural tooling independently walk grammar expressions to derive the
structural facts they need?

**Decision:** No. A parsed grammar now owns one cached `grammar.Analysis`.
Shared structural facts are derived centrally and consumed from that analysis.
Consumers may still interpret grammar expressions directly when that is their
job, and may compile derived facts into local representations optimized for
execution.

The centralized analysis currently owns:

- production names;
- production-referenced token classes;
- grammar-producible AST node types;
- literal terminals structurally contained by each named production;
- one cached newline analysis, with a retained derivation error when the
  Aiki-specific newline prerequisites are unavailable.

The parser's generic EBNF interpretation remains local to the parser.
Enginesmoke's expression rendering remains local presentation logic. Parser
lookup maps for newline filtering remain parser-specific compiled data.
Evaluator exceptions for parser-synthesized nodes such as `TERMINAL` remain
evaluator policy.

**Rationale:** The grammar-authority work made `grammar.ebnfx` the sole source
of syntax, but several consumers still carried their own walkers for answering
the same structural questions. Those walkers did not create a second syntax
specification, but they did create multiple implementations of "what the
grammar says" and therefore another drift surface.

Centralizing the derivation preserves the more important boundary: one
authority for facts, local authority for meaning and policy. It also avoids the
opposite error of moving consumer-specific performance representations or
semantic decisions into the grammar package merely because they use grammar
facts.

**Disposition:** Accepted architectural rule: **parse once; derive shared
structural facts once; consume everywhere.** Apply this rule where concrete
duplicated derivation exists; do not treat it as a mandate for a repository-wide
abstraction campaign.

## D4. `and` and `or` are lazy logical control operators

**Date:** 2026-08-16

**Question:** Should `and` and `or` evaluate both operands like ordinary binary
value operators, or should they short-circuit?

**Decision:** `and` and `or` are lazy logical control operators. They keep their
existing infix surface syntax, but they are evaluated by the evaluator before
ordinary eager binary dispatch. `and` evaluates its left operand first and skips
the right operand when the left value is falsey. `or` evaluates its left operand
first and skips the right operand when the left value is truthy. Function calls
remain eager. Arithmetic and comparison operators remain eager. `not` remains a
unary logical operator over one evaluated operand.

**Rationale:** Aiki's function-call syntax should not lie: `f(a, b)` evaluates
its arguments before calling `f`. Short-circuiting therefore should not be
introduced as `and(a, b)` or `or(a, b)`. But the existing grammar already treats
`and` and `or` as keyword/operator surface forms rather than function calls.
They can therefore honestly belong with evaluation-control constructs such as
`if`, `while`, and `match`.

This preserves left-to-right evaluation while recognizing that logical guards
express conditional dependency between tests. The right operand is not skipped as
an optimization; it is skipped because the logical result is already determined.
This also removes a repeated practical trap in range and bounds checks, where the
right side is only meaningful if the left guard succeeds.

**Disposition:** Accepted semantic correction. Implemented as lazy evaluator
handling for `and` and `or`; no syntax change.

## D5. Natural scalar ordering is shared across comparison and sorting

**Date:** 2026-08-17

**Question:** Should `list.sort(xs)` invent its own notion of natural ordering,
remain limited to numeric `<`, or should Aiki define one scalar ordering relation
shared by comparison operators and sorting?

**Decision:** Aiki defines natural ordering for values of the same scalar type:

- numbers compare by exact numeric value;
- strings compare lexicographically by rune sequence;
- runes compare by Unicode code point;
- symbols compare lexicographically by symbol name.

`<`, `>`, `<=`, and `>=` use this relation. `list.sort(xs)` uses the same relation.
Mixed-type values, booleans, lists, shaped lists, functions, and other composites
remain unordered and produce the ordinary comparison fault. `list.sort(xs, fn)`
supplies an explicit ordering relation when the natural scalar relation is not
applicable.

**Rationale:** A default sort that knows more ordering rules than the comparison
operators would give Aiki two meanings of "comes before." Keeping sorting numeric
only would make ordinary systems work such as filename ordering unnecessarily
ceremonial. The selected scalar domains have straightforward intrinsic orderings;
composites do not need an invented total order merely to make them sortable.

Natural ordering is centralized below the evaluator so comparison operators and
future optimized sorting implementations consume one semantic contract. The
current list implementation remains a stable, non-mutating pure-Aiki merge sort
behind a private implementation seam. A future provider-backed sibling may
accelerate that contract if profiling justifies it, but the portable native path
remains the semantic authority and bare default.

**Disposition:** Accepted semantic refinement. No syntax change and no HAL change.
