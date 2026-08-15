# Milestone 04 — Cut 3 derive newline structure

Status: GATED

Added grammar-derived analysis for expression endings, statement starts,
expression continuations, ambiguous continuations, overblocked continuations,
uncovered expression endings, and impossible declaration entries.

Derived current facts:

    END(expr): NAME NUMBER RUNE STRING SYMBOL ) ] false true }
    CONTINUATION(expr): ( * + - . / < <= > >= [ and or |>
    AMBIGUOUS: ( - [
    OVERBLOCKED: * + . / < <= > >= and or |>
    UNCOVERED_END: }
    DECLARED_IMPOSSIBLE: none

The uncovered `}` showed that no proposed stronger newline "soundness"
invariant could be adopted without changing language behavior. A twelfth smoke,
`newline_function_end_smoke.ai`, pins the existing ability of a function
literal to continue as a call across a newline. D2 records the decision to
preserve policy and defer syntax redesign.

Authoritative `make validate`: passed after correcting the new smoke gold to
Aiki's transcript encoding. Final smoke count for this cut: 46.
