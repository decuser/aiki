# Summary — Grammar as sole syntax authority

This work began with a narrow contradiction: `grammar.ebnfx` permitted syntax
that the token-normalization pass in `parser.go` could silently forbid. The
specific example was a leading pipeline continuation. The implementation had a
private `isComplete` rule that converted physical newlines to synthetic
semicolons after a fixed set of tokens, with hardcoded delimiter suppression.
The grammar therefore described the parser's productions but not the complete
surface policy seen by programmers.

The first move was observational rather than corrective. Eleven Cut 0 behavior
smokes pinned the existing newline policy before implementation changed. That
immediately corrected an assumption in the proposal: `(`, `[`, and `-` are
ambiguous continuations, but they do not fail. Once the normalizer inserts a
semicolon, each can begin a new expression statement. The current policy thus
resolves ambiguity in favor of a new statement. The forms actually rejected
are unambiguous continuations such as leading `.`, `|>`, and `+`.

Cuts 1 and 2 moved the newline policy into `grammar.ebnfx` without changing
behavior. The grammar now declares the newline token, preceding token classes
and lexemes that trigger termination, and delimiter pairs that suppress it.
The parser consumes compiled data from that declaration; the free
`isComplete` function and its Aiki-specific vocabulary were deleted.

Cut 3 made the grammar relationship executable rather than intuitive. It
derived expression endings, statement starts, continuations, ambiguity, and
overblocking. The analysis confirmed `AMBIGUOUS = ( [ -` and derived the
unambiguous overblocked set `* + . / < <= > >= and or |>`. It also exposed a
more important omission: `}` can end a complete expression because a function
literal is a primary, but `}` is not a newline-completion token. A twelfth smoke
pins the consequence: a function literal can continue as a call across a
newline. D2 records the decision not to change either policy in a
behavior-preserving authority refactor.

Cuts 4a–4d generalized the same authority principle to implementation
couplings. Evaluator validation now closes bidirectionally over 32 productions,
six production-referenced token classes, and synthetic `TERMINAL`; dead lexical
handlers disappeared. Formatter coverage became explicit for all 32
productions, with six parent-handled cases, and unknown leaves can no longer be
silently dropped by generic recursion. The linter retained its appropriate
generic traversal but gained checks that every node type it names belongs to
the grammar and that unknown wrappers still traverse their children. Finally,
`BINOP` membership moved out of evaluator helper duplication: syntax owns the
set, semantics owns operator behavior, and the relationship is checked in both
directions.

Cut 5 connected the declared policy back to the programmer. `help("newline")`
reads grammar metadata, general help points to it, and leftover-token errors
caused by synthetic newline termination now explain that relationship. The
error path uses terminator provenance plus grammar-derived continuation data;
it does not introduce another private list. The grammar structural gold was
regenerated because help metadata is deliberately part of the structural
contract.

The final result is stronger than merely deleting one hardcoded function.
Aiki's grammar now owns the newline surface rule, the parser consumes it, the
engine derives its structural consequences, and formatter, linter, evaluator,
and binary-operator boundaries have explicit drift detection appropriate to
their responsibilities. The remaining syntax questions are visible design
choices rather than accidental implementation behavior.

Aiki may choose restrictive syntax. It should not acquire accidental syntax.
